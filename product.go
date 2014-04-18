package main

import (
	"bugzilla"
	"encoding/gob"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"strconv"
)

type ProductCache struct {
	Products []bugzilla.Product `json:"products"`
	ProductMap map[int]bugzilla.Product `json:"product-map"`
	Bugs map[int][]bugzilla.Bug `json:"bugs"`
}

var productCache = ProductCache{
	Bugs: make(map[int][]bugzilla.Bug),
	ProductMap: make(map[int]bugzilla.Product),
}

func (p *ProductCache) Load() {
	if f, err := os.Open(".products-cache"); err == nil {
		dec := gob.NewDecoder(f)
		dec.Decode(p)
		f.Close()
	}
}

func (p *ProductCache) Save() {
	if f, err := os.Create(".products-cache"); err == nil {
		enc := gob.NewEncoder(f)
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

	idv, err := strconv.ParseInt(vars["id"], 10, 32)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := int(idv)

	if product, ok := productCache.ProductMap[id]; ok {
		JsonResponse(w, product)
		return
	}

	product, err := client.Products().Get(client, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	productCache.ProductMap[id] = product
	productCache.Save()

	JsonResponse(w, product)
}

func ProductBugsHandler(w http.ResponseWriter, r *http.Request) {
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

	product, ok := productCache.ProductMap[id]

	if !ok {
		product, err = client.Products().Get(client, id)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		productCache.ProductMap[id] = product
		productCache.Save()
	}

	if bugs, ok := productCache.Bugs[id]; ok {
		JsonResponse(w, bugs)
		return
	}

	bugs, err := product.Bugs(client)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	i := 0
	pbugs := make([]bugzilla.Bug, 0)

	for {
		bug, err := bugs.Get(client, i)

		if err != nil {
			break
		}

		pbugs = append(pbugs, *bug)
		i++
	}

	productCache.Bugs[id] = pbugs
	productCache.Save()

	JsonResponse(w, pbugs)
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
		p, err := list.Get(client, i)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ret = append(ret, p)
	}

	productCache.Products = ret

	for _, p := range ret {
		productCache.ProductMap[p.Id] = p
	}

	productCache.Save()

	JsonResponse(w, ret)
}

func init() {
	router.HandleFunc("/api/product/all", ProductAllHandler)
	router.HandleFunc("/api/product/{id:[0-9]+}", ProductHandler)
	router.HandleFunc("/api/product/{id:[0-9]+}/bugs", ProductBugsHandler)

	productCache.Load()
}
