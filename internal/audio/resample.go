package audio

// Resample returns the samples converted to rate with linear
// interpolation, which is plenty for speech.
func (c *Clip) Resample(rate int) []float32 {
	if c.SampleRate == rate || len(c.Samples) == 0 {
		return c.Samples
	}
	ratio := float64(c.SampleRate) / float64(rate)
	out := make([]float32, int(float64(len(c.Samples))/ratio))
	last := len(c.Samples) - 1
	for i := range out {
		pos := float64(i) * ratio
		j := int(pos)
		if j >= last {
			out[i] = c.Samples[last]
			continue
		}
		f := float32(pos - float64(j))
		out[i] = c.Samples[j]*(1-f) + c.Samples[j+1]*f
	}
	return out
}
