package audio

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/gordonklaus/portaudio"
	"github.com/pkg/errors"
)

var (
	ErrShortBuffer = errors.New("buffer too short")
)

var (
	DefaultRecorder *Recorder
)

type Reader struct {
	stream *portaudio.Stream
	data   []int16
}

func NewReader(rate, channel, frames int) (*Reader, error) {
	r := &Reader{
		data: make([]int16, frames),
	}
	stream, err := portaudio.OpenDefaultStream(channel, 0, float64(rate), len(r.data), r.data)
	if err != nil {
		return nil, fmt.Errorf("Error open default audio stream: %s", err)
	}
	err = stream.Start()
	if err != nil {
		return nil, fmt.Errorf("Error on stream start: %s", err)
	}

	r.stream = stream
	return r, nil
}

func (r *Reader) Read(buf []byte) (int, error) {
	if len(buf) < len(r.data)*2 {
		return 0, ErrShortBuffer
	}
	err := r.stream.Read()
	if err != nil && err != portaudio.InputOverflowed {
		return 0, errors.Wrap(err, "stream.Read")
	}
	for i := 0; i < len(r.data); i++ {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(r.data[i]))
	}
	return len(r.data) * 2, nil
}

func (r *Reader) Close() error {
	return r.stream.Close()
}

func init() {
	err := portaudio.Initialize()
	if err != nil {
		log.Fatalf("Error initialize audio interface: %s", err)
	}
	DefaultRecorder, err = NewRecorder(16000, 1)
	if err != nil {
		log.Fatalf("Error initialize default recorder: %s", err)
	}
}
