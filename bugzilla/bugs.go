package bugzilla

import (
	"errors"
	"fmt"
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

	Comments    []Comment    `json:"comments"`
	Attachments []Attachment `json:"attachments"`
}

type Comment struct {
	Id           int       `xmlrpc:"id" json:"id"`
	BugId        int       `xmlrpc:"bug_id" json:"bug_id"`
	AttachmentId int       `xmlrpc:"attachment_id" json:"attachment_id"`
	Text         string    `xmlrpc:"text" json:"text"`
	Author       string    `xmlrpc:"author" json:"author"`
	Time         time.Time `xmlrpc:"time" json:"time"`
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

func (c *Conn) Bugs() Bugs {
	return Bugs{
		conn: c,
	}
}

func (b Bugs) GetAll(conn *Conn, ids []int) ([]Bug, error) {
	if conn == nil {
		conn = b.conn
	}

	args := struct {
		Ids []int `xmlrpc:"ids" json:"ids"`
	}{
		Ids: ids,
	}

	var ret struct {
		Bugs []Bug `xmlrpc:"bugs" json:"bugs"`
	}

	if err := conn.Call("Bug.get", args, &ret); err != nil {
		return nil, err
	}

	for i := 0; i < len(ret.Bugs); i++ {
		ret.Bugs[i].conn = conn
	}

	return ret.Bugs, nil
}

func (b Bugs) Get(conn *Conn, id int) (Bug, error) {
	ret, err := b.GetAll(conn, []int{id})

	if err != nil {
		return Bug{}, err
	}

	return ret[0], nil
}

func (b Bugs) AddComment(conn *Conn, id int, comment string) (int, error) {
	args := struct {
		Id      int    `xmlrpc:"id"`
		Comment string `xmlrpc:"comment"`
	}{
		Id:      id,
		Comment: comment,
	}

	var ret struct {
		Id int `xmlrpc:"id"`
	}

	if err := conn.Call("Bug.add_comment", args, &ret); err != nil {
		return 0, err
	}

	return ret.Id, nil
}

func (b Bugs) GetAllComments(conn *Conn, ids []int) ([]Comment, error) {
	args := struct {
		Ids []int `xmlrpc:"ids"`
	}{
		Ids: ids,
	}

	var ret struct {
		Bugs map[string]struct {
			Comments []Comment `xmlrpc:"comments"`
		} `xmlrpc:"bugs"`
	}

	if err := conn.Call("Bug.comments", args, &ret); err != nil {
		return nil, err
	}

	retval := make([]Comment, 0)

	for _, b := range ret.Bugs {
		retval = append(retval, b.Comments...)
	}

	return retval, nil
}

func (b Bugs) GetComments(conn *Conn, id int) ([]Comment, error) {
	ret, err := b.GetAllComments(conn, []int{id})

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (b Bugs) GetCommentsAfter(conn *Conn, id int, after time.Time) ([]Comment, error) {
	args := struct {
		Ids      []int     `xmlrpc:"ids"`
		NewSince time.Time `xmlrpc:"new_since"`
	}{
		Ids:      []int{id},
		NewSince: after,
	}

	var ret struct {
		Bugs map[string]struct {
			Comments []Comment `xmlrpc:"comments"`
		} `xmlrpc:"bugs"`
	}

	if err := conn.Call("Bug.comments", args, &ret); err != nil {
		return nil, err
	}

	return ret.Bugs[fmt.Sprintf("%v", id)].Comments, nil
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

func (b *BugList) Get(conn *Conn, i int) (*Bug, error) {
	if conn == nil {
		conn = b.conn
	}

	n := len(b.bugs)

	for i >= n && !b.finished {
		// Fetch next pageSize
		var ret struct {
			Bugs []Bug `xmlrpc:"bugs" json:"bugs"`
		}

		paged := false

		if b.pageSize > 0 {
			if q, ok := b.query.(map[string]interface{}); ok {
				q["limit"] = b.pageSize
				q["offset"] = n

				paged = true
			}
		}

		limit := b.pageSize

		if err := conn.Call("Bug.search", b.query, &ret); err != nil {
			return nil, err
		}

		b.bugs = append(b.bugs, ret.Bugs...)

		b.finished = (len(ret.Bugs) < limit || !paged)
		n = len(b.bugs)
	}

	if i >= n {
		return nil, errors.New("out of bounds")
	}

	b.bugs[i].conn = conn
	return &b.bugs[i], nil
}
