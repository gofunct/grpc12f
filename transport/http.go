package transport

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"net/http"
	"net/http/pprof"
	"time"
)

func NewGWMux() (*http.ServeMux, *runtime.ServeMux) {
	mux := http.NewServeMux()
	gw := runtime.NewServeMux()
	mux.Handle("/", gw)
	check := healthcheck.NewMetricsHandler(prometheus.DefaultRegisterer, "gateway")
	if viper.GetBool("live_endpoint") {
		check.AddLivenessCheck("goroutine_threshold", healthcheck.GoroutineCountCheck(viper.GetInt("routine_threshold")))
		mux.HandleFunc("/live", check.LiveEndpoint)
	}

	if viper.GetBool("ready_endpoint") {
		check.AddReadinessCheck("db_health_check", healthcheck.TCPDialCheck(viper.GetString("db_port"), 1*time.Second))
		mux.HandleFunc("/ready", check.ReadyEndpoint)
	}
	if viper.GetBool("pprof_endpoint") {
		mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	}

	if viper.GetBool("metrics_endpoint") {
		mux.Handle("/metrics", promhttp.Handler())
	}
	return mux, gw
}

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	check := healthcheck.NewMetricsHandler(prometheus.DefaultRegisterer, "runtime")
	if viper.GetBool("live_endpoint") {
		check.AddLivenessCheck("goroutine_threshold", healthcheck.GoroutineCountCheck(viper.GetInt("routine_threshold")))
		mux.HandleFunc("/live", check.LiveEndpoint)
	}

	if viper.GetBool("ready_endpoint") {
		check.AddReadinessCheck("grpc_listener_health_check", healthcheck.TCPDialCheck(viper.GetString("grpc_port"), 1*time.Second))
		check.AddReadinessCheck("db_health_check", healthcheck.TCPDialCheck(viper.GetString("db_port"), 1*time.Second))
		mux.HandleFunc("/ready", check.ReadyEndpoint)
	}
	if viper.GetBool("pprof_endpoint") {
		mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	}

	if viper.GetBool("metrics_endpoint") {
		mux.Handle("/metrics", promhttp.Handler())
	}
	return mux
}
