package main

import (
	"bugzilla"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"os"
)

const CacheHeader = "bugzini"

type CacheItem struct {
	Products   []bugzilla.Product
	ProductMap map[int]bugzilla.Product
	Bugs       map[int][]*bugzilla.Bug

	bugsMap map[int]*bugzilla.Bug
}

type Cache struct {
	Hosts map[string]*CacheItem

	c *CacheItem
}

var cache = Cache{
	Hosts: make(map[string]*CacheItem),
}

func newCacheItem() *CacheItem {
	return &CacheItem{
		Bugs:       make(map[int][]*bugzilla.Bug),
		ProductMap: make(map[int]bugzilla.Product),
		bugsMap:    make(map[int]*bugzilla.Bug),
	}
}

func (c *Cache) readHeader(r io.Reader) (uint32, error) {
	bs := make([]byte, len(CacheHeader))

	if _, err := r.Read(bs); err != nil {
		return 0, err
	}

	if string(bs) != CacheHeader {
		return 0, fmt.Errorf("Invalid cache header")
	}

	var version uint32
	if err := binary.Read(r, binary.LittleEndian, &version); err != nil {
		return 0, err
	}

	return version, nil
}

func (c *Cache) writeHeader(w io.Writer) error {
	if _, err := w.Write([]byte(CacheHeader)); err != nil {
		return err
	}

	var version uint32
	version = BugziniVersion

	if err := binary.Write(w, binary.LittleEndian, version); err != nil {
		return err
	}

	return nil
}

func (c *Cache) Load() {
	if f, err := os.Open(".cache"); err == nil {
		defer f.Close()

		version, err := c.readHeader(f)

		if err == nil && version == BugziniVersion {
			dec := gob.NewDecoder(f)

			if err := dec.Decode(c); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load cache: %s\n", err.Error())
			}

			for _, h := range c.Hosts {
				h.bugsMap = make(map[int]*bugzilla.Bug)

				for _, v := range h.Bugs {
					for _, bug := range v {
						h.bugsMap[bug.Id] = bug
					}
				}
			}
		}
	}

	if cur, ok := c.Hosts[options.Bugzilla.Host]; ok {
		c.c = cur
	} else {
		c.c = newCacheItem()
		c.Hosts[options.Bugzilla.Host] = c.c
	}
}

func (c *Cache) Save() {
	if f, err := os.Create(".cache"); err == nil {
		defer f.Close()

		c.writeHeader(f)
		enc := gob.NewEncoder(f)
		enc.Encode(c)
	}
}
