package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"bugzilla"
	"os"
	"encoding/gob"
)

type BugsCache struct {
	Bugs map[int]*bugzilla.Bug
}

var bugsCache = BugsCache{
	Bugs: make(map[int]*bugzilla.Bug),
}

func (b *BugsCache) Load() {
	if f, err := os.Open(".bugs-cache"); err == nil {
		dec := gob.NewDecoder(f)
		dec.Decode(b)
		f.Close()
	}
}

func (b *BugsCache) Save() {
	if f, err := os.Create(".bugs-cache"); err == nil {
		enc := gob.NewEncoder(f)
		enc.Encode(b)
		f.Close()
	}
}

func BugGet(id int) (*bugzilla.Bug, error) {
	if bug, ok := bugsCache.Bugs[id]; ok {
		return bug, nil
	}

	client, err := Bz()

	if err != nil {
		return nil, err
	}

	bug, err := client.Bugs().Get(client, id)

	if err != nil {
		return nil, err
	}

	bugsCache.Bugs[bug.Id] = &bug
	bugsCache.Save()

	return &bug, nil
}

func BugHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	idl, err := strconv.ParseInt(vars["id"], 10, 32)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := int(idl)

	bug, err := BugGet(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	JsonResponse(w, *bug)
}

func BugCommentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	idl, err := strconv.ParseInt(vars["id"], 10, 32)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := int(idl)

	bug, err := BugGet(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(bug.Comments) != 0 {
		JsonResponse(w, bug.Comments)
		return
	}

	client, err := Bz()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	comments, err := client.Bugs().GetComments(client, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bug.Comments = comments
	bugsCache.Save()

	JsonResponse(w, comments)
}

func init() {
	router.HandleFunc("/api/bug/{id:[0-9]+}/comments", BugCommentsHandler)
	router.HandleFunc("/api/bug/{id:[0-9]+}", BugHandler)

	bugsCache.Load()
}
