FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /data-dashboard cmd/data-dashboard/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /data-dashboard /app/data-dashboard
COPY --from=builder /app/data-dashboard/templates /app/templates
COPY --from=builder /app/data-dashboard/config /app/config

ENV CONFIG_DIR=/app/config
ENV TEMPLATES_DIR=/app/templates
ENV PORT=8084

EXPOSE 8084
ENTRYPOINT ["/app/data-dashboard"]
