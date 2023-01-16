package vals

import "reflect"

type SecretStores struct {
	Vault *Vault `json:"vault,omitempty" yaml:"vault,omitempty"`
}

// A valid SecretStore must has one backend at least
func (ss *SecretStores) IsValid() bool {
	if ss == nil {
		return true
	}

	v := reflect.ValueOf(*ss)
	validStores := 0
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Interface() != nil {
			validStores++
		}
	}

	return validStores >= 1
}

// Vault supports two auth methods, token and approle.
// if token is used, TokenFile or TokeEnv is required;
// if approle is used, RoleID and SecretID are required.
// And supports two kinds of server address, Address or Proto&Host.
type Vault struct {
	Address    string `json:"address" yaml:"address"`
	Proto      string `json:"proto" yaml:"proto"`
	Host       string `json:"host" yaml:"host"`
	Namespace  string `json:"namespace" yaml:"namespace"`
	AuthMethod string `json:"auth_method" yaml:"auth_method"`
	TokenEnv   string `json:"token_env" yaml:"token_env"`
	TokenFile  string `json:"token_file" yaml:"token_file"`
	RoleID     string `json:"role_id" yaml:"role_id"`
	SecretID   string `json:"secret_id" yaml:"secret_id"`
	Version    string `json:"version" yaml:"version"`
}
