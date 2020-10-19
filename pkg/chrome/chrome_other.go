// +build !darwin,!windows,!linux android

package chrome

import (
	"fmt"
	"runtime"
)

func setChromeKeychainPassword(password []byte) []byte {
	return password
}

func decryptValue(encrypted []byte) (string, error) {
	return "", fmt.Errorf("decryptValue not implemented on %q", runtime.GOOS)
}
