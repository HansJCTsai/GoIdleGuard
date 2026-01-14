package main

import (
	_ "embed"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/HanksJCTsai/goidleguard/internal/config"
	"github.com/HanksJCTsai/goidleguard/pkg/logger"
	"github.com/getlantern/systray"
)

//go:embed icon.ico
var iconData []byte

func main() {

	// 1. 確保 Log 檔案跟執行檔在同一層目錄
	ex, _ := os.Executable()
	logPath := filepath.Join(filepath.Dir(ex), "app.log")

	// 2. 開啟檔案
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err) // 如果連檔案都開不了，直接崩潰讓你知道
	}
	defer logFile.Close()

	// --- 除錯檢查點 (Debug Checkpoint) ---
	// 先繞過 logger 套件，直接用 Go 原生功能寫入一行
	// 如果這行沒出現，代表是檔案權限或路徑問題
	logFile.WriteString("=== SYSTEM STARTUP CHECK: File Write OK ===\n")

	// 3. 設定 Logger 輸出
	// 注意：這裡我們設定之後，千萬不要再呼叫 InitLogger()，否則會被重置！
	logger.SetOutput(logFile)

	// 4. 使用 Logger 寫入測試
	logger.LogInfo("Logger configured successfully. Path: ", logPath)

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logger.LogError("Failed to load config:", err)
		os.Exit(1)
	}
	logger.LogInfo("Config loaded successfully")

	// 建立並啟動 DaemonController
	dc := NewController(cfg)
	onReady := func() {
		// Set icon and tooltip / 設定圖示與提示文字
		// Note: You need to implement getIconData() to return []byte of your icon
		systray.SetIcon(iconData)
		systray.SetTitle("GoIdleGuard")
		systray.SetTooltip("GoIdleGuard is running")

		// 1. 開啟一個獨立的 PowerShell 視窗來監控日誌
		mShowLogs := systray.AddMenuItem("Show Logs (Live)", "Open log viewer")
		systray.AddSeparator()
		// 2. Settings: 點擊後直接打開 config.yaml 讓使用者編輯
		mSettings := systray.AddMenuItem("Settings", "Open config.yaml")
		// 3. About: 點擊後彈出 Windows 原生訊息視窗
		mAbout := systray.AddMenuItem("About", "About GoIdleGuard")
		// 4. Handle menu clicks in a goroutine / 在 goroutine 中處理選單點擊事件
		systray.AddSeparator() // 分隔線
		// 5. Add "Quit" menu item / 新增 "退出" 選單選項
		mQuit := systray.AddMenuItem("Quit", "Quit GoIdleGuard")
		// Handle menu clicks in a goroutine / 在 goroutine 中處理選單點擊事件
		go func() {
			for {
				select {
				case <-mShowLogs.ClickedCh:
					// Open PowerShell to tail the log file
					// 開啟 PowerShell 並執行 Get-Content -Wait (類似 Linux tail -f)
					openLogViewer(logPath)
				case <-mSettings.ClickedCh:
					// User clicked Settings / 使用者點擊了設定
					openFile("config.yaml")
				case <-mAbout.ClickedCh:
					// User clicked About / 使用者點擊了關於
					showWindowsAlert("GoIdleGuard v1.0", "Created by Hanks\n\nRunning in background to keep system awake.")
				case <-mQuit.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()

		// Start the Daemon in a separate goroutine / 在獨立的 goroutine 中啟動 Daemon
		// This prevents blocking the systray UI / 這樣可以避免卡住系統匣的介面
		go func() {
			logger.LogInfo("Starting Daemon...")
			dc.StartDaemon()
			// 捕捉系統中斷訊號以優雅關閉
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh
		}()
	}

	// Define onExit logic / 定義程式退出時的邏輯
	onExit := func() {
		logger.LogInfo("Shutdown signal received, stopping daemon...")
		dc.StopDaemon()
		logger.LogInfo("Daemon stopped; exiting.")
	}

	// Start system tray / 啟動系統匣
	// This will block main thread / 這會卡住主執行緒直到 systray.Quit() 被呼叫
	systray.Run(onReady, onExit)
}

// showWindowsAlert calls the native Windows MessageBoxW API
// 呼叫 Windows 底層 API 顯示彈出視窗 (不需要額外 GUI 套件)
func showWindowsAlert(title, message string) {
	// 載入 user32.dll
	user32 := syscall.NewLazyDLL("user32.dll")
	procMessageBox := user32.NewProc("MessageBoxW")

	// 轉換字串為 UTF-16 指標 (Windows API 要求)
	pTitle, _ := syscall.UTF16PtrFromString(title)
	pMessage, _ := syscall.UTF16PtrFromString(message)

	// 呼叫 MessageBox (0 = NULL, MB_OK = 0, MB_ICONINFORMATION = 0x40)
	// uintptr(0) 代表沒有父視窗
	// uintptr(0x40) 代表顯示 "i" (Information) 圖示
	procMessageBox.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(pMessage)),
		uintptr(unsafe.Pointer(pTitle)),
		uintptr(0x40),
	)
}
