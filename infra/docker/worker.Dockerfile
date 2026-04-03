FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY internal ./internal
COPY services ./services
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/worker ./services/worker/cmd/worker

FROM alpine:3.21
WORKDIR /srv/app
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/worker /usr/local/bin/worker
COPY services/api/migrations ./services/api/migrations
ENTRYPOINT ["/usr/local/bin/worker"]

