/**
 * Nanocloud turns any traditional software into a cloud solution, without
 * changing or redeveloping existing source code.
 *
 * Copyright (C) 2016 Nanocloud Software
 *
 * This file is part of Nanocloud.
 *
 * Nanocloud is free software; you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * Nanocloud is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General
 * Public License
 * along with this program.  If not, see
 * <http://www.gnu.org/licenses/>.
 */

// +build windows

package windows

import (
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

type DWord uint32

type luid struct {
	lowPart  uint32
	highPart uint32
}

type luidAndAttributes struct {
	luid       luid
	attributes DWord
}

type tokenPrivileges struct {
	privilegeCount DWord
	privileges     *luidAndAttributes
}

type wtsSessionInfo1 struct {
	execEnvID   DWord
	state       uint32
	sessionID   DWord
	sessionName string
	hostName    string
	userName    string
	domainName  string
	farmName    string
}

type rawWTSSessionInfo1 struct {
	execEnvID   DWord
	state       uint32
	sessionID   DWord
	sessionName *uint16
	hostName    *uint16
	userName    *uint16
	domainName  *uint16
	farmName    *uint16
}

type SID struct {
	Revision            byte
	SubAuthorityCount   byte
	IdentifierAuthority SID_IDENTIFIER_AUTHORITY
}

type SID_IDENTIFIER_AUTHORITY struct {
	Value [6]byte
}

type WTS_PROCESS_INFO struct {
	SessionId    uint32
	ProcessId    uint32
	PProcessName *uint16
	PUserSid     *SID
}

const (
	wtsActive       = 0 // User logged on to WinStation
	wtsConnected    = 1 // WinStation connected to client
	wtsConnectQuery = 2 // In the process of connecting to client
	wtsShadow       = 3 // Shadowing another WinStation
	wtsDisconnected = 4 // WinStation logged on without client
	wtsIdle         = 5 // Waiting for client to connect
	wtsListen       = 6 // WinStation is listening for connection
	wtsReset        = 7 // WinStation is being reset
	wtsDown         = 8 // WinStation is down due to error
	wtsInit         = 9 // WinStation in initialization

	logon32ProviderDefault   = 0
	logonWithProfile         = 1
	logon32LogonInteractive  = 2
	logon32LogonBatch        = 4
	uintptrFlagInherit       = 0x00000001
	createUnicodeEnvironment = 0x00000400

	tokenQuery            = 0x0008
	tokenAdjustPrivileges = 0x0020

	handleFlagInherit = 0x00000001

	startfUseStdHandles = 0x00000100

	sePrivilegeEnabled = DWord(0x00000002)

	normalPriorityClass = 0x00000020
	createNewConsole    = 0x00000010

	MB_OK                = 0x00000000
	MB_OKCANCEL          = 0x00000001
	MB_ABORTRETRYIGNORE  = 0x00000002
	MB_YESNOCANCEL       = 0x00000003
	MB_YESNO             = 0x00000004
	MB_RETRYCANCEL       = 0x00000005
	MB_CANCELTRYCONTINUE = 0x00000006
	MB_ICONHAND          = 0x00000010
	MB_ICONQUESTION      = 0x00000020
	MB_ICONEXCLAMATION   = 0x00000030
	MB_ICONASTERISK      = 0x00000040
	MB_USERICON          = 0x00000080
	MB_ICONWARNING       = MB_ICONEXCLAMATION
	MB_ICONERROR         = MB_ICONHAND
	MB_ICONINFORMATION   = MB_ICONASTERISK
	MB_ICONSTOP          = MB_ICONHAND

	MB_DEFBUTTON1 = 0x00000000
	MB_DEFBUTTON2 = 0x00000100
	MB_DEFBUTTON3 = 0x00000200
	MB_DEFBUTTON4 = 0x00000300
)

//WTSSendMessage func
func WTSSendMessage(sessionID uint, title, content string, style, timeoutSec uint) (int, error) {
	proc, err := loadProc("wtsapi32.dll", "WTSSendMessageA")
	if err != nil {
		return 0, err
	}

	var ret int

	proc.Call(
		0,
		uintptr(sessionID),
		uintptr(unsafe.Pointer(&([]byte(title))[0])),
		uintptr(len(title)),
		uintptr(unsafe.Pointer(&([]byte(content))[0])),
		uintptr(len(content)),
		uintptr(style),
		uintptr(timeoutSec),
		uintptr(unsafe.Pointer(&ret)),
		uintptr(1),
	)

	return ret, nil
}

//WTSGetActiveConsoleSessionID func
func WTSGetActiveConsoleSessionID() (uint, error) {
	proc, err := loadProc("kernel32.dll", "WTSGetActiveConsoleSessionId")
	if err != nil {
		return 0, err
	}
	r0, _, errno := syscall.Syscall(proc.Addr(), 0,
		0,
		0,
		0)
	if r0 == uintptr(0xffffffff) {
		return uint(r0), errors.Errorf("WTSGetActiveConsoleSessionId failed: %v %v", uint(errno), errno)
	}
	return uint(r0), nil
}

//WTSEnumerateProcesses func
func WTSEnumerateProcesses(
	hServer uintptr,
	Reserved, Version uint32,
	ppProcessInfo **WTS_PROCESS_INFO,
	pCount *uint32) (int32, error) {
	proc, err := loadProc("wtsapi32.dll", "WTSEnumerateProcessesW")
	if err != nil {
		return 0, err
	}
	var r1 uintptr
	r1, _, err = proc.Call(
		hServer,
		uintptr(Reserved),
		uintptr(Version),
		uintptr(unsafe.Pointer(ppProcessInfo)),
		uintptr(unsafe.Pointer(pCount)))
	if r1 == 0 {
		return 0, err
	}
	return int32(r1), nil
}

func revertToSelf() error {
	proc, err := loadProc("Advapi32.dll", "RevertToSelf")
	if err != nil {
		return err
	}

	r1, _, err := proc.Call()
	if r1 == 0 {
		return err
	}
	return nil
}

func impersonateLoggedOnUser(token syscall.Token) error {
	proc, err := loadProc("Advapi32.dll", "ImpersonateLoggedOnUser")
	if err != nil {
		return err
	}

	r1, _, err := proc.Call(uintptr(token))
	if r1 == 0 {
		return err
	}
	return nil
}

func getUserProfileDirectory(token syscall.Token) (*uint16, error) {
	proc, err := loadProc("Userenv.dll", "GetUserProfileDirectoryW")
	if err != nil {
		return nil, err
	}

	buffSize := (260 + 1) // (MAX_PATH) * sizeof(WCHAR)
	buff := make([]uint16, buffSize)
	r1, _, err := proc.Call(
		uintptr(token),
		uintptr(unsafe.Pointer(&buff[0])),
		uintptr(unsafe.Pointer(&buffSize)),
	)
	if r1 != 1 {
		return nil, err
	}
	return &buff[0], nil
}

func destroyEnvironmentBlock(env *uint16) error {
	proc, err := loadProc("Userenv.dll", "DestroyEnvironmentBlock")
	if err != nil {
		return err
	}
	r1, _, err := proc.Call(uintptr(unsafe.Pointer(env)))
	if r1 == 0 {
		return err
	}
	return nil
}

func createEnvironmentBlock(token syscall.Token, inherit bool) (*uint16, error) {
	proc, err := loadProc("Userenv.dll", "CreateEnvironmentBlock")
	if err != nil {
		return nil, err
	}

	iInherit := 0
	if inherit {
		iInherit = 1
	}

	var env *uint16

	r1, _, err := proc.Call(
		uintptr(unsafe.Pointer(&env)),
		uintptr(token),
		uintptr(iInherit),
	)

	if r1 == 1 {
		return env, nil
	}
	return nil, err
}

func createProcessWithLogon(
	username string,
	domain string,
	password string,
	logonFlags uint32,
	applicationName string,
	cmd string,
	creationFlags uint32,
	environment uintptr,
	currentDirectory string,
	si *syscall.StartupInfo,
	pi *syscall.ProcessInformation,
) error {
	proc, err := loadProc("advapi32.dll", "CreateProcessWithLogonW")
	if err != nil {
		return err
	}

	r1, _, err := proc.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(username))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(domain))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(password))),
		uintptr(logonFlags),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(applicationName))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(cmd))),
		uintptr(creationFlags),
		environment,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(currentDirectory))),
		uintptr(unsafe.Pointer(si)),
		uintptr(unsafe.Pointer(pi)),
	)
	if r1 == 1 {
		return nil
	}
	return err
}

