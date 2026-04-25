FROM --platform=linux/amd64 golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o proxy .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/proxy .
EXPOSE 8080
CMD ["./proxy"]
