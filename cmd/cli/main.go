package main

import (
	"aetherwave/internal/app/commands"
	"aetherwave/internal/app/config"
	"log"
	"os"
)

func main() {

	cmdStr := os.Args[1]
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	var cmd commands.Command

	for _, c := range commands.All() {
		if c.Name() == cmdStr {
			cmd = c
			break
		}
	}

	if cmd == nil {
		log.Printf("Command not found: %s", cmdStr)
		os.Exit(1)
	}

	cmd.Init(cfg)
	cmd.Run()
}
