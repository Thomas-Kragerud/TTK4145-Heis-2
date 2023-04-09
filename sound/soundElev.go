package sound

import (
	"context"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"log"
	"os"
	"time"
)

var (
	speakerInitialized = false
	songCtx            context.Context
	songCancel         context.CancelFunc
	currentSong        beep.Streamer
	ctrl               *beep.Ctrl
	isPlaying          bool
)

func initSpeaker(sampleRate beep.SampleRate) {
	if !speakerInitialized {
		speaker.Init(sampleRate, sampleRate.N(time.Second/10))
		speakerInitialized = true
	}
}

func AtFloor(floor int) {
	atFloorCtx, atFloorCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer atFloorCancel()

	var filePath string
	switch floor {
	case 0:
		filePath = "/Users/thomas/GolandProjects/TTK4145-Heis-2/sound/SoundEffects/f1.mp3"
	case 1:
		filePath = "/Users/thomas/GolandProjects/TTK4145-Heis-2/sound/SoundEffects/f2.mp3"
	case 2:
		filePath = "/Users/thomas/GolandProjects/TTK4145-Heis-2/sound/SoundEffects/f3.mp3"
	case 3:
		filePath = "/Users/thomas/GolandProjects/TTK4145-Heis-2/sound/SoundEffects/f4.mp3"
	default:
		return
	}
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open mp3: %v", err)
	}
	defer file.Close()

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		log.Fatalf("Failed o decode MP3 file: %v ", err)
	}
	defer streamer.Close()

	initSpeaker(format.SampleRate)

	done := make(chan struct{})
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- struct{}{}
	})))

	select {
	case <-done:
	case <-atFloorCtx.Done():
		speaker.Lock()
		speaker.Clear()
		speaker.Unlock()
	}
}

func IAmBack() {
	iamBackCtx, iamBackCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer iamBackCancel()

	filePath := "/Users/thomas/GolandProjects/TTK4145-Heis-2/sound/SoundEffects/Imback_elevator2.mp3"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open mp3: %v", err)
	}
	defer file.Close()

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		log.Fatalf("Failed o decode MP3 file: %v ", err)
	}
	defer streamer.Close()

	initSpeaker(format.SampleRate)

	done := make(chan struct{})
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- struct{}{}
	})))

	select {
	case <-done:
	case <-iamBackCtx.Done():
		speaker.Lock()
		speaker.Clear()
		speaker.Unlock()
	}
}

func StartCafeteria() {
	StartSong("/Users/thomas/GolandProjects/TTK4145-Heis-2/sound/SoundEffects/LEGO Star Wars II Music - Mos Eisley Cantina.mp3")
}

func StartSong(filePath string) {
	if songCancel != nil {
		songCancel()
	}

	if currentSong != nil && isPlaying {
		return
	}

	songCtx, songCancel = context.WithCancel(context.Background())

	go func() {
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatalf("Failed to open mp3: %v", err)
		}
		defer file.Close()

		streamer, format, err := mp3.Decode(file)
		if err != nil {
			log.Fatalf("Failed o decode MP3 file: %v ", err)
		}
		defer streamer.Close()

		initSpeaker(format.SampleRate)

		done := make(chan struct{})
		ctrl = &beep.Ctrl{Streamer: beep.Loop(-1, streamer)}
		currentSong = beep.Seq(ctrl, beep.Callback(func() {
			done <- struct{}{}
		}))
		speaker.Play(currentSong)
		isPlaying = true

		select {
		case <-done:
			isPlaying = false
		case <-songCtx.Done():
			speaker.Lock()
			speaker.Clear()
			speaker.Unlock()
			isPlaying = false
		}
	}()
}

func StopSong() {
	if songCancel != nil {
		songCancel()
	}
}

func Pause() {
	if ctrl != nil {
		ctrl.Paused = true
	}
}

func Resume() {
	if ctrl != nil {
		ctrl.Paused = false
	}
}
