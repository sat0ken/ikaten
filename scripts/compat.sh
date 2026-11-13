#!/bin/sh
# Build the exact bounce game source with real Ebitengine. Only the API import
# adapter differs. Dependencies stay outside the firmware's main module.
set -eu
cd "$(dirname "$0")/.."
mkdir -p build/compat/bounce
cp games/bounce/game.go build/compat/bounce/game.go
cp compat/import_ebitengine.go.txt build/compat/bounce/import_ebitengine.go
cp compat/main.go.txt build/compat/main.go
cp compat/go.mod.txt build/compat/go.mod
cd build/compat
GOOS=js GOARCH=wasm go build -mod=mod -o ../bounce-ebitengine.wasm .
cmp bounce/game.go ../../games/bounce/game.go
# snake uses ebiten.X directly, so only its import paths are rewritten.
mkdir -p snake snakemain
sed -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
 -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' ../../games/snake/game.go > snake/game.go
cp ../../compat/snake_main.go.txt snakemain/main.go
GOOS=js GOARCH=wasm go build -mod=mod -o ../snake-ebitengine.wasm ./snakemain
# flappy also swaps its image loading and text adapter.
mkdir -p flappy/assets flappymain
sed -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
 -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' ../../games/flappy/game.go > flappy/game.go
cp ../../compat/flappy_platform.go.txt flappy/platform_ebitengine.go
cp ../../games/flappy/assets/*.png flappy/assets/
cp ../../compat/flappy_main.go.txt flappymain/main.go
GOOS=js GOARCH=wasm go build -mod=mod -o ../flappy-ebitengine.wasm ./flappymain
# blocks has several files; all but the platform adapter are shared.
rm -rf blocks blocksmain
mkdir -p blocks/assets blocksmain
for f in ../../games/blocks/*.go; do
 case "$f" in *platform_ikaten.go|*assets.go) continue;; esac
 sed -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
  -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' "$f" > "blocks/$(basename "$f")"
done
cp ../../compat/blocks_platform.go.txt blocks/platform_ebitengine.go
cp ../../games/blocks/assets/*.png blocks/assets/
cp ../../compat/blocks_main.go.txt blocksmain/main.go
GOOS=js GOARCH=wasm go build -mod=mod -o ../blocks-ebitengine.wasm ./blocksmain
# dungeon is written against the Ebitengine API only.
mkdir -p dungeon dungeonmain
sed -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
 -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' ../../games/dungeon/game.go > dungeon/game.go
cp ../../compat/dungeon_main.go.txt dungeonmain/main.go
GOOS=js GOARCH=wasm go build -mod=mod -o ../dungeon-ebitengine.wasm ./dungeonmain
# shooter is written against the Ebitengine API only.
mkdir -p shooter shootermain
for f in ../../games/shooter/*.go; do
 sed -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
  -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' "$f" > "shooter/$(basename "$f")"
done
cp ../../compat/shooter_main.go.txt shootermain/main.go
GOOS=js GOARCH=wasm go build -mod=mod -o ../shooter-ebitengine.wasm ./shootermain
# kart also has a course package and swaps its track loading.
rm -rf kart kartmain
mkdir -p kart/course kart/assets kartmain
for f in ../../games/kart/*.go ../../games/kart/course/*.go; do
 case "$f" in *platform_ikaten.go|*assets.go|*_test.go) continue;; esac
 out="kart/$(basename "$f")"
 case "$f" in */course/*) out="kart/course/$(basename "$f")";; esac
 sed -e 's#"github.com/sat0ken/ikaten/games/kart/course"#"ikaten-compat/kart/course"#' \
  -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
  -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' "$f" > "$out"
done
cp ../../compat/kart_platform.go.txt kart/platform_ebitengine.go
cp ../../games/kart/assets/track.png kart/assets/
cp ../../compat/kart_main.go.txt kartmain/main.go
GOOS=js GOARCH=wasm go build -mod=mod -o ../kart-ebitengine.wasm ./kartmain
# tank also has a field package and swaps its asset loading.
rm -rf tank tankmain
mkdir -p tank/field tank/assets tankmain
for f in ../../games/tank/*.go ../../games/tank/field/*.go; do
 case "$f" in *platform_ikaten.go|*assets.go|*_test.go) continue;; esac
 out="tank/$(basename "$f")"
 case "$f" in */field/*) out="tank/field/$(basename "$f")";; esac
 sed -e 's#"github.com/sat0ken/ikaten/games/tank/field"#"ikaten-compat/tank/field"#' \
  -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
  -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' "$f" > "$out"
done
cp ../../compat/tank_platform.go.txt tank/platform_ebitengine.go
cp ../../games/tank/assets/*.png tank/assets/
cp ../../compat/tank_main.go.txt tankmain/main.go
GOOS=js GOARCH=wasm go build -mod=mod -o ../tank-ebitengine.wasm ./tankmain
# uboat also swaps its asset loading.
rm -rf uboat uboatmain
mkdir -p uboat/assets uboatmain
for f in ../../games/uboat/*.go; do
 case "$f" in *platform_ikaten.go|*assets.go|*_test.go) continue;; esac
 sed -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
  -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' "$f" > "uboat/$(basename "$f")"
done
cp ../../compat/uboat_platform.go.txt uboat/platform_ebitengine.go
cp ../../games/uboat/assets/*.png uboat/assets/
cp ../../compat/uboat_main.go.txt uboatmain/main.go
GOOS=js GOARCH=wasm go build -mod=mod -o ../uboat-ebitengine.wasm ./uboatmain
# train is written against the Ebitengine API only and draws its images in
# code. Its main is written here.
rm -rf train trainmain
mkdir -p train trainmain
for f in ../../games/train/*.go; do
 case "$f" in *_test.go) continue;; esac
 sed -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
  -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' "$f" > "train/$(basename "$f")"
done
cat > trainmain/main.go <<'EOF'
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"ikaten-compat/train"
)

func main() {
	ebiten.SetWindowSize(480, 480)
	ebiten.SetWindowTitle("ikaten train")
	if err := ebiten.RunGame(train.NewGame()); err != nil {
		log.Fatal(err)
	}
}
EOF
GOOS=js GOARCH=wasm go build -mod=mod -o ../train-ebitengine.wasm ./trainmain
# air, like train, draws its images in code. Its main is written here.
rm -rf air airmain
mkdir -p air airmain
for f in ../../games/air/*.go; do
 case "$f" in *_test.go) continue;; esac
 sed -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
  -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' "$f" > "air/$(basename "$f")"
done
cat > airmain/main.go <<'EOF'
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"ikaten-compat/air"
)

func main() {
	ebiten.SetWindowSize(480, 480)
	ebiten.SetWindowTitle("ikaten air")
	if err := ebiten.RunGame(air.NewGame()); err != nil {
		log.Fatal(err)
	}
}
EOF
GOOS=js GOARCH=wasm go build -mod=mod -o ../air-ebitengine.wasm ./airmain
# fish, like train and air, draws its images in code. Its main is written here.
rm -rf fish fishmain
mkdir -p fish fishmain
for f in ../../games/fish/*.go; do
 case "$f" in *_test.go) continue;; esac
 sed -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
  -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' "$f" > "fish/$(basename "$f")"
done
cat > fishmain/main.go <<'EOF'
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"ikaten-compat/fish"
)

func main() {
	ebiten.SetWindowSize(480, 480)
	ebiten.SetWindowTitle("ikaten fish")
	if err := ebiten.RunGame(fish.NewGame()); err != nil {
		log.Fatal(err)
	}
}
EOF
GOOS=js GOARCH=wasm go build -mod=mod -o ../fish-ebitengine.wasm ./fishmain
# bomb, like train, air and fish, draws its images in code. Its main is
# written here.
rm -rf bomb bombmain
mkdir -p bomb bombmain
for f in ../../games/bomb/*.go; do
 case "$f" in *_test.go) continue;; esac
 sed -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
  -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' "$f" > "bomb/$(basename "$f")"
done
cat > bombmain/main.go <<'EOF'
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"ikaten-compat/bomb"
)

func main() {
	ebiten.SetWindowSize(480, 480)
	ebiten.SetWindowTitle("ikaten gopher bomb")
	if err := ebiten.RunGame(bomb.NewGame()); err != nil {
		log.Fatal(err)
	}
}
EOF
GOOS=js GOARCH=wasm go build -mod=mod -o ../bomb-ebitengine.wasm ./bombmain
# jump, like bomb, draws its images in code. Its main is written here.
rm -rf jump jumpmain
mkdir -p jump jumpmain
for f in ../../games/jump/*.go; do
 case "$f" in *_test.go) continue;; esac
 sed -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
  -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' "$f" > "jump/$(basename "$f")"
done
cat > jumpmain/main.go <<'EOF'
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"ikaten-compat/jump"
)

func main() {
	ebiten.SetWindowSize(480, 480)
	ebiten.SetWindowTitle("ikaten gopher jump")
	if err := ebiten.RunGame(jump.NewGame()); err != nil {
		log.Fatal(err)
	}
}
EOF
GOOS=js GOARCH=wasm go build -mod=mod -o ../jump-ebitengine.wasm ./jumpmain
# all imports the games above from this module.
rm -rf all allmain
mkdir -p all allmain
sed -e 's#"github.com/sat0ken/ikaten/games/\([a-z]*\)"#"ikaten-compat/\1"#' \
 -e 's#ebiten "github.com/sat0ken/ikaten"#"github.com/hajimehoshi/ebiten/v2"#' \
 -e 's#"github.com/sat0ken/ikaten/#"github.com/hajimehoshi/ebiten/v2/#' ../../games/all/all.go > all/all.go
cp ../../compat/all_main.go.txt allmain/main.go
GOOS=js GOARCH=wasm go build -mod=mod -o ../all-ebitengine.wasm ./allmain
echo "Ebitengine compatibility build passed (bounce: shared game.go; snake, flappy, blocks, dungeon, shooter, kart, uboat, tank, train, air, fish, bomb, jump and all: import paths rewritten)"
