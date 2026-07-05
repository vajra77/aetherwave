package radio

import (
	"aetherwave/internal/demod"
	"aetherwave/internal/dsp"
)

type ProfileType int

const (
	AMW = iota
	AMN
	USB
	LSB
	CW
	WFM
	NFM
)

type Profile struct {
	Name             string
	Demodulator      demod.Demodulator
	LowCutoff        dsp.Frequency
	HighCutoff       dsp.Frequency
	Taps             int
	TargetSampleRate dsp.SampleRate
}

var Profiles = map[ProfileType]*Profile{
	AMW: &Profile{
		Name:             "AM/W",
		Demodulator:      demod.NewAM(),
		LowCutoff:        0,
		HighCutoff:       3000,
		Taps:             63,
		TargetSampleRate: 250000,
	},
	AMN: &Profile{
		Name:             "AM/N",
		Demodulator:      demod.NewAM(),
		LowCutoff:        0,
		HighCutoff:       1350,
		Taps:             127,
		TargetSampleRate: 250000,
	},
	USB: &Profile{
		Name:             "USB",
		Demodulator:      demod.NewSSB("USB", 250000),
		LowCutoff:        300,
		HighCutoff:       3000,
		Taps:             127,
		TargetSampleRate: 250000,
	},
	LSB: &Profile{
		Name:             "LSB",
		Demodulator:      demod.NewSSB("LSB", 250000),
		LowCutoff:        -3000,
		HighCutoff:       300,
		Taps:             127,
		TargetSampleRate: 250000,
	},
	CW: &Profile{
		Name:             "CW",
		Demodulator:      demod.NewCW(250000, 2400),
		LowCutoff:        -900,
		HighCutoff:       -400,
		Taps:             100,
		TargetSampleRate: 250000,
	},
	WFM: &Profile{
		Name:             "FM/W",
		Demodulator:      demod.NewFM(),
		LowCutoff:        0,
		HighCutoff:       75000,
		Taps:             63,
		TargetSampleRate: 1024000,
	},
	NFM: &Profile{
		Name:             "FM/N",
		Demodulator:      demod.NewFM(),
		LowCutoff:        0,
		HighCutoff:       5500,
		Taps:             127,
		TargetSampleRate: 250000,
	},
}
