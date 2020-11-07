package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/qbart/ohdeer/deer"
	"github.com/qbart/ohdeer/deerstore"
	"github.com/qbart/ohtea/tea"
)

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	e.Logger.Info("Loading config")
	cfg, err := deer.LoadConfig("./ohdeer.hcl")
	if err != nil {
		e.Logger.Fatal(err)
	}

	ctx := context.Background()

	e.Logger.Info("Connecting to store")
	store, err := deerstore.NewTimescaleDB(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer store.Close(ctx)
	e.Logger.Info(ctx, "Init store")
	if err := store.Init(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Info("Starting server")
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	go func() {
		if err := e.Start(":1820"); err != nil {
			e.Logger.Info("Shutting down the server")
		}
	}()

	e.Logger.Info("Starting jobs")
	runner := deer.NewRunner(cfg, store)
	go runner.Start(ctx)

	tea.SysCallWaitDefault()
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	e.Logger.Info("Shutting down the runner")
	runner.Shutdown(timeoutCtx)
	if err := e.Shutdown(timeoutCtx); err != nil {
		e.Logger.Fatal(err)
	}
}
