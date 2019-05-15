package photo

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestThumbnailEngine_Execute(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)

	e := &ThumbnailEngine{
		Size: 100,
	}

	src := dir + "/../_fixtures/src.jpg"
	dst := dir + "/../_fixtures//dst.jpg"
	err = e.Execute(src, dst)
	require.NoError(t, err)
	require.FileExists(t, dst)

	_ = os.Remove(dst)
}
