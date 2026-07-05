FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /identity-hub cmd/identity-hub/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /identity-hub /identity-hub
EXPOSE 8080
ENTRYPOINT ["/identity-hub"]
