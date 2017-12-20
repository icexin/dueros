package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gordonklaus/portaudio"
)

type Writer struct {
	stream *portaudio.Stream
	buf    *bytes.Buffer
	done   chan bool
}

func NewWriter(device *portaudio.DeviceInfo, rate, channel int, buffer []byte) (*Writer, error) {
	w := &Writer{
		buf:  bytes.NewBuffer(buffer),
		done: make(chan bool, 1),
	}
	param := portaudio.LowLatencyParameters(nil, device)
	param.SampleRate = float64(rate)
	param.Output.Channels = channel
	// stream, err := portaudio.OpenDefaultStream(0, channel, float64(rate), portaudio.FramesPerBufferUnspecified, w.callback)
	stream, err := portaudio.OpenStream(param, w.callback)
	if err != nil {
		return nil, fmt.Errorf("Error open default audio stream: %s", err)
	}
	w.stream = stream
	return w, nil
}

func (w *Writer) callback(out []int16) {
	data := out
	if w.buf.Len() < len(out)*2 {
		data = out[:w.buf.Len()/2]
	}
	binary.Read(w.buf, binary.LittleEndian, data)
	if w.buf.Len() == 0 {
		select {
		case w.done <- true:
		default:
		}
	}
}

func (w *Writer) Play() error {
	err := w.stream.Start()
	if err != nil {
		return fmt.Errorf("Error on stream start: %s", err)
	}
	<-w.done
	return w.stream.Close()
}

func (w *Writer) Close() error {
	return w.stream.Close()
}

func play(device *portaudio.DeviceInfo, buf []byte) error {
	w, err := NewWriter(device, 16000, 1, buf)
	if err != nil {
		return err
	}
	err = w.Play()
	if err != nil {
		return err
	}
	return w.Close()
}

func main() {
	err := portaudio.Initialize()
	if err != nil {
		log.Fatalf("Error initialize audio interface: %s", err)
	}

	f, err := os.Open("test.pcm")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	buf, _ := ioutil.ReadAll(f)

	devices, err := portaudio.Devices()
	if err != nil {
		log.Fatal(err)
	}

	for _, device := range devices {
		if device.MaxOutputChannels == 0 {
			continue
		}
		log.Printf("test %s", device.Name)
		err = play(device, buf)
		if err != nil {
			log.Print(err)
		}
	}
}
