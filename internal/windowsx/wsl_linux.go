//go:build linux

package windowsx

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// IsWSL returns true if running inside Windows Subsystem for Linux
var IsWSL = sync.OnceValue(func() bool {
	if _, exists := os.LookupEnv("WSL_DISTRO_NAME"); exists {
		return true
	}
	_, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop")
	return err == nil
})

type winEnv struct {
	username     string
	userProfile  string
	appData      string
	localAppData string
	err          error
}

var getWinEnv = sync.OnceValue(func() winEnv {
	if !IsWSL() {
		return winEnv{err: ErrNotWSL}
	}
	cmd := exec.Command(
		`/mnt/c/Windows/System32/cmd.exe`,
		`/c`,
		``+ // disagreement with the formatter...
			`echo ^%USERNAME^%: %USERNAME% & `+
			`echo ^%USERPROFILE^%: %USERPROFILE% & `+
			`echo ^%AppData^%: %AppData% & `+
			`echo ^%LocalAppData^%: %LocalAppData%`,
	)
	out, err := cmd.Output()
	if err == nil {
		if env := parseWinEnv(string(out)); env.username != "" {
			return env
		}
	}
	// fallback: guess from PATH
	return guessWinEnv()
})

func parseWinEnv(output string) winEnv {
	var env winEnv
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimRight(line, "\r ")
		key, val, ok := strings.Cut(line, ": ")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		switch key {
		case "USERNAME":
			env.username = val
		case "USERPROFILE":
			env.userProfile = winToWSLPath(val)
		case "AppData":
			env.appData = winToWSLPath(val)
		case "LocalAppData":
			env.localAppData = winToWSLPath(val)
		}
	}
	return env
}

// guessWinEnv extracts the Windows username from PATH and derives standard paths
func guessWinEnv() winEnv {
	pathDirs := strings.Split(os.Getenv("PATH"), ":")
	// WSL2 default: interop.appendWindowsPath: true, can be disabled
	windowsAppsPattern := regexp.MustCompile(`/mnt/c/Users/([^/]+)/AppData/Local/Microsoft/WindowsApps`)
	for _, dir := range pathDirs {
		if matches := windowsAppsPattern.FindStringSubmatch(dir); matches != nil {
			username := matches[1]
			profile := filepath.Join(`/mnt/c/Users`, username)
			return winEnv{
				username:     username,
				userProfile:  profile,
				appData:      filepath.Join(profile, `AppData`, `Roaming`),
				localAppData: filepath.Join(profile, `AppData`, `Local`),
			}
		}
	}
	return winEnv{err: ErrNotWSL}
}

// winToWSLPath converts a Windows path like C:\Users\john to /mnt/c/Users/john
func winToWSLPath(winPath string) string {
	if len(winPath) < 3 || winPath[1] != ':' {
		return winPath
	}
	drive := strings.ToLower(string(winPath[0]))
	rest := strings.ReplaceAll(winPath[2:], `\`, `/`)
	return "/mnt/" + drive + rest
}

func Username() (string, error)     { e := getWinEnv(); return e.username, e.err }
func UserProfile() (string, error)  { e := getWinEnv(); return e.userProfile, e.err }
func AppData() (string, error)      { e := getWinEnv(); return e.appData, e.err }
func LocalAppData() (string, error) { e := getWinEnv(); return e.localAppData, e.err }
