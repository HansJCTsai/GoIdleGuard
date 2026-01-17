package schedule

import (
	"strings"
	"time"

	"github.com/HanksJCTsai/goidleguard/internal/config"
)

func parseSessionTime(tStr string, now time.Time) (time.Time, error) {
	layout := "15:04"
	paresd, err := time.ParseInLocation(layout, tStr, now.Location())
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(now.Year(), now.Month(), now.Day(), paresd.Hour(), paresd.Minute(), 0, 0, now.Location()), nil
}

// IsTimeInRange 判斷 target 是否介於 start 與 end 之間。
func IsTimeInRange(target, start, end time.Time) bool {
	// return target.After(start) && target.Before(end)
	return (target.Equal(start) || target.After(start)) && target.Before(end)
}

func CheckWorkTime(cfg *config.APPConfig, now time.Time) bool {
	day := strings.ToLower(now.Weekday().String()) // 例如 "monday"
	sessions, exists := cfg.WorkSchedule[day]
	if !exists || len(sessions) == 0 {
		return false
	}
	for _, session := range sessions {
		start, err := parseSessionTime(session.Start, now)
		if err != nil {
			continue
		}
		end, err := parseSessionTime(session.End, now)
		if err != nil {
			continue
		}
		if IsTimeInRange(now, start, end) {
			return true
		}
	}
	return false
}
