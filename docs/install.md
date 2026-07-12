# Install

Every release ships builds for macOS, Windows and Linux on the [GitHub releases page](https://github.com/TRC-Loop/Pelton/releases). Pick your platform below.

=== "macOS"

    Download the `.dmg` for your Mac (Apple Silicon or Intel) from the latest release, open it and drag Pelton into Applications.

    !!! note "Unsigned app"

        Pelton builds are not notarized by Apple, so macOS blocks the first launch. Getting past that takes one round trip through System Settings:

        1. Open Pelton. A dialog says the app could not be verified; press **Done** (not "Move to Bin").
        2. Open **System Settings, Privacy & Security** and scroll down to where it says Pelton was blocked.
        3. Press **Open Anyway**, then **Open** in the confirmation, and authenticate with Touch ID or your password.

        That is needed once; afterwards Pelton opens normally. The terminal shortcut `xattr -cr /Applications/Pelton.app` achieves the same in one step.

=== "Windows"

    Download `Pelton-<version>-windows-amd64-installer.exe` from the latest release and run it. SmartScreen may warn about an unknown publisher for the same reason as on macOS: the builds are unsigned. Choose **More info** and then **Run anyway**.

=== "Fedora (dnf)"

    Pelton is packaged in [Fedora Copr](https://copr.fedorainfracloud.org/coprs/arnek/Pelton/), so installs and updates go through `dnf` like any other package:

    ```sh
    sudo dnf copr enable arnek/Pelton
    sudo dnf install pelton
    ```

    Updates then arrive with your normal `sudo dnf update`.

=== "Linux (rpm)"

    If you would rather not enable the Copr repo, each release also carries the raw package:

    ```sh
    sudo dnf install ./Pelton-<version>-linux-fedora-x86_64.rpm
    ```

    You will not get automatic updates this way; download the next release yourself or enable the update check in Settings, which compares versions against the GitHub releases API and nothing else.

## Build from source

You need Go, Node with pnpm, and the Wails CLI matching the version pinned in `go.mod`:

```sh
go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0
git clone https://github.com/TRC-Loop/Pelton
cd Pelton
wails build
```

On Linux, install the WebKitGTK toolchain first and build with the matching tag:

```sh
sudo dnf install gtk3-devel webkit2gtk4.1-devel
wails build -platform linux/amd64 -tags webkit2_41
```

The binary lands in `build/bin/`. For a development loop with hot reload, use `make run`; it runs against a separate dev data directory, so it never touches a real install's accounts or settings.
