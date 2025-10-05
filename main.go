// SPDX-License-Identifier: Apache-2.0
// Copyright 2025 the go-dualsense-battery authors

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
)

var (
	powrprof = windows.NewLazyDLL("powrprof.dll")
	kernel32 = windows.NewLazyDLL("kernel32.dll")
	user32   = windows.NewLazyDLL("user32.dll")

	procSetSuspendState  = powrprof.NewProc("SetSuspendState")
	procGetTickCount     = kernel32.NewProc("GetTickCount")
	procGetLastInputInfo = user32.NewProc("GetLastInputInfo")
	procMessageBox       = user32.NewProc("MessageBoxW")

	timeoutMinutes      *uint
	timeoutMilliseconds uint32

	//go:embed icon.ico
	icon []byte
)

func MessageBox(title, text string) int {
	ret, _, _ := procMessageBox.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		0, // 0 = OK
	)
	return int(ret)
}

func getTickCount() uint32 {
	ret, _, _ := procGetTickCount.Call()
	return uint32(ret)
}

func getIdleMilliseconds() (uint32, error) {
	type LASTINPUTINFO struct {
		CbSize uint32
		DwTime uint32
	}

	lii := LASTINPUTINFO{CbSize: uint32(unsafe.Sizeof(LASTINPUTINFO{}))}
	ret, _, err := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	if ret == 0 {
		return 0, fmt.Errorf("error calling GetLastInputInfo: %v", err)
	}

	tickCount := uint32(getTickCount())
	return tickCount - lii.DwTime, nil
}

func hibernate() error {
	ret, _, err := procSetSuspendState.Call(1, 0, 0) // bHibernate = 1
	if ret == 0 {
		return fmt.Errorf("error calling SetSuspendState: %v", err)
	}
	return nil
}

func hasPowerRequests() (bool, bool, error) {
	cmd := exec.Command("powercfg", "/requests")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("error calling 'powercfg /requests': %v\n%s", err, string(out))
		if strings.Contains(string(out), "requires administrator") {
			return true, true, err
		}
		return true, false, err
	}

	if strings.Count(string(out), "None.") == 6 {
		return false, false, nil
	}
	return true, false, nil
}

func main() {
	timeoutMinutes = flag.Uint("timeout", 20,
		"Will hibernate after this amount of minutes have passed of system inactivity")
	flag.Parse()
	timeoutMilliseconds = uint32(*timeoutMinutes * 60 * 1000)

	onExit := func() {}
	systray.Run(onReady, onExit)
}

func onReady() {
	appName := filepath.Base(os.Args[0])

	systray.SetIcon(icon)

	menuItemInfo := systray.AddMenuItem(appName, "")
	menuItemInfo.Disable()

	systray.AddSeparator()

	menuItemExit := systray.AddMenuItem("Exit", "")
	go func() {
		<-menuItemExit.ClickedCh
		systray.Quit()
	}()

	for {
		hasRequests, requiresAdmin, err := hasPowerRequests()
		if err != nil {
			MessageBox(appName, err.Error())
			if requiresAdmin {
				os.Exit(1)
			}
		}

		idle, err := getIdleMilliseconds()
		if err != nil {
			MessageBox(appName, err.Error())
		}

		systray.SetTooltip(fmt.Sprintf("timeout: %d min, idle: %d ms, hasRequests: %v",
			*timeoutMinutes, idle, hasRequests))

		if idle >= timeoutMilliseconds && !hasRequests {
			if err := hibernate(); err != nil {
				MessageBox(appName, err.Error())
			}
		}

		time.Sleep(10 * time.Second)
	}
}
