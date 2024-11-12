package server

import (
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
)

type Config struct {
	DB                 *gorm.DB
	DefaultBackend     entity.Backend
	DefaultSource      entity.Source
	Port               int
	AuthEnabled        bool
	AuthWhitelist      []string
	AuthKeyType        string
	MaxConcurrent      int
	MaxAsyncConcurrent int
	MaxAsyncBuffer     int
	LogFilePath        string
	AutoMigrate        bool
}

func NewConfig() *Config {
	return &Config{}
}
