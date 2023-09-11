FROM golang:1.20-alpine

RUN apk add build-base

EXPOSE 8000

WORKDIR /app
COPY . /app

# RUN go mod download && go mod verify && go build -o ./cmd/api/server.go
RUN go mod download
RUN go mod verify
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

RUN chmod +x ./cmd/api/server.go
RUN chmod +x ./app

RUN go build -o jwtserver ./cmd/api/server.go

CMD ["./jwtserver"]
