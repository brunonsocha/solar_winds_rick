FROM golang:1.26-alpine AS builder
RUN apk add git ca-certificates
WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o service ./cmd

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/service .
EXPOSE 8080
CMD ["./service"]

