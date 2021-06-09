package repository

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	storage "github.com/mahadeva604/audio-storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAudioPostgres_UploadFile(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	r := NewAudioPostgres(db)
	type mockBehavior func(userId int, path string, audioId int)

	testTable := []struct {
		name            string
		userId          int
		path            string
		mockBehavior    mockBehavior
		expectedAudioId int
		expectErr       bool
	}{
		{
			name:   "OK",
			userId: 1,
			path:   "file_path",
			mockBehavior: func(userId int, path string, audioId int) {
				rows := sqlmock.NewRows([]string{"audio_id"}).AddRow(audioId)
				mock.ExpectQuery("INSERT INTO audios").WithArgs(userId, path).WillReturnRows(rows)
			},
			expectedAudioId: 2,
		},
		{
			name:      "Error",
			userId:    1,
			expectErr: true,
			mockBehavior: func(userId int, path string, audioId int) {
				mock.ExpectQuery("INSERT INTO audios").WithArgs(userId, path).WillReturnError(errors.New("path is empty"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.userId, testCase.path, testCase.expectedAudioId)

			gotAudioId, err := r.UploadFile(testCase.userId, testCase.path)
			if testCase.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedAudioId, gotAudioId)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAudioPostgres_DownloadFile(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	r := NewAudioPostgres(db)
	type mockBehavior func(userId int, audioId int, title, filePath string)

	testTable := []struct {
		name              string
		userId            int
		audioId           int
		title             string
		filePath          string
		mockBehavior      mockBehavior
		expectedAudioData storage.DownloadAudio
		expectErr         bool
		expectErrType     error
	}{
		{
			name:     "OK",
			userId:   1,
			audioId:  1,
			title:    "title 1",
			filePath: "file path 1",
			mockBehavior: func(userId int, audioId int, title string, filePath string) {
				rows := sqlmock.NewRows([]string{"title", "file_path"}).AddRow(title, filePath)
				mock.ExpectQuery(`SELECT (.+) FROM audios a LEFT JOIN shares r USING \(audio_id\) WHERE (.+)`).WithArgs(audioId, userId).WillReturnRows(rows)
			},
			expectedAudioData: storage.DownloadAudio{
				Title:    "title 1",
				FilePath: "file path 1",
			},
		},
		{
			name:          "No rows error",
			userId:        1,
			audioId:       2,
			expectErr:     true,
			expectErrType: storage.FileNotFound,
			mockBehavior: func(userId int, audioId int, title string, filePath string) {
				mock.ExpectQuery(`SELECT (.+) FROM audios a LEFT JOIN shares r USING \(audio_id\) WHERE (.+)`).WithArgs(audioId, userId).WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:      "Other error",
			userId:    1,
			audioId:   2,
			expectErr: true,
			mockBehavior: func(userId int, audioId int, title string, filePath string) {
				mock.ExpectQuery(`SELECT (.+) FROM audios a LEFT JOIN shares r USING \(audio_id\) WHERE (.+)`).WithArgs(audioId, userId).WillReturnError(errors.New("other error"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.userId, testCase.audioId, testCase.title, testCase.filePath)

			gotAudioData, err := r.DownloadFile(testCase.userId, testCase.audioId)
			if testCase.expectErr {
				if testCase.expectErrType != nil {
					assert.Equal(t, testCase.expectErrType, err)
				}
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedAudioData, gotAudioData)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAudioPostgres_AddDescription(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	r := NewAudioPostgres(db)
	type mockBehavior func(userId int, audioId int, input storage.UpdateAudio)

	tittle := "title 1"
	duration := 77

	testTable := []struct {
		name            string
		userId          int
		audioId         int
		input           storage.UpdateAudio
		mockBehavior    mockBehavior
		expectedErr     bool
		expectedErrType error
	}{
		{
			name:    "OK",
			userId:  1,
			audioId: 1,
			input: storage.UpdateAudio{
				Title:    &tittle,
				Duration: &duration,
			},
			mockBehavior: func(userId int, audioId int, input storage.UpdateAudio) {
				mock.ExpectExec("UPDATE audios SET (.+) WHERE (.+)").WithArgs(input.Title, input.Duration, userId, audioId).WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:    "Error not updated",
			userId:  1,
			audioId: 1,
			input: storage.UpdateAudio{
				Title:    &tittle,
				Duration: &duration,
			},
			mockBehavior: func(userId int, audioId int, input storage.UpdateAudio) {
				mock.ExpectExec("UPDATE audios SET (.+) WHERE (.+)").WithArgs(input.Title, input.Duration, userId, audioId).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedErr:     true,
			expectedErrType: storage.NotOwner,
		},
		{
			name:    "Error",
			userId:  1,
			audioId: 1,
			input: storage.UpdateAudio{
				Title:    &tittle,
				Duration: &duration,
			},
			mockBehavior: func(userId int, audioId int, input storage.UpdateAudio) {
				mock.ExpectExec("UPDATE audios SET (.+) WHERE (.+)").WithArgs(input.Title, input.Duration, userId, audioId).WillReturnError(errors.New("some error"))
			},
			expectedErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.userId, testCase.audioId, testCase.input)

			err = r.AddDescription(testCase.userId, testCase.audioId, testCase.input)
			if testCase.expectedErr {
				assert.Error(t, err)
				if testCase.expectedErrType != nil {
					assert.Equal(t, testCase.expectedErrType, err)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAudioPostgres_GetAudioList(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	r := NewAudioPostgres(db)
	type mockBehavior func(userId int, input storage.AudioListParam)

	offset, limit := 0, 1

	testTable := []struct {
		name          string
		userId        int
		input         storage.AudioListParam
		mockBehavior  mockBehavior
		expectErr     bool
		expectErrType error
		expectData    storage.AudioListJson
	}{
		{
			name:   "OK owner order",
			userId: 1,
			input: storage.AudioListParam{
				Offset:    &offset,
				Limit:     &limit,
				OrderType: "owner",
			},
			mockBehavior: func(userId int, input storage.AudioListParam) {
				query := `SELECT (.+) FROM \(SELECT (.+) FROM audios (.+) ORDER BY is_owner DESC, name, title OFFSET \$2 LIMIT \$3\) (.+)  ORDER BY is_owner DESC, name, title`
				rows := sqlmock.NewRows([]string{"full_count", "audio_id", "title", "is_owner", "user_id", "name", "shared_to_id", "shared_to_name"}).
					AddRow(10, 1, "audio 1", true, 1, "user 1", 2, "user 2").
					AddRow(10, 1, "audio 1", true, 1, "user 1", 3, "user 3").
					AddRow(10, 2, "audio 2", true, 1, "user 1", 0, "").
					AddRow(10, 3, "audio 3", false, 2, "user 2", 1, "user 1")
				mock.ExpectQuery(query).WithArgs(userId, input.Offset, input.Limit).WillReturnRows(rows)
			},
			expectData: storage.AudioListJson{
				TotalCount: 10,
				Records: []storage.AudioList{
					{
						Id:      1,
						Title:   "audio 1",
						IsOwner: true,
						Owner:   1,
						Name:    "user 1",
						Shares: &[]storage.ShareList{
							{
								UserId: 2,
								Name:   "user 2",
							},
							{
								UserId: 3,
								Name:   "user 3",
							},
						},
					},
					{
						Id:      2,
						Title:   "audio 2",
						IsOwner: true,
						Owner:   1,
						Name:    "user 1",
						Shares:  nil,
					},
					{
						Id:      3,
						Title:   "audio 3",
						IsOwner: false,
						Owner:   2,
						Name:    "user 2",
						Shares: &[]storage.ShareList{
							{
								UserId: 1,
								Name:   "user 1",
							},
						},
					},
				},
			},
		},
		{
			name:   "OK title order",
			userId: 1,
			input: storage.AudioListParam{
				Offset:    &offset,
				Limit:     &limit,
				OrderType: "alphabet",
			},
			mockBehavior: func(userId int, input storage.AudioListParam) {
				query := `SELECT (.+) FROM \(SELECT (.+) FROM audios (.+) ORDER BY title OFFSET \$2 LIMIT \$3\) (.+)  ORDER BY title`
				rows := sqlmock.NewRows([]string{"full_count", "audio_id", "title", "is_owner", "user_id", "name", "shared_to_id", "shared_to_name"}).
					AddRow(10, 1, "audio 1", true, 1, "user 1", 2, "user 2").
					AddRow(10, 1, "audio 1", true, 1, "user 1", 3, "user 3").
					AddRow(10, 2, "audio 2", true, 1, "user 1", 0, "").
					AddRow(10, 3, "audio 3", false, 2, "user 2", 1, "user 1")
				mock.ExpectQuery(query).WithArgs(userId, input.Offset, input.Limit).WillReturnRows(rows)
			},
			expectData: storage.AudioListJson{
				TotalCount: 10,
				Records: []storage.AudioList{
					{
						Id:      1,
						Title:   "audio 1",
						IsOwner: true,
						Owner:   1,
						Name:    "user 1",
						Shares: &[]storage.ShareList{
							{
								UserId: 2,
								Name:   "user 2",
							},
							{
								UserId: 3,
								Name:   "user 3",
							},
						},
					},
					{
						Id:      2,
						Title:   "audio 2",
						IsOwner: true,
						Owner:   1,
						Name:    "user 1",
						Shares:  nil,
					},
					{
						Id:      3,
						Title:   "audio 3",
						IsOwner: false,
						Owner:   2,
						Name:    "user 2",
						Shares: &[]storage.ShareList{
							{
								UserId: 1,
								Name:   "user 1",
							},
						},
					},
				},
			},
		},
		{
			name:   "Error query",
			userId: 1,
			input: storage.AudioListParam{
				Offset:    &offset,
				Limit:     &limit,
				OrderType: "owner",
			},
			mockBehavior: func(userId int, input storage.AudioListParam) {
				query := `SELECT (.+) FROM \(SELECT (.+) FROM audios (.+) ORDER BY is_owner DESC, name, title OFFSET \$2 LIMIT \$3\) (.+)  ORDER BY is_owner DESC, name, title`
				mock.ExpectQuery(query).WithArgs(userId, input.Offset, input.Limit).WillReturnError(errors.New("query error"))
			},
			expectErr:     true,
			expectErrType: errors.New("query error"),
		},
		{
			name:   "Error unknown order",
			userId: 1,
			input: storage.AudioListParam{
				Offset:    &offset,
				Limit:     &limit,
				OrderType: "unknown",
			},
			mockBehavior: func(userId int, input storage.AudioListParam) {
			},
			expectErr:     true,
			expectErrType: errors.New("unknown order type"),
		},
		{
			name:   "Error scan",
			userId: 1,
			input: storage.AudioListParam{
				Offset:    &offset,
				Limit:     &limit,
				OrderType: "owner",
			},
			mockBehavior: func(userId int, input storage.AudioListParam) {
				query := `SELECT (.+) FROM \(SELECT (.+) FROM audios (.+) ORDER BY is_owner DESC, name, title OFFSET \$2 LIMIT \$3\) (.+)  ORDER BY is_owner DESC, name, title`
				rows := sqlmock.NewRows([]string{"wrong_row"}).
					AddRow("wrong_row")
				mock.ExpectQuery(query).WithArgs(userId, input.Offset, input.Limit).WillReturnRows(rows)
			},
			expectErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.userId, testCase.input)

			output, err := r.GetAudioList(testCase.userId, testCase.input)
			if testCase.expectErr {
				assert.Error(t, err)
				if testCase.expectErrType != nil {
					assert.Equal(t, testCase.expectErrType, err)
				}
			} else {
				assert.Equal(t, testCase.expectData, output)
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
