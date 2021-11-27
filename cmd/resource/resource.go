/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"image"
	"io/ioutil"
	"log"
	"os"

	"github.com/tc-hib/winres"
	"github.com/tc-hib/winres/version"
)

var (
	// fixed non-string version. used for launcher version checks
	intVersion = [4]uint16{1, 0, 22, 0}
	// display version
	strVersion = "1.0.22"
)

func getIcon(path string) *winres.Icon {
	// Make an icon group from a png file
	f, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatalln(err)
	}
	f.Close()
	icon, _ := winres.NewIconFromResizedImage(img, nil)
	return icon
}

func main() {
	// First create an empty resource set
	rs := winres.ResourceSet{}

	icon := getIcon("ico/icon_512x512.png")
	icon2 := getIcon("ico/windows-active.png")

	// Add the icon to the resource set, as "APPICON"
	rs.SetIcon(winres.Name("APPICON"), icon)
	rs.SetIcon(winres.Name("ICON_ACTIVE"), icon2)

	// Make a VersionInfo structure
	vi := version.Info{
		FileVersion:    intVersion,
		ProductVersion: intVersion,
	}
	vi.Set(0x0409, version.ProductVersion, strVersion)
	vi.Set(0x0409, version.ProductName, "Mysterium Network Node Launcher")
	vi.Set(0x0409, version.CompanyName, "Mysterium Network")
	vi.Set(0x0409, version.LegalCopyright, "Copyright \u00a9 2021 Mysterium Network")

	// Add the VersionInfo to the resource set
	rs.SetVersionInfo(vi)

	// Add a manifest
	rs.SetManifest(winres.AppManifest{
		DPIAwareness:        winres.DPIPerMonitorV2,
		UseCommonControlsV6: true,
	})

	b, err := ioutil.ReadFile("ico/spinner-18px.gif") // just pass the file name
	if err != nil {
		log.Fatalln(err)
	}
	rs.Set(winres.RT_RCDATA, winres.Name("SPINNER"), 0, b)

	// Create an object file for amd64
	out, err := os.Create("cmd/app/resource.syso")
	if err != nil {
		log.Fatalln(err)
	}
	defer out.Close()

	err = rs.WriteObject(out, winres.ArchAMD64)
	if err != nil {
		log.Fatalln(err)
	}
}
