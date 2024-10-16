package main

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/Francesco99975/qbittal/internal/api"
	"github.com/Francesco99975/qbittal/internal/controllers"
	"github.com/Francesco99975/qbittal/internal/middlewares"
	"github.com/Francesco99975/qbittal/internal/models"
	"github.com/Francesco99975/qbittal/views"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func createRouter() *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RemoveTrailingSlash())
	e.Logger.SetLevel(log.INFO)
	e.GET("/healthcheck", func(c echo.Context) error {
		time.Sleep(5 * time.Second)
		return c.JSON(http.StatusOK, "OK")
	})

	e.Static("/assets", "./static")

	web := e.Group("")

	web.GET("/", controllers.Index())

	admin := web.Group("/admin")
	admin.POST("/login", api.Login())
	admin.POST("/check", api.CheckToken())
	admin.GET("/patterns", api.GetPatterns(), middlewares.IsAuthenticatedAdmin())
	admin.POST("/patterns", api.CreatePattern(), middlewares.IsAuthenticatedAdmin())
	admin.PUT("/patterns/:id", api.UpdatePattern(), middlewares.IsAuthenticatedAdmin())
	admin.DELETE("/patterns/:id", api.DeletePattern(), middlewares.IsAuthenticatedAdmin())

	e.HTTPErrorHandler = serverErrorHandler

	return e
}

func serverErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	data := models.GetDefaultSite("Error")

	buf := bytes.NewBuffer(nil)
	if code < 500 {
		_ = views.ClientError(data, err).Render(context.Background(), buf)

	} else {
		_ = views.ServerError(data, err).Render(context.Background(), buf)
	}

	_ = c.Blob(200, "text/html; charset=utf-8", buf.Bytes())

}
