# Cloudflare WARP Applet

The missing Applet for Cloudflare WARP on Linux for easy connect or disconnect WARP by mouse-clicking on the panel. Sometimes, we want to take a break from typing on the terminal, don't we?

This acts as a frontend for [warp-cli](https://developers.cloudflare.com/warp-client/get-started/linux/).

![demo](./demo.gif)

## Prerequisites

1. Install Cloudflare's [warp-cli](https://developers.cloudflare.com/warp-client/get-started/linux/)

## Install

If you have [Go](https://go.dev/) installed and want to build from source:

```sh
go install github.com/muktihari/warp-applet@latest
```

Or, you can just download the **warp-applet** binary from [Release Assets](https://github.com/muktihari/warp-applet/releases), don't forget to make it executable:

```sh
chmod +x warp-applet
```

## Run

```sh
warp-applet
```

## Create Launcher
Create launcher so you can open it via Applications Launcher (e.g. Show Apps on GNOME).

After the applet is running, check **Settings -> Show in Apps**.

> When this menu is checked, this file *~/.local/share/applications/warp-applet.desktop* will be created. When unchecked, the file will be deleted.

## Autostart at Login:
Enable auto start so the applet is ready right after you login.

After the applet is running, check  **Settings -> Autostart at Login**

> When this menu is checked, this file *~/.config/autostart/warp-applet.desktop* will be created. When unchecked, the file will be deleted.

## Troubleshooting

- If this applet doesn't show on GNOME's panel, please make sure that AppIndicator is installed/enabled. Installing this gnome-extension works for me on Fedora: https://extensions.gnome.org/extension/615/appindicator-support.

  If this applet doesn't show only on autostart while the lock file is exist, it's probably because the applet is run before the AppIndicator is ready, you can try adding delay on autostart file manually e.g.

  `Exec=bash -c "sleep 2 && /<absolute path to>/warp-applet"`

- If this applet is unexpectedly crashed (e.g. force killed) and is in broken state (e.g. can't run the applet due to dangling lock file), please run this following command:

  ```sh
  rm /tmp/warp-applet.lock
  ```

- If you want to completely remove all files created by the applet such as launcher, autostart and lock file, you can run:

  ```sh
  warp-applet -cleanup
  ```
