/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package native

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"path"
	"runtime"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type NodeRunner struct {
	mod        *model.UIModel // from main App
	binpath    string
	configpath string

	cmd *exec.Cmd
}

func NewRunner(mod *model.UIModel) *NodeRunner {

	binpath := getNodeBinDirPath()
	utils.MakeDirectoryIfNotExists(binpath)

	configpath := path.Join(utils.GetUserProfileDir(), `.mysterium`)
	configpath = utils.MakeCanonicalPath(configpath)
	utils.MakeDirectoryIfNotExists(configpath)

	return &NodeRunner{
		mod:        mod,
		binpath:    binpath,
		configpath: configpath,
	}
}

func getNodeBinDirPath() string {
	binpath := path.Join(utils.GetUserProfileDir(), `.mysterium-bin`)
	binpath = utils.MakeCanonicalPath(binpath)
	return binpath
}

func getNodeExePath() string {
	fullExePath := path.Join(getNodeBinDirPath(), getNodeProcessName())
	fullExePath = utils.MakeCanonicalPath(fullExePath)
	return fullExePath
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

func (r *NodeRunner) isRunning() uint32 {
	exename := getNodeProcessName()
	fullpath := path.Join(r.binpath, exename)
	fullpath = utils.MakeCanonicalPath(fullpath)

	p, _ := utils.IsProcessRunningExt(exename, fullpath)
	return p
}

// return values: isRunning
func (r *NodeRunner) IsRunningOrTryStart() bool {

	p := r.isRunning()
	if p == 0 {
		return r.startNode() == nil
	}

	return true
}

func (r *NodeRunner) Stop() {
	if r.cmd != nil && r.cmd.Process != nil {
		r.cmd.Process.Kill()
		r.cmd.Wait()
	} else {
		// if process was started in prev. run of launcher

		p := r.isRunning()
		if p != 0 {
			utils.TerminateProcess(p, 0)
		}
	}
}

func (r *NodeRunner) startNode() error {
	fullExePath := getNodeExePath()

	versionArg := fmt.Sprintf("--launcher.ver=%s", r.mod.GetProductVersionString())
	configDirArg := fmt.Sprintf("--config-dir=%s", r.configpath)
	dataDirArg := fmt.Sprintf("--data-dir=%s", r.configpath)
	userspaceArg := "--userspace"

	args := []string{userspaceArg, versionArg, configDirArg, dataDirArg, "service", "--agreed-terms-and-conditions"}

	switch runtime.GOOS {
	case "windows":
		if err := utils.CmdStart(fullExePath, args...); err != nil {
			log.Println("run node failed:", err)
			return err
		}

	case "darwin":
		if err := utils.CmdStart(fullExePath, args...); err != nil {
			log.Println("run node failed:", err)
			return err
		}

	default:
		return errors.New("unsupported OS: " + runtime.GOOS)
	}

	return nil
}
