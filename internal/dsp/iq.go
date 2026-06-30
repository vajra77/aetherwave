package dsp

import (
	"errors"
	"math"
	"math/cmplx"
)

var (
	ErrNotPowerOfTwo = errors.New("size must be a power of two")
)

type IQSample struct {
	data complex64
}

func NewIQSample(i, q float32) IQSample {
	return IQSample{data: complex(i, q)}
}

func (s IQSample) I() float32 { return real(s.data) }

func (s IQSample) Q() float32 { return imag(s.data) }

// Magnitude calcola l'ampiezza del segnale sfruttando cmplx.Abs
func (s IQSample) Magnitude() float32 {
	// Convertiamo a complex128 solo per il calcolo della libreria standard
	return float32(cmplx.Abs(complex128(s.data)))
}

// Phase calcola la fase istantanea (angolo) usando cmplx.Phase
func (s IQSample) Phase() float32 {
	return float32(cmplx.Phase(complex128(s.data)))
}

// Conjugate restituisce il complesso coniugato sfruttando la funzione nativa (I - jQ)
func (s IQSample) Conjugate() IQSample {
	// Il coniugato di (i, q) si ottiene invertendo il segno della parte immaginaria
	return IQSample{data: complex(real(s.data), -imag(s.data))}
}

// Add esegue una somma vettoriale nativa velocissima
func (s IQSample) Add(other IQSample) IQSample {
	return IQSample{data: s.data + other.data}
}

// Multiply esegue la moltiplicazione complessa nativa (fondamentale per mixing e filtri)
func (s IQSample) Multiply(other IQSample) IQSample {
	return IQSample{data: s.data * other.data}
}

// Scale moltiplica il campione per un guadagno o un fattore di scala reale
func (s IQSample) Scale(factor float32) IQSample {
	return IQSample{data: s.data * complex(factor, 0)}
}

// --- SignalBuffer ---

type SignalBuffer struct {
	size       int
	sampleRate uint32
	centerFreq uint64
	samples    []IQSample
	spectrum   []float32
}

func NewSignalBuffer(size int, sampleRate uint32, centerFreq uint64) *SignalBuffer {
	return &SignalBuffer{
		size:       size,
		sampleRate: sampleRate,
		centerFreq: centerFreq,
		samples:    make([]IQSample, size),
	}
}

func NewSignalBufferFromRawBytes(rate uint32, freq uint64, raw []byte) (*SignalBuffer, error) {
	nIQSamples := len(raw) / 2
	if nIQSamples == 0 || (nIQSamples&(nIQSamples-1)) != 0 {
		return nil, ErrNotPowerOfTwo
	}

	buf := NewSignalBuffer(nIQSamples, rate, freq)

	for i := 0; i < len(raw); i += 2 {
		iFloat := (float32(raw[i]) - 127.5) / 127.5
		qFloat := (float32(raw[i+1]) - 127.5) / 127.5
		buf.samples[i/2] = NewIQSample(iFloat, qFloat)
	}
	return buf, nil
}

func (sb *SignalBuffer) Size() int                   { return sb.size }
func (sb *SignalBuffer) IQSampleRate() uint32        { return sb.sampleRate }
func (sb *SignalBuffer) CenterFreq() uint64          { return sb.centerFreq }
func (sb *SignalBuffer) Samples() []IQSample         { return sb.samples }
func (sb *SignalBuffer) GetSample(n int) IQSample    { return sb.samples[n] }
func (sb *SignalBuffer) SetSample(n int, s IQSample) { sb.samples[n] = s }

func (sb *SignalBuffer) ComputeSpectrum() []float32 {
	N := int(sb.Size())

	// Creiamo un array di lavoro temporaneo (questa sarà l'unica vera allocazione)
	workingIQSamples := make([]IQSample, N)
	for n := 0; n < N; n++ {
		hann := float32(0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(n)/float64(N-1))))
		workingIQSamples[n] = sb.samples[n].Scale(hann)
	}

	// Eseguiamo la FFT in-place sull'array di lavoro
	FFTi(workingIQSamples)

	// Da qui in poi la logica del tuo Shift e dei Decibel rimane identica
	shiftedResult := make([]IQSample, N)
	mid := N / 2
	for i := 0; i < N; i++ {
		shiftedResult[(i+mid)%N] = workingIQSamples[i]
	}

	spectrum := make([]float32, N)
	for i := 0; i < N; i++ {
		mag := shiftedResult[i].Magnitude()
		spectrum[i] = 20 * float32(math.Log10(float64(mag)+1e-10))
	}

	return spectrum
}
