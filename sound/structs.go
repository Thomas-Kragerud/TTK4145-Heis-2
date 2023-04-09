package sound

import "github.com/faiface/beep"

type SoundEvent int

const (
	startIdleSong SoundEvent = 0
	stopIdleSong  SoundEvent = 1
)

type AudioPlayer struct {
	ctrl     *beep.Ctrl
	streamer beep.StreamSeekCloser
	stopCh   chan struct{}
}
