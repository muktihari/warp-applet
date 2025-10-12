package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/muktihari/warp-applet/internal/applet"
	"github.com/muktihari/warp-applet/internal/launcher"
)

func main() {
	var cleanup = flag.Bool("cleanup", false, "remove all files created by applet: launcher, icon, autostart and lock file")
	flag.Parse()

	if *cleanup {
		launcher.Cleanup()
		applet.ForceUnlock()
		return
	}

	quitc := make(chan os.Signal, 1)
	signal.Notify(quitc, os.Interrupt, syscall.SIGTERM)

	execPath, err := whichExecutable()
	if err != nil {
		fmt.Println(err)
		return
	}

	l := launcher.New(execPath)
	a := applet.New(l)

	go func() {
		<-quitc
		a.Quit()
	}()

	a.Launch()
}

func whichExecutable() (path string, err error) {
	path, err = os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not get executable path: %v", err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("could not eval symbolic link of the executable path: %v", err)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("could not resolve absolute path of the executable: %v", err)
	}
	return path, nil
}
