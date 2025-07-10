//go:build !linux

package wsl

import (
	"errors"
)

// wrong platform

// IsWSL always returns false
func IsWSL() bool                      { return false }
func WindowsUsername() (string, error) { return "", errors.New("wrong platform") }
func WSLAppDataRoot() (string, error)  { return "", errors.New("wrong platform") }
