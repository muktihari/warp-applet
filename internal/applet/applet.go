package applet

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
	"unicode"

	"github.com/getlantern/systray"
)

const bin = "warp-cli"

type status string

const (
	statusUnknown       status = "Unknown"
	statusConnecting    status = "Connecting"
	statusConnected     status = "Connected"
	statusDisconnected  status = "Disconnected"
	statusDisconnecting status = "Disconnecting"
	statusNoNetwork     status = "No Network" // Status update: Unable\nReason: No Network
)

type mode string

const (
	modeUnknown              mode = "Unknown"
	modeDnsOverHttps         mode = "DnsOverHttps"
	modeWarp                 mode = "Warp"
	modeWarpWithDnsOverHttps mode = "WarpWithDnsOverHttps"
	modeDnsOverTls           mode = "DnsOverTls"
	modeWarpWithDnsOverTls   mode = "WarpWithDnsOverTls"
	modeWarpProxy            mode = "WarpProxy"
	modeTunnelOnly           mode = "TunnelOnly"
)

func (m mode) cmdArg() string {
	switch m {
	case modeDnsOverHttps:
		return "doh"
	case modeWarp:
		return "warp"
	case modeWarpWithDnsOverHttps:
		return "warp+doh"
	case modeDnsOverTls:
		return "dot"
	case modeWarpWithDnsOverTls:
		return "warp+dot"
	case modeWarpProxy:
		return "proxy"
	case modeTunnelOnly:
		return "tunnel_only"
	}
	return "unknown"
}

// New creates new applet.
func New(lockFilePath string, connectedIcon, disconnectedIcon, unknownStateIcon []byte) *Applet {
	a := new(Applet)
	a.lock.path = lockFilePath
	a.icon.connected = connectedIcon
	a.icon.disconnected = disconnectedIcon
	a.icon.unknownState = unknownStateIcon
	return a
}

// Applet is a fontend for warp-cli (Cloudflare WARP).
type Applet struct {
	buf  bytes.Buffer
	lock struct { // Enable singleton app
		path string
		file *os.File
	}

	icon struct {
		connected    []byte
		disconnected []byte
		unknownState []byte
	}

	menu struct {
		status     *systray.MenuItem
		connect    *systray.MenuItem
		disconnect *systray.MenuItem
		mode       modeMenu
		refresh    *systray.MenuItem
		quit       *systray.MenuItem
	}
}

// Launch launchs the applet. Only one instance of applet can run (singleton) as a guard for users from
// accidentally opening multiple instances. However, if it fails to create a lock file, we'll still allow it
// to run so we don't prevent users from running the applet at all.
func (a *Applet) Launch() {
	var err error
	if a.lock.file, err = os.OpenFile(a.lock.path, os.O_CREATE|os.O_EXCL, 0o644); err != nil {
		if os.IsExist(err) {
			fmt.Printf("%s is already running, lock file exists: %q\n", bin, a.lock.path)
			return
		}
		fmt.Printf("%s is running but without lock, creating lock file failed: %v\n", bin, err)
	}
	systray.Run(a.onReady, a.onExit)
}

// Quit gracefully quits the applet.
func (a *Applet) Quit() { systray.Quit() }

