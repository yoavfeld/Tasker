package lib

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	Port string
}

func LoadConf(path string) (*Config, error) {
	v := viper.New()
	v.AddConfigPath(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "v.ReadInConfig failed")
	}
	c := new(Config)
	if err := v.Unmarshal(c); err != nil {
		return nil, errors.Wrap(err, "v.Unmarshal failed")
	}
	return c, nil
}
