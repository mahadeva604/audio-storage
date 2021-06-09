package repository

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	storage "github.com/mahadeva604/audio-storage"
	"github.com/stretchr/testify/assert"
	"testing"
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
