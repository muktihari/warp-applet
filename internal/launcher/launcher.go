package launcher

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

const name = "Cloudflare WARP Applet"
const perm = 0o644

var (
	home          = os.Getenv("HOME")
	iconPath      = filepath.Join(home, ".local", "share", "icons", "warp-applet.png")
	launcherPath  = filepath.Join(home, ".local", "share", "applications", "warp-applet.desktop")
	autostartPath = filepath.Join(home, ".config", "autostart", "warp-applet.desktop")
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

	if err := os.WriteFile(launcherPath, buf.Bytes(), perm); err != nil {
		return fmt.Errorf("could not create launcher file %q: %v", launcherPath, err)
	}
	fmt.Printf("Launcher file is created: %q\n", launcherPath)

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

	if err := os.WriteFile(autostartPath, buf.Bytes(), perm); err != nil {
		return fmt.Errorf("could not create autostart file %q: %v", autostartPath, err)
	}
	fmt.Printf("Autostart file is created: %q\n", autostartPath)

	return nil
}

// Cleanup completely removes all created files.
func Cleanup(lockFilePath string) {
	const n = 4
	msgs := [n]string{
		fmt.Sprintf("Launcher file is removed: %q\n", launcherPath),
		fmt.Sprintf("Launcher's icon file is removed: %q\n", iconPath),
		fmt.Sprintf("Autostart file is removed: %q\n", autostartPath),
		fmt.Sprintf("Lock file is removed: %q\n", lockFilePath),
	}

	var errs [n]error
	if err := os.Remove(launcherPath); err != nil && !os.IsNotExist(err) {
		errs[0] = fmt.Errorf("could not remove launcher file %q: %v\n", launcherPath, err)
	}
	if err := os.Remove(iconPath); err != nil && !os.IsNotExist(err) {
		errs[1] = fmt.Errorf("could not remove launcher's icon file %q: %v\n", iconPath, err)
	}
	if err := os.Remove(autostartPath); err != nil && !os.IsNotExist(err) {
		errs[2] = fmt.Errorf("could not remove autostart file %q: %v\n", autostartPath, err)
	}
	if err := os.Remove(lockFilePath); err != nil && !os.IsNotExist(err) {
		errs[3] = fmt.Errorf("could not remove lock file %q: %v\n", lockFilePath, err)
	}

	for i := range errs {
		if errs[i] == nil {
			fmt.Print(msgs[i])
		} else {
			fmt.Print(errs[i])
		}
	}
}
