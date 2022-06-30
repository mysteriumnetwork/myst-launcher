/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package native

import (
	"errors"
	"log"
	"os/exec"
	"path"
	"runtime"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type NodeRunner struct {
	cfg     *model.Config // from main App
	binpath string
	cmd     *exec.Cmd
}

func NewRunner(cfg *model.Config) *NodeRunner {

	binpath := path.Join(utils.GetUserProfileDir(), `.mysterium-node\bin\`)
	binpath = utils.MakeCanonicalPath(binpath)
	utils.MakeDirectoryIfNotExists(binpath)

	return &NodeRunner{
		cfg:     cfg,
		binpath: binpath,
	}
}

func (r *NodeRunner) IsRunning() bool {
	exe := getNodeProcessName()
	return utils.IsProcessRunning(exe)
}

func getNodeProcessName() string {
	exe := "myst.exe"
	if runtime.GOOS == "darwin" {
		exe = "myst"
	}
	return exe
}

// return values: isRunning
func (r *NodeRunner) IsRunningOrTryStart() bool {
	exename := getNodeProcessName()

	fullpath := path.Join(r.binpath, exename)
	fullpath = utils.MakeCanonicalPath(fullpath)

	p, err := utils.IsProcessRunningExt(exename, fullpath)
	_ = err
	if p == 0 {
		return r.startNode() == nil
	}

	return true
}

func (r *NodeRunner) Stop() {
	if r.cmd != nil {
		r.cmd.Process.Kill()
		r.cmd.Wait()
	}
}

func (r *NodeRunner) startNode() error {
	log.Println("Start node native >")
	exename := getNodeProcessName()

	switch runtime.GOOS {
	case "windows":
		fullpath := path.Join(r.binpath, exename)
		fullpath = utils.MakeCanonicalPath(fullpath)

		r.cmd = exec.Command(fullpath, "--ui.enable=false", "--usermode", "--consumer", "--tequilapi.port=44050", "--discovery.type=api", "daemon")

	case "darwin":
		// cmd = exec.Command("open", "/Applications/Docker.app/")
	default:
		return errors.New("unsupported OS: " + runtime.GOOS)
	}

	var err error
	if err = r.cmd.Start(); err != nil {
		log.Println("err>", err)
		return err
	}
	return nil
}
