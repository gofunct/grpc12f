package runtime

import (
	"context"
	"github.com/go-pg/pg"
	"github.com/gofunct/grpc12f/store"
	"github.com/gofunct/grpc12f/trace"
	"github.com/gofunct/grpc12f/transport"
	"github.com/gofunct/grpc12f/config"
	"github.com/prometheus/common/log"
	"github.com/soheilhy/cmux"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
)

func init() { config.SetupViper() }

type Runtime struct {
	Server   *grpc.Server
	Router   *http.ServeMux
	Debug    *http.Server
	Store    *pg.DB
	Listener net.Listener
	LogLevel string
	// Whether to log request header
	Closer   io.Closer
}

func NewRuntime() (*Runtime, error) {
	var err error
	closer, err := trace.NewTracer("grpc_server")
	if err != nil {
		return nil, err
	}
	router := transport.NewMux()
	listener, err := transport.NewInsecureListener("grpc_port")
	return &Runtime{
		Server: transport.NewGrpc(),
		Router: router,
		Debug: &http.Server{
			Handler: router,
		},
		Store:    store.NewStore(),
		Listener: listener,
		Closer:   closer,
	}, err
}

func (r *Runtime) Serve(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	m := cmux.New(r.Listener)
	grpcListener := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpListener := m.Match(cmux.HTTP1Fast())
	group.Go(func() error { return r.Server.Serve(grpcListener) })
	log.Info("grpc server started successfully-->", viper.GetString("grpc_port"))
	group.Go(func() error { return r.Debug.Serve(httpListener) })
	log.Info("debug server started successfully-->", viper.GetString("grpc_port"))
	group.Go(func() error { return m.Serve() })


	return group.Wait()
}

func (r *Runtime) Deny(msg string, err error) {
	log.Fatal(msg, zap.Error(err))
}

func (r *Runtime) Shutdown(ctx context.Context) func() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	return func() {
		select {
		case <-signals:
			log.Fatal("signal received, shutting down...")
			r.Server.GracefulStop()
			r.Debug.Shutdown(ctx)
			r.Closer.Close()
		case <-ctx.Done():
			log.Fatal("context done, shutting down...")
			r.Server.GracefulStop()
			r.Debug.Shutdown(ctx)
			r.Closer.Close()
		}
	}
}
