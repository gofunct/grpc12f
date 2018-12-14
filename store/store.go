package store

import (
	"github.com/go-pg/pg"
	"github.com/spf13/viper"
	"time"
)

func NewStore() *pg.DB {

		return pg.Connect(&pg.Options{
			User:                  viper.GetString("db_user"),
			Password:              viper.GetString("db_pass"),
			Database:              viper.GetString("db_name"),
			Addr:                  viper.GetString("db_host") + viper.GetString("db_port"),
			RetryStatementTimeout: true,
			MaxRetries:            4,
			MinRetryBackoff:       250 * time.Millisecond,
		})
}

