package utils

import (
	"ResuMatch/internal/config"
	l "ResuMatch/pkg/logger"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const GenerateURL = "http://gotenberg:3000/forms/chromium/convert/html"

func GeneratePDF(htmlContent string, cfg config.ResumeConfig) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("не удалось создать форму из файла %w", err)
	}

	_, err = part.Write([]byte(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("не удалось записать HTML-содержимое в multipart-форму: %w", err)
	}

	if err = writer.WriteField("paperWidth", cfg.PaperWidth); err != nil {
		return nil, fmt.Errorf("не удалось установить ширину страницы: %w", err)
	}

	if err = writer.WriteField("paperHeight", cfg.PaperHeight); err != nil {
		return nil, fmt.Errorf("не удалось установить высоту страницы: %w", err)
	}

	if err = writer.Close(); err != nil {
		return nil, fmt.Errorf("не удалось завершить multipart-форму: %w", err)
	}

	resp, err := http.Post(
		cfg.GenerateURL,
		writer.FormDataContentType(),
		body,
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось сделать запрос к Gotenberg: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			l.Log.Errorf("Ошибка при закрытии тела ответа: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ошибка Gotenberg (%d): %s", resp.StatusCode, string(errorBody))
	}
	return io.ReadAll(resp.Body)
}
