package states

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"kusionstack.io/kusion/pkg/engine/models"
	"net/url"
	"sort"

	"github.com/didi/gendry/scanner"
	"github.com/jinzhu/copier"
	"gopkg.in/yaml.v3"
	"kusionstack.io/kusion/pkg/engine/dal/mapper"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util"
	jsonutil "kusionstack.io/kusion/pkg/util/json"

	"github.com/didi/gendry/manager"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zclconf/go-cty/cty"
)

func init() {
	AddToBackends("db", NewDBState)
}

func NewDBState() StateStorage {
	result := &DBState{}
	return result
}

type DBState struct {
	DB *sql.DB
}

func (s *DBState) ConfigSchema() cty.Type {
	config := map[string]cty.Type{
		"dbName":     cty.String,
		"dbUser":     cty.String,
		"dbPassword": cty.String,
		"dbHost":     cty.String,
		"dbPort":     cty.Number,
	}
	return cty.Object(config)
}

func (s *DBState) Configure(obj cty.Value) error {
	var dbName, dbUser, dbPassword, dbHost, dbPort cty.Value
	if dbName = obj.GetAttr("dbName"); dbName.IsNull() {
		return fmt.Errorf("dbName must be configure in backend config")
	}
	if dbUser = obj.GetAttr("dbUser"); dbUser.IsNull() {
		return fmt.Errorf("dbUser must be configure in backend config")
	}
	if dbPassword = obj.GetAttr("dbPassword"); dbPassword.IsNull() {
		return fmt.Errorf("dbPassword must be configure in backend config")
	}
	if dbHost = obj.GetAttr("dbHost"); dbHost.IsNull() {
		return fmt.Errorf("dbHost must be configure in backend config")
	}
	if dbPort = obj.GetAttr("dbPort"); dbPort.IsNull() {
		return fmt.Errorf("dbPort must be configure in backend config")
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
	s.DB = db

	return nil
}

// Apply save state in DB by add-only strategy.
func (s *DBState) Apply(state *State) error {
	m := make(map[string]interface{})
	sort.Stable(state.Resources)
	marshal, err := json.Marshal(state)
	util.CheckNotError(err, fmt.Sprintf("marshal state failed:%+v", state))
	err = json.Unmarshal(marshal, &m)
	util.CheckNotError(err, fmt.Sprintf("unmarshal state failed:%+v", marshal))
	m["resources"] = jsonutil.MustMarshal2String(m["resources"])
	delete(m, "gmt_create")
	delete(m, "gmt_modified")
	id, err := mapper.Insert(s.DB, []map[string]interface{}{m})
	state.ID = id
	return err
}

func (s *DBState) Delete(id string) error {
	panic("implement me")
}

func (s *DBState) GetLatestState(q *StateQuery) (*State, error) {
	where := make(map[string]interface{})

	if len(q.Tenant) == 0 {
		msg := "no Tenant in query"
		log.Errorf(msg)
		return nil, fmt.Errorf(msg)
	}
	where["global_tenant"] = q.Tenant

	if len(q.Project) == 0 {
		msg := "no Project in query"
		log.Errorf(msg)
		return nil, fmt.Errorf(msg)
	}
	where["project"] = q.Project

	if len(q.Stack) != 0 {
		where["stack"] = q.Stack
	}
	where["_orderby"] = "serial desc"

	stateDO, err := mapper.GetOne(s.DB, where)
	if errors.Is(err, scanner.ErrEmptyResult) {
		return nil, nil
	}
	res := do2Bo(stateDO)
	return res, err
}

func do2Bo(dbState *mapper.StateDO) *State {
	var resStateList []models.Resource

	// JSON is a subset of YAML. Please check FileSystemState.GetLatestState for detail explanation
	parseErr := yaml.Unmarshal([]byte(dbState.Resources), &resStateList)
	util.CheckNotError(parseErr, fmt.Sprintf("marshall stateDO.resources failed:%v", dbState.Resources))
	res := NewState()
	e := copier.Copy(res, dbState)
	util.CheckNotError(e,
		fmt.Sprintf("copy db_state to State failed. db_state:%v", jsonutil.MustMarshal2String(dbState)))
	res.Resources = resStateList
	return res
}
