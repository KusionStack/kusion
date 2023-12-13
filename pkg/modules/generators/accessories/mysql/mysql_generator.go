package accessories

import (
	"errors"
	"fmt"
	"net"
	"strings"

	v1 "k8s.io/api/core/v1"
	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	workspaceapi "kusionstack.io/kusion/pkg/apis/workspace"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
	"kusionstack.io/kusion/pkg/workspace"
)

const (
	errEmptyProjectName      = "project name must not be empty"
	errUnsupportedTFProvider = "unsupported terraform provider for mysql generator: %s"
	errUnsupportedMySQLType  = "unsupported mysql type: %s"
)

var _ modules.Generator = &mysqlGenerator{}

// mysqlGenerator implements the modules.Generator interface.
type mysqlGenerator struct {
	project  *project.Project
	stack    *stack.Stack
	appName  string
	workload *workload.Workload
	mysql    *mysql.MySQL
	ws       *workspaceapi.Workspace
}

// NewMySQLGenerator returns a new generator for mysql database.
func NewMySQLGenerator(
	project *project.Project,
	stack *stack.Stack,
	appName string,
	workload *workload.Workload,
	mysql *mysql.MySQL,
	ws *workspaceapi.Workspace,
) (modules.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf(errEmptyProjectName)
	}

	return &mysqlGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		mysql:    mysql,
		ws:       ws,
	}, nil
}

// NewMySQLGeneratorFunc returns a new generator function for
// generating a new mysql database.
func NewMySQLGeneratorFunc(
	project *project.Project,
	stack *stack.Stack,
	appName string,
	workload *workload.Workload,
	mysql *mysql.MySQL,
	ws *workspaceapi.Workspace,
) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewMySQLGenerator(project, stack, appName, workload, mysql, ws)
	}
}

// Generate generates a new mysql database instance for the workload.
func (g *mysqlGenerator) Generate(spec *intent.Intent) error {
	if spec.Resources == nil {
		spec.Resources = make(intent.Resources, 0)
	}

	// Patch workspace configurations for mysql generator.
	if err := g.patchWorkspaceConfig(); err != nil {
		if !errors.Is(err, workspace.ErrEmptyModuleConfig) {
			return err
		}
	}

	// Skip rendering for empty mysql instance.
	db := g.mysql
	if db == nil {
		return nil
	}

	var secret *v1.Secret
	var err error
	// Generate the mysql resources based on the type and provider config.
	switch strings.ToLower(db.Type) {
	case "local":
		secret, err = g.generateLocalResources(db, spec)
	case "cloud":
		providerType := g.getTFProviderType()
		switch providerType {
		case "aws":
			secret, err = g.generateAWSResources(db, spec)
		case "alicloud":
			secret, err = g.generateAlicloudResources(db, spec)
		default:
			return fmt.Errorf(errUnsupportedTFProvider, providerType)
		}
	default:
		return fmt.Errorf(errUnsupportedMySQLType, db.Type)
	}

	if err != nil {
		return err
	}

	return g.injectSecret(secret)
}

// patchWorkspaceConfig patches the config items for mysql generator in workspace configurations.
func (g *mysqlGenerator) patchWorkspaceConfig() error

// getTFProviderType returns the type of terraform provider, e.g. "aws", "alicloud" or "azure", etc.
func (g *mysqlGenerator) getTFProviderType() string

// injectSecret injects the mysql instance host address, username and password into
// the containers of the workload as environment variables with kubernetes secret.
func (g *mysqlGenerator) injectSecret(secret *v1.Secret) error

// generateDBSecret generates kubernetes secret resource to store the host address,
// username and password of the mysql database instance.
func (g *mysqlGenerator) generateDBSecret(hostAddress, username, password string, spec *intent.Intent) (*v1.Secret, error)

// generateTFRandomPassword generates terraform random_password resource as the password
// the mysql database instance.
func (g *mysqlGenerator) generateTFRandomPassword(provider *inputs.Provider) (string, intent.Resource)

// isPublicAccessible returns whether the mysql database instance is publicly
// accessible according to the securityIPs.
func isPublicAccessible(securityIPs []string) bool {
	var parsedIP net.IP
	for _, ip := range securityIPs {
		if isIPAddress(ip) {
			parsedIP = net.ParseIP(ip)
		} else if isCIDR(ip) {
			parsedIP, _, _ = net.ParseCIDR(ip)
		}

		if parsedIP != nil && !parsedIP.IsPrivate() {
			return true
		}
	}

	return false
}

// isIPAddress returns whether the input string is a valid ip address.
func isIPAddress(ipStr string) bool {
	ip := net.ParseIP(ipStr)

	return ip != nil
}

// isCIDR returns whether the input string is a valid CIDR record.
func isCIDR(cidrStr string) bool {
	_, _, err := net.ParseCIDR(cidrStr)

	return err == nil
}
