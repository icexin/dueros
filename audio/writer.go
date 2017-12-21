package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gordonklaus/portaudio"
)

func getOutDeviceByName(name string) *portaudio.DeviceInfo {
	devices, err := portaudio.Devices()
	if err != nil {
		log.Print(err)
		return nil
	}
	for _, device := range devices {
		if device.MaxOutputChannels == 0 {
			continue
		}
		if device.Name == name {
			return device
		}
	}
	return nil
}

type Writer struct {
	stream *portaudio.Stream

	rate, channel int

	buf []int16
	pos int32

	mutex sync.Mutex
	cond  *sync.Cond
	done  bool

	paused bool
	closed bool
}

func NewWriter(rate, channel int, buffer []byte) (*Writer, error) {
	w := &Writer{
		rate:    rate,
		channel: channel,
		buf:     make([]int16, len(buffer)/2),
	}
	w.cond = sync.NewCond(&w.mutex)

	var device *portaudio.DeviceInfo
	var err error
	err = binary.Read(bytes.NewBuffer(buffer), binary.LittleEndian, w.buf)
	if err != nil {
		panic(err)
	}
	name := os.Getenv("DUEROS_OUT")
	if name != "" {
		device = getOutDeviceByName(name)
		if device == nil {
			return nil, errors.New("bad out device name")
		}
	} else {
		device, err = portaudio.DefaultOutputDevice()
		if err != nil {
			return nil, err
		}
	}
	param := portaudio.LowLatencyParameters(nil, device)
	param.SampleRate = float64(rate)
	param.Output.Channels = channel
	param.FramesPerBuffer = portaudio.FramesPerBufferUnspecified

	stream, err := portaudio.OpenStream(param, w.callback)
	if err != nil {
		return nil, fmt.Errorf("Error open default audio stream: %s", err)
	}
	w.stream = stream
	return w, nil
}

func (w *Writer) callback(out []int16) {
	if int(w.pos) == len(w.buf) {
		return
	}
	var i = 0
	for i = 0; i < len(out) && i+int(w.pos) < len(w.buf); i++ {
		out[i] = w.buf[int(w.pos)+i]
	}
	atomic.AddInt32(&w.pos, int32(i))
	if int(w.pos) == len(w.buf) {
		go w.playDone()
	}
}

func (w *Writer) playDone() {
	w.mutex.Lock()
	w.done = true
	w.mutex.Unlock()
	w.cond.Broadcast()
}

func (w *Writer) Len() time.Duration {
	frames := len(w.buf) / w.channel
	return time.Duration(frames*1000/w.rate) * time.Millisecond
}

func (w *Writer) SetOffset(offset time.Duration) {
	if w.closed {
		return
	}
	length := w.Len()
	n := int32(offset/length) * int32(len(w.buf))
	if n < int32(len(w.buf)) {
		atomic.StoreInt32(&w.pos, n)
	}
}

func (w *Writer) Offset() time.Duration {
	frames := int(atomic.LoadInt32(&w.pos)) / w.channel
	return time.Duration(frames*1000/w.rate) * time.Millisecond
}

func (w *Writer) Play() error {
	err := w.Start()
	if err != nil {
		return err
	}
	w.Wait()
	return nil
}

func (w *Writer) Start() error {
	if w.closed {
		return errors.New("closed")
	}
	err := w.stream.Start()
	if err != nil {
		return fmt.Errorf("Error on stream start: %s", err)
	}
	return nil
}

func (w *Writer) Wait() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for !w.done {
		w.cond.Wait()
	}
}

func (w *Writer) Pause() {
	if w.paused {
		return
	}
	w.paused = true
	w.stream.Stop()
}

func (w *Writer) Resume() {
	if !w.paused || w.closed {
		return
	}
	w.paused = false
	w.stream.Start()
}

func (w *Writer) Closed() bool {
	return w.closed
}

func (w *Writer) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	w.playDone()
	return w.stream.Close()
}
