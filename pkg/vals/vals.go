package vals

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/variantdev/vals"
)

const (
	VaultPrefix = "ref+vault://"
)

var supported = []string{
	VaultPrefix,
}

var runtime *vals.Runtime

func init() {
	r, err := vals.New(vals.Options{})
	contract.AssertNoErrorf(err, "failed to initialize vals")
	runtime = r
}

func IsSecured(str string) (string, bool) {
	for _, prefix := range supported {
		if strings.HasPrefix(str, prefix) {
			return prefix, true
		}
	}
	return "", false
}

func ParseSecretRef(prefix, src string, ss *SecretStores) (string, error) {
	params := buildParams(prefix, ss)
	fullFormat := constructURI(src, params)
	tmpMap := map[string]interface{}{
		"tmp": fullFormat,
	}
	evalMap, err := runtime.Eval(tmpMap)
	if err != nil {
		return "", err
	}
	restored := evalMap["tmp"].(string)
	return restored, nil
}

var paramsInMemory = map[string]string{}

// buildParams joints all no-nil field value with '&'
func buildParams(prefix string, ss *SecretStores) string {
	if v, ok := paramsInMemory[prefix]; ok {
		return v
	}

	ret := []string{}
	var t reflect.Type
	var v reflect.Value
	switch prefix {
	case VaultPrefix:
		contract.Requiref(ss.Vault != nil, "secret_store.vault is nil", "")
		t = reflect.TypeOf(*ss.Vault)
		v = reflect.ValueOf(*ss.Vault)
	default:
		return "" // Never reach
	}

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Len() == 0 {
			continue
		}
		ret = append(ret, fmt.Sprintf("%s=%s", t.Field(i).Tag.Get("yaml"), v.Field(i).String()))
	}

	params := strings.Join(ret, "&")
	paramsInMemory[prefix] = params

	return params
}

// constructURI transforms "ref+vault://path/to/backend#/key" to:
// "ref+vault://PATH/TO/KV_BACKEND[?address=VAULT_ADDR:PORT&token_file=PATH/TO/FILE&token_env=VAULT_TOKEN&namespace=VAULT_NAMESPACE]#/key" or
// "ref+vault://PATH/TO/KV_BACKEND[?address=VAULT_ADDR:PORT&auth_method=approle&role_id=vault_role&secret_id=vault_secret]#/key"
func constructURI(str string, params string) string {
	splits := strings.Split(str, "#")
	if len(splits) != 2 {
		panic(fmt.Sprintf("invalid format for secret ref: %s", str))
	}
	return fmt.Sprintf("%s?%s#%s", splits[0], params, splits[1])
}
