package logging

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func NewLogger() *zap.Logger {
		var err error
		log, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		log.With(
			zap.Bool("grpc_log", true),
			zap.String("grpc_port", viper.GetString("grpc_port")),
			zap.String("db_port", viper.GetString("db_port")),
			zap.String("db_name", viper.GetString("db_name")),
			zap.String("db_name", viper.GetString("db_user")),
		)

		zap.ReplaceGlobals(log)
		log.Debug("global logger successfully registered")
		return log
	}
}
