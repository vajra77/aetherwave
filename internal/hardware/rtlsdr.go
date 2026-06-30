package hardware

/*
#cgo LDFLAGS: -lrtlsdr -L/opt/homebrew/lib
#cgo CFLAGS: -I/opt/homebrew/include
#include <rtl-sdr.h>
#include <stdlib.h>
*/
import "C"

import (
	"aetherwave/internal/dsp"
	"errors"
	"fmt"
	"unsafe"
)

var ErrRead = errors.New("read error")

type Device struct {
	dev *C.rtlsdr_dev_t // Il puntatore opaco C al dispositivo RTL-SDR
}

// NewDevice apre la chiavetta Nooelec (indice 0) e configura frequenza e sample rate
func NewDevice(freq dsp.Frequency, sampleRate dsp.SampleRate) (*Device, error) {
	var dev *C.rtlsdr_dev_t

	// 1. Apriamo il dispositivo con indice 0 (la prima chiavetta USB inserita)
	// Passiamo l'indirizzo del puntatore a Cgo usando unsafe.Pointer
	res := C.rtlsdr_open((**C.rtlsdr_dev_t)(unsafe.Pointer(&dev)), C.uint32_t(0))
	if res < 0 {
		return nil, fmt.Errorf("impossibile aprire RTL-SDR con indice 0 (codice errore C: %d)", res)
	}

	d := &Device{dev: dev}

	// 2. Impostiamo la Frequenza Centrale
	// C.uint32_t fa il cast dal tuo dsp.Frequency (uint64) al tipo richiesto dal C
	res = C.rtlsdr_set_center_freq(d.dev, C.uint32_t(freq))
	if res < 0 {
		d.Close()
		return nil, errors.New("errore durante l'impostazione della frequenza centrale")
	}

	// 3. Impostiamo la Frequenza di Campionamento (Sample Rate)
	res = C.rtlsdr_set_sample_rate(d.dev, C.uint32_t(sampleRate))
	if res < 0 {
		d.Close()
		return nil, errors.New("errore durante l'impostazione del sample rate")
	}

	// 4. Abilitiamo il guadagno automatico dell'hardware (AGC) per iniziare in modo semplice
	C.rtlsdr_set_tuner_gain_mode(d.dev, 0) // 0 = Automatico, 1 = Manuale

	// 5. Resettiamo i buffer interni per svuotare i vecchi dati accumulati
	C.rtlsdr_reset_buffer(d.dev)

	return d, nil
}

// ReadBlock riempie uno slice di byte in Go attingendo direttamente dalla memoria del C
func (d *Device) ReadBlock(size int) ([]byte, error) {
	// Creiamo lo slice di byte in Go in cui confluiranno i dati I/Q
	buf := make([]byte, size)

	var nRead C.int

	// Chiamiamo la lettura sincrona di librtlsdr.
	// Dobbiamo passare il puntatore al primo elemento dello slice Go castato a un puntatore C (*C.uint8_t)
	res := C.rtlsdr_read_sync(
		d.dev,
		unsafe.Pointer(&buf[0]), // <--- Basta passarlo semplicemente come unsafe.Pointer
		C.int(size),
		&nRead,
	)

	if res < 0 {
		return nil, fmt.Errorf("errore durante la lettura sincrona dall'hardware (codice: %v)", res)
	}

	// Restituiamo il buffer Go, tagliato alla quantità di byte effettivamente letti dal C
	return buf[:int(nRead)], nil
}

func (d *Device) ReadBlockInto(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	var nRead C.int
	res := C.rtlsdr_read_sync(
		d.dev,
		unsafe.Pointer(&buf[0]),
		C.int(len(buf)),
		&nRead,
	)
	if res < 0 {
		return 0, ErrRead
	}

	return int(nRead), nil
}

// Close chiude la connessione con l'hardware e rilascia la memoria USB
func (d *Device) Close() {
	if d.dev != nil {
		C.rtlsdr_close(d.dev)
		d.dev = nil
	}
}
