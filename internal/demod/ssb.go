package demod

import (
	"aetherwave/internal/audio"
	"aetherwave/internal/dsp"
	"math"
)

type SSBDemodulator struct {
	mode       string // "USB" oppure "LSB"
	sampleRate dsp.SampleRate
	phase      float64 // Mantiene lo stato della fase della portante locale (BFO)
}

func NewSSB(mode string, sampleRate dsp.SampleRate) *SSBDemodulator {
	return &SSBDemodulator{
		mode:       mode,
		sampleRate: sampleRate,
	}
}

func (s *SSBDemodulator) Process(in *dsp.Signal, out []audio.Sample) {
	nSamples := in.Size()
	samples := in.Samples()

	// Scegliamo di quanto "spostare" il segnale per far coincidere la banda laterale con l'audio udibile.
	// Di solito per la SSB si sposta la frequenza di circa 1.5 kHz (0.0015 MHz) per centrare la voce.
	var shiftFreq float64 = 1500.0
	if s.mode == "LSB" {
		shiftFreq = -1500.0 // Direzione opposta per la banda inferiore
	}

	// Passo di fase per ogni campione (derivata della frequenza del BFO)
	phaseStep := 2.0 * math.Pi * shiftFreq / float64(s.sampleRate)

	for i := 0; i < nSamples; i++ {
		// 1. Generiamo il campione della portante locale e^(-j*fase)
		bfo := dsp.Sample(complex(float32(math.Cos(s.phase)), float32(-math.Sin(s.phase))))

		// 2. Moltiplicazione complessa (Eterodinaggio): spostiamo lo spettro radio verso l'audio
		shiftedSample := samples[i].Multiply(bfo)

		// 3. Estrazione dell'audio: nella SSB, dopo aver centrato lo spettro,
		// l'informazione audio reale si trova semplicemente nella PARTE REALE (I) del segnale complessivo!
		out[i] = audio.Sample(shiftedSample.I())

		// Incrementiamo la fase della portante locale per il prossimo campione
		s.phase += phaseStep
		if s.phase > 2.0*math.Pi {
			s.phase -= 2.0 * math.Pi
		}
	}

	// NOTA: Per un lavoro perfetto, dopo questo ciclo l'array out andrebbe
	// passato dentro un filtro FIR passa-basso molto stretto (es. 3 kHz) per tagliare
	// i rimasugli dell'altra banda laterale.
}
