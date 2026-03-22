//go:build !windows && !linux

package windowsx

func IsWSL() bool                   { return false }
func Username() (string, error)     { return "", ErrNotWSL }
func UserProfile() (string, error)  { return "", ErrNotWSL }
func AppData() (string, error)      { return "", ErrNotWSL }
func LocalAppData() (string, error) { return "", ErrNotWSL }
