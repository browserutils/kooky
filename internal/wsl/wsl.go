package wsl

import "errors"

var ErrNotWSL = errors.New("not running inside WSL")
