package dsp

import (
	"math"
)

type FIRFilter struct {
	taps    []float32
	history []Sample
}

func NewLowPassFilter(numTaps int, sampleRate SampleRate, cutoffFreqHz Frequency) *FIRFilter {
	// Usa le funzioni matematiche interne allo stesso pacchetto dsp per calcolare i taps
	normCutoff := float64(cutoffFreqHz) / float64(sampleRate)
	taps := ComputeLowPassTaps(numTaps, normCutoff)

	return &FIRFilter{
		taps:    taps,
		history: make([]Sample, numTaps),
	}
}

// ComputeLowPassTaps calcola i coefficienti (taps) per un filtro passa-basso FIR.
// cutoffFreq è la frequenza di taglio normalizzata rispetto alla frequenza di campionamento (f_taglio / f_s).
func ComputeLowPassTaps(numTaps int, cutoffFreq float64) []float32 {
	taps := make([]float32, numTaps)
	M := numTaps - 1

	for n := 0; n < numTaps; n++ {
		// Se siamo al centro del filtro, evitiamo la divisione per zero nella Sinc
		if float64(n) == float64(M)/2.0 {
			taps[n] = float32(2.0 * cutoffFreq)
		} else {
			// Funzione Sinc standard
			arg := 2.0 * math.Pi * cutoffFreq * (float64(n) - float64(M)/2.0)
			sinc := math.Sin(arg) / arg

			// Applichiamo una finestra di Hamming per smussare il filtro
			hamming := 0.54 - 0.46*math.Cos(2.0*math.Pi*float64(n)/float64(M))

			taps[n] = float32(sinc * hamming)
		}
	}
	return taps
}

func (f *FIRFilter) Process(input *SignalBuffer) *SignalBuffer {
	numSamples := input.Size()
	numTaps := len(f.taps)

	// Creiamo il buffer di uscita con gli stessi metadati di frequenza e sample rate
	output := NewSignalBuffer(numSamples, input.SampleRate(), input.CenterFreq())

	// Per ogni campione nel buffer in ingresso
	for n := 0; n < numSamples; n++ {
		// 1. Spostiamo la storia della delay line (scorriamo i vecchi campioni)
		copy(f.history[1:], f.history[:numTaps-1])
		f.history[0] = input.GetSample(n) // Il campione attuale diventa il più recente

		// 2. Eseguiamo la Convoluzione (Moltiplica e Accumula)
		var acc Sample // Inizia a (0,0) grazie al tipo di Go
		for k := 0; k < numTaps; k++ {
			// Moltiplichiamo il campione storico per lo scalare del tap corrispondente
			scaled := f.history[k].Scale(f.taps[k])
			acc = acc.Add(scaled)
		}

		// Salviamo il risultato nel buffer di uscita
		output.SetSample(n, acc)
	}

	return output
}

// SetCutoff permette di cambiare la frequenza di taglio del filtro a runtime.
func (f *FIRFilter) SetCutoff(sampleRate SampleRate, newCutoffHz Frequency) {
	// Ricalcoliamo la frequenza normalizzata
	normCutoff := float64(newCutoffHz) / float64(sampleRate)

	// Generiamo i nuovi coefficienti usando la funzione matematica pura che abbiamo in dsp
	newTaps := ComputeLowPassTaps(len(f.taps), normCutoff)

	// Sostituiamo i taps in modo atomico
	f.taps = newTaps
}
