package server

import (
	"encoding/json"
	"os"
	"strings"

	"kusionstack.io/kusion/pkg/server"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var _ Options = &DatabaseOptions{}

// DatabaseOptions is a Database options struct
type DatabaseOptions struct {
	DatabaseAccessOptions `json:",inline" yaml:",inline"`
	// AutoMigrate will attempt to automatically migrate all tables
	AutoMigrate bool   `json:"autoMigrate,omitempty" yaml:"autoMigrate,omitempty"`
	MigrateFile string `json:"migrateFile,omitempty" yaml:"migrateFile,omitempty"`
}

// NewDatabaseOptions returns a DatabaseOptions instance with the default values
func NewDatabaseOptions() *DatabaseOptions {
	return &DatabaseOptions{
		DatabaseAccessOptions: DatabaseAccessOptions{},
		AutoMigrate:           false,
	}
}

// Validate checks DatabaseOptions and return a slice of found error(s)
func (o *DatabaseOptions) Validate() error {
	if o == nil {
		return errors.Errorf("options is nil")
	}

	if o.AutoMigrate && len(o.MigrateFile) == 0 {
		return errors.Errorf("when --auto-migrate is true, --migrate-file must be specified")
	}

	return o.DatabaseAccessOptions.Validate()
}

// ApplyTo apply database options to the server config
func (o *DatabaseOptions) ApplyTo(config *server.Config) {
	if err := o.DatabaseAccessOptions.ApplyTo(&config.DB); err != nil {
		logrus.Fatalf("Failed to apply database options to server.Config as: %+v", err)
	}

	// AutoMigrate will attempt to automatically migrate all tables
	if o.AutoMigrate {
		logrus.Debugf("AutoMigrate will attempt to automatically migrate all tables from [%s]", o.MigrateFile)
		// Read all content by migrate file
		migrateSQL, err := os.ReadFile(o.MigrateFile)
		if err != nil {
			logrus.Fatalf("Failed to read migrate file: %+v", err)
		}

		// Split multiple SQL statements into individual statements
		stmts := strings.Split(string(migrateSQL), ";")

		// Iterate over all statements and execute them
		for _, stmt := range stmts {
			// Ignore empty statements
			if len(strings.TrimSpace(stmt)) == 0 {
				continue
			}

			// Use gorm.Exec() function to execute SQL statement
			if err = config.DB.Exec(stmt).Error; err != nil {
				logrus.Warnf("Failed to exec migrate sql: %+v", err)
			}
		}
	}
}

// AddFlags adds flags for a specific Option to the specified FlagSet
func (o *DatabaseOptions) AddFlags(fs *pflag.FlagSet) {
	o.DatabaseAccessOptions.AddFlags(fs)

	fs.BoolVar(&o.AutoMigrate, "auto-migrate", o.AutoMigrate, "Whether to enable automatic migration")
	fs.StringVar(&o.MigrateFile, "migrate-file", o.MigrateFile, "The migrate sql file")
}

// MarshalJSON is custom marshalling function for masking sensitive field values
func (o DatabaseOptions) MarshalJSON() ([]byte, error) {
	type tempOptions DatabaseOptions
	o2 := tempOptions(o)
	o2.DBPassword = MaskString
	return json.Marshal(&o2)
}
