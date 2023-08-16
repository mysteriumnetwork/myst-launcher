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
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

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

	configpath := getNodeDirPath(`.mysterium`)
	if _, err := os.Stat(configpath); os.IsNotExist(err) {
		configpath = getNodeDirPath(`.mysterium-node`)
	}
	utils.MakeDirectoryIfNotExists(configpath)

	return &NodeRunner{
		mod:        mod,
		binpath:    binpath,
		configpath: configpath,
	}
}

func getNodeDirPath(profileDir string) string {
	path := path.Join(utils.GetUserProfileDir(), profileDir)
	path = utils.MakeCanonicalPath(path)
	return path
}

func getNodeBinDirPath() string {
	return getNodeDirPath(`.mysterium-bin`)
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

	if r.isRunning() == 0 {
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

// Kill all launcher instances except current
func KillPreviousLauncher() {
	exename := getNodeProcessName()
	err := utils.KillProcessByName(exename)
	if err != nil {
		log.Println("KillAllByName failed:", err)
	}
}

func (r *NodeRunner) startNode() error {
	log.Println("!startNode")
	fullExePath := getNodeExePath()

	versionArg := fmt.Sprintf("--launcher.ver=%s", r.mod.GetProductVersionString())
	configDirArg := fmt.Sprintf("--config-dir=%s", r.configpath)
	dataDirArg := fmt.Sprintf("--data-dir=%s", r.configpath)
	logDirArg := fmt.Sprintf("--log-dir=%s", r.configpath)
	nodeuiDirArg := fmt.Sprintf("--node-ui-dir=%s", path.Join(r.configpath, "nodeui"))

	userspaceArg := "--userspace"

	args2 := make([]string, 0)
	if r.mod.NodeFlags != "" {
		args2 = strings.Split(r.mod.NodeFlags, " ")
	}

	args := []string{userspaceArg, versionArg}
	if len(args2) > 0 {
		args = append(args, args2...)
	} else {
		args = append(args, configDirArg, dataDirArg, logDirArg, nodeuiDirArg, "service", "--agreed-terms-and-conditions")
	}

	switch runtime.GOOS {
	case "windows", "darwin":
		cmd, err := utils.CmdStart(fullExePath, args...)
		if err != nil {
			log.Println("run node failed:", err)
			return err
		}
		r.cmd = cmd

	default:
		return errors.New("unsupported OS: " + runtime.GOOS)
	}

	return nil
}
