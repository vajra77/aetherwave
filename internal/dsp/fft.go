package dsp

import (
	"math"
)

// FFTr è il motore ricorsivo della FFT.
// Lavora direttamente sugli IQSample sfruttando la loro astrazione.
func FFTr(x []IQSample) []IQSample {
	N := len(x)

	if N <= 1 {
		return x
	}

	// 1. Divide: separa gli indici even dai odd
	even := make([]IQSample, N/2)
	odd := make([]IQSample, N/2)
	for i := 0; i < N/2; i++ {
		even[i] = x[2*i]
		odd[i] = x[2*i+1]
	}

	// 2. Impera: calcola la FFT ricorsivamente
	fftEven := FFTr(even)
	fftOdd := FFTr(odd)

	// 3. Combina: unisci i risultati usando i twiddle factors
	X := make([]IQSample, N)
	for k := 0; k < N/2; k++ {
		angle := -2.0 * math.Pi * float64(k) / float64(N)

		// Creiamo il twiddle factor sfruttando il costruttore del nostro tipo
		// e^(-j*θ) = cos(θ) + j*sin(θ)
		twiddle := NewIQSample(float32(math.Cos(angle)), float32(math.Sin(angle)))

		// Moltiplicazione complessa astratta: odd * twiddle
		turnedOdd := fftOdd[k].Multiply(twiddle)

		// Sfruttiamo la simmetria (Farfalla della FFT / Butterfly network)
		X[k] = fftEven[k].Add(turnedOdd)
		X[k+N/2] = fftEven[k].Add(turnedOdd.Scale(-1)) // Sottrazione tramite scale
	}

	return X
}

// FFTi calcola la Fast Fourier Transform in-place (senza allocazioni ricorsive).
// Modifica direttamente l'array passato in input. La dimensione deve essere una potenza di 2.
func FFTi(x []IQSample) {
	n := len(x)
	if n <= 1 {
		return
	}

	// 1. Riordina i campioni prima di iniziare i cicli
	bitReverseReorder(x)

	// 2. Ciclo principale dell'algoritmo iterativo
	// 'size' raddoppia a ogni stadio (2, 4, 8, ..., N)
	for size := 2; size <= n; size <<= 1 {
		halfSize := size >> 1
		// Calcoliamo l'angolo base per lo stadio corrente
		angleStep := -2.0 * math.Pi / float64(size)

		// Per ogni sottogruppo di questa dimensione
		for i := 0; i < n; i += size {
			// Elaborazione a "Farfalla" (Butterfly)
			for k := 0; k < halfSize; k++ {
				angle := float64(k) * angleStep

				// Generiamo il twiddle factor
				twiddle := NewIQSample(float32(math.Cos(angle)), float32(math.Sin(angle)))

				// Indici dei due elementi che si incrociano nella farfalla
				idxPari := i + k
				idxDispari := i + k + halfSize

				// Moltiplichiamo l'elemento dispari per il fattore di rotazione
				tau := x[idxDispari].Multiply(twiddle)

				// Aggiorniamo i valori in-place nello slice originario
				x[idxDispari] = x[idxPari].Add(tau.Scale(-1)) // Sottrazione: Pari - Tau
				x[idxPari] = x[idxPari].Add(tau)              // Somma: Pari + Tau
			}
		}
	}
}

// bitReverseReorder riordina lo slice in-place invertendo i bit degli indici.
func bitReverseReorder(x []IQSample) {
	n := len(x)
	j := 0
	for i := 0; i < n; i++ {
		if i < j {
			// Scambiamo i campioni se l'indice invertito è maggiore
			x[i], x[j] = x[j], x[i]
		}
		bit := n >> 1
		for bit&j != 0 {
			j ^= bit
			bit >>= 1
		}
		j ^= bit
	}
}
