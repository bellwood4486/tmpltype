# CLIリファレンス

`tmpltype`ツールの完全なコマンドラインリファレンスです。

## 目次

- [概要](#概要)
- [オプション](#オプション)
- [ロギング](#ロギング)
- [使用例](#使用例)
- [ディレクトリスキャンの動作](#ディレクトリスキャンの動作)
- [go generateとの統合](#go-generateとの統合)
- [トラブルシューティング](#トラブルシューティング)

## 概要

```bash
tmpltype -dir <directory> -pkg <name> -out <file>
```

指定されたディレクトリ内のテンプレートファイルから型安全なGoコードを生成します。

## オプション

### `-dir` (必須)

**型:** `string`
**説明:** スキャンするテンプレートディレクトリ

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

**スキャン動作:**
- `<dir>/*.tmpl` をスキャン（ルートのフラットなテンプレート）
- `<dir>/*/*.tmpl` をスキャン（グループ化されたテンプレート、深さ1まで）
- 深さ1より深いサブディレクトリはスキャン**されません**

**例:**

```bash
# templates/*.tmpl をスキャン
tmpltype -dir templates -pkg main -out gen.go

# ./web/templates/*.tmpl と ./web/templates/*/*.tmpl をスキャン
tmpltype -dir ./web/templates -pkg web -out templates_gen.go

# 相対パスも使用可能
tmpltype -dir ../shared/templates -pkg shared -out gen.go
```

### `-pkg` (必須)

**型:** `string`
**説明:** 出力パッケージ名

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

生成されるコードの先頭に`package <name>`が付きます。

**例:**

```bash
# package main
tmpltype -dir templates -pkg main -out gen.go

# package templates
tmpltype -dir tpl -pkg templates -out templates.go

# package myapp
tmpltype -dir views -pkg myapp -out views_gen.go
```

### `-out` (必須)

**型:** `string`
**説明:** 出力ファイルパス

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

**ベストプラクティス:**
- 生成されたコードを示すために`_gen.go`サフィックスを使用
- 出力ファイルを同じパッケージディレクトリに配置
- 生成されたコードをコミットしない場合は`.gitignore`に追加（非推奨）

**例:**

```bash
# 標準的な命名
tmpltype -dir templates -pkg main -out template_gen.go

# カスタム命名
tmpltype -dir tpl -pkg web -out web_templates.go

# パス付き
tmpltype -dir ../templates -pkg shared -out ../shared/templates_gen.go
```

## ロギング

`TMPLTYPE_LOG_LEVEL`環境変数を使用してtmpltypeの出力の詳細度を制御します。

### ログレベル

#### `info` (デフォルト)

処理中のテンプレート名のみを表示：

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

**出力:**
```
[template] email
[template] notification
```

#### `debug`

フィールド参照、型推論、スキーマ構築を含む詳細なスキャン情報を表示：

```bash
TMPLTYPE_LOG_LEVEL=debug tmpltype -dir templates -pkg main -out template_gen.go
```

**出力:**
```
[template] email
[scan:ref] count=2
[scan:ref] path=User.Name usage=leaf
[scan:ref] path=Message usage=leaf
[scan:pathinfo] count=3
[scan:pathinfo] hasChild=false path=Message usages=leaf
[scan:pathinfo] hasChild=true path=User usages=
[scan:pathinfo] hasChild=false path=User.Name usages=leaf
[scan:kind] count=3
[scan:kind] kind=String path=Message
[scan:kind] kind=Struct path=User
[scan:kind] kind=String path=User.Name
[scan:schema] fields=2
[scan:schema] field=Message kind=String
[scan:schema] field=User kind=Struct
[scan:schema] field=  Name kind=String
```

### go generateでの使用

`gen.go`ファイルでログレベルを設定：

```go
package main

//go:generate env TMPLTYPE_LOG_LEVEL=debug tmpltype -dir templates -pkg main -out template_gen.go
```

またはセッション全体で一度設定：

```bash
export TMPLTYPE_LOG_LEVEL=debug
go generate ./...
```

### デバッグ出力のフィルタリング

デバッグ出力はgrep向けに設計されています：

```bash
# Kind推論のみ表示
TMPLTYPE_LOG_LEVEL=debug tmpltype ... | grep "\[scan:kind\]"

# 特定のパスの全情報を表示
TMPLTYPE_LOG_LEVEL=debug tmpltype ... | grep "path=User"

# スキーマツリーのみ表示
TMPLTYPE_LOG_LEVEL=debug tmpltype ... | grep "\[scan:schema\]"
```

### ユースケース

**デフォルト (info):**
- プロダクションビルド
- CI/CDパイプライン
- どのテンプレートが処理されているかの迅速なフィードバック

**デバッグ:**
- 型推論の決定を理解する
- フィールドが予期しない型になっている理由のトラブルシューティング
- tmpltypeがテンプレートをどのように解析するかを学ぶ
- 複雑なテンプレート構造のデバッグ

## 使用例

### 基本的な使用

単一ディレクトリから生成：

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

**プロジェクト構造:**
```
myproject/
├── main.go
├── template_gen.go (生成される)
└── templates/
    └── email.tmpl
```

### 複数テンプレート（フラット構造）

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

**プロジェクト構造:**
```
myproject/
├── main.go
├── template_gen.go (生成される)
└── templates/
    ├── email.tmpl
    ├── sms.tmpl
    └── push.tmpl
```

**生成されるもの:**
- `Email`、`Sms`、`Push`型
- `RenderEmail()`、`RenderSms()`、`RenderPush()`関数

### テンプレートグルーピング（サブディレクトリ）

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

**プロジェクト構造:**
```
myproject/
├── main.go
├── template_gen.go (生成される)
└── templates/
    ├── footer.tmpl                    # フラット
    ├── mail_invite/                   # グループ
    │   ├── title.tmpl
    │   └── content.tmpl
    └── mail_account_created/          # グループ
        ├── title.tmpl
        └── content.tmpl
```

**生成されるもの:**
- フラット: `Footer`型、`RenderFooter()`関数
- グループ: `MailInviteTitle`、`MailInviteContent`など
- `Template`構造体内のネストされた名前空間

詳細は[テンプレートグルーピングドキュメント](template-grouping.md)を参照してください。

### 相対パスの使用

```bash
# プロジェクトルートから
tmpltype -dir ./internal/views/templates -pkg views -out ./internal/views/templates_gen.go

# サブディレクトリから
cd internal/views
tmpltype -dir templates -pkg views -out templates_gen.go
```

### 絶対パスの使用

```bash
tmpltype \
  -dir /home/user/project/templates \
  -pkg main \
  -out /home/user/project/template_gen.go
```

## ディレクトリスキャンの動作

### スキャンされるもの

`tmpltype`は自動的に以下をスキャンします：

1. **フラットなテンプレート**: `<dir>/*.tmpl`
2. **グループ化されたテンプレート**: `<dir>/*/*.tmpl`（深さ1のみ）

### 深さ制限

**✅ スキャンされる（深さ0と1）:**
```
templates/
├── email.tmpl              ← 深さ0（スキャンされる）
└── mail/
    └── invite.tmpl         ← 深さ1（スキャンされる）
```

**❌ スキャンされない（深さ2以上）:**
```
templates/
└── mail/
    └── invite/
        └── html.tmpl       ← 深さ2（スキャンされない）
```

### ファイルパターン

`.tmpl`拡張子を持つファイルのみが処理されます：

**✅ 処理される:**
- `email.tmpl`
- `user.tmpl`
- `index.tmpl`

**❌ 無視される:**
- `email.txt`
- `user.html`
- `template.go`
- `.tmpl.bak`

### 隠しファイルとディレクトリ

`.`で始まる隠しファイルとディレクトリは**無視**されます：

**❌ 無視される:**
```
templates/
├── .backup/
│   └── old.tmpl            ← 無視される（隠しディレクトリ）
├── .template.tmpl          ← 無視される（隠しファイル）
└── email.tmpl              ← 処理される
```

## go generateとの統合

### 基本セットアップ

**1. パッケージ内に`gen.go`を作成:**

```go
package main

//go:generate tmpltype -dir templates -pkg main -out template_gen.go
```

**2. 生成を実行:**

```bash
go generate
```

または

```bash
go generate ./...
```

### 複数パッケージ

**プロジェクト構造:**
```
myproject/
├── cmd/
│   └── server/
│       ├── gen.go
│       └── templates/
└── internal/
    └── mailer/
        ├── gen.go
        └── templates/
```

**cmd/server/gen.go:**
```go
package main

//go:generate tmpltype -dir templates -pkg main -out template_gen.go
```

**internal/mailer/gen.go:**
```go
package mailer

//go:generate tmpltype -dir templates -pkg mailer -out template_gen.go
```

**すべて実行:**
```bash
go generate ./...
```

### 生成コードのコミット

**推奨:** 生成されたコードをコミットする

```bash
git add template_gen.go
git commit -m "テンプレート生成コードを更新"
```

**理由:**
- ✅ `tmpltype`へのビルド依存がない
- ✅ CIビルドが高速
- ✅ コードレビューで差分が明確
- ✅ `go get`で動作する

**CI検証:**

CIで生成コードが最新であることを確認：

```bash
#!/bin/bash
# scripts/verify_generated.sh
go generate ./...
git diff --exit-code || (echo "生成コードが古い。'go generate ./...'を実行してください" && exit 1)
```

### 代替案: ビルド時に生成

生成コードをコミットしたくない場合：

**Makefile:**
```makefile
.PHONY: generate
generate:
	go generate ./...

.PHONY: build
build: generate
	go build ./...
```

**ビルド:**
```bash
make build
```

## トラブルシューティング

### エラー: "no templates found"

**原因:** 指定されたディレクトリに`.tmpl`ファイルがない

**解決方法:**
```bash
# テンプレートが存在するか確認
ls -la templates/

# 正しいパスを使用しているか確認
tmpltype -dir templates -pkg main -out gen.go
```

### エラー: "failed to parse template"

**原因:** テンプレート構文エラー

**解決方法:**
```bash
# テンプレート構文を確認
# 閉じていない{{ }}、欠けているendタグなどを探す
```

**よくある問題:**
- `{{ end }}`の欠落
- 閉じていない`{{`または`}}`
- 無効なGoテンプレート構文

### エラー: "cannot infer type"

**原因:** 自動推論できない複雑な型

**解決方法:** `@param`ディレクティブを使用

```html
{{/* @param Items []struct{ID int64; Name string} */}}
{{ range .Items }}
  <li>{{ .ID }}: {{ .Name }}</li>
{{ end }}
```

### 生成コードにコンパイルエラー

**原因:** 通常は`@param`ディレクティブの構文に関連

**解決方法:** `@param`ディレクティブを確認

**よくある問題:**
```go
// ❌ 間違い：構造体内でカンマ
{{/* @param User struct{Name string, Age int} */}}

// ✅ 正しい：構造体内でセミコロン
{{/* @param User struct{Name string; Age int} */}}

// ❌ 間違い：ネストされたスライス
{{/* @param Matrix [][]string */}}

// ✅ 正しい：構造体を使用
{{/* @param Matrix []struct{Row []string} */}}
```

詳細は[`@param`ディレクティブドキュメント](param-directive.md)を参照してください。

### サブディレクトリのテンプレートが見つからない

**原因:** 深さ1より深いサブディレクトリはスキャンされない

**解決方法:** テンプレートを深さ1以内に保つ

```
templates/
├── email.tmpl           ← スキャンされる
└── mail/
    └── invite.tmpl      ← スキャンされる
```

サポートされていない:
```
templates/
└── mail/
    └── invite/
        └── html.tmpl    ← スキャンされない（深すぎる）
```

### パーミッションエラー

**エラー:** `permission denied: templates/`

**解決方法:**
```bash
# ディレクトリパーミッションを確認
ls -ld templates/

# 必要に応じてパーミッションを修正
chmod 755 templates/
chmod 644 templates/*.tmpl
```

### Windowsでのパスの問題

**問題:** Windowsのパス区切り文字

**解決方法:** Windowsでもフォワードスラッシュを使用

```bash
# ✅ Windowsでも動作
tmpltype -dir templates/email -pkg main -out gen.go

# ❌ 問題が発生する可能性
tmpltype -dir templates\email -pkg main -out gen.go
```

## 関連項目

- [はじめに](getting-started.md) - チュートリアルと基本概念
- [テンプレート構文](template-syntax.md) - サポートされるテンプレート構文
- [`@param`ディレクティブ](param-directive.md) - 型ディレクティブリファレンス
- [テンプレートグルーピング](template-grouping.md) - テンプレートの整理
