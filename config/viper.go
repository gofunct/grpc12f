package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
	"os"
)

func SetupViper() {
	viper.SetConfigName("config")           // name of config file (without extension)
	viper.AddConfigPath(os.Getenv("$HOME")) // name of config file (without extension)
	viper.AddConfigPath(".")                // optionally look for config in the working directory
	viper.AutomaticEnv()                    // read in environment variables that match
	viper.SetDefault("tracing", true)
	viper.SetDefault("tls", false)
	viper.SetDefault("metrics_endpoint", true)
	viper.SetDefault("live_endpoint", false)
	viper.SetDefault("ready_endpoint", false)
	viper.SetDefault("pprof_endpoint", true)
	viper.SetDefault("db_host", "localhost")
	viper.SetDefault("db_port", ":5432")
	viper.SetDefault("db_name", "postgresdb")
	viper.SetDefault("db_user", "admin")
	viper.SetDefault("grpc_port", ":8443")
	viper.SetDefault("routine_threshold", 300)
	viper.SetDefault("jaeger_metrics", false)
	viper.SetDefault("monitor_peers", true)
	viper.SetDefault("swaggerfile", "swagger.json")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Println(zap.String("error", "failed to read config file, writing defaults..."))
		if err := viper.WriteConfigAs("config.yaml"); err != nil {
			log.Fatal("failed to write config")
			os.Exit(1)
		}

	} else {
		log.Println("Using config file:", zap.String("config", viper.ConfigFileUsed()))
		if err := viper.WriteConfig(); err != nil {
			log.Fatal("failed to write config file")
			os.Exit(1)
		}
	}

	if viper.GetBool("tls") == true {
		viper.Set("grpc_port", ":443")
		if err := viper.WriteConfig(); err != nil {
			log.Fatal("failed to rewrite config")
			os.Exit(1)
		}
	}
}
