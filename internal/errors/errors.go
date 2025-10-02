package errors

import "errors"

var (
	ErrConfigNotFound   = errors.New("configuration file not found")
	ErrStateFileMissing = errors.New("state file missing")
	ErrTaskNotFound     = errors.New("task not found")
)
