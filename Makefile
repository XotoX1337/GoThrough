.PHONY: build run

build:
	wails build -s

run: build
	./build/bin/GoThrough.exe run configs/gothic2/chapter1.yaml
