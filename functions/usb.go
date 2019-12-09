package functions

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"main/windows"
	"net/http"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/sirupsen/logrus"
	"github.com/winlabs/gowin32"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type processInfo struct {
	id        uint32
	name      string
	sessionID uint32
}

func Usb() {
	// init COM
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, _ := oleutil.CreateObject("WbemScripting.SWbemLocator")
	defer unknown.Release()

	wmi, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, _ := oleutil.CallMethod(wmi, "ConnectServer")
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	// result is a SWBemObjectSet
	resultRaw, _ := oleutil.CallMethod(service, "ExecNotificationQuery", "SELECT * FROM Win32_VolumeChangeEvent")
	result := resultRaw.ToIDispatch()
	defer result.Release()

	// item is a SWbemObject, but really a Win32_Process
	itemRaw, _ := oleutil.CallMethod(result, "NextEvent")
	item := itemRaw.ToIDispatch()
	defer item.Release()
	//asString, _ := oleutil.GetProperty(item, "DriveName")
	//Проверка работы сервера
	sessionID, err := windows.WTSGetActiveConsoleSessionID()
	if err != nil {
		logrus.Error("WTSGetActiveConsoleSessionI: ", err)
		return
	}
	wtsServ := gowin32.OpenWTSServer("localhost")
	defer wtsServ.Close()
	username, err := wtsServ.QuerySessionUserName(sessionID)
	if err != nil {
		logrus.Error("QuerySessionUserName: ", err)
		return
	}
	//Get PID of explorer
	var pProcessInfo *windows.WTS_PROCESS_INFO
	var count uint32
	var ps []processInfo

	_, err = windows.WTSEnumerateProcesses(uintptr(wtsServ.GetWTSHandle()), 0, 1, &pProcessInfo, &count)
	defer windows.WtsFreeMemory(uintptr(unsafe.Pointer(pProcessInfo)))
	if err != nil {
		logrus.Error("WTSEnumerateProcesses: ", err)
	}

	size := unsafe.Sizeof(windows.WTS_PROCESS_INFO{})
	for i := uint32(0); i < count; i++ {
		p := *(*windows.WTS_PROCESS_INFO)(unsafe.Pointer(uintptr(unsafe.Pointer(pProcessInfo)) + uintptr(size)*uintptr(i)))
		ps = append(ps, processInfo{
			id:        p.ProcessId,
			name:      ole.UTF16PtrToString(p.PProcessName),
			sessionID: p.SessionId,
		})
	}
	var pid int
	for _, p := range ps {
		if p.name == "explorer.exe" && p.sessionID == uint32(sessionID) {
			pid = int(p.id)
		}
	}
	//Берет pid юзера службы, то есть СИСТЕМА
	/*currentUser, err := user.Current()
	if err != nil {
		logrus.Error(err)
		return
	}

	splt := strings.Split(currentUser.Username, `\`)
	if len(splt) == 0 {
		logrus.Error("Unable to get the current user")
		return
	}*/

	mes := struct {
		Username string `json:"username"`
		Pid      int    `json:"pid"`
	}{}

	//Нужен username активной сессии
	mes.Username = username //splt[len(splt)-1]
	//Нужен pid explorer.exe, os.Getpid() берет pid этой службы
	mes.Pid = pid

	buf, err := json.Marshal(mes)
	if err != nil {
		logrus.Error(err)
		return
	}

	//Prepare session
	_, err = http.Post("http://127.0.0.1:9090/shells", "application/json",
		bytes.NewReader(buf))
	if err != nil {
		logrus.Error(err)
	}

	mes2 := struct {
		Username   string   `json:"username"`
		Command    []string `json:"command"`
		HideWindow bool     `json:"hidewindow"`
	}{}

	mes2.Username = username //splt[len(splt)-1]
	mes2.Command = []string{`C:\Windows\System32\notepad.exe`}
	mes2.HideWindow = true

	buf, err = json.Marshal(mes2)
	if err != nil {
		logrus.Error(err)
		return
	}

	//Exec notepad
	_, err = http.Post("http://127.0.0.1:9090/exec", "application/json",
		bytes.NewReader(buf))
	if err != nil {
		logrus.Error(err)
	}
	/*
		//Показ формы
		var answer int
		answer = robotgo.ShowAlert("VAS POSETILA CYBER POLICY ICFA", "Na diske "+
			asString.ToString()+"obnarujeno zapreschennoe anime!")

		if err := exec.Command("cmd", "/C", "logoff").Run(); err != nil {
			fmt.Println("Failed to logoff: ", err)
		}
		if answer == 0 { //Ответ ок

		} else { //Ответ отмена

		}*/

}

//Смена кодировки с UTF-8 в ANSI
func utfToAnsi(str string) (string, error) {
	var windows1251 *charmap.Charmap = charmap.Windows1251
	bs := []byte(str)
	readerBs := bytes.NewReader(bs)
	readerWin := transform.NewReader(readerBs, windows1251.NewEncoder())
	bWin, err := ioutil.ReadAll(readerWin)
	if err != nil {
		return "", err
	}
	return string(bWin), nil
}
