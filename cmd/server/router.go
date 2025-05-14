package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Francesco99975/qbittal/cmd/boot"
	"github.com/Francesco99975/qbittal/internal/api"
	"github.com/Francesco99975/qbittal/internal/connections"
	"github.com/Francesco99975/qbittal/internal/controllers"
	"github.com/Francesco99975/qbittal/internal/middlewares"
	"github.com/Francesco99975/qbittal/internal/models"
	"github.com/Francesco99975/qbittal/views"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func createRouter(ctx context.Context) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.RemoveTrailingSlash())
	// Apply Gzip middleware, but skip it for /metrics
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/metrics" // Skip compression for /metrics
		},
	}))
	e.Use(middlewares.MonitoringMiddleware())
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	if os.Getenv("GO_ENV") == "development" {
		e.Logger.SetLevel(log.DEBUG)
		log.SetLevel(log.DEBUG)
	} else {
		e.Logger.SetLevel(log.INFO)
		log.SetLevel(log.INFO)
	}
	e.GET("/healthcheck", func(c echo.Context) error {
		time.Sleep(5 * time.Second)
		return c.JSON(http.StatusOK, "OK")
	})

	e.Static("/assets", "./static")

	wsManager := connections.NewManager(ctx)

	go wsManager.Run()

	e.GET("/ws", wsManager.ServeWS)

	patterns, err := models.GetPatterns()
	if err != nil {
		panic(err)
	}

	boot.SetupCronJobs(patterns, wsManager)

	web := e.Group("")

	web.GET("/", controllers.Index())

	admin := web.Group("/admin")
	admin.POST("/login", api.Login())
	admin.POST("/check", api.CheckToken())
	admin.GET("/patterns", api.GetPatterns(), middlewares.IsAuthenticatedAdmin())
	admin.GET("/execute/:id", api.ExecutePattern(), middlewares.IsAuthenticatedAdmin())
	admin.GET("/progress/:id", api.GetTorrentProgress(), middlewares.IsAuthenticatedAdmin())
	admin.DELETE("/execute/:id", api.DeleteTorrent(), middlewares.IsAuthenticatedAdmin())
	admin.POST("/patterns", api.CreatePattern(wsManager), middlewares.IsAuthenticatedAdmin())
	admin.PUT("/patterns/:id", api.UpdatePattern(wsManager), middlewares.IsAuthenticatedAdmin())
	admin.DELETE("/patterns/:id", api.DeletePattern(), middlewares.IsAuthenticatedAdmin())

	e.HTTPErrorHandler = serverErrorHandler

	return e
}

func serverErrorHandler(err error, c echo.Context) {
	// Default to internal server error (500)
	code := http.StatusInternalServerError
	var message any = "An unexpected error occurred"

	// Check if it's an echo.HTTPError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message
	}

	// Check the Accept header to decide the response format
	if strings.Contains(c.Request().Header.Get("Accept"), "application/json") {
		// Respond with JSON if the client prefers JSON
		_ = c.JSON(code, map[string]any{
			"error":   true,
			"message": message,
			"status":  code,
		})
	} else {
		// Prepare data for rendering the error page (HTML)
		data := models.GetDefaultSite("Error")

		// Buffer to hold the HTML content (in case of HTML response)
		buf := bytes.NewBuffer(nil)

		// Render based on the status code

		_ = views.Error(data, fmt.Sprintf("%d", code), err).Render(context.Background(), buf)

		// Respond with HTML (default) if the client prefers HTML
		_ = c.Blob(code, "text/html; charset=utf-8", buf.Bytes())
	}
}
