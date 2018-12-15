package proxy

import (
	"fmt"
	"github.com/gofunct/grpc12f/logging"
	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Proxy struct {
	Mux       *http.ServeMux
	Gateway   *runtime.ServeMux
	Formatter handlers.LogFormatter
	DialOpts  []grpc.DialOption
	Prefix 		string
}

func NewProxy(ctx context.Context) *Proxy {

	formatter := logging.LogHandlers()

	logrus.Infof("Creating grpc-gateway proxy")
	mux := NewMux()

	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, viper.GetString("proxy.swagger_file"))
	})

	gwmux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(incomingHeaderMatcher),
		runtime.WithOutgoingHeaderMatcher(outgoingHeaderMatcher),
	)
	logrus.Infof("Proxying requests to gRPC service at '%s'", viper.GetString("proxy.backend"))

	opts := NewDialOpts()
	prefix := sanitizeApiPrefix( viper.GetString("proxy.prefix"))


	/*
	// If you get a compilation error that gw.Register${SERVICE}HandlerFromEndpoint
	// does not exist, it's because you haven't added any google.api.http annotations
	// to your proto. Add some!
		err := gw.Register${SERVICE}HandlerFromEndpoint(ctx, gwmux, viper.GetString("proxy.backend"), opts)
		if err != nil {
			logrus.Fatalf("Could not register gateway: %v", err)
		}

		logrus.Infof("API prefix is: %s", prefix)
		mux.Handle(prefix, handlers.CustomLoggingHandler(os.Stdout, http.StripPrefix(prefix[:len(prefix)-1], allowCors(cfg, gwmux)), formatter))
	*/

	return &Proxy{
		Mux:       mux,
		Gateway:   gwmux,
		Formatter: formatter,
		DialOpts:  opts,
		Prefix:    prefix,
	}
}

// SignalRunner runs a runner function until an interrupt signal is received, at which point it
// will call stopper.
func SignalRunner(runner, stopper func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	go func() {
		runner()
	}()

	logrus.Info("hit Ctrl-C to shutdown")
	select {
	case <-signals:
		stopper()
	}
}

func Listen() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	prox := NewProxy(ctx)

	addr := fmt.Sprintf(":%v", viper.GetInt("proxy.port"))
	server := &http.Server{Addr: addr, Handler: prox.Mux}

	SignalRunner(
		func() {
			logrus.Infof("launching http server on %v", server.Addr)
			if err := server.ListenAndServe(); err != nil {
				logrus.Fatalf("Could not start http server: %v", err)
			}
		},
		func() {
			shutdown, _ := context.WithTimeout(ctx, 10*time.Second)
			server.Shutdown(shutdown)
		})
}

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	check := healthcheck.NewMetricsHandler(prometheus.DefaultRegisterer, "gateway")
	if viper.GetBool("proxy.live_endpoint") == true {
		check.AddLivenessCheck("goroutine_threshold", healthcheck.GoroutineCountCheck(viper.GetInt("routine_threshold")))
		mux.HandleFunc("/live", check.LiveEndpoint)
		log.Info("liveness handler registered-->", "/live")
	}

	if viper.GetBool("proxy.ready_endpoint") == true {
		check.AddReadinessCheck("db_health_check", healthcheck.TCPDialCheck(viper.GetString("db_port"), 1*time.Second))
		mux.HandleFunc("/ready", check.ReadyEndpoint)
		log.Info("readiness handler registered-->", "/ready")
	}
	if viper.GetBool("proxy.pprof_endpoint") == true {
		mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
		log.Info("pprof handler registered-->", "/debug/pprof")
	}

	if viper.GetBool("proxy.metrics_endpoint") == true {
		mux.Handle("/metrics", promhttp.Handler())
		log.Info("metrics handler registered-->", "/metrics")
	}
	return mux
}
