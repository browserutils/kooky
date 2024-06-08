//go:build !windows

package utils

import "os"

var openFile = os.Open
