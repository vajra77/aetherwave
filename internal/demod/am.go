package demod

import "aetherwave/internal/dsp"

type AMDemodulator struct{}

func NewAM() *AMDemodulator { return &AMDemodulator{} }

func (am *AMDemodulator) Demodulate(in *dsp.SignalBuffer) []float32 {
	// Sfrutta i metodi esposti da dsp.SignalBuffer
	audioOut := make([]float32, in.Size())
	return audioOut
}
