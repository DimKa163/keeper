package core

type User struct {
	ID       string
	Username string
	Password []byte
	Salt     []byte
}
