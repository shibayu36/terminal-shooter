package stats

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// 接続中のクライアント数
var ActiveClients = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "terminal_shooter_active_clients",
	Help: "The number of active clients",
})

// Publishされたパケット数
var PublishedPackets = promauto.NewCounter(prometheus.CounterOpts{
	Name: "terminal_shooter_published_packets_total",
	Help: "The total number of published packets",
})

// ゲーム更新ループの実行時間
var GameLoopDuration = promauto.NewHistogram(prometheus.HistogramOpts{
	Name: "terminal_shooter_game_loop_duration_seconds",
	Help: "Time spent processing game loop",
	Buckets: []float64{
		0.0001, 0.001, 0.005, 0.01, 0.015,
		0.016667,
		0.02, 0.025, 0.03,
	},
})
