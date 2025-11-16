了解！“写経しながら”着実に組み上げる前提で、**小さい成功を積み重ねる段階的プラン**にします。各フェーズに「達成条件（DoD）」「作業ステップ」「動作確認」を付けました。CI の“再生成差分で落とす”運用も最終段で導入します。

# 要点（ロードマップ）

1. リポ初期化・最小の骨組み（Go module / ディレクトリ構成）
2. MVP：テンプレを埋め込んで `Template()` + `RenderAny()` だけ（型生成なし）
3. AST 解析導入：`text/template/parse` でフィールド経路を厳密収集
4. 型推論（デフォルト string）とスキーマ木の構築（range→slice, index→map）
5. `@param` 上書き（型だけ）適用
6. コード生成：ネスト struct/type・`Params`・`Render`・`RenderWith` 出力
7. CLI：`gen` サブコマンド（`go:generate` で使えるところまで）
8. 例プロジェクト `examples/` と **e2e（go\:generate → go build）** テスト
9. CI：“生成物コミット + 再生成で差分検出”の整合チェック
10. 仕上げ（`--root-type`/`--emit-params`/`--no-render-typed`、README/仕様書整備）

---

# 詳細プラン

## フェーズ0：準備

**DoD**

* Go 1.22+、git、make（任意）が使える。

**手順**

* 新規リポ作成（例: `github.com/you/tmpltype`）
* `go mod init github.com/you/tmpltype`

---

## フェーズ1：骨組み（最小実行）

**DoD**

* `cmd/tmpltype` バイナリがビルドできる。
* `tmpltype gen --help` 相当の最低限パースが動く（まだ生成はしないでもOK）。

**手順**

* ディレクトリ：

  ```
  cmd/tmpltype/
  internal/{magic,scan,gen}/
  examples/{mailtpl/,main.go}
  ```
* `cmd/tmpltype/main.go` にフラグ（`--in --pkg --out`）だけ実装。
* とりあえず `Template()` + `RenderAny()` を吐くコード生成（固定文字列）で出力→コンパイル通ることを目標。

**確認**

```bash
go build ./cmd/templagen
templagen gen --in examples/mailtpl/email.tmpl --pkg mailtpl --out examples/mailtpl/params_gen.go
go build ./examples
```

---

## フェーズ2：MVP（テンプレ埋め込み + RenderAny）

**DoD**

* 生成された `.go` に `//go:embed <tmpl>`、`Template()`、`RenderAny()` がある。
* `go run ./examples` でテンプレが実行される。

**手順**

* `gen.Emit` の最小版：preamble + imports + `//go:embed` + `Template()` + `RenderAny()`
* examples 側で `map[string]any` を渡して動作確認。

**確認**

* サンプルテンプレを簡単な `{{ .User.Name }}` 程度で確認。

---

## フェーズ3：AST 解析（厳密スコープ）

**DoD**

* `scan.ScanTemplate(src)` が `text/template/parse` を使って `.User.Name` / `range .Items` / `with .User` / `index .Meta "k"` を正しく検出し、**フィールド経路**を収集できる。

**手順**

* `internal/scan/scan_ast.go` を実装：

  * `List/Action/If/With/Range` を DFS。
  * `range .Items` → `Items` を配列扱い、ブロック内の `.Title` 等は要素側へ。
  * `index .Meta "k"` → `Meta` を `map[string]string` 扱い。
  * 結果はスキーマ木（struct/slice/map/leaf=string）に落とす。

**確認**

* 単体テスト or `fmt.Printf` デバッグで `Schema` の構造が期待通りか確認。

---

## フェーズ4：デフォルト string の型推論

**DoD**

* 参照された各フィールドが未指定なら `string`。
* `range` から `[]struct{...}`、`index` から `map[string]string` が入る。

**手順**

* フェーズ3のスキーマ木に `KindString/Struct/Slice/Map/Ptr` と `Children/Elem` を持たせる。
* 参照のたびに `ensureStruct/ensurePath/markSliceStruct/markMapString` で木を拡張。

**確認**

* スキーマ木のダンプで `User{Name string}`、`Items []{Title string}` 等が見える。

---

## フェーズ5：`@param` 上書き

**DoD**

