package main

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/qbart/ohdeer/deer"
	"github.com/qbart/ohdeer/deerstatic"
	"github.com/qbart/ohdeer/deerstore"
	"github.com/qbart/ohtea/tea"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	e := echo.New()
	e.Use(middleware.Recover())
	// e.Use(middleware.Logger())
	e.Use(middleware.Secure())
	e.Logger.SetLevel(log.INFO)

	e.Logger.Info("Loading config")
	cfg, err := deer.LoadConfig("./ohdeer.hcl")
	if err != nil {
		e.Logger.Fatal(err)
	}
	if cfg.Tls.Domain != "" {
		if cfg.Tls.CacheDir != "" {
			e.AutoTLSManager.Cache = autocert.DirCache(cfg.Tls.CacheDir)
		}
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(cfg.Tls.Domain)
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

	e.Renderer = &Template{
		templates: template.Must(template.New("index").Parse(deerstatic.IndexTpl)),
	}
	e.Logger.Info("Starting server")
	e.GET("/", func(c echo.Context) error {
		data, err := store.Read(context.Background(), deer.ReadFilter{
			TimeBucket:     1,
			TimeBucketUnit: "hour",
			Interval:       23,
			IntervalUnit:   "hour",
		})
		if err != nil {
			e.Logger.Error(err)
			return c.String(http.StatusInternalServerError, "Failed to fetch metrics")
		} else {
			view := buildIndexView(cfg, data)
			err := c.Render(http.StatusOK, "index", view)
			if err != nil {
				e.Logger.Error(err)
				return c.String(http.StatusInternalServerError, "Failed to render view")
			}
			return nil
		}
	})
	e.GET("/api/v1/config", func(c echo.Context) error {
		return c.JSON(http.StatusOK, buildConfigResp(cfg))
	})
	e.GET("/api/v1/metrics", func(c echo.Context) error {
		rows, err := store.Read(context.Background(), deer.ReadFilter{
			TimeBucket:     1,
			TimeBucketUnit: "hour",
			Interval:       23,
			IntervalUnit:   "hour",
		})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, rows)
	})

	go func() {
		var err error
		if cfg.Tls.Domain != "" {
			err = e.StartAutoTLS(":443")
		} else {
			err = e.Start(":1820")
		}

		if err != nil {
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

func buildIndexView(cfg *deer.Config, data []*deer.Metric) *IndexView {
	monitorNames := map[string]string{}
	serviceNames := map[string]string{}

	for _, m := range cfg.Monitors {
		monitorNames[m.ID] = m.Name
		for _, s := range m.Services {
			serviceNames[s.ID] = s.Name
		}
	}

	view := IndexView{
		Monitors: make([]*IndexViewMonitor, 0),
	}
	pm := ""
	ps := ""
	var monitor *IndexViewMonitor
	var service *IndexViewService

	for _, m := range data {
		if pm != m.MonitorID {
			pm = m.MonitorID
			ps = "" // reset servicej
			monitor = &IndexViewMonitor{
				Name:     monitorNames[m.MonitorID],
				Services: make([]*IndexViewService, 0),
			}
			view.Monitors = append(view.Monitors, monitor)
		}

		if ps != m.ServiceID {
			ps = m.ServiceID
			service = &IndexViewService{
				Name:   serviceNames[m.ServiceID],
				Health: make([]IndexViewHealth, 0),
			}
			monitor.Services = append(monitor.Services, service)
		}
		service.Health = append(service.Health, IndexViewHealth{
			Health: m.Health,
			When:   m.Bucket,
		})
	}
	return &view
}

type IndexView struct {
	Monitors []*IndexViewMonitor
}

type IndexViewMonitor struct {
	Name     string
	Services []*IndexViewService
}

type IndexViewService struct {
	Name   string
	Health []IndexViewHealth
}

type IndexViewHealth struct {
	Health float64
	When   time.Time
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
