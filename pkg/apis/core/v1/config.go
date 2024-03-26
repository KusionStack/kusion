package v1

const ConfigBackends = "backends"

// Config contains configurations for kusion cli, which stores in ${KUSION_HOME}/config.yaml.
type Config struct {
	// Backends contains the configurations for multiple backends.
	Backends *BackendConfigs `yaml:"backends,omitempty" json:"backends,omitempty"`
}

const (
	DefaultBackendName = "default"

	BackendCurrent            = "current"
	BackendType               = "type"
	BackendConfigItems        = "configs"
	BackendLocalPath          = "path"
	BackendMysqlDBName        = "dbName"
	BackendMysqlUser          = "user"
	BackendMysqlPassword      = "password"
	BackendMysqlHost          = "host"
	BackendMysqlPort          = "port"
	BackendGenericOssEndpoint = "endpoint"
	BackendGenericOssAK       = "accessKeyID"
	BackendGenericOssSK       = "accessKeySecret"
	BackendGenericOssBucket   = "bucket"
	BackendGenericOssPrefix   = "prefix"
	BackendS3Region           = "region"

	BackendTypeLocal = "local"
	BackendTypeMysql = "mysql"
	BackendTypeOss   = "oss"
	BackendTypeS3    = "s3"

	EnvBackendMysqlPassword = "KUSION_BACKEND_MYSQL_PASSWORD"
	EnvOssAccessKeyID       = "OSS_ACCESS_KEY_ID"
	EnvOssAccessKeySecret   = "OSS_ACCESS_KEY_SECRET"
	EnvAwsAccessKeyID       = "AWS_ACCESS_KEY_ID"
	EnvAwsSecretAccessKey   = "AWS_SECRET_ACCESS_KEY"
	EnvAwsDefaultRegion     = "AWS_DEFAULT_REGION"
	EnvAwsRegion            = "AWS_REGION"

	DefaultMysqlPort = 3306
)

// BackendConfigs contains the configuration of multiple backends and the current backend.
type BackendConfigs struct {
	// Current is the name of the current used backend.
	Current string `yaml:"current,omitempty" json:"current,omitempty"`

	// Backends contains the types and configs of multiple backends, whose key is the backend name.
	Backends map[string]*BackendConfig `yaml:",omitempty,inline" json:",omitempty,inline"`
}

// BackendConfig contains the type and configs of a backend, which is used to store Spec, State and Workspace.
type BackendConfig struct {
	// Type is the backend type, supports BackendTypeLocal, BackendTypeMysql, BackendTypeOss, BackendTypeS3.
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Configs contains config items of the backend, whose keys differ from different backend types.
	Configs map[string]any `yaml:"configs,omitempty" json:"configs,omitempty"`
}

