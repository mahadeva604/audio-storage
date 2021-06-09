package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	storage "github.com/mahadeva604/audio-storage"
	"io"
)

type Authorization interface {
	CreateUser(user storage.User) (int, error)
	GetUser(username, password string) (storage.User, error)
}

type Audio interface {
	UploadFile(userId int, path string) (int, error)
	AddDescription(userID, audioId int, input storage.UpdateAudio) error
	DownloadFile(userID, audioId int) (storage.DownloadAudio, error)
	GetAudioList(userID int, input storage.AudioListParam) (storage.AudioListJson, error)
}

type Share interface {
	ShareAudio(userID, audioId, shareId int) error
	UnshareAudio(userID, audioId, shareId int) error
	GetSharedList(input storage.ShareListParam) (storage.ShareListJson, error)
}

type Storage interface {
	StoreFile(fileId uuid.UUID, file io.ReadSeeker) error
	GetFile(fileId uuid.UUID) (io.ReadCloser, int64, error)
}

type Repository struct {
	Authorization
	Audio
	Share
	Storage
}

func NewRepository(db *sqlx.DB, dirPath string) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		Audio:         NewAudioPostgres(db),
		Share:         NewSharePostgres(db),
		Storage:       NewStorageFS(dirPath),
	}
}
