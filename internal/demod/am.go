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

// Demodulate prende il SignalBuffer (I/Q ad alta frequenza) e restituisce l'audio reale
func (am *AMDemodulator) Process(in *dsp.SignalBuffer) []audio.Sample {
	nSamples := in.Size()
	samples := in.Samples()

	// Allocchiamo l'array per i campioni audio in uscita (stessa lunghezza temporanea)
	audioOut := make([]audio.Sample, nSamples)

	// 1. Estrazione dell'inviluppo (Magnitude)
	var sum float32
	for i := 0; i < nSamples; i++ {
		// Il metodo Magnitude() usa internamente i complessi nativi!
		audioOut[i] = audio.Sample(samples[i].Magnitude())
		sum += float32(audioOut[i])
	}

	// 2. Rimozione della componente continua (DC Block)
	// Calcoliamo il valore medio dell'ampiezza in questo blocco
	averageDC := audio.Sample(sum / float32(nSamples))
	gainSample := audio.Sample(am.gain)

	for i := 0; i < nSamples; i++ {
		// Sottraiamo la media per centrare il segnale audio sullo 0.0
		audioOut[i] = audioOut[i].Sub(averageDC).Mul(gainSample)
		audioOut[i] = audioOut[i].Clamp(audio.Sample(-1.0), audio.Sample(1.0))
	}

	return audioOut
}
