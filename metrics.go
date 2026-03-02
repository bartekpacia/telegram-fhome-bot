package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const defaultMetricsListenAddr = "127.0.0.1:9108"

const (
	metricsResultLabel = "result"
	metricsResultError = "error"
	metricsResultOK    = "success"
)

type metricsRecorder struct {
	toggleEventsTotal         *prometheus.CounterVec
	toggleDurationSeconds     *prometheus.HistogramVec
	toggleLastUnixTimeSeconds *prometheus.GaugeVec
}

func newMetricsRecorder() *metricsRecorder {
	m := &metricsRecorder{
		toggleEventsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "telegram_fhome_bot_gate_toggle_events_total",
				Help: "Total number of requested gate toggles.",
			},
			[]string{metricsResultLabel},
		),
		toggleDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "telegram_fhome_bot_gate_toggle_duration_seconds",
				Help:    "Duration of sending gate toggle events to F&Home.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{metricsResultLabel},
		),
		toggleLastUnixTimeSeconds: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "telegram_fhome_bot_gate_toggle_last_unix_time_seconds",
				Help: "Unix timestamp of the last gate toggle event.",
			},
			[]string{metricsResultLabel},
		),
	}

	prometheus.MustRegister(
		m.toggleEventsTotal,
		m.toggleDurationSeconds,
		m.toggleLastUnixTimeSeconds,
	)

	return m
}

func (m *metricsRecorder) recordToggle(success bool, duration time.Duration) {
	result := metricsResultError
	if success {
		result = metricsResultOK
	}

	m.toggleEventsTotal.WithLabelValues(result).Inc()
	m.toggleDurationSeconds.WithLabelValues(result).Observe(duration.Seconds())
	m.toggleLastUnixTimeSeconds.WithLabelValues(result).Set(float64(time.Now().Unix()))
}

func startMetricsServer(ctx context.Context, addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil && err != http.ErrServerClosed {
			slog.Warn("error while shutting down metrics server", slog.Any("error", err))
		}
	}()

	go func() {
		slog.Info("metrics server will listen", slog.String("addr", addr), slog.String("endpoint", "/metrics"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("metrics server failed", slog.Any("error", err))
		}
	}()
}
