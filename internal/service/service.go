package service

import (
	"github.com/File-Sharer/file-storage/internal/model"
	"go.uber.org/zap"
)

type Uploader interface {
	Upload(d model.UploadData) (int64, string, error)
	Delete(paths []string) error
	CreateFolder(path string) error
}

type Service struct {
	Uploader
}

func New(logger *zap.Logger) *Service {
	return &Service{
		Uploader: newUploaderService(logger),
	}
}
