package models

import (
	"fmt"
	"strconv"
	"strings"
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

type Source = string

const (
	Nyaa      Source = "nyaa.si"
	PirateBay Source = "thepiratebay.org"
)

func parseSource(source string) (Source, error) {
	switch source {
	case Nyaa, PirateBay:
		return Source(source), nil
	default:
		return source, fmt.Errorf("invalid source: %s", source)
	}
}

func isValidSource(source string) bool {
	_, err := parseSource(source)
	return err == nil
}

type PatternPayload struct {
	Source         string    `json:"source"`
	QueryKeywords  []string  `json:"query_keywords"`
	SearchKeywords []string  `json:"search_keywords"`
	DownloadPath   string    `json:"download_path"`
	Period         string    `json:"period"`
	DayIndicator   string    `json:"day_indicator"`
	FireTime       time.Time `json:"fire_time"`
}

func (p *PatternPayload) Validate() bool {
	return isValidSource(p.Source) && len(p.QueryKeywords) > 0 && len(p.SearchKeywords) > 0 && len(p.DownloadPath) > 0 && isValidPeriod(p.Period) && isValidDayIndicator(p.DayIndicator, p.Period) && p.FireTime.After(time.Now().Add(3*time.Minute))
}

func (p *PatternPayload) ToPattern() (Pattern, error) {
	period, err := parsePeriod(p.Period)
	if err != nil {
		return Pattern{}, err
	}

	source, err := parseSource(p.Source)
	if err != nil {
		return Pattern{}, err
	}
	return Pattern{
		Source:         source,
		QueryKeywords:  p.QueryKeywords,
		SearchKeywords: p.SearchKeywords,
		DownloadPath:   p.DownloadPath,
		Period:         period,
		DayIndicator:   p.DayIndicator,
		FireTime:       p.FireTime,
	}, nil
}

type Pattern struct {
	ID             string    `json:"id"`
	Source         Source    `json:"source"`
	QueryKeywords  []string  `json:"query_keywords"`
	SearchKeywords []string  `json:"search_keywords"`
	DownloadPath   string    `json:"download_path"`
	Period         Period    `json:"period"`
	DayIndicator   string    `json:"day_indicator"`
	FireTime       time.Time `json:"fire_time"`
	Created        time.Time `json:"created"`
	Updated        time.Time `json:"updated"`
}

func GetPatterns() ([]Pattern, error) {
	rows, err := db.Query("SELECT * FROM patterns")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []Pattern

	for rows.Next() {
		var p PatternDB
		err := rows.Scan(&p.ID, &p.Source, &p.Query, &p.Search, &p.DownloadPath, &p.Period, &p.Dayind, &p.FireTime, &p.Created, &p.Updated)
		if err != nil {
			return nil, err
		}

		patterns = append(patterns, p.ToPattern())

		if err := rows.Err(); err != nil {
			return nil, err
		}

	}

	return patterns, nil
}

func GetPattern(id string) (Pattern, error) {
	row := db.QueryRow("SELECT * FROM patterns WHERE id = $1", id)
	var p PatternDB
	err := row.Scan(&p.ID, &p.Source, &p.Query, &p.Search, &p.DownloadPath, &p.Period, &p.Dayind, &p.FireTime, &p.Created, &p.Updated)
	if err != nil {
		return Pattern{}, err
	}

	return p.ToPattern(), nil
}

func AddPattern(p Pattern) error {
	query := "INSERT INTO patterns (source, query, search, dlpath, period, dayind, firetime) VALUES ($1, $2, $3, $4, $5, $6, $7, $9, $10)"
	_, err := db.Exec(query, p.Source, strings.Join(p.QueryKeywords, ","), strings.Join(p.SearchKeywords, ","), p.DownloadPath, p.Period, p.DayIndicator, p.FireTime)
	return err
}

func (p *Pattern) Update() error {
	query := "UPDATE patterns SET source = $1, query = $2, search = $3, dlpath = $4, period = $5, dayind = $6, firetime = $7 WHERE id = $8"
	_, err := db.Exec(query, p.Source, strings.Join(p.QueryKeywords, ","), strings.Join(p.SearchKeywords, ","), p.DownloadPath, p.Period, p.DayIndicator, p.FireTime, p.ID)
	return err
}

func (p *Pattern) Delete() error {
	query := "DELETE FROM patterns WHERE id = $1"
	_, err := db.Exec(query, p.ID)
	return err
}

type PatternDB struct {
	ID           string    `db:"id"`
	Source       string    `db:"source"`
	Query        string    `db:"query"`
	Search       string    `db:"search"`
	DownloadPath string    `db:"dlpath"`
	Period       string    `db:"period"`
	Dayind       string    `db:"dayind"`
	FireTime     time.Time `db:"firetime"`
	Created      time.Time `db:"created"`
	Updated      time.Time `db:"updated"`
}

func (p *PatternDB) ToPattern() Pattern {
	return Pattern{
		ID:             p.ID,
		Source:         Source(p.Source),
		QueryKeywords:  strings.Split(p.Query, ","),
		SearchKeywords: strings.Split(p.Search, ","),
		DownloadPath:   p.DownloadPath,
		Period:         Period(p.Period),
		DayIndicator:   p.Dayind,
		FireTime:       p.FireTime,
		Created:        p.Created,
		Updated:        p.Updated,
	}
}