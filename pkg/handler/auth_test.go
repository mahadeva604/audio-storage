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
	type mockBehavior func(s *mock_service.MockAuthorization, user signInInput)

	testTable := []struct {
		name                 string
		inputBody            string
		inputUser            signInInput
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			inputBody: `{"username":"user_1","password":"qwerty_1"}`,
			inputUser: signInInput{
				Username: "user_1",
				Password: "qwerty_1",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user signInInput) {
				s.EXPECT().GenerateToken(user.Username, user.Password).Return("token_string", nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"token":"token_string"}`,
		},
		{
			name:                 "Empty field",
			inputBody:            "",
			mockBehavior:         func(s *mock_service.MockAuthorization, user signInInput) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid input body"}`,
		},
		{
			name:      "Servise fail",
			inputBody: `{"username":"user_1","password":"qwerty_1"}`,
			inputUser: signInInput{
				Username: "user_1",
				Password: "qwerty_1",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user signInInput) {
				s.EXPECT().GenerateToken(user.Username, user.Password).Return("", errors.New("service error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"service error"}`,
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
			r.POST("/sign-in", handler.signIn)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sign-in", bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}
