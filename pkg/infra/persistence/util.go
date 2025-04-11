package persistence

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"kusionstack.io/kusion/pkg/domain/entity"
)

// MultiString is a custom type for handling arrays of strings with GORM.
type MultiString []string

// Scan implements the Scanner interface for the MultiString type.
func (s *MultiString) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		*s = strings.Split(string(src), ",")
	case string:
		*s = strings.Split(src, ",")
	case nil:
		*s = nil
	default:
		return fmt.Errorf("unsupported type %T", src)
	}
	return nil
}

// Value implements the Valuer interface for the MultiString type.
func (s MultiString) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return strings.Join(s, ","), nil
}

// GormDataType gorm common data type
func (s MultiString) GormDataType() string {
	return "text"
}

// GormDBDataType gorm db data type
func (s MultiString) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	// returns different database type based on driver name
	switch db.Dialector.Name() {
	case "mysql", "sqlite":
		return "text"
	}
	return ""
}

// Create a mock database connection
func GetMockDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	// Create a sqlMock of sql.DB.
	fakeDB, sqlMock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	// common execution for orm
	sqlMock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows(
		[]string{"VERSION()"}).AddRow("5.7.35-log"))

	// Create the gorm database connection with fake db
	fakeGDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      fakeDB,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, nil, err
	}

	return fakeGDB, sqlMock, nil
}

