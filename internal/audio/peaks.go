package audio

import "math"

// Peaks reduces the clip to n loudness values in [0, 1], normalized so
// the loudest bucket is 1 - the bars of a waveform display.
func (c *Clip) Peaks(n int) []float32 {
	if c == nil || len(c.Samples) == 0 || n <= 0 {
		return nil
	}
	out := make([]float32, n)
	var top float32
	for i := range out {
		lo := i * len(c.Samples) / n
		hi := min(max((i+1)*len(c.Samples)/n, lo+1), len(c.Samples))
		var sumSq float64
		var peak float32
		for _, s := range c.Samples[lo:hi] {
			peak = max(peak, float32(math.Abs(float64(s))))
			sumSq += float64(s) * float64(s)
		}
		// Blend RMS and peak: pure peaks look spiky, pure RMS flat.
		rms := float32(math.Sqrt(sumSq / float64(hi-lo)))
		out[i] = 0.6*rms + 0.4*peak
		top = max(top, out[i])
	}
	if top > 0 {
		for i := range out {
			out[i] /= top
		}
	}
	return out
}
