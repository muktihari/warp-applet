package launcher

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
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

// Create creates launcher file.
func Create(execPath string, iconData []byte) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `[Desktop Entry]
Name=%s
Type=Application
Exec=%s
Icon=%s
`, name, execPath, iconPath)

	if err := os.MkdirAll(launcherDir, dirPerm); err != nil && !os.IsExist(err) {
		return fmt.Errorf("could not mkdir %q: %v", launcherDir, err)
	}
	if err := os.WriteFile(launcherPath, buf.Bytes(), perm); err != nil {
		return fmt.Errorf("could not create launcher file %q: %v", launcherPath, err)
	}
	fmt.Printf("Launcher file is created: %q\n", launcherPath)

	if err := os.MkdirAll(iconDir, dirPerm); err != nil && !os.IsExist(err) {
		return fmt.Errorf("could not mkdir %q: %v", iconDir, err)
	}
	if err := os.WriteFile(iconPath, iconData, perm); err != nil {
		return fmt.Errorf("could not create launcher icon file %q: %v", launcherPath, err)
	}
	fmt.Printf("Launcher's icon file is created: %q\n", iconPath)

	return nil
}

// Autostart creates autostart file.
func Autostart(execPath string) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `[Desktop Entry]
Name=%s
Type=Application
Exec=%s
X-GNOME-Autostart-enabled=true
`, name, execPath)

	if err := os.MkdirAll(autostartDir, dirPerm); err != nil && !os.IsExist(err) {
		return fmt.Errorf("could not mkdir %q: %v", autostartDir, err)
	}
	if err := os.WriteFile(autostartPath, buf.Bytes(), perm); err != nil {
		return fmt.Errorf("could not create autostart file %q: %v", autostartPath, err)
	}
	fmt.Printf("Autostart file is created: %q\n", autostartPath)

	return nil
}

// Cleanup completely removes all created files.
func Cleanup() {
	msgs := [...]string{
		fmt.Sprintf("Launcher file is removed: %q\n", launcherPath),
		fmt.Sprintf("Launcher's icon file is removed: %q\n", iconPath),
		fmt.Sprintf("Autostart file is removed: %q\n", autostartPath),
	}

	if err := os.Remove(launcherPath); err != nil && !os.IsNotExist(err) {
		msgs[0] = fmt.Sprintf("could not remove launcher file %q: %v\n", launcherPath, err)
	}
	if err := os.Remove(iconPath); err != nil && !os.IsNotExist(err) {
		msgs[1] = fmt.Sprintf("could not remove launcher's icon file %q: %v\n", iconPath, err)
	}
	if err := os.Remove(autostartPath); err != nil && !os.IsNotExist(err) {
		msgs[2] = fmt.Sprintf("could not remove autostart file %q: %v\n", autostartPath, err)
	}

	for i := range msgs {
		fmt.Print(msgs[i])
	}
}
