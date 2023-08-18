package appconfiguration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"
)

var (
	appString = `workload:
    _type: Job
    containers:
        busybox:
            image: busybox:1.28
            command:
                - /bin/sh
                - -c
                - echo hello
    replicas: 2
    schedule: 0 * * * *
`
	appStruct = AppConfiguration{
		Workload: &workload.Workload{
			Header: workload.Header{
				Type: workload.TypeJob,
			},
			Job: &workload.Job{
				Base: workload.Base{
					Containers: map[string]container.Container{
						"busybox": {
							Image:   "busybox:1.28",
							Command: []string{"/bin/sh", "-c", "echo hello"},
						},
					},
					Replicas: 2,
				},
				Schedule: "0 * * * *",
			},
		},
	}
)

func TestAppConfigurationMarshal(t *testing.T) {
	in := appStruct
	exp := appString
	out, err := yaml.Marshal(in)
	require.NoError(t, err)
	require.Equal(t, exp, string(out))
}

func TestAppConfigurationUnmarshal(t *testing.T) {
	in := appString
	exp := appStruct
	out := AppConfiguration{}
	err := yaml.Unmarshal([]byte(in), &out)
	require.NoError(t, err)
	require.Equal(t, exp, out)
}
