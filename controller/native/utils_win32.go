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

func getFWRuleNameTCP(ver string) string {
	return "myst_launcher" + ver + "_tcp"
}

func getFWRuleNameUDP(ver string) string {
	return "myst_launcher" + ver + "_udp"
}

// where
// ver - version of launcher: "" - legacy, "2" - new
// fullExe - myst exe path
func CheckAndInstallFirewallRules(ver, fullExe string) {

	//rule, err := winapi.FirewallIsEnabled(winapi.NET_FW_PROFILE2_PUBLIC|winapi.NET_FW_PROFILE2_PRIVATE)
	fwRuleNameUDP := getFWRuleNameUDP(ver)
	rule, err := wapi.FirewallRuleGet(fwRuleNameUDP)
	if err != nil || rule.Name == "" {
		_, err := wapi.FirewallRuleCreate(fwRuleNameUDP, "", "", fullExe, "*", wapi.NET_FW_IP_PROTOCOL_UDP)
		log.Println(err)
	}
	fwRuleNameTCP := getFWRuleNameTCP(ver)
	rule, err = wapi.FirewallRuleGet(fwRuleNameTCP)
	if err != nil || rule.Name == "" {
		_, err := wapi.FirewallRuleCreate(fwRuleNameTCP, "", "", fullExe, "*", wapi.NET_FW_IP_PROTOCOL_TCP)
		log.Println(err)
	}
}

// returns true if some firewall rules to be setup
func checkFirewallRules() bool {
	fwRuleNameUDP := getFWRuleNameUDP("")
	fwRuleNameTCP := getFWRuleNameTCP("")

	rule, err := wapi.FirewallRuleGet(fwRuleNameUDP)
	if err != nil || rule.Name == "" {
		return true
	}

	rule, err = wapi.FirewallRuleGet(fwRuleNameTCP)
	if err != nil || rule.Name == "" {
		return true
	}
	return false
}
