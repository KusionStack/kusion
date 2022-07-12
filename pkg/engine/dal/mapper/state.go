package mapper

import (
	"database/sql"
	"time"

	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	"github.com/pkg/errors"
)

type StateDO struct {
	ID            int64     `json:"id"`
	Tenant        string    `json:"tenant"`
	Project       string    `json:"project"`
	Stack         string    `json:"stack"`
	Cluster       string    `json:"cluster,omitempty"`
	Version       int       `json:"version"`
	KusionVersion string    `json:"kusion_version"`
	Serial        uint64    `json:"serial"`
	Operator      string    `json:"operator"`
	Resources     string    `json:"resources"`
	CreateTime    time.Time `json:"create_time"`
	ModifiedTime  time.Time `json:"modified_time"`
}

// GetOne gets one record from table build_task by condition "where"
func GetOne(db *sql.DB, where map[string]interface{}) (*StateDO, error) {
	if nil == db {
		return nil, errors.New("sql.DB is nil")
	}
	cond, values, err := builder.BuildSelect("state", where, nil)
	if nil != err {
		return nil, err
	}
	row, err := db.Query(cond, values...)
	if nil != err || nil == row {
		return nil, err
	}
	defer row.Close()
	var dbRes *StateDO
	scanner.SetTagName("json")
	err = scanner.Scan(row, &dbRes)
	return dbRes, err
}

// Insert inserts an array of data into table StateDO
func Insert(db *sql.DB, data []map[string]interface{}) (int64, error) {
	if nil == db {
		return 0, errors.New("sql.DB is nil")
	}

	cond, values, err := builder.BuildInsert("state", data)
	if nil != err {
		return 0, err
	}

	result, err := db.Exec(cond, values...)
	if nil != err || nil == result {
		return 0, err
	}

	return result.LastInsertId()
}
