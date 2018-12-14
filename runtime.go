package grpc12factor

import (
	"context"
	"github.com/go-pg/pg"
	"github.com/gofunct/grpc12factor/config"
	"github.com/gofunct/grpc12factor/store"
	"github.com/gofunct/grpc12factor/trace"
	"github.com/gofunct/grpc12factor/transport"
	"github.com/soheilhy/cmux"
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
	Log      *zap.Logger
	Server   *grpc.Server
	Router   *http.ServeMux
	Debug    *http.Server
	Store    *pg.DB
	Listener net.Listener
	Closer   io.Closer
}

func NewRuntime() (*Runtime, error) {
	var err error

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	closer, err := trace.NewTracer("grpc_server")
	if err != nil {
		return nil, err
	}

	router := transport.NewMux()

	listener, err := transport.NewInsecureListener("grpc_port")

	return &Runtime{
		Log:    logger,
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