* `{{/* @param Path Type */}}` を複数拾って **最終的な型**に上書きできる。
* サポート：`string/int/int64/float64/bool/time.Time`、`[]T`、`map[string]T`（keyは string 固定）、`*T`、`struct{...}`。

**手順**

* `internal/magic.ParseParams` で正規表現抽出（Path→Type）。
* 簡易型パーサでスキーマ木に上書き（`applyOverride`）。

**確認**

* `@param User.Age int`、`@param Items []struct{ ID int64; Title string }` などで木が変わること。

---

## フェーズ6：フルコード生成

**DoD**

* 生成物に以下が含まれる：

  * `type Params struct { ... }`（トップレベル）
  * ネスト struct/type（要素 struct は自動命名）
  * `Template()` / `Render(Params)` / `RenderAny()` / `RenderWith[T]()`
* `examples` を **型安全な `Params` で実行**できる。

**手順**

* スキーマ木から Go 型文字列を作る `goTypeOf`。
* 命名規則（エクスポート化、要素名 `Item`/`Value` など）を固定し、**安定した順序**で出力。
* `time.Time` を含む場合に `import "time"` を自動追加。

**確認**

* `go generate ./...` → `go build` が通る。
* `examples/main.go` で `mailtpl.Params{...}` を渡して動作。

---

## フェーズ7：CLI（`gen`）の実用化

**DoD**

* `//go:generate templagen gen -in ... -pkg ... -out ...` が動く。
* 生成物をコミットする運用に移行。

**手順**

* `cmd/templagen/main.go` で `--emit-params/--root-type/--no-render-typed` をフラグ化（後続で）。
* まずは `gen` の基本フロー固定：読み込み→解析→生成→書き出し。

**確認**

```bash
(cd examples/mailtpl && go generate ./...)
go build ./examples
```

---

## フェーズ8：examples + e2e テスト

**DoD**

* `go test ./examples -v` で「`go generate` → `go build`」が自動検証される。

**手順**

* `examples/mailtpl/gen.go` に `//go:generate` を記述。
* `examples/e2e_test.go` を作成：`exec.Command` で `go generate` と `go build` を呼ぶ。

**確認**

```bash
go test ./examples -v
```

---

## フェーズ9：CI（差分検出で落とす）

**DoD**

* PR / push 時に「生成物が最新か」を CI が判定し、差分があれば失敗。

**手順**

* スクリプト `scripts/verify_generated.sh`：

  1. ワークツリーがクリーン前提チェック（任意）
  2. `go generate ./...`
  3. `gofmt -s -w .` / `go mod tidy`
  4. `git diff --quiet` で差分検出 → 差分あれば失敗
* GitHub Actions などに組み込み（Go セットアップ → スクリプト実行）。

**確認**

* わざとテンプレを修正して CI が落ちることを確認。

---

## フェーズ10：仕上げ＆拡張

**DoD**

* 開発者ガイドと CLI マニュアル、仕様（`spec.md`/`impl.md`/`cli.md`/`examples.md`）がリポに揃う。
* 任意の追加機能：

  * `--root-type`（外部型エイリアス）
  * `--emit-params=false`（自前型前提・`RenderAny/With` のみ）
  * `--no-render-typed`
  * `templagen check`（必要なら。今回は CI 差分方式が主）

**手順**

* README に「思想／使い方／制限」を簡潔に集約。
* セマンティックバージョニング、タグ付け（`v0.x`）。

---

# 進め方のコツ（写経モード向け）

* **1コミット1小目標**：たとえば「AST で `.User.Name` だけ拾えるようにした」など、粒度を小さく。
* **ダンプを多用**：スキーマ木を `fmt.Printf("%#v\n", tree)` で可視化し、期待と比較。
* **順序の安定化**：生成時は**必ず sort**。import・フィールド・型定義の順序ブレは CI 地獄の元。
* **テンプレは段階的に複雑化**：プレーン→`with`→`range`→`index`→`@param`→ネスト…の順に難度を上げて確認。

---

# 完了の定義（Definition of Done）

* `go generate ./...` → `go build ./...` が問題なく通る。
* `go test ./examples -v` が緑。
* CI の生成差分チェックが動き、差分で確実に落ちる。
* `README` を読めば、非エンジニア／エンジニア双方が使い方に迷わない。

