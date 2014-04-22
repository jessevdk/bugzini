package main

import (
	"bugzilla"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

func ProductHandler(w http.ResponseWriter, r *http.Request) {
	noCache(w)

	vars := mux.Vars(r)

	client, err := Bz()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	idv, err := strconv.ParseInt(vars["id"], 10, 32)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := int(idv)

	if product, ok := cache.ProductMap[id]; ok {
		JsonResponse(w, product)
		return
	}

	product, err := client.Products().Get(client, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cache.ProductMap[id] = product
	cache.Save()

	JsonResponse(w, product)
}

func ProductBugsHandler(w http.ResponseWriter, r *http.Request) {
	noCache(w)

	vars := mux.Vars(r)

	idv, err := strconv.ParseInt(vars["id"], 10, 32)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := int(idv)

	client, err := Bz()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	product, ok := cache.ProductMap[id]

	if !ok {
		product, err = client.Products().Get(client, id)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cache.ProductMap[id] = product
		cache.Save()
	}

	after := r.FormValue("after")

	// Only use cache if not asking for bugs after a certain date
	if len(after) == 0 {
		if bugs, ok := cache.Bugs[id]; ok {
			JsonResponse(w, bugs)
			return
		}
	}

	var bugs *bugzilla.BugList

	if len(after) != 0 {
		afsec, err := strconv.ParseInt(after, 10, 64)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		t := time.Unix(afsec, 0)
		bugs, err = product.BugsAfter(client, t)
	} else {
		bugs, err = product.Bugs(client)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	i := 0
	pbugs := make([]*bugzilla.Bug, 0)

	for {
		bug, err := bugs.Get(client, i)

		if err != nil {
			break
		}

		pbugs = append(pbugs, bug)

		if len(after) != 0 {
			if b, ok := cache.bugsMap[bug.Id]; ok {
				*b = *bug
			} else {
				cache.Bugs[id] = append(cache.Bugs[id], bug)
				cache.bugsMap[bug.Id] = cache.Bugs[id][len(cache.Bugs[id])-1]
			}
		}

		i++
	}

	if len(after) == 0 {
		cache.Bugs[id] = pbugs
	}

	cache.Save()

	JsonResponse(w, pbugs)
}

func ProductAllHandler(w http.ResponseWriter, r *http.Request) {
	noCache(w)

	if cache.Products != nil {
		JsonResponse(w, cache.Products)
		return
	}

	client, err := Bz()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list, err := client.Products().List()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret := make([]bugzilla.Product, 0, list.Len())

	for i := 0; i < list.Len(); i++ {
		p, err := list.Get(client, i)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ret = append(ret, p)
	}

	cache.Products = ret

	for _, p := range ret {
		cache.ProductMap[p.Id] = p
	}

	cache.Save()

	JsonResponse(w, ret)
}

func init() {
	router.HandleFunc("/api/product/all", ProductAllHandler)
	router.HandleFunc("/api/product/{id:[0-9]+}", ProductHandler)
	router.HandleFunc("/api/product/{id:[0-9]+}/bugs", ProductBugsHandler)
}