// Close the gorm database connection
func CloseDB(t *testing.T, gdb *gorm.DB) {
	db, err := gdb.DB()
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

func GetProjectQuery(filter *entity.ProjectFilter) (string, []interface{}) {
	pattern := make([]string, 0)
	args := make([]interface{}, 0)
	if filter.OrgID != 0 {
		pattern = append(pattern, "organization_id = ?")
		args = append(args, fmt.Sprint(filter.OrgID))
	}
	if filter.Name != "" {
		pattern = append(pattern, "name = ?")
		args = append(args, filter.Name)
	}
	if filter.FuzzyName != "" {
		pattern = append(pattern, "name LIKE ?")
		args = append(args, fmt.Sprintf("%%%s%%", filter.FuzzyName))
	}
	return CombineQueryParts(pattern), args
}

func GetStackQuery(filter *entity.StackFilter) (string, []interface{}) {
	pattern := make([]string, 0)
	args := make([]interface{}, 0)
	if filter.OrgID != 0 {
		pattern = append(pattern, "project.organization_id = ?")
		args = append(args, fmt.Sprint(filter.OrgID))
	}
	if filter.ProjectID != 0 {
		pattern = append(pattern, "project_id = ?")
		args = append(args, fmt.Sprint(filter.ProjectID))
	}
	if filter.Path != "" {
		pattern = append(pattern, "stack.path = ?")
		args = append(args, filter.Path)
	}
	return CombineQueryParts(pattern), args
}

func GetWorkspaceQuery(filter *entity.WorkspaceFilter) (string, []interface{}) {
	pattern := make([]string, 0)
	args := make([]interface{}, 0)
	if filter.BackendID != 0 {
		pattern = append(pattern, "backend_id = ?")
		args = append(args, fmt.Sprint(filter.BackendID))
	}
	if filter.Name != "" {
		pattern = append(pattern, "name = ?")
		args = append(args, filter.Name)
	}
	return CombineQueryParts(pattern), args
}

func GetResourceQuery(filter *entity.ResourceFilter) (string, []interface{}) {
	pattern := make([]string, 0)
	args := make([]interface{}, 0)
	if filter.OrgID != 0 {
		pattern = append(pattern, "organization_id = ?")
		args = append(args, fmt.Sprint(filter.OrgID))
	}
	if filter.ProjectID != 0 {
		pattern = append(pattern, "project_id = ?")
		args = append(args, fmt.Sprint(filter.ProjectID))
	}
	if filter.StackID != 0 {
		pattern = append(pattern, "stack_id = ?")
		args = append(args, fmt.Sprint(filter.StackID))
	}
	if filter.ResourcePlane != "" {
		pattern = append(pattern, "resource_plane = ?")
		args = append(args, filter.ResourcePlane)
	}
	if filter.ResourceType != "" {
		pattern = append(pattern, "resource_type = ?")
		args = append(args, filter.ResourceType)
	}
	return CombineQueryParts(pattern), args
}

func GetRunQuery(filter *entity.RunFilter) (string, []interface{}) {
	pattern := make([]string, 0)
	args := make([]interface{}, 0)
	if filter.ProjectID != 0 {
		pattern = append(pattern, "project.ID = ?")
		args = append(args, fmt.Sprint(filter.ProjectID))
	}
	if filter.StackID != 0 {
		pattern = append(pattern, "stack_id = ?")
		args = append(args, filter.StackID)
	}
	if filter.Workspace != "" {
		pattern = append(pattern, "workspace.name = ?")
		args = append(args, filter.Workspace)
	}
	if len(filter.Type) > 0 {
		pattern = append(pattern, "run.type IN (?)")
		args = append(args, filter.Type)
	}
	if len(filter.Status) > 0 {
		pattern = append(pattern, "run.status IN (?)")
		args = append(args, filter.Status)
	}
	if !filter.StartTime.IsZero() && !filter.EndTime.IsZero() {
		start := time.Date(filter.StartTime.Year(),
			filter.StartTime.Month(),
			filter.StartTime.Day(),
			filter.StartTime.Hour(),
			filter.StartTime.Minute(),
			filter.StartTime.Second(),
			0, time.Local)
		end := time.Date(filter.EndTime.Year(),
			filter.EndTime.Month(),
			filter.EndTime.Day(),
			filter.EndTime.Hour(),
			filter.EndTime.Minute(),
			filter.EndTime.Second(),
			0, time.Local)
		pattern = append(pattern, "run.created_at >= ? AND run.created_at <= ?")
		args = append(args, start, end)
	}

	return CombineQueryParts(pattern), args
}

func GetSourceQuery(filter *entity.SourceFilter) (string, []interface{}) {
	pattern := make([]string, 0)
	args := make([]interface{}, 0)
	if filter.SourceName != "" {
		pattern = append(pattern, "source.name LIKE ?")
		args = append(args, fmt.Sprintf("%%%s%%", filter.SourceName))
	}
	return CombineQueryParts(pattern), args
}

func GetModuleQuery(filter *entity.ModuleFilter) (string, []interface{}) {
	pattern := make([]string, 0)
	args := make([]interface{}, 0)
	if filter.ModuleName != "" {
		pattern = append(pattern, "module.name LIKE ?")
		args = append(args, fmt.Sprintf("%%%s%%", filter.ModuleName))
	}
	return CombineQueryParts(pattern), args
}

func GetVariableSetQuery(filter *entity.VariableSetFilter) (string, []interface{}) {
	pattern := make([]string, 0)
	args := make([]interface{}, 0)
	if filter.Name != "" {
		pattern = append(pattern, "variable_set.name = ?")
		args = append(args, filter.Name)
	}
	return CombineQueryParts(pattern), args
}

func GetVariableQuery(filter *entity.VariableFilter) (string, []interface{}) {
	pattern := make([]string, 0)
	args := make([]interface{}, 0)
	if filter.Name != "" {
		pattern = append(pattern, "variable.name = ?")
		args = append(args, filter.Name)
	}
	if filter.VariableSet != "" {
		pattern = append(pattern, "variable.variable_set = ?")
		args = append(args, filter.VariableSet)
	}
	return CombineQueryParts(pattern), args
}

func CombineQueryParts(queryParts []string) string {
	queryString := ""
	if len(queryParts) > 0 {
		queryString = queryParts[0]
		for _, part := range queryParts[1:] {
			queryString += fmt.Sprintf(" AND %s", part)
		}
	}
	return queryString
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&BackendModel{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&SourceModel{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&OrganizationModel{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&ProjectModel{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&StackModel{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&ResourceModel{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&WorkspaceModel{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&ModuleModel{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&RunModel{}); err != nil {
		return err
	}
	return nil
}
