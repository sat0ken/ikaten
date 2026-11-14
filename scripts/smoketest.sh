#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
TINYGO=${TINYGO:-tinygo}
mkdir -p build/smoke
# These are drawn for 240 by 240 only. ESP32 (without S3) has too little
# memory for them, so they are not built for it.
only240="dungeon shooter kart uboat tank train air fish bomb jump all"
for target in badge esp32s3 esp32 pico gopher-badge gopher-board; do
 for example in bounce assets snake flappy blocks dungeon shooter kart uboat tank train air fish bomb jump all probe drawprobe benchmark; do
  if [ "$target" = esp32 ]; then
   case " $only240 " in *" $example "*) echo "Skipping $example for $target (240x240 only)"; continue;; esac
  fi
  # The launcher holds every game's images, about 2.7 MB, more than these
  # boards' 2 MB of flash. Each game still builds for them on its own.
  if [ "$example" = all ] && { [ "$target" = pico ] || [ "$target" = gopher-board ]; }; then
   echo "Skipping $example for $target (2 MB flash)"
   continue
  fi
  echo "Building $example for $target"
  tags=
  if [ "$example" = benchmark ]; then tags=ikaten_benchmark; fi
  "$TINYGO" build -target "./targets/$target.json" -tags "$tags" -size short \
   -o "build/smoke/$example-$target.elf" "./examples/$example"
 done
done
for example in bounce assets snake flappy blocks dungeon shooter kart uboat tank train air fish bomb jump all benchmark; do
 echo "Building $example for wasm"
 tags=
 if [ "$example" = benchmark ]; then tags=ikaten_benchmark; fi
 "$TINYGO" build -target wasm -tags "$tags" -o "build/smoke/$example.wasm" "./examples/$example"
done
echo "smoketest passed"
