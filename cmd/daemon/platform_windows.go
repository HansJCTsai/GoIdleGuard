//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/HanksJCTsai/goidleguard/pkg/logger"
)

// No need for console control constants anymore
// 不再需要控制 Console 的常數，因為我們改用 GUI 模式 + 外部視窗

// openLogViewer spawns a separate PowerShell window to tail the log
// openLogViewer 產生一個獨立的 PowerShell 視窗來跟隨日誌
func openLogViewer(filename string) {
	// 組合 PowerShell 指令
	// -Wait: 持續監聽
	// -Tail 20: 顯示最後 20 行
	// -Encoding UTF8: 確保中文不會亂碼 (VS Code 截圖顯示你是 UTF-8)
	psCommand := fmt.Sprintf("Write-Host 'Monitoring: %s' -ForegroundColor Cyan; Get-Content -Path '%s' -Wait -Tail 20 -Encoding UTF8", filename, filename)

	// 使用 cmd /c start 啟動，這樣會彈出一個全新的視窗，與主程式完全脫鉤
	// 參數結構： cmd /c start "視窗標題" powershell -NoExit -Command "..."
	cmd := exec.Command("cmd", "/c", "start", "GoIdleGuard Logs", "powershell", "-NoLogo", "-NoExit", "-Command", psCommand)

	// 這裡不需要 HideWindow，因為我們就是要彈出視窗
	err := cmd.Start()
	if err != nil {
		logger.LogError("Failed to open log viewer:", err)
	}
}

// openFile 使用預設關聯程式開啟檔案
func openFile(filename string) {
	cmd := exec.Command("cmd", "/c", "start", filename)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Start()
}

// 彈出提示視窗
func showAlert(title, message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	procMessageBox := user32.NewProc("MessageBoxW")
	pTitle, _ := syscall.UTF16PtrFromString(title)
	pMessage, _ := syscall.UTF16PtrFromString(message)
	procMessageBox.Call(0, uintptr(unsafe.Pointer(pMessage)), uintptr(unsafe.Pointer(pTitle)), 0x40)
}
