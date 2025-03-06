package delivery

import (
	requests "auth/request"
	"auth/usecase"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mailru/easyjson"
)

type IApi interface {
	Signin(w http.ResponseWriter, r *http.Request)
	Signup(w http.ResponseWriter, r *http.Request)
	LogoutSession(w http.ResponseWriter, r *http.Request)
	AuthAccept(w http.ResponseWriter, r *http.Request)
}

type API struct {
	core usecase.ICore
	mx   *http.ServeMux
}

func (a *API) ListenAndServe() error {
	err := http.ListenAndServe(":8081", a.mx)
	if err != nil {
		return fmt.Errorf("listen and serve error: %w", err)
	}

	return nil
}

func GetApi(c *usecase.Core) *API {
	api := &API{
		core: c,
		mx:   http.NewServeMux(),
	}

	api.mx.HandleFunc("/signin", api.Signin)
	api.mx.HandleFunc("/signup", api.Signup)
	api.mx.HandleFunc("/logout", api.LogoutSession)
	api.mx.HandleFunc("/authcheck", api.AuthAccept)

	return api
}

func (a *API) LogoutSession(w http.ResponseWriter, r *http.Request) {

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		return
	}

	found, _ := a.core.FindActiveSession(r.Context(), session.Value)
	if !found {
		return
	} else {
		err := a.core.KillSession(r.Context(), session.Value)
		if err != nil {
			fmt.Printf("failed to kill session", err.Error())
		}
		session.Expires = time.Now().AddDate(0, 0, -1)
		http.SetCookie(w, session)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (a *API) AuthAccept(w http.ResponseWriter, r *http.Request) {

	var authorized bool

	session, err := r.Cookie("session_id")
	if err == nil && session != nil {
		authorized, _ = a.core.FindActiveSession(r.Context(), session.Value)
	}

	if !authorized {
		return
	}
	login, err := a.core.GetUserName(r.Context(), session.Value)
	if err != nil {
		return
	}

	role, err := a.core.GetUserRole(login)
	if err != nil {
		return
	}

	jsonResponse, err := easyjson.Marshal(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonResponse)
	if err != nil {
		fmt.Printf("failed to send response", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (a *API) Signin(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		return
	}

	var request requests.SigninRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	if err = easyjson.Unmarshal(body, &request); err != nil {
		return
	}

	user, found, err := a.core.FindUserAccount(request.Login, request.Password)
	if err != nil {
		return
	}
	if !found {
		return
	} else {
		sid, session, _ := a.core.CreateSession(r.Context(), user.Login)
		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    sid,
			Path:     "/",
			Expires:  session.Expires,
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (a *API) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	var request requests.SignupRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Signup error", err.Error())
		return
	}

	err = easyjson.Unmarshal(body, &request)
	if err != nil {
		return
	}

	found, err := a.core.FindUserByLogin(request.Login)
	if err != nil {
		return
	}

	if found {
		return
	}
	err = a.core.CreateUserAccount(request.Login, request.Password, request.Name, request.BirthDate, request.Email)
	if err == usecase.InvalideEmail {
		w.WriteHeader(http.StatusBadRequest)
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
