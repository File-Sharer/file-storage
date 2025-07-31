package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/File-Sharer/file-storage/internal/model"
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

	var filePath string
	path = strings.TrimSpace(path)
	if path != "" {
		dirPath := filepath.Join(DEFAULT_FILE_PATH_PREFIX, path)
		filePath = filepath.Join(dirPath, fileHeader.Filename)

		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			s.logger.Sugar().Errorf("failed to create directories: %s", err.Error())
			return 0, "", err
		}
	} else {
		filePath = filepath.Join(DEFAULT_FILE_PATH_PREFIX, fileHeader.Filename)
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
		path = filepath.Clean(path)

		if err := os.Remove(path); err != nil {
			s.logger.Sugar().Errorf("failed to remove path(%s): %s", path, err.Error())
		}
	}

	return nil
}

func (s *uploaderService) CreateFolder(path string) error {
	path = filepath.Join(DEFAULT_FILE_PATH_PREFIX, path)
	path = filepath.Clean(path)
	return os.MkdirAll(path, os.ModePerm)
}

func (s *uploaderService) GetZippedFolder(folderPath string) ([]byte, error) {
	basePath := filepath.Join(DEFAULT_FILE_PATH_PREFIX, folderPath)
	basePath = filepath.Clean(basePath)

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	err := filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(filepath.Dir(basePath), path)
		if err != nil {
			return err
		}
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		if info.IsDir() {
			if relPath != "" {
				_, err := w.Create(relPath + "/")
				if err != nil {
					return err
				}
			}
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := w.Create(relPath)
		if err != nil {
			return err
		}

		if _, err := io.Copy(writer, file); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.logger.Sugar().Errorf("failed to walk path(%s): %s", basePath, err.Error())
		w.Close()
		return nil, ErrInternal
	}

	if err := w.Close(); err != nil {
		s.logger.Sugar().Errorf("failed to close zip writer: %s", err.Error())
		return nil, ErrInternal
	}

	return buf.Bytes(), nil
}
