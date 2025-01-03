package project

import (
	"fmt"
	"net/url"
	"strings"
)

// GenerateDefaultSourceName generates a default source name based on the remote URL
func GenerateDefaultSourceName(remoteURL string) (string, error) {
	parsedURL, err := url.Parse(remoteURL)
	if err != nil {
		return "", ErrDefaultSourceRemoteInvalid
	}

	// Clean up the path - remove leading/trailing slashes and special characters
	path := strings.Trim(parsedURL.Path, "/")
	path = strings.ReplaceAll(path, "/", "-")

	// Generate source name
	sourceName := fmt.Sprintf("default-%s", parsedURL.Host)
	if path != "" {
		sourceName = fmt.Sprintf("default-%s-%s", parsedURL.Host, path)
	}

	return sourceName, nil
}
