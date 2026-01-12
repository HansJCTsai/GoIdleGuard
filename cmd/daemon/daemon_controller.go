package main

import (
	"strings"
	"time"

	"github.com/HanksJCTsai/goidleguard/internal/config"
	"github.com/HanksJCTsai/goidleguard/internal/preventidle"
	"github.com/HanksJCTsai/goidleguard/internal/schedule"
	"github.com/HanksJCTsai/goidleguard/pkg/logger"
)

type Controller struct {
	cfg        *config.APPConfig
	scheduler  *schedule.Scheduler
	healthStop chan struct{}
}

func NewController(cfg *config.APPConfig) *Controller {
	return &Controller{
		cfg:        cfg,
		scheduler:  schedule.InitialScheduler(cfg),
		healthStop: make(chan struct{}),
	}
}

func (c *Controller) StartDaemon() {
	logger.LogInfo("StartDaemon: will wait for idle >=", c.cfg.IdlePrevention.Interval)
	task := func() {
		if schedule.CheckWorkTime(c.cfg, time.Now()) {
			logger.LogInfo("StartDaemon: idle threshold met, starting prevention")

			idle, err := preventidle.GetIdleTime()
			if err != nil {
				logger.LogError("WaitForIdle:", err)
				return
			}
			logger.LogInfo("WaitForIdle: idle=%v/%v", idle, c.cfg.IdlePrevention.Interval)

			if idle >= c.cfg.IdlePrevention.Interval {
				err := preventidle.SimulateActivity(c.cfg.IdlePrevention.Mode)
				if err != nil {
					logger.LogError("Scheduled SimulateActivity error:", err)
					return
				}
			}
		} else {
			logger.LogInfo("It's not working time now: %s", strings.ToLower(time.Now().Weekday().String()))
			idle, _ := preventidle.GetIdleTime()
			logger.LogInfo("WaitForIdle: idle=%v/%v", idle, c.cfg.IdlePrevention.Interval)
		}
	}

	c.scheduler.ScheduleTask(task)
	// 啟動健康檢查
	go c.healthCheckLoop()
}

func (c *Controller) StopDaemon() {
	logger.LogInfo("Stopping daemon...")
	// 停健康檢查
	close(c.healthStop)
	// 停排程與持續輸入模擬
	c.scheduler.StopScheduler()
	// c.idleCtl.StopIdlePrevention()
}

func (c *Controller) RestartDaemon() {
	logger.LogInfo("Restarting daemon...")
	c.StopDaemon()
	// 確保資源釋放
	time.Sleep(100 * time.Millisecond)
	c.healthStop = make(chan struct{})
	c.scheduler = schedule.InitialScheduler(c.cfg)
	c.StartDaemon()
}

func (c *Controller) healthCheckLoop() {
	ticker := time.NewTicker(c.cfg.Scheduler.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-c.healthStop:
			logger.LogInfo("Health check stopped")
			return
		case <-ticker.C:
			if schedule.CheckWorkTime(c.cfg, time.Now()) {
				idleTime, err := preventidle.GetIdleTime()
				if err != nil {
					logger.LogError("HealthCheck: failed to get idle time:", err)
					continue
				}
				// 如果閒置時間過長（例如 10 分鐘以上），可能代表模擬失效，嘗試重啟
				if idleTime > c.cfg.IdlePrevention.Interval+(5*time.Minute) {
					logger.LogError("HealthCheck: idle time too long (", idleTime, "), restarting prevention")
					c.RestartDaemon()
				} else {
					logger.LogInfo("HealthCheck: idle time healthy (", idleTime, ")")
				}
			}
		}
	}
}
