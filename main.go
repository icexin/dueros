package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/icexin/dueros/audio"
	"github.com/icexin/dueros/auth"
	"github.com/icexin/dueros/duer"
	"github.com/icexin/dueros/iface"
)

var (
	wakeupMethod = flag.String("wakeup", "keyword", "wakeup method(keyboard|keyword)")
)

func setuplog() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	f, err := os.OpenFile("duer.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
}

func setuphttp() {
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
}

func waitToken() {
	_, err := auth.GetToken()
	if err == nil {
		return
	}
	fmt.Println("open browser, type: http://pi.local:8080/login")
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	for range ticker.C {
		_, err := auth.GetToken()
		if err == nil {
			return
		}
	}
}

func main() {
	flag.Parse()

	setuplog()
	setuphttp()
	// 等待access token被设置好
	waitToken()

	duer.OS = duer.NewDuerOS(iface.DefaultRegistry)
	wakeup := NewWakeupListener(*wakeupMethod)
	player := audio.NewPlayer()
	voiceInput := iface.DefaultRegistry.GetService("ai.dueros.device_interface.voice_input").(*iface.VoiceInput)
	for {
		fmt.Println(">>> 等待唤醒")
		wakeup.ListenAndWakeup()
		player.LoadAndPlay("resource/du.mp3")
		voiceInput.Listen(nil)
	}
}
