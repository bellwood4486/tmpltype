# CLI Reference

Complete command-line reference for the `tmpltype` tool.

## Table of Contents

- [Synopsis](#synopsis)
- [Options](#options)
- [Usage Examples](#usage-examples)
- [Directory Scanning Behavior](#directory-scanning-behavior)
- [Integration with go generate](#integration-with-go-generate)
- [Troubleshooting](#troubleshooting)

## Synopsis

```bash
tmpltype -dir <directory> -pkg <name> -out <file>
```

Generate type-safe Go code from template files in the specified directory.

## Options

### `-dir` (required)

**Type:** `string`
**Description:** Template directory to scan

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

**Scanning Behavior:**
- Scans `<dir>/*.tmpl` (flat templates in the root)
- Scans `<dir>/*/*.tmpl` (grouped templates, 1 level deep)
- Does **not** scan deeper than 1 level of subdirectories

**Examples:**

```bash
# Scans templates/*.tmpl
tmpltype -dir templates -pkg main -out gen.go

# Scans ./web/templates/*.tmpl and ./web/templates/*/*.tmpl
tmpltype -dir ./web/templates -pkg web -out templates_gen.go

# Relative paths work
tmpltype -dir ../shared/templates -pkg shared -out gen.go
```

### `-pkg` (required)

**Type:** `string`
**Description:** Output package name

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

The generated code will have `package <name>` at the top.

**Examples:**

```bash
# package main
tmpltype -dir templates -pkg main -out gen.go

# package templates
tmpltype -dir tpl -pkg templates -out templates.go

# package myapp
tmpltype -dir views -pkg myapp -out views_gen.go
```

### `-out` (required)

**Type:** `string`
**Description:** Output file path

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

**Best Practices:**
- Use the `_gen.go` suffix to indicate generated code
- Place the output file in the same package directory
- Add to `.gitignore` if you don't want to commit generated code (not recommended)

**Examples:**

```bash
# Standard naming
tmpltype -dir templates -pkg main -out template_gen.go

# Custom naming
tmpltype -dir tpl -pkg web -out web_templates.go

# With path
tmpltype -dir ../templates -pkg shared -out ../shared/templates_gen.go
```

## Usage Examples

### Basic Usage

Generate from a single directory:

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

**Project structure:**
```
myproject/
├── main.go
├── template_gen.go (generated)
└── templates/
    └── email.tmpl
```

### Multiple Templates (Flat Structure)

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

**Project structure:**
```
myproject/
├── main.go
├── template_gen.go (generated)
└── templates/
    ├── email.tmpl
    ├── sms.tmpl
    └── push.tmpl
```

**Generated:**
- `Email`, `Sms`, `Push` types
- `RenderEmail()`, `RenderSms()`, `RenderPush()` functions

### Template Grouping (Subdirectories)

```bash
tmpltype -dir templates -pkg main -out template_gen.go
```

**Project structure:**
```
myproject/
├── main.go
├── template_gen.go (generated)
└── templates/
    ├── footer.tmpl                    # Flat
    ├── mail_invite/                   # Group
    │   ├── title.tmpl
    │   └── content.tmpl
    └── mail_account_created/          # Group
        ├── title.tmpl
        └── content.tmpl
```

**Generated:**
- Flat: `Footer` type, `RenderFooter()` function
- Groups: `MailInviteTitle`, `MailInviteContent`, etc.
- Nested namespace in `Template` struct

See [Template Grouping documentation](template-grouping.md) for details.

### Using with Relative Paths

```bash
# From project root
tmpltype -dir ./internal/views/templates -pkg views -out ./internal/views/templates_gen.go

# From subdirectory
cd internal/views
tmpltype -dir templates -pkg views -out templates_gen.go
```

### Using Absolute Paths

```bash
tmpltype \
  -dir /home/user/project/templates \
  -pkg main \
  -out /home/user/project/template_gen.go
```

## Directory Scanning Behavior

### What Gets Scanned

`tmpltype` automatically scans:

1. **Flat templates**: `<dir>/*.tmpl`
2. **Grouped templates**: `<dir>/*/*.tmpl` (1 level deep only)

### Depth Limit

**✅ Scanned (depth 0 and 1):**
```
templates/
├── email.tmpl              ← depth 0 (scanned)
└── mail/
    └── invite.tmpl         ← depth 1 (scanned)
```

**❌ Not scanned (depth 2+):**
```
templates/
└── mail/
    └── invite/
        └── html.tmpl       ← depth 2 (not scanned)
```

### File Pattern

Only files with `.tmpl` extension are processed:

**✅ Processed:**
- `email.tmpl`
- `user.tmpl`
- `index.tmpl`

**❌ Ignored:**
- `email.txt`
- `user.html`
- `template.go`
- `.tmpl.bak`

### Hidden Files and Directories

Hidden files and directories (starting with `.`) are **ignored**:

**❌ Ignored:**
```
templates/
├── .backup/
│   └── old.tmpl            ← ignored (hidden directory)
├── .template.tmpl          ← ignored (hidden file)
└── email.tmpl              ← processed
```

## Integration with go generate

### Basic Setup

**1. Create `gen.go` in your package:**

```go
package main

//go:generate tmpltype -dir templates -pkg main -out template_gen.go
```

**2. Run generation:**

```bash
go generate
```

or

```bash
go generate ./...
```

### Multiple Packages

**Project structure:**
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

**Run all:**
```bash
go generate ./...
```

### Committing Generated Code

**Recommended:** Commit the generated code

```bash
git add template_gen.go
git commit -m "Update generated templates"
```

**Why?**
- ✅ No build dependency on `tmpltype`
- ✅ Faster CI builds
- ✅ Clear diff in code reviews
- ✅ Works with `go get`

**CI Verification:**

Ensure generated code is up-to-date in CI:

```bash
#!/bin/bash
# scripts/verify_generated.sh
go generate ./...
git diff --exit-code || (echo "Generated code is out of date. Run 'go generate ./...'" && exit 1)
```

### Alternative: Generate on Build

If you prefer not to commit generated code:

**Makefile:**
```makefile
.PHONY: generate
generate:
	go generate ./...

.PHONY: build
build: generate
	go build ./...
```

**Build:**
```bash
make build
```

## Troubleshooting

### Error: "no templates found"

**Cause:** No `.tmpl` files in the specified directory

**Solution:**
```bash
# Check if templates exist
ls -la templates/

# Verify you're using the correct path
tmpltype -dir templates -pkg main -out gen.go
```

### Error: "failed to parse template"

**Cause:** Template syntax error

**Solution:**
```bash
# Check template syntax
# Look for unclosed {{ }}, missing end tags, etc.
```

**Common issues:**
- Missing `{{ end }}`
- Unclosed `{{` or `}}`
- Invalid Go template syntax

### Error: "cannot infer type"

**Cause:** Complex type that cannot be automatically inferred

**Solution:** Use `@param` directive

```html
{{/* @param Items []struct{ID int64; Name string} */}}
{{ range .Items }}
  <li>{{ .ID }}: {{ .Name }}</li>
{{ end }}
```

### Generated Code Has Compile Errors

**Cause:** Usually related to `@param` directive syntax

**Solution:** Check your `@param` directives

**Common issues:**
```go
// ❌ Wrong: commas in struct
{{/* @param User struct{Name string, Age int} */}}

// ✅ Correct: semicolons in struct
{{/* @param User struct{Name string; Age int} */}}

// ❌ Wrong: nested slices
{{/* @param Matrix [][]string */}}

// ✅ Correct: use struct
{{/* @param Matrix []struct{Row []string} */}}
```

See [`@param` Directive documentation](param-directive.md) for details.

### Templates Not Found in Subdirectories

**Cause:** Subdirectories deeper than 1 level are not scanned

**Solution:** Keep templates within 1 level of subdirectories

```
templates/
├── email.tmpl           ← scanned
└── mail/
    └── invite.tmpl      ← scanned
```

Not supported:
```
templates/
└── mail/
    └── invite/
        └── html.tmpl    ← NOT scanned (too deep)
```

### Permission Errors

**Error:** `permission denied: templates/`

**Solution:**
```bash
# Check directory permissions
ls -ld templates/

# Fix permissions if needed
chmod 755 templates/
chmod 644 templates/*.tmpl
```

### Path Issues on Windows

**Issue:** Windows path separators

**Solution:** Use forward slashes even on Windows

```bash
# ✅ Works on Windows
tmpltype -dir templates/email -pkg main -out gen.go

# ❌ May cause issues
tmpltype -dir templates\email -pkg main -out gen.go
```

## See Also

- [Getting Started Guide](getting-started.md) - Tutorial and basic concepts
- [Template Syntax](template-syntax.md) - Supported template constructs
- [`@param` Directive](param-directive.md) - Type directive reference
- [Template Grouping](template-grouping.md) - Organizing templates
