package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/File-Sharer/file-storage/internal/model"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	DEFAULT_FILE_PATH_PREFIX = "files/"
	FILE_URL_STRING = "%s/%s"
)

type uploaderService struct {
	logger *zap.Logger
}

func newUploaderService(logger *zap.Logger) Uploader {
	return &uploaderService{
		logger: logger,
	}
}

func (s *uploaderService) saveFile(path string, file multipart.File, fileHeader *multipart.FileHeader) (int64, string, error) {
	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		return 0, "", ErrFileMustHaveAValidExtension
	}

	fileID := uuid.New()
	var filePath string
	path = strings.TrimSpace(path)
	if path != "" {
		dirPath := filepath.Join(DEFAULT_FILE_PATH_PREFIX, path)
		filePath = filepath.Join(dirPath, fileID.String() + ext)

		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			s.logger.Sugar().Errorf("failed to create directories: %s", err.Error())
			return 0, "", err
		}
	} else {
		filePath = filepath.Join(DEFAULT_FILE_PATH_PREFIX, fileID.String() + ext)
	}

	createdFile, err := os.Create(filePath)
	if err != nil {
		s.logger.Sugar().Errorf("failed to create file: %s", err.Error())
		return 0, "", err
	}
	defer createdFile.Close()

	if _, err := io.Copy(createdFile, file); err != nil {
		s.logger.Sugar().Errorf("failed to copy src: %s", err.Error())
		return 0, "", err
	}

	filePath = strings.ReplaceAll(filePath, "\\", "/")

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		s.logger.Sugar().Errorf("failed to get file(%s) info: %s", filePath, err.Error())
		return 0, "", err
	}

	fileURL := fmt.Sprintf(FILE_URL_STRING, viper.GetString("app.origin"), filePath)

	return fileInfo.Size(), fileURL, nil
}

func (s *uploaderService) Upload(d model.UploadData) (int64, string, error) {
	buff := make([]byte, 512)
	if _, err := d.File.Read(buff); err != nil {
		s.logger.Sugar().Errorf("error while uploading a file: %s", err.Error())
		return 0, "", err
	}

	if _, err := d.File.Seek(0, io.SeekStart); err != nil {
		s.logger.Sugar().Errorf("error while uploading a file: %s", err.Error())
		return 0, "", err
	}

	return s.saveFile(d.Path, d.File, d.FileHeader)
}

func (s *uploaderService) Delete(paths []string) error {
	for _, path := range paths {
		path = filepath.Join(DEFAULT_FILE_PATH_PREFIX, path)

		if err := os.Remove(path); err != nil {
			s.logger.Sugar().Errorf("failed to remove path(%s): %s", path, err.Error())
		}
	}

	return nil
}
