package config

import (
	"flag"
	"os"
)

type Config struct {
	Frequency uint64
	Profile   string
}

func New() (*Config, error) {
	config := Config{}

	fs := flag.NewFlagSet("config", flag.ExitOnError)
	fs.StringVar(&config.Profile, "profile", "AMW", "Radio mode (AMW, AMN, USB, LSB, CW, WFM, NFM)")
	fs.Uint64Var(&config.Frequency, "freq", 1950, "Radio frequency in kHz")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return nil, err
	}

	return &config, nil
}
