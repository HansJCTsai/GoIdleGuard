package main

import (
	_ "embed"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/HanksJCTsai/goidleguard/internal/config"
	"github.com/HanksJCTsai/goidleguard/pkg/logger"
	"github.com/getlantern/systray"
	"gopkg.in/natefinch/lumberjack.v2"
)

//go:embed icon.ico
var iconData []byte

func main() {
	appTitle := "GoIdleGuard"
	appTooltip := "GoIdleGuard is running"
	if runtime.GOOS == "darwin" {
		appTitle = ""
		appTooltip = ""
	}
	// 1. 確保 Log 檔案跟執行檔在同一層目錄
	ex, _ := os.Executable()
	logPath := filepath.Join(filepath.Dir(ex), "app.log")

	// 2. 開啟檔案
	logFile := &lumberjack.Logger{
		Filename:   logPath, // 檔案路徑
		MaxSize:    10,      // 每個 Log 檔案最大 10 MB (超過就切割)
		MaxBackups: 5,       // 最多保留 5 個舊檔案 (超過就刪最舊的)
		MaxAge:     30,      // 舊檔案最多保留 30 天
		Compress:   true,    // 是否壓縮舊檔案 (變成 .gz 以節省空間)
	}
	// 記得在程式結束時關閉它
	defer logFile.Close()
	// 3. 設定 Logger 輸出
	logger.SetOutput(io.MultiWriter(os.Stdout, logFile))

	logger.LogInfo("=== SYSTEM STARTUP CHECK: File Write OK ===\n")
	logger.LogInfo("=== Log System: Rotation Enabled (10MB limit). Path: ", logPath, " ===")

	configPath := filepath.Join(filepath.Dir(ex), "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.LogError("Failed to load config:", err)
		os.Exit(1)
	}
	logger.LogInfo("Config loaded successfully. Path: ", configPath)

	// 建立並啟動 DaemonController
	dc := NewController(cfg)
	onReady := func() {
		// Set icon and tooltip / 設定圖示與提示文字
		// Note: You need to implement getIconData() to return []byte of your icon
		systray.SetIcon(iconData)
		systray.SetTitle(appTitle)
		systray.SetTooltip(appTooltip)

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
					openFile(configPath)
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
