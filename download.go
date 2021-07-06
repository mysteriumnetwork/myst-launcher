package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type PrintProgressCallback func(progress int)

// WriteCounter counts the number of bytes written to it. It implements to the io.Writer interface
// and we can pass this into io.TeeReader() which will report progress on each write cycle.
type WriteCounter struct {
	total         uint64
	contentLength uint64
	progress      int
	cb            PrintProgressCallback
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.total += uint64(n)
	//wc.PrintProgress()
	newProgress := int(100 * wc.total / wc.contentLength)
	if newProgress > wc.progress {
		wc.progress = newProgress
		//wc.PrintProgress()
		if wc.cb != nil {
			wc.cb(wc.progress)
		}
	}

	return n, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory. We pass an io.TeeReader
// into Copy() to report progress on the download.
func DownloadFile(filepath string, url string, cb PrintProgressCallback) error {

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
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

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{cb: cb}
	counter.contentLength = uint64(resp.ContentLength)
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Print("\n")

	// Close the file without defer so it can happen before Rename()
	out.Close()

	if err = os.Rename(filepath+".tmp", filepath); err != nil {
		return err
	}
	return nil
}
