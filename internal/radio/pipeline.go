package radio

import (
	"aetherwave/internal/audio"
	"aetherwave/internal/demod"
	"aetherwave/internal/dsp"
	"aetherwave/internal/hardware"
)

const (
	ModeAM  = "AM"
	ModeFM  = "FM"
	ModeUSB = "USB"
	ModeLSB = "LSB"
	ModeCW  = "CW"
)

const BlockSize = 32768

type Stage interface {
	// Process prende i dati dallo stadio precedente e restituisce il risultato per il successivo.
	// Usiamo l'interfaccia {} o un tipo custom per permettere passaggi da SignalBuffer a []float32.
	Process(input interface{}) (interface{}, error)
}

type Pipeline struct {
	frequency  dsp.Frequency
	sampleRate dsp.SampleRate
	mode       string

	isRunning bool
	stopChan  chan struct{}

	// --- Componenti interni della Catena (Ora nello Stato!) ---
	sdr         *hardware.Device
	filter      *dsp.FIRFilter
	demodulator demod.Demodulator
	resampler   *dsp.Resampler

	// --- Canali di Output per l'esterno (GUI / Audio Player) ---
	fftOutputChan chan []dsp.DBPower
	audioChan     chan audio.Sample
}

// NewPipeline inizializza la catena in base alla configurazione richiesta
func NewPipeline(freq dsp.Frequency, rate dsp.SampleRate, mode string) *Pipeline {
	const audioSampleRate audio.SampleRate = 48000

	// 1. Selezione polimorfica del demodulatore
	var d demod.Demodulator
	switch mode {
	case "AM":
		d = demod.NewAM()
	case "FM":
		d = demod.NewFM()
	default:
		// Gestione di fallback o log di errore
		d = demod.NewAM()
	}

	return &Pipeline{
		frequency:  freq,
		sampleRate: rate,
		mode:       mode,

		stopChan:      make(chan struct{}),
		fftOutputChan: make(chan []dsp.DBPower, 10),
		audioChan:     make(chan audio.Sample, 1024),

		demodulator: d,
		filter:      dsp.NewLowPassFilter(63, rate, 10000),
		resampler:   dsp.NewResampler(rate, audioSampleRate),
	}
}

// SetFilterBandwidth permette alla GUI (es. tramite uno slider) o alla CLI
// di cambiare la larghezza di banda del filtro passa-basso in tempo reale.
func (p *Pipeline) SetFilterBandwidth(bandwidthHz dsp.Frequency) {
	// In un filtro passa-basso reale simmetrico (I/Q), la frequenza di taglio
	// è pari a metà della larghezza di banda totale desiderata.
	cutoff := bandwidthHz / 2

	// Aggiorniamo il filtro interno usando il sample rate corrente della pipeline
	p.filter.SetCutoff(p.sampleRate, cutoff)
}

// Start avvia il loop di campionamento ed elaborazione in una goroutine separata
func (p *Pipeline) Start() error {
	if p.isRunning {
		return nil
	}

	// Inizializza l'hardware realmente qui
	sdr, err := hardware.NewDevice(p.frequency, p.sampleRate)
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
func (p *Pipeline) SetFrequency(newFreq dsp.Frequency) {
	p.frequency = newFreq
	// Comunica il cambio all'hardware
	// p.sdr.SetCenterFreq(newFreq)
}

// FFTChannel permette alla GUI di ricevere i dati dello spettro in tempo reale
func (p *Pipeline) FFTChannel() <-chan []dsp.DBPower {
	return p.fftOutputChan
}

// AudioChannel permette al player audio di attingere ai campioni pronti
func (p *Pipeline) AudioChannel() <-chan audio.Sample {
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
			rawSignal, _ := dsp.NewSignalBufferFromRawBytes(p.sampleRate, p.frequency, rawBytes)

			// ---- HOOK PER LA GUI (SPETTRO) ----
			// Calcoliamo la FFT sul buffer grezzo e la spariamo sul canale.
			// Se il canale è pieno (la GUI è lenta a disegnare), saltiamo il blocco per non rallentare l'audio.
			select {
			case p.fftOutputChan <- dsp.ComputeSpectrum(rawSignal):
			default: // Non bloccante
			}

			// 3. Catena DSP e Demodulazione
			filteredSignal := p.filter.Process(rawSignal)
			hiFreqAudio := p.demodulator.Process(filteredSignal)
			loFreqAudio := p.resampler.Process(hiFreqAudio)

			for _, sample := range loFreqAudio {
				select {
				case p.audioChan <- sample:
				case <-p.stopChan:
					return
				}
			}
		}
	}
}
