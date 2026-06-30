package demod

import (
	"aetherwave/internal/audio"
	"aetherwave/internal/dsp"
	"math"
)

type CWDemodulator struct {
	sampleRate dsp.SampleRate
	phase      float64 // Mantiene lo stato della fase per evitare clic audio tra i blocchi
	toneFreq   float64 // La frequenza del "bip" (es. 700 Hz)
}

// NewCW istanzia il demodulatore impostando la nota acustica desiderata
func NewCW(sampleRate dsp.SampleRate, toneHz float64) *CWDemodulator {
	return &CWDemodulator{
		sampleRate: sampleRate,
		toneFreq:   toneHz,
		phase:      0.0,
	}
}

func (cw *CWDemodulator) Demodulate(in *dsp.Signal) []audio.Sample {
	nSamples := in.Size()
	samples := in.Samples()
	audioOut := make([]audio.Sample, nSamples)

	// Calcoliamo di quanto deve avanzare la fase della nostra nota a ogni campione
	phaseStep := 2.0 * math.Pi * cw.toneFreq / float64(cw.sampleRate)

	for i := 0; i < nSamples; i++ {
		// 1. Generiamo il vettore dell'oscillatore locale alla frequenza del tono (es. 700 Hz)
		bfo := dsp.Sample(complex(float32(math.Cos(cw.phase)), float32(math.Sin(cw.phase))))

		// 2. Moltiplicazione complessa: eterodiniamo il segnale radio con il nostro tono
		mixedSample := samples[i].Multiply(bfo)

		// 3. Estraiamo la componente reale (I) che ora oscilla a 700 Hz se la portante era presente
		audioOut[i] = audio.Sample(mixedSample.I())

		// 4. Incrementiamo la fase mantenendola nel range [0, 2*PI]
		cw.phase += phaseStep
		if cw.phase > 2.0*math.Pi {
			cw.phase -= 2.0 * math.Pi
		}
	}

	return audioOut
}
