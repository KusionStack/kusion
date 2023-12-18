package accessories

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	errEmptyCloudInfo        = "empty cloud info in module config"
)

const (
	dbResSuffix      = "-db"
	dbEngine         = "mysql"
	dbHostAddressEnv = "KUSION_DB_HOST"
	dbUsernameEnv    = "KUSION_DB_USERNAME"
	dbPasswordEnv    = "KUSION_DB_PASSWORD"
)

const (
	randomPassword = "random_password"
)

var (
	defaultUsername       string   = "root"
	defaultCategory       string   = "Basic"
	defaultSecurityIPs    []string = []string{"0.0.0.0/0"}
	defaultPrivateRouting bool     = true
	defaultSize           int      = 10
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
		if !errors.Is(err, workspace.ErrEmptyModuleConfigBlock) {
			return err
		}
	}

	// Skip empty or invalid mysql database instance.
	db := g.mysql
	if db == nil {
		return nil
	}
	if err := db.Validate(); err != nil {
		return err
	}

	var secret *v1.Secret
	var err error
	// Generate the mysql resources based on the type and provider config.
	switch strings.ToLower(db.Type) {
	case mysql.LocalDBType:
		secret, err = g.generateLocalResources(db, spec)
	case mysql.CloudDBType:
		providerType, err := g.getTFProviderType()
		if err != nil {
			return err
		}

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
func (g *mysqlGenerator) patchWorkspaceConfig() error {
	// Get the workspace configurations for mysql database instance of the workload.
	mysqlCfgs, ok := g.ws.Modules[dbEngine]
	if !ok {
		return workspace.ErrEmptyModuleConfigBlock
	}

	mysqlCfg, err := workspace.GetProjectModuleConfig(mysqlCfgs, g.project.Name)
	if err != nil {
		return err
	}

	// Patch workspace configurations for mysql generator.
	if username, ok := mysqlCfg["username"]; ok {
		g.mysql.Username = username.(string)
	} else {
		g.mysql.Username = defaultUsername
	}

	if category, ok := mysqlCfg["category"]; ok {
		g.mysql.Category = category.(string)
	} else {
		g.mysql.Category = defaultCategory
	}

	if securityIPs, ok := mysqlCfg["securityIPs"]; ok {
		g.mysql.SecurityIPs = securityIPs.([]string)
	} else {
		g.mysql.SecurityIPs = defaultSecurityIPs
	}

	if privateRouting, ok := mysqlCfg["privateRouting"]; ok {
		g.mysql.PrivateRouting = privateRouting.(bool)
	} else {
		g.mysql.PrivateRouting = defaultPrivateRouting
	}

	if size, ok := mysqlCfg["size"]; ok {
		g.mysql.Size = size.(int)
	} else {
		g.mysql.Size = defaultSize
	}

	if instanceType, ok := mysqlCfg["instanceType"]; ok {
		g.mysql.InstanceType = instanceType.(string)
	}

	if subnetID, ok := mysqlCfg["subnetID"]; ok {
		g.mysql.SubnetID = subnetID.(string)
	}

	return nil
}

// getTFProviderType returns the type of terraform provider, e.g. "aws", "alicloud" or "azure", etc.
func (g *mysqlGenerator) getTFProviderType() (string, error) {
	// Get the workspace configurations for mysql database instance of the workload.
	mysqlCfgs, ok := g.ws.Modules[dbEngine]
	if !ok {
		return "", workspace.ErrEmptyModuleConfigBlock
	}

	mysqlCfg, err := workspace.GetProjectModuleConfig(mysqlCfgs, g.project.Name)
	if err != nil {
		return "", err
	}

	if cloud, ok := mysqlCfg["cloud"]; ok {
		return cloud.(string), nil
	}

	return "", fmt.Errorf(errEmptyCloudInfo)
}

// injectSecret injects the mysql instance host address, username and password into
// the containers of the workload as environment variables with kubernetes secret.
func (g *mysqlGenerator) injectSecret(secret *v1.Secret) error {
	secEnvs := yaml.MapSlice{
		{
			Key:   dbHostAddressEnv,
			Value: "secret://" + secret.Name + "/hostAddress",
		},
		{
			Key:   dbUsernameEnv,
			Value: "secret://" + secret.Name + "/username",
		},
		{
			Key:   dbPasswordEnv,
			Value: "secret://" + secret.Name + "/password",
		},
	}

	// Inject the database information into the containers of service/job workload.
	if g.workload.Service != nil {
		for k, v := range g.workload.Service.Containers {
			v.Env = append(secEnvs, v.Env...)
			g.workload.Service.Containers[k] = v
		}
	} else if g.workload.Job != nil {
		for k, v := range g.workload.Job.Containers {
			v.Env = append(secEnvs, v.Env...)
			g.workload.Service.Containers[k] = v
		}
	}

	return nil
}

// generateDBSecret generates kubernetes secret resource to store the host address,
// username and password of the mysql database instance.
func (g *mysqlGenerator) generateDBSecret(hostAddress, username, password string, spec *intent.Intent) (*v1.Secret, error) {
	// Create the data map of k8s secret storing the database host address, username
	// and password.
	data := make(map[string]string)
	data["hostAddress"] = hostAddress
	data["username"] = username
	data["password"] = password

	// Create the k8s secret and append to the spec.
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.appName + dbResSuffix,
			Namespace: g.project.Name,
		},
		StringData: data,
	}

	return secret, modules.AppendToIntent(
		intent.Kubernetes,
		modules.KubernetesResourceID(secret.TypeMeta, secret.ObjectMeta),
		spec,
		secret,
	)
}

// generateTFRandomPassword generates terraform random_password resource as the password
// the mysql database instance.
func (g *mysqlGenerator) generateTFRandomPassword(provider *inputs.Provider) (string, intent.Resource) {
	pswAttrs := map[string]interface{}{
		"length":           16,
		"special":          true,
		"override_special": "_",
	}

	id := modules.TerraformResourceID(provider, randomPassword, g.appName+dbResSuffix)
	pvdExts := modules.ProviderExtensions(provider, nil, randomPassword)

	return id, modules.TerraformResource(id, nil, pswAttrs, pvdExts)
}

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
