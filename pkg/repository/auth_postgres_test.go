package repository

import (
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	storage "github.com/mahadeva604/audio-storage"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAuthPostgres_CreateUser(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	r := NewAuthPostgres(db)
	type mockBehavior func(user storage.User, userId int)

	testTable := []struct {
		name            string
		user            storage.User
		mockBehavior    mockBehavior
		expectedUserId  int
		expectedErr     bool
		expectedErrType error
	}{
		{
			name: "OK",

			user: storage.User{
				Name:     "user_1",
				Username: "User One",
				Password: "secret",
			},
			mockBehavior: func(user storage.User, userId int) {
				rows := sqlmock.NewRows([]string{"user_id"}).AddRow(userId)
				mock.ExpectQuery("INSERT INTO users").WithArgs(user.Name, user.Username, user.Password).WillReturnRows(rows)
			},
			expectedUserId: 1,
		},
		{
			name: "User exists",
			user: storage.User{
				Name:     "user_1",
				Username: "User One",
				Password: "secret",
			},
			mockBehavior: func(user storage.User, userId int) {
				mock.ExpectQuery("INSERT INTO users").WithArgs(user.Name, user.Username, user.Password).WillReturnError(&pq.Error{Code: "23505"})
			},
			expectedErr:     true,
			expectedErrType: storage.UserExists,
		},
		{
			name: "Error",
			user: storage.User{
				Name:     "user_1",
				Username: "User One",
				Password: "secret",
			},
			mockBehavior: func(user storage.User, userId int) {
				mock.ExpectQuery("INSERT INTO users").WithArgs(user.Name, user.Username, user.Password).WillReturnError(errors.New("query error"))
			},
			expectedErr:     true,
			expectedErrType: errors.New("query error"),
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.user, testCase.expectedUserId)

			gotUserId, err := r.CreateUser(testCase.user)
			if testCase.expectedErr {
				assert.Error(t, err)
				if testCase.expectedErrType != nil {
					assert.Equal(t, testCase.expectedErrType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedUserId, gotUserId)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthPostgres_GetUser(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	r := NewAuthPostgres(db)
	type mockBehavior func(userName, password string, userId int)

	testTable := []struct {
		name            string
		userName        string
		password        string
		userId          int
		mockBehavior    mockBehavior
		expectedData    storage.User
		expectedErr     bool
		expectedErrType error
	}{
		{
			name:     "OK",
			userName: "user_1",
			password: "secret",
			userId:   1,
			mockBehavior: func(userName, password string, userId int) {
				rows := sqlmock.NewRows([]string{"user_id"}).AddRow(userId)
				mock.ExpectQuery("select user_id FROM users").WithArgs(userName, password).WillReturnRows(rows)
			},
			expectedData: storage.User{
				Id: 1,
			},
		},
		{
			name:     "User not found",
			userName: "user_1",
			password: "secret",
			mockBehavior: func(userName, password string, userId int) {
				rows := sqlmock.NewRows([]string{"user_id"})
				mock.ExpectQuery("select user_id FROM users").WithArgs(userName, password).WillReturnRows(rows)
			},
			expectedErr:     true,
			expectedErrType: storage.UserNotFound,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.userName, testCase.password, testCase.userId)

			gotData, err := r.GetUser(testCase.userName, testCase.password)
			if testCase.expectedErr {
				assert.Error(t, err)
				if testCase.expectedErrType != nil {
					assert.Equal(t, testCase.expectedErrType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedData, gotData)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthPostgres_SetRefreshToken(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	r := NewAuthPostgres(db)
	type mockBehavior func(userId int, refreshToken string, refreshTokenTTL time.Duration)

	testTable := []struct {
		name            string
		userId          int
		refreshToken    string
		refreshTokenTTL time.Duration
		mockBehavior    mockBehavior
		expectedErr     bool
	}{
		{
			name:            "OK",
			userId:          1,
			refreshToken:    "refresh_token_string",
			refreshTokenTTL: time.Second,
			mockBehavior: func(userId int, refreshToken string, refreshTokenTTL time.Duration) {
				query := fmt.Sprintf(`INSERT INTO refresh_tokens VALUES \(\$1, \$2, now\(\) \+ interval \'%d seconds\'\)
											ON CONFLICT \(user_id\) DO UPDATE SET (.+)`, int64(refreshTokenTTL.Seconds()))
				mock.ExpectExec(query).WithArgs(userId, refreshToken).WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:            "Error",
			userId:          1,
			refreshToken:    "refresh_token_string",
			refreshTokenTTL: time.Second,
			mockBehavior: func(userId int, refreshToken string, refreshTokenTTL time.Duration) {
				query := fmt.Sprintf(`INSERT INTO refresh_tokens VALUES \(\$1, \$2, now\(\) \+ interval \'%d seconds\'\)
											ON CONFLICT \(user_id\) DO UPDATE SET (.+)`, int64(refreshTokenTTL.Seconds()))
				mock.ExpectExec(query).WithArgs(userId, refreshToken).WillReturnError(errors.New("query error"))
			},
			expectedErr: true,
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.userId, testCase.refreshToken, testCase.refreshTokenTTL)

			err := r.SetRefreshToken(testCase.userId, testCase.refreshToken, testCase.refreshTokenTTL)
			if testCase.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthPostgres_UpdateRefreshToken(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	r := NewAuthPostgres(db)
	type mockBehavior func(userId int, oldRefreshToken string, newRefreshToken string, refreshTokenTTL time.Duration)

	testTable := []struct {
		name            string
		userId          int
		oldRefreshToken string
		newRefreshToken string
		refreshTokenTTL time.Duration
		mockBehavior    mockBehavior
		expectedErr     bool
		expectedErrType error
	}{
		{
			name:            "OK",
			userId:          1,
			oldRefreshToken: "old_refresh_token",
			newRefreshToken: "new_refresh_token",
			refreshTokenTTL: time.Second,
			mockBehavior: func(userId int, oldRefreshToken string, newRefreshToken string, refreshTokenTTL time.Duration) {
				query := fmt.Sprintf(`UPDATE refresh_tokens SET refresh_token = \$1, expires_in = now\(\) \+ interval '%d seconds'
								WHERE refresh_token = \$2 AND expires_in > now\(\) RETURNING user_id`, int64(refreshTokenTTL.Seconds()))
				rows := sqlmock.NewRows([]string{"user_id"}).AddRow(userId)
				mock.ExpectQuery(query).WithArgs(newRefreshToken, oldRefreshToken).WillReturnRows(rows)
			},
		},
		{
			name:            "Error no rows",
			userId:          1,
			oldRefreshToken: "old_refresh_token",
			newRefreshToken: "new_refresh_token",
			refreshTokenTTL: time.Second,
			mockBehavior: func(userId int, oldRefreshToken string, newRefreshToken string, refreshTokenTTL time.Duration) {
				query := fmt.Sprintf(`UPDATE refresh_tokens SET refresh_token = \$1, expires_in = now\(\) \+ interval '%d seconds'
								WHERE refresh_token = \$2 AND expires_in > now\(\) RETURNING user_id`, int64(refreshTokenTTL.Seconds()))
				rows := sqlmock.NewRows([]string{"user_id"})
				mock.ExpectQuery(query).WithArgs(newRefreshToken, oldRefreshToken).WillReturnRows(rows)
			},
			expectedErr:     true,
			expectedErrType: storage.WrongRefreshToken,
		},
		{
			name:            "Error query",
			userId:          1,
			oldRefreshToken: "old_refresh_token",
			newRefreshToken: "new_refresh_token",
			refreshTokenTTL: time.Second,
			mockBehavior: func(userId int, oldRefreshToken string, newRefreshToken string, refreshTokenTTL time.Duration) {
				query := fmt.Sprintf(`UPDATE refresh_tokens SET refresh_token = \$1, expires_in = now\(\) \+ interval '%d seconds'
								WHERE refresh_token = \$2 AND expires_in > now\(\) RETURNING user_id`, int64(refreshTokenTTL.Seconds()))
				mock.ExpectQuery(query).WithArgs(newRefreshToken, oldRefreshToken).WillReturnError(errors.New("query error"))
			},
			expectedErr:     true,
			expectedErrType: errors.New("query error"),
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.userId, testCase.oldRefreshToken, testCase.newRefreshToken, testCase.refreshTokenTTL)

			gotUserId, err := r.UpdateRefreshToken(testCase.oldRefreshToken, testCase.newRefreshToken, testCase.refreshTokenTTL)
			if testCase.expectedErr {
				assert.Error(t, err)
				if testCase.expectedErrType != nil {
					assert.Equal(t, testCase.expectedErrType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.userId, gotUserId)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
