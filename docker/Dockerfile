FROM golang:1.24-alpine

WORKDIR /app

COPY ./go.mod /app
COPY ./go.sum /app
RUN go mod download

COPY ./ /app

RUN go build -o app ./cmd/api/main.go

EXPOSE 8082
CMD ["./app"]
