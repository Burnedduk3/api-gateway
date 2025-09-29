FROM golang:1.25-alpine3.22 AS build

WORKDIR /build

COPY go.mod go.sum main.go .

RUN go mod download

COPY . .

RUN go build -o api-gateway .

FROM alpine:3.22 AS app

WORKDIR /app

COPY --from=build /build/api-gateway /app

COPY configs/config-docker.yaml /etc/api-gateway/config.yaml

ENTRYPOINT ["/app/api-gateway", "server"]
