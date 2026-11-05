# IKA1アセット形式

`ikaten-asset` はPNGをRGB565と任意の1bitマスクへ変換します。色はstraight RGBへ戻した後にRGB565へ丸め、アルファ128未満は透明にします。ディザリングは行いません。

| オフセット | 長さ | 内容 |
| --- | --- | --- |
| 0 | 4 | ASCII `IKA1` |
| 4 | 2 | 幅、little endian、1〜32767 |
| 6 | 2 | 高さ、little endian、1〜32767 |
| 8 | 2 | flags、little endian、bit0=マスクあり |
| 10 | 2 | 予約、0 |
| 12 | 幅×高さ×2 | RGB565、big endian、行優先 |
| 続き | ceil(幅×高さ/8) | flags=1のときのみ。不透明=1、LSB first |

マスクは行をまたいで連続します。全ピクセルが不透明ならマスクを省きます。ファイル末尾の余分なデータや不正なヘッダーはエラーです。

`NewImageFromBytes` と `NewImageFromString` はデータをコピーせず、画像とバッファの管理情報だけを確保します。SubImageも読み取り専用のまま共有します。データは画像の寿命中保持されます。`NewImageFromString` は `unsafe.StringData` と `unsafe.Slice` で読み取り専用ビューを作り、書き込みAPIでは必ず拒否します。アラインメントを必要とする `[]uint16` へのキャストは使いません。

TinyGoのフラッシュ配置には `//go:embed` の `string` を使用してください。`[]byte` の埋め込みは書き換え可能なためRAMに配置されます。大きな画像をグローバル初期化で `NewImageFromImage` / `NewImage` へ展開するのも避けてください。

```go
flashImage, err := ikaten.NewImageFromString(embeddedData)
if err != nil { return err }
w, h := flashImage.Size()
ramImage := ikaten.NewImage(w, h)
ramImage.DrawImage(flashImage, nil)
```

PNGを実行時に読む場合はGo標準の `image/png` でDecodeし、`NewImageFromImage` で変換できます。その場合は展開中のRGBA画像とRGB565画像の両方がRAMを使います。
