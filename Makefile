test:
	go test ./...

build:
	mkdir -p bin/
	go build -o bin/ohdeer

run:
	DATABASE_URL=postgres://ohdeer:secret@localhost:5432/deer?sslmode=disable go run main.go
