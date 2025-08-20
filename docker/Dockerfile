FROM golang:1.24-alpine

WORKDIR /app

COPY ./ /app

RUN go mod download

RUN go build -o app ./cmd/api

EXPOSE 8082
CMD ["./app"]
