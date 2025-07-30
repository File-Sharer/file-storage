package service

import "errors"

var (
	ErrInternal = errors.New("internal server error")
	ErrFileMustHaveAValidExtension = errors.New("file must have a valid extension")
)
