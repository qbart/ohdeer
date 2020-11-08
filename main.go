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

	e.Logger.Info("Connecting to store")
	store, err := deerstore.NewTimescaleDB(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer store.Close(context.Background())
	e.Logger.Info(context.Background(), "Init store")
	if err := store.Init(context.Background()); err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Info("Starting server")
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Oh! Deer!")
	})
	e.GET("/api/v1/config", func(c echo.Context) error {
		return c.JSON(http.StatusOK, buildConfigResp(cfg))
	})
	e.GET("/api/v1/metrics", func(c echo.Context) error {
		rows, err := store.Read(context.Background())
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, rows)
	})

	go func() {
		if err := e.Start(":1820"); err != nil {
			e.Logger.Info("Shutting down the server")
		}
	}()

	e.Logger.Info("Starting jobs")
	runner := deer.NewRunner(cfg, store)
	go runner.Start(context.Background())

	tea.SysCallWaitDefault()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	e.Logger.Info("Shutting down the runner")
	runner.Shutdown(ctx)
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

type configResp struct {
	Monitors []configMonitorResp `json:"monitors"`
}
type configMonitorResp struct {
	ID       string              `json:"id"`
	Name     string              `json:"name"`
	Services []configServiceResp `json:"services"`
}
type configServiceResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func buildConfigResp(cfg *deer.Config) *configResp {
	r := configResp{Monitors: make([]configMonitorResp, len(cfg.Monitors))}

	for mi, m := range cfg.Monitors {
		r.Monitors[mi] = configMonitorResp{
			ID:       m.ID,
			Name:     m.Name,
			Services: make([]configServiceResp, len(m.Services)),
		}
		for si, s := range m.Services {
			r.Monitors[mi].Services[si] = configServiceResp{
				ID:   s.ID,
				Name: s.Name,
			}
		}
	}

	return &r
}
