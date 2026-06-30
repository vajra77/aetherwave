package commands

import (
	"aetherwave/internal/app/config"
)

type scanCmd struct {
	config *config.Config
}

func (c *scanCmd) Name() string        { return "scan" }
func (c *scanCmd) Description() string { return "Scan for available frequencies" }
func (c *scanCmd) Usage() string       { return "" }

func (c *scanCmd) Init(config *config.Config) {
	c.config = config
}

func (c *scanCmd) Run() {

}
