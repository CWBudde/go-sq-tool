package wav

import (
	"math"
	"path/filepath"
	"testing"
)

func TestReadWAVChannels_StereoRoundTrip(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "stereo.wav")

	in := &AudioData{
		SampleRate: 44100,
		Samples: [][]float64{
			{0.0, 0.5, -0.5, 1.0, -1.0, 0.25, -0.25},
			{0.1, -0.1, 0.9, -0.9, 0.0, 0.75, -0.75},
		},
		NumSamples: 7,
	}

	if err := WriteStereoWAV(filename, in); err != nil {
		t.Fatalf("WriteStereoWAV() error = %v", err)
	}

	out, err := ReadWAVChannels(filename, 2)
	if err != nil {
		t.Fatalf("ReadWAVChannels() error = %v", err)
	}

	if out.SampleRate != in.SampleRate {
		t.Fatalf("SampleRate = %d, want %d", out.SampleRate, in.SampleRate)
	}
	if out.NumSamples != in.NumSamples {
		t.Fatalf("NumSamples = %d, want %d", out.NumSamples, in.NumSamples)
	}
	if got := len(out.Samples); got != 2 {
		t.Fatalf("len(Samples) = %d, want 2", got)
	}

	const tol = 2.0 / 32767.0 // allow ~2 LSB error for float conversion / rounding
	for ch := 0; ch < 2; ch++ {
		for i := 0; i < in.NumSamples; i++ {
			got := out.Samples[ch][i]
			want := in.Samples[ch][i]
			if math.Abs(got-want) > tol {
				t.Fatalf("sample[%d][%d] = %.8f, want %.8f (tol %.8f)", ch, i, got, want, tol)
			}
		}
	}
}

func TestReadWAVChannels_ChannelMismatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "stereo.wav")

	data := &AudioData{
		SampleRate: 44100,
		Samples: [][]float64{
			{0.0, 0.0, 0.0, 0.0},
			{0.0, 0.0, 0.0, 0.0},
		},
		NumSamples: 4,
	}

	if err := WriteStereoWAV(filename, data); err != nil {
		t.Fatalf("WriteStereoWAV() error = %v", err)
	}

	if _, err := ReadWAVChannels(filename, 4); err == nil {
		t.Fatalf("ReadWAVChannels() expected error, got nil")
	}
}
