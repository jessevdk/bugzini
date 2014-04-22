package bugzilla

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

func (c *Conn) Users() Users {
	return Users{
		conn: c,
	}
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
		Id int `xmlrpc:"id"`
	}

	if err := u.conn.Call("User.login", args, &ret); err != nil {
		return User{}, err
	}

	return u.Get(ret.Id)
}

func (u Users) Logout() error {
	var ret struct{}
	var args struct{}

	return u.conn.Call("User.logout", args, &ret)
}

func (u Users) Get(id int) (User, error) {
	args := struct {
		Ids []int `xmlrpc:"ids"`
	}{
		Ids: []int{id},
	}

	var ret struct {
		Users []User `xmlrpc:"users"`
	}

	if err := u.conn.Call("User.get", args, &ret); err != nil {
		return User{}, err
	}

	return ret.Users[0], nil
}
