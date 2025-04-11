FROM golang:1.23-alpine

RUN apk update && apk add --no-cache git curl build-base

ENV PATH="/go/bin:$PATH"

RUN go install github.com/air-verse/air@latest

WORKDIR /app

COPY . .

EXPOSE 8080

CMD ["air"]
