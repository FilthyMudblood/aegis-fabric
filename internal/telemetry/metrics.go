package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// PeerCVPScore tracks the localized trust score of each neighbor.
	PeerCVPScore = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "afp_peer_cvp_score",
			Help: "Current Coordination Viability Probability (CVP) score evaluated locally",
		},
		[]string{"peer_id"},
	)

	// IngressActionTotal categorizes the physical enforcements executed by the Single Execution Authority.
	IngressActionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "afp_ingress_actions_total",
			Help: "Total number of physical actions enforced on ingress traffic",
		},
		[]string{"action"}, // e.g., fast_path, slow_path, drop_isolated, drop_probation, stranger_tax_reject
	)

	// InjectedDelayDuration measures the structural dampening latency via Slow Path.
	InjectedDelayDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "afp_injected_delay_milliseconds",
			Help:    "Physical latency injected by the Slow Path",
			Buckets: []float64{100, 500, 1000, 1500, 2000, 3000},
		},
		[]string{"peer_id"},
	)

	// CVPPenaltyEventsTotal tracks observed CVP downgrades during ingress evaluations.
	CVPPenaltyEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "afp_cvp_penalty_events_total",
			Help: "Total number of CVP downgrade events observed by the sidecar",
		},
		[]string{"peer_id"},
	)

	// CVPPenaltyAmount tracks the magnitude of each observed CVP downgrade.
	CVPPenaltyAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "afp_cvp_penalty_amount",
			Help:    "Magnitude of CVP downgrade per observed event",
			Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.4, 0.8},
		},
		[]string{"peer_id"},
	)
)
