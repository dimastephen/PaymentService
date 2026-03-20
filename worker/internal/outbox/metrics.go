package outbox

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	eventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "outboxeventsProcessed",
		Help: "Events processed by outbox worker",
	}, []string{"type"})

	eventsDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "outboxeventsDuration",
		Help: "Duration of events processed by outbox worker",
	})
)
