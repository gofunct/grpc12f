package grpc12factor

import (
	"context"
	"fmt"
	"github.com/go-pg/pg"
	"github.com/soheilhy/cmux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

type Runtime struct {
	Log     *zap.Logger
	RootCmd *cobra.Command
	Server  *grpc.Server
	Store   *pg.DB
	Router  *http.ServeMux
	Closer  io.Closer
}

func NewRuntime() *Runtime {
	return &Runtime{}
}

func Compose(r *Runtime) *Runtime {

	if r.Log == nil {
		o := WithLogger()
		r = o(r)
	}
	if r.RootCmd == nil {
		o := WithRootCmd()
		r = o(r)
	}
	if r.Server == nil {
		o := WithServer()
		r = o(r)
	}
	if r.Store == nil {
		o := WithStore()
		r = o(r)
	}
	if r.Router == nil {
		o := WithRouter()
		r = o(r)
	}
	if r.Closer == nil {
		r.Log.Debug("Failed to compose runtime closer, using default...")
		o := WithTracer()
		r = o(r)
	}

	return r
}

func init() {
	viper.SetConfigName("config")           // name of config file (without extension)
	viper.AddConfigPath(os.Getenv("$HOME")) // name of config file (without extension)
	viper.AddConfigPath(".")                // optionally look for config in the working directory
	viper.AutomaticEnv()                    // read in environment variables that match
	viper.SetDefault("tracing", true)
	viper.SetDefault("tls", false)
	viper.SetDefault("metrics_endpoint", true)
	viper.SetDefault("live_endpoint", false)
	viper.SetDefault("ready_endpoint", false)
	viper.SetDefault("pprof_endpoint", true)
	viper.SetDefault("db_host", "localhost")
	viper.SetDefault("db_port", ":5432")
	viper.SetDefault("db_name", "postgresdb")
	viper.SetDefault("db_user", "admin")
	viper.SetDefault("grpc_port", ":8443")
	viper.SetDefault("routine_threshold", 300)
	viper.SetDefault("jaeger_metrics", false)
	viper.SetDefault("monitor_peers", true)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Println(zap.String("error", "failed to read config file, writing defaults..."))
		if err := viper.WriteConfigAs("config.yaml"); err != nil {
			log.Fatal("failed to write config")
			os.Exit(1)
		}

	} else {
		log.Println("Using config file:", zap.String("config", viper.ConfigFileUsed()))
		if err := viper.WriteConfig(); err != nil {
			log.Fatal("failed to write config file")
			os.Exit(1)
		}
	}

	if viper.GetBool("tls") == true {
		viper.Set("grpc_port", ":443")
		if err := viper.WriteConfig(); err != nil {
			log.Fatal("failed to rewrite config")
			os.Exit(1)
		}
	}

}

func (r *Runtime) Execute() {
	if err := r.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (r *Runtime) Serve(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	server := &http.Server{
		Handler: r.Router,
	}
	listener, err := net.Listen("tcp", viper.GetString("grpc_port"))
	if err != nil {
		log.Fatal(err)
	}
	if viper.GetString("grpc_port") == ":443" {
		var x = viper.GetStringSlice("domains")

		if len(x) < 1 {
			r.Log.Debug("failed to create tls certificates, must add domains to config.yaml before enabling tls")
		} else {
			r.Log.Debug("creating tls certificates and registering listener...")
			listener = autocert.NewListener(viper.GetStringSlice("domains")...)
		}
	}

	m := cmux.New(listener)
	grpcListener := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpListener := m.Match(cmux.HTTP1Fast())

	r.Log.Debug("Starting grpc service..", zap.String("grpc_port", viper.GetString("grpc_port")))
	group.Go(func() error { return r.Server.Serve(grpcListener) })

	r.Log.Debug("Starting debug service..", zap.String("grpc_port", viper.GetString("grpc_port")))
	group.Go(func() error { return server.Serve(httpListener) })

	group.Go(func() error { return m.Serve() })

	return group.Wait()
}

func (r *Runtime) Deny(msg string, err error) {
	r.Log.Fatal(msg, zap.Error(err))
}
