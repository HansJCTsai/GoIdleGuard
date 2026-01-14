//go:build !windows

package main

import (
	"os/exec"
	"runtime"

	"github.com/HanksJCTsai/goidleguard/pkg/logger"
)

// Unix 系統通常不支援動態隱藏/顯示 Console
const SupportsConsoleControl = false

func initPlatform() {
	// Do nothing on Unix
}

// 為了滿足 main.go 的呼叫，即使不支援也要定義空函式
func showConsole() {}
func hideConsole() {}

func openFile(filename string) {
	var cmdName string
	var args []string

	if runtime.GOOS == "darwin" {
		cmdName = "open"
		args = []string{filename}
	} else {
		// Linux
		cmdName = "xdg-open"
		args = []string{filename}
	}

	cmd := exec.Command(cmdName, args...)
	err := cmd.Start()
	if err != nil {
		logger.LogError("Failed to open file:", err)
	}
}

func showAlert(title, message string) {
	if runtime.GOOS == "darwin" {
		// macOS: 使用 AppleScript 顯示對話框
		script := `display dialog "` + message + `" with title "` + title + `" buttons {"OK"} default button "OK" with icon note`
		exec.Command("osascript", "-e", script).Run()
	} else {
		// Linux: 嘗試使用 zenity (如果有的話)，否則只寫 Log
		// 你也可以考慮安裝 'notify-send'
		if _, err := exec.LookPath("zenity"); err == nil {
			exec.Command("zenity", "--info", "--title="+title, "--text="+message).Run()
		} else {
			logger.LogInfo("ALERT [" + title + "]: " + message)
		}
	}
}
