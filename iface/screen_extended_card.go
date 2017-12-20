package iface

import (
	"fmt"

	"github.com/icexin/dueros/proto"
)

type ScreenExtendedCard struct {
}

func (s *ScreenExtendedCard) RenderPlayerInfo(m *proto.Message) error {
	content := m.PayloadJSON.Get("content")
	fmt.Printf(">>> 正在播放 %s/%s/%s\n", content.Get("title"),
		content.Get("titleSubtext1"), content.Get("titleSubtext2"))
	return nil
}

func init() {
	RegisterService(new(ScreenExtendedCard), "ai.dueros.device_interface.screen_extended_card")
}
