package main

import (
	"fmt"
	"math"
	"os"
	"unsafe"

	"github.com/go-vgo/robotgo"

	"github.com/sirupsen/logrus"
)

const (
	whMouseLl      = 14
	wmLButtonDown  = 513 //0x0201
	wmLButtonUp    = 514 //0x0202
	wmMouseMove    = 512 //0x0200
	wmMouseWheel   = 522 //0x020A
	wmMouseMHWheel = 526 //0x020E
	wmRButtonDown  = 516 //0x0204
	wmRButtonUp    = 517 //0x0205
	targetX        = 100
	targetY        = 100
)

var moved bool
var oldX, oldY int32

// POINT ...
type POINT struct {
	X, Y int32
}

// MSLLHOOKSTRUCT ...
type MSLLHOOKSTRUCT struct {
	Pt          POINT
	MouseData   uintptr
	Flags       uintptr
	Time        uintptr
	DwExtraInfo uintptr
}

//LowLevelMouseProcess ...
func LowLevelMouseProcess(nCode int, wparam uintptr, lparam uintptr) uintptr {
	var temporaryKeyPtr uintptr
	var mouse *MSLLHOOKSTRUCT
	if nCode == 0 && wparam == wmMouseMove {
		mouse = (*MSLLHOOKSTRUCT)(unsafe.Pointer(lparam))
		x := mouse.Pt.X
		y := mouse.Pt.Y
		logrus.Infoln("X: ", x, "Y: ", y)
		fmt.Println("X: ", x, "Y: ", y)
	}
	return CallNextHookEx(temporaryKeyPtr, nCode, wparam, lparam)
}

//run in main()
func mouseStart() {
	// Set Log output to a file
	filename := fmt.Sprintf("C:\\mouse.txt")
	logFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		logrus.Info("Can't set log file")
	}
	logrus.SetOutput(logFile)
	//Start()
	defer user32.Release()
	var msg MSG
	mouseHook := SetWindowsHookEx(whMouseLl, LowLevelMouseProcess, 0, 0)
	for GetMessage(&msg, 0, 0, 0) != 0 {
	}
	UnhookWindowsHookEx(mouseHook)
}

func mouse() {
	//mouseStart()
	var oldX, oldY int
	var targetX, targetY int
	const r = 300
	for {
		x, y := robotgo.GetMousePos()
		if math.Abs(float64(oldX-x)) > 5 || math.Abs(float64(oldY-y)) > 5 {
			oldX = x
			oldY = y
			//Circle
			for i := 0.0; i < 2.0; i = i + 0.15 {
				targetX = int(math.Sin(i*math.Pi)*r) + 800
				targetY = int(math.Cos(i*math.Pi)*r) + 500
				robotgo.MoveMouseSmooth(targetX, targetY)
			}
		}
	}
}
