package metadata

import (
	"fmt"
	"strings"
)

const (
	// AnnotationSource is the OpenContainers annotation for specifying
	// the upstream source of an OCI artifact.
	AnnotationSource = "org.opencontainers.image.source"

	// AnnotationRevision is the OpenContainers annotation for specifying
	// the upstream source revision of an OCI artifact.
	AnnotationRevision = "org.opencontainers.image.revision"

	// AnnotationCreated is the OpenContainers annotation for specifying
	// the date and time on which the OCI artifact was built (RFC 3339).
	AnnotationCreated = "org.opencontainers.image.created"

	// AnnotationVersion is the OpenContainers annotation for specifying
	// the semantic version of an artifact.
	AnnotationVersion = "org.opencontainers.image.version"

	// AnnotationTitle is the OpenContainers annotation for specifying
	// the human-readable title of an artifact.
	AnnotationTitle = "org.opencontainers.image.title"
)

// Metadata holds the upstream information about on artifact's source.
// https://github.com/opencontainers/image-spec/blob/main/annotations.md
type Metadata struct {
	Created     string
	Source      string
	Revision    string
	Digest      string
	URL         string
	Annotations map[string]string
}

// ToAnnotations returns the OpenContainers annotations map.
func (m *Metadata) ToAnnotations() map[string]string {
	annotations := map[string]string{
		AnnotationCreated:  m.Created,
		AnnotationSource:   m.Source,
		AnnotationRevision: m.Revision,
	}

	for k, v := range m.Annotations {
		annotations[k] = v
	}

	return annotations
}

// MetadataFromAnnotations parses the OpenContainers annotations and returns a Metadata object.
func MetadataFromAnnotations(annotations map[string]string) *Metadata {
	return &Metadata{
		Created:     annotations[AnnotationCreated],
		Source:      annotations[AnnotationSource],
		Revision:    annotations[AnnotationRevision],
		Annotations: annotations,
	}
}

// ParseAnnotations parses the annotations string in key=value format
// and returns the OpenContainers annotations.
func ParseAnnotations(annotationsStr []string) (map[string]string, error) {
	annotations := map[string]string{}
	for _, annotation := range annotationsStr {
		kv := strings.Split(annotation, "=")
		if len(kv) != 2 {
			return annotations, fmt.Errorf("invalid annotation %s, must be in the format key=value", annotation)
		}
		annotations[kv[0]] = kv[1]
	}

	return annotations, nil
}
