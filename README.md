# tmpltype

[![Go Reference](https://pkg.go.dev/badge/github.com/bellwood4486/tmpltype.svg)](https://pkg.go.dev/github.com/bellwood4486/tmpltype)

**[English](#english)** | **[日本語](#japanese)**

---

<a name="english"></a>
## English

A Go code generator that creates type-safe template rendering functions from Go template files.

### What is tmpltype?

`tmpltype` eliminates runtime errors in Go templates by generating type-safe structs and render functions. It analyzes your template files and automatically infers parameter types, or you can specify them explicitly.

**Before (runtime errors):**
```go
// ❌ Typo in field name - fails at runtime
tmpl.Execute(w, map[string]any{"Nmae": "Alice"})
```

**After (compile-time safety):**
```go
// ✅ Compile error if field name is wrong
RenderEmail(w, Email{Name: "Alice", Message: "Welcome!"})
```

### Quick Start

**Install:**
```bash
go install github.com/bellwood4486/tmpltype/cmd/tmpltype@latest
```

**1. Create a template** (`templates/email.tmpl`):
```html
<h1>Hello {{ .User.Name }}</h1>
<p>{{ .Message }}</p>
```

**2. Add go:generate directive** (`gen.go`):
```go
package main

//go:generate tmpltype -dir templates -pkg main -out template_gen.go
```

**3. Generate and use:**
```bash
go generate
```

```go
package main

import (
    "bytes"
    "fmt"
)

func main() {
    var buf bytes.Buffer
    _ = RenderEmail(&buf, Email{
        User:    EmailUser{Name: "Alice"},
        Message: "Welcome!",
    })
    fmt.Println(buf.String())
}
```

### Key Features

- **🔒 Type Safety**: Catch template errors at compile time, not runtime
- **🤖 Type Inference**: Automatically infers types from template syntax
- **📝 Explicit Types**: Use `@param` directives for complex types (int, pointers, custom structs)
- **📁 Template Grouping**: Organize templates in subdirectories with nested namespaces
- **🎨 Custom Functions**: Use any custom template functions with functional option pattern
- **🔧 go generate**: Seamless integration with Go's standard workflow
- **💡 IDE Support**: Full autocompletion for template parameters

### Documentation

#### Getting Started
- **[Getting Started Guide](docs/getting-started.md)** - Step-by-step tutorial
- **[Examples](examples/)** - Working code examples for common patterns

#### Reference
- **[CLI Reference](docs/cli-reference.md)** - Command-line options and usage
- **[Template Syntax](docs/template-syntax.md)** - Supported Go template constructs
- **[`@param` Directive](docs/param-directive.md)** - Complete type directive reference
- **[Template Grouping](docs/template-grouping.md)** - Organize templates in subdirectories

#### 日本語ドキュメント
- **[はじめに](docs/ja/getting-started.md)** - ステップバイステップのチュートリアル
- **[CLIリファレンス](docs/ja/cli-reference.md)** - コマンドラインオプションと使い方
- **[テンプレート構文](docs/ja/template-syntax.md)** - サポートされるGoテンプレート構文
- **[`@param`ディレクティブ](docs/ja/param-directive.md)** - 型ディレクティブの完全リファレンス
- **[テンプレートグルーピング](docs/ja/template-grouping.md)** - サブディレクトリでテンプレートを整理

### Examples

Explore working examples in the [`examples/`](examples/) directory:

- [`01_basic`](examples/01_basic/) - Basic type inference
- [`02_param_directive`](examples/02_param_directive/) - Using `@param` for complex types
- [`03_multi_template`](examples/03_multi_template/) - Multiple templates
- [`04_comprehensive_template`](examples/04_comprehensive_template/) - All template syntax patterns
- [`05_all_param_types`](examples/05_all_param_types/) - Complete `@param` reference
- [`07_grouping`](examples/07_grouping/) - Template grouping with subdirectories
- [`08_custom_functions`](examples/08_custom_functions/) - Custom template functions

Run an example:
```bash
cd examples/01_basic
go generate
go run .
```

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<a name="japanese"></a>
## 日本語

Goテンプレートファイルから型安全なテンプレート描画関数を生成するGoコードジェネレータです。

### tmpltypeとは？

`tmpltype`はGoテンプレートでのランタイムエラーを排除し、型安全な構造体とレンダー関数を生成します。テンプレートファイルを解析してパラメータの型を自動推論するか、明示的に指定することができます。

**従来の方法（ランタイムエラー）:**
```go
// ❌ フィールド名のタイポ - 実行時に失敗
tmpl.Execute(w, map[string]any{"Nmae": "Alice"})
```

**tmpltype使用後（コンパイル時の安全性）:**
```go
// ✅ フィールド名が間違っているとコンパイルエラー
RenderEmail(w, Email{Name: "Alice", Message: "ようこそ！"})
```

### クイックスタート

**インストール:**
```bash
go install github.com/bellwood4486/tmpltype/cmd/tmpltype@latest
```

**1. テンプレートを作成** (`templates/email.tmpl`):
```html
<h1>こんにちは {{ .User.Name }}</h1>
<p>{{ .Message }}</p>
```

**2. go:generateディレクティブを追加** (`gen.go`):
```go
package main

//go:generate tmpltype -dir templates -pkg main -out template_gen.go
```

**3. 生成して使用:**
```bash
go generate
```

```go
package main

import (
    "bytes"
    "fmt"
)

func main() {
    var buf bytes.Buffer
    _ = RenderEmail(&buf, Email{
        User:    EmailUser{Name: "太郎"},
        Message: "ようこそ！",
    })
    fmt.Println(buf.String())
}
```

### 主な機能

- **🔒 型安全性**: テンプレートエラーを実行時ではなくコンパイル時に検出
- **🤖 型推論**: テンプレート構文から自動的に型を推論
- **📝 明示的な型指定**: 複雑な型（int、ポインタ、カスタム構造体）には`@param`ディレクティブを使用
- **📁 テンプレートグルーピング**: サブディレクトリでテンプレートを整理し、ネストされた名前空間を生成
- **🎨 カスタム関数**: functional optionパターンで任意のカスタムテンプレート関数を使用可能
- **🔧 go generate**: Goの標準ワークフローにシームレスに統合
- **💡 IDE サポート**: テンプレートパラメータの完全な自動補完

### ドキュメント

#### はじめに
- **[はじめに](docs/ja/getting-started.md)** - ステップバイステップのチュートリアル
- **[サンプル](examples/)** - よくあるパターンの動作するコード例

#### リファレンス
- **[CLIリファレンス](docs/ja/cli-reference.md)** - コマンドラインオプションと使い方
- **[テンプレート構文](docs/ja/template-syntax.md)** - サポートされるGoテンプレート構文
- **[`@param`ディレクティブ](docs/ja/param-directive.md)** - 型ディレクティブの完全リファレンス
- **[テンプレートグルーピング](docs/ja/template-grouping.md)** - サブディレクトリでテンプレートを整理

#### English Documentation
- **[Getting Started](docs/getting-started.md)** - Step-by-step tutorial
- **[CLI Reference](docs/cli-reference.md)** - Command-line options and usage
- **[Template Syntax](docs/template-syntax.md)** - Supported Go template constructs
- **[`@param` Directive](docs/param-directive.md)** - Complete type directive reference
- **[Template Grouping](docs/template-grouping.md)** - Organize templates in subdirectories

### サンプル

[`examples/`](examples/)ディレクトリの動作するサンプルをご覧ください:

- [`01_basic`](examples/01_basic/) - 基本的な型推論
- [`02_param_directive`](examples/02_param_directive/) - 複雑な型に対する`@param`の使用
- [`03_multi_template`](examples/03_multi_template/) - 複数テンプレート
- [`04_comprehensive_template`](examples/04_comprehensive_template/) - すべてのテンプレート構文パターン
- [`05_all_param_types`](examples/05_all_param_types/) - `@param`の完全リファレンス
- [`07_grouping`](examples/07_grouping/) - サブディレクトリでのテンプレートグルーピング
- [`08_custom_functions`](examples/08_custom_functions/) - カスタムテンプレート関数

サンプルの実行:
```bash
cd examples/01_basic
go generate
go run .
```

### コントリビューション

コントリビューションを歓迎します！プルリクエストを自由に提出してください。

### ライセンス

このプロジェクトはMITライセンスのもとでライセンスされています - 詳細は[LICENSE](LICENSE)ファイルを参照してください。
