package repository

import (
	"github.com/google/uuid"
	storage "github.com/mahadeva604/audio-storage"
	"io"
	"os"
)

type StorageFS struct {
	dirPath string
}

func NewStorageFS(dirPath string) *StorageFS {
	return &StorageFS{dirPath: dirPath}
}

func (r StorageFS) StoreFile(fileId uuid.UUID, file io.ReadSeeker) error {
	// Read two bytes to check magic number

	buffer := make([]byte, 2)
	_, err := io.ReadFull(file, buffer)
	if err != nil {
		return err
	}

	if ok := storage.Aac(buffer); !ok {
		return storage.NotAacFile
	}

	file.Seek(0, io.SeekStart)

	newFileName := fileId.String() + storage.FileExt

	out, err := os.Create(r.dirPath + newFileName)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)

	if err != nil {
		return err
	}

	return nil
}

func (r StorageFS) GetFile(fileId uuid.UUID) (io.ReadCloser, int64, error) {
	file, err := os.Open(r.dirPath + fileId.String() + storage.FileExt)
	if err != nil {
		return nil, 0, err
	}
	fileStat, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}

	return file, fileStat.Size(), nil
}
