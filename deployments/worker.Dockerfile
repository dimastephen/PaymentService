FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN cd worker && go build -o /app/worker-bin ./cmd/main.go

FROM alpine AS runner
WORKDIR /app
COPY --from=builder /app/worker-bin ./worker
CMD ["./worker"]