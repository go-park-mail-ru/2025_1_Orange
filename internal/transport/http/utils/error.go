package utils

import "net/http"

func NewError(w http.ResponseWriter, status int, err error) {
	http.Error(w, err.Error(), status)
}
