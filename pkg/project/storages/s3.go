package storages

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Storage is an implementation of graph.Storage which uses s3 as storage.
type S3Storage struct {
	s3     *s3.S3
	bucket string

	// The prefix to store the project folders' directory.
	prefix string
}

// NewS3Storage creates a new S3Storage instance.
func NewS3Storage(s3 *s3.S3, bucket, prefix string) *S3Storage {
	s := &S3Storage{
		s3:     s3,
		bucket: bucket,
		prefix: prefix,
	}
	return s
}

// Get returns a project map which key is workspace name and value is its belonged project list.
func (s *S3Storage) Get() (map[string][]string, error) {
	projects := map[string][]string{}
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(s.prefix + "/"),
		Delimiter: aws.String("/"),
	}
	for {
		// List all the project objects
		output, err := s.s3.ListObjectsV2(input)
		if err != nil {
			return nil, fmt.Errorf("list projects directory from s3 failed: %w", err)
		}

		for _, projectPrefix := range output.CommonPrefixes {
			// Get project name
			projectDir := strings.TrimPrefix(*projectPrefix.Prefix, s.prefix+"/")
			projectDir = strings.TrimSuffix(projectDir, "/")

			// List workspaces under the project prefix
			listWorkspacesInput := &s3.ListObjectsV2Input{
				Bucket:    aws.String(s.bucket),
				Prefix:    projectPrefix.Prefix,
				Delimiter: aws.String("/"),
			}

			workspaceOutput, err := s.s3.ListObjectsV2(listWorkspacesInput)
			if err != nil {
				return nil, fmt.Errorf("list project's workspaces directory from s3 failed: %w", err)
			}

			for _, workspacePrefix := range workspaceOutput.CommonPrefixes {
				// Get each of the workspace name
				workspaceDir := strings.TrimPrefix(*workspacePrefix.Prefix, *projectPrefix.Prefix)
				workspaceDir = strings.TrimSuffix(workspaceDir, "/")
				// Store workspace name as key, project name as value
				projects[workspaceDir] = append(projects[workspaceDir], projectDir)
			}
		}

		// Break if there are no more results
		if *output.IsTruncated {
			input.ContinuationToken = output.NextContinuationToken
		} else {
			break
		}
	}
	return projects, nil
}
