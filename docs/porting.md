# Ebitengineからの移植

まずimportを `ebiten "github.com/sat0ken/ikaten"` に変更し、`inpututil` / `ebitenutil` / `vector` も対応するikatenパッケージへ変更します。起動時に `SetHardware(hardware.Device)` が必要です。

`games/bounce/game.go` は両エンジンで同じ内容を使用します。`make compat` はこのファイルを `build/compat/bounce` にコピーし、`compat/import_ebitengine.go.txt` をimportアダプターとして追加して、本物のEbitengine v2.10.5でWASMをビルドします。ゲーム本体の一致も `cmp` で検証します。ikaten側のアダプターは `games/bounce/import_ikaten.go` です。起動用mainはハードウェア選択の有無に合わせて別に用意しています。

## 移植例: snake

`games/snake/game.go` はEbitengine v2.10.5の `examples/snake/main.go` を移植したものです。Apache License 2.0の著作権表示を残しています。`ebiten.X` の呼び出しは書き換えていません。`make compat` はimportパスだけを戻して本物のEbitengineでビルドします。元のソースとの差分は次の全行です。

```diff
-package main
+// This file is modified from Ebitengine v2.10.5 examples/snake/main.go for ikaten.
+// The changes are listed in docs/porting.md.
+
+package snake
-	"log"
-	"github.com/hajimehoshi/ebiten/v2"
-	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
-	"github.com/hajimehoshi/ebiten/v2/inpututil"
-	"github.com/hajimehoshi/ebiten/v2/vector"
+	ebiten "github.com/sat0ken/ikaten"
+	"github.com/sat0ken/ikaten/ebitenutil"
+	"github.com/sat0ken/ikaten/inpututil"
+	"github.com/sat0ken/ikaten/vector"
-	screenWidth        = 640
-	screenHeight       = 480
-	gridSize           = 10
+	screenWidth        = 120
+	screenHeight       = 120
+	gridSize           = 5
-func newGame() *Game {
+func NewGame() *Game {
-func main() {
-	ebiten.SetWindowSize(screenWidth, screenHeight)
-	ebiten.SetWindowTitle("Snake (Ebitengine Demo)")
-	if err := ebiten.RunGame(newGame()); err != nil {
-		log.Fatal(err)
-	}
-}
```

`main` は `examples/snake/main.go` に移し、`SetHardware(hardware.Device)` を追加しました。解像度を120x120にしたのは、液晶より大きい `Layout` に対応しておらず、240x240ではESP32無印のヒープに入らないためです。そのため、スコアの文字は途中で切れます。

## 移植例: flappy

`games/flappy/game.go` はEbitengine v2.10.5の `examples/flappy/main.go` を移植したものです。snakeと違い、import行以外の変更も必要でした。元のソースとの全差分は [porting-flappy.diff](porting-flappy.diff) にあります。

| 変更 | 理由 |
| --- | --- |
| 640x480を240x240に、タイル、物理量、文字の大きさを半分に | 液晶の解像度に合わせるため。速度と重力も半分にし、見た目の動きを元と同じにした |
| カメラの初期位置を-240から-90に | 画面幅に対するGopherの位置を元と同じ約37%にするため |
| 音声(`audio`)を削除 | ikatenに音声APIがないため |
| CRTシェーダーと `-crt` フラグを削除 | シェーダーに対応しないため |
| `text/v2` とPressStart2Pを `drawText` に置き換え | ikatenは `text/v2` に非互換。ikaten側はtinyfont(freemono、TomThumb)で描く |
| 画像の読み込みを `loadImages` に置き換え | ikaten側はフラッシュ上の `.ika` を参照し、PNGを展開しない |
| `SubImage` を `init` で3つ作り、毎フレームの切り出しをやめた | `SubImage` は毎回ヒープを確保するため |
| `op.Filter = ebiten.FilterLinear` を削除 | ikatenは `FilterNearest` のみ |

`drawText` と `loadImages` は `games/flappy/platform_ikaten.go` にあります。`make compat` は、Ebitengine用の実装 `compat/flappy_platform.go.txt` に差し替え、importパスを戻してビルドします。そのため、`game.go` は両エンジンで同じ内容です。

画像は `games/flappy/assets/gen.go` で元の画像を半分に縮小し、`go run ./cmd/ikaten-asset` で `.ika` に変換しています。Gopherの画像はRenee Frenchのデザインで、CC BY 4.0です(タイトル画面の表示は元のソースのまま)。タイルの画像はCC0です。

ゲーム側では `fmt.Sprintf` が毎フレーム呼ばれます(元のソースのまま)。そのため、ヒープが約3KB/秒で増え、約70秒ごとにGCが走ります。

