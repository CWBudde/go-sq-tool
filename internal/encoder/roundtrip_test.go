package encoder_test

import (
	"math"
	"testing"

	"github.com/cwbudde/go-sq-decoder/internal/decoder"
	"github.com/cwbudde/go-sq-decoder/internal/encoder"
)

func TestEncodeDecodeRoundTrip_FrontChannels(t *testing.T) {
	t.Parallel()

	const (
		blockSize = 1024
		overlap   = 512
		n         = 10 * overlap
	)

	lf := make([]float64, n)
	rf := make([]float64, n)
	for i := 0; i < n; i++ {
		lf[i] = 0.6 * math.Sin(2.0*math.Pi*float64(i)/97.0)
		rf[i] = 0.4 * math.Cos(2.0*math.Pi*float64(i)/131.0)
	}

	quad := [][]float64{
		lf,
		rf,
		make([]float64, n),
		make([]float64, n),
	}

	sqEnc := encoder.NewSQEncoderWithParams(blockSize, overlap)
	sqStereo, err := sqEnc.Process(quad)
	if err != nil {
		t.Fatalf("encoder.Process() error = %v", err)
	}
	if got := len(sqStereo); got != 2 {
		t.Fatalf("encoded channels = %d, want 2", got)
	}

	sqDec := decoder.NewSQDecoderWithParams(blockSize, overlap)
	decoded, err := sqDec.Process(sqStereo)
	if err != nil {
		t.Fatalf("decoder.Process() error = %v", err)
	}
	if got := len(decoded); got != 4 {
		t.Fatalf("decoded channels = %d, want 4", got)
	}

	// Both encoder and decoder use inputOffset=overlap/4 when mapping samples.
	// For front-only content, this effectively results in an overall shift of overlap/2.
	shift := overlap / 2
	const tol = 1e-12

	for i := 0; i < n-shift; i++ {
		wantLF := lf[i+shift]
		wantRF := rf[i+shift]

		if math.Abs(decoded[0][i]-wantLF) > tol {
			t.Fatalf("LF[%d] = %.15f, want %.15f", i, decoded[0][i], wantLF)
		}
		if math.Abs(decoded[1][i]-wantRF) > tol {
			t.Fatalf("RF[%d] = %.15f, want %.15f", i, decoded[1][i], wantRF)
		}
	}
}
