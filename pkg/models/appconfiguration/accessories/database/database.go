package database

const moduleName = "database"

// As an important supporting accessory, Database describes the attributes to
// locally deploy or create a cloud provider managed database instance for the workload.
type Database struct {
	// The local deployment mode or the specific cloud vendor that provides the
	// relational database service (rds).
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// The database engine to use.
	Engine string `json:"engine,omitempty" yaml:"engine,omitempty"`
	// The database engine version to use.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// The type of the rds instance.
	InstanceType string `json:"instanceType,omitempty" yaml:"instanceType,omitempty"`
	// The allocated storage size of the rds instance provided by the cloud vendor in GB.
	Size int `json:"size,omitempty" yaml:"size,omitempty"`
	// The edition of the rds instance provided by the cloud vendor.
	Category string `json:"category,omitempty" yaml:"category,omitempty"`
	// The operation account for the database.
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	// The list of IP addresses allowed to access the rds instance provided by the cloud vendor.
	SecurityIPs []string `json:"securityIPs,omitempty" yaml:"securityIPs,omitempty"`
	// The virtual subnet ID associated with the VPC that the rds instance will be created in.
	SubnetID string `json:"subnetID,omitempty" yaml:"subnetID,omitempty"`
	// Whether the host address of the rds instance for the workload to connect with is via
	// public network or priviate network of the cloud vendor.
	PrivateRouting bool `json:"privateRouting,omitempty" yaml:"privateRouting,omitempty"`
	// The diversified rds configuration items from different cloud vendors.
	ExtraMap map[string]string `json:"extraMap,omitempty" yaml:"extraMap,omitempty"`
}

func (d *Database) ModuleName() string {
	return moduleName
}
