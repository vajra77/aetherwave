package demod

import (
	"aetherwave/internal/audio"
	"aetherwave/internal/dsp"
)

type AMDemodulator struct {
	// Possiamo aggiungere parametri di configurazione qui, come un volume software
	gain float32
}

// NewAM istanzia il demodulatore AM con un guadagno predefinito
func NewAM() *AMDemodulator {
	return &AMDemodulator{
		gain: 1.0,
	}
}

// Demodulate prende il Signal (I/Q ad alta frequenza) e restituisce l'audio reale
func (am *AMDemodulator) Process(in *dsp.Signal, out []audio.Sample) {
	out = out[:cap(out)]
	nSamples := in.Size()
	samples := in.Samples()

	// 1. Estrazione dell'inviluppo (Magnitude)
	var sum float32
	for i := 0; i < nSamples; i++ {
		// Il metodo Magnitude() usa internamente i complessi nativi!
		out[i] = audio.Sample(samples[i].Magnitude())
		sum += float32(out[i])
	}

	// 2. Rimozione della componente continua (DC Block)
	// Calcoliamo il valore medio dell'ampiezza in questo blocco
	averageDC := audio.Sample(sum / float32(nSamples))
	gainSample := audio.Sample(am.gain)

	for i := 0; i < nSamples; i++ {
		// Sottraiamo la media per centrare il segnale audio sullo 0.0
		out[i] = out[i].Sub(averageDC).Mul(gainSample)
		out[i] = out[i].Clamp(audio.Sample(-1.0), audio.Sample(1.0))
	}
}
