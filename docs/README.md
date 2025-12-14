# Documentation

Complete documentation for tmpltype.

## Getting Started

Start here if you're new to tmpltype:

- **[Getting Started Guide](getting-started.md)** - Step-by-step tutorial from installation to first code generation

## Reference Documentation

### For Users

- **[CLI Reference](cli-reference.md)** - Complete command-line reference
  - All command-line options
  - Usage examples
  - Integration with `go generate`
  - Logging and debugging
  - Troubleshooting guide

- **[Template Syntax](template-syntax.md)** - Supported Go template constructs
  - Field references (`.Field`, `.Nested.Field`)
  - Control flow (`if`, `with`, `range`)
  - Collections (slices, maps)
  - Type inference rules

- **[`@param` Directive](param-directive.md)** - Complete type directive reference
  - When to use `@param`
  - Supported types (primitives, slices, maps, structs, pointers)
  - Syntax examples
  - Common patterns

- **[Template Grouping](template-grouping.md)** - Organize templates in subdirectories
  - Flat vs grouped templates
  - Namespace generation
  - File structure examples

### For Developers

- **[Multi-Template Design](multi-template-design.md)** - Internal design documentation
- **[Plan](plan.md)** - Project planning and architecture notes

## Quick Reference

### "I want to..."

| Goal | Document |
|------|----------|
| Get started from scratch | [Getting Started](getting-started.md) |
| Understand command-line options | [CLI Reference](cli-reference.md) |
| See debug output | [CLI Reference - Logging](cli-reference.md#logging) |
| Learn supported template syntax | [Template Syntax](template-syntax.md) |
| Use specific types (int, pointers, etc.) | [`@param` Directive](param-directive.md) |
| Organize templates in folders | [Template Grouping](template-grouping.md) |
| Troubleshoot errors | [CLI Reference - Troubleshooting](cli-reference.md#troubleshooting) |

## Language

- **English**: Current directory
- **日本語**: [`ja/`](ja/) directory

## Examples

Working code examples are available in the [`examples/`](../examples/) directory. See the [examples README](../examples/README.md) for a guide to all available examples.
