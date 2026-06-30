package radio

import (
	"aetherwave/internal/dsp"
	"aetherwave/internal/hardware"
)

const BlockSize = 32768

type Stage interface {
	// Process prende i dati dallo stadio precedente e restituisce il risultato per il successivo.
	// Usiamo l'interfaccia {} o un tipo custom per permettere passaggi da SignalBuffer a []float32.
	Process(input interface{}) (interface{}, error)
}

type Config struct {
	Frequency  uint64
	SampleRate uint32
	Mode       string // "AM", "FM", "RAW"
}

type Pipeline struct {
	config    Config
	sdr       *hardware.Device
	isRunning bool
	stopChan  chan struct{}

	// Hooks per la GUI (Canali di sola lettura per l'esterno)
	fftOutputChan chan []float32 // La GUI si collega qui per disegnare il Waterfall
	audioChan     chan []float32 // La GUI o il Player leggono qui per l'audio
}

// NewPipeline inizializza la catena in base alla configurazione richiesta
func NewPipeline(cfg Config) *Pipeline {
	return &Pipeline{
		config:        cfg,
		stopChan:      make(chan struct{}),
		fftOutputChan: make(chan []float32, 10),
		audioChan:     make(chan []float32, 10),
	}
}

// Start avvia il loop di campionamento ed elaborazione in una goroutine separata
func (p *Pipeline) Start() error {
	if p.isRunning {
		return nil
	}

	// Inizializza l'hardware realmente qui
	sdr, err := hardware.NewDevice(p.config.Frequency, p.config.SampleRate)
	if err != nil {
		return err
	}
	p.sdr = sdr
	p.isRunning = true

	// Avvia l'orchestrazione asincrona
	go p.run()
	return nil
}

// Stop arresta bruscamente la radio e libera l'hardware
func (p *Pipeline) Stop() {
	if !p.isRunning {
		return
	}
	close(p.stopChan)
	p.sdr.Close()
	p.isRunning = false
}

// SetFrequency permette alla GUI di cambiare frequenza al volo (es. girando una manopola virtuale)
func (p *Pipeline) SetFrequency(newFreq uint64) {
	p.config.Frequency = newFreq
	// Comunica il cambio all'hardware
	// p.sdr.SetCenterFreq(newFreq)
}

// FFTChannel permette alla GUI di ricevere i dati dello spettro in tempo reale
func (p *Pipeline) FFTChannel() <-chan []float32 {
	return p.fftOutputChan
}

// AudioChannel permette al player audio di attingere ai campioni pronti
func (p *Pipeline) AudioChannel() <-chan []float32 {
	return p.audioChan
}

func (p *Pipeline) run() {
	// Inizializzazione degli stub dei componenti dsp/demod interni
	// (Vengono configurati internamente in base a p.config.Mode)

	for {
		select {
		case <-p.stopChan:
			return
		default:
			// 1. Leggi dall'hardware
			rawBytes, _ := p.sdr.ReadBlock(BlockSize)

			// 2. Converti in SignalBuffer
			rawSignal, _ := dsp.NewSignalBufferFromRawBytes(p.config.SampleRate, p.config.Frequency, rawBytes)

			// ---- HOOK PER LA GUI (SPETTRO) ----
			// Calcoliamo la FFT sul buffer grezzo e la spariamo sul canale.
			// Se il canale è pieno (la GUI è lenta a disegnare), saltiamo il blocco per non rallentare l'audio.
			select {
			case p.fftOutputChan <- rawSignal.Spectrum():
			default: // Non bloccante
			}

			// 3. Catena DSP e Demodulazione
			//filteredSignal := filter.Process(bufferGrezzo)
			// audioAltaFreq := demodulatore.Demodulate(bufferFiltrato)
			// audioPronto := resampler.Process(audioAltaFreq)

			// ---- HOOK PER L'AUDIO ----
			var audioData []float32 // Risultato finale degli stub sopra
			select {
			case p.audioChan <- audioData:
			case <-p.stopChan:
				return
			}
		}
	}
}
