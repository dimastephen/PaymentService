FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN cd api && go build -o /app/api-bin ./cmd/main.go

FROM alpine AS runner
WORKDIR /app
COPY --from=builder /app/api-bin ./api
CMD ["./api"]
