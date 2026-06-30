package hardware

type Device struct {
	freq       uint64
	sampleRate uint32
}

func NewDevice(freq uint64, sampleRate uint32) (*Device, error) {
	return &Device{}, nil
}

func (d *Device) ReadBlock(size int) ([]byte, error) {
	return nil, nil
}

func (d *Device) Close() {
}
