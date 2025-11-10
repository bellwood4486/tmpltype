# tmpltype を始める

このガイドでは、`tmpltype`のインストールから最初の型安全なテンプレート作成まで、ステップバイステップで説明します。

## 目次

- [インストール](#インストール)
- [最初のテンプレート](#最初のテンプレート)
- [型推論を理解する](#型推論を理解する)
- [型ディレクティブを使う](#型ディレクティブを使う)
- [複数テンプレートを扱う](#複数テンプレートを扱う)
- [次のステップ](#次のステップ)

## インストール

`go install`を使って`tmpltype`をインストールします：

```bash
go install github.com/bellwood4486/tmpltype/cmd/tmpltype@latest
```

またはプロジェクトに追加：

```bash
go get github.com/bellwood4486/tmpltype
```

インストールの確認：

```bash
tmpltype -h
```

## 最初のテンプレート

シンプルなメールテンプレートを作成して、型安全なコードを生成してみましょう。

### ステップ1: プロジェクト構造を作成

```bash
mkdir myproject
cd myproject
go mod init myproject
```

templatesディレクトリを作成：

```bash
mkdir templates
```

### ステップ2: テンプレートファイルを作成

`templates/email.tmpl`を作成：

```html
<h1>こんにちは {{ .User.Name }}</h1>
<p>{{ .Message }}</p>
<p>送信日時: {{ .Timestamp }}</p>
```

### ステップ3: go:generateディレクティブを追加

プロジェクトルートに`gen.go`を作成：

```go
package main

//go:generate tmpltype -dir templates -pkg main -out template_gen.go
```

### ステップ4: コードを生成

コード生成を実行：

```bash
go generate
```

これにより`template_gen.go`が作成され、以下が含まれます：
- `User`、`Message`、`Timestamp`フィールドを持つ`Email`構造体
- `Name`フィールドを持つネストされた`EmailUser`構造体
- 型安全なレンダリングのための`RenderEmail()`関数
- テンプレート初期化と管理関数

### ステップ5: 生成されたコードを使用

`main.go`を作成：

```go
package main

import (
    "bytes"
    "fmt"
    "time"
)

func main() {
    var buf bytes.Buffer

    err := RenderEmail(&buf, Email{
        User: EmailUser{
            Name: "太郎",
        },
        Message:   "tmpltypeへようこそ！",
        Timestamp: time.Now().Format(time.RFC3339),
    })

    if err != nil {
        panic(err)
    }

    fmt.Println(buf.String())
}
```

プログラムを実行：

```bash
go run .
```

**出力:**
```html
<h1>こんにちは 太郎</h1>
<p>tmpltypeへようこそ！</p>
<p>送信日時: 2025-11-10T15:04:05Z</p>
```

## 型推論を理解する

`tmpltype`はテンプレート構文から自動的に型を推論します：

| テンプレート構文 | 推論される型 | 理由 |
|----------------|---------------|-----|
| `{{ .Name }}` | `string` | シンプルなフィールド参照 |
| `{{ .User.Name }}` | `string`フィールドを持つネストされた構造体 | ネストされたフィールド参照 |
| `{{ range .Items }}{{ .Title }}{{ end }}` | `[]struct{Title string}` | コレクションのrange |
| `{{ index .Meta "key" }}` | `map[string]string` | indexを使ったマップアクセス |

### 例: 複雑なテンプレート

```html
<!-- templates/blog.tmpl -->
<article>
  <h1>{{ .Title }}</h1>
  <p>著者: {{ .Author.Name }} ({{ .Author.Email }})</p>

  <ul>
  {{ range .Tags }}
    <li>{{ . }}</li>
  {{ end }}
  </ul>

  <div>{{ .Content }}</div>
</article>
```

生成される構造体：

```go
type BlogAuthor struct {
    Email string
    Name  string
}

type Blog struct {
    Author  BlogAuthor
    Content string
    Tags    []string
    Title   string
}
```

**ポイント:**
- すべてのフィールド参照はデフォルトで`string`になります
- ネストされたパスはネストされた構造体を作成します
- `range`はスライスを作成します
- フィールドはアルファベット順にソートされます

## 型ディレクティブを使う

特定の型（`string`以外）が必要な場合は、テンプレートコメントで`@param`ディレクティブを使用します。

### 例: 型付きフィールドを持つユーザープロフィール

`templates/profile.tmpl`を作成：

```html
{{/* @param User.ID int64 */}}
{{/* @param User.Age int */}}
{{/* @param User.Bio *string */}}
{{/* @param Items []struct{ID int64; Title string; Price float64} */}}

<div class="profile">
  <h2>ユーザー #{{ .User.ID }}: {{ .User.Name }}</h2>
  <p>年齢: {{ .User.Age }}</p>

  {{ if .User.Bio }}
  <blockquote>{{ .User.Bio }}</blockquote>
  {{ end }}

  <h3>最近の購入</h3>
  <ul>
  {{ range .Items }}
    <li>#{{ .ID }}: {{ .Title }} - ${{ .Price }}</li>
  {{ end }}
  </ul>
</div>
```

生成される型：

```go
type ProfileItemsItem struct {
    ID    int64
    Price float64
    Title string
}

type ProfileUser struct {
    Age  int
    Bio  *string  // オプショナルフィールド（ポインタ）
    ID   int64
    Name string
}

type Profile struct {
    Items []ProfileItemsItem
    User  ProfileUser
}
```

使用例：

```go
bio := "Goエンジニア、コーヒー好き"

var buf bytes.Buffer
err := RenderProfile(&buf, Profile{
    User: ProfileUser{
        ID:   123,
        Name: "太郎",
        Age:  30,
        Bio:  &bio,  // オプショナルフィールド
    },
    Items: []ProfileItemsItem{
        {ID: 1, Title: "Go プログラミング本", Price: 49.99},
        {ID: 2, Title: "メカニカルキーボード", Price: 129.99},
    },
})
```

### よくある型パターン

```go
// 基本型
{{/* @param Age int */}}
{{/* @param Price float64 */}}
{{/* @param Active bool */}}

// オプショナルフィールド（ポインタ）
{{/* @param Email *string */}}
{{/* @param Score *int */}}

// コレクション
{{/* @param Tags []string */}}
{{/* @param Scores []int */}}

// マップ
{{/* @param Metadata map[string]string */}}
{{/* @param Counters map[string]int */}}

// 複雑な構造
{{/* @param Items []struct{ID int64; Name string; Active bool} */}}

// ネストされたフィールド
{{/* @param Config.Database.Host string */}}
{{/* @param Config.Database.Port int */}}
```

完全な型リファレンスについては、[`@param`ディレクティブドキュメント](param-directive.md)を参照してください。

## 複数テンプレートを扱う

`tmpltype`は`-dir`オプションを使って複数のテンプレートを一度に処理できます。

### プロジェクト構造の例

```
myproject/
├── gen.go
├── main.go
├── template_gen.go (生成される)
└── templates/
    ├── email.tmpl
    ├── sms.tmpl
    └── push.tmpl
```

### 単一の生成コマンド

```go
//go:generate tmpltype -dir templates -pkg main -out template_gen.go
```

これにより以下が生成されます：
- `Email`、`Sms`、`Push`構造体
- `RenderEmail()`、`RenderSms()`、`RenderPush()`関数
- 動的レンダリング用の`Template`名前空間
- 汎用の`Render()`関数

### 型安全なレンダリング

```go
// 型安全：フィールド名が間違っているとコンパイルエラー
var buf bytes.Buffer
_ = RenderEmail(&buf, Email{...})
_ = RenderSms(&buf, Sms{...})
```

### 動的レンダリング

```go
// 動的：実行時のテンプレート選択
var buf bytes.Buffer
templateName := getTemplateNameFromConfig()
_ = Render(&buf, templateName, data)
```

サブディレクトリでのテンプレート整理については、[テンプレートグルーピングドキュメント](template-grouping.md)を参照してください。

## 次のステップ

基本を理解したら、以下を探索してください：

- **[CLIリファレンス](cli-reference.md)** - すべてのコマンドラインオプション
- **[テンプレート構文](template-syntax.md)** - サポートされるGoテンプレート構文
- **[`@param`ディレクティブ](param-directive.md)** - 型ディレクティブの完全リファレンス
- **[テンプレートグルーピング](template-grouping.md)** - サブディレクトリでのテンプレート整理
- **[サンプル](../../examples/)** - 動作するコード例

### 推奨される学習パス

1. **[サンプル 01: 基本](../../examples/01_basic/)** - 基本的な型推論から始める
2. **[サンプル 02: Paramディレクティブ](../../examples/02_param_directive/)** - `@param`の使い方を学ぶ
3. **[サンプル 03: 複数テンプレート](../../examples/03_multi_template/)** - 複数テンプレート
4. **[サンプル 07: グルーピング](../../examples/07_grouping/)** - テンプレート整理

各サンプルは`go generate && go run .`で実行できます。
