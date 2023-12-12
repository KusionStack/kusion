package postgres

import "fmt"

const (
	CloudDBType = "cloud"
	LocalDBType = "local"
)

const (
	ErrEmptyInstanceTypeForCloudDB = "empty instance type for cloud managed postgresql instance"
)

// PostgreSQL describes the attributes to locally deploy or create a cloud provider
// managed postgresql database instance for the workload.
type PostgreSQL struct {
	// The deployment mode of the postgresql database.
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// The postgresql database version to use.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// The type of the postgresql instance.
	InstanceType string `json:"instanceType,omitempty" yaml:"instanceType,omitempty"`
	// The allocated storage size of the postgresql instance.
	Size int `json:"size,omitempty" yaml:"size,omitempty"`
	// The edition of the postgresql instance provided by the cloud vendor.
	Category string `json:"category,omitempty" yaml:"category,omitempty"`
	// The operation account for the postgresql database.
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	// The list of IP addresses allowed to access the postgresql instance provided by the cloud vendor.
	SecurityIPs []string `json:"securityIPs,omitempty" yaml:"securityIPs,omitempty"`
	// The virtual subnet ID associated with the VPC that the cloud postgresql instance will be created in.
	SubnetID string `json:"subnetID,omitempty" yaml:"subnetID,omitempty"`
	// Whether the host address of the cloud postgresql instance for the workload to connect with is via
	// public network or private network of the cloud vendor.
	PrivateRouting bool `json:"privateRouting,omitempty" yaml:"privateRouting,omitempty"`
	// The specified name of the postgresql database instance, composed with `dbKey` and `suffix`.
	DatabaseName string `json:"databaseName,omitempty" yaml:"databaseName,omitempty"`
}

// Validate validates whether the input of a postgresql database instance is valid.
func (db *PostgreSQL) Validate() error {
	if db.Type == CloudDBType && db.InstanceType == "" {
		return fmt.Errorf(ErrEmptyInstanceTypeForCloudDB)
	}

	return nil
}
