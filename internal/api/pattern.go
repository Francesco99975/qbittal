package api

import (
	"github.com/Francesco99975/qbittal/internal/helpers"
	"github.com/Francesco99975/qbittal/internal/models"
	"github.com/Francesco99975/qbittal/internal/util"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func GetPatterns() echo.HandlerFunc {
	return func(c echo.Context) error {
		patterns, err := models.GetPatterns()
		if err != nil {
			return c.JSON(500, models.JSONErrorResponse{Code: 500, Message: "Error while getting patterns from DB", Errors: []string{err.Error()}})
		}

		return c.JSON(200, patterns)
	}
}

func CreatePattern() echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload models.PatternPayload

		if err := c.Bind(&payload); err != nil {
			log.Errorf("Error while binding request body: %v", err)
			return c.JSON(400, models.JSONErrorResponse{Code: 400, Message: "Invalid request body", Errors: []string{err.Error()}})
		}

		log.Infof("Payload: %v", payload)

		if !payload.Validate() {
			return c.JSON(400, models.JSONErrorResponse{Code: 400, Message: "Invalid payload", Errors: []string{"Invalid period"}})
		}

		newPattern, err := payload.ToPattern()
		if err != nil {
			return c.JSON(400, models.JSONErrorResponse{Code: 400, Message: "Invalid pattern", Errors: []string{err.Error()}})
		}

		err = models.AddPattern(newPattern)
		if err != nil {
			log.Errorf("Error while creating pattern: %v", err)
			return c.JSON(500, models.JSONErrorResponse{Code: 500, Message: "Error while creating pattern", Errors: []string{err.Error()}})
		}

		err = util.AddJob(newPattern.ID, helpers.ConvertPeriodToCron(newPattern.Period, newPattern.DayIndicator, newPattern.FireHour, newPattern.FireMinute), func() { util.Scraper(newPattern) })
		if err != nil {
			return c.JSON(500, models.JSONErrorResponse{Code: 500, Message: "Error while creating job", Errors: []string{err.Error()}})
		}

		updatedPatterns, err := models.GetPatterns()
		if err != nil {
			return c.JSON(500, models.JSONErrorResponse{Code: 500, Message: "Error while getting patterns from DB", Errors: []string{err.Error()}})
		}

		return c.JSON(200, updatedPatterns)
	}
}

func UpdatePattern() echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload models.PatternPayload
		if err := c.Bind(&payload); err != nil {
			return c.JSON(400, models.JSONErrorResponse{Code: 400, Message: "Invalid request body", Errors: []string{err.Error()}})
		}

		if !payload.Validate() {
			return c.JSON(400, models.JSONErrorResponse{Code: 400, Message: "Invalid request body", Errors: []string{"Invalid period"}})
		}

		id := c.Param("id")
		pattern, err := models.GetPattern(id)
		if err != nil {
			return c.JSON(404, models.JSONErrorResponse{Code: 404, Message: "Pattern not found", Errors: []string{err.Error()}})
		}

		pattern.QueryKeywords = payload.QueryKeywords
		pattern.SearchKeywords = payload.SearchKeywords
		pattern.DownloadPath = payload.DownloadPath
		pattern.Period = payload.Period
		pattern.DayIndicator = payload.DayIndicator

		err = pattern.Update()
		if err != nil {
			return c.JSON(500, models.JSONErrorResponse{Code: 500, Message: "Error while updating pattern", Errors: []string{err.Error()}})
		}

		err = util.UpdateJob(pattern.ID, helpers.ConvertPeriodToCron(pattern.Period, pattern.DayIndicator, pattern.FireHour, pattern.FireMinute), func() { util.Scraper(pattern) })
		if err != nil {
			return c.JSON(500, models.JSONErrorResponse{Code: 500, Message: "Error while updating job", Errors: []string{err.Error()}})
		}

		updatedPatterns, err := models.GetPatterns()
		if err != nil {
			return c.JSON(500, models.JSONErrorResponse{Code: 500, Message: "Error while getting patterns from DB", Errors: []string{err.Error()}})
		}

		return c.JSON(200, updatedPatterns)
	}
}

func DeletePattern() echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		pattern, err := models.GetPattern(id)
		if err != nil {
			return c.JSON(404, models.JSONErrorResponse{Code: 404, Message: "Pattern not found", Errors: []string{err.Error()}})
		}
		err = pattern.Delete()
		if err != nil {
			return c.JSON(500, models.JSONErrorResponse{Code: 500, Message: "Error while deleting pattern", Errors: []string{err.Error()}})
		}

		util.RemoveJob(pattern.ID)

		updatedPatterns, err := models.GetPatterns()
		if err != nil {
			return c.JSON(500, models.JSONErrorResponse{Code: 500, Message: "Error while getting patterns from DB", Errors: []string{err.Error()}})
		}

		return c.JSON(200, updatedPatterns)
	}
}

func ExecutePattern() echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		pattern, err := models.GetPattern(id)
		if err != nil {
			return c.JSON(404, models.JSONErrorResponse{Code: 404, Message: "Pattern not found", Errors: []string{err.Error()}})
		}

		util.Scraper(pattern)
		return c.JSON(200, "OK")
	}
}

// Endpoint for Flutter to get torrent progress
func GetTorrentProgress() echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		util.Mu.RLock()
		dlTorrent, exists := util.DownloadingTorrents[id]
		util.Mu.RUnlock()

		if !exists {
			return c.JSON(404, models.JSONErrorResponse{Code: 404, Message: "Torrent Downloaded", Errors: []string{"Torrent not found"}})
		}

		type TorrentProgress struct {
			Progress float64 `json:"progress"`
		}

		log.Infof("Torrent <%s> progress: %v", dlTorrent.Hash, dlTorrent.Progress)

		return c.JSON(200, TorrentProgress{Progress: dlTorrent.Progress})

	}
}

func DeleteTorrent() echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		util.DeleteTorrent(id, true)

		return c.JSON(200, "OK")

	}
}
