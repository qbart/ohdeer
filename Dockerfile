FROM golang:1.15.5-alpine3.12 AS build

WORKDIR /src
ADD . /src
RUN go mod download
RUN CGO_ENABLED=0 go build -o /bin/ohdeer

FROM scratch
COPY --from=build /bin/ohdeer /bin/ohdeer
ENTRYPOINT ["/bin/ohdeer"]

