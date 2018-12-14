package transport

import (
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"time"
)

func NewDialOpts(log *zap.Logger) []grpc.DialOption {

	opts := []grpc_zap.Option{
		grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Int64("grpc.time_ns", duration.Nanoseconds())
		}),
	}
	streamInterceptors := grpc.StreamClientInterceptor(grpc_middleware.ChainStreamClient(
		grpc_zap.StreamClientInterceptor(log, opts...),
		grpc_opentracing.StreamClientInterceptor(grpc_opentracing.WithTracer(opentracing.GlobalTracer())),
	))

	unaryInterceptors := grpc.UnaryClientInterceptor(grpc_middleware.ChainUnaryClient(
		grpc_zap.UnaryClientInterceptor(log, opts...),
		grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(opentracing.GlobalTracer())),
	))

	dopts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(unaryInterceptors),
		grpc.WithStreamInterceptor(streamInterceptors),
	}
	return dopts
}
