package gateway

import (
	"context"
	"github.com/go-pg/pg"
	"github.com/gofunct/grpc12factor/config"
	"github.com/gofunct/grpc12factor/store"
	"github.com/gofunct/grpc12factor/trace"
	"github.com/gofunct/grpc12factor/transport"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"net/http"
)

func init() { config.SetupViper() }

type Runtime struct {
	Log      *zap.Logger
	Server 	*http.Server
	Store    *pg.DB
	Router   *http.ServeMux
	Gate     *runtime.ServeMux
	DialOpts []grpc.DialOption
	Closer   io.Closer
}

func NewRuntime() (*Runtime, error) {
	var err error
	r := &Runtime{}
	r.Log, err = zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	r.DialOpts = transport.NewDialOpts(r.Log)

	r.Router, r.Gate = transport.NewGWMux()

	r.Server = &http.Server{
		Handler: r.Router,
	}
	r.Closer, err = trace.NewTracer("grpc_gateway")
	if err != nil {
		return nil, err
	}
	r.Store = store.NewStore()

	return r, err
}

func (r *Runtime) NewConnection(ctx context.Context) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, viper.GetString("grpc_port"), r.DialOpts...)
}

func (r *Runtime) NewServer(ctx context.Context) *http.Server {
	return &http.Server{
		Handler: r.Router,
	}
}