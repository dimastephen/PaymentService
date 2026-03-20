package handler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	eventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "projection_events_proccesed",
		Help: "Total events proceesed in projection worker",
	}, []string{"status"})
	eventsDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "projection_events_duration",
		Help: "Duration of events processed",
	})
)
