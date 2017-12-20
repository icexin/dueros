package main

import (
	"fmt"
	"io"

	snowboy "github.com/brentnd/go-snowboy"
	"github.com/icexin/dueros/audio"
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
	k.detector.HandleFunc(snowboy.NewHotword("resource/wakeup.pmdl", 0.5), k.onWakeup)
	return k
}

func (k *keywordWakeupListener) onWakeup(string) {
	k.recordReader.Close()
}

func (k *keywordWakeupListener) ListenAndWakeup() {
	k.recordReader = audio.NewRecordStream()
	k.detector.ReadAndDetect(k.recordReader)
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
