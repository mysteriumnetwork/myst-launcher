package util

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/native"
)

func PanicHandler(threadName string) {
	if panic := recover(); panic != nil {

		dir, _ := os.Getwd()
		log.Println(dir)

		fmt.Printf("Panic: %v\n", panic)
		fmt.Printf("Stacktrace %s: %s\n", threadName, debug.Stack())
		fname := fmt.Sprintf("%s/launcher_trace_%d.txt", dir, time.Now().Unix())
		f, err := os.Create(fname)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()

		var bu bytes.Buffer

		bu.WriteString(fmt.Sprintf("Panic: %v\n", panic))
		bu.WriteString(fmt.Sprintf("Stacktrace %s: \n", threadName))
		bu.Write(debug.Stack())
		f.Write(bu.Bytes())
	}
}

func OpenUrlInBrowser(url string) {
	native.ShellExecuteAndWait(
		0,
		"",
		"rundll32",
		"url.dll,FileProtocolHandler "+url,
		"",
		syscall.SW_NORMAL)
}

func ThisPath() (string, error) {
	thisExec, err := os.Executable()
	if err != nil {
		return "", err
	}
	thisPath, err := filepath.Abs(thisExec)
	if err != nil {
		return "", err
	}
	return thisPath, nil
}

func ReadConsole() string {
	b, _ := bufio.NewReader(os.Stdin).ReadBytes('\n')
	k := strings.TrimSuffix(string(b), "\r\n")
	return k
}
