package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/adverax/echo"
	"net/http"
	"path"
	"path/filepath"
	"repo/photo"
	"strings"
)

// Upload embedded file
func actUploadFromJSON(
	ctx echo.Context,
	filer photo.FileManager,
) error {
	req := ctx.Request()
	var params map[string]string
	if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
		if ute, ok := err.(*json.UnmarshalTypeError); ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset)).SetInternal(err)
		} else if se, ok := err.(*json.SyntaxError); ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error())).SetInternal(err)
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	name, ok := params["name"]
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Has no param 'name'")
	}

	name, err := normName(name)
	if err != nil {
		return err
	}

	if filepath.Ext(name) != ".jpg" {
		return ctx.JSON(http.StatusBadRequest, false)
	}

	data, ok := params["data"]
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Has no param 'data'")
	}

	d, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}

	err = filer.Append(name, bytes.NewReader(d))
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, true)
}

// Upload attached file
func actUploadFromMultipartForm(
	ctx echo.Context,
	filer photo.FileManager,
) error {
	_, err := ctx.MultipartForm()
	if err != nil {
		return err
	}

	file, handler, err := ctx.Request().FormFile("data")
	if err != nil {
		return err
	}
	defer file.Close()

	basename := filepath.Base(handler.Filename)
	if path.Ext(basename) != ".jpg" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file extension")
	}
	if name := ctx.FormValue("name"); name != "" {
		basename, err = normName(name)
		if err != nil {
			return err
		}
	}

	err = filer.Append(basename, file)
	if err != nil {
		return err
	}

	return ctx.String(http.StatusOK, "File is uploaded successfully")
}

// Upload file from external site
func actUploadFromForm(
	ctx echo.Context,
	manager photo.FileManager,
) error {
	name, err := normName(ctx.FormValue("name"))
	if err != nil {
		return err
	}

	url := ctx.FormValue("url")
	if url == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Has no param 'url'")
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = manager.Append(name, resp.Body)
	if err != nil {
		return err
	}

	return ctx.String(http.StatusOK, "File is uploaded successfully")
}

// Http handler for upload file.
// Supported following HeaderContentType
// * MIMEApplicationJSON - handle json request
// * MIMEMultipartForm - handle http multipart form data
func actionUpload(
	manager photo.FileManager,
) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		req := ctx.Request()
		if req.ContentLength == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "Request body can't be empty")
		}

		ctype := req.Header.Get(echo.HeaderContentType)
		switch {
		case strings.HasPrefix(ctype, echo.MIMEApplicationJSON):
			return actUploadFromJSON(ctx, manager)
		case strings.HasPrefix(ctype, echo.MIMEApplicationForm):
			return actUploadFromForm(ctx, manager)
		case strings.HasPrefix(ctype, echo.MIMEMultipartForm):
			return actUploadFromMultipartForm(ctx, manager)
		default:
			return echo.ErrUnsupportedMediaType
		}
	}
}

func normName(name string) (string, error) {
	ext := path.Ext(name)
	if ext == "" {
		return name + ".jpg", nil
	}
	if ext != ".jpg" {
		return "", echo.NewHTTPError(
			http.StatusBadRequest,
			"Invalid target file extension",
		)
	}
	return name, nil
}
