package cmd

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewKusionctlCmd(t *testing.T) {
	NewDefaultKusionctlCommand()
}

func TestIsDevVersion(t *testing.T) {
	stableVer, _ := semver.ParseTolerant("1.0.0")
	assert.False(t, isDevVersion(stableVer))

	devVer, _ := semver.ParseTolerant("v1.0.0-dev")
	assert.True(t, isDevVersion(devVer))

	alphaVer, _ := semver.ParseTolerant("v1.0.0-alpha.1590772212+g4ff08363.dirty")
	assert.True(t, isDevVersion(alphaVer))

	betaVer, _ := semver.ParseTolerant("v1.0.0-beta.1590772212")
	assert.True(t, isDevVersion(betaVer))

	rcVer, _ := semver.ParseTolerant("v1.0.0-rc.1")
	assert.True(t, isDevVersion(rcVer))

	cmVer, _ := semver.ParseTolerant("v0.7.1+3d300d71")
	assert.True(t, isDevVersion(cmVer))
}
