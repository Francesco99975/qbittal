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

	// Create a root ctx and a CancelFunc which can be used to cancel retentionMap goroutine
	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	port := os.Getenv("PORT")

	adminPassword := models.Setup(os.Getenv("DSN"))

	err = boot.VerifyQbittorrentConnection()
	if err != nil {
		panic(err)
	}

	e := createRouter(ctx)

	go func() {
		e.Logger.Infof("Running Environment: %s", os.Getenv("GO_ENV"))
		e.Logger.Infof("Qbittorrent Server: %s", os.Getenv("QBITTORRENT_API"))
		if adminPassword != "" {
			e.Logger.Infof("Admin password: %s", adminPassword)
		}
		e.Logger.Fatal(e.Start(":" + port))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
