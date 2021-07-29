package model

import (
	"fmt"
	"time"
)

type Config struct {
	AutoStart bool `json:"auto_start"`
	Enabled   bool `json:"enabled"`

	// allow auto-upgrades
	AutoUpgrade bool `json:"auto_upgrade"`
	// the last time we checked for upgrade, Unix timestamp, [second]
	LastUpgradeCheck int64  `json:"last_upgrade_check"`
	LastSeenUpgrade  string `json:"last_seen_upgrade"`
}

func (c *Config) RefreshLastUpgradeCheck() {
	c.LastUpgradeCheck = time.Now().Unix()
}

// Check if 24 hours passed since last upgrade check
func (c *Config) NeedToCheckUpgrade() bool {
	t := time.Unix(c.LastUpgradeCheck, 0)
	fmt.Println("t", t, t.Add(24*time.Hour))
	return t.Add(24 * time.Hour).Before(time.Now())
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