package db

import (
	"errors"
	"net/url"

	"github.com/didi/gendry/manager"
	"github.com/zclconf/go-cty/cty"
	"kusionstack.io/kusion/pkg/engine/states"
)

type DBBackend struct {
	DBState
}

func NewDBBackend() states.Backend {
	return &DBBackend{}
}

// ConfigSchema returns a description of the expected configuration
// structure for the receiving backend.
func (b *DBBackend) ConfigSchema() cty.Type {
	config := map[string]cty.Type{
		"dbName":     cty.String,
		"dbUser":     cty.String,
		"dbPassword": cty.String,
		"dbHost":     cty.String,
		"dbPort":     cty.Number,
	}
	return cty.Object(config)
}

// Configure uses the provided configuration to set configuration fields
// within the DBState backend.
func (b *DBBackend) Configure(obj cty.Value) error {
	var dbName, dbUser, dbPassword, dbHost, dbPort cty.Value
	if dbName = obj.GetAttr("dbName"); dbName.IsNull() {
		return errors.New("dbName must be configure in backend config")
	}
	if dbUser = obj.GetAttr("dbUser"); dbUser.IsNull() {
		return errors.New("dbUser must be configure in backend config")
	}
	if dbPassword = obj.GetAttr("dbPassword"); dbPassword.IsNull() {
		return errors.New("dbPassword must be configure in backend config")
	}
	if dbHost = obj.GetAttr("dbHost"); dbHost.IsNull() {
		return errors.New("dbHost must be configure in backend config")
	}
	if dbPort = obj.GetAttr("dbPort"); dbPort.IsNull() {
		return errors.New("dbPort must be configure in backend config")
	}
	port, _ := dbPort.AsBigFloat().Int64()

	db, err := manager.New(dbName.AsString(), dbUser.AsString(), dbPassword.AsString(), dbHost.AsString()).Set(
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
func (b *DBBackend) StateStorage() states.StateStorage {
	return &DBState{b.DB}
}
