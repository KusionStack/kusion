package accessories

import (
	v1 "k8s.io/api/core/v1"
	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
)

// generateAlicloudResources generates alicloud provided mysql database instance.
func (g *mysqlGenerator) generateAlicloudResources(db *mysql.MySQL, spec *intent.Intent) (*v1.Secret, error)