func (a *Applet) onReady() {
	systray.SetTooltip("Cloudflare WARP Applet")
	systray.SetIcon(a.icon.disconnected)

	a.menu.status = systray.AddMenuItem("Initializing...", "")

	systray.AddSeparator()

	a.menu.connect = systray.AddMenuItem("Connect", "")
	a.menu.disconnect = systray.AddMenuItem("Disconnect", "")

	a.menu.mode.MenuItem = systray.AddMenuItem("Mode", "")
	a.menu.mode.doh = a.menu.mode.AddSubMenuItemCheckbox("  DNS over HTTPS (DoH)", "", false)
	a.menu.mode.warp = a.menu.mode.AddSubMenuItemCheckbox("  WARP", "", false)
	a.menu.mode.warpPlusDoh = a.menu.mode.AddSubMenuItemCheckbox("  WARP + DoH", "", false)
	a.menu.mode.dot = a.menu.mode.AddSubMenuItemCheckbox("  DNS over TLS (DoT)", "", false)
	a.menu.mode.warpPlusDot = a.menu.mode.AddSubMenuItemCheckbox("  WARP + DoT", "", false)
	a.menu.mode.proxy = a.menu.mode.AddSubMenuItemCheckbox("  Proxy", "", false)
	a.menu.mode.tunnelOnly = a.menu.mode.AddSubMenuItemCheckbox("  Tunnel Only", "", false)

	a.menu.refresh = systray.AddMenuItem("Refresh", "")
	a.menu.quit = systray.AddMenuItem("Quit", "")

	a.menu.status.Disable()
	a.menu.disconnect.Hide()

	go func() {
		for {
			select {
			case <-a.menu.connect.ClickedCh:
				a.connect()
			case <-a.menu.disconnect.ClickedCh:
				a.disconnect()
			case <-a.menu.refresh.ClickedCh:
				a.refresh()
			case <-a.menu.mode.doh.ClickedCh:
				a.changeMode(modeDnsOverHttps)
			case <-a.menu.mode.warp.ClickedCh:
				a.changeMode(modeWarp)
			case <-a.menu.mode.warpPlusDoh.ClickedCh:
				a.changeMode(modeWarpWithDnsOverHttps)
			case <-a.menu.mode.dot.ClickedCh:
				a.changeMode(modeDnsOverTls)
			case <-a.menu.mode.warpPlusDot.ClickedCh:
				a.changeMode(modeWarpWithDnsOverTls)
			case <-a.menu.mode.proxy.ClickedCh:
				a.changeMode(modeWarpProxy)
			case <-a.menu.mode.tunnelOnly.ClickedCh:
				a.changeMode(modeTunnelOnly)
			}
		}
	}()

	go func() {
		<-a.menu.quit.ClickedCh
		systray.Quit()
	}()

	a.updateModeFromSettings()
	// On startup, the computer may take a while to connect to Wi-Fi.
loop:
	for {
		switch a.updateStatus() {
		case statusConnecting, statusDisconnecting, statusNoNetwork:
			time.Sleep(time.Second)
		default:
			break loop
		}
	}
}

func (a *Applet) onExit() {
	a.disconnect()
	if a.lock.file == nil {
		return
	}
	a.lock.file.Close()
	a.lock.file = nil
	if err := os.Remove(a.lock.path); err != nil && !os.IsNotExist(err) {
		fmt.Printf("could not remove lock file %q: %v\n", a.lock.path, err)
		return
	}
}

func (a *Applet) connect() {
	a.buf.Reset()
	cmd := exec.Command(bin, "connect")
	cmd.Stderr = &a.buf
	if err := cmd.Run(); err != nil {
		a.showError(a.buf.Bytes(), err)
		return
	}
	a.mustUpdateStatus()
}

func (a *Applet) disconnect() {
	a.buf.Reset()
	cmd := exec.Command(bin, "disconnect")
	cmd.Stderr = &a.buf
	if err := cmd.Run(); err != nil {
		a.showError(a.buf.Bytes(), err)
		return
	}
	a.mustUpdateStatus()
}

func (a *Applet) changeMode(mode mode) {
	defer a.menu.mode.check(mode)
	if mode == a.menu.mode.checked() {
		return
	}
	a.buf.Reset()
	cmd := exec.Command(bin, "mode", mode.cmdArg())
	cmd.Stderr = &a.buf
	if err := cmd.Run(); err != nil {
		a.showError(a.buf.Bytes(), err)
		return
	}
}

func (a *Applet) refresh() {
	a.updateModeFromSettings()
	a.mustUpdateStatus()
}

// mustUpdateStatusExpect checks status on loop with interval. It only returns
// when it gets status statusConnected, statusDisconnected or statusUnknown.
func (a *Applet) mustUpdateStatus() {
	const add = 100 * time.Millisecond
	const max = 2 * time.Second
	dur := 200 * time.Millisecond
	time.Sleep(dur) // The update is not immediate, let's wait for a moment first.
	for {
		switch a.updateStatus() {
		case statusConnected, statusDisconnected, statusUnknown:
			return
		}
		if dur < max {
			dur += add
		}
		time.Sleep(dur)
		continue
	}
}

