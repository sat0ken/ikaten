#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
TINYGO=${TINYGO:-tinygo}
mkdir -p build/smoke
for target in badge esp32s3 esp32 pico gopher-badge gopher-board; do
 for example in bounce assets snake flappy blocks dungeon shooter kart uboat tank train air fish all probe benchmark; do
  echo "Building $example for $target"
  tags=
  if [ "$example" = benchmark ]; then tags=ikaten_benchmark; fi
  "$TINYGO" build -target "./targets/$target.json" -tags "$tags" -size short \
   -o "build/smoke/$example-$target.elf" "./examples/$example"
 done
done
for example in bounce assets snake flappy blocks dungeon shooter kart uboat tank train air fish all benchmark; do
 echo "Building $example for wasm"
 tags=
 if [ "$example" = benchmark ]; then tags=ikaten_benchmark; fi
 "$TINYGO" build -target wasm -tags "$tags" -o "build/smoke/$example.wasm" "./examples/$example"
done
echo "smoketest passed"
