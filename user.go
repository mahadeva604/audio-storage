package storage

type User struct {
	Id       int    `json:"-" db:"user_id"`
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
