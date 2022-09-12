package docker

import (
	"github.com/blang/semver/v4"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCompare(t *testing.T) {

	{
		fileVer := "5.10.16"
		installedVer := "5.10.16"

		semverFileVer, err := semver.Parse(fileVer)
		assert.NoError(t, err)
		semverInstalledVer, err := semver.Parse(installedVer)
		assert.NoError(t, err)

		log.Println("IsWSLUpdated > semverFileVer, semverInstalledVer >", semverFileVer, semverInstalledVer)
		assert.Equal(t, 0, semverInstalledVer.Compare(semverFileVer))
	}

	{
		fileVer := "5.11.16"
		installedVer := "5.11.16"

		semverFileVer, err := semver.Parse(fileVer)
		assert.NoError(t, err)
		semverInstalledVer, err := semver.Parse(installedVer)
		assert.NoError(t, err)

		log.Println("IsWSLUpdated > semverFileVer, semverInstalledVer >", semverFileVer, semverInstalledVer)
		assert.Equal(t, true, semverInstalledVer.Compare(semverFileVer) >= 0)
	}

}
