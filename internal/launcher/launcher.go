package launcher

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/muktihari/warp-applet/internal/icon"
)

const (
	name    = "Cloudflare WARP Applet"
	dirPerm = 0o775
	perm    = 0o644
)

var (
	home = os.Getenv("HOME")

	iconDir  = filepath.Join(home, ".local", "share", "icons")
	iconPath = filepath.Join(iconDir, "warp-applet.png")

	launcherDir  = filepath.Join(home, ".local", "share", "applications")
	launcherPath = filepath.Join(launcherDir, "warp-applet.desktop")

	autostartDir  = filepath.Join(home, ".config", "autostart")
	autostartPath = filepath.Join(autostartDir, "warp-applet.desktop")
)

type Launcher struct {
	buf      []byte
	execPath string
}

// New creates new launcher instance.
func New(execPath string) *Launcher {
	return &Launcher{
		buf:      make([]byte, 0, 512),
		execPath: execPath,
	}
}

// CreateLauncher creates launcher file.
func (l *Launcher) CreateLauncher() error {
	if err := os.MkdirAll(launcherDir, dirPerm); err != nil && !os.IsExist(err) {
		return fmt.Errorf("could not mkdir %q: %v", launcherDir, err)
	}
	b := fmt.Appendf(l.buf[:0], "[Desktop Entry]\nName=%s\nType=Application\nExec=%s\nIcon=%s\n", name, l.execPath, iconPath)
	if err := os.WriteFile(launcherPath, b, perm); err != nil {
		return fmt.Errorf("could not create launcher file %q: %v", launcherPath, err)
	}
	if err := os.MkdirAll(iconDir, dirPerm); err != nil && !os.IsExist(err) {
		return fmt.Errorf("could not mkdir %q: %v", iconDir, err)
	}
	if err := os.WriteFile(iconPath, icon.Connected, perm); err != nil {
		return fmt.Errorf("could not create launcher icon file %q: %v", launcherPath, err)
	}
	return nil
}

func (l *Launcher) RemoveLauncher() error {
	if err := os.Remove(launcherPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove launcher file %q: %v\n", launcherPath, err)
	}
	if err := os.Remove(iconPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove launcher's icon file %q: %v\n", iconPath, err)
	}
	return nil
}

func (l *Launcher) IsLauncherExist() bool {
	_, err := os.Stat(launcherPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Printf("check if exist: %s: %v\n", launcherPath, err)
		}
		return false
	}
	return true
}

// CreateAutostartAtLogin creates autostart file.
func (l *Launcher) CreateAutostartAtLogin() error {
	if err := os.MkdirAll(autostartDir, dirPerm); err != nil && !os.IsExist(err) {
		return fmt.Errorf("could not mkdir %q: %v", autostartDir, err)
	}
	b := fmt.Appendf(l.buf[:0], "[Desktop Entry]\nName=%s\nType=Application\nExec=%s\n", name, l.execPath)
	if err := os.WriteFile(autostartPath, b, perm); err != nil {
		return fmt.Errorf("could not create autostart file %q: %v", autostartPath, err)
	}
	return nil
}

func (l *Launcher) RemoveAutostartAtLogin() error {
	if err := os.Remove(autostartPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove autostart file %q: %v\n", autostartPath, err)
	}
	return nil
}

func (l *Launcher) IsAutostartExist() bool {
	_, err := os.Stat(autostartPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Printf("check if exist: %s: %v\n", autostartPath, err)
		}
		return false
	}
	return true
}

// Cleanup completely removes all created files.
func Cleanup() {
	msgs := [...]string{
		fmt.Sprintf("Launcher file is removed: %q", launcherPath),
		fmt.Sprintf("Launcher's icon file is removed: %q", iconPath),
		fmt.Sprintf("Autostart file is removed: %q", autostartPath),
	}

	if err := os.Remove(launcherPath); err != nil && !os.IsNotExist(err) {
		msgs[0] = fmt.Sprintf("could not remove launcher file %q: %v", launcherPath, err)
	}
	if err := os.Remove(iconPath); err != nil && !os.IsNotExist(err) {
		msgs[1] = fmt.Sprintf("could not remove launcher's icon file %q: %v", iconPath, err)
	}
	if err := os.Remove(autostartPath); err != nil && !os.IsNotExist(err) {
		msgs[2] = fmt.Sprintf("could not remove autostart file %q: %v", autostartPath, err)
	}

	for i := range msgs {
		fmt.Println(msgs[i])
	}
}
