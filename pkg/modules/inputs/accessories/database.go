package database

import (
	"encoding/json"
	"errors"

	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/postgres"
)

type Type string

const (
	TypeMySQL      = "MySQL"
	TypePostgreSQL = "PostgreSQL"
)

type Header struct {
	Type Type `yaml:"_type" json:"_type"`
}

type Database struct {
	Header               `yaml:",inline" json:",inline"`
	*mysql.MySQL         `yaml:",inline" json:",inline"`
	*postgres.PostgreSQL `yaml:",inline" json:",inline"`
}

func (db Database) MarshalJSON() ([]byte, error) {
	switch db.Header.Type {
	case TypeMySQL:
		return json.Marshal(struct {
			Header       `yaml:",inline" json:",inline"`
			*mysql.MySQL `yaml:",inline" json:",inline"`
		}{
			Header: Header{db.Header.Type},
			MySQL:  db.MySQL,
		})
	case TypePostgreSQL:
		return json.Marshal(struct {
			Header               `yaml:",inline" json:",inline"`
			*postgres.PostgreSQL `yaml:",inline" json:",inline"`
		}{
			Header:     Header{db.Header.Type},
			PostgreSQL: db.PostgreSQL,
		})
	default:
		return nil, errors.New("unknown database type")
	}
}

func (db *Database) UnmarshalJSON(data []byte) error {
	var databaseData Header
	err := json.Unmarshal(data, &databaseData)
	if err != nil {
		return err
	}

	db.Header.Type = databaseData.Type
	switch db.Header.Type {
	case TypeMySQL:
		var v mysql.MySQL
		err = json.Unmarshal(data, &v)
		db.MySQL = &v
	case TypePostgreSQL:
		var v postgres.PostgreSQL
		err = json.Unmarshal(data, &v)
		db.PostgreSQL = &v
	default:
		err = errors.New("unknown database type")
	}

	return err
}

func (db Database) MarshalYAML() (interface{}, error) {
	switch db.Header.Type {
	case TypeMySQL:
		return struct {
			Header       `yaml:",inline" json:",inline"`
			*mysql.MySQL `yaml:",inline" json:",inline"`
		}{
			Header: Header{db.Header.Type},
			MySQL:  db.MySQL,
		}, nil
	case TypePostgreSQL:
		return struct {
			Header               `yaml:",inline" json:",inline"`
			*postgres.PostgreSQL `yaml:",inline" json:",inline"`
		}{
			Header:     Header{db.Header.Type},
			PostgreSQL: db.PostgreSQL,
		}, nil
	default:
		return nil, errors.New("unknown database type")
	}
}

func (db *Database) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var databaseData Header
	err := unmarshal(&databaseData)
	if err != nil {
		return err
	}

	db.Header.Type = databaseData.Type
	switch db.Header.Type {
	case TypeMySQL:
		var v mysql.MySQL
		err = unmarshal(&v)
		db.MySQL = &v
	case TypePostgreSQL:
		var v postgres.PostgreSQL
		err = unmarshal(&v)
		db.PostgreSQL = &v
	default:
		err = errors.New("unknown database type")
	}

	return err
}
