package api

import (
	"github.com/Francesco99975/qbittal/internal/models"
	"github.com/labstack/echo/v4"
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
			return c.JSON(400, models.JSONErrorResponse{Code: 400, Message: "Invalid request body", Errors: []string{err.Error()}})
		}

		if !payload.Validate() {
			return c.JSON(400, models.JSONErrorResponse{Code: 400, Message: "Invalid request body", Errors: []string{"Invalid period"}})
		}

		newPattern, err := payload.ToPattern()
		if err != nil {
			return c.JSON(400, models.JSONErrorResponse{Code: 400, Message: "Invalid request body", Errors: []string{err.Error()}})
		}

		err = models.AddPattern(newPattern)
		if err != nil {
			return c.JSON(500, models.JSONErrorResponse{Code: 500, Message: "Error while creating pattern", Errors: []string{err.Error()}})
		}

		return c.JSON(200, newPattern)
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
		pattern.FireTime = payload.FireTime

		err = pattern.Update()
		if err != nil {
			return c.JSON(500, models.JSONErrorResponse{Code: 500, Message: "Error while updating pattern", Errors: []string{err.Error()}})
		}

		return c.JSON(200, pattern)
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
		return c.JSON(200, "Pattern deleted")
	}
}
