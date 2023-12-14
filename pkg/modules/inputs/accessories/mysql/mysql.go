package mysql

import "fmt"

const (
	CloudDBType = "cloud"
	LocalDBType = "local"
)

const (
	ErrEmptyInstanceTypeForCloudDB = "empty instance type for cloud managed mysql instance"
)

// MySQL describes the attributes to locally deploy or create a cloud provider
// managed mysql database instance for the workload.
type MySQL struct {
	// The deployment mode of the mysql database.
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// The mysql database version to use.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// The type of the mysql instance.
	InstanceType string `json:"instanceType,omitempty" yaml:"instanceType,omitempty"`
	// The allocated storage size of the mysql instance.
	Size int `json:"size,omitempty" yaml:"size,omitempty"`
	// The edition of the mysql instance provided by the cloud vendor.
	Category string `json:"category,omitempty" yaml:"category,omitempty"`
	// The operation account for the mysql database.
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	// The list of IP addresses allowed to access the mysql instance provided by the cloud vendor.
	SecurityIPs []string `json:"securityIPs,omitempty" yaml:"securityIPs,omitempty"`
	// The virtual subnet ID associated with the VPC that the cloud mysql instance will be created in.
	SubnetID string `json:"subnetID,omitempty" yaml:"subnetID,omitempty"`
	// Whether the host address of the cloud mysql instance for the workload to connect with is via
	// public network or private network of the cloud vendor.
	PrivateRouting bool `json:"privateRouting,omitempty" yaml:"privateRouting,omitempty"`
}

// Validate validates whether the input of a mysql database instance is valid.
func (db *MySQL) Validate() error {
	if db.Type == CloudDBType && db.InstanceType == "" {
		return fmt.Errorf(ErrEmptyInstanceTypeForCloudDB)
	}

	return nil
}
