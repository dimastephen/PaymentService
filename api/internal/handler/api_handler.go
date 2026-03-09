package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/payment-service/shared/domain"
	"github.com/payment-service/shared/kafka"
	pb "github.com/payment-service/shared/proto/gen/proto/payment"
	"github.com/payment-service/shared/storage"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/proto"
)

var tracer = otel.Tracer("api-service")

type Handler struct {
	mux          *chi.Mux
	producer     kafka.Producer
	repo         storage.PaymentRepository
	commandTopic string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func NewHandler(commandTopic string, producer kafka.Producer, repo storage.PaymentRepository) *Handler {
	mux := chi.NewMux()
	mux.Use(MetricsMiddleware)
	handler := &Handler{
		commandTopic: commandTopic,
		producer:     producer,
		repo:         repo,
		mux:          mux,
	}
	mux.Get("/payments/{key}", handler.GetPayment)
	mux.Post("/payments", handler.PostPayment)
	mux.Handle("/metrics", promhttp.Handler())
	return handler
}

func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetPayment")
	defer span.End()
	slog.Info("get payment", "idempotency", chi.URLParam(r, "key"))
	key := chi.URLParam(r, "key")
	if key == "" {
		slog.Error("get payment: no key provided")
		span.RecordError(errors.New("no key provided"))
		span.SetStatus(codes.Error, "no key provided")
		writeError(w, http.StatusBadRequest, errors.New("no key provided"))
		return
	}
	resp, err := h.repo.Get(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		writeError(w, http.StatusInternalServerError, err)
		slog.Error("failed to write response", "idempotency", resp.IdempotencyKey, "err", err)
		return
	}

}
func (h *Handler) PostPayment(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "PostPayment")
	defer span.End()
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	cmd := &domain.CreatePayment{}
	err := decoder.Decode(cmd)
	slog.Info("post payment", "Idempotency_key", cmd.IdempotencyKey)
	if err != nil {
		slog.Error("failed to parse cmd", "cmd", cmd)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		writeError(w, http.StatusBadRequest, err)
		return
	}
	_, err = domain.NewPayment(cmd)
	if err != nil {
		slog.Error("failed to validate cmd", "cmd", cmd)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		writeError(w, http.StatusBadRequest, err)
		return
	}
	command := &pb.CreatePaymentCommand{
		IdempotencyKey: cmd.IdempotencyKey,
		Amount:         cmd.Amount,
		Currency:       string(cmd.Currency),
		MerchantId:     cmd.MerchantID,
	}
	req, err := proto.Marshal(command)
	if err != nil {
		slog.Error("failed to marshal command", "command", command)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	err = h.producer.SendMessage(ctx, h.commandTopic, []byte(cmd.IdempotencyKey), req)
	if err != nil {
		slog.Error("failed to send command", "command", cmd)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	slog.Info("Successfully sent command", "command", cmd)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	return
}

func writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
