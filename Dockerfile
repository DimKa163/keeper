FROM golang:latest AS builder

LABEL authors="Games"

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app ./cmd/server

FROM alpine:latest AS run

COPY --from=builder /app /app

WORKDIR /app
EXPOSE 8080

CMD ["./server"]