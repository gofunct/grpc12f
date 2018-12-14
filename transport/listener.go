package transport

import (
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
	"net"
	"errors"
)

func NewSecureListener() (net.Listener, error) {
	switch {
	case len(viper.GetStringSlice("domains")) == 0:
		return nil, errors.New("failed to create secure listener- a list of domains must be provided in your config file")
	case viper.GetBool("tls") != true:
		return nil, errors.New("failed to create secure listener- must set tls to true in config file")
	}
	return autocert.NewListener(viper.GetString("domains")), nil
}

func NewInsecureListener(key string) (net.Listener, error) {
	listener, err := net.Listen("tcp", viper.GetString(key))
	if err != nil {
		return nil, err
	}
	return listener, err
}
