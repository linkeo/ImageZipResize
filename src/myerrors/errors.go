package myerrors

import (
	"errors"
)

var ErrNotDirectory = errors.New("path is not directory")
