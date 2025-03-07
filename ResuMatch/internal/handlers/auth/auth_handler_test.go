package auth

import (
	"ResuMatch/internal/data"
	"ResuMatch/internal/models"
	"ResuMatch/internal/repository/profile"
	"ResuMatch/internal/repository/session"
	"ResuMatch/internal/usecase"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestRequest(method, url string, body string) (*http.Request, error) {
	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func clearSessions() {

	for k := range session.Sessions {
		delete(session.Sessions, k)
	}
}

func TestSignup_Success(t *testing.T) {

	signupPayload := `{"email":"testuser@mail.ru", "password":"testpassword", "firstName":"Test User", "lastName": "test User", "companyName":"Company", "companyAddress": "Address" }`

	req, err := newTestRequest("POST", "/signup", signupPayload)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()

	core := usecase.NewCore(session.Sessionrepo{}, profile.UserRepo{})
	handler := NewMyHandler(core)

	handler.Signup(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Signup handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	expected := `{"message":"User created successfully"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Signup handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	if _, ok := data.Users["testuser"]; !ok {
		t.Error("User was not added to the Users map")
	}
}

func TestEmail_Success(t *testing.T) {

	testEmail := "testlemail@examle.com"
	testPass := "testpass"
	data.Users[testEmail] = models.User{
		ID:             6,
		Email:          testEmail,
		Password:       testPass,
		FirstName:      "Test Username",
		LastName:       "Test Username2",
		CompanyName:    "Test company",
		CompanyAddress: "Test address",
	}
	defer delete(data.Users, testEmail)

	emailPayload := `{"email":"testemail@example.com", "password":"testpass"}`

	req, err := newTestRequest("POST", "/signin", emailPayload)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()

	core := usecase.NewCore(session.Sessionrepo{}, profile.UserRepo{})
	handler := NewMyHandler(core)

	handler.Signin(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Login handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}
	sessionID, ok := resp["session_id"]
	if !ok || sessionID == "" {
		t.Errorf("Login handler did not return session_id in response")
	}

	cookies := rr.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Errorf("Login handler did not set session cookie")
	}

	clearSessions()
}

func TestLogout_Success(t *testing.T) {
	testEmail := "testemail@example.com"
	testPass := "testpass"
	data.Users[testEmail] = models.User{
		ID:             6,
		Email:          testEmail,
		Password:       testPass,
		FirstName:      "Test Username",
		LastName:       "Test Username2",
		CompanyName:    "Test company",
		CompanyAddress: "Test address",
	}
	defer delete(data.Users, testEmail)

	sid, err := usecase.CreateSessionID()
	if err != nil {
		t.Fatalf("Couldn't create session ID: %v", err)
	}

	req, err := newTestRequest("GET", "/logout", "")
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sid,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()
	core := usecase.NewCore(session.Sessionrepo{}, profile.UserRepo{})
	handler := NewMyHandler(core)

	handler.Logout(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Logout handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"message":"Logged out successfully"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Logout handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	cookies := rr.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_id" {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil || sessionCookie.Value != sid || !sessionCookie.Expires.Before(time.Now()) {
		t.Errorf("Logout handler did not expire the session cookie")
	}

}

func makeHTTPReq(method, path string, body string) (*http.Request, error) {
	var bodyReader *bytes.Reader
	if body != "" {
		bodyReader = bytes.NewReader([]byte(body))
	}
	req, err := http.NewRequest(method, path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func makeTestCoreAndHandler() (*usecase.Core, *MyHandler) {
	staticSessionRepo := session.Sessionrepo{}
	staticUserRepo := profile.UserRepo{}
	core := usecase.NewCore(staticSessionRepo, staticUserRepo)
	handler := NewMyHandler(core)
	return core, handler
}

func setUserInMap(email, password string, id uint64) {
	data.Users[email] = models.User{
		ID:             id,
		Email:          email,
		Password:       password,
		FirstName:      "Test Username",
		LastName:       "Test Username2",
		CompanyName:    "Test company",
		CompanyAddress: "Test address",
	}
}

func TestSignin(t *testing.T) {

	email := "testuser@example.com"
	password := "testpassword"
	userID := uint64(1)

	setUserInMap(email, password, userID)
	defer delete(data.Users, email)

	jsonBody := []byte(`{"email":"testuser", "password":"testpassword"}`)
	req, err := makeHTTPReq("POST", "/signin", string(jsonBody))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()

	_, handler := makeTestCoreAndHandler()

	handler.Signin(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	sessionID, ok := response["session_id"]
	if !ok {
		t.Fatalf("session_id not found in response")
	}

	if sessionID == "" {
		t.Errorf("session_id is empty")
	}

	cookieFound := false
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == "session_id" {
			cookieFound = true
			if cookie.Value != sessionID {
				t.Errorf("Session cookie value is incorrect")
			}
			break
		}
	}

	if !cookieFound {
		t.Errorf("Session cookie not found")
	}

	clearSessions()
}

func TestLogout(t *testing.T) {
	email := "testuser@example.com"
	password := "testpassword"
	userID := uint64(1)

	setUserInMap(email, password, userID)
	defer delete(data.Users, email)

	core, handler := makeTestCoreAndHandler()

	sessionID, err := core.CreateSession(context.Background(), userID)
	if err != nil {
		t.Fatalf("Could not create session: %v", err)
	}

	req, err := makeHTTPReq("GET", "/logout", "")
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	cookie := &http.Cookie{
		Name:  "session_id",
		Value: sessionID,
	}
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()

	handler.Logout(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	expectedMessage := "Logged out successfully"
	if response["message"] != expectedMessage {
		t.Errorf("handler returned unexpected body: got %v want %v", response["message"], expectedMessage)
	}

	sessionCookieFound := false
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == "session_id" {
			sessionCookieFound = true
			if cookie.Value != sessionID {
				t.Errorf("Session cookie value is incorrect")
			}
			if !cookie.Expires.Before(time.Now()) {
				t.Errorf("Session cookie expiration time is incorrect")
			}
			break
		}
	}

	if !sessionCookieFound {
		t.Errorf("Session cookie not found")
	}

	clearSessions()
}

func TestSignup(t *testing.T) {
	email := "testuser@example.com"
	jsonBody := []byte(`{"email":"` + email + `", "password":"testpassword", "name":"Test User", "birthdate":"1990-01-01", "email":"test@example.com"}`)
	req, err := makeHTTPReq("POST", "/signup", string(jsonBody))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()

	_, handler := makeTestCoreAndHandler()

	handler.Signup(rr, req)
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	expectedMessage := "User created successfully"
	if response["message"] != expectedMessage {
		t.Errorf("handler returned unexpected body: got %v want %v", response["message"], expectedMessage)
	}

	_, userExists := data.Users[email]
	if !userExists {
		t.Errorf("User not found in Users map")
	}
	delete(data.Users, email)

}
