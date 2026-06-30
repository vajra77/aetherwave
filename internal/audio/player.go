package audio

import (
	"fmt"

	"github.com/gordonklaus/portaudio"
)

type Player struct {
	stream     *portaudio.Stream
	sampleRate SampleRate
	audioChan  <-chan Sample
}

// NewPlayer inizializza il sottosistema audio e apre lo stream sulla scheda audio predefinita
func NewPlayer(sampleRate SampleRate, audioChan <-chan Sample) (*Player, error) {
	// 1. Inizializziamo PortAudio (richiede i driver del sistema operativo)
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("impossibile inizializzare PortAudio: %w", err)
	}

	p := &Player{
		sampleRate: sampleRate,
		audioChan:  audioChan,
	}

	// 2. Apriamo lo stream predefinito: 0 ingressi, 1 canale di uscita (Mono), frequenza impostata
	// Passiamo la funzione di callback interna 'processAudio'
	stream, err := portaudio.OpenDefaultStream(
		0, // Canali di input
		1, // Canali di output (Mono per la radio)
		float64(sampleRate),
		0, // Dimensione del buffer dinamica impostata da PortAudio
		p.processAudio,
	)
	if err != nil {
		portaudio.Terminate()
		return nil, fmt.Errorf("impossibile aprire lo stream audio: %w", err)
	}

	p.stream = stream
	return p, nil
}

// Start avvia la riproduzione acustica
func (p *Player) Start() error {
	return p.stream.Start()
}

// processAudio è la callback invocata direttamente dal thread della scheda audio.
// Deve essere velocissima e non deve bloccarsi!
func (p *Player) processAudio(out []float32) {
	for i := 0; i < len(out); i++ {
		select {
		case sample := <-p.audioChan:
			// Convertiamo il nostro tipo audio.Sample nel float32 nativo per la scheda audio
			out[i] = float32(sample)
		default:
			// Se il canale è temporaneamente vuoto (il DSP è in ritardo),
			// scriviamo 0.0 (silenzio) per evitare rumori digitali molesti
			out[i] = 0.0
		}
	}
}

// Close ferma l'audio e rilascia le risorse hardware del PC
func (p *Player) Close() {
	if p.stream != nil {
		p.stream.Stop()
		p.stream.Close()
	}
	portaudio.Terminate()
}
