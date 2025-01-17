package apply

import "fmt"

func WrappedErr(err error, context string) error {
	return fmt.Errorf("%s: %w", context, err)
}
