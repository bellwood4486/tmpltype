# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

tmpltype is a Go code generator that creates type-safe template rendering functions from Go template files. It analyzes `.tmpl` files and generates Go structs and render functions, enabling compile-time type safety for template parameters.

## Build and Test Commands

```bash
# Run all tests
go test ./...

# Run a specific test
go test ./internal/gen -run TestEmit_BasicScaffoldAndTypes

# Build the CLI
go build ./cmd/tmpltype

# Install the CLI
go install ./cmd/tmpltype

# Run code generation in examples
cd examples/01_basic && go generate && go run .
```

## Architecture

```mermaid
flowchart LR
    subgraph Input
        TMPL["*.tmpl files"]
    end

    subgraph cmd["cmd/tmpltype"]
        CMD_MAIN["flags解析<br/>ファイル走査"]
    end

    subgraph scan["internal/scan"]
        SCAN_MAIN["AST解析<br/>スコープ追跡<br/>種別(Kind)推論"]
        KIND["Kind:<br/>String | Struct<br/>Slice | Map"]
    end

    subgraph typing["internal/typing"]
        TYPE_MAIN["Kind→Go型変換<br/>@paramオーバーライド<br/>名前付き型抽出<br/>import収集"]
        subgraph magic["magic"]
            MAGIC["@paramパース<br/>型表現解析"]
        end
    end

    subgraph gen["internal/gen"]
        GEN_MAIN["構造体生成<br/>Render関数<br/>テンプレートソース出力"]
    end

    subgraph Output
        OUT1["*_gen.go<br/>(型定義, Render)"]
        OUT2["*_sources_gen.go<br/>(テンプレート文字列)"]
    end

    TMPL --> CMD_MAIN
    CMD_MAIN -->|"*.tmpl"| SCAN_MAIN
    SCAN_MAIN --> KIND
    SCAN_MAIN -->|"Schema"| TYPE_MAIN
    TYPE_MAIN --> MAGIC
    TYPE_MAIN -->|"TypedSchema"| GEN_MAIN
    GEN_MAIN --> OUT1
    GEN_MAIN --> OUT2
```

The code generation pipeline flows through four internal packages:

1. **cmd/tmpltype** - CLI entry point
   - Parses flags (`-dir`, `-pkg`, `-out`)
   - Scans template files (supports flat `dir/*.tmpl` and grouped `dir/*/*.tmpl`)
   - Invokes the generation pipeline

2. **internal/scan** - Template AST analysis
   - Parses Go templates and walks the AST
   - Tracks dot (`.`) scope through `with`, `range`, `if` blocks
   - Infers schema: leaf fields → `string`, `range` → `[]struct{}`, `index` → `map[string]string`
   - Handles unknown custom functions by dynamically adding dummies during parse

3. **internal/typing** - Type resolution
   - Applies default type inference from scan results
   - Processes `@param` directive overrides (via `internal/typing/magic`)
   - Extracts named types for struct generation
   - Collects required imports (e.g., `time` for `time.Time`)

4. **internal/gen** - Code emission
   - Generates two output files:
     - Main file: type definitions, `InitTemplates()`, `Render*()` functions
     - Sources file: template string literals
   - Supports template grouping (subdirectories become nested namespaces)
   - Uses functional options pattern (`WithFuncs`) for custom template functions

## Key Patterns

- **@param directive**: Use `{{/* @param Path.To.Field type */}}` in templates to override inferred types
- **Template grouping**: Templates in `templates/email/*.tmpl` generate `Template.Email.Welcome` style namespaces
- **Generated files**: Output includes `*_gen.go` (types/functions) and `*_sources_gen.go` (template strings)
