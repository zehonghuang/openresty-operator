FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o reload-agent ./cmd

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/reload-agent /usr/local/bin/reload-agent

EXPOSE 8080

WORKDIR /app
ENTRYPOINT ["reload-agent"]
