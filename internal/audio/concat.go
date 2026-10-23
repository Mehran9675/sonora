package audio

// Concat joins clips of the same sample rate, separated by gap seconds
// of silence.
func Concat(clips []*Clip, gap float64) *Clip {
	if len(clips) == 0 {
		return &Clip{SampleRate: 24000}
	}
	out := &Clip{SampleRate: clips[0].SampleRate}
	silence := make([]float32, int(gap*float64(out.SampleRate)))
	for i, c := range clips {
		if i > 0 {
			out.Samples = append(out.Samples, silence...)
		}
		out.Samples = append(out.Samples, c.Resample(out.SampleRate)...)
	}
	return out
}
