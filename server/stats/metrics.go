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
