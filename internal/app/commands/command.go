package commands

import (
	"aetherwave/internal/app/config"
)

type Command interface {
	Name() string
	Description() string
	Usage() string
	Init(config *config.Config)
	Run()
}

func All() []Command {
	return []Command{
		&listenCmd{},
		&scanCmd{},
	}
}
