package bugzilla

import (
	"errors"
	"time"
)

type Products struct {
	conn *Conn
}

type Product struct {
	conn *Conn

	Id          int    `xmlrpc:"id" json:"id"`
	Name        string `xmlrpc:"name" json:"name"`
	Description string `xmlrpc:"description" json:"description"`
}

type ProductList struct {
	conn *Conn

	Ids      []int `xmlrpc:"ids" json:"ids"`
	products []Product
}

func (c *Conn) Products() Products {
	return Products{
		conn: c,
	}
}

func (p Products) GetAll(conn *Conn, ids []int) ([]Product, error) {
	if conn == nil {
		conn = p.conn
	}

	args := struct {
		Ids []int `xmlrpc:"ids" json:"ids"`
	}{
		Ids: ids,
	}

	var ret struct {
		Products []Product `xmlrpc:"products" json:"products"`
	}

	if err := conn.Call("Product.get", args, &ret); err != nil {
		return nil, err
	}

	return ret.Products, nil
}

func (p Products) Get(conn *Conn, id int) (Product, error) {
	ret, err := p.GetAll(conn, []int{id})

	if err != nil {
		return Product{}, err
	}

	return ret[0], nil
}

func (p Products) List() (*ProductList, error) {
	var result ProductList

	if err := p.conn.Call("Product.get_accessible_products", nil, &result); err != nil {
		return nil, err
	}

	result.conn = p.conn
	result.products = make([]Product, 0, len(result.Ids))

	return &result, nil
}

func (p *ProductList) Len() int {
	return len(p.Ids)
}

func (p *ProductList) Get(conn *Conn, i int) (Product, error) {
	if conn == nil {
		conn = p.conn
	}

	if i >= len(p.Ids) {
		return Product{}, errors.New("out of bounds")
	}

	N := 300
	n := len(p.products)

	for i >= n {
		// Fetch next N
		var ret struct {
			Products []Product `xmlrpc:"products" json:"products"`
		}

		limit := N

		if n+limit > len(p.Ids) {
			limit = len(p.Ids) - n
		}

		pids := struct {
			Ids []int `xmlrpc:"ids" json:"ids"`
		}{
			Ids: p.Ids[n : n+limit],
		}

		if err := conn.Call("Product.get", pids, &ret); err != nil {
			return Product{}, err
		}

		p.products = append(p.products, ret.Products...)
		n = len(p.products)
	}

	p.products[i].conn = conn
	return p.products[i], nil
}

func (p *Product) Bugs(conn *Conn) (*BugList, error) {
	if conn == nil {
		conn = p.conn
	}

	return conn.Bugs().SearchPage(map[string]interface{}{
		"product":    []string{p.Name},
		"resolution": "",
	}, 300)
}

func (p *Product) BugsAfter(conn *Conn, after time.Time) (*BugList, error) {
	if conn == nil {
		conn = p.conn
	}

	return conn.Bugs().SearchPage(map[string]interface{}{
		"product":          []string{p.Name},
		"last_change_time": after,
		"resolution":       "",
	}, 300)
}
