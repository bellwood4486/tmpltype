package scan

import (
	"fmt"
	"regexp"
	"text/template"
	"text/template/parse"
)

// usage はフィールドがテンプレート内でどのように使われたかを表します。
type usage int

const (
	// usageLeaf は {{ .Foo }} のように葉として参照された場合
	usageLeaf usage = iota
	// usageRange は {{ range .Foo }} の対象になった場合
	usageRange
	// usageRangeMap は {{ range $k, $v := .Foo }} の対象になった場合
	usageRangeMap
	// usageIndex は {{ index .Foo "key" }} の対象になった場合
	usageIndex
	// usageScope は {{ with .Foo }} や {{ if .Foo }} のスコープ基点になった場合
	usageScope
)

// fieldRef はテンプレート内でのフィールド参照を表します。
type fieldRef struct {
	path  []string // スコープ解決済みの絶対パス
	usage usage
}

// inspection はテンプレートを検査した結果です。
type inspection struct {
	refs []fieldRef
}

// inspectCtx は検査中のドットスコープを追跡します。
type inspectCtx struct {
	dot []string
}

func (c inspectCtx) with(prefix []string) inspectCtx {
	dup := make([]string, len(c.dot))
	copy(dup, c.dot)
	return inspectCtx{dot: append(dup, prefix...)}
}

// inspect はテンプレートを検査してフィールド参照を収集します。
// この段階では型の決定は行わず、どのフィールドがどのように使われたかのみを記録します。
func inspect(src string) (inspection, error) {
	tmpl, err := parseTemplateWithDynamicFuncs(src)
	if err != nil {
		return inspection{}, fmt.Errorf("failed to parse template: %w", err)
	}
	if tmpl.Tree == nil || tmpl.Tree.Root == nil {
		return inspection{}, fmt.Errorf("template not found: %s", "tpl")
	}

	var refs []fieldRef
	collectRefs(tmpl.Tree.Root, &refs, inspectCtx{})
	return inspection{refs: refs}, nil
}

// collectRefs はテンプレート AST を DFS して全フィールド参照を収集します。
func collectRefs(n parse.Node, refs *[]fieldRef, c inspectCtx) {
	switch x := n.(type) {
	case *parse.ListNode:
		for _, nn := range x.Nodes {
			collectRefs(nn, refs, c)
		}

	case *parse.ActionNode:
		collectFromPipeRefs(x.Pipe, refs, c, usageLeaf)

	case *parse.IfNode:
		// if のパイプに出るフィールドは存在チェック用途（スコープ基点）
		base := baseFieldFromPipeNode(x.Pipe)
		if len(base) > 0 {
			*refs = append(*refs, fieldRef{
				path:  append(c.dot, base...),
				usage: usageScope,
			})
		}
		collectFromPipeRefs(x.Pipe, refs, c, usageLeaf)
		if x.List != nil {
			collectRefs(x.List, refs, c)
		}
		if x.ElseList != nil {
			collectRefs(x.ElseList, refs, c)
		}

	case *parse.WithNode:
		// with では基点フィールドがスコープ基点になる
		base := baseFieldFromPipeNode(x.Pipe)
		if len(base) > 0 {
			*refs = append(*refs, fieldRef{
				path:  append(c.dot, base...),
				usage: usageScope,
			})
		}
		nc := c
		if len(base) > 0 {
			nc = c.with(base)
		}
		if x.List != nil {
			collectRefs(x.List, refs, nc)
		}
		if x.ElseList != nil {
			collectRefs(x.ElseList, refs, c)
		}

	case *parse.RangeNode:
		base := baseFieldFromPipeNode(x.Pipe)
		if len(base) > 0 {
			// 2変数なら map、1変数/0変数なら slice
			if len(x.Pipe.Decl) == 2 {
				*refs = append(*refs, fieldRef{
					path:  append(c.dot, base...),
					usage: usageRangeMap,
				})
			} else {
				*refs = append(*refs, fieldRef{
					path:  append(c.dot, base...),
					usage: usageRange,
				})
			}
		}
		nc := c
		if len(base) > 0 {
			nc = c.with(base)
		}
		if x.List != nil {
			collectRefs(x.List, refs, nc)
		}
		if x.ElseList != nil {
			collectRefs(x.ElseList, refs, c)
		}
	}
}

