package winutil

import (
	"fmt"
	"log"

	"golang.org/x/sys/windows/registry"
)

func GetWindowsVersion() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	releaseIdStr, _, err := k.GetStringValue("ProductName")
	if err != nil {
		log.Fatal(err)
	}
	сurrentBuildStr, _, err := k.GetStringValue("CurrentBuild")
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%s (build %s)", releaseIdStr, сurrentBuildStr)
}
