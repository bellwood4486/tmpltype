package scan

import (
	"fmt"
	"regexp"
	"text/template"
)

// Kind は推論されたフィールド種別を表します。
type Kind int

const (
	KindString Kind = iota
	KindStruct
	KindSlice
	KindMap
)

// Field は推論スキーマ木のノードです。
type Field struct {
	Name     string
	Kind     Kind
	Elem     *Field            // Slice/Map の要素
	Children map[string]*Field // Struct の子
}

// Schema はトップレベル（Params直下）のフィールド集合です。
type Schema struct {
	Fields map[string]*Field
}

// ScanTemplate は Go テンプレートを AST 解析して、.(ドット）スコープを追跡して
// フィールド参照からスキーマ木を推論します。
// 既定では葉はすべて string として扱い、 range は []struct{} (子フィールドがあれば) または []string (なければ),
// index は map[string]string を推論します。
func ScanTemplate(src string) (Schema, error) {
	insp, err := Inspect(src)
	if err != nil {
		return Schema{}, err
	}
	return BuildSchema(insp), nil
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
