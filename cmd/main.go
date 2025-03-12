package main

import (
	"ResuMatch/internal/middleware"
	"ResuMatch/internal/router"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Создаем маршрутизатор
	mux := router.NewRouter()

	handler := middleware.CORS(mux)

	// Запускаем сервер
	port := ":8000"
	fmt.Println("Сервер запущен на http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, handler))

}
