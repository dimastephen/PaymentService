package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/payment-service/api/internal/config"
	"github.com/payment-service/api/internal/handler"
	"github.com/payment-service/shared/kafka"
	"github.com/payment-service/shared/storage"
	"github.com/payment-service/shared/tracing"
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
	tp, err := tracing.InitTracer(ctx, "api-service", cfg.JaegerURL())
	if err != nil {
		slog.Error("failed to init tracer", "err", err)
	}
	defer tp.Shutdown(ctx)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	pool, err := storage.NewPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	paymentRepo := storage.NewPaymentRepository(pool)

	producer, err := kafka.NewSyncProducer(cfg.Brokers())
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	apiHandler := handler.NewHandler(cfg.CommandTopic(), producer, paymentRepo)

	server := http.Server{
		Addr:         cfg.Address(),
		Handler:      apiHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("starting server at " + cfg.Address())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Error starting server", "error", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	slog.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)
}
