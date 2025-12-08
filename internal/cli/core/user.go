package core

type User struct {
	ID       string
	Username string
	Password []byte
	Salt     []byte
}

type Server struct {
	ID       int32
	Address  string
	Login    string
	Password string
	Active   bool
}
