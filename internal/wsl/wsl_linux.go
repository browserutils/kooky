//go:build linux

package wsl

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// IsWSL returns true if running inside Windows Subsystem for Linux
func IsWSL() bool {
	if _, exists := os.LookupEnv("WSL_DISTRO_NAME"); exists {
		return true
	}

	_, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop")
	return err == nil
}

func WindowsUsername() (string, error) {
	if !IsWSL() {
		return "", errors.New("not running inside WSL")
	}

	// First try to extract username from PATH
	pathDirs := strings.Split(os.Getenv("PATH"), ":")
	windowsAppsPattern := regexp.MustCompile(`/mnt/c/Users/([^/]+)/AppData/Local/Microsoft/WindowsApps`)
	for _, dir := range pathDirs {
		if matches := windowsAppsPattern.FindStringSubmatch(dir); matches != nil {
			return matches[1], nil
		}
	}

	// Fall back to PowerShell command
	cmd := exec.Command("powershell.exe", "$env:UserName")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get username: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// parent directory of %AppData% and %LocalAppData%
func WSLAppDataRoot() (string, error) {
	username, err := WindowsUsername()
	if err != nil {
		return ``, err
	}
	appData := filepath.Join(`/mnt/c/Users`, username, `AppData`)
	return appData, nil
}
