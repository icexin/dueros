package audio

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestRecordAndPlay(t *testing.T) {
	r, err := NewReader(16000, 1, 160)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	membuf := new(bytes.Buffer)
	buf := make([]byte, 320)
	fmt.Println("recording...")
	begin := time.Now()
	for {
		_, err := r.Read(buf)
		if err != nil {
			t.Error(err)
		}
		membuf.Write(buf)
		if time.Now().Sub(begin) > 3*time.Second {
			break
		}
	}
	fmt.Println("play...")
	w, err := NewWriter(16000, 1, membuf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	err = w.Play()
	if err != nil {
		t.Error(err)
	}
}

func TestPlayMP3(t *testing.T) {
	p := NewPlayer()
	w, err := p.LoadMP3File("testdata/Dota2_music_ui_main_02.mp3")
	if err != nil {
		t.Error(err)
	}
	defer w.Close()
	err = w.Play()
	if err != nil {
		t.Error(err)
	}
}
