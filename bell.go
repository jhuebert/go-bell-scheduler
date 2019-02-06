package main

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func GetPlayBellFunc(path string) func() {
	return func() {
		log.Info("Playing bell")
		playBell(path)
	}
}

func playBell(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Errorf("Sound file not found - %v", err)
		return
	}

	s, format, err := wav.Decode(f)
	if err != nil {
		log.Errorf("Cannot decode sound file - %v", err)
		return
	}

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Errorf("Cannot initialize speaker - %v", err)
		return
	}

	done := make(chan struct{})

	t := beep.Loop(5, s)

	speaker.Play(beep.Seq(t, beep.Callback(func() {
		close(done)
	})))

	<-done
}
