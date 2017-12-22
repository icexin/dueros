package main

import (
	"flag"
	"fmt"
	"io"

	snowboy "github.com/brentnd/go-snowboy"
	"github.com/icexin/dueros/audio"
)

var (
	wakeupSensitivity = flag.Float64("sens", 0.4, "wakeup detector sensitivity")
)

const (
	KeyboardListener = "keyboard"
	KeywordListener  = "keyword"
)

type WakeupListener interface {
	ListenAndWakeup()
	Close() error
}

type keyboardWakeupListener struct {
}

func newKeyboardWakeupListener() WakeupListener {
	return new(keyboardWakeupListener)
}

func (k keyboardWakeupListener) ListenAndWakeup() {
	fmt.Scanln()
}

func (k keyboardWakeupListener) Close() error {
	return nil
}

type keywordWakeupListener struct {
	detector     snowboy.Detector
	recordReader io.ReadCloser
}

func newKeywordWakeupListener() WakeupListener {
	k := &keywordWakeupListener{
		detector: snowboy.NewDetector("resource/common.res"),
	}
	k.detector.HandleFunc(snowboy.NewHotword("resource/wakeup.pmdl", float32(*wakeupSensitivity)), k.onWakeup)
	return k
}

func (k *keywordWakeupListener) onWakeup(string) {
	fmt.Println(">>> wakeup")
	k.recordReader.Close()
}

func (k *keywordWakeupListener) ListenAndWakeup() {
	k.recordReader = audio.NewRecordStream()
	k.detector.ReadAndDetect(k.recordReader)
	k.detector.Reset()
}

func (k *keywordWakeupListener) Close() error {
	return k.detector.Close()
}

func NewWakeupListener(method string) WakeupListener {
	switch method {
	case KeyboardListener:
		return newKeyboardWakeupListener()
	case KeywordListener:
		return newKeywordWakeupListener()
	default:
		panic("wakeup method not found: " + method)
	}
}
