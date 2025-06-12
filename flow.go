package main

/*
#cgo CFLAGS: -Wall
#cgo LDFLAGS: -luser32 -lgdi32
#cgo CFLAGS: -I./winapi
#include "lib/winapi/system_api.c"

// Forward declaration
extern void OnMouseMove(int x, int y);
extern void OnEnabledHotKey();

#define MOD_ALT     0x0001
#define MOD_CONTROL 0x0002
#define MOD_SHIFT   0x0004
#define VK_W        0x57
*/
import "C"

//TODO: move all C / Proc code to system_api.c

import (
	"errors"
	"flag"
	"fmt"
	"github.com/karalabe/hid"
	"log"
	"reflect"
	"strconv"
	"strings"
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

	vendorID  uint16 = 0x046D  //Logitech
	productID uint16 = 0x_c548 //Receiver/devices?
	usage            = 0x02    //short command Logitech API
	page             = 0xff00  //default Logitech custom commands Page

	switchHostFeatureId uint16 = 0x1814 //see https://github.com/Logitech/cpg-docs/blob/master/hidpp20/README.rst
	switchHostFunction         = 0x01
	softwareId                 = 0x0c
	reportId                   = 0x11
	commandLength              = 20
)

const hotkeyID = 1

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

type byteSlice []byte

func (s *byteSlice) String() string {
	return fmt.Sprint(*s)
}

var device hid.Device

var commandsLeft = [][]byte{}
var commandsRight = [][]byte{}
var lastSwitched = time.Now()
var switchThreshold, _ = time.ParseDuration("2s")
var user32 = syscall.NewLazyDLL("user32.dll")
var procSendInput = user32.NewProc("SendInput")
var procSystemMetrics = user32.NewProc("GetSystemMetrics")
var leftBorder = 0
var rightBorder = 0
var enabled = true
var offline = true

func OnHotKey() {
	enabled = !enabled
	fmt.Printf("Enabled: %v\n", enabled)
}

//export OnMouseMove
func OnMouseMove(x C.int, y C.int) {
	xPos := int(x)
	//log.Printf("Mouse moved to: X=%d, Y=%d\n", int(x), int(y))
	if !enabled || lastSwitched.Add(switchThreshold).After(time.Now()) {
		return
	}
	if offline {
		log.Println("Back online")
		lastSwitched = time.Now()
		offline = false
	}
	if xPos <= leftBorder {
		log.Printf("Switching to the left\n")
		onSwitch()
		for _, command := range commandsLeft {
			SwitchHost(command)
		}
		MoveMouse(leftBorder+100, 0)
	} else if xPos >= rightBorder {
		log.Printf("Switching to the right\n")
		onSwitch()
		for _, command := range commandsRight {
			SwitchHost(command)
		}
		MoveMouse(rightBorder-100, 0)
	}
}

func onSwitch() {
	lastSwitched = time.Now()
	offline = true
}

func SwitchHost(command []byte) {
	_, err := device.Write(command)
	if err != nil {
		log.Printf("Unable to send command [%s] to the Device: %v\n", command, err)
	}
}

func MoveMouse(x int, y int) {
	width, _, _ := procSystemMetrics.Call(C.SM_CXSCREEN)
	height, _, _ := procSystemMetrics.Call(C.SM_CYSCREEN)

	rightBorder = int(width)

	var input INPUT
	input.Type = INPUT_MOUSE
	input.MI.Dx = int32(x * 65535 / int(width))
	input.MI.Dy = int32(y * 65535 / int(height))
	input.MI.DwFlags = MOUSEEVENTF_MOVE | MOUSEEVENTF_ABSOLUTE

	_, _, _ = procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&input)),
		unsafe.Sizeof(input),
	)
}

func (s *byteSlice) Set(value string) error {
	for _, part := range strings.Split(value, ",") {
		i, err := strconv.ParseUint(strings.TrimSpace(part), 10, 8)
		if err != nil {
			return err
		}
		*s = append(*s, byte(i))
	}
	return nil
}

func parseArgs() {
	var deviceIds byteSlice
	var leftChannels byteSlice
	var rightChannels byteSlice
	flag.Var(&deviceIds, "devices", "Comma-separated list of local device indexes")
	flag.Var(&leftChannels, "left-channels", "Comma-separated list of channels for devices on the left of current")
	flag.Var(&rightChannels, "right-channels", "Comma-separated list of channels for devices on the right of current")
	flag.Parse()

	fmt.Println("Devices: ", deviceIds)
	fmt.Println("Left Channels: ", leftChannels)
	fmt.Println("Right Channels: ", rightChannels)

	if len(deviceIds) == 0 {
		log.Fatalln("Device list must be provided")
	}

	for idx, dev := range deviceIds {
		if idx < len(leftChannels) {
			ch := leftChannels[idx]
			cmd, err := buildChangeHostCommand(dev, ch)
			if err != nil {
				log.Printf("Unable to build Change Host Command for the Device %v: %v\n", dev, err)
			} else {
				commandsLeft = append(commandsLeft, cmd)
			}
		}
		if idx < len(rightChannels) {
			ch := rightChannels[idx]
			cmd, err := buildChangeHostCommand(dev, ch)
			if err != nil {
				log.Printf("Unable to build Change Host Command for the Device %v: %v\n", dev, err)
			} else {
				commandsRight = append(commandsRight, cmd)
			}
		}
	}

	if len(commandsLeft) == 0 && len(commandsRight) == 0 {
		log.Fatalln("No left or right channels configured")
	}

	logCommands("Left Commands", commandsLeft)
	logCommands("Right Commands", commandsRight)
}

