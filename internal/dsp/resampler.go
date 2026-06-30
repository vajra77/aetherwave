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
func (r *Resampler) Process(audioIn []audio.Sample) []audio.Sample {
	if len(audioIn) == 0 {
		return nil
	}

	// 1. Calcoliamo il rapporto di conversione
	ratio := float64(r.inputRate) / float64(r.outputRate)

	// 2. Calcoliamo quanti campioni conterrà il buffer di uscita
	newSize := int(float64(len(audioIn)) / ratio)
	audioOut := make([]audio.Sample, newSize)

	// 3. Algoritmo di campionamento
	sourceIndex := 0.0
	for i := 0; i < newSize; i++ {
		// Arrotondiamo l'indice float all'intero più vicino
		nearestInputIndex := int(sourceIndex)

		// Protezione di sicurezza per non sforare l'array originale
		if nearestInputIndex >= len(audioIn) {
			nearestInputIndex = len(audioIn) - 1
		}

		// Copiamo il campione nel nuovo buffer a bassa frequenza
		audioOut[i] = audioIn[nearestInputIndex]

		// Facciamo avanzare l'indice del nostro passo fisso (ratio)
		sourceIndex += ratio
	}

	return audioOut
}
