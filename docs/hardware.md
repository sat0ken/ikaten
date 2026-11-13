# ハードウェア設定と実測

## TinyGo Conference 2026 badge

`targets/badge.json` は、既存バッジプロジェクトと同じ `esp32s3-box-3` を継承し、`ikaten_badge` タグで専用バックエンドを選択します。無線や音声は初期化しません。

配線の参照元は、バッジの既存ファームウェアの `firmware/badge/{badge,display,input}.go` です。このディレクトリへの実行時依存はありません。

| 機能 | GPIO / 設定 |
| --- | --- |
| ST7789 SCK / MOSI | 12 / 11、SPI0 |
| CS / DC / RESET | 10 / 5 / 4 |
| バックライト | 3V3直結、NoPin |
| 液晶 | 240×240、回転なし、60Hz設定 |
| SPI要求周波数 | 80MHz（実測と目視確認の結果。既存ファームウェアは40MHz） |
| SW1 / SW2 | 13 / 14、active low、Z / X |
| スティック押下 | 15、active low、Enter |
| スティック X / Y | ADC GPIO7 / GPIO6、矢印キーへ変換 |

起動時の液晶初期化後、約80msの間スティックを中立に保ってください。その間の16サンプルから中心を測ります。Xは値が下がる方向を右、Yは値が上がる方向を上にします。中心から6000カウントで押下、4000未満で解放するヒステリシスを入れています。値と向きは `hardware/board_badge.go` で調整できます。デジタルボタンは5msの時間ベースのデバウンスです。

仮想ゲームパッドIDは0、A/BはStandardGamepadButtonRightBottom / RightRight、方向はD-pad。左スティックAPIは方向キーから-1 / 0 / 1を返します。ADCの精密なアナログ値は公開していません。

## その他のターゲット

| ターゲット | 物理サイズ | 論理サイズ | SPI | SCK/MOSI | CS/DC/RST/BL |
| --- | --- | --- | --- | --- | --- |
| esp32s3 | 240×240 | 240×240 | SPI0、40MHz | 12/11 | 10/5/4/NoPin |
| esp32 | 240×240 | 120×120 | SPI2、40MHz | 18/23 | 5/2/4/NoPin |
| pico | 240×240 | 120×120 | SPI0、62.5MHz | 18/19 | 17/20/21/NoPin |
| gopher-badge | 320×240 | 120×120 | SPI0、62.5MHz | 基板のTFT定数 | 基板のTFT定数 |
| gopher-board | 240×240 | 120×120 | SPI1、62.5MHz | 10/11 | 13/12/9/14 |

