package demod

import "aetherwave/internal/dsp" // Ora importa dsp invece di iq

type Demodulator interface {
	// Accetta il buffer definito dentro il pacchetto dsp
	Demodulate(in *dsp.SignalBuffer) []float32
}
