package scan

import (
	"fmt"
	"text/template/parse"
)

// Usage はフィールドがテンプレート内でどのように使われたかを表します。
type Usage int

const (
	// UsageLeaf は {{ .Foo }} のように葉として参照された場合
	UsageLeaf Usage = iota
	// UsageRange は {{ range .Foo }} の対象になった場合
	UsageRange
	// UsageRangeMap は {{ range $k, $v := .Foo }} の対象になった場合
	UsageRangeMap
	// UsageIndex は {{ index .Foo "key" }} の対象になった場合
	UsageIndex
	// UsageScope は {{ with .Foo }} や {{ if .Foo }} のスコープ基点になった場合
	UsageScope
)

// FieldRef はテンプレート内でのフィールド参照を表します。
type FieldRef struct {
	Path  []string // スコープ解決済みの絶対パス
	Usage Usage
}

// Inspection はテンプレートを検査した結果です。
type Inspection struct {
	Refs []FieldRef
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

// Inspect はテンプレートを検査してフィールド参照を収集します。
// この段階では型の決定は行わず、どのフィールドがどのように使われたかのみを記録します。
func Inspect(src string) (Inspection, error) {
	tmpl, err := parseTemplateWithDynamicFuncs(src)
	if err != nil {
		return Inspection{}, fmt.Errorf("failed to parse template: %w", err)
	}
	if tmpl.Tree == nil || tmpl.Tree.Root == nil {
		return Inspection{}, fmt.Errorf("template not found: %s", "tpl")
	}

	var refs []FieldRef
	collectRefs(tmpl.Tree.Root, &refs, inspectCtx{})
	return Inspection{Refs: refs}, nil
}

// collectRefs はテンプレート AST を DFS して全フィールド参照を収集します。
func collectRefs(n parse.Node, refs *[]FieldRef, c inspectCtx) {
	switch x := n.(type) {
	case *parse.ListNode:
		for _, nn := range x.Nodes {
			collectRefs(nn, refs, c)
		}

	case *parse.ActionNode:
		collectFromPipeRefs(x.Pipe, refs, c, UsageLeaf)

	case *parse.IfNode:
		// if のパイプに出るフィールドは存在チェック用途（スコープ基点）
		base := baseFieldFromPipeNode(x.Pipe)
		if len(base) > 0 {
			*refs = append(*refs, FieldRef{
				Path:  append(c.dot, base...),
				Usage: UsageScope,
			})
		}
		collectFromPipeRefs(x.Pipe, refs, c, UsageLeaf)
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
			*refs = append(*refs, FieldRef{
				Path:  append(c.dot, base...),
				Usage: UsageScope,
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
				*refs = append(*refs, FieldRef{
					Path:  append(c.dot, base...),
					Usage: UsageRangeMap,
				})
			} else {
				*refs = append(*refs, FieldRef{
					Path:  append(c.dot, base...),
					Usage: UsageRange,
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
func collectFromPipeRefs(p *parse.PipeNode, refs *[]FieldRef, c inspectCtx, defaultUsage Usage) {
	if p == nil {
		return
	}

	for _, cmd := range p.Cmds {
		// index .Meta "key" → Meta は UsageIndex
		if len(cmd.Args) >= 2 {
			if id, ok := cmd.Args[0].(*parse.IdentifierNode); ok && id.Ident == "index" {
				if fn, ok := cmd.Args[1].(*parse.FieldNode); ok {
					*refs = append(*refs, FieldRef{
						Path:  append(c.dot, fn.Ident...),
						Usage: UsageIndex,
					})
				}
			}
		}
		// 通常のフィールド参照
		for _, a := range cmd.Args {
			if f, ok := a.(*parse.FieldNode); ok {
				*refs = append(*refs, FieldRef{
					Path:  append(c.dot, f.Ident...),
					Usage: defaultUsage,
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