// collectFromPipeRefs はパイプ内のフィールド参照を収集します。
func collectFromPipeRefs(p *parse.PipeNode, refs *[]fieldRef, c inspectCtx, defaultUsage usage) {
	if p == nil {
		return
	}

	for _, cmd := range p.Cmds {
		// index .Meta "key" → Meta は usageIndex
		if len(cmd.Args) >= 2 {
			if id, ok := cmd.Args[0].(*parse.IdentifierNode); ok && id.Ident == "index" {
				if fn, ok := cmd.Args[1].(*parse.FieldNode); ok {
					*refs = append(*refs, fieldRef{
						path:  append(c.dot, fn.Ident...),
						usage: usageIndex,
					})
				}
			}
		}
		// 通常のフィールド参照
		for _, a := range cmd.Args {
			if f, ok := a.(*parse.FieldNode); ok {
				*refs = append(*refs, fieldRef{
					path:  append(c.dot, f.Ident...),
					usage: defaultUsage,
				})
			}
		}
	}
}

// baseFieldFromPipeNode はパイプ内で最初に現れるフィールドノードの識別子スライスを返します。
func baseFieldFromPipeNode(p *parse.PipeNode) []string {
	if p == nil {
		return nil
	}

	for _, cmd := range p.Cmds {
		for _, a := range cmd.Args {
			if f, ok := a.(*parse.FieldNode); ok && len(f.Ident) > 0 {
				return f.Ident
			}
		}
	}

	return nil
}

// parseTemplateWithDynamicFuncs はテンプレートをパースし、未定義関数があれば動的にダミー関数を追加してリトライします。
func parseTemplateWithDynamicFuncs(src string) (*template.Template, error) {
	funcs := dummyFuncMap()
	var tmpl *template.Template
	var err error

	// 未定義関数エラーの場合、関数名を抽出してダミー関数を追加し、最大10回リトライ
	maxRetries := 10
	for i := 0; i <= maxRetries; i++ {
		tmpl, err = template.New("tpl").Funcs(funcs).Parse(src)
		if err == nil {
			break
		}

		funcName := extractUndefinedFuncName(err.Error())
		if funcName == "" {
			// 未定義関数エラーではない、または関数名が抽出できない
			break
		}

		if i == maxRetries {
			// 最大リトライ回数に到達
			return nil, fmt.Errorf("exceeded max retries (%d) while resolving undefined functions: %w", maxRetries, err)
		}

		// ダミー関数を追加して次回リトライ
		funcs[funcName] = dummyFunc
	}

	return tmpl, err
}

// dummyFunc は未定義カスタム関数の代替として使用されるダミー関数です。
// 任意の引数を受け取り、空文字列を返します。
var dummyFunc = func(args ...interface{}) interface{} {
	return ""
}

// dummyFuncMap は未定義カスタム関数によるパースエラーを回避するためのダミー関数マップを返します。
// スキャン時は AST を取得するだけで実際に関数を実行しないため、ダミー実装で十分です。
func dummyFuncMap() template.FuncMap {
	// よく使われるカスタム関数名をプリセット
	return template.FuncMap{
		"upper":          dummyFunc,
		"lower":          dummyFunc,
		"title":          dummyFunc,
		"trim":           dummyFunc,
		"trimSpace":      dummyFunc,
		"formatDate":     dummyFunc,
		"formatDateTime": dummyFunc,
		"formatTime":     dummyFunc,
		"nl2br":          dummyFunc,
		"default":        dummyFunc,
		"join":           dummyFunc,
		"split":          dummyFunc,
		"add":            dummyFunc,
		"sub":            dummyFunc,
		"mul":            dummyFunc,
		"div":            dummyFunc,
		"mod":            dummyFunc,
		"comma":          dummyFunc,
		"json":           dummyFunc,
		"yaml":           dummyFunc,
		"base64":         dummyFunc,
		"urlEncode":      dummyFunc,
		"urlDecode":      dummyFunc,
		"htmlEscape":     dummyFunc,
		"htmlUnescape":   dummyFunc,
		"contains":       dummyFunc,
		"hasPrefix":      dummyFunc,
		"hasSuffix":      dummyFunc,
		"replace":        dummyFunc,
		"repeat":         dummyFunc,
		"reverse":        dummyFunc,
		"truncate":       dummyFunc,
	}
}

// undefinedFuncPattern は未定義関数エラーから関数名を抽出するための正規表現です。
var undefinedFuncPattern = regexp.MustCompile(`function "([^"]+)" not defined`)

// extractUndefinedFuncName はエラーメッセージから未定義関数名を抽出します。
// エラーメッセージ形式: template: tpl:1: function "functionName" not defined
func extractUndefinedFuncName(errMsg string) string {
	matches := undefinedFuncPattern.FindStringSubmatch(errMsg)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
