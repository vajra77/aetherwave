package config

import (
	"flag"
	"os"
)

type Config struct {
	Frequency uint64
	Mode      string
}

func New() (*Config, error) {
	config := Config{}

	fs := flag.NewFlagSet("config", flag.ExitOnError)
	fs.StringVar(&config.Mode, "mode", "AM", "Radio mode (AM, FM, USB, LSB, CW)")
	fs.Uint64Var(&config.Frequency, "freq", 1000000, "Radio frequency")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return nil, err
	}

	return &config, nil
}
