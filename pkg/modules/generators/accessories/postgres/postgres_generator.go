package postgres

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
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/postgres"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
	"kusionstack.io/kusion/pkg/workspace"
)

const (
	errUnsupportedTFProvider     = "unsupported terraform provider for postgres generator: %s"
	errUnsupportedPostgreSQLType = "unsupported postgres type: %s"
	errEmptyCloudInfo            = "empty cloud info in module config"
)

const (
	dbEngine         = "postgres"
	dbResSuffix      = "-postgres"
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

var _ modules.Generator = &postgresGenerator{}

// postgresGenerator implements the modules.Generator interface.
type postgresGenerator struct {
	project       *apiv1.Project
	stack         *apiv1.Stack
	appName       string
	workload      *workload.Workload
	postgres      *postgres.PostgreSQL
	moduleConfigs map[string]apiv1.GenericConfig
	tfConfigs     apiv1.TerraformConfig
	namespace     string
	dbKey         string
}

// NewPostgreGenerator returns a new generator for postgres database.
func NewPostgresGenerator(ctx modules.GeneratorContext, dbKey string, db *postgres.PostgreSQL) (modules.Generator, error) {
	return &postgresGenerator{
		project:       ctx.Project,
		stack:         ctx.Stack,
		appName:       ctx.Application.Name,
		workload:      ctx.Application.Workload,
		postgres:      db,
		moduleConfigs: ctx.ModuleInputs,
		tfConfigs:     ctx.TerraformConfig,
		namespace:     ctx.Namespace,
		dbKey:         dbKey,
	}, nil
}

// NewPostgresGeneratorFunc returns a new generator function for
// generating a new postgres database.
func NewPostgresGeneratorFunc(ctx modules.GeneratorContext, dbKey string, db *postgres.PostgreSQL) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewPostgresGenerator(ctx, dbKey, db)
	}
}

// Generate generates a new postgres database instance for the workload.
func (g *postgresGenerator) Generate(spec *apiv1.Intent) error {
	if spec.Resources == nil {
		spec.Resources = make(apiv1.Resources, 0)
	}

	// Skip empty postgres database instance.
	db := g.postgres
	if db == nil {
		return nil
	}

	// Patch workspace configurations for postgres generator.
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
	// Generate the postgres resources based on the type and provider config.
	switch strings.ToLower(db.Type) {
	case postgres.LocalDBType:
		secret, err = g.generateLocalResources(db, spec)
	case postgres.CloudDBType:
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
		return fmt.Errorf(errUnsupportedPostgreSQLType, db.Type)
	}

	if err != nil {
		return err
	}

	return g.injectSecret(secret)
}

// patchWorkspaceConfig patches the config items for postgres generator in workspace configurations.
func (g *postgresGenerator) patchWorkspaceConfig() error {
	// Get the workspace configurations for postgres database instance of the workload.
	postgresCfg, ok := g.moduleConfigs[dbEngine]
	if !ok {
		g.postgres.Username = defaultUsername
		g.postgres.Category = defaultCategory
		g.postgres.SecurityIPs = defaultSecurityIPs
		g.postgres.PrivateRouting = defaultPrivateRouting
		g.postgres.Size = defaultSize
		g.postgres.DatabaseName = g.dbKey

		return workspace.ErrEmptyModuleConfigBlock
	}

	// Patch workspace configurations for postgres generator.
	if username, ok := postgresCfg["username"]; ok {
		g.postgres.Username = username.(string)
	} else {
		g.postgres.Username = defaultUsername
	}

	if category, ok := postgresCfg["category"]; ok {
		g.postgres.Category = category.(string)
	} else {
		g.postgres.Category = defaultCategory
	}

	if securityIPs, ok := postgresCfg["securityIPs"]; ok {
		g.postgres.SecurityIPs = securityIPs.([]string)
	} else {
		g.postgres.SecurityIPs = defaultSecurityIPs
	}

	if privateRouting, ok := postgresCfg["privateRouting"]; ok {
		g.postgres.PrivateRouting = privateRouting.(bool)
	} else {
		g.postgres.PrivateRouting = defaultPrivateRouting
	}

	if size, ok := postgresCfg["size"]; ok {
		g.postgres.Size = size.(int)
	} else {
		g.postgres.Size = defaultSize
	}

	if instanceType, ok := postgresCfg["instanceType"]; ok {
		g.postgres.InstanceType = instanceType.(string)
	}

	if subnetID, ok := postgresCfg["subnetID"]; ok {
		g.postgres.SubnetID = subnetID.(string)
	}

	if suffix, ok := postgresCfg["suffix"]; ok {
		g.postgres.DatabaseName = g.dbKey + suffix.(string)
	} else {
		g.postgres.DatabaseName = g.dbKey
	}

	return nil
}

// getTFProviderType returns the type of terraform provider, e.g. "aws" or "alicloud", etc.
func (g *postgresGenerator) getTFProviderType() (string, error) {
	// Get the workspace configurations for postgres database instance of the workload.
	postgresCfg, ok := g.moduleConfigs[dbEngine]
	if !ok {
		return "", workspace.ErrEmptyModuleConfigBlock
	}

	if cloud, ok := postgresCfg["cloud"]; ok {
		return cloud.(string), nil
	}

	return "", fmt.Errorf(errEmptyCloudInfo)
}

// injectSecret injects the postgres instance host address, username and password into
// the containers of the workload as environment variables with kubernetes secret.
func (g *postgresGenerator) injectSecret(secret *v1.Secret) error {
	secEnvs := yaml.MapSlice{
		{
			Key:   dbHostAddressEnv + "_" + strings.ToUpper(strings.ReplaceAll(g.postgres.DatabaseName, "-", "_")),
			Value: "secret://" + secret.Name + "/hostAddress",
		},
		{
			Key:   dbUsernameEnv + "_" + strings.ToUpper(strings.ReplaceAll(g.postgres.DatabaseName, "-", "_")),
			Value: "secret://" + secret.Name + "/username",
		},
		{
			Key:   dbPasswordEnv + "_" + strings.ToUpper(strings.ReplaceAll(g.postgres.DatabaseName, "-", "_")),
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
// username and password of the postgres database instance.
func (g *postgresGenerator) generateDBSecret(hostAddress, username, password string, spec *apiv1.Intent) (*v1.Secret, error) {
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
			Name:      g.postgres.DatabaseName + dbResSuffix,
			Namespace: g.project.Name,
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
// the postgres database instance.
func (g *postgresGenerator) generateTFRandomPassword(provider *inputs.Provider) (string, apiv1.Resource) {
	pswAttrs := map[string]interface{}{
		"length":           16,
		"special":          true,
		"override_special": "_",
	}

	id := modules.TerraformResourceID(provider, randomPassword, g.postgres.DatabaseName+dbResSuffix)
	pvdExts := modules.ProviderExtensions(provider, nil, randomPassword)

	return id, modules.TerraformResource(id, nil, pswAttrs, pvdExts)
}

// isPublicAccessible returns whether the postgres database instance is publicly
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
