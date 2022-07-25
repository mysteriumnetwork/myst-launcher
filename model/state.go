/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package model

import (
	"encoding/json"
	"log"
	"os"
	"time"

	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type InitialState int

// 0 - undef,
// 1 - stage1, // after welcome dialogue; and elevation of rights
// 2 - stage2, // after features / WSL update; and restart
// 3 - first run after install (notify)
// 4 - normal (start minimized)

const (
	InitialStateUndefined            = InitialState(0)
	InitialStateStage1               = InitialState(1)
	InitialStateStage2               = InitialState(2)
	InitialStateFirstRunAfterInstall = InitialState(3)
	InitialStateNormalRun            = InitialState(4)
)

type Config struct {
	AutoStart              bool `json:"auto_start"`
	Enabled                bool `json:"enabled"`
	CheckVMSettingsConfirm bool `json:"check_vm_settings_confirm"`

	InitialState InitialState `json:"state"`

	// autoupgrade node
	AutoUpgrade    bool   `json:"auto_upgrade"`
	NodeExeDigest  string `json:"node_exe_digest"`
	NodeExeVersion string `json:"node_exe_version"`
	NodeLatestTag  string `json:"node_latest_tag"` // cache latest tag
	// the last time we checked for upgrade of Myst / Exe, Unix timestamp, [second]
	LastUpgradeCheck int64  `json:"last_upgrade_check"` // once a day
	Backend          string `json:"backend"`            // runner: docker | native

	// Networking mode
	EnablePortForwarding bool `json:"enable_port_forwarding"`
	PortRangeBegin       int  `json:"port_range_begin"`
	PortRangeEnd         int  `json:"port_range_end"`

	Network string `json:"network"`
}

func (c *Config) GetLatestImageTag() string {
	if c.Network == "" {
		return "latest"
	}
	return c.Network
}

func (c *Config) GetFullImageName() string {
	return _const.ImageNamePrefix + ":" + c.GetLatestImageTag()
}

func (c *Config) GetImageNamePrefix() string {
	return _const.ImageNamePrefix
}

func (c *Config) GetNetworkCaption() string {
	switch c.Network {
	case "mainnet":
		return "MainNet"
	case "testnet3":
		return "TestNet3"
	default:
		return "MainNet"
	}
}

func (c *Config) RefreshLastUpgradeCheck() {
	c.LastUpgradeCheck = time.Now().Unix()
}

const upgradeCheckPeriod = 24 * time.Hour

// Check if 24 hours passed since last upgrade check
func (c *Config) NeedToCheckUpgrade() bool {
	t := time.Unix(c.LastUpgradeCheck, 0)
	return t.Add(upgradeCheckPeriod).Before(time.Now())
}

func (c *Config) getDefaultValues() {
	c.Enabled = true
	c.EnablePortForwarding = false
	c.PortRangeBegin = 42000
	c.PortRangeEnd = 42100
	c.Backend = "native"
}

func (c *Config) Read() {
	f := utils.GetUserProfileDir() + "/.myst_node_launcher"
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		c.getDefaultValues()
		c.Save()
		return
	}

	file, err := os.Open(f)
	if err != nil {
		return
	}

	c.getDefaultValues()
	json.NewDecoder(file).Decode(&c)
}

func (c *Config) Save() {
	f := utils.GetUserProfileDir() + "/.myst_node_launcher"
	file, err := os.Create(f)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", " ")
	enc.Encode(&c)
}
