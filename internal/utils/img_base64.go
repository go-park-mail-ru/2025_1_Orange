package utils

import (
	l "ResuMatch/pkg/logger"
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

func ConvertImageToBase64(imageUrl string) (string, error) {
	if imageUrl == "" {
		return "", nil
	}

	imageUrl = strings.Replace(imageUrl, "localhost", "minio", 1)
	resp, err := http.Get(imageUrl)
	if err != nil {
		return "", err
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			l.Log.Errorf("Ошибка при закрытии тела ответа: %v", closeErr)
		}
	}()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	mimeType := http.DetectContentType(buf.Bytes())

	var prefix string
	switch mimeType {
	case "image/jpeg":
		prefix = "data:image/jpeg;base64,"
	case "image/png":
		prefix = "data:image/png;base64,"
	default:
		return "", fmt.Errorf("unsupported media type: %s", mimeType)
	}

	return prefix + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
