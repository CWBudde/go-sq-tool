package wav

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

// AudioData represents multi-channel audio data
type AudioData struct {
	SampleRate uint32
	Samples    [][]float64 // [channel][sample]
	NumSamples int
}

// ReadWAV reads a stereo WAV file and returns the audio data
func ReadWAV(filename string) (*AudioData, error) {
	return ReadWAVChannels(filename, 2)
}

// ReadWAVFromReader reads a WAV stream with a specific channel count.
func ReadWAVFromReader(r io.Reader, channels int) (*AudioData, error) {
	audioData, err := readWAV(r, channels)
	if err != nil {
		return nil, fmt.Errorf("failed to read WAV: %w", err)
	}
	return audioData, nil
}

// ReadWAVBytes reads a WAV payload with a specific channel count.
func ReadWAVBytes(data []byte, channels int) (*AudioData, error) {
	return ReadWAVFromReader(bytes.NewReader(data), channels)
}

// ReadWAVChannels reads a WAV file with a specific channel count
func ReadWAVChannels(filename string, channels int) (*AudioData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAV file: %w", err)
	}
	defer file.Close()

	return ReadWAVFromReader(file, channels)
}

// WriteWAV writes 4-channel audio data to a WAV file
func WriteWAV(filename string, data *AudioData) error {
	return writeWAVPCM16(filename, data, 4)
}

// WriteStereoWAV writes 2-channel audio data to a WAV file
func WriteStereoWAV(filename string, data *AudioData) error {
	return writeWAVPCM16(filename, data, 2)
}

func writeWAVPCM16(filename string, data *AudioData, channels int) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create WAV file: %w", err)
	}
	defer file.Close()

	return writeWAVPCM16ToWriter(file, data, channels)
}

// WriteWAVToWriter writes 4-channel audio data to a WAV stream in 16-bit PCM.
func WriteWAVToWriter(w io.Writer, data *AudioData) error {
	return writeWAVPCM16ToWriter(w, data, 4)
}

// WriteStereoWAVToWriter writes 2-channel audio data to a WAV stream in 16-bit PCM.
func WriteStereoWAVToWriter(w io.Writer, data *AudioData) error {
	return writeWAVPCM16ToWriter(w, data, 2)
}

