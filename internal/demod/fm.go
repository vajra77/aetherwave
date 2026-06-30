package demod

import (
	"aetherwave/internal/audio"
	"aetherwave/internal/dsp"
	"math"
)

type FMDemodulator struct {
	lastPhase float32 // Memorizza la fase del blocco precedente per mantenere la continuità
}

func NewFM() *FMDemodulator {
	return &FMDemodulator{}
}

func (fm *FMDemodulator) Process(in *dsp.SignalBuffer) []audio.Sample {
	nSamples := in.Size()
	samples := in.Samples()
	audioOut := make([]audio.Sample, nSamples)

	for i := 0; i < nSamples; i++ {
		// 1. Calcoliamo la fase istantanea del campione corrente
		// Il tuo metodo Phase() restituisce float32 (radianti da -PI a +PI)
		currentPhase := samples[i].Phase()

		// 2. Calcoliamo la variazione di fase (derivata) rispetto al campione precedente
		phaseDiff := currentPhase - fm.lastPhase

		// 3. Phase Unwrapping (Correzione del salto di fase)
		// Se la differenza supera PI, significa che il vettore ha "scavalcato" l'asse.
		// Riportiamo la differenza nel range corretto aggiungendo o togliendo 2*PI.
		if phaseDiff > math.Pi {
			phaseDiff -= 2 * math.Pi
		} else if phaseDiff < -math.Pi {
			phaseDiff += 2 * math.Pi
		}

		// 4. Salviamo l'ampiezza dell'audio (la deviazione di frequenza è l'audio!)
		// Normalizziamo dividendo per PI per avere un valore idealmente compreso tra -1.0 e 1.0
		audioOut[i] = audio.Sample(phaseDiff / math.Pi)

		// Conserviamo la fase corrente per il prossimo ciclo (o il prossimo blocco hardware)
		fm.lastPhase = currentPhase
	}

	// NOTA PER IL FUTURO: La FM commerciale ha una forte enfasi sulle alte frequenze
	// all'emissione (Pre-emphasis). Per sentire la radio in modo perfetto, qui andrà
	// applicato un filtro dsp molto semplice chiamato "De-emphasis" (un passa-basso leggero).

	return audioOut
}
