package config

import (
	"github.com/prometheus/common/log"
	"github.com/spf13/viper"
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
	viper.SetDefault("live_endpoint", true)
	viper.SetDefault("ready_endpoint", true)
	viper.SetDefault("pprof_endpoint", true)
	viper.SetDefault("db_host", "localhost")
	viper.SetDefault("db_port", ":5432")
	viper.SetDefault("db_name", "postgresdb")
	viper.SetDefault("db_user", "admin")
	viper.SetDefault("grpc_port", ":8443")
	viper.SetDefault("routine_threshold", 300)
	viper.SetDefault("jaeger_metrics", true)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Info("failed to read config file, writing defaults...")
		if err := viper.WriteConfigAs("config.yaml"); err != nil {
			log.Fatal("failed to write config")
			os.Exit(1)
		}

	} else {
		log.Info("Using config file-->", viper.ConfigFileUsed())
		if err := viper.WriteConfig(); err != nil {
			log.Fatal("failed to write config file")
			os.Exit(1)
		}
	}

	if viper.GetBool("tls") == true {
		viper.Set("gw_port", ":443")
		if err := viper.WriteConfig(); err != nil {
			log.Fatal("failed to rewrite config")
			os.Exit(1)
		}
	}
}
