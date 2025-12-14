# Examples

Working code examples demonstrating tmpltype features.

## Quick Start

Each example is a complete, runnable Go project:

```bash
cd 01_basic
go generate
go run .
```

## Examples Overview

### Getting Started

| Example | What it demonstrates | When to use |
|---------|---------------------|-------------|
| **[01_basic](01_basic/)** | Basic type inference from template | Start here - simplest example |
| **[02_param_directive](02_param_directive/)** | Using `@param` for complex types | When you need specific types (int, pointers, etc.) |
| **[03_multi_template](03_multi_template/)** | Multiple templates in one package | Managing multiple templates |

### Comprehensive Examples

| Example | What it demonstrates | When to use |
|---------|---------------------|-------------|
| **[04_comprehensive_template](04_comprehensive_template/)** | All template syntax patterns | Reference for all supported constructs |
| **[05_all_param_types](05_all_param_types/)** | Complete `@param` type reference | Reference for all type directive patterns |

### Advanced Features

| Example | What it demonstrates | When to use |
|---------|---------------------|-------------|
| **[07_grouping](07_grouping/)** | Template grouping with subdirectories | Organizing templates by category |
| **[08_custom_functions](08_custom_functions/)** | Custom template functions | Using helper functions in templates |
| **[06_non_ascii_filename](06_non_ascii_filename/)** | Non-ASCII filenames | International file naming |

## Learning Path

### 1. First Steps
Start with **[01_basic](01_basic/)** to understand the basics:
- How tmpltype infers types
- Generated code structure
- Basic usage pattern

### 2. Type Control
Move to **[02_param_directive](02_param_directive/)** to learn:
- When automatic inference isn't enough
- How to specify exact types
- `@param` directive syntax

### 3. Multiple Templates
Try **[03_multi_template](03_multi_template/)** to see:
- Managing multiple templates
- How generated code scales
- Namespace organization

### 4. Reference Examples
Use these when you need specific patterns:
- **[04_comprehensive_template](04_comprehensive_template/)** - Template syntax reference
- **[05_all_param_types](05_all_param_types/)** - Type directive reference

### 5. Production Patterns
Explore advanced features:
- **[07_grouping](07_grouping/)** - Large-scale template organization
- **[08_custom_functions](08_custom_functions/)** - Extending template functionality

## Quick Reference by Use Case

### "I want to..."

| Goal | Example |
|------|---------|
| Get started quickly | [01_basic](01_basic/) |
| Use specific types (int, time.Time, etc.) | [02_param_directive](02_param_directive/) |
| Work with multiple templates | [03_multi_template](03_multi_template/) |
| Understand all template syntax | [04_comprehensive_template](04_comprehensive_template/) |
| See all `@param` patterns | [05_all_param_types](05_all_param_types/) |
| Organize templates by category | [07_grouping](07_grouping/) |
| Use custom helper functions | [08_custom_functions](08_custom_functions/) |

## Running an Example

```bash
# Navigate to example directory
cd 01_basic

# Generate code
go generate

# Run the example
go run .

# See debug output (optional)
TMPLTYPE_LOG_LEVEL=debug go generate
```

## Example Structure

Each example follows this structure:

```
01_basic/
├── gen.go              # go:generate directive
├── main.go             # Example usage code
├── templates/          # Template files
│   └── email.tmpl
├── template_gen.go     # Generated types and render functions (generated)
└── template_sources_gen.go  # Generated template strings (generated)
```

## Next Steps

- Read the [documentation](../docs/) for detailed explanations
- Check the [CLI reference](../docs/cli-reference.md) for all options
- See [template syntax guide](../docs/template-syntax.md) for supported constructs
