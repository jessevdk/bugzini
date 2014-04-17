package bugzilla

import (
	"errors"
	"time"
)

type Bug struct {
	conn *Conn

	Alias          string    `xmlrpc:"alias" json:"alias"`
	AssignedTo     string    `xmlrpc:"assigned_to" json:"assigned_to"`
	Component      string    `xmlrpc:"component" json:"component"`
	CreationTime   time.Time `xmlrpc:"creation_time" json:"creation_time"`
	DupeOf         int       `xmlrpc:"dupe_of" json:"dupe_of"`
	Id             int       `xmlrpc:"id" json:"id"`
	IsOpen         bool      `xmlrpc:"is_open" json:"is_open"`
	LastChangeTime time.Time `xmlrpc:"last_change_time" json:"last_change_time"`
	Priority       string    `xmlrpc:"priority" json:"priority"`
	Product        string    `xmlrpc:"product" json:"product"`
	Resolution     string    `xmlrpc:"resolution" json:"resolution"`
	Severity       string    `xmlrpc:"severity" json:"severity"`
	Status         string    `xmlrpc:"status" json:"status"`
	Summary        string    `xmlrpc:"summary" json:"summary"`

	attachments []Attachment
}

type Attachment struct {
	CreationTime   time.Time `xmlrpc:"creation_time" json:"creation_time"`
	LastChangeTime time.Time `xmlrpc:"last_change_time" json:"last_change_time"`
	Id             int       `xmlrpc:"id" json:"id"`
	BugId          int       `xmlrpc:"bug_id" json:"bug_id"`
	FileName       string    `xmlrpc:"file_name" json:"file_name"`
	Description    string    `xmlrpc:"description" json:"description"`
	ContentType    string    `xmlrpc:"content_type" json:"content_type"`
	IsPrivate      bool      `xmlrpc:"is_private" json:"is_private"`
	IsObsolete     bool      `xmlrpc:"is_obsolete" json:"is_obsolete"`
	IsUrl          bool      `xmlrpc:"is_url" json:"is_url"`
	IsPatch        bool      `xmlrpc:"is_patch" json:"is_patch"`
	Attacher       string    `xmlrpc:"attacher" json:"attacher"`
}

type Bugs struct {
	conn *Conn
}

type BugList struct {
	conn     *Conn
	query    interface{}
	pageSize int

	bugs []Bug

	finished bool
}

type Query struct {
	Limit  int
	Offset int
}

func (c *Conn) Bugs() Bugs {
	return Bugs{
		conn: c,
	}
}

func (b Bugs) GetAll(ids []int) ([]Bug, error) {
	args := struct {
		Ids []int `xmlrpc:"ids" json:"ids"`
	}{
		Ids: ids,
	}

	var ret struct {
		Bugs []Bug `xmlrpc:"bugs" json:"bugs"`
	}

	if err := b.conn.Call("Bug.get", args, &ret); err != nil {
		return nil, err
	}

	for i := 0; i < len(ret.Bugs); i++ {
		ret.Bugs[i].conn = b.conn
	}

	return ret.Bugs, nil
}

func (b Bugs) Get(id int) (Bug, error) {
	ret, err := b.GetAll([]int{id})

	if err != nil {
		return Bug{}, err
	}

	return ret[0], nil
}

func (p Bugs) SearchPage(query interface{}, pageSize int) (*BugList, error) {
	return &BugList{
		conn:     p.conn,
		query:    query,
		pageSize: pageSize,
	}, nil
}

func (p Bugs) Search(query interface{}) (*BugList, error) {
	return p.SearchPage(query, 0)
}

func (b *BugList) Get(i int) (*Bug, error) {
	n := len(b.bugs)

	for i >= n && !b.finished {
		// Fetch next pageSize
		var ret struct {
			Bugs []Bug `xmlrpc:"bugs" json:"bugs"`
		}

		paged := false

		if b.pageSize > 0 {
			if q, ok := b.query.(*Query); ok {
				q.Limit = b.pageSize
				q.Offset = n

				paged = true
			}
		}

		limit := b.pageSize

		if err := b.conn.Call("Bug.search", b.query, &ret); err != nil {
			return nil, err
		}

		b.bugs = append(b.bugs, ret.Bugs...)

		b.finished = (len(ret.Bugs) < limit || !paged)
		n = len(b.bugs)
	}

	if i >= n {
		return nil, errors.New("out of bounds")
	}

	b.bugs[i].conn = b.conn
	return &b.bugs[i], nil
}

func (b *Bug) Attachments() ([]Attachment, error) {
	if b.attachments != nil {
		return b.attachments, nil
	}

	args := struct {
		Ids []int `xmlrpc:"ids" json:"ids"`
	}{
		Ids: []int{b.Id},
	}

	var ret struct {
		Bugs        map[int][]Attachment `xmlrpc:"bugs" json:"bugs"`
		Attachments map[int]Attachment   `xmlrpc:"attachments" json:"attachments"`
	}

	if err := b.conn.Call("Bug.attachments", args, &ret); err != nil {
		return nil, err
	}

	b.attachments = ret.Bugs[b.Id]
	return b.attachments, nil
}
