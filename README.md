# ikaten

Ebitengineと似たAPIでTinyGo向けの2Dゲームを書くためのパッケージです。RGB565のソフトウェア描画を使い、ESP32-S3 / ESP32 / RP2040のST7789液晶と、ブラウザのWASMプレビューに対応します。

主ターゲットは **TinyGo Conference 2026 badge（ESP32-S3、240×240）**。バッジで表示、入力、転送速度を確認済みで、240×240、スプライト30枚で約43 FPS / 60 TPS(目標は30 FPS)です。Gopher Board(RP2040)でも、120×120の2倍表示、スプライト15枚で約32 FPSを確認しました。

オリジナルのゲームとランチャーは240×240専用に描き直しています(移行中)。画面だけで約122KBを使うため、ESP32無印には対応しません。詳しくは[ハードウェア設定](docs/hardware.md)を見てください。

## 実行

Go 1.24以上、TinyGo **0.44.0**を使用します。driversは **v0.37.0**、tinyfontは **v0.7.0**に固定しています。

```sh
make serve                  # TinyGo WASMをビルドし http://127.0.0.1:8080 でプレビュー
make build                  # バッジ用: build/bounce-badge.elf
make build TARGET=esp32     # 120×120を2倍表示
make build TARGET=pico
make wasm GAME=assets       # マスク付き画像・回転・図形・文字のサンプル
make wasm GAME=flappy       # Ebitengineのflappyを移植したもの
make wasm GAME=blocks       # Ebitengineのblocks(落ちものパズル)を移植したもの
make wasm GAME=dungeon      # オリジナルのレイキャスティング方式の一人称シューティング
make wasm GAME=shooter      # オリジナルの縦スクロールシューティング
make wasm GAME=kart         # オリジナルのカートレース(地面を回転・拡大して描く方式)
make wasm GAME=uboat        # オリジナルの潜水艦ゲーム(潜望鏡で船団を狙う)
make wasm GAME=tank         # オリジナルの戦車戦(砲手の照準器で狙い、装甲の向きと弾の落下を考える)
make wasm GAME=train        # オリジナルの電車の運転(運転台から見た路線で、停止位置と時刻表を守る)
make wasm GAME=air          # オリジナルの空戦(操縦席から見た3D。零戦の直掩、九九艦爆の急降下爆撃、九七艦攻の雷撃)
make wasm GAME=fish         # オリジナルの船釣り(海の断面を横から見て、棚を探し、アタリで合わせ、糸の張りを見て巻く)
make wasm GAME=bomb         # オリジナルの爆弾迷路の対戦(Go Gopherが主人公。CPUのGopher3匹と最後の1匹を目指す)
make wasm GAME=jump         # オリジナルの横スクロールアクション(Go Gopherが走り、跳び、ダッシュで土を壊す。3ステージ)
make wasm GAME=all          # 全ゲームを選んで遊べるランチャー(A+Bを1秒でメニューへ)
```

バッジへの書き込みは接続を確認して実行します。

```sh
tinygo flash -target ./targets/badge.json ./examples/bounce
```

`TARGET` は `badge` / `esp32s3` / `esp32` / `pico` / `gopher-badge` / `gopher-board`。配線と解像度は[ハードウェア設定](docs/hardware.md)に記載しています。通常のGoビルドは画面を開かないHeadlessバックエンドを使用します。

バッジではジョイスティックで移動、SW1で色変更、SW2で一時停止。ブラウザでは矢印、Z、Xが対応します。ブラウザ画面のボタンでも操作できます。

## ゲームの基本形

```go
package main

import (
    "image/color"

    ebiten "github.com/sat0ken/ikaten"
    "github.com/sat0ken/ikaten/hardware"
)

type Game struct{ sprite *ebiten.Image }

func (g *Game) Update() error { return nil }
func (g *Game) Layout(w, h int) (int, int) {
    return hardware.LogicalWidth, hardware.LogicalHeight
}
func (g *Game) Draw(screen *ebiten.Image) {
    var op ebiten.DrawImageOptions
    op.GeoM.Translate(20, 30)
    screen.DrawImage(g.sprite, &op)
}
func main() {
    sprite := ebiten.NewImage(16, 16) // バッファ確保はmain以降
    sprite.Fill(color.RGBA{R: 255, A: 255})
    ebiten.SetHardware(hardware.Device)
    if err := ebiten.RunGame(&Game{sprite: sprite}); err != nil {
        println(err.Error())
    }
}
```

