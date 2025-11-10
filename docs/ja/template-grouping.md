# テンプレートグルーピング

> **📖 英語版の詳細ドキュメント:** [Template Grouping (English)](../template-grouping.md)

サブディレクトリでテンプレートを論理的に整理して、ネストされた名前空間を作成し、プロジェクト構造を改善します。

## 目次

- [概要](#概要)
- [なぜテンプレートグルーピングを使うのか](#なぜテンプレートグルーピングを使うのか)
- [ディレクトリ構造](#ディレクトリ構造)
- [生成されるコード構造](#生成されるコード構造)
- [使用パターン](#使用パターン)
- [命名規則](#命名規則)
- [ベストプラクティス](#ベストプラクティス)
- [完全な例](#完全な例)

## 概要

テンプレートグルーピングを使用すると、関連するテンプレートをサブディレクトリに整理できます。各サブディレクトリは生成されるコードでネストされた名前空間になり、より良い整理と名前の衝突回避を提供します。

**基本コンセプト:**
```
templates/
├── footer.tmpl          # フラットなテンプレート → Footer
└── mail_invite/         # グループ → MailInvite名前空間
    ├── title.tmpl       #   → MailInvite.Title
    └── content.tmpl     #   → MailInvite.Content
```

## なぜテンプレートグルーピングを使うのか

### ✅ 論理的な整理

関連するテンプレートをまとめる：

```
templates/
├── mail_invite/
│   ├── title.tmpl
│   └── content.tmpl
├── mail_welcome/
│   ├── title.tmpl
│   └── content.tmpl
└── mail_reset_password/
    ├── title.tmpl
    └── content.tmpl
```

すべての招待メールテンプレートは1つのディレクトリに、ウェルカムメールテンプレートは別のディレクトリに、など。

### ✅ 名前の衝突回避

グルーピングなしでは、すべてのテンプレートに一意の名前が必要：

```
# ❌ グルーピングなし - 冗長な名前
templates/
├── mail_invite_title.tmpl
├── mail_invite_content.tmpl
├── mail_welcome_title.tmpl
├── mail_welcome_content.tmpl
└── mail_reset_password_title.tmpl
```

グルーピングを使用すると、各グループ内でシンプルな名前を再利用できます：

```
# ✅ グルーピングあり - クリーンな名前
templates/
├── mail_invite/
│   ├── title.tmpl
│   └── content.tmpl
└── mail_welcome/
    ├── title.tmpl
    └── content.tmpl
```

### ✅ より良いナビゲーション

テンプレートを素早く見つける：
- すべてのメール関連テンプレートは`mail_*/`以下
- すべてのダッシュボード関連テンプレートは`dashboard_*/`以下
- 共有テンプレートはルートに

### ✅ 型安全な名前空間

生成されるコードは整理を反映：

```go
// ネストされた名前空間を通じてテンプレートにアクセス
Template.MailInvite.Title
Template.MailInvite.Content
Template.MailWelcome.Title
Template.MailWelcome.Content
```

## ディレクトリ構造

### フラットとグループ化されたテンプレート

同じプロジェクトで両方のアプローチを混在させることができます：

```
templates/
├── header.tmpl                    # フラット（ルートレベル）
├── footer.tmpl                    # フラット（ルートレベル）
├── mail_invite/                   # グループ
│   ├── title.tmpl
│   └── content.tmpl
└── mail_account_created/          # グループ
    ├── title.tmpl
    └── content.tmpl
```

### スキャン深度

**重要:** `tmpltype`は**深度0と1のみ**のテンプレートをスキャンします。

**✅ スキャンされる:**
```
templates/
├── email.tmpl              ← 深度0（フラット）
└── mail/
    └── invite.tmpl         ← 深度1（グループ化）
```

**❌ スキャンされない:**
```
templates/
└── mail/
    └── invite/
        └── html.tmpl       ← 深度2（スキャンされない）
```

### 推奨される構造

**パターン:** `<category>_<name>/`

```
templates/
├── shared_header.tmpl              # 共有テンプレート（フラット）
├── shared_footer.tmpl
├── mail_invite/                    # メール: 招待
│   ├── title.tmpl
│   └── content.tmpl
├── mail_welcome/                   # メール: ウェルカム
│   ├── title.tmpl
│   └── content.tmpl
├── dashboard_summary/              # ダッシュボード: サマリー
│   └── widget.tmpl
└── dashboard_activity/             # ダッシュボード: アクティビティ
    └── widget.tmpl
```

**利点:**
- 明確なカテゴリ化（mail、dashboardなど）
- 関連するテンプレートを簡単に見つけられる
- 順序付けのための数値プレフィックス（オプション）

## 生成されるコード構造

### テンプレート名前空間

`tmpltype`はネストされた`Template`構造体を生成：

```go
var Template = struct {
    // フラットなテンプレート
    SharedFooter TemplateName
    SharedHeader TemplateName

    // グループ化されたテンプレート
    MailInvite struct {
        Title   TemplateName
        Content TemplateName
    }
    MailWelcome struct {
        Title   TemplateName
        Content TemplateName
    }
}{
    SharedFooter: "shared_footer",
    SharedHeader: "shared_header",
    MailInvite: struct {
        Title   TemplateName
        Content TemplateName
    }{
        Title:   "mail_invite/title",
        Content: "mail_invite/content",
    },
    // ...
}
```

### 型安全なレンダー関数

各テンプレートは独自のレンダー関数を取得：

**フラットなテンプレート:**
```go
func RenderSharedHeader(w io.Writer, p SharedHeader) error
func RenderSharedFooter(w io.Writer, p SharedFooter) error
```

**グループ化されたテンプレート:**
```go
func RenderMailInviteTitle(w io.Writer, p MailInviteTitle) error
func RenderMailInviteContent(w io.Writer, p MailInviteContent) error
func RenderMailWelcomeTitle(w io.Writer, p MailWelcomeTitle) error
func RenderMailWelcomeContent(w io.Writer, p MailWelcomeContent) error
```

### 型名

**パターン:** `<GroupName><TemplateName>`

| テンプレートパス | 型名 | レンダー関数 |
|--------------|-----------|-----------------|
| `footer.tmpl` | `Footer` | `RenderFooter()` |
| `mail_invite/title.tmpl` | `MailInviteTitle` | `RenderMailInviteTitle()` |
| `mail_invite/content.tmpl` | `MailInviteContent` | `RenderMailInviteContent()` |
| `dashboard_summary/widget.tmpl` | `DashboardSummaryWidget` | `RenderDashboardSummaryWidget()` |

## 使用パターン

### 型安全なレンダリング（推奨）

生成された型安全な関数を使用：

```go
var buf bytes.Buffer

// グループ化されたテンプレートをレンダー
err := RenderMailInviteTitle(&buf, MailInviteTitle{
    SiteName:    "MyApp",
    InviterName: "太郎",
})

// フラットなテンプレートをレンダー
err = RenderSharedFooter(&buf, SharedFooter{
    Year:    2025,
    Company: "MyCompany",
})
```

**利点:**
- ✅ コンパイル時の型チェック
- ✅ IDEの自動補完
- ✅ エラーを早期に検出

### 動的レンダリング

`Template`名前空間で汎用の`Render()`関数を使用：

```go
var buf bytes.Buffer

// テンプレートを動的に選択
templateName := getTemplateFromConfig()

// 型安全なテンプレート名にTemplate名前空間を使用
err := Render(&buf, Template.MailInvite.Title, data)
```

**ユースケース:**
- 設定駆動のテンプレート選択
- ユーザー設定に基づく動的テンプレート切り替え
- 異なるテンプレートでのA/Bテスト

## 命名規則

### ディレクトリ命名

**パターン:** `lowercase_with_underscores`

```
templates/
├── mail_invite/              ✅ 良い
├── dashboard_summary/        ✅ 良い
├── user_profile/             ✅ 良い
├── MailInvite/               ❌ 避ける（PascalCase）
├── mail-invite/              ❌ 避ける（ハイフン）
```

### 数値プレフィックス

順序付けのための数値プレフィックスを使用（生成される名前から削除されます）：

```
templates/
├── 01_mail_invite/
│   └── title.tmpl
├── 02_mail_welcome/
│   └── title.tmpl
└── 03_mail_password_reset/
    └── title.tmpl
```

**生成される名前:**
- `MailInvite.Title`（01_が削除される）
- `MailWelcome.Title`（02_が削除される）
- `MailPasswordReset.Title`（03_が削除される）

### テンプレートファイル命名

各グループ内でシンプルで説明的な名前を使用：

```
mail_invite/
├── title.tmpl        ✅ シンプルで明確
├── content.tmpl      ✅ シンプルで明確
├── html.tmpl         ✅ シンプルで明確
```

冗長性を避ける（グループ名がすでにコンテキストを提供）：

```
mail_invite/
├── mail_invite_title.tmpl      ❌ 冗長
├── invite_email_content.tmpl   ❌ 冗長
```

## ベストプラクティス

### ✅ 推奨

**関連するテンプレートをグループ化:**
```
templates/
├── mail_invite/
│   ├── title.tmpl
│   ├── content.tmpl
│   └── footer.tmpl
```

**一貫した命名を使用:**
```
templates/
├── mail_invite/
├── mail_welcome/
└── mail_reset_password/
```

**必要に応じてフラットとグループを混在:**
```
templates/
├── shared_header.tmpl      # すべてのページで使用
├── shared_footer.tmpl      # すべてのページで使用
└── mail_invite/            # 招待に特化
    └── content.tmpl
```

**順序付けのための数値プレフィックスを使用:**
```
templates/
├── 01_header/
├── 02_nav/
└── 03_footer/
```

### ❌ 非推奨

**深さ1より深くネストしない:**
```
templates/
└── mail/
    └── invite/
        └── html.tmpl    ❌ スキャンされない（深度2）
```

**命名スタイルを混在させない:**
```
templates/
├── mail_invite/         # アンダースコア
├── mail-welcome/        # ハイフン
└── MailResetPassword/   # PascalCase
```

## 完全な例

### ディレクトリ構造

```
myproject/
├── gen.go
├── main.go
├── template_gen.go (生成される)
└── templates/
    ├── footer.tmpl
    ├── 01_mail_invite/
    │   ├── title.tmpl
    │   └── content.tmpl
    ├── 02_mail_account_created/
    │   ├── title.tmpl
    │   └── content.tmpl
    └── 03_mail_article_created/
        ├── title.tmpl
        └── content.tmpl
```

### コードを生成

**gen.go:**
```go
package main

//go:generate tmpltype -dir templates -pkg main -out template_gen.go
```

実行:
```bash
go generate
```

### 生成されたコードを使用

**main.go:**
```go
package main

import (
    "bytes"
    "fmt"
)

func main() {
    var buf bytes.Buffer

    // 型安全なレンダリング
    _ = RenderMailInviteTitle(&buf, MailInviteTitle{
        SiteName:    "MyApp",
        InviterName: "太郎",
    })
    fmt.Println("招待タイトル:", buf.String())

    buf.Reset()

    // 動的レンダリング
    _ = Render(&buf, Template.MailAccountCreated.Title, MailAccountCreatedTitle{
        SiteName: "MyApp",
        UserName: "次郎",
    })
    fmt.Println("アカウント作成タイトル:", buf.String())
}
```

### 完全な例を見る

動作する例を実行：

```bash
cd examples/07_grouping
go generate
go run .
```

この例では以下を示します：
- フラットとグループ化されたテンプレートの混在
- 順序付けのための数値プレフィックス
- 型安全と動的レンダリング
- ネストされた名前空間構造

## 関連項目

- [はじめに](getting-started.md) - チュートリアルと基本
- [CLIリファレンス](cli-reference.md) - コマンドラインオプション（ディレクトリスキャン）
- [テンプレート構文](template-syntax.md) - サポートされるテンプレート構文
- [サンプル03: Multi Template](../../examples/03_multi_template/) - 複数のフラットなテンプレート
- [サンプル07: Grouping](../../examples/07_grouping/) - 完全なグルーピング例
- [English version (詳細)](../template-grouping.md) - Complete reference in English
