package server

import (
	"strconv"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kusionstack.io/kusion/pkg/cmd/server/util"

	gomysql "github.com/go-sql-driver/mysql"
	"github.com/spf13/pflag"
)

var (
	ErrDBHostNotSpecified = errors.New("--db-host must be specified")
	ErrDBNameNotSpecified = errors.New("--db-name must be specified")
	ErrDBUserNotSpecified = errors.New("--db-user must be specified")
	ErrDBPortNotSpecified = errors.New("--db-port must be specified")
)

// DatabaseAccessOptions holds the database access layer configurations.
type DatabaseAccessOptions struct {
	DBName     string `json:"dbName,omitempty" yaml:"dbName,omitempty"`
	DBUser     string `json:"dbUser,omitempty" yaml:"dbUser,omitempty"`
	DBPassword string `json:"dbPassword,omitempty" yaml:"dbPassword,omitempty"`
	DBHost     string `json:"dbHost,omitempty" yaml:"dbHost,omitempty"`
	DBPort     int    `json:"dbPort,omitempty" yaml:"dbPort,omitempty"`
}

// NewDatabaseAccessOptions returns a DatabaseAccessOptions struct with the default values
func NewDatabaseAccessOptions() *DatabaseAccessOptions {
	return &DatabaseAccessOptions{
		DBHost: "127.0.0.1",
		DBPort: 3306,
	}
}

// InstallDB uses the run options to generate and open a db session.
func (o *DatabaseAccessOptions) InstallDB() (*gorm.DB, error) {
	// Generate go-sql-driver.mysql config to format DSN
	config := gomysql.NewConfig()
	config.User = o.DBUser
	config.Passwd = o.DBPassword
	config.Addr = o.DBHost + ":" + strconv.Itoa(o.DBPort)
	config.DBName = o.DBName
	config.Net = "tcp"
	config.ParseTime = true
	config.InterpolateParams = true
	config.Params = map[string]string{
		"charset": "utf8",
		"loc":     "Asia/Shanghai",
	}
	dsn := config.FormatDSN()
	// silence log output
	cfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	return gorm.Open(mysql.Open(dsn), cfg) // todo: add db connection check to healthz check
}

// ApplyTo uses the run options to generate and open a db session.
func (o *DatabaseAccessOptions) ApplyTo(db **gorm.DB) error {
	d, err := o.InstallDB()
	if err != nil {
		return err
	}
	*db = d
	return nil
}

// Validate checks validation of DatabaseAccessOptions
func (o *DatabaseAccessOptions) Validate() error {
	var errs []error
	if len(o.DBHost) == 0 {
		errs = append(errs, ErrDBHostNotSpecified)
	}
	if len(o.DBName) == 0 {
		errs = append(errs, ErrDBNameNotSpecified)
	}
	if len(o.DBUser) == 0 {
		errs = append(errs, ErrDBUserNotSpecified)
	}
	if o.DBPort == 0 {
		errs = append(errs, ErrDBPortNotSpecified)
	}
	if errs != nil {
		err := util.AggregateError(errs)
		return errors.Wrap(err, "invalid db options")
	}
	return nil
}

// AddFlags adds flags related to DB to a specified FlagSet
func (o *DatabaseAccessOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.DBName, "db-name", o.DBName, "the database name")
	fs.StringVar(&o.DBUser, "db-user", o.DBUser, "the user name used to access database")
	fs.StringVar(&o.DBPassword, "db-pass", o.DBPassword, "the user password used to access database")
	fs.StringVar(&o.DBHost, "db-host", o.DBHost, "database host")
	fs.IntVar(&o.DBPort, "db-port", o.DBPort, "database port")
}
