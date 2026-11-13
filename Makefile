GO ?= go
TINYGO ?= tinygo
TARGET ?= badge
GAME ?= bounce
export GOFLAGS ?= -buildvcs=false

.PHONY: test check build wasm serve smoketest benchmark probe drawprobe allocs compat clean

test:
	$(GO) test ./...

check: test
	$(GO) vet ./...
	$(GO) test -tags ikaten_benchmark ./...

build:
	mkdir -p build
	$(TINYGO) build -target ./targets/$(TARGET).json -size short -o build/$(GAME)-$(TARGET).elf ./examples/$(GAME)

wasm:
	$(TINYGO) build -target wasm -o static/game.wasm ./examples/$(GAME)
	cp "$$($(TINYGO) env TINYGOROOT)/targets/wasm_exec.js" static/wasm_exec.js

serve: wasm
	python3 -m http.server 8080 --directory static --bind 127.0.0.1

smoketest:
	TINYGO=$(TINYGO) sh scripts/smoketest.sh

probe:
	mkdir -p build
	$(TINYGO) build -target ./targets/$(TARGET).json -size short -o build/probe-$(TARGET).elf ./examples/probe

drawprobe:
	mkdir -p build
	$(TINYGO) build -target ./targets/$(TARGET).json -size short -o build/drawprobe-$(TARGET).elf ./examples/drawprobe

benchmark:
	mkdir -p build
	$(TINYGO) build -target ./targets/$(TARGET).json -tags ikaten_benchmark -size short -o build/benchmark-$(TARGET).elf ./examples/benchmark

allocs:
	mkdir -p build
	$(TINYGO) build -target ./targets/$(TARGET).json -print-allocs='github.com/sat0ken/ikaten.*' -o build/allocs-$(TARGET).elf ./examples/bounce

compat:
	sh scripts/compat.sh

clean:
	rm -rf build static/game.wasm static/wasm_exec.js
