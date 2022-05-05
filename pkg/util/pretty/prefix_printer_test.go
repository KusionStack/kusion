package pretty

import (
	"fmt"
	"testing"
)

func TestPrefixPrinter(t *testing.T) {
	fmt.Println("PrefixPrinter:")
	defer fmt.Println("")

	Info.Println("This is a Info message")
	Warning.Println("This is a Warning message")
	Success.Println("This is a Success message")
	Error.Println("This is a Error message")
	FatalN.Println("This is a FatalN message")
	Debug.Println("This is a Debug message")
}