func logonUser(username, domain, password string, logonType, logonProvider uint32) (hd syscall.Token, err error) {
	proc, err := loadProc("advapi32.dll", "LogonUserW")
	if err != nil {
		return
	}
	r1, _, err := proc.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(username))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(domain))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(password))),
		uintptr(logonType),
		uintptr(logonProvider),
		uintptr(unsafe.Pointer(&hd)),
	)
	if r1 == 1 {
		err = nil
	}
	return
}

func WtsFreeMemory(ptr uintptr) (err error) {
	proc, err := loadProc("Wtsapi32.dll", "WTSFreeMemory")
	if err != nil {
		return
	}

	proc.Call(ptr)
	return
}

func UTF16FromUTF16Ptr(str *uint16) (rt []uint16) {
	for *str != 0 {
		rt = append(rt, *str)
		str = (*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(str)) + 2))
	}
	return
}

func wtsEnumerateSessionsEx(server syscall.Handle) ([]wtsSessionInfo1, error) {
	proc, err := loadProc("Wtsapi32.dll", "WTSEnumerateSessionsExW")
	if err != nil {
		return nil, err
	}

	count := uint32(0)

	var sessionInfo *rawWTSSessionInfo1

	level := uint32(1)
	r1, _, err := proc.Call(
		uintptr(server),
		uintptr(unsafe.Pointer(&level)),
		uintptr(0),
		uintptr(unsafe.Pointer(&sessionInfo)),
		uintptr(unsafe.Pointer(&count)),
	)

	if r1 != 1 {
		return nil, err
	}

	defer WtsFreeMemory(uintptr(unsafe.Pointer(sessionInfo)))

	rt := make([]wtsSessionInfo1, count)

	i := uint32(0)
	for i < count {
		info := wtsSessionInfo1{
			execEnvID: sessionInfo.execEnvID,
			state:     sessionInfo.state,
			sessionID: sessionInfo.sessionID,
		}

		if sessionInfo.sessionName != nil {
			info.sessionName = syscall.UTF16ToString(UTF16FromUTF16Ptr(sessionInfo.sessionName))
		}

		if sessionInfo.hostName != nil {
			info.hostName = syscall.UTF16ToString(UTF16FromUTF16Ptr(sessionInfo.hostName))
		}

		if sessionInfo.userName != nil {
			info.userName = syscall.UTF16ToString(UTF16FromUTF16Ptr(sessionInfo.userName))
		}

		if sessionInfo.domainName != nil {
			info.domainName = syscall.UTF16ToString(UTF16FromUTF16Ptr(sessionInfo.domainName))
		}

		if sessionInfo.farmName != nil {
			info.farmName = syscall.UTF16ToString(UTF16FromUTF16Ptr(sessionInfo.farmName))
		}

		rt[i] = info
		sessionInfo = (*rawWTSSessionInfo1)(unsafe.Pointer(uintptr(unsafe.Pointer(sessionInfo)) + unsafe.Sizeof(*sessionInfo)))
		i++
	}
	return rt, nil
}

