package audio

import (
	"io"
	"sync"
)

type streamReader struct {
	r      *Recorder
	closed bool
}

func newStreamReader(r *Recorder) *streamReader {
	return &streamReader{
		r: r,
	}
}

func (s *streamReader) Read(b []byte) (int, error) {
	if s.closed {
		return 0, io.EOF
	}

	buf := b
	for {
		n, err := s.r.read(buf)
		if err == ErrShortBuffer {
			break
		}
		if err != nil {
			return 0, err
		}
		buf = buf[n:]
	}
	n := len(b) - len(buf)
	if n == 0 {
		return 0, ErrShortBuffer
	}
	return n, nil
}

func (s *streamReader) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	s.r.onStreamClosed()
	return nil
}

type Recorder struct {
	r *Reader

	mutex sync.Mutex
}

func NewRecorder(rate, channel int) (*Recorder, error) {
	r, err := NewReader(rate, channel, 160)
	if err != nil {
		return nil, err
	}

	return &Recorder{
		r: r,
	}, nil
}

func (r *Recorder) NewStream() io.ReadCloser {
	r.mutex.Lock()
	return newStreamReader(r)
}

func (r *Recorder) read(b []byte) (int, error) {
	return r.r.Read(b)
}

func (r *Recorder) onStreamClosed() {
	r.mutex.Unlock()
}

func NewRecordStream() io.ReadCloser {
	return DefaultRecorder.NewStream()
}
