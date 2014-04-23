package main

import (
	"bugzilla"
	"net/http"
)

func UserCurrentHandler(w http.ResponseWriter, r *http.Request) {
	noCache(w)

	u := bugzilla.CurrentUser()

	if u == nil {
		http.Error(w, "Not logged in", http.StatusNoContent)
		return
	}

	JsonResponse(w, u)
}

func UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	noCache(w)

	user := r.FormValue("user")
	password := r.FormValue("password")

	if len(user) == 0 || len(password) == 0 {
		http.Error(w, "User or password not provided", http.StatusBadRequest)
		return
	}

	client, err := Bz()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u, err := client.Users().Login(user, password)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	JsonResponse(w, u)
}

func UserLogoutHandler(w http.ResponseWriter, r *http.Request) {
	noCache(w)

	if bugzilla.CurrentUser() == nil {
		return
	}

	client, err := Bz()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := client.Users().Logout(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bz = nil
}

func init() {
	router.HandleFunc("/api/user/current", UserCurrentHandler)
	router.HandleFunc("/api/user/login", UserLoginHandler)
	router.HandleFunc("/api/user/logout", UserLogoutHandler)
}
