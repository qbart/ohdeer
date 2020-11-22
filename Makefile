test:
	DATABASE_URL=postgres://ohdeer:secret@localhost:5433/deer_test?sslmode=disable \
	go test -timeout 60s ./... 

build:
	mkdir -p bin/
	CGO_ENABLED=0 GOOS=linux go build -o bin/ohdeer

run:
	DATABASE_URL=postgres://ohdeer:secret@localhost:5432/deer?sslmode=disable go run main.go

dc:
	docker-compose up
