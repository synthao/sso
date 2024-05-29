package domain

type User struct {
	ID       int
	Nickname string
	Password string
}

type Token struct {
	ID     int
	UserID int
}
