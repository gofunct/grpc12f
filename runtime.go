package grpc12factor

import (
	"context"
	"fmt"
	"github.com/go-pg/pg"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
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
	"os/signal"
)

func init() { SetupViper() }

type Runtime struct {
	Log      *zap.Logger
	RootCmd  *cobra.Command
	Server   *grpc.Server
	Debug    *http.Server
	Store    *pg.DB
	Router   *http.ServeMux
	Gate     *runtime.ServeMux
	DialOpts []grpc.DialOption
	Closer   io.Closer
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
	if r.DialOpts == nil {
		o := WithDialer()
		r = o(r)
	}
	if r.Closer == nil {
		o := WithTracer()
		r = o(r)
	}

	return r
}

func (r *Runtime) Execute() {
	if err := r.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (r *Runtime) Serve(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

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
	group.Go(func() error { return r.Debug.Serve(httpListener) })

	group.Go(func() error { return m.Serve() })

	return group.Wait()
}

func (r *Runtime) Deny(msg string, err error) {
	r.Log.Fatal(msg, zap.Error(err))
}

func (r *Runtime) Shutdown(ctx context.Context) func() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	return func() {
		select {
		case <-signals:
			r.Log.Debug("signal received, shutting down...")
			r.Server.GracefulStop()
			r.Debug.Shutdown(ctx)
			r.Closer.Close()
		case <-ctx.Done():
			r.Log.Debug("context done, shutting down...")
			r.Server.GracefulStop()
			r.Debug.Shutdown(ctx)
			r.Closer.Close()
		}
	}
}
