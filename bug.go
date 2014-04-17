package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func BugHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	client, err := Bz()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := strconv.ParseInt(vars["id"], 10, 32)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bug, err := client.Bugs().Get(int(id))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	JsonResponse(w, bug)
}

func init() {
	router.HandleFunc("/api/bug/{id:[0-9]+}", BugHandler)
}
