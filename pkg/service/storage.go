package service

import (
	"github.com/google/uuid"
	"github.com/mahadeva604/audio-storage/pkg/repository"
	"io"
)

type StorageService struct {
	repo repository.Storage
}

func NewStorageService(repo repository.Storage) *StorageService {
	return &StorageService{repo: repo}
}

func (s StorageService) StoreFile(fileId uuid.UUID, file io.ReadSeeker) error {
	return s.repo.StoreFile(fileId, file)
}

func (s StorageService) GetFile(fileId uuid.UUID) (io.ReadCloser, int64, error) {
	return s.repo.GetFile(fileId)
}
