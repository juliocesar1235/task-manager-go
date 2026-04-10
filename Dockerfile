FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:latest

# Non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/server .

USER appuser

EXPOSE 8080
CMD ["./server"]