# Oh! Deer!

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
