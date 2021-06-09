package repository

import (
	"bytes"
	"errors"
	"github.com/google/uuid"
	storage "github.com/mahadeva604/audio-storage"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestStorageFS_StoreFile(t *testing.T) {
	testTable := []struct {
		name            string
		fileId          uuid.UUID
		file            io.ReadSeeker
		expectedErr     bool
		expectedErrType error
	}{
		{
			name:   "OK",
			fileId: uuid.New(),
			file:   bytes.NewReader([]byte{0xFF, 0xF1}),
		},
		{
			name:            "EOF",
			fileId:          uuid.New(),
			file:            bytes.NewReader([]byte{}),
			expectedErr:     true,
			expectedErrType: errors.New("EOF"),
		},
		{
			name:            "Error not aac file",
			fileId:          uuid.New(),
			file:            bytes.NewReader([]byte{0x12, 0x34, 0x56}),
			expectedErr:     true,
			expectedErrType: storage.NotAacFile,
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			tmpdir := t.TempDir()
			s := NewStorageFS(tmpdir)
			err := s.StoreFile(testCase.fileId, testCase.file)
			if testCase.expectedErr {
				assert.Error(t, err)
				if testCase.expectedErrType != nil {
					assert.Equal(t, testCase.expectedErrType, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
