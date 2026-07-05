package radio

import (
	"aetherwave/internal/audio"
	"aetherwave/internal/demod"
	"aetherwave/internal/dsp"
	"aetherwave/internal/hardware"
	"fmt"
	"math"
	"sync"
)

const (
	ModeAM  = "AM"
	ModeFM  = "FM"
	ModeUSB = "USB"
	ModeLSB = "LSB"
	ModeCW  = "CW"
)

const MaxBuffers = 32
const BlockSize = 16384
const RadioSampleRate dsp.SampleRate = 1024000
const AudioSampleRate audio.SampleRate = 48000

type Stage interface {
	// Process prende i dati dallo stadio precedente e restituisce il risultato per il successivo.
	// Usiamo l'interfaccia {} o un tipo custom per permettere passaggi da Signal a []float32.
	Process(input interface{}) (interface{}, error)
}

type Receiver struct {
	frequency dsp.Frequency
	profile   *Profile

	isRunning bool
	stopChan  chan struct{}

	// --- Componenti interni della Catena (Ora nello Stato!) ---
	sdr         *hardware.Device
	filter      *dsp.FIRFilter
	demodulator demod.Demodulator
	resampler   *dsp.Resampler

	// --- buffer fissi ---
	decimatedSignal *dsp.Signal
	filteredSignal  *dsp.Signal
	hiFreqAudio     []audio.Sample
	readyAudio      []audio.Sample

	// --- Canali di Output per l'esterno (GUI / Audio Player) ---
	fftOutputChan chan []dsp.DBPower
	audioChan     chan audio.Sample

	// --- Canali per la gestione concorrente
	rawDataChan chan []byte
	bufferPool  chan []byte

	mu sync.RWMutex
}

// NewReceiver inizializza la catena in base alla configurazione richiesta
func NewReceiver(freq dsp.Frequency, p *Profile) *Receiver {
	// 1. Calcolo dei fattori di conversione
	decFactor := int(math.Round(float64(RadioSampleRate) / float64(p.TargetSampleRate)))
	if decFactor < 1 {
		decFactor = 1
	}

	// Supponiamo che la dimensione del blocco hardware sia fissa a 32768 byte
	hardwareBlockBytes := BlockSize * 2
	rawSamples := hardwareBlockBytes / 2 // 16384

	// 2. Calcolo delle dimensioni esatte
	decimatedSize := rawSamples / decFactor // 4096

	// Per l'audio in uscita dal resampler (250000 / 48000 = 5.2083)
	resampleRatio := float64(p.TargetSampleRate) / float64(AudioSampleRate)
	// Aggiungiamo un piccolo margine di sicurezza (+32) per gli arrotondamenti dei filtri interni del resampler
	readyAudioSize := int(float64(decimatedSize)/resampleRatio) + 32

	rcv := Receiver{
		frequency: freq,
		profile:   p,

		stopChan:      make(chan struct{}),
		fftOutputChan: make(chan []dsp.DBPower, 10),
		audioChan:     make(chan audio.Sample, 16384),

		demodulator: p.Demodulator,
		filter:      dsp.NewLowPassFilter(p.Taps, p.TargetSampleRate, p.HighCutoff),
		resampler:   dsp.NewResampler(p.TargetSampleRate, AudioSampleRate),

		decimatedSignal: dsp.NewEmptySignal(p.TargetSampleRate, decimatedSize),
		filteredSignal:  dsp.NewEmptySignal(p.TargetSampleRate, decimatedSize),
		hiFreqAudio:     make([]audio.Sample, decimatedSize),
		readyAudio:      make([]audio.Sample, readyAudioSize),
		rawDataChan:     make(chan []byte, MaxBuffers),
		bufferPool:      make(chan []byte, MaxBuffers),
	}

	for i := 0; i < MaxBuffers; i++ {
		rcv.bufferPool <- make([]byte, BlockSize*2)
	}

	return &rcv
}

