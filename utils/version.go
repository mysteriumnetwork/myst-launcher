package utils

import (
	"os"

	"github.com/mysteriumnetwork/go-fileversion"
)

func GetProductVersion() (fileversion.Info, error) {
	fullExe_, err := os.Executable()
	if err != nil {
		return fileversion.Info{}, err
	}
	return fileversion.New(fullExe_)
}
