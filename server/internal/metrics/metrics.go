// Package metrics defines the Prometheus metrics exposed on /metrics.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tictactoe_active_connections",
		Help: "Number of currently connected WebSocket clients.",
	})

	ConnectionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tictactoe_connections_total",
		Help: "Total number of WebSocket connections accepted.",
	})

	ReconnectsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tictactoe_reconnects_total",
		Help: "Total number of connections that resumed an existing participant identity.",
	})

	DisconnectsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tictactoe_disconnects_total",
		Help: "Total number of WebSocket connections lost (clean close, error, or stale-heartbeat sweep).",
	})

	ActiveGames = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tictactoe_active_games",
		Help: "Number of games currently in progress.",
	})

	GamesStartedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tictactoe_games_started_total",
		Help: "Total number of games started.",
	})

	GamesCompletedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tictactoe_games_completed_total",
		Help: "Total number of games completed, labeled by outcome.",
	}, []string{"outcome"}) // outcome: "win" or "draw"

	MovesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tictactoe_moves_total",
		Help: "Total number of moves processed.",
	})

	DatabaseLatencySeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tictactoe_database_latency_seconds",
		Help:    "Latency of PostgreSQL operations.",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})

	WSMessageLatencySeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "tictactoe_ws_message_latency_seconds",
		Help:    "Time to process an inbound WebSocket message, from receipt to dispatch completion.",
		Buckets: prometheus.DefBuckets,
	})
)

// Handler exposes the Prometheus scrape endpoint.
func Handler() http.Handler {
	return promhttp.Handler()
}