---

必要なら、このプランに沿って\*\*コミット単位のタスクリスト（チェックボックス）\*\*も用意します。

---

# 今後の課題と優先順位（Future Enhancements）

このセクションは、現在の実装で未対応のGo template機能について、実用性と影響度の観点から優先順位付けを行ったものです。

## 優先順位の基準

- **影響度**: ツールの実用性に与える影響の大きさ（★1〜5）
- **実用性**: 実際のGoテンプレート利用パターンでの使用頻度（★1〜5）
- **実装難易度**: 低・中・高で評価

---

## 🔴 最優先 (Critical) - 実用上の大きな制約になっている

### 1. カスタム関数サポート (Custom FuncMap)

**影響度**: ★★★★★
**実用性**: ★★★★★
**実装難易度**: 中

#### 現状の問題

```go
// ❌ 現状: ユーザー定義関数を一切使えない
{{/* templates/email.tmpl */}}
<h1>{{ .Title | upper }}</h1>
<p>Created: {{ formatDate .CreatedAt }}</p>
<p>{{ nl2br .Message }}</p>
```

- ユーザー定義関数が使えず、標準関数以外を使うとランタイムエラー
- Goテンプレートの最も一般的なパターンが使えない
- 既存プロジェクトのテンプレートをそのまま移行できない

#### 必要な対応

1. **CLI オプション追加**
   ```bash
   tmpltype -dir templates -pkg main -out gen.go -funcs GetCustomFuncs
   ```

2. **生成コード修正** (internal/gen/emit.go:415)
   ```go
   func newTemplate(name TemplateName, source string) *template.Template {
       t := template.New(string(name)).Funcs(GetCustomFuncs())
       return template.Must(t.Option("missingkey=error").Parse(source))
   }
   ```

3. **テンプレートスキャン時の対応** (internal/scan/scan.go:51)
   ```go
   // ダミー関数マップで未定義エラーを回避
   tmpl, err := template.New("tpl").Funcs(dummyFuncMap()).Parse(src)
   ```

#### 実装ステップ

- [ ] CLI に `-funcs` フラグを追加
- [ ] 生成コードで関数マップを適用するロジックを実装
- [ ] スキャン時にダミー関数マップを使用して解析エラーを回避
- [ ] examples に custom functions を使うサンプルを追加
- [ ] ドキュメント更新

---

### 2. 変数の代入と参照 (Variable Assignments)

**影響度**: ★★★★☆
**実用性**: ★★★★★
**実装難易度**: 高

#### 現状の問題

```go
// ❌ 現状: 変数経由のアクセスが型推論されない
{{ $user := .CurrentUser }}
{{ $user.Name }} - {{ $user.Email }}

{{ range $i, $item := .Products }}
  {{ $i }}: {{ $item.Price }}
{{ end }}
```

- 変数を使った参照が型推論に反映されない
- rangeのインデックス・要素変数が追跡されない
- 複雑なテンプレートで必須のパターン

#### 必要な対応

1. **スコープ管理の実装**
   - `scan.go`の`ctx`構造体に変数スコープマップを追加
   - 変数のライフタイム（スコープの開始・終了）を追跡

2. **AST Walkerの拡張**
   - `VariableNode`の処理を追加
   - 変数への代入（`:=`）と参照を解決
   - `$var.Field`のようなパス解決

#### 実装ステップ

- [ ] ctx構造体に変数スコープ管理を追加
- [ ] VariableNodeの処理ロジックを実装
- [ ] range構文での変数代入を追跡
- [ ] 変数経由のフィールドアクセス解析
- [ ] テストケースとサンプルを追加

#### 参考

- 現在の制限事項: docs/template-syntax.md:421-428
- 関連コード: internal/scan/scan.go

---

## 🟡 高優先度 (High) - よく使われるが回避策がある

### 3. template/define/block サポート

**影響度**: ★★★★☆
**実用性**: ★★★☆☆
**実装難易度**: 中〜高

#### 現状の問題

```go
// ❌ 現状: サブテンプレートのフィールドが推論されない
{{/* templates/page.tmpl */}}
{{ define "header" }}
  <h1>{{ .PageTitle }}</h1>
{{ end }}

{{ template "header" . }}
<main>{{ .Content }}</main>
```

