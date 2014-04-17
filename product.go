package main

import (
	"bugzilla"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"strconv"
)

type ProductCache struct {
	Products []bugzilla.Product `json:"products"`
}

var productCache ProductCache

func (p *ProductCache) Load() {
	if f, err := os.Open(".products-cache"); err == nil {
		dec := json.NewDecoder(f)
		dec.Decode(p)
		f.Close()
	}
}

func (p *ProductCache) Save() {
	if f, err := os.Create(".products-cache"); err == nil {
		enc := json.NewEncoder(f)
		enc.Encode(p)
		f.Close()
	}
}

func ProductHandler(w http.ResponseWriter, r *http.Request) {
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

	product, err := client.Products().Get(int(id))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	JsonResponse(w, product)
}

func ProductAllHandler(w http.ResponseWriter, r *http.Request) {
	if productCache.Products != nil {
		JsonResponse(w, productCache.Products)
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
		p, err := list.Get(i)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ret = append(ret, p)
	}

	productCache.Products = ret
	productCache.Save()

	JsonResponse(w, ret)
}

func init() {
	router.HandleFunc("/api/product/all", ProductAllHandler)
	router.HandleFunc("/api/product/{id:[0-9]+}", ProductHandler)

	productCache.Load()
}
