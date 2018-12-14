package transport

import (
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"time"
)

func NewGrpc() *grpc.Server {
		grpc_zap.ReplaceGrpcLogger(zap.L())
		zopts := []grpc_zap.Option{
			grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
				return zap.Int64("grpc.time_ns", duration.Nanoseconds())
			}),
		}
		// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
		grpc_zap.ReplaceGrpcLogger(zap.L())
		s := grpc.NewServer(
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(opentracing.GlobalTracer())),
				grpc_zap.StreamServerInterceptor(zap.L(), zopts...),
			)),
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(opentracing.GlobalTracer())),
				grpc_zap.UnaryServerInterceptor(zap.L(), zopts...),
			)),
		)

		grpc_health_v1.RegisterHealthServer(s, health.NewServer())
		return s

	}
