package commands

import (
	"aetherwave/internal/app/config"
	"aetherwave/internal/audio"
	"aetherwave/internal/dsp"
	"aetherwave/internal/radio"
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

	// 1. Configura la pipeline astratta
	rcvr := radio.NewReceiver(
		dsp.Frequency(c.config.Frequency),
		dsp.SampleRate(48000),
		c.config.Mode,
	)

	// 2. Avvia il flusso in background
	rcvr.Start()
	defer rcvr.Stop()

	// 3. La CLI è interessata solo a suonare l'audio, quindi svuota il canale audio nel player
	player, _ := audio.New()

	for audioData := range rcvr.AudioChannel() {
		player.Play(audioData)
	}
}
