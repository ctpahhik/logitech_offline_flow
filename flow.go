package main

/*
#cgo CFLAGS: -Wall
#cgo LDFLAGS: -luser32
#cgo CFLAGS: -I./winapi
#include "winapi/hook.c"
#include "winapi/hook.h"

// Forward declaration
extern void onMouseMove(int x, int y);
*/
import "C"
import (
	"fmt"
	"github.com/karalabe/hid"
	"log"
	"syscall"
	"time"
	"unsafe"
)

const (
	INPUT_MOUSE                 = 0
	MOUSEEVENTF_MOVE            = 0x0001
	MOUSEEVENTF_ABSOLUTE        = 0x8000
	MOUSEEVENTF_MOVE_NOCOALESCE = 0x2000
	MOUSEEVENTF_VIRTUALDESK     = 0x4000
)

type MOUSEINPUT struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type INPUT struct {
	Type uint32
	MI   MOUSEINPUT
}

var device hid.Device
var switchM = []byte{0x10, 0x04, 0x0a, 0x1d, 0x00, 0x00, 0x00}
var lastSwitched = time.Now()
var threshold, _ = time.ParseDuration("1s")
var user32 = syscall.NewLazyDLL("user32.dll")
var procSendInput = user32.NewProc("SendInput")

//export onMouseMove
func onMouseMove(x C.int, y C.int) {
	xPos := int(x)
	fmt.Printf("Mouse moved to: X=%d, Y=%d\n", int(x), int(y))
	if xPos < 0 && lastSwitched.Add(threshold).Before(time.Now()) {
		fmt.Printf("Switching Mouse\n")
		device.Write(switchM)
		moveMouse(50, 0)
		lastSwitched = time.Now()
	}
}

func moveMouse(x int32, y int32) {
	var input INPUT
	input.Type = INPUT_MOUSE
	input.MI.Dx = x
	input.MI.Dy = y
	input.MI.DwFlags = MOUSEEVENTF_MOVE | MOUSEEVENTF_MOVE_NOCOALESCE

	_, _, _ = procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&input)),
		unsafe.Sizeof(input),
	)
}

func main() {
	const vendorID uint16 = 0x046D
	const productID uint16 = 0x_c548
	const usage = 0x01
	const page = 0xff00

	devicesInfos, _ := hid.Enumerate(vendorID, productID)

	for _, deviceInfo := range devicesInfos {
		if deviceInfo.Usage == usage && deviceInfo.UsagePage == page {
			dev, err := deviceInfo.Open()
			fmt.Printf("Opened: %s\n", deviceInfo.Path)
			if err != nil {
				log.Fatal("Failed to open device: %v", err)
			} else {
				device = dev
			}
			defer dev.Close()
			break
		}
	}
	if device == nil {
		log.Fatal("No device available.")
	}
	C.SetMouseHook()
}
