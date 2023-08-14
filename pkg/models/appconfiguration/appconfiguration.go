package appconfiguration

import "kusionstack.io/kusion/pkg/models/appconfiguration/workload"

// AppConfiguration is a developer-centric definition that describes
// how to run an Application.
// This application model builds upon a decade of experience at
// AntGroup running super large scale internal developer platform,
// combined with best-of-breed ideas and practices from the community.
//
// Example:
//
//	 import models.schema.v1 as ac
//	 import models.schema.v1.workload as wl
//	 import models.schema.v1.workload.container as c
//
//		appConfiguration = ac.AppConfiguration {
//		    workload: wl.Service {
//		        containers: {
//		            "nginx": c.Container {
//		                image: "nginx:v1"
//		            }
//		        }
//		    }
//		}
type AppConfiguration struct {
	// Workload defines how to run your application code.
	Workload *workload.Workload `json:"workload" yaml:"workload"`

	// Labels and annotations can be used to attach arbitrary metadata
	// as key-value pairs to resources.
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}