func wtsQueryUserToken(sessionID DWord) (hd syscall.Token, err error) {
	proc, err := loadProc("Wtsapi32.dll", "WTSQueryUserToken")
	if err != nil {
		return
	}
	r1, _, err := proc.Call(
		uintptr(sessionID),
		uintptr(unsafe.Pointer(&hd)),
	)
	if r1 == 1 {
		err = nil
	}
	return
}

func createProcessAsUser(
	token syscall.Token,
	applicationName *uint16,
	cmd *uint16,
	procSecurity *syscall.SecurityAttributes,
	threadSecurity *syscall.SecurityAttributes,
	inheritHandles bool,
	creationFlags uint32,
	environment *uint16,
	currentDir *uint16,
	startupInfo *syscall.StartupInfo,
	outProcInfo *syscall.ProcessInformation,
) error {
	proc, err := loadProc("advapi32.dll", "CreateProcessAsUserW")
	if err != nil {
		return err
	}

	iInheritHandles := 0
	if inheritHandles {
		iInheritHandles = 1
	}

	r1, _, err := proc.Call(
		uintptr(token),
		uintptr(unsafe.Pointer(applicationName)),
		uintptr(unsafe.Pointer(cmd)),
		uintptr(unsafe.Pointer(procSecurity)),
		uintptr(unsafe.Pointer(threadSecurity)),
		uintptr(iInheritHandles),
		uintptr(creationFlags),
		uintptr(unsafe.Pointer(environment)),
		uintptr(unsafe.Pointer(currentDir)),
		uintptr(unsafe.Pointer(startupInfo)),
		uintptr(unsafe.Pointer(outProcInfo)),
	)

	if r1 == 1 {
		return nil
	}
	return err
}