func writeWAVPCM16ToWriter(w io.Writer, data *AudioData, channels int) error {
	if len(data.Samples) != channels {
		return fmt.Errorf("output must have %d channels, got %d", channels, len(data.Samples))
	}
	if data.NumSamples < 0 {
		return fmt.Errorf("NumSamples must be >= 0")
	}
	for ch := 0; ch < channels; ch++ {
		if len(data.Samples[ch]) < data.NumSamples {
			return fmt.Errorf("channel %d has %d samples, want at least %d", ch, len(data.Samples[ch]), data.NumSamples)
		}
	}

	bw := bufio.NewWriter(w)

	numChannels := uint16(channels)
	bitsPerSample := uint16(16)
	blockAlign := numChannels * (bitsPerSample / 8)
	byteRate := data.SampleRate * uint32(blockAlign)
	audioFormat := uint16(1) // PCM
	dataSize := uint32(data.NumSamples) * uint32(blockAlign)

	// RIFF header
	if err := writeString(bw, "RIFF"); err != nil {
		return fmt.Errorf("failed to write RIFF header: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, uint32(36+dataSize)); err != nil {
		return fmt.Errorf("failed to write file size: %w", err)
	}
	if err := writeString(bw, "WAVE"); err != nil {
		return fmt.Errorf("failed to write WAVE header: %w", err)
	}

	// fmt chunk
	if err := writeString(bw, "fmt "); err != nil {
		return fmt.Errorf("failed to write fmt chunk ID: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, uint32(16)); err != nil {
		return fmt.Errorf("failed to write fmt chunk size: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, audioFormat); err != nil {
		return fmt.Errorf("failed to write audio format: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, numChannels); err != nil {
		return fmt.Errorf("failed to write num channels: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, data.SampleRate); err != nil {
		return fmt.Errorf("failed to write sample rate: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, byteRate); err != nil {
		return fmt.Errorf("failed to write byte rate: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, blockAlign); err != nil {
		return fmt.Errorf("failed to write block align: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, bitsPerSample); err != nil {
		return fmt.Errorf("failed to write bits per sample: %w", err)
	}

	// data chunk
	if err := writeString(bw, "data"); err != nil {
		return fmt.Errorf("failed to write data chunk ID: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, dataSize); err != nil {
		return fmt.Errorf("failed to write data size: %w", err)
	}

	// Interleaved PCM16 samples
	for i := 0; i < data.NumSamples; i++ {
		for ch := 0; ch < channels; ch++ {
			sample := floatToPCM16(data.Samples[ch][i])
			if err := binary.Write(bw, binary.LittleEndian, sample); err != nil {
				return fmt.Errorf("failed to write sample data: %w", err)
			}
		}
	}
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("failed to flush WAV data: %w", err)
	}

	return nil
}

// WriteFloat32WAV writes 4-channel audio data to a WAV file in 32-bit IEEE float format
func WriteFloat32WAV(filename string, data *AudioData) error {
	return writeWAVFloat32(filename, data, 4)
}

// WriteStereoFloat32WAV writes 2-channel audio data to a WAV file in 32-bit IEEE float format
func WriteStereoFloat32WAV(filename string, data *AudioData) error {
	return writeWAVFloat32(filename, data, 2)
}

func writeWAVFloat32(filename string, data *AudioData, channels int) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create WAV file: %w", err)
	}
	defer file.Close()

	return writeWAVFloat32ToWriter(file, data, channels)
}

// WriteFloat32WAVToWriter writes 4-channel audio data to a WAV stream in 32-bit IEEE float format.
func WriteFloat32WAVToWriter(w io.Writer, data *AudioData) error {
	return writeWAVFloat32ToWriter(w, data, 4)
}

// WriteStereoFloat32WAVToWriter writes 2-channel audio data to a WAV stream in 32-bit IEEE float format.
func WriteStereoFloat32WAVToWriter(w io.Writer, data *AudioData) error {
	return writeWAVFloat32ToWriter(w, data, 2)
}

func writeWAVFloat32ToWriter(w io.Writer, data *AudioData, channels int) error {
	if len(data.Samples) != channels {
		return fmt.Errorf("output must have %d channels, got %d", channels, len(data.Samples))
	}
	if data.NumSamples < 0 {
		return fmt.Errorf("NumSamples must be >= 0")
	}
	for ch := 0; ch < channels; ch++ {
		if len(data.Samples[ch]) < data.NumSamples {
			return fmt.Errorf("channel %d has %d samples, want at least %d", ch, len(data.Samples[ch]), data.NumSamples)
		}
	}

	bw := bufio.NewWriter(w)

	numChannels := uint16(channels)
	bitsPerSample := uint16(32)
	byteRate := data.SampleRate * uint32(numChannels) * uint32(bitsPerSample/8)
	blockAlign := numChannels * (bitsPerSample / 8)
	audioFormat := uint16(3) // IEEE float
	dataSize := uint32(data.NumSamples) * uint32(numChannels) * uint32(bitsPerSample/8)

	// Write RIFF header
	if err := writeString(bw, "RIFF"); err != nil {
		return fmt.Errorf("failed to write RIFF header: %w", err)
	}
	// File size - 8 (will be updated at the end if needed)
	if err := binary.Write(bw, binary.LittleEndian, uint32(36+dataSize)); err != nil {
		return fmt.Errorf("failed to write file size: %w", err)
	}
	if err := writeString(bw, "WAVE"); err != nil {
		return fmt.Errorf("failed to write WAVE header: %w", err)
	}

	// Write fmt chunk
	if err := writeString(bw, "fmt "); err != nil {
		return fmt.Errorf("failed to write fmt chunk ID: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, uint32(16)); err != nil { // fmt chunk size
		return fmt.Errorf("failed to write fmt chunk size: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, audioFormat); err != nil {
		return fmt.Errorf("failed to write audio format: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, numChannels); err != nil {
		return fmt.Errorf("failed to write num channels: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, data.SampleRate); err != nil {
		return fmt.Errorf("failed to write sample rate: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, byteRate); err != nil {
		return fmt.Errorf("failed to write byte rate: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, blockAlign); err != nil {
		return fmt.Errorf("failed to write block align: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, bitsPerSample); err != nil {
		return fmt.Errorf("failed to write bits per sample: %w", err)
	}

	// Write data chunk
	if err := writeString(bw, "data"); err != nil {
		return fmt.Errorf("failed to write data chunk ID: %w", err)
	}
	if err := binary.Write(bw, binary.LittleEndian, dataSize); err != nil {
		return fmt.Errorf("failed to write data size: %w", err)
	}

	// Write interleaved float32 samples
	for i := 0; i < data.NumSamples; i++ {
		for ch := 0; ch < channels; ch++ {
			val := data.Samples[ch][i]
			// Clamp to [-1.0, 1.0] to prevent invalid float values
			if val > 1.0 {
				val = 1.0
			} else if val < -1.0 {
				val = -1.0
			} else if math.IsNaN(val) || math.IsInf(val, 0) {
				val = 0.0
			}

			if err := binary.Write(bw, binary.LittleEndian, float32(val)); err != nil {
				return fmt.Errorf("failed to write sample data: %w", err)
			}
		}
	}

	if err := bw.Flush(); err != nil {
		return fmt.Errorf("failed to flush WAV data: %w", err)
	}

	return nil
}

// writeString writes a string to the writer without a null terminator
func writeString(w io.Writer, s string) error {
	_, err := w.Write([]byte(s))
	return err
}

type wavFormat struct {
	audioFormat   uint16
	numChannels   uint16
	sampleRate    uint32
	byteRate      uint32
	blockAlign    uint16
	bitsPerSample uint16
}

func readWAV(r io.Reader, expectedChannels int) (*AudioData, error) {
	br := bufio.NewReader(r)

	var riff [4]byte
	if _, err := io.ReadFull(br, riff[:]); err != nil {
		return nil, fmt.Errorf("read RIFF header: %w", err)
	}
	if string(riff[:]) != "RIFF" {
		return nil, fmt.Errorf("not a RIFF file")
	}

	var _riffSize uint32
	if err := binary.Read(br, binary.LittleEndian, &_riffSize); err != nil {
		return nil, fmt.Errorf("read RIFF size: %w", err)
	}

	var wave [4]byte
	if _, err := io.ReadFull(br, wave[:]); err != nil {
		return nil, fmt.Errorf("read WAVE header: %w", err)
	}
	if string(wave[:]) != "WAVE" {
		return nil, fmt.Errorf("not a WAVE file")
	}

	var fmtChunk *wavFormat
	for {
		var chunkID [4]byte
		if _, err := io.ReadFull(br, chunkID[:]); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read chunk id: %w", err)
		}
		var chunkSize uint32
		if err := binary.Read(br, binary.LittleEndian, &chunkSize); err != nil {
			return nil, fmt.Errorf("read chunk size: %w", err)
		}

		switch string(chunkID[:]) {
		case "fmt ":
			if chunkSize < 16 {
				return nil, fmt.Errorf("invalid fmt chunk size %d", chunkSize)
			}
			f := &wavFormat{}
			if err := binary.Read(br, binary.LittleEndian, &f.audioFormat); err != nil {
				return nil, fmt.Errorf("read audio format: %w", err)
			}
			if err := binary.Read(br, binary.LittleEndian, &f.numChannels); err != nil {
				return nil, fmt.Errorf("read num channels: %w", err)
			}
			if err := binary.Read(br, binary.LittleEndian, &f.sampleRate); err != nil {
				return nil, fmt.Errorf("read sample rate: %w", err)
			}
			if err := binary.Read(br, binary.LittleEndian, &f.byteRate); err != nil {
				return nil, fmt.Errorf("read byte rate: %w", err)
			}
			if err := binary.Read(br, binary.LittleEndian, &f.blockAlign); err != nil {
				return nil, fmt.Errorf("read block align: %w", err)
			}
			if err := binary.Read(br, binary.LittleEndian, &f.bitsPerSample); err != nil {
				return nil, fmt.Errorf("read bits per sample: %w", err)
			}

			remaining := int64(chunkSize) - 16
			if remaining > 0 {
				if _, err := io.CopyN(io.Discard, br, remaining); err != nil {
					return nil, fmt.Errorf("skip fmt extension: %w", err)
				}
			}

			fmtChunk = f

		case "data":
			if fmtChunk == nil {
				return nil, fmt.Errorf("data chunk before fmt chunk")
			}
			if int(fmtChunk.numChannels) != expectedChannels {
				return nil, fmt.Errorf("input must have %d channels, got %d channels", expectedChannels, fmtChunk.numChannels)
			}
			if fmtChunk.blockAlign == 0 {
				return nil, fmt.Errorf("invalid blockAlign=0")
			}
			if chunkSize%uint32(fmtChunk.blockAlign) != 0 {
				return nil, fmt.Errorf("data chunk not aligned to block size")
			}

			numFrames := int(chunkSize / uint32(fmtChunk.blockAlign))
			samplesByChannel := make([][]float64, expectedChannels)
			for ch := 0; ch < expectedChannels; ch++ {
				samplesByChannel[ch] = make([]float64, numFrames)
			}

			switch fmtChunk.audioFormat {
			case 1: // PCM
				switch fmtChunk.bitsPerSample {
				case 16:
					for i := range numFrames {
						for ch := range expectedChannels {
							var v int16
							if err := binary.Read(br, binary.LittleEndian, &v); err != nil {
								return nil, fmt.Errorf("read PCM16 sample: %w", err)
							}
							samplesByChannel[ch][i] = float64(v) / 32768.0
						}
					}
				case 24:
					for i := range numFrames {
						for ch := range expectedChannels {
							v, err := readPCM24Sample(br)
							if err != nil {
								return nil, fmt.Errorf("read PCM24 sample: %w", err)
							}
							samplesByChannel[ch][i] = float64(v) / 8388608.0
						}
					}
				default:
					return nil, fmt.Errorf("unsupported PCM bit depth %d", fmtChunk.bitsPerSample)
				}

			case 3: // IEEE float
				if fmtChunk.bitsPerSample != 32 {
					return nil, fmt.Errorf("unsupported IEEE float bit depth %d", fmtChunk.bitsPerSample)
				}
				for i := 0; i < numFrames; i++ {
					for ch := 0; ch < expectedChannels; ch++ {
						var v float32
						if err := binary.Read(br, binary.LittleEndian, &v); err != nil {
							return nil, fmt.Errorf("read float32 sample: %w", err)
						}
						fv := float64(v)
						if math.IsNaN(fv) || math.IsInf(fv, 0) {
							fv = 0
						}
						if fv > 1.0 {
							fv = 1.0
						} else if fv < -1.0 {
							fv = -1.0
						}
						samplesByChannel[ch][i] = fv
					}
				}

			default:
				return nil, fmt.Errorf("unsupported WAV audio format %d", fmtChunk.audioFormat)
			}

			// Chunks are word-aligned; if size is odd, a pad byte follows.
			if chunkSize%2 == 1 {
				if _, err := br.ReadByte(); err != nil {
					return nil, fmt.Errorf("read data pad byte: %w", err)
				}
			}

			return &AudioData{
				SampleRate: fmtChunk.sampleRate,
				Samples:    samplesByChannel,
				NumSamples: numFrames,
			}, nil

		default:
			// Skip unknown chunk (plus pad byte if needed)
			if _, err := io.CopyN(io.Discard, br, int64(chunkSize)); err != nil {
				return nil, fmt.Errorf("skip chunk %q: %w", string(chunkID[:]), err)
			}
			if chunkSize%2 == 1 {
				if _, err := br.ReadByte(); err != nil {
					return nil, fmt.Errorf("read pad byte: %w", err)
				}
			}
		}
	}

	return nil, fmt.Errorf("no data chunk found")
}

func floatToPCM16(v float64) int16 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		v = 0
	}
	if v >= 1.0 {
		return 32767
	}
	if v <= -1.0 {
		return -32768
	}
	return int16(math.Round(v * 32767.0))
}

func readPCM24Sample(r io.Reader) (int32, error) {
	var b [3]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	v := int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16
	if v&0x800000 != 0 {
		v |= ^0xffffff
	}
	return v, nil
}
