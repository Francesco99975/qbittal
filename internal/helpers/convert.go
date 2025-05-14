package helpers

import (
	"fmt"
	"strconv"

	"github.com/Francesco99975/qbittal/internal/models"
	"github.com/labstack/gommon/log"
)

func ConvertPeriodToCron(period models.Period, dayIndicator string, fireHour int, fireMinute int) string {
	var cron string

	day, err := strconv.Atoi(dayIndicator)
	if err != nil {
		log.Errorf("Error while converting day to int <- %v", err)
	}

	switch period {
	case models.Daily:
		cron = fmt.Sprintf("%d %d * * *", fireMinute, fireHour)
	case models.Weekly:
		cron = fmt.Sprintf("%d %d * * %d", fireMinute, fireHour, day+1)
	case models.Monthly:
		cron = fmt.Sprintf("%d %d %d * *", fireMinute, fireHour, day+1)
	}

	return cron
}
