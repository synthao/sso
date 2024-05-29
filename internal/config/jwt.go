package config

import "go.uber.org/config"

type JWT struct {
	Secret string `yaml:"secret"`
}

func NewJWTConfig() (*JWT, error) {
	provider, err := config.NewYAML(config.File(filename))
	if err != nil {
		return nil, err
	}

	var c JWT

	err = provider.Get("jwt").Populate(&c)
	if err != nil {
		panic(err)
	}

	return &c, nil
}
