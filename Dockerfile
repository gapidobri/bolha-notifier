FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bolha-notifier

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/bolha-notifier /app

ENTRYPOINT ["/app/bolha-notifier"]