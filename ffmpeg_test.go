package ffmpeg

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"testing"
	"time"
)

func TestEncoderUsage(t *testing.T) {

	f, err := os.Create("test.mpeg")
	if err != nil {
		log.Panicf("Unable to open output file: %q", err)
	}

	im := image.NewRGBA(image.Rect(0, 0, 640, 480))

	e, err := NewEncoder(CODEC_ID_MPEG4, im, f)
	if err != nil {
		log.Panicf("Unable to start encoder: %q", err)
	}

	start := time.Now()

	for i := 0; i < 25*5; i++ {
		c := color.RGBA{0, 0, uint8(i % 255), 255}
		// uint8(i%255), uint8(i%255), 255}
		draw.Draw(im, im.Bounds(), &image.Uniform{c}, image.ZP, draw.Src)

		err := e.WriteFrame()
		if err != nil {
			log.Panicf("Problem writing frame: %q", err)
		}
	}

	e.Close()
	f.Close()

	log.Printf("Took %s", time.Since(start))
}
