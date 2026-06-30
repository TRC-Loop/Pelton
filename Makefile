# Pelton - email client (Wails + Svelte)

.PHONY: build build-mac build-win build-linux run app-dev dev clean tidy deps licenses icon

# version string injected into the binary. it prefers the latest git tag (with a
# short commit suffix on untagged commits) and falls back to "dev". it is wired
# into main.version via ldflags and shown in the about section.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

# production build into build/bin
build:
	wails build -ldflags "$(LDFLAGS)"

# compile the macOS Liquid Glass app icon (build/darwin/pelton.icon) into the
# built .app via actool. needs a one-time `sudo xcodebuild -runFirstLaunch`.
icon:
	scripts/build-icon.sh

# macOS build that also installs the Liquid Glass icon into the .app.
build-mac: build icon

# windows build (amd64). cross-compiling from another OS needs the appropriate
# toolchain (mingw-w64) and webview2; run on Windows for a no-fuss build.
build-win:
	wails build -platform windows/amd64 -ldflags "$(LDFLAGS)"

# linux build (amd64), then drop the .desktop launcher next to the binary so it
# is easy to install into ~/.local/share/applications (or a package). building
# from macOS needs the gtk/webkit2gtk toolchain; run on Linux for a clean build.
build-linux:
	wails build -platform linux/amd64 -ldflags "$(LDFLAGS)"
	cp build/linux/pelton.desktop build/bin/pelton.desktop
	@echo "linux binary + pelton.desktop in build/bin (install the .desktop and an icon named 'pelton')"

# run the whole app in dev mode: make sure go + npm deps are present, regenerate
# the typescript bindings from the go methods, then launch wails dev with hot
# reload for both the go backend and the svelte frontend.
run: deps
	wails generate module
	wails dev -ldflags "$(LDFLAGS)"

dev: run

# alias kept for discoverability; identical to run.
app-dev: run

# sync go + npm dependencies (frontend uses pnpm)
deps:
	go mod download
	cd frontend && pnpm install

tidy:
	go mod tidy

# build licenses/manifest.json (embedded and shown in the about section).
licenses:
	node scripts/collect-licenses.mjs

clean:
	go clean
	wails build -clean || true
