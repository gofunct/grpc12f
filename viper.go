package runtime

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
)

var logV = LogViper()

func VString(key string) string {
	s := viper.GetString(key)
	return s
}

func Viperize() func(cmd *cobra.Command, args []string) error {

	return func(cmd *cobra.Command, args []string) error {
		viper.SetConfigName("gotasks")          // name of config file (without extension)
		viper.AddConfigPath(os.Getenv("$HOME")) // name of config file (without extension)
		viper.AddConfigPath(".")                // optionally look for config in the working directory
		viper.AutomaticEnv()                    // read in environment variables that match
		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err != nil {
			logV.Debug(`failed to locate config file. place a "gotasks.yaml" config file in your current or home directory`)
			return err
		} else {
			logV.Debug("Using config file:", zap.String("config", viper.ConfigFileUsed()))
		}
		return nil
	}

}