- `define`内のフィールド参照が親スキーマに反映されない
- `template`や`block`の呼び出しが追跡されない
- マルチテンプレート構成で型安全性が失われる

#### 現在の回避策

各テンプレートを別ファイルに分割してフラット構造で管理する（template grouping機能を使用）

#### 必要な対応

1. **AST Walkerの拡張**
   - `walk`関数に`DefineNode`, `TemplateNode`, `BlockNode`のケース追加

2. **テンプレート間の依存関係解決**
   - サブテンプレートの名前と定義を管理
   - 呼び出し時のデータフロー追跡

#### 実装ステップ

- [ ] DefineNode/TemplateNode/BlockNodeの処理を追加
- [ ] サブテンプレート定義の収集
- [ ] テンプレート呼び出し時のデータ伝播解析
- [ ] 単一ファイル内での複数テンプレート定義のサポート
- [ ] サンプルとドキュメント追加

---

### 4. index関数の正確な型推論 (スライス vs マップ)

**影響度**: ★★★☆☆
**実用性**: ★★★★☆
**実装難易度**: 低〜中

#### 現状の問題

```go
// ❌ 現状: 常に map[string]string と推論される
{{ index .Items 0 }}      {{/* 実際は []T だが map として推論 */}}
{{ index .Matrix 1 2 }}   {{/* 多次元アクセスも未対応 */}}
```

#### 問題箇所

internal/scan/scan.go:136-140
```go
if id, ok := cmd.Args[0].(*tplparse.IdentifierNode); ok && id.Ident == "index" {
    if fn, ok := cmd.Args[1].(*tplparse.FieldNode); ok {
        markMapString(s, append(c.dot, fn.Ident...))  // 常にマップ
    }
}
```

#### 必要な対応

1. **引数の型判定**
   - 第2引数が`StringNode`なら`map[string]T`
   - 第2引数が`NumberNode`なら`[]T`
   - 既存のスライス定義を尊重

2. **多次元アクセスの考慮**
   - `index .Matrix 1 2`のようなケースの処理

#### 実装ステップ

- [ ] index関数の引数型を判定するロジックを追加
- [ ] StringNode/NumberNodeによる分岐処理
- [ ] スライスの場合の型推論ロジック
- [ ] 多次元アクセスの処理（将来）
- [ ] テストケース追加

---

## 🟢 中優先度 (Medium) - あると便利だが使用頻度は低い

### 5. パイプライン型推論の強化

**影響度**: ★★☆☆☆
**実用性**: ★★★☆☆
**実装難易度**: 高

#### 現状の問題

```go
// ⚠️ 現状: .Title は推論されるが、後続が不正確
{{ .Title | printf "Title: %s" }}
{{ .Items | len }}  {{/* len の結果が int だが追跡されない */}}
```

- 関数の戻り値型が考慮されない
- パイプライン後のフィールドアクセスが失われる

#### 必要な対応

1. ビルトイン関数のシグネチャマップ
2. パイプライン結果の型伝播
3. カスタム関数のシグネチャ指定（将来）

#### 実装ステップ

- [ ] ビルトイン関数の型シグネチャ定義
- [ ] パイプライン結果の型追跡ロジック
- [ ] 型システムの拡張検討
- [ ] テストケースとドキュメント

---

### 6. 論理・比較演算子の型推論

**影響度**: ★★☆☆☆
**実用性**: ★★☆☆☆
**実装難易度**: 中

#### 現状の問題

```go
// ⚠️ 現状: bool型として扱うべきフィールドが string 推論
{{ if and .User.IsActive .User.IsPremium }}
{{ if eq .Status "published" }}
```

- `and`/`or`/`not`の引数がboolと推論されない
- `eq`/`ne`等の比較結果が型に反映されない

#### 必要な対応

1. 演算子コンテキストでの型推論
2. `@param IsActive bool`を推奨する警告

#### 実装ステップ

- [ ] 論理演算子のコンテキスト検出
- [ ] bool型の推論ルール追加
- [ ] 型不一致時の警告メッセージ
- [ ] ドキュメント更新

---

### 7. break/continue のサポート

