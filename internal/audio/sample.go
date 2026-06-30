package audio

type Sample float32

func (s Sample) Add(v Sample) Sample { return Sample(float32(s) + float32(v)) }
func (s Sample) Sub(v Sample) Sample { return Sample(float32(s) - float32(v)) }
func (s Sample) Mul(v Sample) Sample { return Sample(float32(s) * float32(v)) }

func (s Sample) Clamp(min, max Sample) Sample {
	if s < min {
		return min
	}
	if s > max {
		return max
	}
	return s
}

type SampleRate uint32
