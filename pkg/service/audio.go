package service

import (
	storage "github.com/mahadeva604/audio-storage"
	"github.com/mahadeva604/audio-storage/pkg/repository"
)

type AudioService struct {
	repo repository.Audio
}

func NewAudioService(repo repository.Audio) *AudioService {
	return &AudioService{repo: repo}
}

func (s *AudioService) UploadFile(userId int, path string) (int, error) {
	return s.repo.UploadFile(userId, path)
}

func (s *AudioService) DownloadFile(userID, audioId int) (storage.DownloadAudio, error) {
	return s.repo.DownloadFile(userID, audioId)
}

func (s *AudioService) AddDescription(userID, audioId int, input storage.UpdateAudio) error {
	return s.repo.AddDescription(userID, audioId, input)
}

func (s *AudioService) GetAudioList(userID int, input storage.AudioListParam) (storage.AudioListJson, error) {
	return s.repo.GetAudioList(userID, input)
}
