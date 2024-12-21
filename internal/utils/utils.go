package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func ListFilesInDir(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range dirEntries {
		if !entry.IsDir() &&
			entry.Name()[0] != '.' &&
			!strings.HasSuffix(entry.Name(), ".lock") {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

func CreateBackup(filename string) {
	filenameBackup := filename + "_" + time.Now().Format("20060102150405")
	fmt.Println(filenameBackup)
}

func Stat(filename string) error {
	_, err := os.Stat(filename)
	if err != nil {
		return err
	}
	return nil
}

func CleanName(path string) string {
	ext := filepath.Ext(path)
	basename := filepath.Base(path)
	return basename[:len(basename)-len(ext)]
}

func GetDir(file string) string {
	return filepath.Dir(file)
}

func AbsolutePath(file string) (string, error) {
	if strings.HasPrefix(file, "~") {
		u, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("Failed to get current user: %v", err)
		}
		file = filepath.Join(u.HomeDir, file[1:])
	}

	absolutePath, err := filepath.Abs(file)
	if err != nil {
		return "", fmt.Errorf("Failed to resolve absolute path: %v", err)
	}

	return absolutePath, nil
}

func IsSingleEntry[T any](m map[string]T) bool {
	if len(m) == 1 {
		return true
	}
	return false
}

func GetSingleKey[T any](m map[string]T) string {
	for key := range m {
		return key
	}
	return ""
}
