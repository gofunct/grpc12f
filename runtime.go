package runtime

import (
	"fmt"
	"github.com/go-pg/pg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

type Runtime struct {
	Log     *zap.Logger
	Metrics *MetricsIntercept
	RootCmd *cobra.Command
	Server  *grpc.Server
	Store   *pg.DB
	Router  *http.ServeMux
	Closer  io.Closer
}

func Compose(opts ...Option) *Runtime {
	o := &Runtime{}
	for _, opt := range opts {
		o = opt(o)
	}
	return o
}

func init() {
	viper.SetConfigName("config")           // name of config file (without extension)
	viper.AddConfigPath(os.Getenv("$HOME")) // name of config file (without extension)
	viper.AddConfigPath(".")                // optionally look for config in the working directory
	viper.AutomaticEnv()                    // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(zap.String("error", "failed to read config file."))
	} else {
		log.Println("Using config file:", zap.String("config", viper.ConfigFileUsed()))
	}
}

func (r *Runtime) Execute() {
	if err := r.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (r *Runtime) Serve() {
	lis, err := net.Listen("tcp", viper.GetString("grpc_port"))
	if err != nil {
		r.Deny("Failed to listen:"+viper.GetString("grpc_port"), err)
	}
	defer r.Kill()
	r.Log.Debug("Starting debug service..", zap.String("grpc_debug_port", viper.GetString("grpc_debug_port")))
	go func() { http.ListenAndServe(viper.GetString("grpc_debug_port"), r.Router) }()

	r.Log.Debug("Starting grpc service..", zap.String("grpc_port", viper.GetString("grpc_port")))
	r.Server.Serve(lis)
}

func (r *Runtime) Deny(msg string, err error) {
	r.Log.Fatal(msg, zap.Error(err))
}
func (r *Runtime) Kill() {
	r.Closer.Close()
}
