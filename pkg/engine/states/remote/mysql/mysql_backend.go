package mysql

import (
	"errors"
	"net/url"

	"github.com/didi/gendry/manager"
	"github.com/zclconf/go-cty/cty"

	"kusionstack.io/kusion/pkg/engine/states"
)

type MysqlBackend struct {
	MysqlState
}

func NewMysqlBackend() states.Backend {
	return &MysqlBackend{}
}

// ConfigSchema returns a description of the expected configuration
// structure for the receiving backend.
func (b *MysqlBackend) ConfigSchema() cty.Type {
	config := map[string]cty.Type{
		"dbName":   cty.String,
		"user":     cty.String,
		"password": cty.String,
		"host":     cty.String,
		"port":     cty.Number,
	}
	return cty.Object(config)
}

// Configure uses the provided configuration to set configuration fields
// within the MysqlState backend.
func (b *MysqlBackend) Configure(obj cty.Value) error {
	var dbName, dbUser, dbPassword, dbHost, dbPort cty.Value
	if dbName = obj.GetAttr("dbName"); dbName.IsNull() {
		return errors.New("dbName must be configure in backend config")
	}
	if dbUser = obj.GetAttr("user"); dbUser.IsNull() {
		return errors.New("user must be configure in backend config")
	}
	if dbHost = obj.GetAttr("host"); dbHost.IsNull() {
		return errors.New("host must be configure in backend config")
	}
	if dbPort = obj.GetAttr("port"); dbPort.IsNull() {
		return errors.New("port must be configure in backend config")
	}

	port, _ := dbPort.AsBigFloat().Int64()
	var password string
	if dbPassword = obj.GetAttr("password"); !dbPassword.IsNull() {
		password = dbPassword.AsString()
	}
	db, err := manager.New(dbName.AsString(), dbUser.AsString(), password, dbHost.AsString()).Set(
		manager.SetCharset("utf8"),
		manager.SetParseTime(true),
		manager.SetInterpolateParams(true),
		manager.SetLoc(url.QueryEscape("Asia/Shanghai"))).Port(int(port)).Open(true)
	if err != nil {
		return err
	}
	b.DB = db

	return nil
}

// StateStorage return a StateStorage to manage State stored in db
func (b *MysqlBackend) StateStorage() states.StateStorage {
	return &MysqlState{b.DB}
}
