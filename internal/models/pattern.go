package models

import (
	"fmt"
	"strconv"
	"time"
)

type Period = string

const (
	Daily   Period = "daily"
	Weekly  Period = "weekly"
	Monthly Period = "monthly"
)

func parsePeriod(period string) (Period, error) {
	switch period {
	case Daily, Weekly, Monthly:
		return Period(period), nil
	default:
		return period, fmt.Errorf("invalid period: %s", period)
	}
}

func isValidPeriod(period string) bool {
	_, err := parsePeriod(period)
	return err == nil
}

func isValidDayIndicator(dayIndicator string, period Period) bool {
	day, err := strconv.Atoi(dayIndicator)
	switch period {
	case Daily:
		return err == nil && day == -1
	case Weekly:
		return err == nil && day >= 0 && day <= 6
	case Monthly:
		return err == nil && day >= 1 && day <= 28
	default:
		return false
	}
}

type PatternPayload struct {
	QueryKeywords  []string  `json:"query_keywords"`
	SearchKeywords []string  `json:"search_keywords"`
	DownloadPath   string    `json:"download_path"`
	Period         string    `json:"period"`
	DayIndicator   string    `json:"day_indicator"`
	FireTime       time.Time `json:"fire_time"`
}

func (p *PatternPayload) Validate() bool {
	return len(p.QueryKeywords) > 0 && len(p.SearchKeywords) > 0 && len(p.DownloadPath) > 0 && isValidPeriod(p.Period) && isValidDayIndicator(p.DayIndicator, p.Period) && p.FireTime.After(time.Now().Add(3*time.Minute))
}

func (p *PatternPayload) ToPattern() (Pattern, error) {
	period, err := parsePeriod(p.Period)
	if err != nil {
		return Pattern{}, err
	}
	return Pattern{
		QueryKeywords:  p.QueryKeywords,
		SearchKeywords: p.SearchKeywords,
		DownloadPath:   p.DownloadPath,
		Period:         period,
		DayIndicator:   p.DayIndicator,
		FireTime:       p.FireTime,
	}, nil
}

type Pattern struct {
	ID             int       `json:"id"`
	QueryKeywords  []string  `json:"query_keywords"`
	SearchKeywords []string  `json:"search_keywords"`
	DownloadPath   string    `json:"download_path"`
	Period         Period    `json:"period"`
	DayIndicator   string    `json:"day_indicator"`
	FireTime       time.Time `json:"fire_time"`
	Created        time.Time `json:"created"`
	Updated        time.Time `json:"updated"`
}
