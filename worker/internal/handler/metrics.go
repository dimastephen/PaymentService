package handler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	commandsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_commands_processed",
		Help: "total commands processed",
	}, []string{"status"})
	commandDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "worker_commands_duration_seconds",
		Help:    "duration of commands send to worker",
		Buckets: prometheus.DefBuckets,
	})
)
