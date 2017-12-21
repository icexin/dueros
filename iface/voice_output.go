package iface

import (
	"github.com/icexin/dueros/audio"
	"github.com/icexin/dueros/proto"
)

type VoiceOutput struct {
	p *audio.Player
}

func NewVoiceOutput() *VoiceOutput {
	return &VoiceOutput{
		p: audio.NewPlayer(),
	}
}

func (v *VoiceOutput) Speak(m *proto.Message) error {
	defer m.Attach.Close()
	w, err := v.p.LoadMP3Reader(m.Attach)
	if err != nil {
		return err
	}
	defer w.Close()
	player := DefaultRegistry.GetService("ai.dueros.device_interface.audio_player").(*AudioPlayer)
	if player != nil {
		player.Pause(nil)
		defer player.Resume(nil)
	}
	err = w.Play()
	if err != nil {
		return err
	}
	return nil
}

func (v *VoiceOutput) Pause(m *proto.Message) error {
	return nil
}

func init() {
	RegisterService(NewVoiceOutput(), "ai.dueros.device_interface.voice_output")
}
