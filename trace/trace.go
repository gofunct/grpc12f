package trace

import (
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	jconfig "github.com/uber/jaeger-client-go/config"
	"io"
)

func NewTracer(service string) (io.Closer, error) {
		var err error
		var tracer opentracing.Tracer
		cfg, err := jconfig.FromEnv()
		if err != nil {
			return nil, err
		}
		cfg.ServiceName = service
		cfg.RPCMetrics = viper.GetBool("jaeger_metrics")

		tracer, closer, err := cfg.NewTracer()
		if err != nil {
			return nil, err
		}
		opentracing.SetGlobalTracer(tracer)

		return closer, err
	}

}

