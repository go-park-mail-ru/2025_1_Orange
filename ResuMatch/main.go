package main

import (
	"ResuMatch/router"
	"fmt"
	"net/http"
)

func main() {
	// Создаем маршрутизатор
	mux := router.NewRouter()

	// Запускаем сервер
	port := ":8000"
	fmt.Println("Сервер запущен на http://localhost" + port)
	http.ListenAndServe(port, mux)
}
