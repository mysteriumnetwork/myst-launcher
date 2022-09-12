//go:build windows
// +build windows

package native

import (
	"log"

	"github.com/artdarek/go-unzip"
	wapi "github.com/iamacarpet/go-win64api"
)

func extractNodeBinary(src, dest string) error {
	return unzip.New(src, dest).Extract()
}

const (
	fwRuleNameTCP = "myst_launcher_tcp"
	fwRuleNameUDP = "myst_launcher_udp"
)

func CheckAndInstallFirewallRules() {
	fullExe := getNodeExePath()

	//rule, err := winapi.FirewallIsEnabled(winapi.NET_FW_PROFILE2_PUBLIC|winapi.NET_FW_PROFILE2_PRIVATE)
	rule, err := wapi.FirewallRuleGet(fwRuleNameUDP)
	if err != nil || rule.Name == "" {
		_, err := wapi.FirewallRuleCreate(fwRuleNameUDP, "", "", fullExe, "*", wapi.NET_FW_IP_PROTOCOL_UDP)
		log.Println(err)
	}
	rule, err = wapi.FirewallRuleGet(fwRuleNameTCP)
	if err != nil || rule.Name == "" {
		_, err := wapi.FirewallRuleCreate(fwRuleNameTCP, "", "", fullExe, "*", wapi.NET_FW_IP_PROTOCOL_TCP)
		log.Println(err)
	}
}
