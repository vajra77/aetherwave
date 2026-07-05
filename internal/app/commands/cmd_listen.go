package commands

import (
	"aetherwave/internal/app/config"
	"aetherwave/internal/audio"
	"aetherwave/internal/dsp"
	"aetherwave/internal/radio"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type listenCmd struct {
	config *config.Config
}

func (c *listenCmd) Name() string        { return "listen" }
func (c *listenCmd) Description() string { return "Listen on a given frequancy and modulation mode" }
func (c *listenCmd) Usage() string {
	return "  -freq <frequency>\n" +
		" -mode <mode>\n"
}

func (c *listenCmd) Init(config *config.Config) {
	c.config = config
}

func (c *listenCmd) Run() {

	fmt.Println("🎶 Radio is starting, please wait...")
	fmt.Printf("frequency: %d kHz\n", c.config.Frequency)
	// 1. Configura la pipeline astratta
	rcvr := radio.NewReceiver(
		dsp.Frequency(c.config.Frequency*1000),
		dsp.SampleRate(250000),
		c.config.Mode,
	)

	player, err := audio.NewPlayer(48000, rcvr.AudioChannel())
	if err != nil {
		fmt.Printf("❌ Audio Error: %v\n", err)
		return
	}
	defer player.Close()

	// 4. Avviamo i motori!
	if err := rcvr.Start(); err != nil {
		fmt.Printf("❌ SDR HW Error: %v\n", err)
		return
	}
	defer rcvr.Stop()

	if err := player.Start(); err != nil {
		fmt.Printf("❌ Audio Start Error: %v\n", err)
		return
	}

	fmt.Println("🎶 Radio is playing, press CTRL+C to stop.")

	// 5. Mettiamo il main thread in attesa del segnale di chiusura del sistema (CTRL+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nRadio shutting off, thank you for listening! 👋")
}
