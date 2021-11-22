/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package model

import (
	"encoding/json"
	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"log"
	"os"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type Config struct {
	AutoStart              bool `json:"auto_start"`
	Enabled                bool `json:"enabled"`
	CheckVMSettingsConfirm bool `json:"check_vm_settings_confirm"`

	// allow auto-upgrades
	AutoUpgrade bool `json:"auto_upgrade"`
	// the last time we checked for upgrade, Unix timestamp, [second]
	LastUpgradeCheck int64 `json:"last_upgrade_check"` // once a day

	// Networking mode
	EnablePortForwarding bool `json:"enable_port_forwarding"`
	PortRangeBegin       int  `json:"port_range_begin"`
	PortRangeEnd         int  `json:"port_range_end"`

	Network string `json:"network"`
}

func (c *Config) GetImageTag() string {
	if c.Network == "" {
		return "latest"
	}
	return c.Network
}

func (c *Config) GetFullImageName() string {
	return _const.ImageNamePrefix + ":" + c.GetImageTag()
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
