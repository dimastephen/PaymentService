package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/payment-service/shared/kafka"
	"github.com/payment-service/shared/psp"
	"github.com/payment-service/shared/storage"
	"github.com/payment-service/shared/tracing"
	"github.com/payment-service/worker/internal/config"
	"github.com/payment-service/worker/internal/handler"
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

	tp, err := tracing.InitTracer(ctx, "worker", cfg.JaegerURL())
	if err != nil {
		slog.Error("tracer init err", "err", err)
	}
	defer tp.Shutdown(ctx)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	pool, err := storage.NewPool(ctx, cfg.PostgresDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	slog.Info("connected to postgres database")

	eventRepo := storage.NewEventRepository(pool)
	commandRepo := storage.NewCommandRepository(pool)
	paymentRepo := storage.NewPaymentRepository(pool)
	store := storage.NewStorage(pool)

	syncProducer, err := kafka.NewSyncProducer(cfg.Brokers())
	if err != nil {
		log.Fatal(err)
	}
	defer syncProducer.Close()

	pspMock := psp.NewFakeProvider()
	worker := handler.NewWorkerHandler(pspMock, eventRepo, commandRepo, paymentRepo, syncProducer, store, cfg.EventTopic())
	consumer := kafka.NewConsumer(cfg.Brokers(), cfg.Group(), cfg.CommandTopic(), worker.Handle)
	defer consumer.Close()
	slog.Info("consumer started")

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9090", mux)
	}()

	err = consumer.Start(ctx)
	if err != nil {
		slog.Error("Error starting consumer", "error", err)
		os.Exit(1)
	}
}
