package handler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	storage "github.com/mahadeva604/audio-storage"
	"github.com/mahadeva604/audio-storage/pkg/service"
	mock_service "github.com/mahadeva604/audio-storage/pkg/service/mocks"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestHandler_shareAudio(t *testing.T) {
	type mockBehavior func(s *mock_service.MockShare, userId int, audioId int, shareTo int)

	testTable := []struct {
		name                 string
		userId               int
		audioId              int
		wrongAudioID         bool
		inputBody            string
		inputShare           storage.ShareInput
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			userId:    1,
			audioId:   1,
			inputBody: `{"share_to":2}`,
			inputShare: storage.ShareInput{
				ShareTo: 2,
			},
			mockBehavior: func(s *mock_service.MockShare, userId int, audioId int, input int) {
				s.EXPECT().ShareAudio(userId, audioId, input).Return(nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"status":"ok"}`,
		},
		{
			name:                 "User not found",
			mockBehavior:         func(s *mock_service.MockShare, userId int, audioId int, input int) {},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"user id not found"}`,
		},
		{
			name:                 "Invalid audio id",
			userId:               1,
			audioId:              0,
			mockBehavior:         func(s *mock_service.MockShare, userId int, audioId int, input int) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid audio id param"}`,
		},
		{
			name:                 "Invalid input",
			userId:               1,
			audioId:              1,
			mockBehavior:         func(s *mock_service.MockShare, userId int, audioId int, input int) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid input body"}`,
		},
		{
			name:      "Service error",
			userId:    1,
			audioId:   1,
			inputBody: `{"share_to":2}`,
			inputShare: storage.ShareInput{
				ShareTo: 2,
			},
			mockBehavior: func(s *mock_service.MockShare, userId int, audioId int, input int) {
				s.EXPECT().ShareAudio(userId, audioId, input).Return(errors.New("service error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"service error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			share := mock_service.NewMockShare(c)
			testCase.mockBehavior(share, testCase.userId, testCase.audioId, testCase.inputShare.ShareTo)

			services := &service.Service{Share: share}
			handler := NewHandler(services)

			r := gin.New()
			if testCase.userId != 0 {
				r.POST("/share/:id", func(c *gin.Context) {
					c.Set(userCtx, testCase.userId)
				}, handler.shareAudio)
			} else {
				r.POST("/share/:id", handler.shareAudio)
			}

			w := httptest.NewRecorder()
			target := fmt.Sprintf("/share/%d", testCase.audioId)
			if testCase.audioId == 0 {
				target = "/share/wrong_id"
			}
			req := httptest.NewRequest("POST", target, bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_unshareAudio(t *testing.T) {
	type mockBehavior func(s *mock_service.MockShare, userId int, audioId int, shareTo int)

	testTable := []struct {
		name                 string
		userId               int
		audioId              int
		inputBody            string
		inputShare           storage.ShareInput
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			userId:    1,
			audioId:   1,
			inputBody: `{"share_to":2}`,
			inputShare: storage.ShareInput{
				ShareTo: 2,
			},
			mockBehavior: func(s *mock_service.MockShare, userId int, audioId int, shareTo int) {
				s.EXPECT().UnshareAudio(userId, audioId, shareTo).Return(nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"status":"ok"}`,
		},
		{
			name:                 "User not found",
			userId:               0,
			mockBehavior:         func(s *mock_service.MockShare, userId int, audioId int, shareTo int) {},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"user id not found"}`,
		},
		{
			name:                 "Invalid audio id",
			userId:               1,
			audioId:              0,
			mockBehavior:         func(s *mock_service.MockShare, userId int, audioId int, shareTo int) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid audio id param"}`,
		},
		{
			name:                 "Invalid input",
			userId:               1,
			audioId:              1,
			mockBehavior:         func(s *mock_service.MockShare, userId int, audioId int, input int) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid input body"}`,
		},
		{
			name:      "Service error",
			userId:    1,
			audioId:   1,
			inputBody: `{"share_to":2}`,
			inputShare: storage.ShareInput{
				ShareTo: 2,
			},
			mockBehavior: func(s *mock_service.MockShare, userId int, audioId int, input int) {
				s.EXPECT().UnshareAudio(userId, audioId, input).Return(errors.New("service error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"service error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			share := mock_service.NewMockShare(c)
			testCase.mockBehavior(share, testCase.userId, testCase.audioId, testCase.inputShare.ShareTo)

			services := &service.Service{Share: share}
			handler := NewHandler(services)

			r := gin.New()
			if testCase.userId != 0 {
				r.POST("/unshare/:id", func(c *gin.Context) {
					c.Set(userCtx, testCase.userId)
				}, handler.unshareAudio)
			} else {
				r.POST("/unshare/:id", handler.unshareAudio)
			}

			w := httptest.NewRecorder()
			target := fmt.Sprintf("/unshare/%d", testCase.audioId)
			if testCase.audioId == 0 {
				target = "/unshare/wrong_id"
			}
			req := httptest.NewRequest("POST", target, bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_getSharedAudio(t *testing.T) {
	type mockBehavior func(s *mock_service.MockShare, input storage.ShareListParam)

	offset, limit := 0, 2

	testTable := []struct {
		name                 string
		offset               string
		limit                string
		inputShare           storage.ShareListParam
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:   "OK",
			offset: strconv.Itoa(offset),
			limit:  strconv.Itoa(limit),
			inputShare: storage.ShareListParam{
				Offset: &offset,
				Limit:  &limit,
			},
			mockBehavior: func(s *mock_service.MockShare, input storage.ShareListParam) {
				s.EXPECT().GetSharedList(input).Return(storage.ShareListJson{
					Count: 100,
					Users: []storage.ShareListCount{
						{
							UserId:     1,
							Name:       "user 1",
							ShareCount: 10,
						},
						{
							UserId:     2,
							Name:       "user 2",
							ShareCount: 20,
						},
					},
				}, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"total_count":100,"users":[{"id":1,"name":"user 1","shared_records":10},{"id":2,"name":"user 2","shared_records":20}]}`,
		},
		{
			name:                 "Invalid input",
			offset:               "bad field",
			limit:                strconv.Itoa(limit),
			mockBehavior:         func(s *mock_service.MockShare, input storage.ShareListParam) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid input body"}`,
		},
		{
			name:   "Service fail",
			offset: strconv.Itoa(offset),
			limit:  strconv.Itoa(limit),
			inputShare: storage.ShareListParam{
				Offset: &offset,
				Limit:  &limit,
			},
			mockBehavior: func(s *mock_service.MockShare, input storage.ShareListParam) {
				s.EXPECT().GetSharedList(input).Return(storage.ShareListJson{}, errors.New("service fail"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"service fail"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			share := mock_service.NewMockShare(c)
			testCase.mockBehavior(share, testCase.inputShare)

			services := &service.Service{Share: share}
			handler := NewHandler(services)

			r := gin.New()
			r.GET("/shares", handler.getSharedAudio)

			w := httptest.NewRecorder()
			params := url.Values{"offset": {testCase.offset}, "limit": {testCase.limit}}.Encode()
			req := httptest.NewRequest("GET", "/shares?"+params, nil)

			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}
