package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/icexin/dueros-go/audio"
)

func main() {
	f, err := os.Create(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	ar, err := audio.NewReader(16000, 1, 1024)
	if err != nil {
		log.Fatal(err)
	}
	defer ar.Close()
	fmt.Println(">>> 按回车后开始录制")
	fmt.Scanln()
	go func() {
		fmt.Println(">>> 开始录制")
		io.Copy(f, ar)
	}()
	fmt.Println(">>> 按回车后结束录制")
	fmt.Scanln()
	ar.Pause()
	fmt.Printf(">>> 录制结束, %s saved\n", os.Args[1])
	ar.Close()
}
