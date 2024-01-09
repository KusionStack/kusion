package mysql

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
	"kusionstack.io/kusion/pkg/workspace"
)

const (
	errUnsupportedTFProvider = "unsupported terraform provider for mysql generator: %s"
	errUnsupportedMySQLType  = "unsupported mysql type: %s"
	errEmptyCloudInfo        = "empty cloud info in module config"
)

const (
	dbEngine         = "mysql"
	dbResSuffix      = "-mysql"
	dbHostAddressEnv = "KUSION_DB_HOST"
	dbUsernameEnv    = "KUSION_DB_USERNAME"
	dbPasswordEnv    = "KUSION_DB_PASSWORD"
)

const (
	defaultRandomProviderURL = "registry.terraform.io/hashicorp/random/3.5.1"
	randomPassword           = "random_password"
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
	project       *apiv1.Project
	stack         *apiv1.Stack
	appName       string
	workload      *workload.Workload
	mysql         *mysql.MySQL
	moduleConfigs map[string]apiv1.GenericConfig
	tfConfigs     apiv1.TerraformConfig
	namespace     string
	dbKey         string
}

// NewMySQLGenerator returns a new generator for mysql database.
func NewMySQLGenerator(ctx modules.GeneratorContext, dbKey string, db *mysql.MySQL) (modules.Generator, error) {
	return &mysqlGenerator{
		project:       ctx.Project,
		stack:         ctx.Stack,
		appName:       ctx.Application.Name,
		workload:      ctx.Application.Workload,
		mysql:         db,
		moduleConfigs: ctx.ModuleInputs,
		tfConfigs:     ctx.TerraformConfig,
		namespace:     ctx.Namespace,
		dbKey:         dbKey,
	}, nil
}

// NewMySQLGeneratorFunc returns a new generator function for
// generating a new mysql database.
func NewMySQLGeneratorFunc(ctx modules.GeneratorContext, dbKey string, db *mysql.MySQL) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewMySQLGenerator(ctx, dbKey, db)
	}
}

// Generate generates a new mysql database instance for the workload.
func (g *mysqlGenerator) Generate(spec *apiv1.Intent) error {
	if spec.Resources == nil {
		spec.Resources = make(apiv1.Resources, 0)
	}

	// Skip empty mysql database instance.
	db := g.mysql
	if db == nil {
		return nil
	}

	// Patch workspace configurations for mysql generator.
	if err := g.patchWorkspaceConfig(); err != nil {
		if !errors.Is(err, workspace.ErrEmptyModuleConfigBlock) {
			return err
		}
	}

	// Validate the complete mysql database module input.
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
		var providerType string
		providerType, err = g.getTFProviderType()
		if err != nil {
			return err
		}

		switch strings.ToLower(providerType) {
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
	mysqlCfg, ok := g.moduleConfigs[dbEngine]
	if !ok {
		g.mysql.Username = defaultUsername
		g.mysql.Category = defaultCategory
		g.mysql.SecurityIPs = defaultSecurityIPs
		g.mysql.PrivateRouting = defaultPrivateRouting
		g.mysql.Size = defaultSize
		g.mysql.DatabaseName = g.dbKey

		return workspace.ErrEmptyModuleConfigBlock
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

	if suffix, ok := mysqlCfg["suffix"]; ok {
		g.mysql.DatabaseName = g.dbKey + suffix.(string)
	} else {
		g.mysql.DatabaseName = g.dbKey
	}

	return nil
}

// getTFProviderType returns the type of terraform provider, e.g. "aws" or "alicloud", etc.
func (g *mysqlGenerator) getTFProviderType() (string, error) {
	// Get the workspace configurations for mysql database instance of the workload.
	mysqlCfg, ok := g.moduleConfigs[dbEngine]
	if !ok {
		return "", workspace.ErrEmptyModuleConfigBlock
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
			Key:   dbHostAddressEnv + "_" + strings.ToUpper(strings.ReplaceAll(g.mysql.DatabaseName, "-", "_")),
			Value: "secret://" + secret.Name + "/hostAddress",
		},
		{
			Key:   dbUsernameEnv + "_" + strings.ToUpper(strings.ReplaceAll(g.mysql.DatabaseName, "-", "_")),
			Value: "secret://" + secret.Name + "/username",
		},
		{
			Key:   dbPasswordEnv + "_" + strings.ToUpper(strings.ReplaceAll(g.mysql.DatabaseName, "-", "_")),
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
			g.workload.Job.Containers[k] = v
		}
	}

	return nil
}

// generateDBSecret generates kubernetes secret resource to store the host address,
// username and password of the mysql database instance.
func (g *mysqlGenerator) generateDBSecret(hostAddress, username, password string, spec *apiv1.Intent) (*v1.Secret, error) {
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
			Name:      g.mysql.DatabaseName + dbResSuffix,
			Namespace: g.namespace,
		},
		StringData: data,
	}

	return secret, modules.AppendToIntent(
		apiv1.Kubernetes,
		modules.KubernetesResourceID(secret.TypeMeta, secret.ObjectMeta),
		spec,
		secret,
	)
}

// generateTFRandomPassword generates terraform random_password resource as the password
// the mysql database instance.
func (g *mysqlGenerator) generateTFRandomPassword(provider *inputs.Provider) (string, apiv1.Resource) {
	pswAttrs := map[string]interface{}{
		"length":           16,
		"special":          true,
		"override_special": "_",
	}

	id := modules.TerraformResourceID(provider, randomPassword, g.mysql.DatabaseName+dbResSuffix)
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
