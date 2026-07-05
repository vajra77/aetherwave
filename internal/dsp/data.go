package dsp

import (
	"errors"
	"fmt"
	"math"
	"math/cmplx"
)

var (
	ErrNotPowerOfTwo = errors.New("size must be a power of two")
)

type Frequency float64
type SampleRate uint32

// ---- Metodi di utilità per Frequency ----

// MHz restituisce il valore della frequenza in MegaHertz (float64)
func (f Frequency) MHz() float64 {
	return float64(f) / 1e6
}

// String formatta la frequenza in modo leggibile (es. "100.00 MHz") per la CLI o la GUI
func (f Frequency) String() string {
	if f >= 1e6 {
		return fmt.Sprintf("%.2f MHz", f.MHz())
	}
	return fmt.Sprintf("%.2f kHz", float64(f)/1e3)
}

// ---- Metodi di utilità per SampleRate ----

// Period restituisce la durata temporale in secondi di un singolo campione (1 / SampleRate)
func (sr SampleRate) Period() float64 {
	return 1.0 / float64(sr)
}

func (sr SampleRate) String() string {
	return fmt.Sprintf("%d Hz", sr)
}

// NormalizeFreq calcola la frequenza normalizzata rispetto a questo SampleRate.
// Il risultato è un float64 compreso tra 0.0 e 0.5 (limite di Nyquist).
func (sr SampleRate) NormalizeFreq(cutoff Frequency) float64 {
	return float64(cutoff) / float64(sr)
}

type Sample complex64

func (s Sample) I() float32 { return real(s) }

func (s Sample) Q() float32 { return imag(s) }

// Magnitude calcola l'ampiezza del segnale sfruttando cmplx.Abs
func (s Sample) Magnitude() float32 {
	// Convertiamo a complex128 solo per il calcolo della libreria standard
	return float32(cmplx.Abs(complex128(s)))
}

// Phase calcola la fase istantanea (angolo) usando cmplx.Phase
func (s Sample) Phase() float32 {
	return float32(cmplx.Phase(complex128(s)))
}

// Conjugate restituisce il complesso coniugato sfruttando la funzione nativa (I - jQ)
func (s Sample) Conjugate() Sample {
	// Il coniugato di (i, q) si ottiene invertendo il segno della parte immaginaria
	return Sample(complex(real(s), -imag(s)))
}

// Add esegue una somma vettoriale nativa velocissima
func (s Sample) Add(other Sample) Sample {
	return Sample(s + other)
}

// Multiply esegue la moltiplicazione complessa nativa (fondamentale per mixing e filtri)
func (s Sample) Multiply(other Sample) Sample {
	return Sample(s * other)
}

// Scale moltiplica il campione per un guadagno o un fattore di scala reale
func (s Sample) Scale(factor float32) Sample {
	return Sample(complex64(s) * complex(factor, 0))
}

// --- Signal ---

type Signal struct {
	size       int
	sampleRate SampleRate
	samples    []Sample
	spectrum   []float32
}

func NewEmptySignal(sampleRate SampleRate, size int) *Signal {
	return new(Signal{
		size:       size,
		sampleRate: sampleRate,
		samples:    make([]Sample, size),
	})
}

func NewSignalFromSamples(sampleRate SampleRate, samples []Sample) *Signal {
	return &Signal{
		size:       len(samples),
		sampleRate: sampleRate,
		samples:    samples,
	}
}

func DecimateRawBytesIntoSignal(inRate, tgtRate SampleRate, raw []byte, out *Signal) error {
	out.samples = out.samples[:cap(out.samples)]

	nSamples := len(raw) / 2
	if nSamples == 0 || (nSamples&(nSamples-1)) != 0 {
		return ErrNotPowerOfTwo
	}

	decFactor := int(math.Round(float64(inRate) / float64(tgtRate)))

	dstIdx := 0
	step := 2 * decFactor
	for i := 0; i < len(raw); i += step {
		if i+1 >= len(raw) || dstIdx >= out.size {
			break
		}
		iFloat := (float32(raw[i]) - 127.5) / 127.5
		qFloat := (float32(raw[i+1]) - 127.5) / 127.5
		out.SetSample(dstIdx, Sample(complex(iFloat, qFloat)))
		dstIdx++
	}

	out.size = dstIdx
	out.samples = out.samples[:dstIdx]

	return nil
}

func (s *Signal) Size() int {
	return s.size
}

func (s *Signal) Samples() []Sample {
	return s.samples
}

func (s *Signal) SampleRate() SampleRate {
	return s.sampleRate
}

func (s *Signal) Spectrum() []float32 {
	return s.spectrum
}

func (s *Signal) GetSample(i int) Sample {
	return s.samples[i]
}

func (s *Signal) SetSample(i int, v Sample) {
	s.samples[i] = v
}
