package main

import (
	"errors"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/HanksJCTsai/goidleguard/internal/config"
	"github.com/HanksJCTsai/goidleguard/internal/schedule"
)

type Controller struct {
	cfg       *config.APPConfig
	daemonCmd *exec.Cmd
	scheduler *schedule.Scheduler
	mu        sync.Mutex
}

// NewController 會一併建立 Scheduler
func NewController(cfg *config.APPConfig) *Controller {
	return &Controller{
		cfg:       cfg,
		scheduler: schedule.InitialScheduler(cfg),
	}
}

func updateTime(clock *widget.Label) {
	formatted := time.Now().Format("Time: 03:04:05")
	clock.SetText(formatted)
}

func (c *Controller) IsDaemonRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.daemonCmd == nil || c.daemonCmd.Process == nil {
		return false
	}
	// 在 Windows/Linux 上可以用 Signal(0) 檢查
	err := c.daemonCmd.Process.Signal(syscall.Signal(0))
	if err == nil {
		return true
	}

	if errors.Is(err, syscall.EPERM) {
		return true
	}

	return false
}

func (c *Controller) StartDaemonAndSchedule(onError func(error), onInfo func(string)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.daemonCmd == nil {
		cmd := exec.Command("./bin/app-daemon", "-config=./bin/config.yaml")
		if err := cmd.Start(); err != nil {
			onError(err)
			return
		}
		c.daemonCmd = cmd
	}
	onInfo("Daemon started!")
}

func (c *Controller) StopDaemonAndSchedule(onError func(error), onInfo func(string)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1) 停排程
	c.scheduler.StopScheduler()
	onInfo("Scheduler stopped")

	// 2) 停 Daemon 進程
	if c.daemonCmd != nil && c.daemonCmd.Process != nil {
		if err := c.daemonCmd.Process.Kill(); err != nil {
			onError(err)
			return
		}
		c.daemonCmd = nil
		onInfo("Daemon stopped")
	}
}
