package config

import (
	"github.com/pkg/errors"
)

// now not all implementation return those error codes
var (
	KeyNotFoundError      = errors.New("the key is not found")
	InvalidValueTypeError = errors.New("the value is not expected type")
)
