package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/qbart/ohdeer/deer"
	"github.com/qbart/ohdeer/deerstatic"
	"github.com/qbart/ohdeer/deerstore"
	"github.com/qbart/ohtea/tea"
)

func main() {
	configPath := flag.String("C", "./ohdeer.hcl", "config file path")
	flag.Parse()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.Secure())
	e.Logger.SetLevel(log.INFO)

	err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	})
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer sentry.Flush(2 * time.Second)

	e.Use(sentryecho.New(sentryecho.Options{}))

	e.Logger.Info("Loading config")
	cfg, err := deer.LoadConfig(*configPath)
	if err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Info("Connecting to store")
	store, err := deerstore.NewTimescaleDB(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer store.Close(context.Background())
	e.Logger.Info(context.Background(), "Migrate store")
	if err := store.Migrate(context.Background()); err != nil {
		e.Logger.Fatal(err)
	}

	e.Renderer = &myTemplate{
		templates: template.Must(template.New("index").Parse(deerstatic.IndexTpl)),
	}
	e.Logger.Info("Starting server")
	e.GET("/", func(c echo.Context) error {
		if err = c.Render(http.StatusOK, "index", cfg); err != nil {
			e.Logger.Error(err)
			return err
		}
		return nil
	})
	e.GET("/api/v1/config", func(c echo.Context) error {
		return c.JSON(http.StatusOK, buildConfigResp(cfg))
	})
	e.GET("/api/v1/metrics/default/:monitor/:service", func(c echo.Context) error {
		active := activeFilter(c.Param("monitor"), c.Param("service"))
		since := time.Now().Add(-time.Duration(89) * 24 * time.Hour)

		metrics, err := store.Read(context.Background(), &deer.ReadFilter{
			Since:          since,
			TimeBucket:     1,
			TimeBucketUnit: "day",
			Interval:       89,
			IntervalUnit:   "day",
			ActiveServices: active,
		})

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, defaultMetrics{
			Metrics: metrics,
			Uptime:  calcUptimeString(metrics),
		})
	})
	e.GET("/api/v1/metrics/details/:monitor/:service", func(c echo.Context) error {
		active := activeFilter(c.Param("monitor"), c.Param("service"))
		since, err := time.Parse(time.RFC3339, c.QueryParam("since"))

		if err != nil {
			return c.String(http.StatusUnprocessableEntity, err.Error())
		}
		metrics, err := store.Read(context.Background(), &deer.ReadFilter{
			Since:          since,
			TimeBucket:     1,
			TimeBucketUnit: "hour",
			Interval:       1,
			IntervalUnit:   "day",
			ActiveServices: active,
		})

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, defaultMetrics{
			Metrics: metrics,
			Uptime:  calcUptimeString(metrics),
		})
	})

	go func() {
		var err error
		if cfg.IsTLSConfigured() {
			err = e.StartTLS(cfg.Server.BindAddress, cfg.Server.TLSCertFile, cfg.Server.TLSKeyFile)
		} else {
			err = e.Start(cfg.Server.BindAddress)
		}

		if err != nil {
			if err.Error() != "http: Server closed" {
				e.Logger.Error(err)
			}
		}
	}()

	e.Logger.Info("Starting jobs")
	runner := deer.NewRunner(cfg, store)
	go runner.Start(context.Background())

	loop := tea.NewLoop()
	loop.OnShutdown(func(ctx context.Context) {

		e.Logger.Info("Shutting down the runner")
		runner.Shutdown(ctx)
		e.Logger.Info("Shutting down the server")
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}

	})
	loop.Run()
}

type defaultMetrics struct {
	Uptime  string         `json:"uptime"`
	Metrics []*deer.Metric `json:"metrics"`
}

func activeFilter(monitor, service string) map[string][]string {
	active := map[string][]string{}
	active[monitor] = []string{service}
	return active
}

func calcUptimeString(metrics []*deer.Metric) string {
	sum := 0.0
	count := 0.0
	uptime := "no data"
	for _, m := range metrics {
		sum += float64(m.PassedChecks)
		count += float64(m.PassedChecks + m.FailedChecks)
	}
	if count > 0.0 {
		upt := sum * 100.0 / count
		uptime = fmt.Sprintf("%0.2f%%", upt)
	}
	return uptime
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

type myTemplate struct {
	templates *template.Template
}

func (t *myTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
