package boot

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"github.com/Francesco99975/qbittal/internal/helpers"
	"github.com/Francesco99975/qbittal/internal/models"
	"github.com/Francesco99975/qbittal/internal/util"
	"github.com/joho/godotenv"
	"github.com/labstack/gommon/log"
)

func LoadEnvVariables() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("cannot load environment variables")
	}

	return err
}

func SetupCronJobs(patterns []models.Pattern) {
	for _, pattern := range patterns {
		err := util.AddJob(pattern.ID, helpers.ConvertPeriodToCron(pattern.Period, pattern.DayIndicator, pattern.FireHour, pattern.FireMinute), func() {
			err := util.Scraper(pattern)
			if err != nil {
				log.Errorf("Error while scraping: %v", err)
			}
		})
		if err != nil {
			log.Errorf("Error while creating job: %v", err)
		}
	}
}

func VerifyQbittorrentConnection() error {
	var jar, _ = cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	err := util.QbittLogin(client)
	if err != nil {
		return fmt.Errorf("failed to login to qbittorrent <- %w", err)
	}

	err = util.QbittLogout(client)
	if err != nil {
		return fmt.Errorf("failed to logout from qbittorrent <- %w", err)
	}

	return nil
}
