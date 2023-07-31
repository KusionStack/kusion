package job

import "kusionstack.io/kusion/pkg/models/appconfiguration/component/container"

type Job struct {
	container.Container
	// The schedule in Cron format
	schedule string
}
