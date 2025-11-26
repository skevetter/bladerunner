FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bladerunner ./cmd/bladerunner

FROM alpine:latest

WORKDIR /app

COPY --from=builder /bladerunner /app/bladerunner
# Copy migrations if we have any static files, though we might embed them later
# COPY --from=builder /app/migrations /app/migrations

EXPOSE 8090

CMD ["/app/bladerunner", "serve", "--http=0.0.0.0:8090"]
