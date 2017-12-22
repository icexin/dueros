package iface

import (
	"fmt"
	"io"

	"github.com/icexin/dueros/audio"
	"github.com/icexin/dueros/duer"
	"github.com/icexin/dueros/proto"
	uuid "github.com/satori/go.uuid"
)

type VoiceInput struct {
	stream io.ReadCloser
}

func NewVoiceInput() *VoiceInput {
	return &VoiceInput{}
}

func (v *VoiceInput) Listen(m *proto.Message) error {
	if v.stream != nil {
		v.stream.Close()
	}
	v.slience()
	fmt.Println(">>> 正在倾听")
	v.stream = audio.NewRecordStream()
	ctxid := uuid.NewV4().String()
	message := proto.NewMessage("ai.dueros.device_interface.voice_input.ListenStarted", map[string]string{
		"format": "AUDIO_L16_RATE_16000_CHANNELS_1",
	})
	message.Header.DialogRequestId = ctxid
	message.Attach = v.stream
	duer.OS.PostEvent(message)
	return nil
}

func (v *VoiceInput) StopListen(m *proto.Message) error {
	if v.stream != nil {
		v.stream.Close()
	}
	player := DefaultRegistry.GetService("ai.dueros.device_interface.audio_player").(*AudioPlayer)
	if player != nil {
		player.Resume(nil)
	}
	return nil
}

func (v *VoiceInput) slience() {
	player := DefaultRegistry.GetService("ai.dueros.device_interface.audio_player").(*AudioPlayer)
	if player != nil {
		player.Pause(nil)
	}
}

func init() {
	RegisterService(NewVoiceInput(), "ai.dueros.device_interface.voice_input")
}
