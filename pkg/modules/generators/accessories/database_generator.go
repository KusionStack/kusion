package database

import (
	"fmt"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/generators/accessories/postgres"
	database "kusionstack.io/kusion/pkg/modules/inputs/accessories"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

const (
	errUnsupportedDatabaseType = "unsupported database type: %s"
)

var _ modules.Generator = &databaseGenerator{}

type databaseGenerator struct {
	project       *apiv1.Project
	stack         *apiv1.Stack
	appName       string
	workload      *workload.Workload
	database      map[string]*database.Database
	moduleConfigs map[string]apiv1.GenericConfig
	tfConfigs     apiv1.TerraformConfig
	namespace     string

	// for internal generator
	context modules.GeneratorContext
}

func NewDatabaseGenerator(ctx modules.GeneratorContext) (modules.Generator, error) {
	return &databaseGenerator{
		project:       ctx.Project,
		stack:         ctx.Stack,
		appName:       ctx.Application.Name,
		workload:      ctx.Application.Workload,
		database:      ctx.Application.Database,
		moduleConfigs: ctx.ModuleInputs,
		tfConfigs:     ctx.TerraformConfig,
		namespace:     ctx.Namespace,
		context:       ctx,
	}, nil
}

func NewDatabaseGeneratorFunc(ctx modules.GeneratorContext) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewDatabaseGenerator(ctx)
	}
}

func (g *databaseGenerator) Generate(spec *apiv1.Intent) error {
	if spec.Resources == nil {
		spec.Resources = make(apiv1.Resources, 0)
	}

	if len(g.database) > 0 {
		var gfs []modules.NewGeneratorFunc

		for dbKey, db := range g.database {
			switch db.Header.Type {
			case database.TypeMySQL:
				gfs = append(gfs, mysql.NewMySQLGeneratorFunc(g.context, dbKey, db.MySQL))
			case database.TypePostgreSQL:
				gfs = append(gfs, postgres.NewPostgresGeneratorFunc(g.context, dbKey, db.PostgreSQL))
			default:
				return fmt.Errorf(errUnsupportedDatabaseType, db.Header.Type)
			}
		}

		if err := modules.CallGenerators(spec, gfs...); err != nil {
			return err
		}
	}
	return nil
}
