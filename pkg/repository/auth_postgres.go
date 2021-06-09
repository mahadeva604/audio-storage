package repository

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	storage "github.com/mahadeva604/audio-storage"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) CreateUser(user storage.User) (int, error) {
	var userId int

	query := fmt.Sprintf("INSERT INTO %s (name, username, password_hash) VALUES($1, $2, crypt($3, gen_salt('bf'))) RETURNING user_id", usersTable)
	err := r.db.Get(&userId, query, user.Name, user.Username, user.Password)

	if _, ok := err.(*pq.Error); ok && err.(*pq.Error).Code == "23505" {
		err = storage.UserExists
	}

	return userId, err
}

func (r *AuthPostgres) GetUser(username, password string) (storage.User, error) {
	var user storage.User
	query := fmt.Sprintf("select user_id FROM %s WHERE username = $1 and password_hash = crypt( $2, password_hash)", usersTable)

	err := r.db.Get(&user, query, username, password)

	if err == sql.ErrNoRows {
		err = storage.UserNotFound
	}

	return user, err
}
