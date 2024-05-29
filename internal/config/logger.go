package config

import "go.uber.org/config"

type Logger struct {
	Level string `yaml:"level"`
}

func NewLoggerConfig() (*Logger, error) {
	provider, err := config.NewYAML(config.File(filename))
	if err != nil {
		return nil, err
	}

	var c Logger

	err = provider.Get("logger").Populate(&c)
	if err != nil {
		panic(err)
	}

	return &c, nil
}
