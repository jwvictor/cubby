package users

type User struct {
	Id          string
	Email       string
	DisplayName string
}

type UserProvider interface {
	Authenticate(userEmail string, userPass string) (*User, error)
	SignUp(userEmail, userPass, displayName string) (*User, error)
	GetById(id string) (*User, error)
	GetByDisplayName(name string) (*User, error)
}
