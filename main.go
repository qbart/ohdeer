package main

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/qbart/ohdeer/deer"
	"github.com/qbart/ohtea/tea"
)

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	e.Logger.Info("Loading config")
	cfg, err := deer.LoadConfig("./ohdeer.hcl")
	if err != nil {
		e.Logger.Fatal(err)
		return
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
	runner := deer.NewRunner(cfg)
	go func() {
		<-runner.Start()
	}()

	tea.SysCallWaitDefault()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	e.Logger.Info("Shutting down the runner")
	runner.Shutdown()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
