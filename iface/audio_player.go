package iface

import (
	"time"

	"github.com/icexin/dueros/audio"
	"github.com/icexin/dueros/duer"
	"github.com/icexin/dueros/proto"
	"github.com/tidwall/gjson"
)

const (
	AudioStatePlaying  = "PLAYING"
	AudioStateStoped   = "STOPED"
	AudioStatePaused   = "PAUSED"
	AudioStateFinished = "FINISHED"
)

type AudioPlayer struct {
	p             *audio.Player
	currWriter    *audio.Writer
	currAudioItem gjson.Result
	state         string
}

func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{
		p:     audio.NewPlayer(),
		state: AudioStateFinished,
	}
}

func (a *AudioPlayer) Play(m *proto.Message) error {
	payload := &m.PayloadJSON
	stream := payload.Get("audioItem.stream")
	url := stream.Get("url").String()
	token := stream.Get("token").String()
	w, err := a.p.LoadMP3(url)
	if err != nil {
		return err
	}

	// 关闭前一个播放的音乐，同时等待结束
	a.state = AudioStateStoped
	if a.currWriter != nil {
		a.currWriter.Close()
		a.currWriter.Wait()
	}

	a.currWriter = w
	err = a.currWriter.Start()
	if err != nil {
		return err
	}
	a.currAudioItem = payload.Get("audioItem")
	a.sendPlaybackStarted(token)

	go a.reportProgress(w, m)
	go func() {
		w.Wait()
		w.Close()
		if a.state != AudioStateStoped {
			a.sendPlaybackNearlyFinished(token)
		}
		a.sendPlaybackFinished(token)
	}()
	return nil
}

func (a *AudioPlayer) Stop(m *proto.Message) error {
	a.state = AudioStateStoped
	if a.currWriter != nil {
		a.currWriter.Close()
	}
	return nil
}

func (a *AudioPlayer) Context() *proto.Message {
	token := a.currAudioItem.Get("stream.token").String()
	var offset int32
	if a.currWriter != nil {
		offset = int32(a.currWriter.Offset() / time.Millisecond)
	}
	return proto.NewMessage("ai.dueros.device_interface.audio_player.PlaybackState", map[string]interface{}{
		"token":                token,
		"offsetInMilliseconds": offset,
		"playerActivity":       a.state,
	})
}

func (a *AudioPlayer) sendPlaybackNearlyFinished(token string) {
	duer.OS.PostEvent(proto.NewMessage("ai.dueros.device_interface.audio_player.PlaybackNearlyFinished", map[string]string{
		"token": token,
	}))
}

func (a *AudioPlayer) sendPlaybackStarted(token string) {
	a.state = AudioStatePlaying
	duer.OS.PostEvent(proto.NewMessage("ai.dueros.device_interface.audio_player.PlaybackStarted", map[string]string{
		"token": token,
	}))
}

func (a *AudioPlayer) sendPlaybackFinished(token string) {
	a.state = AudioStateFinished
	duer.OS.PostEvent(proto.NewMessage("ai.dueros.device_interface.audio_player.PlaybackFinished", map[string]string{
		"token": token,
	}))
}

func (a *AudioPlayer) reportProgress(w *audio.Writer, m *proto.Message) {
	stream := m.PayloadJSON.Get("audioItem.stream")
	interval := stream.Get("progressReport.progressReportIntervalInMilliseconds").Int()
	if interval == 0 {
		return
	}
	token := stream.Get("token").String()

	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		if w.Closed() {
			break
		}
		duer.OS.PostEvent(proto.NewMessage("ai.dueros.device_interface.audio_player.ProgressReportIntervalElapsed", map[string]interface{}{
			"token":                token,
			"offsetInMilliseconds": w.Offset() / time.Millisecond,
		}))
	}
}

func (a *AudioPlayer) Pause(m *proto.Message) error {
	a.state = AudioStatePaused
	if a.currWriter != nil {
		a.currWriter.Pause()
	}
	return nil
}

func (a *AudioPlayer) Resume(m *proto.Message) error {
	if a.currWriter != nil {
		a.state = AudioStatePlaying
		a.currWriter.Resume()
		return nil
	}
	a.state = AudioStateStoped
	return nil
}

func init() {
	RegisterService(NewAudioPlayer(), "ai.dueros.device_interface.audio_player")
}
