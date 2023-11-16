package accessories

import (
	"fmt"
	"net"
	"strings"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/accessories/database"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/projectstack"
)

const (
	dbResSuffix       = "-db"
	randomPassword    = "random_password"
	randomProviderURL = "registry.terraform.io/hashicorp/random/3.5.1"
	dbHostAddressEnv  = "KUSION_DB_HOST"
	dbUsernameEnv     = "KUSION_DB_USERNAME"
	dbPasswordEnv     = "KUSION_DB_PASSWORD"
)

type databaseGenerator struct {
	project  *projectstack.Project
	stack    *projectstack.Stack
	appName  string
	workload *workload.Workload
	database *database.Database
}

func NewDatabaseGenerator(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	workload *workload.Workload,
	database *database.Database,
) (appconfiguration.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	return &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}, nil
}

func NewDatabaseGeneratorFunc(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	workload *workload.Workload,
	database *database.Database,
) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewDatabaseGenerator(project, stack, appName, workload, database)
	}
}

func (g *databaseGenerator) Generate(spec *models.Intent) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	// Skip rendering for empty database instance.
	db := g.database
	if db == nil {
		return nil
	}

	var secret *v1.Secret
	var err error
	// Generate the database resources based on the type.
	switch strings.ToLower(db.Type) {
	case "aws":
		secret, err = g.generateAWSResources(db, spec)
	case "alicloud":
		secret, err = g.generateAlicloudResources(db, spec)
	case "local":
		secret, err = g.generateLocalResources(db, spec)
	default:
		return fmt.Errorf("unsupported database type: %s", db.Type)
	}

	if err != nil {
		return err
	}

	// Inject the database host address, username and password into the containers
	// of the workload as environment variables with Kubernetes Secret.
	return g.injectSecret(secret)
}

func (g *databaseGenerator) injectSecret(secret *v1.Secret) error {
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

func (g *databaseGenerator) generateTFRandomPassword(provider *models.Provider) (string, models.Resource) {
	pswAttrs := map[string]interface{}{
		"length":           16,
		"special":          true,
		"override_special": "_",
	}

	id := appconfiguration.TerraformResourceID(provider, randomPassword, g.appName+dbResSuffix)
	pvdExts := appconfiguration.ProviderExtensions(provider, nil, randomPassword)

	return id, appconfiguration.TerraformResource(id, nil, pswAttrs, pvdExts)
}

func (g *databaseGenerator) generateDBSecret(hostAddress, username, password string, spec *models.Intent) (*v1.Secret, error) {
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

	return secret, appconfiguration.AppendToSpec(
		models.Kubernetes,
		appconfiguration.KubernetesResourceID(secret.TypeMeta, secret.ObjectMeta),
		spec,
		secret,
	)
}

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

func isIPAddress(ipStr string) bool {
	ip := net.ParseIP(ipStr)

	return ip != nil
}

func isCIDR(cidrStr string) bool {
	_, _, err := net.ParseCIDR(cidrStr)

	return err == nil
}
