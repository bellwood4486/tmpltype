package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bellwood4486/tmpltype/internal/gen"
)

func main() {
	dir := flag.String("dir", "", "template directory (required)")
	pkg := flag.String("pkg", "", "output package name (required)")
	out := flag.String("out", "", "output .go file path (required)")
	flag.Parse()

	if *dir == "" || *pkg == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: tmpltype -dir <directory> -pkg <name> -out <file>")
		os.Exit(2)
	}

	// ディレクトリの存在確認
	if _, err := os.Stat(*dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory not found: %s\n", *dir)
		os.Exit(1)
	}

	// テンプレートファイルをスキャン
	files, err := scanTemplateFiles(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("failed to scan directory: %w", err))
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no .tmpl files found in %s/\n", *dir)
		os.Exit(1)
	}

	// 複数のテンプレートを処理
	units := make([]gen.Unit, 0, len(files))
	outDir := filepath.Dir(*out)

	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("failed to read %s: %w", file, err))
			os.Exit(1)
		}

		// テンプレート名を抽出
		templateName, err := extractTemplateName(file, *dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("failed to extract template name from %s: %w", file, err))
			os.Exit(1)
		}

		// ファイルパスを計算（出力ディレクトリからの相対パス）
		relPath, err := filepath.Rel(outDir, file)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("failed to get relative path for %s: %w", file, err))
			os.Exit(1)
		}

		units = append(units, gen.Unit{
			TemplateName: templateName,
			Pkg:          *pkg,
			FilePath:     relPath,
			Source:       string(src),
		})
	}

	// コード生成
	result, err := gen.Emit(units)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("failed to emit: %w", err))
		os.Exit(1)
	}

	// 警告を出力
	for _, warning := range result.Warnings {
		fmt.Fprintln(os.Stderr, warning)
	}

	// メインファイルを書き込み
	if err := os.WriteFile(*out, []byte(result.MainCode), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// テンプレート文字列リテラルファイルを書き込み
	sourcesPath := generateSourcesPath(*out)
	if err := os.WriteFile(sourcesPath, []byte(result.SourcesCode), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// generateSourcesPath は出力ファイルパスからテンプレート文字列リテラルファイルのパスを生成する
// 例: "template_gen.go" -> "template_sources_gen.go"
// 例: "foo/bar_gen.go" -> "foo/bar_sources_gen.go"
func generateSourcesPath(outPath string) string {
	dir := filepath.Dir(outPath)
	base := filepath.Base(outPath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// "_gen" を "_sources_gen" に置き換える、なければ "_sources_gen" を追加
	var sourcesName string
	if strings.HasSuffix(nameWithoutExt, "_gen") {
		sourcesName = strings.TrimSuffix(nameWithoutExt, "_gen") + "_sources_gen"
	} else {
		sourcesName = nameWithoutExt + "_sources_gen"
	}

	return filepath.Join(dir, sourcesName+ext)
}

// scanTemplateFiles はディレクトリから.tmplファイルをスキャンする
// dir/*.tmpl (フラット) と dir/*/*.tmpl (グループ) のみを対象とする
func scanTemplateFiles(dir string) ([]string, error) {
	var files []string

	// フラットなテンプレート: dir/*.tmpl
	flatPattern := filepath.Join(dir, "*.tmpl")
	flatFiles, err := filepath.Glob(flatPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan flat templates: %w", err)
	}
	files = append(files, flatFiles...)

	// グループ化されたテンプレート: dir/*/*.tmpl (1階層のみ)
	groupPattern := filepath.Join(dir, "*", "*.tmpl")
	groupFiles, err := filepath.Glob(groupPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan grouped templates: %w", err)
	}
	files = append(files, groupFiles...)

	return files, nil
}

// extractTemplateName はファイルパスからテンプレート名を抽出する
// basedir からの相対パスでグループ判定を行う
// 例: basedir="templates", path="templates/footer.tmpl" -> "footer" (フラット)
// 例: basedir="templates", path="templates/email/welcome.tmpl" -> "email/welcome" (グループ)
func extractTemplateName(path string, basedir string) (string, error) {
	// basedir からの相対パスを取得
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	absBasedir, err := filepath.Abs(basedir)
	if err != nil {
		return "", err
	}

	relPath, err := filepath.Rel(absBasedir, absPath)
	if err != nil {
		return "", fmt.Errorf("path %s is not under basedir %s", path, basedir)
	}

	// 拡張子を削除
	pathWithoutExt := strings.TrimSuffix(relPath, filepath.Ext(relPath))

	// ディレクトリ区切りで分割
	parts := strings.Split(filepath.ToSlash(pathWithoutExt), "/")

	// 各パーツから数字プレフィックスを削除してクリーンアップ
	for i, part := range parts {
		parts[i] = cleanName(part)
	}

	// 階層チェック（フラット=1パーツ、グループ=2パーツ、それ以上はエラー）
	if len(parts) > 2 {
		return "", fmt.Errorf("template nesting too deep: %s (max 1 level of grouping)", relPath)
	}

	// パスとして結合
	return strings.Join(parts, "/"), nil
}

// cleanName は名前から数字プレフィックスを削除し、ハイフンをアンダースコアに変換する
func cleanName(name string) string {
	// 数字プレフィックスを削除（例: "01_header" -> "header", "1-mail" -> "mail"）
	re := regexp.MustCompile(`^\d+[-_]`)
	name = re.ReplaceAllString(name, "")

	// ハイフンをアンダースコアに変換
	name = strings.ReplaceAll(name, "-", "_")

	return name
}
