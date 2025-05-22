package utils

import (
	"ResuMatch/internal/entity"
	"fmt"
	"github.com/mailru/easyjson"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, response easyjson.Marshaler) error {
	w.Header().Set("Content-Type", "application/json")
	jsonData, err := easyjson.Marshal(response)
	if err != nil {
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка сериализации http ответа: %w", err),
		)

	}
	if _, err := w.Write(jsonData); err != nil {
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка записи http ответа: %w", err),
		)
	}
	return nil
}

func ReadJSON(r *http.Request, v easyjson.Unmarshaler) error {
	err := easyjson.UnmarshalFromReader(r.Body, v)
	if err != nil {
		return entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("невалидный json: %w", err),
		)
	}
	return nil
}
