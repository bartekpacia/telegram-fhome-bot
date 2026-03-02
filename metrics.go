package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

const defaultMetricsListenAddr = "127.0.0.1:9108"

type metricsRecorder struct {
	toggleSuccessTotal      atomic.Uint64
	toggleErrorTotal        atomic.Uint64
	toggleSuccessDurationNs atomic.Uint64
	toggleErrorDurationNs   atomic.Uint64
	toggleSuccessLastUnix   atomic.Int64
	toggleErrorLastUnix     atomic.Int64
}

func newMetricsRecorder() *metricsRecorder {
	return &metricsRecorder{}
}

func (m *metricsRecorder) recordToggle(success bool, duration time.Duration) {
	durNs := uint64(duration.Nanoseconds())

	if success {
		m.toggleSuccessTotal.Add(1)
		m.toggleSuccessDurationNs.Add(durNs)
		m.toggleSuccessLastUnix.Store(time.Now().Unix())
		return
	}

	m.toggleErrorTotal.Add(1)
	m.toggleErrorDurationNs.Add(durNs)
	m.toggleErrorLastUnix.Store(time.Now().Unix())
}

func (m *metricsRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("method not allowed"))
		return
	}

	successTotal := m.toggleSuccessTotal.Load()
	errorTotal := m.toggleErrorTotal.Load()
	successDurationSecondsTotal := float64(m.toggleSuccessDurationNs.Load()) / float64(time.Second)
	errorDurationSecondsTotal := float64(m.toggleErrorDurationNs.Load()) / float64(time.Second)
	successLastUnix := m.toggleSuccessLastUnix.Load()
	errorLastUnix := m.toggleErrorLastUnix.Load()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	_, _ = fmt.Fprintln(w, "# HELP telegram_fhome_bot_gate_toggle_events_total Total number of requested gate toggles.")
	_, _ = fmt.Fprintln(w, "# TYPE telegram_fhome_bot_gate_toggle_events_total counter")
	_, _ = fmt.Fprintf(w, "telegram_fhome_bot_gate_toggle_events_total{result=\"success\"} %d\n", successTotal)
	_, _ = fmt.Fprintf(w, "telegram_fhome_bot_gate_toggle_events_total{result=\"error\"} %d\n", errorTotal)

	_, _ = fmt.Fprintln(w, "# HELP telegram_fhome_bot_gate_toggle_duration_seconds_total Total duration spent sending gate toggle events.")
	_, _ = fmt.Fprintln(w, "# TYPE telegram_fhome_bot_gate_toggle_duration_seconds_total counter")
	_, _ = fmt.Fprintf(w, "telegram_fhome_bot_gate_toggle_duration_seconds_total{result=\"success\"} %g\n", successDurationSecondsTotal)
	_, _ = fmt.Fprintf(w, "telegram_fhome_bot_gate_toggle_duration_seconds_total{result=\"error\"} %g\n", errorDurationSecondsTotal)

	_, _ = fmt.Fprintln(w, "# HELP telegram_fhome_bot_gate_toggle_duration_seconds_count Number of measured gate toggle durations.")
	_, _ = fmt.Fprintln(w, "# TYPE telegram_fhome_bot_gate_toggle_duration_seconds_count counter")
	_, _ = fmt.Fprintf(w, "telegram_fhome_bot_gate_toggle_duration_seconds_count{result=\"success\"} %d\n", successTotal)
	_, _ = fmt.Fprintf(w, "telegram_fhome_bot_gate_toggle_duration_seconds_count{result=\"error\"} %d\n", errorTotal)

	_, _ = fmt.Fprintln(w, "# HELP telegram_fhome_bot_gate_toggle_last_unix_time_seconds Unix timestamp of the last gate toggle event.")
	_, _ = fmt.Fprintln(w, "# TYPE telegram_fhome_bot_gate_toggle_last_unix_time_seconds gauge")
	_, _ = fmt.Fprintf(w, "telegram_fhome_bot_gate_toggle_last_unix_time_seconds{result=\"success\"} %d\n", successLastUnix)
	_, _ = fmt.Fprintf(w, "telegram_fhome_bot_gate_toggle_last_unix_time_seconds{result=\"error\"} %d\n", errorLastUnix)
}

func startMetricsServer(ctx context.Context, addr string, metrics *metricsRecorder) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics)

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
