# Example 08: Custom Functions

This example demonstrates how to use custom template functions with tmpltype.

## Overview

Custom functions allow you to extend Go templates with your own logic for:
- String manipulation (upper, lower, trim, etc.)
- Date/time formatting
- Number formatting
- HTML generation
- And more...

## How It Works

### 1. Define Custom Functions

Create a function that returns `template.FuncMap`:

```go
func GetTemplateFuncs() template.FuncMap {
    return template.FuncMap{
        "upper": strings.ToUpper,
        "formatDate": func(t time.Time) string {
            return t.Format("2006-01-02")
        },
        // ... more functions
    }
}
```

### 2. Use Functions in Templates

```html
<h1>{{ .Title | upper }}</h1>
<p>Date: {{ formatDate .CreatedAt }}</p>
```

### 3. Initialize with Functions

```go
func main() {
    // Initialize templates with custom functions
    InitTemplates(WithFuncs(GetTemplateFuncs()))

    // Now you can render
    RenderEmail(&buf, data)
}
```

## Functional Option Pattern

The `InitTemplates` function uses the functional option pattern for flexibility:

```go
// Without custom functions
InitTemplates()

// With custom functions
InitTemplates(WithFuncs(GetTemplateFuncs()))

// Future: Multiple options
InitTemplates(
    WithFuncs(funcs),
    WithOtherOption(value),
)
```

## Running the Example

```bash
# Generate code
go generate

# Run
go run .
```

## Key Features

1. **Type Safety**: Template parameters are still type-safe
2. **Flexible Initialization**: Choose whether to use custom functions
3. **Clear Errors**: Forgetting to call `InitTemplates()` gives a clear error message
4. **No CLI Flags**: No need to specify function names in `go:generate` directive

## Custom Functions in This Example

- `upper` - Convert string to uppercase
- `lower` - Convert string to lowercase
- `formatDate` - Format time.Time as date
- `formatDateTime` - Format time.Time as datetime
- `nl2br` - Convert newlines to HTML `<br>` tags
- `default` - Provide default value if empty
- `comma` - Add thousands separator to numbers
