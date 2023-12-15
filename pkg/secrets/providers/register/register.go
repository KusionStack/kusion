package register

// packages imported here are registered to the secret store provider registry.

import (
	_ "kusionstack.io/kusion/pkg/secrets/providers/alicloud/secretsmanager"
	_ "kusionstack.io/kusion/pkg/secrets/providers/aws/secretsmanager"
	_ "kusionstack.io/kusion/pkg/secrets/providers/hashivault"
)
