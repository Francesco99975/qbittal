package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/Francesco99975/qbittal/cmd/boot"
	"github.com/Francesco99975/qbittal/internal/models"
)

func main() {
	err := boot.LoadEnvVariables()
	if err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")

	adminPassword := models.Setup(os.Getenv("DSN"))

	patterns, err := models.GetPatterns()
	if err != nil {
		panic(err)
	}

	boot.SetupCronJobs(patterns)

	e := createRouter()

	go func() {
		e.Logger.Infof("Running Environment: %s", os.Getenv("GO_ENV"))
		if adminPassword != "" {
			e.Logger.Infof("Admin password: %s", adminPassword)
		}
		e.Logger.Fatal(e.Start(":" + port))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
