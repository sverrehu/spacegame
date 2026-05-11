package client

import (
	_ "embed"
	"log"
	"os"

	"github.com/gogpu/gogpu/sound"
)

//go:embed audio/PhaserHit.wav
var phaserHit []byte

//go:embed audio/BigExplosion.wav
var bigExplosion []byte

//go:embed audio/SmallExplosion.wav
var smallExplosion []byte

var phaserHitFile string
var bigExplosionFile string
var smallExplosionFile string
var soundsEnabled bool

func initSounds() {
	sound.SetEnabled(true)
	soundsEnabled = true
	phaserHitFile = saveTemp(phaserHit)
	bigExplosionFile = saveTemp(bigExplosion)
	smallExplosionFile = saveTemp(smallExplosion)
}

func teardownSounds() {
	_ = os.Remove(phaserHitFile)
	_ = os.Remove(bigExplosionFile)
	_ = os.Remove(smallExplosionFile)
}

func playPhaserHit() {
	playFile(phaserHitFile)
}

func playBigExplosion() {
	playFile(bigExplosionFile)
}

func playSmallExplosion() {
	playFile(smallExplosionFile)
}

func playFile(filename string) {
	if !soundsEnabled {
		return
	}
	err := sound.PlayFile(filename)
	if err != nil {
		log.Printf("Error playing sound file: %v", err)
		soundsEnabled = false
	}
}

func saveTemp(bytes []byte) string {
	f, err := os.CreateTemp("", "spacegame-*.wav")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write(bytes); err != nil {
		log.Fatal(err)
	}
	_ = f.Close()
	return f.Name()
}
