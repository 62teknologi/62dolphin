package models

type User struct {
	Id       uint   `json:"id" gorm:"primary_key"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
