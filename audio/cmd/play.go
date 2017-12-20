package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/icexin/dueros-go/audio"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	buf, _ := ioutil.ReadAll(f)
	w, err := audio.NewWriter(16000, 1, buf)
	if err != nil {
		log.Fatal(err)
	}
	w.Play()
	w.Close()
}
