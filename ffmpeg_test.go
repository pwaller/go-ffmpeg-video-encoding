package ffmpeg

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"testing"
	"time"
)

func TestEncoderUsage(t *testing.T) {

	f, err := os.Create("test2.mp4")
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
		draw.Draw(im, im.Bounds().Add(image.Pt(i*5, 0)), image.NewUniform(c), image.ZP, draw.Src)

		err := e.WriteFrame(im)
		if err != nil {
			log.Panicf("Problem writing frame: %q", err)
		}
	}

	e.Close()
	f.Close()

	log.Printf("Took %s", time.Since(start))
}

func TestEncoderUsage2(t *testing.T) {
	// Start encoding
	f, err := os.Create("output/test3.mp4")
	if err != nil {
		log.Panicf("Unable to open output file: %q", err)
	}

	im := image.NewRGBA(image.Rect(0, 0, 352, 288))
	bi, wi := image.NewRGBA(im.Bounds()), image.NewRGBA(im.Bounds())
	draw.Draw(bi, bi.Bounds(), image.Black, image.ZP, draw.Src)
	draw.Draw(wi, wi.Bounds(), image.White, image.ZP, draw.Src)

	e, err := NewEncoder(CODEC_ID_MPEG4, im, f)
	if err != nil {
		log.Panicf("Unable to start encoder: %q", err)
	}

	start := time.Now()
	draw.Draw(im, im.Bounds(), image.Black, image.ZP, draw.Src)

	// listen on channel for tweened frames
	na := []int{0, 16, 32, 48, 64, 80, 96, 112, 128, 144, 160, 176, 192, 208, 224, 240, 256, 272, 288, 304, 320, 336, 352}
	r := image.Rect(0, 0, 352, 288)
	img := image.NewRGBA(r)

	for i := 0; i < len(na); i++ {
		pt := image.Pt(na[i], 0)

		draw.Draw(img, r, wi, image.ZP, draw.Src)
		draw.Draw(img, r.Sub(pt), bi, image.ZP, draw.Src)

		fmt.Println(img.At(i<<4, 5), r.Sub(pt))

		err := e.WriteFrame(img)
		if err != nil {
			log.Panicf("Problem writing frame: %q", err)
		}
	}

	e.Close()
	f.Close()

	log.Printf("Took %s", time.Since(start))
}

func TestEncoderUsage3(t *testing.T) {
	// Start encoding
	f, err := os.Create("output/test4.vp8")
	if err != nil {
		log.Panicf("Unable to open output file: %q", err)
	}

	im := image.NewRGBA(image.Rect(0, 0, 352, 288))
	bi, wi := image.NewRGBA(im.Bounds()), image.NewRGBA(im.Bounds())
	draw.Draw(bi, bi.Bounds(), image.Black, image.ZP, draw.Src)
	draw.Draw(wi, wi.Bounds(), image.White, image.ZP, draw.Src)

	e, err := NewEncoder(CODEC_ID_VP8, im, f)
	if err != nil {
		log.Panicf("Unable to start encoder: %q", err)
	}

	start := time.Now()
	draw.Draw(im, im.Bounds(), image.Black, image.ZP, draw.Src)

	// listen on channel for tweened frames
	na := []int{0, 16, 32, 48, 64, 80, 96, 112, 128, 144, 160, 176, 192, 208, 224, 240, 256, 272, 288, 304, 320, 336, 352}
	r := image.Rect(0, 0, 352, 288)
	img := image.NewRGBA(r)

	for i := 0; i < len(na); i++ {
		pt := image.Pt(na[i], 0)

		draw.Draw(img, r, wi, image.ZP, draw.Src)
		draw.Draw(img, r.Sub(pt), bi, image.ZP, draw.Src)

		fmt.Println(img.At(i<<4, 5), r.Sub(pt))

		err := e.WriteFrame(img)
		if err != nil {
			log.Panicf("Problem writing frame: %q", err)
		}
	}

	e.Close()
	f.Close()

	log.Printf("Took %s", time.Since(start))
}
