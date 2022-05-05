package pretty

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSpinnerT(t *testing.T) {
	fmt.Println("SpinnerT:")
	defer fmt.Println("")

	sp, err := SpinnerT.Start("Starting ...")
	assert.Nil(t, err)
	time.Sleep(time.Second * 1)
	sp.Success("Success")

	sp, err = SpinnerT.Start("Starting ...")
	assert.Nil(t, err)
	time.Sleep(time.Second * 1)
	sp.Fail("Fail")

	sp, err = SpinnerT.Start("Starting ...")
	assert.Nil(t, err)
	time.Sleep(time.Second * 1)
	sp.Warning("Warning")
}
