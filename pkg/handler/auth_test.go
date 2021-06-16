package handler

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	storage "github.com/mahadeva604/audio-storage"
	"github.com/mahadeva604/audio-storage/pkg/service"
	mock_service "github.com/mahadeva604/audio-storage/pkg/service/mocks"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestHandler_sigUp(t *testing.T) {
	type mockBehavior func(s *mock_service.MockAuthorization, user storage.User)

	testTable := []struct {
		name                string
		inputBody           string
		inputUser           storage.User
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"name":"User 1","username":"user_1","password":"qwerty_1"}`,
			inputUser: storage.User{
				Name:     "User 1",
				Username: "user_1",
				Password: "qwerty_1",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user storage.User) {
				s.EXPECT().CreateUser(user).Return(1, nil)
			},
			expectedStatusCode:  200,
			expectedRequestBody: `{"id":1}`,
		},
		{
			name:                "Empty field",
			inputBody:           "",
			mockBehavior:        func(s *mock_service.MockAuthorization, user storage.User) {},
			expectedStatusCode:  400,
			expectedRequestBody: `{"message":"invalid input body"}`,
		},
		{
			name:      "User exists",
			inputBody: `{"name":"User 1","username":"user_1","password":"qwerty_1"}`,
			inputUser: storage.User{
				Name:     "User 1",
				Username: "user_1",
				Password: "qwerty_1",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user storage.User) {
				s.EXPECT().CreateUser(user).Return(0, storage.UserExists)
			},
			expectedStatusCode:  400,
			expectedRequestBody: `{"message":"user exists"}`,
		},
		{
			name:      "Servise fail",
			inputBody: `{"name":"User 1","username":"user_1","password":"qwerty_1"}`,
			inputUser: storage.User{
				Name:     "User 1",
				Username: "user_1",
				Password: "qwerty_1",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user storage.User) {
				s.EXPECT().CreateUser(user).Return(0, errors.New("service error"))
			},
			expectedStatusCode:  500,
			expectedRequestBody: `{"message":"service error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			auth := mock_service.NewMockAuthorization(c)
			testCase.mockBehavior(auth, testCase.inputUser)

			services := &service.Service{Authorization: auth}
			handler := NewHandler(services)

			r := gin.New()
			r.POST("/sign-up", handler.signUp)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sign-up", bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}

func TestHandler_sigIn(t *testing.T) {
	type mockBehavior func(s *mock_service.MockAuthorization, user signInInput, userId int)

	testTable := []struct {
		name                 string
		userId               int
		inputBody            string
		inputUser            signInInput
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			userId:    1,
			inputBody: `{"username":"user_1","password":"qwerty_1"}`,
			inputUser: signInInput{
				Username: "user_1",
				Password: "qwerty_1",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user signInInput, userId int) {
				s.EXPECT().GenerateAccessToken(user.Username, user.Password).Return(userId, "token_string", nil)
				s.EXPECT().GenerateRefreshToken(userId).Return("refresh_token_string", nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"token":"token_string","refresh_token":"refresh_token_string"}`,
		},
		{
			name:                 "Empty field",
			inputBody:            "",
			mockBehavior:         func(s *mock_service.MockAuthorization, user signInInput, userId int) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid input body"}`,
		},
		{
			name:      "Service fail 1",
			userId:    1,
			inputBody: `{"username":"user_1","password":"qwerty_1"}`,
			inputUser: signInInput{
				Username: "user_1",
				Password: "qwerty_1",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user signInInput, userId int) {
				s.EXPECT().GenerateAccessToken(user.Username, user.Password).Return(userId, "token_string", nil)
				s.EXPECT().GenerateRefreshToken(userId).Return("", errors.New("service error 1"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"service error 1"}`,
		},
		{
			name:      "Service fail 2",
			inputBody: `{"username":"user_1","password":"qwerty_1"}`,
			inputUser: signInInput{
				Username: "user_1",
				Password: "qwerty_1",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user signInInput, userId int) {
				s.EXPECT().GenerateAccessToken(user.Username, user.Password).Return(0, "", errors.New("service error 2"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"service error 2"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			auth := mock_service.NewMockAuthorization(c)
			testCase.mockBehavior(auth, testCase.inputUser, testCase.userId)

			services := &service.Service{Authorization: auth}
			handler := NewHandler(services)

			r := gin.New()
			r.POST("/sign-in", handler.signIn)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sign-in", bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_refreshTokens(t *testing.T) {
	type mockBehavior func(s *mock_service.MockAuthorization, refreshToken refreshTokensInput, userId int)

	testTable := []struct {
		name                 string
		userId               int
		inputBody            string
		refreshToken         refreshTokensInput
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			userId:    1,
			inputBody: `{"refresh_token":"refresh_token"}`,
			refreshToken: refreshTokensInput{
				RefreshTooken: "refresh_token",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, refreshToken refreshTokensInput, userId int) {
				s.EXPECT().UpdateRefreshToken(refreshToken.RefreshTooken).Return(userId, "new_refresh_token_string", nil)
				s.EXPECT().UpdateAccessToken(userId).Return("new_token_string", nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"refresh_token":"new_refresh_token_string","token":"new_token_string"}`,
		},
		{
			name: "Error invalid body",
			mockBehavior: func(s *mock_service.MockAuthorization, refreshToken refreshTokensInput, userId int) {
			},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid input body"}`,
		},
		{
			name:      "Error service 1",
			userId:    1,
			inputBody: `{"refresh_token":"refresh_token"}`,
			refreshToken: refreshTokensInput{
				RefreshTooken: "refresh_token",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, refreshToken refreshTokensInput, userId int) {
				s.EXPECT().UpdateRefreshToken(refreshToken.RefreshTooken).Return(0, "", errors.New("service error 1"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"service error 1"}`,
		},
		{
			name:      "Error service 2",
			userId:    1,
			inputBody: `{"refresh_token":"refresh_token"}`,
			refreshToken: refreshTokensInput{
				RefreshTooken: "refresh_token",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, refreshToken refreshTokensInput, userId int) {
				s.EXPECT().UpdateRefreshToken(refreshToken.RefreshTooken).Return(userId, "new_refresh_token_string", nil)
				s.EXPECT().UpdateAccessToken(userId).Return("", errors.New("service error 2"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"service error 2"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			auth := mock_service.NewMockAuthorization(c)
			testCase.mockBehavior(auth, testCase.refreshToken, testCase.userId)

			services := &service.Service{Authorization: auth}
			handler := NewHandler(services)

			r := gin.New()
			r.POST("/refresh", handler.refreshTokens)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/refresh", bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}
