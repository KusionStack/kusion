package entity

import (
	"reflect"
	"testing"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func TestSecretStore_ConvertToKusionSecretStore(t *testing.T) {
	secretStore := &SecretStore{
		ProviderType: SecretProviderTypeAliCloud,
		Provider: map[string]any{
			"accessKey": "access-key",
			"secretKey": "secret-key",
			"region":    "cn-hangzhou",
		},
	}

	expected := &v1.SecretStore{
		Provider: &v1.ProviderSpec{
			Alicloud: &v1.AlicloudProvider{
				Region: "cn-hangzhou",
			},
		},
	}

	result, err := secretStore.ConvertToKusionSecretStore()
	if err != nil {
		t.Errorf("ConvertToKusionSecretStore() returned an unexpected error: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ConvertToKusionSecretStore() returned unexpected result.\nExpected: %v\nGot: %v", expected, result)
	}
}