240x240の画面が必要なため、ESP32無印ではヒープに入らず動きません(ビルドは成功します)。

## 移植例: blocks

`games/blocks` はEbitengine v2.10.5の `examples/blocks/blocks`(落ちものパズル)を移植したものです。全差分は [porting-blocks.diff](porting-blocks.diff) にあります。`font.go` は、ikaten用の `platform_ikaten.go` とEbitengine用の `compat/blocks_platform.go.txt` に置き換えました。それ以外のファイルは両エンジンで共通です。

| 変更 | 理由 |
| --- | --- |
| 画面幅を256から240に | 液晶に合わせるため。右側の表示枠の幅が96から80に縮むだけで、配置は元のまま |
| 場面切り替えをクロスフェードから黒経由のフェードに | クロスフェードには画面サイズの画像が2枚必要で、RAMに入らないため |
| 半透明の枠と背景のグレー化を、背景画像に焼き込み | 画面サイズの画像を常に持つとRAMが足りず、ikatenの画像は半透明を保持できないため。`assets/gen.go` で作る |
| 枠のラベルとゲームオーバー画面を毎フレーム描く | 上と同じ理由で、事前に描いた画像を持たない。ゲームオーバーの半透明の黒は1x1の画像に `ColorScale` を掛けて描く |
| 行消去の点滅を `colorm` から `ColorScale` と `BlendLighter` に | ikatenに `colorm` がないため。透明度を下げた上に赤を加算し、元と同じ計算になるようにした |
| ブロックの `SubImage` を `init` で作る | 毎フレーム200回近く確保していたため |
| 文字をtinyfontに(ikaten側) | `text/v2` に非互換のため。ロゴはfreemono、それ以外はproggy。影は不透明の黒になる |

バッジではジョイスティックで移動と落下、SW1で左回転、SW2で右回転です。タイトルではSW1かSW2を2回押すと始まります(1回目でゲームパッドとして認識されます)。キーボードではSpaceで開始と右回転、Zで左回転です。

背景の写真 `gophers.jpg` はChris Nokleberg撮影で、CC BY 4.0です。ブロックとタイトル背景の画像はEbitengineの `license.md` に記載がなく、Ebitengineのリポジトリ全体のライセンス(Apache License 2.0)に従うものとして扱っています。

## 画像と描画

- 画像のサイズと座標はint、GeoMの引数はfloat64、内部行列はfloat32です。平行移動・整数拡大は高速で、回転・反転・非整数拡大は逆写像を使います。ピクセルのサンプリング計算は整数です。
- `screen.DrawImage(sprite, op)` の向きはEbitengineと同じです。SubImageのBoundsは親座標を保持し、DrawImageのソースの左上は原点へ正規化します。
- SubImageは共有するため、描画先とソースが同じバッファを参照するとpanicします。別の作業画像を用意してください。
- NewImageは透明です。Set/Fill/WritePixelsなどはアルファを128で二値化します。RGBA入出力はpremultiplied形式で、ReadPixels/Atの不透明色はRGB565から展開した値です。
- ColorScaleはpremultiplied RGBAの乗算です。フェードには`ScaleAlpha`を使用します。合成後のアルファは二値なので、半透明の中間レイヤーは事前に背景へ合成してください。
- FilterNearestと3種類のBlendだけを公開します。シェーダー・ColorM・DrawTrianglesのAPIはありません。描画結果を変える未対応機能は移植時に見直してください。
- vectorは `FillRect`、`FillCircle`、`StrokeLine` と非推奨の旧名だけで、アンチエイリアスは使えません。text/v2の代わりに `ikaten/text.Draw` とtinyfontのフォントを使います。DebugPrintの字形もEbitengineとは異なります。

## 入力とゲームループ

キーはゲームのUpdate境界で確定します。catch-up中に3回UpdateしてもJustPressedは最初の1回だけです。ボードで割り当てた入力だけが有効です。マウス・タッチは座標0、押下なし、IDリストへの追加なしを返します。`inpututil` のマウスとタッチの関数も同じく何も報告しません。実行後にハードウェアを変更したり、複数のgoroutineから画像・入力・描画APIを操作することはできません。

Layoutは物理画面に入る正のサイズを返してください。小さい画面は整数倍で拡大します。液晶より大きな論理解像度の縮小は未対応です。既定のTPSは60、SetTPSは1〜1000。`SyncWithFPS` やTPS=0には対応しません。処理落ちが大きいと最大3Updateの制限によりゲームが遅くなります。

画像、SubImage、文字列、アセットは起動時に作って保持してください。SubImageとLayout変更は割り当てを伴います。CPUとSPIの性能判断は実機で行います。RP2040では小さいスプライトの平行移動を中心にする設計を想定しています。
