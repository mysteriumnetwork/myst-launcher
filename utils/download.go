/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package utils

import (
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

func init() {
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

type PrintProgressCallback func(progress int)

type WriteCounter struct {
	total         uint64
	contentLength uint64
	progress      int
	cb            PrintProgressCallback
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.total += uint64(n)
	newProgress := int(100 * wc.total / wc.contentLength)
	if newProgress > wc.progress {
		wc.progress = newProgress
		if wc.cb != nil {
			wc.cb(wc.progress)
		}
	}
	return n, nil
}

func DownloadFile(filepath string, url string, cb PrintProgressCallback) error {
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		out.Close()
		return err
	}
	defer resp.Body.Close()

	counter := &WriteCounter{cb: cb}
	counter.contentLength = uint64(resp.ContentLength)
	counter.cb(0)
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}
	out.Close()
	if err = os.Rename(filepath+".tmp", filepath); err != nil {
		return err
	}
	return nil
}
