package app

import "errors"

// Common errors for the app layer
var (
	ErrKeyExistsInEnv  = errors.New("key already exists in env file")
	ErrEnvFileNotExist = errors.New("env file does not exist")
)
