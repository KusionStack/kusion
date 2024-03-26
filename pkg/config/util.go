package config

import (
	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

// GetConfig returns the structured config stored in the config file. The validation of the config is not checked.
func GetConfig() (*v1.Config, error) {
	o, err := newOperator()
	if err != nil {
		return nil, err
	}
	return o.config, nil
}

// GetEncodedConfigItem gets the value of config item after encoding from the config file, where the key is the
// corresponding config YAML keys connected by dot. The validation of the config is not checked.
func GetEncodedConfigItem(key string) (string, error) {
	o, err := newOperator()
	if err != nil {
		return "", err
	}
	return o.getEncodedConfigItem(key)
}

// SetEncodedConfigItem sets the config item with the encoded value in the config file, where the key is the
// corresponding config YAML keys connected by dot. The validation of the config is checked.
func SetEncodedConfigItem(key, strValue string) error {
	o, err := newOperator()
	if err != nil {
		return err
	}
	if err = o.setEncodedConfigItem(key, strValue); err != nil {
		return err
	}
	return o.writeConfig()
}

// DeleteConfigItem deletes the config item in the config file, where the key is the corresponding config YAML keys
// connected by dot. The validation of the config is not checked.
func DeleteConfigItem(key string) error {
	o, err := newOperator()
	if err != nil {
		return err
	}
	if err = o.deleteConfigItem(key); err != nil {
		return err
	}
	return o.writeConfig()
}
