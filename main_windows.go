// +build windows

/*
 * Nanocloud Community, a comprehensive platform to turn any application
 * into a cloud solution.
 *
 * Copyright (C) 2015 Nanocloud Software
 *
 * This file is part of Nanocloud community.
 *
 * Nanocloud community is free software; you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * Nanocloud community is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strings"
	"syscall"

	"main/router"

	"main/windows"
	"main/windows/service"

	"github.com/sirupsen/logrus"
)

// initPlatform should be called when plaza is running as an agent.
func initPlatform() {
	err := windows.SetWinlogonShell("plaza.exe")
	if err != nil {
		log.Fatal(err)
	}
}

// sendShellInfo tries to contact the plaza agent running on the machine to
// inform it about the session open. It just sends his pid and the username of
// the session's owner.
func sendShellInfo() {
	// TODO: send the domain name of the user if any

	currentUser, err := user.Current()
	if err != nil {
		logrus.Error(err)
		return
	}

	splt := strings.Split(currentUser.Username, `\`)
	if len(splt) == 0 {
		logrus.Error("Unable to get the current user")
		return
	}

	mes := struct {
		Username string `json:"username"`
		Pid      int    `json:"pid"`
	}{}

	// Here we should take the current user to fork, because the current user
	// is administrator. But if ldap is activated, it is possibly not Administrator
	// and it's a problem:
	// if the fork is made with another session, all people wanting to connect it
	// througth the administrator would be blocked, because the fork is hosted by
	// another user
	mes.Username = "kek" //splt[len(splt)-1]
	mes.Pid = os.Getpid()

	buf, err := json.Marshal(mes)
	if err != nil {
		logrus.Error(err)
		return
	}

	_, err = http.Post("http://127.0.0.1:9090/shells", "application/json",
		bytes.NewReader(buf))
	if err != nil {
		logrus.Error(err)
	}
}

// main is the main function of plaza.
// plaza takes only one argument that determines in which mode plaza should be
// launched:
//  - install: This will copy plaza to the right plaza (C:\Windows\plaza.exe),
//             create the Windows service for plaza and start it.
//  - service: This will launch plaza as a Windows service. It should only be
//             called by the Windows service manager; never directly.
//  - server: This will launch the plaza server normaly launched in the process.
//            This option exists mainly for test purposes.
//  - shell: This will launch plaza in shell mode. It should only be called by
//           plaza itself.
// If launched without argument, we asume Windows launched plaza as the shell
// application of the session. The shell application of a sessions, is the
// application launched when the user logon to his Windows session. Usually it's
// `C:\Windows\explorer.exe`. This value is set in the registry at:
// `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\Shell`.
// If so, plaza will just execute itself in shell mode in a detached process
// and exit. This; in order to hide the console windows created by Windows for
// the shell application.
func main() {

	if len(os.Args) > 1 {
		// Set Log output to a file
		filename := fmt.Sprintf("C:\\plaza-%s.txt", os.Args[1])
		logFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			logrus.Info("Can't set log file")
		}
		logrus.SetOutput(logFile)

		switch os.Args[1] {

		case "install":
			initPlatform()
			logrus.Info("(re)Installing service")
			err = service.InstallItSelf()
		case "remove":
			logrus.Info("Removing service")
			err = service.RemoveService()
		case "service":
			initPlatform()
			logrus.Info("Run service")
			err = service.Run()
		case "server":
			initPlatform()
			logrus.Info("Start server")
			router.Start()

		case "shell":
			sendShellInfo()

			c := make(chan os.Signal, 1)
			signal.Notify(c)
			<-c
			return
		default:
			err = fmt.Errorf("Invalid action %s. Must be \"install\" or \"service\".", os.Args[1])
		}

		if err != nil {
			logrus.Error(err)
		}
		return
	}

	{
		cmd := exec.Command(`C:\Windows\plaza.exe`, "shell")
		detachedProcess := uint32(0x00000008)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcess,
		}

		err := cmd.Start()
		if err != nil {
			logrus.Error(err)
		}
	}
}
