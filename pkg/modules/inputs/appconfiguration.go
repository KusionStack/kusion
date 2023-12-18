package inputs

import (
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/inputs/monitoring"
	"kusionstack.io/kusion/pkg/modules/inputs/trait"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

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
	// OpsRule specifies collection of rules that will be checked for Day-2 operation.
	OpsRule    *trait.OpsRule      `json:"opsRule,omitempty" yaml:"opsRule,omitempty"`
	Monitoring *monitoring.Monitor `json:"monitoring,omitempty" yaml:"monitoring,omitempty"`

	// MySQL defines a locally deployed or a cloud provider managed
	// mysql database instance for the workload.
	MySQL *mysql.MySQL `json:"mysql,omitempty" yaml:"mysql,omitempty"`

	// Labels and annotations can be used to attach arbitrary metadata
	// as key-value pairs to resources.
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}