// SetFilterBandwidth permette alla GUI (es. tramite uno slider) o alla CLI
// di cambiare la larghezza di banda del filtro passa-basso in tempo reale.
func (r *Receiver) SetFilterBandwidth(bandwidthHz dsp.Frequency) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// In un filtro passa-basso reale simmetrico (I/Q), la frequenza di taglio
	// è pari a metà della larghezza di banda totale desiderata.
	cutoff := bandwidthHz / 2

	// Aggiorniamo il filtro interno usando il sample rate corrente della pipeline
	r.filter.SetCutoff(r.profile.TargetSampleRate, cutoff)
}

// Start avvia il loop di campionamento ed elaborazione in una goroutine separata
func (r *Receiver) Start() error {
	if r.isRunning {
		return nil
	}

	// Inizializza l'hardware realmente qui
	sdr, err := hardware.NewDevice(r.frequency, RadioSampleRate)
	if err != nil {
		return fmt.Errorf("unable to init SDR hardware: %w", err)
	}

	r.sdr = sdr
	r.isRunning = true

	go r.hwReaderLoop()
	go r.dspProcessorLoop()

	return nil
}

// Stop arresta bruscamente la radio e libera l'hardware
func (r *Receiver) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isRunning {
		return
	}
	close(r.stopChan)
	r.sdr.Close()
	r.isRunning = false
}

// SetFrequency permette alla GUI di cambiare frequenza al volo (es. girando una manopola virtuale)
func (r *Receiver) SetFrequency(newFreq dsp.Frequency) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.frequency = newFreq
	// Comunica il cambio all'hardware
	// r.sdr.SetCenterFreq(newFreq)
}

func (r *Receiver) hwReaderLoop() {
	for r.isRunning {
		// Prendiamo un buffer vuoto pre-allocato dal pool
		var buf []byte
		select {
		case buf = <-r.bufferPool:
		default:
			// Se il pool è vuoto significa che il DSP è troppo lento!
			// Per non crashare, saltiamo o riutilizziamo l'ultimo (scarto controllato)
			buf = make([]byte, BlockSize*2)
		}

		// Travaso Cgo sincrono direttamente nel buffer pre-allocato
		// Modifichiamo ReadBlock per accettare il buffer invece di crearlo
		n, err := r.sdr.ReadBlockInto(buf)
		if err != nil {
			r.bufferPool <- buf // Restituiamo il buffer in caso di errore
			continue
		}

		// Spediamo il blocco pieno al DSP
		r.rawDataChan <- buf[:n]
	}
}

func (r *Receiver) dspProcessorLoop() {

	for {
		r.mu.RLock()
		running := r.isRunning
		r.mu.RUnlock()

		if !running {
			return
		}

		select {
		case <-r.stopChan:
			return
		case rawBytes := <-r.rawDataChan:
			r.mu.RLock()
			currentRate := r.profile.TargetSampleRate
			r.mu.RUnlock()

			_ = dsp.DecimateRawBytesIntoSignal(RadioSampleRate, currentRate, rawBytes, r.decimatedSignal)
			r.bufferPool <- rawBytes[:cap(rawBytes)]

			r.filter.Process(r.decimatedSignal, r.filteredSignal)

			r.demodulator.Process(r.filteredSignal, r.hiFreqAudio)

			nAudioSamples := r.resampler.Process(r.hiFreqAudio, r.readyAudio)

			for _, sample := range r.readyAudio[:nAudioSamples] {
				r.audioChan <- sample
			}
		}
	}
}

// FFTChannel permette alla GUI di ricevere i dati dello spettro in tempo reale
func (r *Receiver) FFTChannel() <-chan []dsp.DBPower {
	return r.fftOutputChan
}

// AudioChannel permette al player audio di attingere ai campioni pronti
func (r *Receiver) AudioChannel() <-chan audio.Sample {
	return r.audioChan
}
