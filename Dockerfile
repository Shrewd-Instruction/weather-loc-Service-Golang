FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /weather-loc-service .
FROM golang:1.24-alpine

WORKDIR /root/

COPY --from=builder /weather-loc-service .

EXPOSE 8080

CMD ["./weather-loc-service"]
