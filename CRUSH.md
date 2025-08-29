# CRUSH Configuration for surflog

## Build/Test Commands
- **Build**: `go build -o surflog .`
- **Run**: `go run .` or `./surflog`
- **Test all**: `go test ./...`
- **Test single package**: `go test ./cmd/journal`
- **Test with coverage**: `go test -cover ./...`
- **Lint**: `go vet ./...`
- **Format**: `go fmt ./...`
- **Mod tidy**: `go mod tidy`

## Code Style Guidelines

### Imports
- Standard library first, then third-party, then local packages
- Use aliases for long package names: `tea "github.com/charmbracelet/bubbletea"`

### Naming Conventions
- Packages: lowercase, single word when possible
- Types: PascalCase (Entry, Model, Service)
- Functions/methods: PascalCase for exported, camelCase for unexported
- Variables: camelCase, descriptive names
- Constants: PascalCase for exported, camelCase for unexported

### Types & Structs
- Use struct tags for JSON serialization: `json:"field_name"`
- Unexported fields for internal state, exported for public API
- Group related fields together in structs

### Error Handling
- Return errors as last return value
- Use descriptive error messages
- Handle errors at call site, don't ignore
- Use `errors.New()` for simple error strings

### Comments
- Only add comments when explicitly requested by user
- Keep code self-documenting through clear naming