//go:build !windows

package main

import (
	"fmt"
	"os/exec"

	"github.com/HanksJCTsai/goidleguard/pkg/logger"
)

// openLogViewer: 彈出獨立的 Terminal 視窗執行 tail -f
func openLogViewer(filename string) {
	// 1. 準備 Shell 指令
	// echo: 顯示標題
	// tail -f: 持續監聽檔案 (等同 PowerShell 的 -Wait)
	// -n 20: 顯示最後 20 行 (等同 PowerShell 的 -Tail 20)
	// 注意：使用單引號 '%s' 包住路徑，防止路徑有空白出錯
	shellCmd := fmt.Sprintf("echo 'Monitoring Log: %s'; tail -f -n 20 '%s'", filename, filename)

	// 2. 準備 AppleScript
	// tell application "Terminal" to do script "..." -> 這會讓 Terminal 開一個新視窗(Tab)跑指令
	// activate -> 把 Terminal 拉到最上層讓你看見
	appleScript := fmt.Sprintf(`tell application "Terminal" to do script "%s"`, shellCmd)
	appleScript += "\ntell application \"Terminal\" to activate"

	// 3. 執行 AppleScript
	cmd := exec.Command("osascript", "-e", appleScript)

	err := cmd.Start()
	if err != nil {
		logger.LogError("Failed to open log viewer:", err)
	}
}

// openFile: macOS 使用 'open' 指令來開啟檔案 (會使用系統預設編輯器)
func openFile(filename string) {
	// 對應 Windows 的 'start'
	cmd := exec.Command("open", filename)
	err := cmd.Start()
	if err != nil {
		logger.LogError("Failed to open file:", err)
	}
}

// showWindowsAlert: 使用 AppleScript 顯示原生對話框
func showWindowsAlert(title, message string) {
	// display dialog: 顯示訊息
	// buttons {"OK"}: 只顯示 OK 按鈕 (預設會有 Cancel)
	// with icon note: 顯示 "i" 圖示
	script := fmt.Sprintf(`display dialog "%s" with title "%s" buttons {"OK"} default button "OK" with icon note`, message, title)
	exec.Command("osascript", "-e", script).Run()
}
