package model

import (
	"time"
)

type Config struct {
	AutoStart              bool `json:"auto_start"`
	Enabled                bool `json:"enabled"`
	CheckVMSettingsConfirm bool `json:"check_vm_settings_confirm"`

	// allow auto-upgrades
	AutoUpgrade bool `json:"auto_upgrade"`
	// the last time we checked for upgrade, Unix timestamp, [second]
	LastUpgradeCheck int64 `json:"last_upgrade_check"` // once a day
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

type AppInterface interface {
	ReadConfig()
	SaveConfig()

	Publish(topic string, args ...interface{})
	Subscribe(topic string, fn interface{}) error
	TriggerAction(action string)

	GetInTray() bool
	GetConfig() *Config
	GetImageName() string
}
