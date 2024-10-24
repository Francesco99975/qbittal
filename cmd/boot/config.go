package boot

import (
	"fmt"
	"log"

	"github.com/Francesco99975/qbittal/internal/helpers"
	"github.com/Francesco99975/qbittal/internal/models"
	"github.com/Francesco99975/qbittal/internal/util"
	"github.com/joho/godotenv"
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
		err := util.AddJob(pattern.ID, helpers.ConvertPeriodToCron(pattern.Period, pattern.DayIndicator, pattern.FireHour, pattern.FireMinute), func() { util.Scraper(pattern) })
		if err != nil {
			log.Fatalf("Error while creating job: %v", err)
		}
	}
}
