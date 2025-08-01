// Copyright (c) 2025 Ergo Authors
// released under the MIT license

package irc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics contains all Prometheus metrics for the IRC server
type Metrics struct {
	// Connection metrics
	ConnectionsTotal    prometheus.Counter
	ConnectionsActive   prometheus.Gauge
	ConnectionDuration  prometheus.Histogram
	
	// Message metrics
	MessagesReceived    *prometheus.CounterVec
	MessagesSent        *prometheus.CounterVec
	
	// Channel metrics
	ChannelsActive      prometheus.Gauge
	ChannelUsers        prometheus.Gauge
	
	// User metrics
	UsersOnline         prometheus.Gauge
	UsersRegistered     prometheus.Gauge
	AuthenticationTotal *prometheus.CounterVec
	
	// Server performance metrics
	CommandDuration     *prometheus.HistogramVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		ConnectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "ergo_connections_total",
			Help: "Total number of connections",
		}),
		ConnectionsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ergo_connections_active",
			Help: "Number of active connections",
		}),
		ConnectionDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "ergo_connection_duration_seconds",
			Help:    "Connection duration in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		MessagesReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "ergo_messages_received_total",
			Help: "Total number of messages received",
		}, []string{"command"}),
		MessagesSent: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "ergo_messages_sent_total",
			Help: "Total number of messages sent",
		}, []string{"command"}),
		ChannelsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ergo_channels_active",
			Help: "Number of active channels",
		}),
		ChannelUsers: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ergo_channel_users_total",
			Help: "Total number of users in all channels",
		}),
		UsersOnline: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ergo_users_online",
			Help: "Number of online users",
		}),
		UsersRegistered: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ergo_users_registered",
			Help: "Number of registered users",
		}),
		AuthenticationTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "ergo_authentication_total",
			Help: "Total number of authentication attempts",
		}, []string{"method", "result"}),
		CommandDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "ergo_command_duration_seconds",
			Help:    "Command execution duration in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"command"}),
	}

	// Register all metrics
	prometheus.MustRegister(
		m.ConnectionsTotal,
		m.ConnectionsActive,
		m.ConnectionDuration,
		m.MessagesReceived,
		m.MessagesSent,
		m.ChannelsActive,
		m.ChannelUsers,
		m.UsersOnline,
		m.UsersRegistered,
		m.AuthenticationTotal,
		m.CommandDuration,
	)

	return m
}

// StartMetricsServer starts the Prometheus metrics HTTP server
func (server *Server) StartMetricsServer() error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	
	metricsServer := &http.Server{
		Addr:         ":6060",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	
	server.logger.Info("server", fmt.Sprintf("Starting Prometheus metrics server on %s", metricsServer.Addr))
	
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			server.logger.Error("server", "Metrics server error", err.Error())
		}
	}()
	
	return nil
}

// UpdateMetrics updates Prometheus metrics with current server state
func (server *Server) UpdateMetrics() {
	if server.metrics == nil {
		return
	}
	
	// Update connection metrics
	clients := server.clients.AllClients()
	server.metrics.ConnectionsActive.Set(float64(len(clients)))
	
	// Update channel metrics
	channels := server.channels.Channels()
	channelCount := len(channels)
	totalChannelUsers := 0
	for _, channel := range channels {
		memberCount, _, _ := channel.listData()
		totalChannelUsers += memberCount
	}
	server.metrics.ChannelsActive.Set(float64(channelCount))
	server.metrics.ChannelUsers.Set(float64(totalChannelUsers))
	
	// Update user metrics
	server.metrics.UsersOnline.Set(float64(len(clients)))
}