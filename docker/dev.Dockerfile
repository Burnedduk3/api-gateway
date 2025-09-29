FROM golang:1.25-alpine3.22 AS build

WORKDIR /build

COPY go.mod go.sum main.go .

RUN go mod download

COPY . .

RUN go build -o product-service .

FROM alpine:3.22 AS app

WORKDIR /app

COPY --from=build /build/product-service /app

COPY configs/product-config.yaml /etc/product-service/config.yaml

ENTRYPOINT ["/app/product-service", "server"]
