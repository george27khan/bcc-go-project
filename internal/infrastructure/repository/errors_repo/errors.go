package errors_repo

import "errors"

var (
	ErrTaskNotExist = errors.New("task not exist")
	ErrFileNotExist = errors.New("file not exist")
)
