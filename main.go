package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/muktihari/warp-applet/internal/applet"
	"github.com/muktihari/warp-applet/internal/icon"
	"github.com/muktihari/warp-applet/internal/launcher"
)

const lockFilePath = "/tmp/warp-applet.lock"

func main() {
	var (
		createLauncher = flag.Bool("create-launcher", false, "create launcher")
		autostart      = flag.Bool("autostart", false, "enable autostart on login")
		cleanup        = flag.Bool("cleanup", false, "remove all files created by applet: launcher, icon, autostart and lock file")
	)
	flag.Parse()

	switch {
	case *createLauncher, *autostart:
		execPath, err := whichExecutable()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Which executable: %q\n", execPath)
		if *createLauncher {
			if err := launcher.Create(execPath, icon.Connected); err != nil {
				fmt.Println(err)
				return
			}
		}
		if *autostart {
			if err := launcher.Autostart(execPath); err != nil {
				fmt.Println(err)
				return
			}
		}
	case *cleanup:
		launcher.Cleanup(lockFilePath)
	default:
		// Launch if not flags specified
		quitSignal := make(chan os.Signal, 1)
		signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)

		a := applet.New(lockFilePath, icon.Connected, icon.Disconnected, icon.UnknownState)

		go func() {
			<-quitSignal
			a.Quit()
		}()

		a.Launch()
	}
}

// whichExecutable returns absolute path of warp-applet executable.
// If not found, it falls back to this executable path instead.
func whichExecutable() (path string, err error) {
	var buf strings.Builder
	cmd := exec.Command("which", "warp-applet")
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err = cmd.Run(); err != nil {
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

	return strings.TrimRight(buf.String(), "\n"), nil
}
