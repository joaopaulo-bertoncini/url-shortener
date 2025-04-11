FROM golang:1.23.4-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o url-shortener ./cmd/main.go

EXPOSE 8080

CMD ["./url-shortener"]
