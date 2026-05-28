FROM golang:1.25-alpine AS dev
WORKDIR /app
RUN go install github.com/air-verse/air@latest
ENV PATH="/root/go/bin:${PATH}"
COPY go.mod go.sum ./
RUN go mod download
CMD ["air"]

FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o capital-pipefy ./cmd/api

FROM alpine:3.20 AS prod
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/capital-pipefy .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8282
CMD ["./capital-pipefy"]
