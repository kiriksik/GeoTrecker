package models

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"hash"`
	Role     string `json:"role"`
}
