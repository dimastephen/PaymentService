FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN cd projection && go build -o /app/projection-bin ./cmd/main.go

FROM alpine AS runner
WORKDIR /app
COPY --from=builder /app/projection-bin ./projection
CMD ["./projection"]