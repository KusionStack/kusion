package kfile

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
)

const (
	EnvKusionPath = "KUSION_PATH"
)

func Stat(filename string) (fileInfo os.FileInfo, err error) {
	// Golang's official os.Stat() function is case-insensitive in some systems, which treats /var/folder and /var/FoldER as same path.
	// That is apparently insufficient. In that case, we define kfile.Stat() to make up that deficiency.
	// See: https://github.com/golang/go/issues/25786
	fileInfo, err = os.Stat(filename)
	if runtime.GOOS == "linux" {
		return
	}
	if err != nil {
		return
	}
	if filename[len(filename)-1] == '/' {
		filename = filename[:len(filename)-1]
	}

	dirPath := filepath.Dir(filename)
	base := filepath.Base(filename)

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if f.Name() == base {
			return
		}
	}

	return nil, os.ErrNotExist
}

func FileExists(filename string) (bool, error) {
	info, err := Stat(filename)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

// Get the kusion data directory of the current user
func KusionDataFolder() (string, error) {
	var kusionDataFolder string

	if kusionPath := os.Getenv(EnvKusionPath); kusionPath != "" {
		kusionDataFolder = kusionPath
	} else {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		kusionDataFolder = path.Join(usr.HomeDir, ".kusion")
		if exist, _ := FileExists(kusionDataFolder); !exist {
			err = os.MkdirAll(kusionDataFolder, os.ModePerm)
			if err != nil {
				return "", err
			}
		}
	}

	return kusionDataFolder, nil
}

// Get the file name of the kusion credentials file
func KusionCredentialsFilename() string {
	return "credentials.json"
}

// GetCredentialsToken returns the token from credentials file
func GetCredentialsToken() string {
	// Get token from credentials.json in kusion data folder
	credentials, err := GetCredentials()
	if err != nil {
		return ""
	}
	return credentials["token"].(string)
}

// Get the kusion credentials data
func GetCredentials() (map[string]interface{}, error) {
	// Get kusion data folder
	kusionDataFolder, err := KusionDataFolder()
	if err != nil {
		return nil, err
	}
	// Get kusion credentials data from credentials.json in kusion data folder
	credentialsFilepath := filepath.Join(kusionDataFolder, KusionCredentialsFilename())
	data, err := ioutil.ReadFile(credentialsFilepath)
	if err != nil {
		return nil, err
	}
	var credentials map[string]interface{}
	err = json.Unmarshal(data, &credentials)
	if err != nil {
		return nil, err
	}
	return credentials, nil
}
