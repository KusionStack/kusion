package metadata

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
)

// Metadata holds the upstream information about on artifact's source.
// https://github.com/opencontainers/image-spec/blob/main/annotations.md
type Metadata struct {
	Created     string            `json:"created,omitempty"`
	Source      string            `json:"source_url,omitempty"`
	Revision    string            `json:"source_revision,omitempty"`
	Digest      string            `json:"digest"`
	URL         string            `json:"url"`
	Annotations map[string]string `json:"annotations,omitempty"`
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
