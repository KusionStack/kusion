package accessories

import (
	v1 "k8s.io/api/core/v1"
	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
)

// generateLocalResources generates locally deployed mysql database instance.
func (g *mysqlGenerator) generateLocalResources(db *mysql.MySQL, spec *intent.Intent) (*v1.Secret, error)
