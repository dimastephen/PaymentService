FROM golang:1.25-alpine
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.24.1
COPY migrations/ /migrations/
CMD goose -dir /migrations postgres "$POSTGRES_DSN" up