# Build stage (ELEGANT)
FROM golang:1.21.0-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o jwtserver ./cmd/api/main.go

# Final stage (ELEGANT)
FROM alpine:latest

WORKDIR /go/src/app

COPY --from=builder /build/ .

CMD ["./jwtserver"]

