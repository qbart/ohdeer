package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/franela/goblin"
	"github.com/qbart/ohdeer/deer"
	"github.com/qbart/ohdeer/deerstore"
)

func TestRunner(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Runner", func() {
		g.It("Performs healthchecks based on config and saves metrics", func() {
			g.Timeout(10 * time.Second)

			c, err := deer.ParseConfig("http.hcl", []byte(`
			monitor "test" {
				name = "Test"

				service "api" {
					name = "API"

					http {
						interval = 1
						timeout  = 2
						addr     = "https://doesnotexist.ohdeer.dev"

						expect "status" {
							in = [408]
						}
					}
				}
			}
			`))

			if err != nil {
				t.Errorf("Error when parsing %v", err)
				return
			}

			store, err := deerstore.NewTimescaleDB(context.Background(), os.Getenv("DATABASE_URL"))
			if err != nil {
				t.Errorf("Error openning database %v", err)
				return
			}
			defer store.Close(context.Background())

			err = store.Migrate(context.Background())
			if err != nil {
				t.Errorf("Error migrating database %v", err)
				return
			}

			err = store.Truncate(context.Background())
			if err != nil {
				t.Errorf("Error truncating database %v", err)
				return
			}

			// start runner and keep it running for about 5 secs
			{
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				runner := deer.NewRunner(c, store)
				runner.Start(ctx)

			OUTER:
				for {
					select {
					case <-ctx.Done():
						break OUTER
					}
				}
				runner.Shutdown(context.Background())

			}

			metrics, err := store.Read(context.Background(), deer.ReadFilter{
				TimeBucket:     1,
				TimeBucketUnit: "hour",
				Interval:       23,
				IntervalUnit:   "hour",
			})
			if err != nil {
				t.Errorf("Error fetching data %v", err)
				return
			}

			if len(metrics) != 24 {
				t.Errorf("Not enough metrics %d (!=24)", len(metrics))
				return
			}

			for i := 0; i < len(metrics); i++ {
				m := metrics[i]
				g.Assert(m.MonitorID).Eql("test")
				g.Assert(m.ServiceID).Eql("api")
				if i == len(metrics)-1 {
					g.Assert(m.Health).Eql(float64(0))
				} else {
					g.Assert(m.Health).Eql(float64(-1))
				}
			}

			fmt.Println("hello")
			for i := len(metrics) - 1; i >= 1; i-- {
				prev := metrics[i-1]
				curr := metrics[i]

				d := curr.Bucket.Sub(prev.Bucket)
				g.Assert(d.Hours()).Eql(1.0)
			}
		})
	})
}
