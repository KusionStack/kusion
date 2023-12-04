package auth

import (
	"context"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2cfg "github.com/aws/aws-sdk-go-v2/config"
)

// NewV2Config returns an aws.Config for AWS SDK v2, using the default options.
func NewV2Config(ctx context.Context, region, profile string) (awsv2.Config, error) {
	var optFns []func(*awsv2cfg.LoadOptions) error
	if region != "" {
		optFns = append(optFns, awsv2cfg.WithRegion(region))
	}
	if profile != "" {
		optFns = append(optFns, awsv2cfg.WithSharedConfigProfile(profile))
	}

	return awsv2cfg.LoadDefaultConfig(ctx, optFns...)
}
