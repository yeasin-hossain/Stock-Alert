package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	LoginURL   string `yaml:"login_url"`
	SignalRURL string `yaml:"signalr_url"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
