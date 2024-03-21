package util

import (
	"errors"
)

var (
	ErrNotNoArgs  = errors.New("no arg is accepted")
	ErrNotOneArgs = errors.New("only one arg is accepted")
	ErrNotTwoArgs = errors.New("only two args are accepted")
	ErrEmptyItem  = errors.New("empty config item name")
	ErrEmptyValue = errors.New("empty config item value")
)

// GetItemFromArgs returns config item name specified by args.
func GetItemFromArgs(args []string) (string, error) {
	if len(args) != 1 {
		return "", ErrNotOneArgs
	}
	return args[0], nil
}

// GetItemValueFromArgs returns config item name and value specified by args.
func GetItemValueFromArgs(args []string) (string, string, error) {
	if len(args) != 2 {
		return "", "", ErrNotTwoArgs
	}
	return args[0], args[1], nil
}

// ValidateNoArg returns true if there is no arg.
func ValidateNoArg(args []string) error {
	if len(args) != 0 {
		return ErrNotNoArgs
	}
	return nil
}

// ValidateItem returns the config item name is valid or not.
func ValidateItem(item string) error {
	if item == "" {
		return ErrEmptyItem
	}
	return nil
}

// ValidateValue returns the config item value is valid or not.
func ValidateValue(value string) error {
	if value == "" {
		return ErrEmptyValue
	}
	return nil
}
