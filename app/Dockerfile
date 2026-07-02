FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /orders-service ./cmd/server

FROM alpine:3.20
RUN addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app
COPY --from=builder /orders-service /app/orders-service
EXPOSE 8080
ENTRYPOINT ["/app/orders-service"]
