package api

import (
	"fmt"
	"net/http"

	"github.com/Francesco99975/qbittal/internal/helpers"
	"github.com/Francesco99975/qbittal/internal/models"
	"github.com/labstack/echo/v4"
)

func Login() echo.HandlerFunc {
	return func(c echo.Context) error {

		var payload models.LoginPayload
		if err := c.Bind(&payload); err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body", Errors: []string{err.Error()}})
		}

		admin, err := models.GetAdminFromDB()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: fmt.Sprintf("Error while getting admin from DB. Cause -> %v", err)})
		}

		err = admin.VerifyPassword(payload.Password)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, models.JSONErrorResponse{Code: http.StatusUnauthorized, Message: fmt.Sprintf("Unauthorized: wrong password. Cause -> %v", err)})
		}

		token, err := admin.GenerateToken()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.JSONErrorResponse{Code: http.StatusInternalServerError, Message: fmt.Sprintf("Error while generating token. Cause -> %v", err)})
		}

		return c.JSON(http.StatusOK, models.LoginInfo{Token: token})
	}
}

func CheckToken() echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload models.TokenInfo
		if err := c.Bind(&payload); err != nil {
			return c.JSON(http.StatusBadRequest, models.JSONErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body", Errors: []string{err.Error()}})
		}

		_, err := helpers.ValidateToken(payload.Token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, models.JSONErrorResponse{Code: http.StatusUnauthorized, Message: fmt.Sprintf("Unauthorized. Cause -> %v", err)})
		}

		return c.JSON(http.StatusOK, models.CheckResponse{Valid: true})
	}
}
