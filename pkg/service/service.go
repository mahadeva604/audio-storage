package service

import (
	"github.com/google/uuid"
	storage "github.com/mahadeva604/audio-storage"
	"github.com/mahadeva604/audio-storage/pkg/repository"
	"io"
	"time"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Authorization interface {
	CreateUser(user storage.User) (int, error)
	GenerateAccessToken(username, password string) (int, string, error)
	UpdateAccessToken(userId int) (string, error)
	GenerateRefreshToken(userId int) (string, error)
	UpdateRefreshToken(oldRefreshToken string) (int, string, error)
	ParseToken(token string) (int, error)
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

type Service struct {
	Authorization
	Audio
	Share
	Storage
}

func NewService(repos *repository.Repository, secretKey []byte, accessTokenTTL, refreshTokenTTL time.Duration) *Service {
	return &Service{
		Authorization: NewAuthService(repos, secretKey, accessTokenTTL, refreshTokenTTL),
		Audio:         NewAudioService(repos),
		Share:         NewShareService(repos),
		Storage:       NewStorageService(repos),
	}
}
