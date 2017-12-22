package iface

import (
	"fmt"

	"github.com/icexin/dueros/proto"
)

type Screen struct {
}

func (s *Screen) RenderVoiceInputText(m *proto.Message) error {
	fmt.Printf("\r>>> %-40s", m.PayloadJSON.Get("text"))
	if m.PayloadJSON.Get("type").String() == "FINAL" {
		fmt.Println("\n>>> 倾听完毕")
	}

	return nil
}

func init() {
	RegisterService(new(Screen), "ai.dueros.device_interface.screen")
}
