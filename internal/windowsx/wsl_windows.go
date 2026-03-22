//go:build windows

package windowsx

import (
	"errors"
	"os"
)

func IsWSL() bool { return false }

var (
	username     = os.Getenv(`USERNAME`)
	userProfile  = os.Getenv(`USERPROFILE`)
	appData      = os.Getenv(`AppData`)
	localAppData = os.Getenv(`LocalAppData`)
)

func Username() (string, error) {
	if len(username) == 0 {
		return ``, errors.New(`%USERNAME% is empty`)
	}
	return username, nil
}

func UserProfile() (string, error) {
	if len(userProfile) == 0 {
		return ``, errors.New(`%USERPROFILE% is empty`)
	}
	return userProfile, nil
}

func AppData() (string, error) {
	if len(appData) == 0 {
		return ``, errors.New(`%AppData% is empty`)
	}
	return appData, nil
}

func LocalAppData() (string, error) {
	if len(localAppData) == 0 {
		return ``, errors.New(`%LocalAppData% is empty`)
	}
	return localAppData, nil
}