func logCommands(header string, commands [][]byte) {
	if len(commands) == 0 {
		return
	}
	var message = ""
	for _, cmd := range commands {
		message = message + fmt.Sprintf("\t[% x]\n", cmd)
	}
	log.Printf("%s: [\n%s]\n", header, message)
}

// see ./references/x0000_root_v2.pdf
func findFeatureIndex(deviceIdx byte, featureId uint16) (byte, error) {
	msb := byte(featureId >> 8)
	lsb := byte(featureId & 0xFF)
	lookupCommand := make([]byte, commandLength)
	lookupCommand[0] = reportId
	lookupCommand[1] = deviceIdx
	lookupCommand[2] = 0x00
	lookupCommand[3] = softwareId
	lookupCommand[4] = msb
	lookupCommand[5] = lsb
	//log.Printf("Lookup request: % x\n", lookupCommand)
	_, wErr := device.Write(lookupCommand)
	if wErr != nil {
		return 0x00, errors.Join(errors.New("failed to write lookup request"), wErr)
	}

	response := make([]byte, 64)
	var responded = false
	for !responded {
		rn, rErr := device.ReadTimeout(response, 3000)
		if rErr != nil || rn < 5 {
			log.Printf("last read: [% x] header not as expected: [% x]", response, lookupCommand[:4])
			return 0x00, errors.Join(errors.New("failed to read lookup response"), rErr)
		}
		if reflect.DeepEqual(lookupCommand[:4], response[:4]) {
			responded = true
		}
	}
	//log.Printf("Lookup response: % x\n", response)
	return response[4], nil
}

// see ./references/logitech_hidpp_2.0_specification_draft_2012-06-04.pdf
func buildCommand(deviceIdx byte, featureId uint16, function byte, params []byte) ([]byte, error) {
	if params == nil || len(params) > commandLength-4 {
		return nil, errors.New("params is too big")
	}
	idx, err := findFeatureIndex(deviceIdx, featureId)
	if err != nil {
		return nil, err
	}
	result := make([]byte, commandLength)
	result[0] = reportId
	result[1] = deviceIdx
	result[2] = idx
	result[3] = (function << 4) | softwareId
	for pi, p := range params {
		result[4+pi] = p
	}
	return result, nil
}

// see ./references/x1814_change_host_v0.pdf
func buildChangeHostCommand(deviceIdx byte, targetHost byte) ([]byte, error) {
	return buildCommand(deviceIdx, switchHostFeatureId, switchHostFunction, []byte{targetHost})
}

func main() {
	devicesInfos, _ := hid.Enumerate(vendorID, productID)

	for _, deviceInfo := range devicesInfos {
		if deviceInfo.Usage == usage && deviceInfo.UsagePage == page {
			dev, err := deviceInfo.Open()
			fmt.Printf("Opened: %+v\n", deviceInfo)
			if err != nil {
				log.Fatalf("Failed to open device: %v\n", err)
			} else {
				device = dev
			}
			defer func(dev hid.Device) {
				err := dev.Close()
				if err != nil {
					log.Printf("Failed to close device: %v\n", err)
				}
			}(dev)
			break
		}
	}
	if device == nil {
		log.Fatal("No device available.")
	}

	parseArgs()

	width, _, _ := procSystemMetrics.Call(C.SM_CXSCREEN)
	rightBorder = int(width)

	go func() {
		C.SetMouseHook()
	}()

	ok := C.RegisterHotKey(nil, C.int(hotkeyID), C.MOD_CONTROL|C.MOD_SHIFT, C.VK_W)
	if ok == 0 {
		log.Fatal("Failed to register hotkey")
	}
	defer C.UnregisterHotKey(nil, C.int(hotkeyID))

	go func() {
		var msg C.MSG
		for {
			ret := C.GetMessageW(&msg, nil, 0, 0)
			if ret == -1 {
				log.Println("Error in GetMessage")
			} else if msg.message == C.WM_HOTKEY && msg.wParam == hotkeyID {
				OnHotKey()
			}
		}
	}()

	select {}
}
