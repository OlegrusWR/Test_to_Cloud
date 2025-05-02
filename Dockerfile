FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/balancer ./cmd/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/balancer .
COPY config.yaml .

EXPOSE 8080  
EXPOSE 8888  

VOLUME /app/data

CMD ["./balancer"]