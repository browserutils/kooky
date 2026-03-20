//go:build windows

package windowsx

import "os"

func IsWSL() bool { return false }

var (
	username     = os.Getenv(`USERNAME`)
	userProfile  = os.Getenv(`USERPROFILE`)
	appData      = os.Getenv(`AppData`)
	localAppData = os.Getenv(`LocalAppData`)
)

func Username() (string, error)     { return username, nil }
func UserProfile() (string, error)  { return userProfile, nil }
func AppData() (string, error)      { return appData, nil }
func LocalAppData() (string, error) { return localAppData, nil }
