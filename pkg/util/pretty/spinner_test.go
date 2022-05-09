package pretty

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSpinner(t *testing.T) {
	fmt.Println("Spinner:")
	defer fmt.Println("")

	sp, err := Spinner.Start("Starting ...")
	assert.Nil(t, err)
	time.Sleep(time.Second * 1)
	sp.Success("Success")

	sp, err = Spinner.Start("Starting ...")
	assert.Nil(t, err)
	time.Sleep(time.Second * 1)
	sp.Fail("Fail")

	sp, err = Spinner.Start("Starting ...")
	assert.Nil(t, err)
	time.Sleep(time.Second * 1)
	sp.Warning("Warning")
}
