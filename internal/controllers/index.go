package controllers

import (
	"net/http"

	"github.com/Francesco99975/qbittal/internal/helpers"
	"github.com/Francesco99975/qbittal/internal/models"
	"github.com/Francesco99975/qbittal/views"
	"github.com/labstack/echo/v4"
)

func Index() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := models.GetDefaultSite("Home")

		html := helpers.MustRenderHTML(views.Index(data))

		return c.Blob(http.StatusOK, "text/html; charset=utf-8", html)
	}
}
