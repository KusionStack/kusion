package database

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/postgres"
)

func TestDatabase_MarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		data          *Database
		expected      string
		expectedError error
	}{
		{
			name: "Valid MarshalJSON for MySQL",
			data: &Database{
				Header: Header{
					Type: TypeMySQL,
				},
				MySQL: &mysql.MySQL{
					Type:    "local",
					Version: "8.0",
				},
				PostgreSQL: &postgres.PostgreSQL{
					Type:    "cloud",
					Version: "15.5",
				},
			},
			expected:      `{"_type": "MySQL", "type": "local", "version": "8.0"}`,
			expectedError: nil,
		},
		{
			name: "Valid MarshalJSON for PostgreSQL",
			data: &Database{
				Header: Header{
					Type: TypePostgreSQL,
				},
				MySQL: &mysql.MySQL{
					Type:    "local",
					Version: "8.0",
				},
				PostgreSQL: &postgres.PostgreSQL{
					Type:    "cloud",
					Version: "15.5",
				},
			},
			expected:      `{"_type": "PostgreSQL", "type": "cloud", "version": "15.5"}`,
			expectedError: nil,
		},
		{
			name: "Unknown Type",
			data: &Database{
				Header: Header{
					Type: Type("Unknown"),
				},
				MySQL: &mysql.MySQL{
					Type:    "local",
					Version: "8.0",
				},
				PostgreSQL: &postgres.PostgreSQL{
					Type:    "cloud",
					Version: "15.5",
				},
			},
			expected:      "",
			expectedError: errors.New("unknown database type"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, actualErr := json.Marshal(test.data)
			if test.expectedError == nil {
				assert.JSONEq(t, test.expected, string(actual))
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, test.expectedError.Error())
			}
		})
	}
}

func TestDatabase_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		data          string
		expected      Database
		expectedError error
	}{
		{
			name: "Valid UnmarshalJSON for MySQL",
			data: `{"_type": "MySQL", "type": "local", "version": "8.0"}`,
			expected: Database{
				Header: Header{
					Type: TypeMySQL,
				},
				MySQL: &mysql.MySQL{
					Type:    "local",
					Version: "8.0",
				},
			},
			expectedError: nil,
		},
		{
			name: "Valid UnmarshalJSON for PostgreSQL",
			data: `{"_type": "PostgreSQL", "type": "cloud", "version": "15.5"}`,
			expected: Database{
				Header: Header{
					Type: TypePostgreSQL,
				},
				PostgreSQL: &postgres.PostgreSQL{
					Type:    "cloud",
					Version: "15.5",
				},
			},
			expectedError: nil,
		},
		{
			name: "Unknown Type",
			data: `{"_type": "Unknown", "type": "local", "version": "15.5"}`,
			expected: Database{
				Header: Header{
					Type: "Unknown",
				},
			},
			expectedError: errors.New("unknown database type"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Database
			actualErr := json.Unmarshal([]byte(test.data), &actual)
			if test.expectedError == nil {
				assert.Equal(t, test.expected, actual)
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, test.expectedError.Error())
			}
		})
	}
}

func TestDatabase_MarshalYAML(t *testing.T) {
	tests := []struct {
		name          string
		data          Database
		expected      string
		expectedError error
	}{
		{
			name: "Valid MarshalYAML for MySQL",
			data: Database{
				Header: Header{
					Type: TypeMySQL,
				},
				MySQL: &mysql.MySQL{
					Type:    "local",
					Version: "8.0",
				},
				PostgreSQL: &postgres.PostgreSQL{
					Type:    "cloud",
					Version: "15.5",
				},
			},
			expected: `_type: MySQL
type: local
version: "8.0"`,
			expectedError: nil,
		},
		{
			name: "Valid MarshalYAML for PostgreSQL",
			data: Database{
				Header: Header{
					Type: TypePostgreSQL,
				},
				MySQL: &mysql.MySQL{
					Type:    "local",
					Version: "8.0",
				},
				PostgreSQL: &postgres.PostgreSQL{
					Type:    "cloud",
					Version: "15.5",
				},
			},
			expected: `_type: PostgreSQL
type: cloud
version: "15.5"`,
			expectedError: nil,
		},
		{
			name: "Unknown Type",
			data: Database{
				Header: Header{
					Type: Type("Unknown"),
				},
				MySQL: &mysql.MySQL{
					Type:    "local",
					Version: "8.0",
				},
				PostgreSQL: &postgres.PostgreSQL{
					Type:    "cloud",
					Version: "15.5",
				},
			},
			expected:      "",
			expectedError: errors.New("unknown database type"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, actualErr := yaml.Marshal(test.data)
			if test.expectedError == nil {
				assert.YAMLEq(t, test.expected, string(actual))
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, test.expectedError.Error())
			}
		})
	}
}

func TestDatabase_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name          string
		data          string
		expected      Database
		expectedError error
	}{
		{
			name: "Valid UnmarshalYAML for MySQL",
			data: `_type: MySQL
type: "local"
version: "8.0"`,
			expected: Database{
				Header: Header{
					Type: TypeMySQL,
				},
				MySQL: &mysql.MySQL{
					Type:    "local",
					Version: "8.0",
				},
			},
			expectedError: nil,
		},
		{
			name: "Valid UnmarshalYAML for PostgreSQL",
			data: `_type: PostgreSQL
type: "cloud"
version: "15.5"`,
			expected: Database{
				Header: Header{
					Type: TypePostgreSQL,
				},
				PostgreSQL: &postgres.PostgreSQL{
					Type:    "cloud",
					Version: "15.5",
				},
			},
			expectedError: nil,
		},
		{
			name: "Unknown Type",
			data: `_type: Unknown
type: "local"
version: "15.5"`,
			expected:      Database{},
			expectedError: errors.New("unknown database type"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Database
			actualErr := yaml.Unmarshal([]byte(test.data), &actual)
			if test.expectedError == nil {
				assert.Equal(t, test.expected, actual)
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, test.expectedError.Error())
			}
		})
	}
}
