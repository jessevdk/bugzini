package bugzilla

import (
	"errors"
	"strconv"
)

type Users struct {
	conn *Conn
}

type User struct {
	Id           int    `xmlrpc:"id" json:"id"`
	RealName     string `xmlrpc:"real_name" json:"real_name"`
	Email        string `xmlrpc:"email" json:"email"`
	Name         string `xmlrpc:"name" json:"name"`
	CanLogin     bool   `xmlrpc:"can_login" json:"can_login"`
	EmailEnabled bool   `xmlrpc:"email_enabled" json:"email_enabled"`
}

type UserAuth struct {
	Id    int
	Token string
}

var currentUser *User
var AuthUser *UserAuth

func (c *Conn) Users() Users {
	return Users{
		conn: c,
	}
}

func CurrentUser() *User {
	return currentUser
}

func (u *UserAuth) GetToken() string {
	if u == nil {
		return ""
	}

	return u.Token
}

func (u Users) CheckCurrentLogin() (User, error) {
	if AuthUser != nil {
		us, err := u.Get(AuthUser.Id)

		if err == nil {
			currentUser = &us
		}

		return us, err
	}

	cookies := u.conn.Client.Cookies()

	for _, c := range cookies {
		if c.Name == "Bugzilla_login" {
			id, err := strconv.ParseInt(c.Value, 10, 64)

			if err != nil {
				return User{}, err
			}

			us, err := u.Get(int(id))

			if err == nil {
				currentUser = &us
			}

			return us, err
		}
	}

	return User{}, errors.New("Not logged in")
}

func (u Users) Login(user string, passwd string) (User, error) {
	args := struct {
		Login    string `xmlrpc:"login"`
		Password string `xmlrpc:"password"`
	}{
		Login:    user,
		Password: passwd,
	}

	var ret struct {
		Id    int    `xmlrpc:"id"`
		Token string `xmlrpc:"token"`
	}

	if err := u.conn.Call("User.login", args, &ret); err != nil {
		return User{}, err
	}

	us, err := u.getWithToken(ret.Id, ret.Token)

	if err != nil {
		return us, err
	}

	AuthUser = &UserAuth{
		Id:    ret.Id,
		Token: ret.Token,
	}

	currentUser = &us
	return us, err
}

func (u Users) Logout() error {
	var ret struct{}

	args := struct {
		Token string `xmlrpc:"token"`
	}{
		Token: AuthUser.GetToken(),
	}

	if err := u.conn.Call("User.logout", args, &ret); err != nil {
		return err
	}

	AuthUser = nil
	currentUser = nil
	return nil
}

func (u Users) getWithToken(id int, token string) (User, error) {
	args := struct {
		Ids   []int  `xmlrpc:"ids"`
		Token string `xmlrpc:"token"`
	}{
		Ids:   []int{id},
		Token: token,
	}

	var ret struct {
		Users []User `xmlrpc:"users"`
	}

	if err := u.conn.Call("User.get", args, &ret); err != nil {
		return User{}, err
	}

	return ret.Users[0], nil
}

func (u Users) Get(id int) (User, error) {
	return u.getWithToken(id, AuthUser.GetToken())
}
