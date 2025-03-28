# syntax=docker/dockerfile:1

FROM golang:1.24.1

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./cmd ./cmd

RUN CGO_ENABLED=0 GOOS=linux go build -o ./main ./cmd/peverel/

EXPOSE 8080

CMD ["./main", "--config", "./config.yml"]