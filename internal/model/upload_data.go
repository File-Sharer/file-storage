package model

import "mime/multipart"

type UploadData struct {
	Path       string                `json:"path"`
	File       multipart.File        `json:"file"`
	FileHeader *multipart.FileHeader `json:"file_header"`
}
