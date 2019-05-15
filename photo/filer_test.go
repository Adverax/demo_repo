package photo

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

type thumbnailEngine struct {
	src string
	dst string
}

func (e *thumbnailEngine) Execute(src, dst string) error {
	e.src = src
	e.dst = dst
	return nil
}

func TestFileEngine_Append(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)

	thumbnails := &thumbnailEngine{}

	e := &FileEngine{
		ThumbnailManager: thumbnails,
		Images:           dir + "/../static/images/",
		Thumbnails:       dir + "/../static/thumbnails/",
	}

	name := "dog.jpg"
	image := e.Images + name
	thumbnail := e.Thumbnails + name

	file, err := os.Open(dir + "/../_fixtures/src.jpg")
	require.NoError(t, err)
	defer file.Close()

	err = e.Append(name, file)
	require.NoError(t, err)
	require.FileExists(t, image)
	assert.Equal(t, image, thumbnails.src)
	assert.Equal(t, thumbnail, thumbnails.dst)

	_ = os.Remove(image)
}
