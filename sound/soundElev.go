package sound

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"log"
	"os"
	"sync"
	"time"
)

var speakerInitialized = false

func initSpeaker(sampleRate beep.SampleRate) {
	if !speakerInitialized {
		speaker.Init(sampleRate, sampleRate.N(time.Second/10))
		speakerInitialized = true
	}
}

func AtFloor(floor int) {
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

	// Decode the MP3 file
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		log.Fatalf("Failed o decode MP3 file: %v ", err)
	}
	defer streamer.Close()

	// Initialize the speaker, if not already initialized
	initSpeaker(format.SampleRate)

	var wg sync.WaitGroup
	wg.Add(1)

	// Use a buffered channel to avoid blocking the speaker
	done := make(chan struct{}, 1)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- struct{}{}
	})))

	// Use a non-blocking select to wait for the audio to finish playing
	select {
	case <-done:
		wg.Done()
	case <-time.After(5 * time.Second): // Timeout based on the length of your audio file
	}

	wg.Wait()
}

func IAmBack() {
	filePath := "/Users/thomas/GolandProjects/TTK4145-Heis-2/sound/SoundEffects/Imback_elevator2.mp3"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open mp3: %v", err)
	}
	defer file.Close()
	// Decode the MP3 file
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		log.Fatalf("Failed o decode MP3 file: %v ", err)
	}
	defer streamer.Close()

	// Initialize the speaker, if not already initialized
	initSpeaker(format.SampleRate)

	var wg sync.WaitGroup
	wg.Add(1)

	// Use a buffered channel to avoid blocking the speaker
	done := make(chan struct{}, 1)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- struct{}{}
	})))

	// Use a non-blocking select to wait for the audio to finish playing
	select {
	case <-done:
		wg.Done()
	case <-time.After(10 * time.Second): // Timeout based on the length of your audio file
	}
	wg.Wait()
}
