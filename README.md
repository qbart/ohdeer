# Oh! Deer!

[![LICENSE](https://img.shields.io/github/license/qbart/ohdeer)](https://github.com/qbart/ohdeer/blob/master/LICENSE)
[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/qbart/ohdeer)](https://goreportcard.com/report/github.com/qbart/ohdeer)
[![Last commit](https://img.shields.io/github/last-commit/qbart/ohdeer)](https://github.com/qbart/ohdeer/commits/master)

## Example config

```hcl
monitor "aws:eu-west-1" {
  name = "AWS Europe"

  service "api" {
    name = "API"

    http {
      addr     = "https://ohdeer.dev"
      interval = 5
      timeout  = 10

      expect "status" {
        in = [200]
      }
    }
  }
}
```
