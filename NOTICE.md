# ライセンスと帰属表示

ikaten の大部分は [MIT License](LICENSE) です。ただし、次に挙げるファイルは元の作品のライセンスに従い、MIT License の対象外です。

## Ebitengine のサンプルから移植したコード(Apache License 2.0)

[Ebitengine](https://github.com/hajimehoshi/ebiten) v2.10.5 の examples を改変したものです。ライセンス本文は [LICENSE-APACHE](LICENSE-APACHE) にあります。

| ファイル | 元のファイル | 著作権表示 |
| --- | --- | --- |
| `games/snake/` | `examples/snake` | Copyright 2020 The Ebiten Authors |
| `games/flappy/` | `examples/flappy` | Copyright 2018 The Ebiten Authors |
| `games/blocks/` | `examples/blocks/blocks` | Copyright 2014, 2015 Hajime Hoshi |
| `compat/snake_main.go.txt`、`compat/flappy_main.go.txt`、`compat/flappy_platform.go.txt`、`compat/blocks_main.go.txt`、`compat/blocks_platform.go.txt` | 上の各サンプル | 上と同じ |
| `docs/porting-flappy.diff`、`docs/porting-blocks.diff` | 上の各サンプルとの差分 | 上と同じ |

元のコードを改変したファイルには、先頭に元の著作権表示とライセンス表示を残し、改変した旨を書いています。変更点は [docs/porting.md](docs/porting.md) にまとめています。各ディレクトリで ikaten 用に新しく書いたファイル(`platform_ikaten.go`、`assets.go` など)も、ディレクトリ内のライセンスをそろえるため Apache License 2.0 で提供します。

## 画像

| ファイル | 出典と作者 | ライセンス | 加工 |
| --- | --- | --- | --- |
| `games/flappy/assets/gopher.png`、`gopher.ika` | Ebitengine `examples/resources/images/flappy/gopher.png`。The Go gopher was designed by Renee French (http://reneefrench.blogspot.com/). | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) | 縦横を半分に縮小し、IKA1形式に変換 |
| `games/bomb/art.go` のGopherのドット絵 | ikaten 用に新しく描いたもの。The Go gopher was designed by Renee French (http://reneefrench.blogspot.com/). | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) | Go Gopherのデザインをもとに、9×11ピクセルのドット絵としてコードで描いた。色違いの3匹も同じ |
| `games/flappy/assets/tiles.png`、`tiles.ika` | Ebitengine `examples/resources/images/flappy/tiles.png`。https://opengameart.org/content/super-mario-world-style-tiles | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/) | 縦横を半分に縮小し、IKA1形式に変換 |
| `games/blocks/assets/blocks.png`、`background.png` と各 `.ika` | Ebitengine `examples/resources/images/blocks/`。Copyright 2014 Hajime Hoshi | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) | PNGは無加工。IKA1形式に変換 |
| `games/blocks/assets/gamebg.png`、`gamebg.ika` | Ebitengine `examples/resources/images/gophers.jpg`。Photograph by Chris Nokleberg (https://go.dev/blog/2years) | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) | 画面の大きさに拡大縮小し、明るい灰色に変換し、半透明のウィンドウを焼き込み、IKA1形式に変換 |

これ以外の画像(`games/assets`、`games/kart/assets/track.png`、`games/uboat/assets`、`games/tank/assets`、dungeon・shooter・kart・uboat・tank・train・air・fishのスプライト、bombのGopher以外のスプライトなど)は ikaten 用に新しく作ったもので、MIT License です。

## 依存ライブラリ

次のライブラリはリポジトリには含めず、Go モジュールとして取得します。ファームウェアや WASM のバイナリにはこれらが組み込まれるため、バイナリを配布するときはそれぞれのライセンスに従ってください。

| ライブラリ | ライセンス |
| --- | --- |
| [tinygo.org/x/drivers](https://github.com/tinygo-org/drivers) v0.37.0 | BSD 3-Clause |
| [tinygo.org/x/tinyfont](https://github.com/tinygo-org/tinyfont) v0.7.0 | BSD 3-Clause。フォントのデータは、それぞれ元のフォントのライセンスに従います。ikaten は `proggy`(DebugPrint)、`freemono`(flappy、blocks)、`TomThumb`(flappy)を使います |
| [Ebitengine](https://github.com/hajimehoshi/ebiten) v2.10.5 | Apache License 2.0。`make compat` で本物の Ebitengine 向けにビルドするときだけ使います |
