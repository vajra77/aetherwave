package dsp

import (
	"aetherwave/internal/audio"
)

type Resampler struct {
	inputRate  SampleRate       // es. 2048000
	outputRate audio.SampleRate // es. 48000 (standard per il pacchetto audio)
}

// NewResampler inizializza il convertitore di frequenza
func NewResampler(from SampleRate, to audio.SampleRate) *Resampler {
	return &Resampler{
		inputRate:  from,
		outputRate: to,
	}
}

// Process prende i campioni audio ad alta frequenza e restituisce un nuovo array
// ridimensionato e pronto per la scheda audio
func (r *Resampler) Process(in []audio.Sample, out []audio.Sample) int {
	out = out[:cap(out)]
	if len(in) == 0 {
		return 0
	}

	// 1. Calcoliamo il rapporto di conversione
	ratio := float64(r.inputRate) / float64(r.outputRate)

	// 2. Calcoliamo quanti campioni conterrà il buffer di uscita
	newSize := int(float64(len(in)) / ratio)

	// 3. Algoritmo di campionamento
	sourceIndex := 0.0
	for i := 0; i < newSize; i++ {
		// Arrotondiamo l'indice float all'intero più vicino
		nearestInputIndex := int(sourceIndex)

		// Protezione di sicurezza per non sforare l'array originale
		if nearestInputIndex >= len(in) {
			nearestInputIndex = len(in) - 1
		}

		// Copiamo il campione nel nuovo buffer a bassa frequenza
		out[i] = in[nearestInputIndex]

		// Facciamo avanzare l'indice del nostro passo fisso (ratio)
		sourceIndex += ratio
	}

	return newSize
}
