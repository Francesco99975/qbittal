package middlewares

import (
	"fmt"
	"net/http"

	"github.com/Francesco99975/qbittal/internal/helpers"
	"github.com/Francesco99975/qbittal/internal/models"
	"github.com/labstack/echo/v4"
)

func IsAuthenticatedAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("token")
			if err != nil {
				return c.JSON(http.StatusUnauthorized, models.JSONErrorResponse{Code: http.StatusUnauthorized, Message: "Unauthorized. Cause -> Token not provided"})
			}

			token := cookie.Value
			if token == "" {
				return c.JSON(http.StatusUnauthorized, models.JSONErrorResponse{Code: http.StatusUnauthorized, Message: "Unauthorized. Cause -> Token not provided"})
			}

			_, err = helpers.ValidateToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, models.JSONErrorResponse{Code: http.StatusUnauthorized, Message: fmt.Sprintf("Unauthorized. Cause -> %v", err), Errors: []string{err.Error()}})
			}

			return next(c)
		}
	}
}
