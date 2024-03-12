package storages

import (
	"strconv"

	gomysql "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/engine/state"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
)

// MysqlStorage is an implementation of backend.Backend which uses mysql as storage.
type MysqlStorage struct {
	db *gorm.DB
}

func NewMysqlStorage(config *v1.BackendMysqlConfig) (*MysqlStorage, error) {
	c := gomysql.NewConfig()
	c.User = config.User
	c.Passwd = config.Password
	c.Addr = config.Host + ":" + strconv.Itoa(config.Port)
	c.DBName = config.DBName
	c.Net = "tcp"
	c.ParseTime = true
	c.InterpolateParams = true
	c.Params = map[string]string{
		"charset": "utf8",
		"loc":     "Asia/Shanghai",
	}
	db, err := gorm.Open(mysql.Open(c.FormatDSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &MysqlStorage{db: db}, nil
}

func (s *MysqlStorage) StateStorage(project, stack, workspace string) state.Storage {
	return statestorages.NewMysqlStorage(s.db, project, stack, workspace)
}
