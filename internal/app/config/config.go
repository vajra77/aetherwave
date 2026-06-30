package config

import "flag"

type Config struct {
	Frequency uint64
	Mode      string
}

func New() *Config {

	config := Config{}

	flag.StringVar(&config.Mode, "mode", "AM", "Radio mode (AM, FM, USB, LSB, CW)")
	flag.Uint64Var(&config.Frequency, "freq", 1000000000, "Radio frequency")
	flag.Parse()

	return &config
}