// BackendLocalConfig contains the config of using local file system as backend, which can be converted
// from BackendConfig if Type is BackendTypeLocal.
type BackendLocalConfig struct {
	// Path of the directory to store the files.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// BackendMysqlConfig contains the config of using mysql database as backend, which can be converted
// from BackendConfig if Type is BackendMysqlConfig.
type BackendMysqlConfig struct {
	// DBName is the database name.
	DBName string `yaml:"dbName" json:"dbName"`

	// User of the database.
	User string `yaml:"user" json:"user"`

	// Password of the database.
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// Host of the database.
	Host string `yaml:"host" json:"host"`

	// Port of the database. If not set, then it will be set to DeprecatedDefaultMysqlPort.
	Port int `yaml:"port,omitempty" json:"port,omitempty"`
}

// BackendOssConfig contains the config of using OSS as backend, which can be converted from BackendConfig
// if Type is BackendOssConfig.
type BackendOssConfig struct {
	*GenericBackendObjectStorageConfig `yaml:",inline" json:",inline"` // OSS asks for non-empty endpoint
}

// BackendS3Config contains the config of using S3 as backend, which can be converted from BackendConfig
// if Type is BackendS3Config.
type BackendS3Config struct {
	*GenericBackendObjectStorageConfig `yaml:",inline" json:",inline"`

	// Region of S3.
	Region string `yaml:"region,omitempty" json:"region,omitempty"`
}

// GenericBackendObjectStorageConfig contains generic configs which can be reused by BackendOssConfig and
// BackendS3Config.
type GenericBackendObjectStorageConfig struct {
	// Endpoint of the object storage service.
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`

	// AccessKeyID of the object storage service.
	AccessKeyID string `yaml:"accessKeyID,omitempty" json:"accessKeyID,omitempty"`

	// AccessKeySecret of the object storage service.
	AccessKeySecret string `yaml:"accessKeySecret,omitempty" json:"accessKeySecret,omitempty"`

	// Bucket of the object storage service.
	Bucket string `yaml:"bucket" json:"bucket"`

	// Prefix of the key to store the files.
	Prefix string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
}

// ToLocalBackend converts BackendConfig to structured BackendLocalConfig, works only when the Type
// is BackendTypeLocal, and the Configs are with correct type, or return nil.
func (b *BackendConfig) ToLocalBackend() *BackendLocalConfig {
	if b.Type != BackendTypeLocal {
		return nil
	}
	path, _ := b.Configs[BackendLocalPath].(string)
	return &BackendLocalConfig{
		Path: path,
	}
}

// ToMysqlBackend converts BackendConfig to structured BackendMysqlConfig, works only when the Type
// is BackendTypeMysql, and the Configs are with correct type, or return nil.
func (b *BackendConfig) ToMysqlBackend() *BackendMysqlConfig {
	if b.Type != BackendTypeMysql {
		return nil
	}
	dbName, _ := b.Configs[BackendMysqlDBName].(string)
	user, _ := b.Configs[BackendMysqlUser].(string)
	password, _ := b.Configs[BackendMysqlPassword].(string)
	host, _ := b.Configs[BackendMysqlHost].(string)
	port, _ := b.Configs[BackendMysqlPort].(int)
	return &BackendMysqlConfig{
		DBName:   dbName,
		User:     user,
		Password: password,
		Host:     host,
		Port:     port,
	}
}

// ToOssBackend converts BackendConfig to structured BackendOssConfig, works only when the Type is
// BackendTypeOss, and the Configs are with correct type, or return nil.
func (b *BackendConfig) ToOssBackend() *BackendOssConfig {
	if b.Type != BackendTypeOss {
		return nil
	}
	endpoint, _ := b.Configs[BackendGenericOssEndpoint].(string)
	accessKeyID, _ := b.Configs[BackendGenericOssAK].(string)
	accessKeySecret, _ := b.Configs[BackendGenericOssSK].(string)
	bucket, _ := b.Configs[BackendGenericOssBucket].(string)
	prefix, _ := b.Configs[BackendGenericOssPrefix].(string)
	return &BackendOssConfig{
		&GenericBackendObjectStorageConfig{
			Endpoint:        endpoint,
			AccessKeyID:     accessKeyID,
			AccessKeySecret: accessKeySecret,
			Bucket:          bucket,
			Prefix:          prefix,
		},
	}
}

// ToS3Backend converts BackendConfig to structured BackendS3Config, works only when the Type is
// BackendTypeS3, and the Configs are with correct type, or return nil.
func (b *BackendConfig) ToS3Backend() *BackendS3Config {
	if b.Type != BackendTypeS3 {
		return nil
	}
	endpoint, _ := b.Configs[BackendGenericOssEndpoint].(string)
	accessKeyID, _ := b.Configs[BackendGenericOssAK].(string)
	accessKeySecret, _ := b.Configs[BackendGenericOssSK].(string)
	bucket, _ := b.Configs[BackendGenericOssBucket].(string)
	prefix, _ := b.Configs[BackendGenericOssPrefix].(string)
	region, _ := b.Configs[BackendS3Region].(string)
	return &BackendS3Config{
		GenericBackendObjectStorageConfig: &GenericBackendObjectStorageConfig{
			Endpoint:        endpoint,
			AccessKeyID:     accessKeyID,
			AccessKeySecret: accessKeySecret,
			Bucket:          bucket,
			Prefix:          prefix,
		},
		Region: region,
	}
}
