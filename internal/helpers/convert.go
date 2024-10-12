package helpers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Francesco99975/qbittal/internal/models"
	"github.com/labstack/gommon/log"
)

func ConvertPeriodToCron(period models.Period, dayIndicator string, fireTime time.Time) string {
	var cron string
	hour := fireTime.Hour()
	minute := fireTime.Minute()

	day, err := strconv.Atoi(dayIndicator)
	if err != nil {
		log.Errorf("Error while converting day to int <- %v", err)
	}

	switch period {
	case models.Daily:
		cron = fmt.Sprintf("%d %d * * *", minute, hour)
	case models.Weekly:
		cron = fmt.Sprintf("%d %d * * %d", minute, hour, day)
	case models.Monthly:
		cron = fmt.Sprintf("%d %d %d * *", minute, hour, day)
	}

	return cron
}
