package photo

import (
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"image/jpeg"
	"log"
	"os"
)

type ThumbnailManager interface {
	Execute(src, dst string) error
}

type ThumbnailEngine struct {
	Size int // Size of thumbnail (pixels)
}

// Make thumbnail from src image.
func (engine *ThumbnailEngine) Execute(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	originalImg, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	bounds := originalImg.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	if width > height {
		width = height
	} else {
		height = width
	}

	croppedImg, err := cutter.Crop(originalImg, cutter.Config{
		Width:   width,
		Height:  height,
		Mode:    cutter.Centered,
		Options: cutter.Copy, // Copy is useless here
	})

	m := resize.Resize(uint(engine.Size), uint(engine.Size), croppedImg, resize.Lanczos3)

	out, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	return jpeg.Encode(out, m, nil)
}
