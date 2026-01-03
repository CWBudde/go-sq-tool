# SQ Quadrophonic Encoder/Decoder

A high-quality SQ (Stereo Quadrophonic) encoder/decoder written in Go, implementing the FFT-based SQ² algorithm for accurate 90° phase shifting and superior channel separation.

## Overview

**SQ (Stereo Quadraphonic)** is a matrix encoding system developed by CBS in the 1970s for encoding 4-channel quadrophonic audio into a 2-channel stereo signal. This tool can decode SQ-encoded stereo into four channels and encode four-channel quad audio into SQ-compatible stereo.

### Features

- ✅ **FFT-based Hilbert Transform**: Accurate 90° phase shift across all frequencies
- ✅ **High-quality decoding**: Good channel separation using frequency-domain processing
- ✅ **SQ encoding**: Convert quad audio into SQ-compatible stereo
- ✅ **Simple CLI interface**: Easy to use command-line tool
- ✅ **WAV file support**: Standard WAV file I/O for compatibility
- ✅ **Configurable parameters**: Adjustable block size and overlap for quality/performance tuning

## Algorithm

This implementation is based on the **SQ² decoder** algorithm which uses:

1. **FFT-based Hilbert Transform** for precise 90° phase shifting
2. **SQ Decode Matrix**:
   - LF (Left Front) = LT (direct output)
   - RF (Right Front) = RT (direct output)
   - LB (Left Back) = 0.707 × H(LT) - 0.707 × RT
   - RB (Right Back) = 0.707 × LT - 0.707 × H(RT)

Where `H()` denotes the Hilbert transform (90° phase shift) and 0.707 ≈ √2/2.
Note: LT/RT are mixes of all four channels in the SQ encoder, so LF/RF here are direct LT/RT outputs and therefore include rear-channel crosstalk. This is a basic matrix decoder without logic steering.

See [`sq-decoder-explained.md`](./sq-decoder-explained.md) for comprehensive technical documentation.

## Installation

### Prerequisites

- Go 1.21 or later

### Build from source

```bash
git clone https://github.com/cwbudde/go-sq-tool.git
cd go-sq-tool
go mod download
go build -o go-sq-tool
```

## Usage

### Basic Usage (Decode)

```bash
go-sq-tool input.wav output.wav
```

**Input**: 2-channel stereo WAV file (SQ-encoded)
**Output**: 4-channel quadrophonic WAV file (LF, RF, LB, RB)

### Decode (Explicit)

```bash
go-sq-tool decode input.wav output.wav
```

### Encode (Quad to SQ Stereo)

```bash
go-sq-tool encode quad_input.wav sq_output.wav
```

**Input**: 4-channel quadrophonic WAV file (LF, RF, LB, RB)
**Output**: 2-channel stereo WAV file (LT, RT)

### Verbose Output

```bash
go-sq-tool -v input.wav output.wav
```

Shows detailed information about processing:

- Input file properties (sample rate, duration)
- Decoder configuration (block size, latency)
- Processing status

### Custom Parameters

```bash
go-sq-tool -b 2048 -o 1024 input.wav output.wav
```

- `-b, --block-size`: FFT block size (default: 1024, must be power of 2)
- `-o, --overlap`: Overlap in samples (default: 512, typically blockSize/2)
- `--logic`: Enable CBS-style logic steering for improved separation (adds dynamic steering)

### Analyze Channel Separation

```bash
go-sq-tool analyze quad_input.wav
```

This runs an encode -> decode loop on isolated channels from a 4-channel input and reports RMS-based separation in dB. Results depend on program material and decoder settings (including `--logic`).

Optional analysis flags:

- `--leak-mode` (`max` or `avg`): how to aggregate leakage across non-target channels
- `--fmin`, `--fmax`: band-limit the RMS computation (Hz)
- `--pair-mode` (`isolated` or `full`): compute pair separation using isolated channels or the full mix

### Generate Test File

```bash
go-sq-tool generate-test test_quad.wav
```

Generates a 4-channel WAV with tones at 100/200/400/800 Hz (LF/RF/LB/RB) plus low-level white noise for quick separation checks.

### Help

```bash
go-sq-tool --help
```

## WASM Demo

A simple browser demo lives in `web/` and runs the decoder entirely client-side.

```bash
if [ -f "$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/; else cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/; fi
GOOS=js GOARCH=wasm go build -buildvcs=false -o web/sqdecoder.wasm .
python3 -m http.server --directory web 8080
```

Open `http://localhost:8080` and drop a stereo WAV to decode to quad.

## Examples

### Decode SQ record to quadrophonic

```bash
# Basic decode
go-sq-tool my_sq_recording.wav quad_output.wav

# Verbose mode to see processing details
go-sq-tool -v my_sq_recording.wav quad_output.wav

# Higher quality (larger FFT, more latency)
go-sq-tool -b 2048 -o 1024 input.wav output.wav
```

