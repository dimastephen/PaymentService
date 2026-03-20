package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/payment-service/projection/internal/config"
	"github.com/payment-service/projection/internal/handler"
	"github.com/payment-service/shared/kafka"
	"github.com/payment-service/shared/storage"
	"github.com/payment-service/shared/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	tp, err := tracing.InitTracer(ctx, "projection", cfg.JaegerURL())
	if err != nil {
		slog.Error("Failed to init tracer", "err", err)
	}
	defer tp.Shutdown(ctx)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	pool, err := storage.NewPool(ctx, cfg.PostgresDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	paymentRepo := storage.NewPaymentRepository(pool)
	projection := handler.NewProjectionHandler(paymentRepo)
	consumer := kafka.NewConsumer(cfg.Brokers(), cfg.Group(), cfg.EventTopic(), projection.Handle)
	defer consumer.Close()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(":9091", mux)
		if err != nil {
			log.Fatal(err)
		}
	}()

	err = consumer.Start(ctx)
	if err != nil {
		slog.Error("Failed to start consumer", "err", err)
		os.Exit(1)
	}
}
