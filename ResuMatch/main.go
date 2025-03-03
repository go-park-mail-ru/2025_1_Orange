package main

import (
	"ResuMatch/router"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Создаем маршрутизатор
	mux := router.NewRouter()

	// Запускаем сервер
	port := ":8000"
	fmt.Println("Сервер запущен на http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, mux))

}