### Encode quad to SQ stereo

```bash
# Encode 4-channel WAV to SQ stereo
go-sq-tool encode quad_input.wav sq_output.wav

# Float output for headroom
go-sq-tool encode --float32 quad_input.wav sq_output.wav
```

## Technical Details

### Decoder Characteristics

| Parameter              | Value                           |
| ---------------------- | ------------------------------- |
| **Algorithm**          | FFT-based SQ²                   |
| **Default Block Size** | 1024 samples                    |
| **Default Overlap**    | 512 samples (50%)               |
| **Latency**            | 768 samples (~17.4ms @ 44.1kHz) |
| **Phase Shift Method** | Hilbert transform via FFT       |
| **Input Channels**     | 2 (stereo)                      |
| **Output Channels**    | 4 (quadrophonic)                |

### Encoder Characteristics

| Parameter              | Value                           |
| ---------------------- | ------------------------------- |
| **Algorithm**          | FFT-based SQ                    |
| **Default Block Size** | 1024 samples                    |
| **Default Overlap**    | 512 samples (50%)               |
| **Latency**            | 768 samples (~17.4ms @ 44.1kHz) |
| **Phase Shift Method** | Hilbert transform via FFT       |
| **Input Channels**     | 4 (quadrophonic)                |
| **Output Channels**    | 2 (stereo)                      |

### Channel Layout

**Input (SQ-encoded stereo)**:

- Channel 0: LT (Left Total)
- Channel 1: RT (Right Total)

**Output (Quadrophonic)**:

- Channel 0: LF (Left Front)
- Channel 1: RF (Right Front)
- Channel 2: LB (Left Back)
- Channel 3: RB (Right Back)

**Output (SQ-encoded stereo)**:

- Channel 0: LT (Left Total)
- Channel 1: RT (Right Total)

### Dependencies

- [`github.com/MeKo-Christian/algo-fft`](https://github.com/MeKo-Christian/algo-fft) - FFT implementation
- [`github.com/spf13/cobra`](https://github.com/spf13/cobra) - CLI framework
- [`github.com/youpy/go-wav`](https://github.com/youpy/go-wav) - WAV file I/O

## Performance

The FFT-based SQ² decoder provides:

**Advantages**:

- ✅ Accurate 90° phase shift across audio spectrum
- ✅ Predictable frequency response
- ✅ Superior quality for archival restoration

**Tradeoffs**:

- ⏱️ Higher latency (~17ms) vs. recursive filter approach
- 💻 Moderate CPU usage (FFT operations)
- 📊 Block-based processing

For real-time applications requiring minimal latency, consider implementing the simpler recursive filter variant (not included in this tool).

## Separation Measurements

Channel separation is reported by the built-in analyzer, which runs an encode -> decode loop and measures RMS leakage. The numbers are content-dependent, so use the analyzer on your own material (or the generated test file) instead of relying on estimates.

Example workflow:

```bash
go-sq-tool generate-test ./assets/temp_quad.wav --noise-level 0
go-sq-tool analyze ./assets/temp_quad.wav --pair-mode full
```

Example output (basic matrix decode, full mix, no noise):

```
Channel  TargetRMS   LeakRMS  Sep(dB)
LF       0.424032  0.299836    3.01
RF       0.424019  0.299827    3.01
LB       0.211999  0.299812   -3.01
RB       0.212004  0.299819   -3.01

Pair separation (dB)
LF->RF: -0.00  RF->LF: 0.00  LB->RB: 0.00  RB->LB: -0.00
```

Tips:

- Add `--logic` to measure CBS-style logic steering behavior.
- Use `--leak-mode avg` for average leakage instead of max-leak.
- Use `--fmin`/`--fmax` for band-limited separation.

## Contributing

Contributions welcome! Please feel free to submit issues or pull requests.

### Areas for Enhancement

- [ ] Add basic recursive filter decoder (low-latency variant)
- [ ] Support for other audio formats (FLAC, MP3, etc.)
- [ ] Real-time processing mode
- [ ] GUI frontend
- [ ] Batch processing

## References

- **CBS SQ System**: Original matrix quadraphonic system from 1970s
- **Hilbert Transform**: Linear operator for 90° phase shifting
- **FFT**: Fast Fourier Transform for frequency-domain processing

See [`sq-decoder-explained.md`](./sq-decoder-explained.md) for detailed technical documentation including:

- Complete algorithm explanation
- Mathematical foundations
- Historical context
- Implementation comparisons

## License

MIT License - feel free to use for any purpose.

---

**Author**: Christian Budde (cwbudde)
**Version**: 1.0.0
**Date**: 2026-01-02
