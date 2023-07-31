package appconfiguration

import "kusionstack.io/kusion/pkg/models/appconfiguration/component"

// AppConfiguration is a developer-centric definition that describes how to run an Application.
// This application model builds upon a decade of experience at AntGroup running super large scale
// internal developer platform, combined with best-of-breed ideas and practices from the community.
//
// Example
//
//	components:
//	  proxy:
//	    containers:
//	      nginx:
//	        image: nginx:v1
//	        command:
//	        - /bin/sh
//	        - -c
//	        - echo hi
//	        args:
//	        - /bin/sh
//	        - -c
//	        - echo hi
//	        env:
//	          env1: VALUE
//	          env2: secret://sec-name/key
//	        workingDir: /tmp
//	    replicas: 2
type AppConfiguration struct {
	// Component defines the delivery artifact of one application.
	// Each application can be composed by multiple components.
	Components map[string]component.Component `json:"components,omitempty" yaml:"components,omitempty"`
}