`gopher-board` は [Gopher Board](https://github.com/sat0ken/gopher-board-workshop)(RP2040-Zero)です。ピンはワークショップの回路図と `08_st7789_txt` から取りました。液晶は `ROTATION_90` です。スイッチはどれも押すとGNDに落ち、上GP3、左GP4、右GP5、下GP6、A GP26(Z)、B GP15(X)に割り当てています(koebitenの `gopher-board-spi` と同じ)。ワークショップの例は16MHzですが、62.5MHzで表示が乱れないことを目視で確認しました。

genericターゲットは接続例です。入力は配線を確認して `Device.Config.Buttons` に追加します。Gopher BadgeはA/Bと十字キーを設定済みです。液晶の回転・オフセット・サイズは `Device.Config.Panel`、周波数は `Device.Config.Frequency` を `RunGame` の前に変更できます。設定コードは `tinygo` と対象ボードのタグが付いたファイルに置いてください。WASMとHeadlessのDeviceにはConfigはありません。

240×240画面のRAM画像はピクセル115,200Bとマスク7,200Bです。120×120は28,800Bと1,800B。表示側には240px幅で480B（1倍）/960B（2倍）の行バッファ、論理行数×5Bのハッシュと有効フラグが必要です。画像は `main` 以降に確保します。

オリジナルのゲーム(train、air、fish、bomb、jump、dungeon、shooter、kart、uboat、tank)とランチャーは240×240専用に描き直しています(移行中。進み具合は[検証記録](validation.md#240240への移行))。画面の画像だけで約122KBを使うため、ESP32無印には対応しません。RP2040のボードはビルドできますが、実機では確かめていません。

## 転送方式

drivers v0.37.0の `DrawRGBBitmap8` でビッグエンディアンの画素を同期転送します。1倍ではフレームバッファをコピーせず、連続する変更行を1回で送ります。2倍では同じ行を2行分に展開して送ります。さらに大きい整数倍でも最大2物理行のバッファに収めて分割します。

毎フレーム論理画像を行単位の32bit FNV-1a(16bitピクセル単位)で比較し、変化がある行だけ転送します。初回と解像度変更時は物理画面全体を黒で消します。Writerがエラーを返した行はハッシュを確定せず、次回再送します。ドライバ内部のSPI Txエラーはv0.37.0では返されないため、その部分のエラー検出は未対応です。ハッシュ衝突による変更の見逃しの可能性は残ります。

## P0測定

```sh
make probe TARGET=badge
tinygo flash -target ./targets/badge.json ./examples/probe
tinygo monitor -target ./targets/badge.json
```

シリアルポートを読むプログラムを、書き込みと同時に動かさないでください。ブートローダーの応答を奪い、書き込みが止まります。

シリアルに `kind,requested_hz,width,height,band_rows,iterations,total_us,bytes` のCSVを出します。S3は20/40/80MHz、ESP32は20/40MHz、RP2040は31.25/62.5MHzです。周波数は要求値で、実際のSPIクロックはmachine実装の分周設定に依存します。1行/4行転送の平均時間は `total_us / iterations` で計算します。16×16スプライトの描画はRAMとフラッシュ、不透明とマスク付きの4種類を各10,000回測ります。`present_unchanged` は変更のないフレームの `Present` で、行の読み出しとハッシュの計算時間です。

赤の単色表示を目視し、色・向き・乱れを記録してください。このプローブは全周波数で赤を送るため、個々の送信間隔よりパネルの安定性を重点的に確認します。ピンやパネル設定に問題がある場合は先に修正します。

### RP2040のシリアル出力

RP2040では `tinygo monitor` を使ってください。`cat` で読むと出力が取れないことがありました。

TinyGo 0.44.0のRP2040のUSBシリアルは、ポートが開いた状態のまま読まれなくなると、送信バッファが空くまで `println` が待ち続けます。そのため、probeやbenchmarkの実行中に `tinygo monitor` を閉じると、ゲームも止まります。読み直すと再開します。

さらに、起動直後からシリアルを読んでいても、出力が途中で止まることがありました。flappyとblocksの計測版は、毎秒1行を出力すると約80秒止まりませんでした。一方、bounceのbenchmarkは、同じ対策をしても約12秒で止まりました。原因は特定していません。30分の連続実行には、シリアルを使わず数値を画面に表示する版を使いました。通常のサンプルはシリアルに出力しないため、この影響を受けません。

## 描画の種類ごとの時間

```sh
make drawprobe TARGET=badge
tinygo flash -target ./targets/badge.json ./examples/drawprobe
tinygo monitor -target ./targets/badge.json
```

`examples/drawprobe` は、RAM上の240×240の画像に、塗り、不透明とマスク付きの絵、左右反転、拡大と縮小、縦長の画像、文字を繰り返し描き、`name,iterations,us_per_op` を出します。1行のコピーと塗りの書き方(組み込みの `copy`、16ビットと32ビットのループ)の比較も出します。液晶には送りません。起動から3秒待ってから出力し、行ごとに少し待ちます(続けて出すと行が抜けました)。

ESP32-S3では、組み込みの `copy` が呼ぶROMの `memmove` が遅く、構造体のポインタを通してスライスを読むループは画素ごとに読み直すため遅くなります。ホストで速い書き方がバッジでも速いとは限らないので、描画の変更はこれで確かめてください。結果は[検証記録](validation.md#バッジでの描画の速さ)にあります。

## 30分の性能測定

```sh
make benchmark TARGET=badge
tinygo flash -target ./targets/badge.json -tags ikaten_benchmark ./examples/benchmark
```

`-ldflags="-X main.spiHz=40000000"` でSPI周波数を変えられます。`-X main.masked=1` でスプライトをマスク付きにします。

240pxターゲットは30枚、120pxターゲットは15枚の16×16スプライトを描画し、10秒ごとにTPS/FPS、Update/Draw/Presentの平均時間、HeapAllocを出します。約30分で正常終了します。表示時間にはSPI完了待ちが含まれます。初回の確保とシリアル出力は計測値に影響します。比較区間は初期化後に揃えてください。

バッジの実測値は[検証記録](validation.md)にあります。バッジとGopher Board(RP2040)は性能目標を達成しています。DMA、第2コア、音声、三角形APIの採否は、ESP32無印の実測まで保留します。
