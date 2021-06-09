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

func TestSharePostgres_ShareAudio(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	r := NewSharePostgres(db)
	type mockBehavior func(shareId, audioId, userID int)

	testTable := []struct {
		name            string
		shareID         int
		audioId         int
		userId          int
		mockBehavior    mockBehavior
		expectedErr     bool
		expectedErrType error
	}{
		{
			name:    "OK",
			shareID: 1,
			audioId: 2,
			userId:  3,
			mockBehavior: func(shareId, audioId, userID int) {
				result := sqlmock.NewResult(0, 1)
				mock.ExpectExec("INSERT INTO shares SELECT (.+) FROM audios").WithArgs(shareId, audioId, userID).WillReturnResult(result)
			},
		},
		{
			name:    "Error query",
			shareID: 1,
			audioId: 2,
			userId:  3,
			mockBehavior: func(shareId, audioId, userID int) {
				mock.ExpectExec("INSERT INTO shares SELECT (.+) FROM audios").WithArgs(shareId, audioId, userID).WillReturnError(errors.New("query error"))
			},
			expectedErr:     true,
			expectedErrType: errors.New("query error"),
		},
		{
			name:    "Error share exists",
			shareID: 1,
			audioId: 2,
			userId:  3,
			mockBehavior: func(shareId, audioId, userID int) {
				mock.ExpectExec("INSERT INTO shares SELECT (.+) FROM audios").WithArgs(shareId, audioId, userID).WillReturnError(&pq.Error{Code: "23505"})
			},
			expectedErr:     true,
			expectedErrType: storage.ShareExists,
		},
		{
			name:    "Error user share with not exists",
			shareID: 1,
			audioId: 2,
			userId:  3,
			mockBehavior: func(shareId, audioId, userID int) {
				mock.ExpectExec("INSERT INTO shares SELECT (.+) FROM audios").WithArgs(shareId, audioId, userID).WillReturnError(&pq.Error{Code: "23503"})
			},
			expectedErr:     true,
			expectedErrType: storage.ShareUserNotExists,
		},
		{
			name:    "Error not owner",
			shareID: 1,
			audioId: 2,
			userId:  3,
			mockBehavior: func(shareId, audioId, userID int) {
				result := sqlmock.NewResult(0, 0)
				mock.ExpectExec("INSERT INTO shares SELECT (.+) FROM audios").WithArgs(shareId, audioId, userID).WillReturnResult(result)
			},
			expectedErr:     true,
			expectedErrType: storage.NotOwner,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.shareID, testCase.audioId, testCase.userId)

			err := r.ShareAudio(testCase.userId, testCase.audioId, testCase.shareID)
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

func TestSharePostgres_UnshareAudio(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	r := NewSharePostgres(db)
	type mockBehavior func(shareId, audioId, userID int)

	testTable := []struct {
		name            string
		shareID         int
		audioId         int
		userId          int
		mockBehavior    mockBehavior
		expectedErr     bool
		expectedErrType error
	}{
		{
			name:    "OK",
			shareID: 1,
			audioId: 2,
			userId:  3,
			mockBehavior: func(shareId, audioId, userID int) {
				result := sqlmock.NewResult(0, 1)
				mock.ExpectExec(`DELETE FROM shares WHERE audio_id = \(SELECT audio_id FROM (.+)\)`).WithArgs(audioId, userID, shareId).WillReturnResult(result)
			},
		},
		{
			name:    "Error query",
			shareID: 1,
			audioId: 2,
			userId:  3,
			mockBehavior: func(shareId, audioId, userID int) {
				mock.ExpectExec(`DELETE FROM shares WHERE audio_id = \(SELECT audio_id FROM (.+)\)`).WithArgs(audioId, userID, shareId).WillReturnError(errors.New("query error"))
			},
			expectedErr:     true,
			expectedErrType: errors.New("query error"),
		},
		{
			name:    "Error not owner",
			shareID: 1,
			audioId: 2,
			userId:  3,
			mockBehavior: func(shareId, audioId, userID int) {
				result := sqlmock.NewResult(0, 0)
				mock.ExpectExec(`DELETE FROM shares WHERE audio_id = \(SELECT audio_id FROM (.+)\)`).WithArgs(audioId, userID, shareId).WillReturnResult(result)
			},
			expectedErr:     true,
			expectedErrType: storage.NotOwner,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.shareID, testCase.audioId, testCase.userId)

			err := r.UnshareAudio(testCase.userId, testCase.audioId, testCase.shareID)
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

func TestSharePostgres_GetSharedList(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	r := NewSharePostgres(db)
	type mockBehavior func(limit, offset int)

	testTable := []struct {
		name            string
		limit           int
		offset          int
		mockBehavior    mockBehavior
		expectedOut     storage.ShareListJson
		expectedErr     bool
		expectedErrType error
	}{
		{
			name:   "OK",
			limit:  5,
			offset: 2,
			mockBehavior: func(limit, offset int) {
				rows := sqlmock.NewRows([]string{"full_count", "user_id", "name", "count"}).
					AddRow(limit, 1, "User One", 3).
					AddRow(limit, 2, "User Two", 2).
					AddRow(limit, 3, "User Three", 1).
					AddRow(limit, 4, "User Four", 5).
					AddRow(limit, 5, "User Five", 4)
				query := `SELECT (.+) FROM shares s
						JOIN audios a USING \(audio_id\)
						JOIN users u ON a.user_id = u.user_id
						GROUP BY a.user_id, name ORDER BY name
						OFFSET \$1 LIMIT \$2`
				mock.ExpectQuery(query).WithArgs(offset, limit).WillReturnRows(rows)
			},
			expectedOut: storage.ShareListJson{
				Count: 5,
				Users: []storage.ShareListCount{
					{
						UserId:     1,
						Name:       "User One",
						ShareCount: 3,
					},
					{
						UserId:     2,
						Name:       "User Two",
						ShareCount: 2,
					},
					{
						UserId:     3,
						Name:       "User Three",
						ShareCount: 1,
					},
					{
						UserId:     4,
						Name:       "User Four",
						ShareCount: 5,
					},
					{
						UserId:     5,
						Name:       "User Five",
						ShareCount: 4,
					},
				},
			},
		},
		{
			name:   "Error query",
			limit:  5,
			offset: 2,
			mockBehavior: func(limit, offset int) {
				query := `SELECT (.+) FROM shares s
						JOIN audios a USING \(audio_id\)
						JOIN users u ON a.user_id = u.user_id
						GROUP BY a.user_id, name ORDER BY name
						OFFSET \$1 LIMIT \$2`
				mock.ExpectQuery(query).WithArgs(offset, limit).WillReturnError(errors.New("query error"))
			},
			expectedErr:     true,
			expectedErrType: errors.New("query error"),
		},
		{
			name:   "Error scan",
			limit:  5,
			offset: 2,
			mockBehavior: func(limit, offset int) {
				rows := sqlmock.NewRows([]string{"wrong_row"}).AddRow("wrong row")
				query := `SELECT (.+) FROM shares s
						JOIN audios a USING \(audio_id\)
						JOIN users u ON a.user_id = u.user_id
						GROUP BY a.user_id, name ORDER BY name
						OFFSET \$1 LIMIT \$2`
				mock.ExpectQuery(query).WithArgs(offset, limit).WillReturnRows(rows)
			},
			expectedErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.limit, testCase.offset)

			gotData, err := r.GetSharedList(storage.ShareListParam{
				Offset: &testCase.offset,
				Limit:  &testCase.limit,
			})

			if testCase.expectedErr {
				assert.Error(t, err)
				if testCase.expectedErrType != nil {
					assert.Equal(t, testCase.expectedErrType, err)
				}
			} else {
				assert.Equal(t, testCase.expectedOut, gotData)
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
