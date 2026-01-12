package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/HanksJCTsai/goidleguard/internal/config"
)

func main() {
	a := app.NewWithID("com.example.goidleguard")
	w := a.NewWindow("GoIdleGuard GUI")

	cfg, err := config.LoadConfig("./bin/config.yaml")
	if err != nil {
		dialog.ShowError(err, w)
	}

	ctrl := NewController(cfg)

	startAllBtn := widget.NewButton("Start Daemon & Scheduler", func() {
		ctrl.StartDaemonAndSchedule(
			func(err error) { dialog.ShowError(err, w) },
			func(msg string) { dialog.ShowInformation("Info", msg, w) },
		)
	})

	stopAllBtn := widget.NewButton("Stop Daemon & Scheduler", func() {
		ctrl.StopDaemonAndSchedule(
			func(err error) { dialog.ShowError(err, w) },
			func(msg string) { dialog.ShowInformation("Info", msg, w) },
		)
	})

	w.SetContent(container.NewVBox(
		// ... 你的 config 欄位
		container.NewHBox(startAllBtn, stopAllBtn),
	))

	w.ShowAndRun()
}
