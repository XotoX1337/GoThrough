.PHONY: build run run-config fresh

build:
	wails build -s

# Launch the picker (default double-click behavior).
run: build
	./build/bin/GoThrough.exe

# Launch day1 ignoring saved progress (fresh start).
fresh: build
	./build/bin/GoThrough.exe run --fresh configstore/configs/poe/act1.yaml

# Launch a specific config directly via CLI (dev shortcut).
run-config: build
	./build/bin/GoThrough.exe run configstore/configs/gothic2/day1.yaml
