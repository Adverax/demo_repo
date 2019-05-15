package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/adverax/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

type filerMock struct {
	name string
}

func (filer *filerMock) Append(name string, file io.Reader) error {
	filer.name = name
	return nil
}

func fileToBase64(src string) (string, error) {
	file, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", err
	}

	data := make([]byte, stat.Size())

	return base64.StdEncoding.EncodeToString(data), nil
}

func newFileUploadRequest(
	uri string,
	params map[string]string,
	paramName, filePath string,
) (*http.Request, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, fi.Name())
	if err != nil {
		return nil, err
	}

	_, err = part.Write(fileContents)
	if err != nil {
		return nil, err
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	return request, nil
}

func TestActUploadFromJSON(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)

	src := dir + "/../_fixtures/src.jpg"
	data, err := fileToBase64(src)
	require.NoError(t, err)

	e := echo.New()
	filer := new(filerMock)
	handler := actionUpload(filer)
	req := httptest.NewRequest(
		http.MethodPost,
		"/upload",
		strings.NewReader(
			fmt.Sprintf(`{"name":"dog.jpg","data":%q}`, data),
		),
	)
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	err = handler(ctx)
	assert.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	resp := strings.TrimSpace(rec.Body.String())
	assert.Equal(t, "true", resp)
	assert.Equal(t, "dog.jpg", filer.name)
}

func TestActUploadFromMultipartForm(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)

	src := dir + "/../_fixtures/src.jpg"

	e := echo.New()
	filer := new(filerMock)
	handler := actionUpload(filer)
	req, err := newFileUploadRequest(
		"/upload",
		map[string]string{
			"name": "dog.jpg",
		},
		"data",
		src,
	)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	err = handler(ctx)
	assert.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "dog.jpg", filer.name)
}

func TestActionUploadFromForm(t *testing.T) {
	e := echo.New()
	filer := new(filerMock)
	handler := actionUpload(filer)
	req := httptest.NewRequest(
		http.MethodPost,
		"/upload",
		strings.NewReader(
			(&url.Values{
				"name": []string{"dog.jpg"},
				"url":  []string{"https://www.infoniac.ru/upload/medialibrary/409/4099172ff56fa1e8d0a35b52bf908726.jpg"},
			}).Encode(),
		),
	)
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	err := handler(ctx)
	assert.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "dog.jpg", filer.name)
}
