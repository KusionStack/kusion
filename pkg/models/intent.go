package models

// Intent represents the operational intentions that you aim to deliver using Kusion.
// These intentions are expected to contain all components throughout the software development lifecycle (SDLC),
// including resources (workload, database, load balancer, etc.), dependencies, and policies. The Kusion module generators are responsible
// for converting all AppConfigurations and environment configurations into the Intent.
// Once the Intent is generated, the Kusion Engine takes charge of updating the actual infrastructures to match the Intent.
type Intent struct {
	Resources Resources `json:"resources" yaml:"resources"`
}
