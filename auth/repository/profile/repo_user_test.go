// package main

// import (
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"example.com/your_project/data"
// 	"example.com/your_project/models"
// )

// func TestGetUser_Success(t *testing.T) {

// 	expectedUser := data.Users[0]

// 	req, err := http.NewRequest("GET", "/users?login="+expectedUser.Login+"&password="+expectedUser.Password, nil)
// 	if err != nil {
// 		t.Fatalf("Не удалось создать запрос: %v", err)
// 	}

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(handleGetUser(core))

// 	handler.ServeHTTP(rr, req)

// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("Обработчик вернул неверный код статуса: получили %v, ожидали %v", status, http.StatusOK)
// 	}

// 	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
// 		t.Errorf("Обработчик вернул неверный Content-Type: получили %v, ожидали %v", contentType, "application/json")
// 	}

// 	var actualUser models.User
// 	err = json.NewDecoder(rr.Body).Decode(&actualUser)
// 	if err != nil {
// 		t.Fatalf("Не удалось декодировать тело ответа: %v", err)
// 	}

// 	if actualUser.Id != expectedUser.Id {
// 		t.Errorf("ID пользователя не совпадает: получили %v, ожидали %v", actualUser.Id, expectedUser.Id)
// 	}
// 	if actualUser.Login != expectedUser.Login {
// 		t.Errorf("Login пользователя не совпадает: получили %v, ожидали %v", actualUser.Login, expectedUser.Login)
// 	}
// 	if actualUser.Name != expectedUser.Name {
// 		t.Errorf("Name пользователя не совпадает: получили %v, ожидали %v", actualUser.Name, expectedUser.Name)
// 	}

// }

// func handleGetUser(core *Core) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		core.GetUser(w, r)
// 	}
// }