func (a *Applet) updateStatus() status {
	a.buf.Reset()
	cmd := exec.Command(bin, "status")
	cmd.Stderr = &a.buf
	cmd.Stdout = &a.buf
	if err := cmd.Run(); err != nil {
		a.showError(a.buf.Bytes(), err)
		return statusUnknown
	}

	output := a.buf.Bytes()
	switch {
	case bytes.Contains(output, []byte(statusConnecting)):
		a.menu.status.SetTitle("Connecting...")
		return statusConnecting
	case bytes.Contains(output, []byte(statusDisconnecting)):
		a.menu.status.SetTitle("Disconnecting...")
		return statusDisconnecting
	case bytes.Contains(output, []byte(statusConnected)):
		a.menu.status.SetTitle("Connected!")
		a.menu.connect.Hide()
		a.menu.disconnect.Show()
		systray.SetIcon(a.icon.connected)
		return statusConnected
	case bytes.Contains(output, []byte(statusDisconnected)):
		a.menu.status.SetTitle("Disconnected!")
		a.menu.disconnect.Hide()
		a.menu.connect.Show()
		systray.SetIcon(a.icon.disconnected)
		return statusDisconnected
	case bytes.Contains(output, []byte(statusNoNetwork)):
		a.menu.status.SetTitle("No Network")
		a.menu.disconnect.Hide()
		a.menu.connect.Show()
		systray.SetIcon(a.icon.disconnected)
		return statusNoNetwork
	default:
		systray.SetIcon(a.icon.unknownState)
		a.showError(output, fmt.Errorf("Status: %s", statusUnknown))
		return statusUnknown
	}
}

func (a *Applet) updateModeFromSettings() {
	a.buf.Reset()
	cmd := exec.Command(bin, "settings")
	cmd.Stdout = &a.buf
	cmd.Stderr = &a.buf
	if err := cmd.Run(); err != nil {
		a.showError(a.buf.Bytes(), err)
		return
	}

	const sep = "Mode:" // e.g. Mode: WarpProxy on port 40000
	for {
		b, err := a.buf.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		i := bytes.Index(b, []byte(sep))
		if i == -1 {
			continue
		}
		b = bytes.TrimSpace(b[i+len(sep):])
		for i, v := range string(b) {
			if unicode.IsSpace(v) {
				b = b[:i]
				break
			}
		}
		a.menu.mode.check(mode(b)) // e.g. WarpProxy
		break
	}
}

func (a *Applet) showError(msg []byte, err error) {
	s := fmt.Sprintf("ERROR!\n%s\n%v", msg, err)
	a.menu.status.SetTitle(s)
}

type modeMenu struct {
	*systray.MenuItem
	doh         *systray.MenuItem
	warp        *systray.MenuItem
	warpPlusDoh *systray.MenuItem
	dot         *systray.MenuItem
	warpPlusDot *systray.MenuItem
	proxy       *systray.MenuItem
	tunnelOnly  *systray.MenuItem
}

func (m *modeMenu) checked() mode {
	switch {
	case m.doh.Checked():
		return modeDnsOverHttps
	case m.warp.Checked():
		return modeWarp
	case m.warpPlusDoh.Checked():
		return modeWarpWithDnsOverHttps
	case m.dot.Checked():
		return modeDnsOverTls
	case m.warpPlusDot.Checked():
		return modeWarpWithDnsOverTls
	case m.proxy.Checked():
		return modeWarpProxy
	case m.tunnelOnly.Checked():
		return modeTunnelOnly
	}
	return modeUnknown
}

func (m *modeMenu) check(mode mode) {
	m.doh.Uncheck()
	m.warp.Uncheck()
	m.warpPlusDoh.Uncheck()
	m.dot.Uncheck()
	m.warpPlusDot.Uncheck()
	m.proxy.Uncheck()
	m.tunnelOnly.Uncheck()

	switch mode {
	case modeDnsOverHttps:
		m.doh.Check()
	case modeWarp:
		m.warp.Check()
	case modeWarpWithDnsOverHttps:
		m.warpPlusDoh.Check()
	case modeDnsOverTls:
		m.dot.Check()
	case modeWarpWithDnsOverTls:
		m.warpPlusDot.Check()
	case modeWarpProxy:
		m.proxy.Check()
	case modeTunnelOnly:
		m.tunnelOnly.Check()
	}
}
