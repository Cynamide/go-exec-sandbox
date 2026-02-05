FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/evaluator ./cmd/evaluator

FROM alpine:latest

RUN apk --no-cache add ca-certificates docker-cli

WORKDIR /app

COPY --from=builder /app/evaluator .

EXPOSE 8080

ENTRYPOINT ["./evaluator"]
