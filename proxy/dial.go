package proxy

import (
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"log"
	"time"
)

func NewDialOpts() []grpc.DialOption {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("failed to setup logger for grpc interceptor")
	}
	opts := []grpc_zap.Option{
		grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Int64("grpc.time_ns", duration.Nanoseconds())
		}),
	}
	grpc_zap.ReplaceGrpcLogger(logger)
	streamInterceptors := grpc.StreamClientInterceptor(grpc_middleware.ChainStreamClient(
		grpc_opentracing.StreamClientInterceptor(grpc_opentracing.WithTracer(opentracing.GlobalTracer())),
		grpc_prometheus.StreamClientInterceptor,
		grpc_zap.StreamClientInterceptor(logger, opts...),
	))

	unaryInterceptors := grpc.UnaryClientInterceptor(grpc_middleware.ChainUnaryClient(
		grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(opentracing.GlobalTracer())),
		grpc_prometheus.UnaryClientInterceptor,
		grpc_zap.UnaryClientInterceptor(logger, opts...),
	))

	dopts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(unaryInterceptors),
		grpc.WithStreamInterceptor(streamInterceptors),
	}
	return dopts
}
