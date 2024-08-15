package entity

import (
	"fmt"
	"net/url"
	"time"

	"kusionstack.io/kusion/pkg/domain/constant"
)

// Source represents the specific configuration code source,
// which should be a specific instance of the source provider.
type Source struct {
	// ID is the id of the source.
	ID uint `yaml:"id" json:"id"`
	// SourceProvider is the type of the source provider.
	SourceProvider constant.SourceProviderType `yaml:"sourceProvider" json:"sourceProvider"`
	// Remote is the source URL, including scheme.
	Remote *url.URL `yaml:"remote" json:"remote"`
	// Description is a human-readable description of the source.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Labels are custom labels associated with the source.
	Labels []string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Owners is a list of owners for the source.
	Owners []string `yaml:"owners,omitempty" json:"owners,omitempty"`
	// CreationTimestamp is the timestamp of the created for the source.
	CreationTimestamp time.Time `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	// UpdateTimestamp is the timestamp of the updated for the source.
	UpdateTimestamp time.Time `yaml:"updateTimestamp,omitempty" json:"updateTimestamp,omitempty"`
}

// Validate checks if the source is valid.
// It returns an error if the source is not valid.
func (s *Source) Validate() error {
	if s == nil {
		return fmt.Errorf("source is nil")
	}

	if s.SourceProvider == "" {
		return fmt.Errorf("source must have a source provider")
	}

	if s.Remote == nil {
		return fmt.Errorf("source must have a remote")
	}

	return nil
}

func (s *Source) Summary() string {
	return fmt.Sprintf("[%s][%s]", string(s.SourceProvider), s.Remote.String())
}
