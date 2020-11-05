package deer

import (
	"testing"

	. "github.com/franela/goblin"
)

func TestParseConfig(t *testing.T) {
	g := Goblin(t)
	g.Describe(".ParseConfig", func() {
		g.Describe("Valid config", func() {
			c, err := ParseConfig("http.hcl", []byte(`
			monitor "aws:eu-west-1" {
				name = "AWS Europe"

				service "core-api" {
					name = "Core API"

					http {
						interval = 100
						timeout  = 10
						addr     = "http://a.local"

						expect "status" {
							in = [200]
						}
					}
				}
			}
			`))

			if err != nil {
				t.Errorf("Error when parsing %v", err)
			}

			g.It("Parses service definition correctly", func() {
				g.Assert(len(c.Monitors)).Equal(1)
				g.Assert(len(c.Monitors[0].Services)).Equal(1)
				g.Assert(len(c.Monitors[0].Services[0].HttpChecks)).Equal(1)

				g.Assert(c.Monitors[0].ID).Equal("aws:eu-west-1")
				g.Assert(c.Monitors[0].Name).Equal("AWS Europe")

				g.Assert(c.Monitors[0].Services[0].ID).Equal("core-api")
				g.Assert(c.Monitors[0].Services[0].Name).Equal("Core API")

				http := c.Monitors[0].Services[0].HttpChecks[0]
				g.Assert(http.Addr).Equal("http://a.local")
				g.Assert(http.TimeoutSec).Equal(10)
				g.Assert(http.IntervalSec).Equal(100)

			})

			g.Describe("Inclusion", func() {
				g.It("Parses http check", func() {
					expect := c.Monitors[0].Services[0].HttpChecks[0].Expectations[0]

					g.Assert(expect.Subject).Equal("status")
					g.Assert(expect.Inclusion).Equal([]int{200})
				})
			})
		})

		g.Describe("Invalid timeout", func() {
			g.It("Fails", func() {
				_, err := ParseConfig("http.hcl", []byte(`
				monitor "a" {
					name = "a"
					service "b" {
						name = "b"
						http {
							interval = 10
							timeout  = 0
							addr     = "http://a.local"

							expect "status" {
								in = [200]
							}
						}
					}
				}
				`))

				g.Assert(err.Error()).Equal("Timeout must be > 0")
			})
		})

		g.Describe("Invalid interval", func() {
			g.It("Fails", func() {
				_, err := ParseConfig("http.hcl", []byte(`
				monitor "a" {
					name = "a"
					service "b" {
						name = "b"
						http {
							interval = 0
							timeout  = 10
							addr     = "http://a.local"

							expect "status" {
								in = [200]
							}
						}
					}
				}
				`))

				g.Assert(err.Error()).Equal("Interval must be > 0")
			})
		})

		g.Describe("Invalid address", func() {
			g.It("Fails", func() {
				_, err := ParseConfig("http.hcl", []byte(`
				monitor "a" {
					name = "a"
					service "b" {
						name = "b"
						http {
							interval = 10
							timeout  = 10
							addr     = ""

							expect "status" {
								in = [200]
							}
						}
					}
				}
				`))

				g.Assert(err.Error()).Equal("Addr cannot be empty")
			})
		})

		g.Describe("Missing monitor ID", func() {
			g.It("Fails", func() {
				_, err := ParseConfig("http.hcl", []byte(`
				monitor "" {
				  name = "a"
				}
				`))

				g.Assert(err.Error()).Equal("Monitor cannot have empty ID")
			})
		})

		g.Describe("Missing monitor name", func() {
			g.It("Fails", func() {
				_, err := ParseConfig("http.hcl", []byte(`
				monitor "a" {
				  name = ""
				}
				`))

				g.Assert(err.Error()).Equal("Monitor cannot have empty name")
			})
		})

		g.Describe("Missing service ID", func() {
			g.It("Fails", func() {
				_, err := ParseConfig("http.hcl", []byte(`
				monitor "a" {
				  name = "a"
				  service "" {
	     			name = "b"
				  }
				}
				`))

				g.Assert(err.Error()).Equal("Service in monitor a cannot have empty ID")
			})
		})

		g.Describe("Missing service name", func() {
			g.It("Fails", func() {
				_, err := ParseConfig("http.hcl", []byte(`
				monitor "a" {
				  name = "a"
				  service "b" {
	     			name = ""
				  }
				}
				`))

				g.Assert(err.Error()).Equal("Service in monitor a cannot have empty name")
			})
		})
	})
}
