package runtime

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"net"
	"net/http"
)

func Serve() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		lis, err := net.Listen("tcp", VString("grpc_port"))
		if err != nil {
			log.Fatal("Failed to listen:"+VString("grpc_port"), err)
		}
		tracer, closer, err := Trace(log.JZap)
		if err != nil {
			log.Fatal("Cannot initialize Jaeger Tracer %s", zap.Error(err))
		}
		defer closer.Close()

		// Set GRPC Interceptors
		server := NewServer(tracer)

		//api.RegisterTodoServiceServer(server, &db.Store{DB: NewDB()})

		mux := NewMux()
		log.Zap.Debug("Starting debug service..", zap.String("grpc_debug_port", VString("grpc_debug_port")))
		go func() { http.ListenAndServe(VString("grpc_debug_port"), mux) }()

		log.Zap.Debug("Starting grpc service..", zap.String("grpc_port", VString("grpc_port")))
		server.Serve(lis)
	}
}
