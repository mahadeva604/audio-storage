package handler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	storage "github.com/mahadeva604/audio-storage"
	"github.com/mahadeva604/audio-storage/pkg/service"
	mock_service "github.com/mahadeva604/audio-storage/pkg/service/mocks"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestHandler_getAllAudio(t *testing.T) {
	type mockBehavior func(s *mock_service.MockAudio, userId int, audioParam storage.AudioListParam)

	offset, limit := 0, 10

	testTable := []struct {
		name                 string
		offset               string
		limit                string
		orderType            string
		userId               int
		audio                storage.AudioListParam
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			offset:    strconv.Itoa(offset),
			limit:     strconv.Itoa(limit),
			orderType: "owner",
			userId:    1,
			audio: storage.AudioListParam{
				Offset:    &offset,
				Limit:     &limit,
				OrderType: "owner",
			},
			mockBehavior: func(s *mock_service.MockAudio, userId int, audioParam storage.AudioListParam) {
				s.EXPECT().GetAudioList(userId, audioParam).Return(storage.AudioListJson{
					TotalCount: 10,
					Records: []storage.AudioList{
						{
							Id:      1,
							Title:   "title 1",
							IsOwner: true,
							Owner:   1,
							Name:    "user 1",
							Shares: &[]storage.ShareList{
								{
									UserId: 2,
									Name:   "user 2",
								},
							},
						},
						{
							Id:      2,
							Title:   "title 2",
							IsOwner: true,
							Owner:   1,
							Name:    "user 1",
						},
					},
				}, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"total_count":10,"records":[{"id":1,"name":"title 1","is_owner":true,"owner_id":1,"owner_name":"user 1","shared_to":[{"id":2,"name":"user 2"}]},{"id":2,"name":"title 2","is_owner":true,"owner_id":1,"owner_name":"user 1"}]}`,
		},
		{
			name:                 "User no found",
			offset:               strconv.Itoa(offset),
			limit:                strconv.Itoa(limit),
			orderType:            "owner",
			mockBehavior:         func(s *mock_service.MockAudio, userId int, audioParam storage.AudioListParam) {},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"user id not found"}`,
		},
		{
			name:      "Invalid param",
			offset:    strconv.Itoa(offset),
			limit:     strconv.Itoa(limit),
			orderType: "UNKNOWN",
			userId:    1,
			audio: storage.AudioListParam{
				Offset:    &offset,
				Limit:     &limit,
				OrderType: "owner",
			},
			mockBehavior:         func(s *mock_service.MockAudio, userId int, audioParam storage.AudioListParam) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid query"}`,
		},
		{
			name:      "Service error",
			offset:    strconv.Itoa(offset),
			limit:     strconv.Itoa(limit),
			orderType: "owner",
			userId:    1,
			audio: storage.AudioListParam{
				Offset:    &offset,
				Limit:     &limit,
				OrderType: "owner",
			},
			mockBehavior: func(s *mock_service.MockAudio, userId int, audioParam storage.AudioListParam) {
				s.EXPECT().GetAudioList(userId, gomock.AssignableToTypeOf(audioParam)).Return(storage.AudioListJson{}, errors.New("service error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"service error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			audio := mock_service.NewMockAudio(c)
			testCase.mockBehavior(audio, testCase.userId, testCase.audio)

			services := &service.Service{Audio: audio}
			handler := NewHandler(services)

			r := gin.New()

			if testCase.userId != 0 {
				r.GET("/", func(c *gin.Context) {
					c.Set(userCtx, testCase.userId)
				}, handler.getAllAudio)
			} else {
				r.GET("/", handler.getAllAudio)
			}

			w := httptest.NewRecorder()
			params := url.Values{"offset": {testCase.offset}, "limit": {testCase.limit}, "order_type": {testCase.orderType}}.Encode()
			req := httptest.NewRequest("GET", "/?"+params, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}
func TestHandler_addDescription(t *testing.T) {
	type mockBehavior func(s *mock_service.MockAudio, userId, audioId int, audioParam storage.UpdateAudio)

	title := "new audio"
	duration := 67

	testTable := []struct {
		name      string
		inputBody string
		userId    int
		audioId   int

		audioParam           storage.UpdateAudio
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			inputBody: fmt.Sprintf(`{"title":"%s","duration":%d}`, title, duration),
			userId:    1,
			audioId:   1,
			audioParam: storage.UpdateAudio{
				Title:    &title,
				Duration: &duration,
			},
			mockBehavior: func(s *mock_service.MockAudio, userId, audioId int, audioParam storage.UpdateAudio) {
				s.EXPECT().AddDescription(userId, audioId, audioParam).Return(nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"status":"ok"}`,
		},
		{
			name:                 "User not found",
			userId:               0,
			mockBehavior:         func(s *mock_service.MockAudio, userId, audioId int, audioParam storage.UpdateAudio) {},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"user id not found"}`,
		},
		{
			name:                 "Invalid audio id",
			userId:               1,
			audioId:              0,
			mockBehavior:         func(s *mock_service.MockAudio, userId, audioId int, audioParam storage.UpdateAudio) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid audio id param"}`,
		},
		{
			name:                 "Invalid input",
			inputBody:            ``,
			userId:               1,
			audioId:              1,
			mockBehavior:         func(s *mock_service.MockAudio, userId, audioId int, audioParam storage.UpdateAudio) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid input body"}`,
		},
		{
			name:                 "Empty input",
			inputBody:            `{}`,
			userId:               1,
			audioId:              1,
			mockBehavior:         func(s *mock_service.MockAudio, userId, audioId int, audioParam storage.UpdateAudio) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"update structure has no values"}`,
		},
		{
			name:      "Service fail",
			inputBody: fmt.Sprintf(`{"title":"%s","duration":%d}`, title, duration),
			userId:    1,
			audioId:   1,
			audioParam: storage.UpdateAudio{
				Title:    &title,
				Duration: &duration,
			},
			mockBehavior: func(s *mock_service.MockAudio, userId, audioId int, audioParam storage.UpdateAudio) {
				s.EXPECT().AddDescription(userId, audioId, audioParam).Return(errors.New("service fail"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"service fail"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			audio := mock_service.NewMockAudio(c)
			testCase.mockBehavior(audio, testCase.userId, testCase.audioId, testCase.audioParam)

			services := &service.Service{Audio: audio}
			handler := NewHandler(services)

			r := gin.New()

			if testCase.userId != 0 {
				r.POST("/audio/:id", func(c *gin.Context) {
					c.Set(userCtx, testCase.userId)
				}, handler.addDescription)
			} else {
				r.POST("/audio/:id", handler.addDescription)
			}

			w := httptest.NewRecorder()
			url := fmt.Sprintf("/audio/%d", testCase.audioId)
			if testCase.audioId == 0 {
				url = "/audio/wrong_id"
			}
			req := httptest.NewRequest("POST", url, bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_uploadAudio(t *testing.T) {
	type mockBehavior func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId int)

	testTable := []struct {
		name                 string
		userId               int
		wrongFormKey         bool
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:   "OK",
			userId: 1,
			mockBehavior: func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId int) {
				s1.EXPECT().UploadFile(userId, gomock.AssignableToTypeOf("")).Return(1, nil)
				ioInterface := reflect.TypeOf((*io.ReadCloser)(nil)).Elem()
				s2.EXPECT().StoreFile(gomock.AssignableToTypeOf(uuid.UUID{}), gomock.AssignableToTypeOf(ioInterface)).Return(nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"id":1}`,
		},
		{
			name:                 "Wrong form key",
			userId:               1,
			wrongFormKey:         true,
			mockBehavior:         func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId int) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"http: no such file"}`,
		},
		{
			name:                 "User not found",
			mockBehavior:         func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId int) {},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"user id not found"}`,
		},
		{
			name:   "Save file error",
			userId: 1,
			mockBehavior: func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId int) {
				ioInterface := reflect.TypeOf((*io.ReadCloser)(nil)).Elem()
				s2.EXPECT().StoreFile(gomock.AssignableToTypeOf(uuid.UUID{}), gomock.AssignableToTypeOf(ioInterface)).Return(errors.New("save file error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"save file error"}`,
		},
		{
			name:   "Store data to DB error",
			userId: 1,
			mockBehavior: func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId int) {
				s1.EXPECT().UploadFile(userId, gomock.AssignableToTypeOf("")).Return(0, errors.New("store data to DB error"))
				ioInterface := reflect.TypeOf((*io.ReadCloser)(nil)).Elem()
				s2.EXPECT().StoreFile(gomock.AssignableToTypeOf(uuid.UUID{}), gomock.AssignableToTypeOf(ioInterface)).Return(nil)
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"store data to DB error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			audio := mock_service.NewMockAudio(c)
			strg := mock_service.NewMockStorage(c)

			testCase.mockBehavior(audio, strg, testCase.userId)

			services := &service.Service{Audio: audio, Storage: strg}
			handler := NewHandler(services)

			r := gin.New()
			if testCase.userId != 0 {
				r.POST("/upload", func(c *gin.Context) {
					c.Set(userCtx, testCase.userId)
				}, handler.uploadAudio)
			} else {
				r.POST("/upload", handler.uploadAudio)
			}

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			formKey := "file"
			if testCase.wrongFormKey {
				formKey = "no_file"
			}
			part, err := writer.CreateFormFile(formKey, "test.file")
			if err != nil {
				t.Error(err)
			}
			_, err = io.Copy(part, bytes.NewBuffer([]byte("file content")))

			writer.Close()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_downloadAudio(t *testing.T) {
	type mockBehavior func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId, audioId int, fileId uuid.UUID, fileContent string)

	testTable := []struct {
		name                 string
		userId               int
		audioId              int
		fileId               uuid.UUID
		fileContent          string
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedLenBody      int
		expectedResponseBody string
	}{
		{
			name:        "OK",
			userId:      1,
			audioId:     1,
			fileId:      uuid.New(),
			fileContent: "file content",
			mockBehavior: func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId, audioId int, fileId uuid.UUID, fileContent string) {
				s1.EXPECT().DownloadFile(userId, audioId).Return(storage.DownloadAudio{Title: "audio", FilePath: fileId.String()}, nil)
				r := io.NopCloser(strings.NewReader(fileContent))
				s2.EXPECT().GetFile(fileId).Return(r, int64(len(fileContent)), nil)
			},
			expectedStatusCode:   200,
			expectedLenBody:      len("file content"),
			expectedResponseBody: "file content",
		},
		{
			name: "User not found",
			mockBehavior: func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId, audioId int, fileId uuid.UUID, fileContent string) {
			},
			expectedStatusCode:   500,
			expectedLenBody:      len(`{"message":"user id not found"}`),
			expectedResponseBody: `{"message":"user id not found"}`,
		},
		{
			name:    "Invalid audio id",
			userId:  1,
			audioId: 0,
			mockBehavior: func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId, audioId int, fileId uuid.UUID, fileContent string) {
			},
			expectedStatusCode:   400,
			expectedLenBody:      len(`{"message":"invalid audio id param"}`),
			expectedResponseBody: `{"message":"invalid audio id param"}`,
		},
		{
			name:        "Can't get audio data",
			userId:      1,
			audioId:     1,
			fileId:      uuid.New(),
			fileContent: "file content",
			mockBehavior: func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId, audioId int, fileId uuid.UUID, fileContent string) {
				s1.EXPECT().DownloadFile(userId, audioId).Return(storage.DownloadAudio{}, errors.New("service not work"))
			},
			expectedStatusCode:   500,
			expectedLenBody:      len(`{"message":"service not work"}`),
			expectedResponseBody: `{"message":"service not work"}`,
		},
		{
			name:        "Wrong uuid",
			userId:      1,
			audioId:     1,
			fileId:      uuid.New(),
			fileContent: "file content",
			mockBehavior: func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId, audioId int, fileId uuid.UUID, fileContent string) {
				s1.EXPECT().DownloadFile(userId, audioId).Return(storage.DownloadAudio{Title: "audio", FilePath: "wrong uuid"}, nil)
			},
			expectedStatusCode:   500,
			expectedLenBody:      len(`{"message":"invalid UUID length: 10"}`),
			expectedResponseBody: `{"message":"invalid UUID length: 10"}`,
		},
		{
			name:        "Can't get file",
			userId:      1,
			audioId:     1,
			fileId:      uuid.New(),
			fileContent: "file content",
			mockBehavior: func(s1 *mock_service.MockAudio, s2 *mock_service.MockStorage, userId, audioId int, fileId uuid.UUID, fileContent string) {
				s1.EXPECT().DownloadFile(userId, audioId).Return(storage.DownloadAudio{Title: "audio", FilePath: fileId.String()}, nil)
				s2.EXPECT().GetFile(fileId).Return(nil, int64(0), errors.New("can't get file"))
			},
			expectedStatusCode:   500,
			expectedLenBody:      len(`{"message":"can't get file"}`),
			expectedResponseBody: `{"message":"can't get file"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			audio := mock_service.NewMockAudio(c)
			strg := mock_service.NewMockStorage(c)

			testCase.mockBehavior(audio, strg, testCase.userId, testCase.audioId, testCase.fileId, testCase.fileContent)

			services := &service.Service{Audio: audio, Storage: strg}
			handler := NewHandler(services)

			r := gin.New()
			if testCase.userId != 0 {
				r.GET("/download/:id", func(c *gin.Context) {
					c.Set(userCtx, testCase.userId)
				}, handler.downloadAudio)
			} else {
				r.GET("/download/:id", handler.downloadAudio)
			}

			w := httptest.NewRecorder()

			url := fmt.Sprintf("/download/%d", testCase.audioId)
			if testCase.audioId == 0 {
				url = fmt.Sprintf("/download/%s", "wrong_id")
			}
			req := httptest.NewRequest("GET", url, nil)

			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedLenBody, w.Body.Len())
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}
