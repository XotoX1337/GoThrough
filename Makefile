.PHONY: build run run-config

build:
	wails build -s

# Launch the picker (default double-click behavior).
run: build
	./build/bin/GoThrough.exe

# Launch a specific config directly via CLI (dev shortcut).
run-config: build
	./build/bin/GoThrough.exe run configstore/configs/gothic2/chapter1.yaml