func lookupPrivilegeValue(systemName string, name string) (*luid, error) {
	proc, err := loadProc("advapi32.dll", "LookupPrivilegeValueW")
	if err != nil {
		return nil, err
	}

	l := luid{}

	wsSystemName := uintptr(0)
	if len(systemName) > 0 {
		wsSystemName = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(systemName)))
	}

	r1, _, err := proc.Call(
		wsSystemName,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name))),
		uintptr(unsafe.Pointer(&l)),
	)
	if r1 == 1 {
		return &l, nil
	}
	return nil, err
}

func EnablePrivilege(token syscall.Token, privilege string) error {
	uid, err := lookupPrivilegeValue("", privilege)
	if err != nil {
		return err
	}

	return adjustTokenPrivileges(token, *uid)
}

func adjustTokenPrivileges(token syscall.Token, uid luid) error {
	proc, err := loadProc("advapi32.dll", "AdjustTokenPrivileges")
	if err != nil {
		return err
	}

	newState := tokenPrivileges{
		privilegeCount: 1,
		privileges: &luidAndAttributes{
			luid:       uid,
			attributes: sePrivilegeEnabled,
		},
	}

	r1, _, err := proc.Call(
		uintptr(token),
		uintptr(0),
		uintptr(unsafe.Pointer(&newState)),
		uintptr(unsafe.Sizeof(newState)),
		uintptr(0),
		uintptr(0),
	)
	if r1 == 1 {
		return nil
	}
	return err
}

// regSetKeyValue exposes the `RegSetKeyValueW` function in `advapi32.dll`.
// Refer to Windows references for more informations.
func regSetKeyValue(key syscall.Handle, subKey *uint16, valueName *uint16,
	valueType uint32, buf *byte, bufLen uint32) error {
	proc, err := loadProc("advapi32.dll", "RegSetKeyValueW")
	if err != nil {
		return err
	}

	r1, _, err := proc.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(subKey)),
		uintptr(unsafe.Pointer(valueName)),
		uintptr(valueType),
		uintptr(unsafe.Pointer(buf)),
		uintptr(bufLen),
	)
	if r1 == 0 {
		return nil
	}
	return err
}
