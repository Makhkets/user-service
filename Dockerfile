FROM golang:1.20-alpine

RUN apk add build-base

EXPOSE 8000

WORKDIR /app
COPY . /app

# RUN go mod download && go mod verify && go build -o ./cmd/api/server.go
RUN go mod download
RUN go mod verify
RUN go build -o app ./cmd/api/server.go

CMD ["./app"]