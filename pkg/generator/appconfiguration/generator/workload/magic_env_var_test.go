package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMagicEnvVar(t *testing.T) {
	// Test raw environment variable
	rawEnv := MagicEnvVar("key", "value")
	assert.NotNil(t, rawEnv, "Raw environment variable should not be nil")
	assert.Equal(t, "key", rawEnv.Name, "Expected raw environment variable name to be 'key'")
	assert.Equal(t, "value", rawEnv.Value, "Expected raw environment variable value to be 'value'")

	// Test secret-based environment variable
	secretEnv := MagicEnvVar("secret_key", "secret://my_secret/my_key")
	assert.NotNil(t, secretEnv, "Secret-based environment variable should not be nil")
	assert.Equal(t, "secret_key", secretEnv.Name, "Expected secret-based environment variable name to be 'secret_key'")
	assert.Equal(t, "my_secret", secretEnv.ValueFrom.SecretKeyRef.LocalObjectReference.Name, "Expected secret name to be 'my_secret'")
	assert.Equal(t, "my_key", secretEnv.ValueFrom.SecretKeyRef.Key, "Expected secret key to be 'my_key'")
}
