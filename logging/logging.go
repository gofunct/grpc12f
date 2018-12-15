package logging

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net"
	"github.com/gorilla/handlers"
	"time"
	"io"
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


func LogHandlers() handlers.LogFormatter {

	// Setup logrus
	logrus.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "@timestamp",
		},
	})
	level, err := logrus.ParseLevel(viper.GetString("proxy.log_level"))
	if err != nil {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(level)
	}

	return func(writer io.Writer, params handlers.LogFormatterParams) {

		host, _, err := net.SplitHostPort(params.Request.RemoteAddr)
		if err != nil {
			host = params.Request.RemoteAddr
		}

		uri := params.Request.RequestURI

		// Requests using the CONNECT method over HTTP/2.0 must use
		// the authority field (aka r.Host) to identify the target.
		// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
		if params.Request.ProtoMajor == 2 && params.Request.Method == "CONNECT" {
			uri = params.Request.Host
		}
		if uri == "" {
			uri = params.URL.RequestURI()
		}

		duration := int64(time.Now().Sub(params.TimeStamp) / time.Millisecond)

		fields := logrus.Fields{
			"host":       host,
			"url":        uri,
			"duration":   duration,
			"status":     params.StatusCode,
			"method":     params.Request.Method,
			"request":    params.Request.RequestURI,
			"remote":     params.Request.RemoteAddr,
			"size":       params.Size,
			"referer":    params.Request.Referer(),
			"user_agent": params.Request.UserAgent(),
			"request_id": params.Request.Header.Get("x-request-id"),
		}

			if headers, err := json.Marshal(params.Request.Header); err == nil {
				fields["headers"] = string(headers)
			} else {
				fields["header_error"] = err.Error()
			}

		logrus.WithFields(fields).WithTime(params.TimeStamp).Infof("%s %s %d", params.Request.Method, uri, params.StatusCode)
	}
}

