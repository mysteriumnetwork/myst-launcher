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
	mod     *model.UIModel // from main App
	binpath string
	cmd     *exec.Cmd
}

func NewRunner(mod *model.UIModel) *NodeRunner {

	binpath := path.Join(utils.GetUserProfileDir(), `.mysterium-node`, `bin`)
	binpath = utils.MakeCanonicalPath(binpath)
	utils.MakeDirectoryIfNotExists(binpath)

	return &NodeRunner{
		mod:     mod,
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
	exename := getNodeProcessName()
	fullpath := path.Join(r.binpath, exename)
	fullpath = utils.MakeCanonicalPath(fullpath)

	const reportLauncherVersionFlag = "--launcher.ver"
	versionArg := fmt.Sprintf("%s=%s", reportLauncherVersionFlag, r.mod.GetProductVersionString())

	switch runtime.GOOS {
	case "windows":

		r.cmd = exec.Command(fullpath, versionArg /*"--userspace",*/, "service", "--agreed-terms-and-conditions")

	case "darwin":
		// cmd = exec.Command("open", "/Applications/Docker.app/")
		r.cmd = exec.Command(fullpath, versionArg /*"--userspace",*/, "service", "--agreed-terms-and-conditions")

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
