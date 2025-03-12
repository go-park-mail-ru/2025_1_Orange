package profile

import (
	"testing"

	"ResuMatch/internal/models"
)

func TestCreateUser(t *testing.T) {
	storage := NewUserStorage()

	// Тест 1: Успешное создание пользователя
	user, err := storage.CreateUser("test@example.com", "password", "John", "Doe", "Company", "Address")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", user.Email)
	}

	// Тест 2: Попытка создать пользователя с существующим email
	_, err = storage.CreateUser("test@example.com", "password", "Jane", "Doe", "Company", "Address")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "email already exists" {
		t.Errorf("Expected error 'email already exists', got %v", err)
	}
}

func TestGetUserByEmail(t *testing.T) {
	storage := NewUserStorage()

	// Добавляем тестового пользователя
	storage.Users["test@example.com"] = models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: "password",
	}

	// Тест 1: Поиск существующего пользователя
	user, exists := storage.GetUserByEmail("test@example.com")
	if !exists {
		t.Error("Expected user to exist")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", user.Email)
	}

	// Тест 2: Поиск несуществующего пользователя
	_, exists = storage.GetUserByEmail("nonexisting@example.com")
	if exists {
		t.Error("Expected user to not exist")
	}
}
