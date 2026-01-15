package main

import (
	_ "embed"
	"fmt"
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

const (
	AppName      = "GoIdleGuard"
	AppTooltip   = "GoIdleGuard is running"
	LogFileName  = "app.log"
	ConfFileName = "config.yaml"
)

func main() {
	// 1. 確保 Log 檔案跟執行檔在同一層目錄
	appRoot := resolveAppRoot()

	// 2. 設定路徑 (統一使用 appRoot)
	logPath := filepath.Join(appRoot, LogFileName)
	configPath := filepath.Join(appRoot, ConfFileName)

	// [關鍵修正] 確保 Log 目錄存在！
	// lumberjack 不會自己建立目錄，如果目錄不存在，寫入會失敗
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log dir: %v\n", err)
	}

	// 2. 開啟檔案
	logFile := &lumberjack.Logger{
		Filename:   logPath, // 檔案路徑
		MaxSize:    1,       // 每個 Log 檔案最大 1 MB (超過就切割)
		MaxBackups: 5,       // 最多保留 5 個舊檔案 (超過就刪最舊的)
		MaxAge:     30,      // 舊檔案最多保留 30 天
		Compress:   true,    // 是否壓縮舊檔案 (變成 .gz 以節省空間)
	}
	defer logFile.Close()
	// 讓 logFile 優先被寫入，這樣就算 os.Stdout 在 GUI 模式下報錯也不會影響檔案紀錄
	logger.SetOutput(io.MultiWriter(logFile, os.Stdout))

	logger.LogInfo("=== App Started ===")
	logger.LogInfo("App Root detected: ", appRoot)
	logger.LogInfo("Log Path: ", logPath)
	logger.LogInfo("Config Path: ", configPath)
	logger.LogInfo("=== Log System: Rotation Enabled (10MB limit). Path: ", logPath, " ===")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.LogError("Failed to load config:", err)
		os.Exit(1)
	}
	logger.LogInfo("Config loaded successfully. Path: ", configPath)

	// 建立並啟動 DaemonController
	dc := NewController(cfg)
	onReady := func() {
		setupTrayItems(dc, logPath, configPath)
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

// 輔助函式：為了讓 main 更乾淨，可以把 systray 設定放這裡
func setupTrayItems(dc *Controller, logPath, configPath string) {
	systray.SetIcon(iconData)
	if runtime.GOOS != "darwin" {
		systray.SetTitle("GoIdleGuard")
	}
	systray.SetTooltip("GoIdleGuard is running")

	mShowLogs := systray.AddMenuItem("Show Logs (Live)", "Open log viewer")
	systray.AddSeparator()
	mSettings := systray.AddMenuItem("Settings", "Open config.yaml")
	mAbout := systray.AddMenuItem("About", "About GoIdleGuard")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit GoIdleGuard")

	go func() {
		for {
			select {
			case <-mShowLogs.ClickedCh:
				openLogViewer(logPath)
			case <-mSettings.ClickedCh:
				openFile(configPath)
			case <-mAbout.ClickedCh:
				showWindowsAlert("About", fmt.Sprintf("GoIdleGuard running.\nLog: %s", logPath))
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	go func() {
		logger.LogInfo("Starting Daemon...")
		dc.StartDaemon()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		systray.Quit()
	}()
}

// resolveAppRoot 尋找正確的應用程式目錄
// 優先順序：
// 1. 執行檔所在目錄 (適合正式部署，已有 config.yaml)
// 2. 執行檔的上一層目錄 (適合開發模式，config.yaml 在專案根目錄)
// 3. 當前工作目錄 (Fallback)
func resolveAppRoot() string {
	ex, err := os.Executable()
	if err != nil {
		return "."
	}
	exDir := filepath.Dir(ex)

	// 檢查 1: config.yaml 是否在執行檔旁邊? (例如 bin/config.yaml)
	if _, err := os.Stat(filepath.Join(exDir, "config.yaml")); err == nil {
		return exDir
	}

	// 檢查 2: config.yaml 是否在上一層? (例如 GoIdleGuard/config.yaml)
	// 這能解決你遇到的 "變成 ...\bin" 的問題
	parentDir := filepath.Dir(exDir)
	if _, err := os.Stat(filepath.Join(parentDir, "config.yaml")); err == nil {
		return parentDir
	}

	// 檢查 3: config.yaml 是否在上一層? (例如 GoIdleGuard/cmd/daemon/config.yaml)
	// 這能解決你遇到的 "變成 ...\bin" 的問題
	parentDir2 := filepath.Dir(parentDir)
	if _, err := os.Stat(filepath.Join(parentDir2, "config.yaml")); err == nil {
		return parentDir2
	}

	// 預設回傳執行檔目錄 (即使沒找到，也只好用這裡)
	return exDir
}
