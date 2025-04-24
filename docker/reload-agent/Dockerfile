FROM --platform=linux/amd64 golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o reload-agent ./cmd

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/reload-agent /usr/local/bin/reload-agent

EXPOSE 19091

WORKDIR /app
ENTRYPOINT ["reload-agent"]
