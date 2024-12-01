package fs

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

func findFile(dir string, filename string) string {
	path := filepath.Join(dir, filename)
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}

func findExeDir(filename string) (dir string, filePath string) {
	if exePath, err := os.Executable(); err == nil {
		dir = filepath.Dir(exePath)
		filePath = findFile(dir, filename)
	}
	return
}

func findAppDataDir(filename string) (dir string, filePath string) {
	if appDataDir := os.Getenv("APPDATA"); appDataDir != "" {
		dir = filepath.Join(appDataDir, "mfa")
		filePath = findFile(dir, filename)
	} else {
		slog.Debug("Environment variable APPDATA is not defined and cannot be used as configuration location")
	}
	return
}

func findXDGConfig(filename string) (dir string, filePath string) {
	if xdgConfigDir := os.Getenv("XDG_CONFIG_HOME"); xdgConfigDir != "" {
		dir = filepath.Join(xdgConfigDir, "mfa")
		filePath = findFile(filePath, filename)
	}
	return
}

func findHomeConfigDir(filename string) (dir string, filePath string) {
	home, err := Dir()
	if err != nil {
		slog.Debug("Home directory lookup failed and cannot be used as configuration location")
		return
	} else if home == "" {
		slog.Debug("Home directory not defined and cannot be used as configuration location")
		return
	}
	dir = filepath.Join(home, ".config", "mfa")
	filePath = findFile(dir, filename)
	return
}

func MakeFilenamePath(filename string) (filePath string) {
	var (
		dir           string
		defaultDir    string
		homeConfigDir string
	)
	if _, filePath = findExeDir(filename); filePath != "" {
		return
	}
	if runtime.GOOS == "windows" {
		if defaultDir, filePath = findAppDataDir(filename); filePath != "" {
			return
		}
	}
	if dir, filePath = findXDGConfig(filename); filePath != "" {
		return
	}
	if runtime.GOOS != "windows" {
		defaultDir = dir
	}
	if homeConfigDir, filePath = findHomeConfigDir(filename); filePath != "" {
		return
	}
	slog.Debug(fmt.Sprintf("no existing %s found, create a new one", filename))
	if defaultDir != "" {
		filePath = filepath.Join(defaultDir, filename)
		if err := os.MkdirAll(defaultDir, os.ModePerm); err == nil {
			return
		}
	} else if homeConfigDir != "" {
		filePath = filepath.Join(homeConfigDir, filename)
		if err := os.MkdirAll(homeConfigDir, os.ModePerm); err == nil {
			return
		}
	}
	return filename
}
