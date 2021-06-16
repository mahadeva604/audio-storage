package repository

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	storage "github.com/mahadeva604/audio-storage"
	"time"
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

func (r *AuthPostgres) SetRefreshToken(userId int, refreshToken string, refreshTokenTTL time.Duration) error {
	query := fmt.Sprintf(`INSERT INTO %s VALUES ($1, $2, now() + interval '%d seconds')
								ON CONFLICT (user_id) DO
								UPDATE SET refresh_token = excluded.refresh_token,
								expires_in = excluded.expires_in`, tokenTable, int64(refreshTokenTTL.Seconds()))
	_, err := r.db.Exec(query, userId, refreshToken)
	return err
}

func (r *AuthPostgres) UpdateRefreshToken(oldRefreshToken string, newRefreshToken string, refreshTokenTTL time.Duration) (int, error) {
	var userId int
	query := fmt.Sprintf(`UPDATE %s
								SET refresh_token = $1, expires_in = now() + interval '%d seconds'
								WHERE refresh_token = $2 AND expires_in > now() RETURNING user_id`, tokenTable, int64(refreshTokenTTL.Seconds()))
	err := r.db.Get(&userId, query, newRefreshToken, oldRefreshToken)

	if err == sql.ErrNoRows {
		err = storage.WrongRefreshToken
	}

	if err != nil {
		return 0, err
	}

	return userId, nil
}
