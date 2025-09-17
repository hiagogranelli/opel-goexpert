FROM golang:latest AS builder
WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o service-a .

FROM golang:1.19-alpine
COPY --from=builder /app/service-a /app/service-a

CMD ["/app/service-a"]
