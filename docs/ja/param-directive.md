# `@param`ディレクティブリファレンス

> **📖 英語版の詳細ドキュメント:** [`@param` Directive Reference (English)](../param-directive.md)

`@param`ディレクティブを使用すると、テンプレートパラメータの型を明示的に指定でき、自動型推論を上書きできます。これは、特定の整数サイズ、オプショナルフィールド（ポインタ）、構造化データなどの複雑な型に不可欠です。

## 目次

- [構文](#構文)
- [なぜ@paramを使うのか](#なぜparamを使うのか)
- [サポートされる型](#サポートされる型)
- [既知の制限事項](#既知の制限事項)
- [ベストプラクティス](#ベストプラクティス)
- [完全な例](#完全な例)

## 構文

```go
{{/* @param <フィールドパス> <型> */}}
```

**パラメータ:**
- `<フィールドパス>`: ドット区切りのフィールドパス（例: `User.Name`、`Items`、`Config.Database.Host`）
- `<型>`: Go型式（以下のサポートされる型を参照）

**例:**
```go
{{/* @param User.Age int */}}
{{/* @param Items []struct{ID int64; Title string} */}}
```

## なぜ@paramを使うのか

デフォルトでは、`tmpltype`はすべてのフィールドを`string`として推論します。以下の場合に`@param`を使用します：

### ✅ 特定の数値型

```go
{{/* @param User.ID int64 */}}        // データベースID
{{/* @param Price float64 */}}         // 小数精度
{{/* @param Count int */}}             // 整数カウント
```

### ✅ オプショナルフィールド

```go
{{/* @param Email *string */}}         // nilの可能性がある
{{/* @param Score *int */}}            // nilの可能性がある
```

### ✅ 複雑な構造

```go
{{/* @param Items []struct{ID int64; Name string; Price float64} */}}
{{/* @param Config map[string]int */}}
```

### ✅ 型安全性

```go
// @paramなし: intを渡すと実行時エラー
{{/* テンプレートで {{ .Age }} を使用 */}}
RenderTemplate(w, Params{Age: "25"})  // OK（string）
RenderTemplate(w, Params{Age: 25})    // ❌ コンパイルエラー

// @paramあり: コンパイル時の安全性
{{/* @param Age int */}}
{{/* テンプレートで {{ .Age }} を使用 */}}
RenderTemplate(w, Params{Age: 25})    // ✅ OK（int）
RenderTemplate(w, Params{Age: "25"})  // ❌ コンパイルエラー
```

## サポートされる型

### 基本型

すべてのGo基本型がサポートされています：

```go
{{/* @param Age int */}}
{{/* @param UserID int64 */}}
{{/* @param Score int32 */}}
{{/* @param Tiny int8 */}}

{{/* @param Count uint */}}
{{/* @param Size uint64 */}}

{{/* @param Price float64 */}}
{{/* @param Rating float32 */}}

{{/* @param Active bool */}}

{{/* @param Data byte */}}
{{/* @param Char rune */}}

{{/* @param Value any */}}  // interface{}相当
```

**注意:** `string`はデフォルトの推論型なので、`@param`で明示的に宣言する必要はありません。

### ポインタ型（オプショナル/Null許可）

任意の型を`*`でラップしてオプショナルにできます：

```go
{{/* @param Email *string */}}
{{/* @param Age *int */}}
{{/* @param Score *int64 */}}
{{/* @param Price *float64 */}}
{{/* @param Active *bool */}}
```

### スライス

任意の基本型のスライス：

```go
{{/* @param Tags []string */}}
{{/* @param IDs []int */}}
{{/* @param Scores []int64 */}}
{{/* @param Prices []float64 */}}
{{/* @param Flags []bool */}}
```

### マップ

`string`キーを持つマップ：

```go
{{/* @param Metadata map[string]string */}}
{{/* @param Counters map[string]int */}}
{{/* @param Scores map[string]int64 */}}
{{/* @param Prices map[string]float64 */}}
{{/* @param Flags map[string]bool */}}
```

**⚠️ 重要:** マップキーは常に`string`である必要があります。他のキー型はサポートされていません。

### ネストされた構造体フィールド（ドット記法）

ドット記法を使用してネストされた構造体フィールドを定義：

```go
{{/* @param User.ID int64 */}}
{{/* @param User.Age int */}}

{{/* @param Config.Database.Port int */}}
```

`User.Name`や`User.Email`、`Config.Database.Host`などのstring型フィールドは、デフォルトで`string`なので`@param`宣言は不要です。

### 構造体のスライス

スライス要素のインライン構造体型を定義：

```go
{{/* @param Items []struct{ID int64; Price float64} */}}
{{/* @param Records []struct{Tags []string; Score *int} */}}
```

`Title`や`Name`などのstring型フィールドは、デフォルトで`string`なので宣言は不要です。

**⚠️ 重要:** 構造体フィールドの区切りには**セミコロン**（`;`）を使用し、カンマは使用しません。

```go
// ❌ 間違い - カンマは使用不可
{{/* @param Item struct{ID int, Price float64} */}}

// ✅ 正しい - セミコロンを使用
{{/* @param Item struct{ID int; Price float64} */}}
```

## 既知の制限事項

### ❌ ネストされたスライス/マップ

直接ネストされたスライスとマップはサポートされていません：

```go
// ❌ 動作しない - 無効な構文を生成
{{/* @param Matrix [][]string */}}
{{/* @param Groups map[string][]string */}}
{{/* @param Data []map[string]int */}}
```

**回避策:** 構造体のスライスを使用：

```go
// ✅ 動作する - 構造体でラップ
{{/* @param Matrix []struct{Row []string} */}}
{{/* @param Groups []struct{Key string; Values []string} */}}
{{/* @param Data []struct{Items map[string]int} */}}
```

### ❌ トップレベルでのインライン構造体

トップレベルで直接`struct{...}`を使用できません：

```go
// ❌ 動作しない - 無効なGoコードを生成
{{/* @param User struct{ID int64; Name string} */}}
```

**回避策:** ドット記法を使用：

```go
// ✅ 動作する
{{/* @param User.ID int64 */}}
```

### ❌ 深くネストされたパスとインライン構造体

深いパスとインライン構造体定義を組み合わせることはできません：

```go
// ❌ 動作しない - ドットを含む型名を生成
{{/* @param Complex.Nested.User struct{ID int64; Name string} */}}
```

**回避策:** 構造をフラット化：

```go
// ✅ 動作する
{{/* @param Complex.Nested.User.ID int64 */}}
{{/* @param Complex.Nested.User.Name string */}}
```

### ❌ 文字列以外のマップキー

マップキーは常に`string`である必要があります：

```go
// ❌ サポートされていない
{{/* @param Lookup map[int]string */}}
{{/* @param Index map[int64]bool */}}
```

## ベストプラクティス

### ✅ 推奨

**`@param`ディレクティブには常にトリムマーカー（`{{-`と`-}}`）を使用:**

`@param`ディレクティブは出力を生成しないため、常にGoテンプレートのトリムマーカーを使用して、レンダリング出力に空行が含まれないようにする必要があります。

```go
// ✅ 推奨: 出力に空行が含まれない
{{- /* @param User.Name string */ -}}
{{- /* @param User.Age int */ -}}
{{- /* @param Items []struct{ID int64; Title string} */ -}}
<div>コンテンツがここから始まります</div>

// ❌ 非推奨: 空行が生成される
{{/* @param User.Name string */}}
{{/* @param User.Age int */}}
{{/* @param Items []struct{ID int64; Title string} */}}
<div>コンテンツがここから始まります</div>  {{/* この前に3行の空行 */}}
```

**ネストされた構造にはドット記法を使用:**
```go
{{- /* @param User.Age int */ -}}
{{- /* @param Config.Database.Port int */ -}}
```

**複雑なコレクションには`[]struct{...}`を使用:**
```go
{{- /* @param Items []struct{ID int64; Price float64} */ -}}
```

**オプショナルフィールドにはポインタ型を使用:**
```go
{{- /* @param Email *string */ -}}
{{- /* @param Score *int */ -}}
```

**フィールドパスは比較的フラットに（1〜2レベル）:**
```go
// ✅ 良い
{{/* @param User.Age int */}}
{{/* @param Config.Port int */}}

// ⚠️ 動作するが冗長
{{/* @param App.Config.Database.Connection.Pool.MaxSize int */}}
```

**構造体フィールドにはセミコロンを使用:**
```go
{{/* @param Item struct{ID int; Price float64} */}}
```

### ❌ 非推奨

**トップレベルでインライン`struct{...}`を使用しない:**
```go
// ❌ 間違い
{{/* @param User struct{ID int64} */}}

// ✅ 正しい
{{/* @param User.ID int64 */}}
```

**スライス/マップを直接ネストしない:**
```go
// ❌ 間違い
{{/* @param Matrix [][]string */}}

// ✅ 正しい
{{/* @param Matrix []struct{Row []string} */}}
```

## 完全な例

### 例1: Eコマース商品

```html
{{/* @param Product.ID int64 */}}
{{/* @param Product.Price float64 */}}
{{/* @param Product.InStock bool */}}
{{/* @param Product.Description *string */}}
{{/* @param Tags []string */}}
{{/* @param Reviews []struct{Rating int} */}}

<div class="product">
  <h2>{{ .Product.Name }} (#{{ .Product.ID }})</h2>
  <p class="price">¥{{ .Product.Price }}</p>

  {{ if .Product.InStock }}
    <span class="badge">在庫あり</span>
  {{ else }}
    <span class="badge out">在庫切れ</span>
  {{ end }}

  {{ if .Product.Description }}
    <p>{{ .Product.Description }}</p>
  {{ end }}

  <div class="tags">
    {{ range .Tags }}
      <span class="tag">{{ . }}</span>
    {{ end }}
  </div>

  <div class="reviews">
    {{ range .Reviews }}
      <div class="review">
        <span class="rating">{{ .Rating }}/5</span>
        <p>{{ .Comment }}</p>
        <small>- {{ .Author }}</small>
      </div>
    {{ end }}
  </div>
</div>
```

### 例2: 完全なリファレンス

すべてのサポートされる型パターンと制限事項を示す包括的で実行可能な例については、[サンプル05: All Param Types](../../examples/05_all_param_types/)を参照してください。

サンプルの実行：
```bash
cd examples/05_all_param_types
go generate
go run .
```

## 関連項目

- [はじめに](getting-started.md) - チュートリアルと基本
- [テンプレート構文](template-syntax.md) - サポートされるテンプレート構文
- [CLIリファレンス](cli-reference.md) - コマンドラインオプション
- [サンプル02: Paramディレクティブ](../../examples/02_param_directive/) - 基本的な`@param`の使用
- [サンプル05: All Param Types](../../examples/05_all_param_types/) - 完全なリファレンス
- [English version (詳細)](../param-directive.md) - Complete reference in English