`Update` は既定60TPS。遅れた場合は最大3回更新してから1回描画し、中間の描画を省きます。3回を超える遅れは捨てるため、その場合はゲーム時間も遅くなります。`Layout` が変わると画面を再確保します。物理画面に収まる最大の整数倍で中央に表示し、余白は黒にします。`Termination` を返すと正常終了します。

## アセット

```sh
go run ./cmd/ikaten-asset -o sprite.ika sprite.png
go generate ./games/assets   # 同梱サンプルを再生成
```

フラッシュに置く場合は文字列として埋め込みます。TinyGoでは書き換え可能な `[]byte` の埋め込みが静的RAMに配置されるためです。

```go
import (
    _ "embed"
    "github.com/sat0ken/ikaten"
)

//go:embed sprite.ika
var spriteData string

// mainやゲームの初期化で呼ぶ
func loadSprite() (*ikaten.Image, error) {
    return ikaten.NewImageFromString(spriteData)
}
```

画像のピクセルはコピーされません。`NewImageFromBytes` も同じ形式を参照しますが、渡したスライスの配置場所は呼び出し側が管理します。読み取り専用画像への `Fill` / `Set` / `DrawImage` などはpanicします。RAMコピーが必要なら `NewImage` に描き込みます。形式は[アセット仕様](docs/assets.md)を参照してください。

## 対応範囲

- `Game`、固定TPS、実測TPS/FPS、`Image`、共有`SubImage`、RGBA入出力
- `GeoM`の平行移動・拡大・回転・逆行列、`ColorScale`、Nearest、SourceOver / Copy / Lighter
- キー入力、`inpututil`、GPIOからの仮想ゲームパッド1台
- tinyfont文字描画、DebugPrint、塗り矩形・円・直線
- 行ハッシュによる変更行転送、整数倍表示、計測用ビルドタグ

色はRGB565に丸め、画像のアルファは1bitです。`Set` / `Fill` / PNG変換などはアルファ128を境に透明・不透明へ丸めます。一様アルファを使う合成は描画先の色へ混ぜ、その結果を不透明ピクセルにします。半透明の中間画像を保持することはできません。

`DrawImage` は同じピクセルを共有する画像同士では使えません。`vector`のアンチエイリアス引数は`false`のみ対応し、色のアルファは画像と同じ二値化です。マウス・タッチAPIは空の入力を返し、ウィンドウ設定APIは何もしません。シェーダー・任意三角形・音声は未対応です。[移植ガイド](docs/porting.md)に差分をまとめています。

## 検証と計測

```sh
make check                   # 単体テスト・vet・計測タグ付きテスト
go test -race ./...
make smoketest               # 18サンプル×6ボード + 17サンプル×WASM = 125ビルド
make compat                  # bounce、snake、flappy、blocks、dungeon、shooter、kart、uboat、tank、train、air、fish、bomb、jump、allをEbitengine v2.10.5でWASMビルド
make allocs                  # TinyGoの確保箇所を表示
make probe                   # SPI周波数別の測定用ファームウェア
make benchmark               # 30分実行用の計測ファームウェア
```

`make compat` の依存モジュールは `build/compat` に分離しています。初回はダウンロードが必要です。`make serve` 実行後に `/browsercheck.html` を開くと、WASMのアニメーションとキー操作の自動検証が走ります。

## ライセンス

ikaten は [MIT License](LICENSE) です。ただし、Ebitengine のサンプルを移植した `games/snake`、`games/flappy`、`games/blocks` と関連する互換ビルド・差分のファイルは [Apache License 2.0](LICENSE-APACHE)、flappy と blocks の画像の一部は CC BY 4.0 または CC0 です。対象のファイル、著作権表示、帰属表示は [NOTICE.md](NOTICE.md) にまとめています。

ライブラリ本体、dungeon、shooter、kart、uboat、tank、train、air、fish、bomb、jump、ランチャー(all)は ikaten 用に新しく書いたもので、koebiten と Ebitengine の実装コードはコピーしていません。API の名前と形は Ebitengine に合わせています。依存ライブラリにはそれぞれのライセンスが適用されます。
