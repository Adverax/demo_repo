package server

import (
	"github.com/adverax/echo"
	"github.com/adverax/echo/middleware"
	"log"
	"os"
	"repo/photo"
)

const thumbnailSize = 100

func Setup() {
	workdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	filer := &photo.FileEngine{
		Images:     workdir + "/static/images/",
		Thumbnails: workdir + "/static/thumbnails/",
		ThumbnailManager: &photo.ThumbnailEngine{
			Size: thumbnailSize,
		},
	}

	e := echo.New()
	router := e.Router()

	router.Use(middleware.Static("/static"))

	router.Post(
		"/upload",
		actionUpload(
			filer,
		),
	)

	log.Fatal(e.Start(":8080"))
}