**影響度**: ★☆☆☆☆
**実用性**: ★★☆☆☆
**実装難易度**: 低

#### 現状の問題

```go
// ⚠️ 現状: 動作はするが AST walker が認識しない
{{ range .Items }}
  {{ if .Skip }}{{ continue }}{{ end }}
  {{ .Name }}
{{ end }}
```

- 制御フローの早期終了が追跡されない
- 将来的な最適化の障害になる可能性

#### 必要な対応

`walk`関数に`BreakNode`/`ContinueNode`のケースを追加（認識するだけでOK）

#### 実装ステップ

- [ ] BreakNode/ContinueNodeのケース追加
- [ ] 将来の最適化に向けた準備
- [ ] テストケース追加

---

## 🔵 低優先度 (Low) - ニッチなユースケース

### 8. カスタムデリミタ

**影響度**: ★☆☆☆☆
**実用性**: ★☆☆☆☆
**実装難易度**: 低

#### 現状の問題

```go
// ❌ 現状: {{ }} 以外は未対応
<<- .Field ->>  {{/* Delims("<<", ">>") */}}
```

#### 必要な対応

- CLI オプション: `-left-delim`, `-right-delim`
- スキャン時のデリミタ設定

---

### 9. call関数のサポート

**影響度**: ★☆☆☆☆
**実用性**: ★☆☆☆☆
**実装難易度**: 高

#### 現状の問題

```go
// ❌ 現状: メソッド呼び出しが追跡されない
{{ call .GetTitle }}
{{ .Method arg1 arg2 }}
```

#### 必要な対応

- メソッドシグネチャの型推論
- 引数と戻り値の追跡
- リフレクションベースの型解析

---

## 📊 優先順位マトリックス

```
影響度×実用性マトリックス:

     実用性 →
影響度  低    中    高    最高
  ↓  ┌─────┬─────┬─────┬─────┐
最高 │  -  │  -  │  -  │ [1] │ 1. カスタム関数
  ↑  ├─────┼─────┼─────┼─────┤
  高 │  -  │  -  │ [3] │ [2] │ 2. 変数追跡
     ├─────┼─────┼─────┼─────┤ 3. template/define
  中 │ [8] │ [6] │ [4] │ [5] │ 4. index正確化
     ├─────┼─────┼─────┼─────┤ 5. パイプライン
  低 │ [9] │ [7] │  -  │  -  │ 6. bool型推論
     └─────┴─────┴─────┴─────┘ 7. break/continue
                                 8. カスタムデリミタ
                                 9. call関数
```

---

## 🎯 推奨実装ロードマップ

### Phase 1: 実用性の向上
**目標**: 既存プロジェクトでの採用障壁を下げる

1. **カスタム関数サポート** (Issue #TBD)
   - 最も影響が大きく、既存テンプレートの移行に必須
   - 実装難易度も中程度で取り組みやすい

2. **index関数の正確化** (Issue #TBD)
   - 実装コストが低く効果が高い
   - 早期に対応可能

### Phase 2: 高度な型推論
**目標**: 複雑なテンプレートパターンへの対応

3. **変数追跡** (Issue #TBD)
   - 実装は複雑だが価値が高い
   - 段階的なアプローチで実装可能

4. **template/define サポート** (Issue #TBD)
   - 大規模プロジェクト対応
   - 現在のgrouping機能との統合を検討

### Phase 3: エッジケース対応
**目標**: 完全性の向上

5. パイプライン型推論
6. bool型推論と警告
7. break/continue認識

### 将来検討 (Need-driven)
**目標**: ユーザーからの要望に応じて検討

8. カスタムデリミタ
9. call関数

---

## 📝 関連ドキュメント

- 現在の制限事項: [docs/template-syntax.md](./template-syntax.md#limitations)
- Go template仕様: https://pkg.go.dev/text/template
- 実装箇所:
  - スキャナー: `internal/scan/scan.go`
  - コード生成: `internal/gen/emit.go`
  - 型推論: `internal/typing/`

---

## 🤝 コントリビューション

これらの機能実装に興味がある方は、各Issueでディスカッションに参加してください。
実装の前に設計の議論を行い、段階的なアプローチで進めることを推奨します。
