package config

import (
	"fmt"
	"go.uber.org/config"
	"os"
)

type DB struct {
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
	User     string `yaml:"user"`
	Host     string `yaml:"host"`
	Charset  string `yaml:"charset"`
	Params   string `yaml:"params"`
}

func NewDBConfig() (*DB, error) {
	provider, err := config.NewYAML(config.File(filename))
	if err != nil {
		return nil, err
	}

	var c DB

	err = provider.Get("db").Populate(&c)
	if err != nil {
		panic(err)
	}

	return &c, nil
}

func (c *DB) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
}
