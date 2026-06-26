# Pelton - email client (Wails + Svelte)

.PHONY: build run dev clean tidy deps licenses

# production build into build/bin
build:
	wails build

# run in dev mode with hot reload
run:
	wails dev

dev: run

# sync go + npm dependencies
deps:
	go mod download
	cd frontend && npm install

tidy:
	go mod tidy

# collect third-party license texts into licenses/
licenses:
	bash scripts/collect-licenses.sh

clean:
	go clean
	wails build -clean || true
