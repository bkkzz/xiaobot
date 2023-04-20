package main

import (
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

func main() {
	wAddr := flag.String("w", "127.0.0.1:9997", "webAddr")
	flag.Parse()
	h := Router()
	go func() {
		time.Sleep(time.Second)
		_ = open(fmt.Sprintf("http://%s/", *wAddr))
	}()
	err := http.ListenAndServe(*wAddr, h)
	if err != nil {
		panic(err)
	}
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var (
		cmd  string
		args []string
	)

	switch runtime.GOOS {
	case "windows":
		cmd, args = "cmd", []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		// "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
