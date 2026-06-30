package demod

import (
	"aetherwave/internal/audio"
	"aetherwave/internal/dsp"
) // Ora importa dsp invece di iq

type Demodulator interface {
	Process(in *dsp.Signal) []audio.Sample
}
