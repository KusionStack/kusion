package project

import (
	"fmt"
	"net/url"
	"strings"

	"kusionstack.io/kusion/pkg/domain/constant"
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

func validateProjectSortOptions(sortBy string) (string, error) {
	if sortBy == "" {
		return constant.SortByID, nil
	}
	if sortBy != constant.SortByID && sortBy != constant.SortByName && sortBy != constant.SortByCreateTimestamp {
		return "", fmt.Errorf("invalid sort option: %s. Can only sort by id, name or create timestamp", sortBy)
	}
	switch sortBy {
	case constant.SortByCreateTimestamp:
		return "created_at", nil
	case constant.SortByModifiedTimestamp:
		return "updated_at", nil
	}
	return sortBy, nil
}
