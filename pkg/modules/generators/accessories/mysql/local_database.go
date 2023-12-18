package accessories

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
)

var (
	localDatabaseName     = "local-database"
	localSecretSuffix     = "-local-secret"
	localPVCSuffix        = "-local-pvc"
	localDeploymentSuffix = "-local-deployment"
	localServiceSuffix    = "-local-service"
	localMatchLabels      = map[string]string{"accessory": localDatabaseName}
)

// generateLocalResources generates locally deployed mysql database instance.
func (g *mysqlGenerator) generateLocalResources(db *mysql.MySQL, spec *intent.Intent) (*v1.Secret, error) {
	// Build k8s secret for local database's random password.
	password, err := g.generateLocalSecret(spec)
	if err != nil {
		return nil, err
	}

	// Build k8s persistentvolumeclaim for local database.
	if err = g.generateLocalPVC(db, spec); err != nil {
		return nil, err
	}

	// Build k8s deployment for local database.
	if err = g.generateLocalDeployment(db, spec); err != nil {
		return nil, err
	}

	// Build k8s service for local database.
	hostAddress, err := g.generateLocalService(db, spec)
	if err != nil {
		return nil, err
	}

	return g.generateDBSecret(hostAddress, db.Username, password, spec)
}

func (g *mysqlGenerator) generateLocalSecret(spec *intent.Intent) (string, error) {
	password := g.generateLocalPassword(16)

	data := make(map[string]string)
	data["password"] = password

	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.appName + dbResSuffix + localSecretSuffix,
			Namespace: g.project.Name,
		},
		StringData: data,
	}
	secID := modules.KubernetesResourceID(secret.TypeMeta, secret.ObjectMeta)

	// Fixme: return $kusion_path with `stringData.password` of local database secret id.
	return password, modules.AppendToIntent(
		intent.Kubernetes,
		secID,
		spec,
		secret,
	)
}

func (g *mysqlGenerator) generateLocalPVC(db *mysql.MySQL, spec *intent.Intent) error {
	// Create the k8s pvc with the storage size of `db.Size`.
	pvc := &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.appName + dbResSuffix + localPVCSuffix,
			Namespace: g.project.Name,
			Labels:    localMatchLabels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					v1.ResourceStorage: resource.MustParse(strconv.Itoa(db.Size) + "Gi"),
				},
			},
		},
	}

	return modules.AppendToIntent(
		intent.Kubernetes,
		modules.KubernetesResourceID(pvc.TypeMeta, pvc.ObjectMeta),
		spec,
		pvc,
	)
}

func (g *mysqlGenerator) generateLocalDeployment(db *mysql.MySQL, spec *intent.Intent) error {
	// Prepare the pod spec for specific local database.
	podSpec, err := g.generateLocalPodSpec(db)
	if err != nil {
		return err
	}

	// Create the k8s deployment for local database.
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.appName + dbResSuffix + localDeploymentSuffix,
			Namespace: g.project.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: localMatchLabels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: localMatchLabels,
				},
				Spec: podSpec,
			},
		},
	}

	return modules.AppendToIntent(
		intent.Kubernetes,
		modules.KubernetesResourceID(deployment.TypeMeta, deployment.ObjectMeta),
		spec,
		deployment,
	)
}

func (g *mysqlGenerator) generateLocalService(db *mysql.MySQL, spec *intent.Intent) (string, error) {
	// Prepare the service port for specific local database.
	svcPort := g.generateLocalSvcPort(db)

	svcName := g.appName + dbResSuffix + localServiceSuffix
	// Create the k8s service for local database.
	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: g.project.Name,
			Labels:    localMatchLabels,
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports:     svcPort,
			Selector:  localMatchLabels,
		},
	}

	return svcName, modules.AppendToIntent(
		intent.Kubernetes,
		modules.KubernetesResourceID(service.TypeMeta, service.ObjectMeta),
		spec,
		service,
	)
}

func (g *mysqlGenerator) generateLocalPodSpec(db *mysql.MySQL) (v1.PodSpec, error) {
	var env []v1.EnvVar
	var ports []v1.ContainerPort
	var volumes []v1.Volume
	var volumeMounts []v1.VolumeMount
	var podSpec v1.PodSpec

	image := dbEngine + ":" + db.Version
	secretName := g.appName + dbResSuffix + localSecretSuffix
	ports = []v1.ContainerPort{
		{
			Name:          localDatabaseName,
			ContainerPort: int32(3306),
		},
	}
	volumes = []v1.Volume{
		{
			Name: localDatabaseName,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: g.appName + dbResSuffix + localPVCSuffix,
				},
			},
		},
	}
	volumeMounts = []v1.VolumeMount{
		{
			Name:      localDatabaseName,
			MountPath: "/var/lib/mysql",
		},
	}

	if db.Username != "root" {
		env = []v1.EnvVar{
			{
				Name:  "MYSQL_USER",
				Value: db.Username,
			},
			{
				Name: "MYSQL_PASSWORD",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: secretName,
						},
						Key: "password",
					},
				},
			},
		}
	} else {
		env = []v1.EnvVar{
			{
				Name: "MYSQL_ROOT_PASSWORD",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: secretName,
						},
						Key: "password",
					},
				},
			},
		}
	}

	podSpec = v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:         localDatabaseName,
				Image:        image,
				Env:          env,
				Ports:        ports,
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	}

	return podSpec, nil
}

func (g *mysqlGenerator) generateLocalSvcPort(db *mysql.MySQL) []v1.ServicePort {
	svcPort := []v1.ServicePort{
		{
			Port: int32(3306),
		},
	}

	return svcPort
}

func (g *mysqlGenerator) generateLocalPassword(n int) string {
	hashInput := g.appName + g.project.Name + g.stack.Name
	hash := md5.Sum([]byte(hashInput))

	hashString := hex.EncodeToString(hash[:])

	return hashString[:n]
}
