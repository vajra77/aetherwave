package radio

import (
	"aetherwave/internal/audio"
	"aetherwave/internal/demod"
	"aetherwave/internal/dsp"
	"aetherwave/internal/hardware"
	"fmt"
	"sync"
)

const (
	ModeAM  = "AM"
	ModeFM  = "FM"
	ModeUSB = "USB"
	ModeLSB = "LSB"
	ModeCW  = "CW"
)

const MaxBuffers = 16
const BlockSize = 16384

type Stage interface {
	// Process prende i dati dallo stadio precedente e restituisce il risultato per il successivo.
	// Usiamo l'interfaccia {} o un tipo custom per permettere passaggi da Signal a []float32.
	Process(input interface{}) (interface{}, error)
}

type Receiver struct {
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

	// --- Canali per la gestione concorrente
	rawDataChan chan []byte
	bufferPool  chan []byte

	mu sync.RWMutex
}

// NewReceiver inizializza la catena in base alla configurazione richiesta
func NewReceiver(freq dsp.Frequency, rate dsp.SampleRate, mode string) *Receiver {
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

	r := Receiver{
		frequency:  freq,
		sampleRate: rate,
		mode:       mode,

		stopChan:      make(chan struct{}),
		fftOutputChan: make(chan []dsp.DBPower, 10),
		audioChan:     make(chan audio.Sample, 1024),

		demodulator: d,
		filter:      dsp.NewLowPassFilter(63, rate, 10000),
		resampler:   dsp.NewResampler(rate, audioSampleRate),

		rawDataChan: make(chan []byte, MaxBuffers),
		bufferPool:  make(chan []byte, MaxBuffers),
	}

	for i := 0; i < MaxBuffers; i++ {
		r.bufferPool <- make([]byte, BlockSize*2)
	}

	return &r
}

// SetFilterBandwidth permette alla GUI (es. tramite uno slider) o alla CLI
// di cambiare la larghezza di banda del filtro passa-basso in tempo reale.
func (p *Receiver) SetFilterBandwidth(bandwidthHz dsp.Frequency) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// In un filtro passa-basso reale simmetrico (I/Q), la frequenza di taglio
	// è pari a metà della larghezza di banda totale desiderata.
	cutoff := bandwidthHz / 2

	// Aggiorniamo il filtro interno usando il sample rate corrente della pipeline
	p.filter.SetCutoff(p.sampleRate, cutoff)
}

// Start avvia il loop di campionamento ed elaborazione in una goroutine separata
func (p *Receiver) Start() error {
	if p.isRunning {
		return nil
	}

	// Inizializza l'hardware realmente qui
	sdr, err := hardware.NewDevice(p.frequency, p.sampleRate)
	if err != nil {
		return fmt.Errorf("unable to init SDR hardware: %w", err)
	}
	p.sdr = sdr
	p.isRunning = true

	go p.hwReaderLoop()
	go p.dspProcessorLoop()

	return nil
}

// Stop arresta bruscamente la radio e libera l'hardware
func (p *Receiver) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isRunning {
		return
	}
	close(p.stopChan)
	p.sdr.Close()
	p.isRunning = false
}

// SetFrequency permette alla GUI di cambiare frequenza al volo (es. girando una manopola virtuale)
func (p *Receiver) SetFrequency(newFreq dsp.Frequency) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.frequency = newFreq
	// Comunica il cambio all'hardware
	// p.sdr.SetCenterFreq(newFreq)
}

func (p *Receiver) hwReaderLoop() {
	for p.isRunning {
		// Prendiamo un buffer vuoto pre-allocato dal pool
		var buf []byte
		select {
		case buf = <-p.bufferPool:
		default:
			// Se il pool è vuoto significa che il DSP è troppo lento!
			// Per non crashare, saltiamo o riutilizziamo l'ultimo (scarto controllato)
			buf = make([]byte, BlockSize*2)
		}

		// Travaso Cgo sincrono direttamente nel buffer pre-allocato
		// Modifichiamo ReadBlock per accettare il buffer invece di crearlo
		n, err := p.sdr.ReadBlockInto(buf)
		if err != nil {
			p.bufferPool <- buf // Restituiamo il buffer in caso di errore
			continue
		}

		// Spediamo il blocco pieno al DSP
		p.rawDataChan <- buf[:n]
	}
}

func (p *Receiver) dspProcessorLoop() {

	for {
		p.mu.RLock()
		running := p.isRunning
		p.mu.RUnlock()

		if !running {
			return
		}

		select {
		case <-p.stopChan:
			return
		case rawBytes := <-p.rawDataChan:
			p.mu.RLock()
			currentFreq := p.frequency
			currentRate := p.sampleRate
			p.mu.RUnlock()

			// 1. Trasformiamo in Signal e applichiamo la catena DSP
			rawSignal, _ := dsp.NewSignalFromRawBytes(currentRate, currentFreq, rawBytes)
			filteredSignal := p.filter.Process(rawSignal)
			hiFreqAudio := p.demodulator.Process(filteredSignal)
			readyAudio := p.resampler.Process(hiFreqAudio)

			// 2. IMPORTANTISSIMO: Restituiamo il buffer originale al pool
			// così il Produttore può riutilizzarlo, evitando il Garbage Collector!
			// Dobbiamo estrarre lo slice originario (resettando la slice expression)
			p.bufferPool <- rawBytes[:cap(rawBytes)]

			// 3. Spingiamo l'audio verso il player
			for _, sample := range readyAudio {
				p.audioChan <- sample
			}
		}
	}
}

// FFTChannel permette alla GUI di ricevere i dati dello spettro in tempo reale
func (p *Receiver) FFTChannel() <-chan []dsp.DBPower {
	return p.fftOutputChan
}

// AudioChannel permette al player audio di attingere ai campioni pronti
func (p *Receiver) AudioChannel() <-chan audio.Sample {
	return p.audioChan
}
