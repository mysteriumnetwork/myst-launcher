/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package utils

import (
	"fmt"
	"time"

	"github.com/cavaliercoder/grab"
)

type PrintProgressCallback func(progress int)

func DownloadFile(filepath string, url string, cb PrintProgressCallback) error {
	client := grab.NewClient()
	req, _ := grab.NewRequest(filepath, url)
	// workaround for ErrBadLength
	req.NoResume = true

	resp := client.Do(req)
	fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	t := time.NewTicker(1000 * time.Millisecond)
	defer t.Stop()
	percent := 0
	cb(percent)
Loop:
	for {
		select {
		case <-t.C:
			if int(100*resp.Progress()) > percent {
				percent = int(100 * resp.Progress())
				cb(percent)
			}

		case <-resp.Done:
			if int(100*resp.Progress()) > percent {
				percent = int(100 * resp.Progress())
				cb(percent)
			}
			break Loop
		}
	}

	return resp.Err()
}
