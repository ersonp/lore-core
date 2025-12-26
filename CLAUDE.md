# Lore-Core Project Guidelines

## Project Overview

Lore-Core is a factual consistency database for fictional worlds. It extracts, stores, and queries facts using Qdrant (vector DB) and Claude/OpenAI (LLM).

**Tech Stack:** Go, Qdrant, Claude API, OpenAI Embeddings

---

# Go Coding Standards

## Formatting

- All code must be formatted with [`goimports`](https://godoc.org/golang.org/x/tools/cmd/goimports)
- Use `make format` to format all source files
- All files must end with a newline character
- No trailing whitespace on any line (applies to all files, not just `.go`)

---

## Naming Conventions

### Packages
- Short, lowercase, single-word names
- No underscores or mixedCaps
- Package name must match directory name

```go
// Good
package extraction
package qdrant

// Bad
package fact_extraction
package QdrantClient
```

### Filenames
- Lowercase only
- Use underscores for multiword: `fact_service.go`, `llm_client.go`

### Variables and Functions
- Camel case always — no underscores, no ALLCAPS
- Unexported: `lowerCamelCase`
- Exported: `UpperCamelCase`
- Acronyms consistent case: `ID`, `HTTP`, `URL`
- Names should be meaningful

```go
// Good
var userID string
func GetUserID() string
type HTTPClient struct{}

// Bad
var UserId string       // Wrong
var user_id string      // Wrong
var USERID string       // Wrong
```

### Interfaces
- Single-method interfaces end with `-er`: `Reader`, `Writer`, `Extractor`
- Interfaces describe behavior, not data

```go
// Good
type Extractor interface {
    Extract(text string) ([]Fact, error)
}

// Bad
type IExtractor interface{}  // No "I" prefix
type FactData interface{}    // Interfaces are for behavior
```

---

## Struct Instantiation

Use field names, each on its own line:

```go
// Good
var foo = Foo{
    Bar: 1,
    Baz: "cat",
}

// Bad
var foo = Foo{1, "cat"}
var foo = Foo{Bar: 1, Baz: "cat"}  // All on one line
```

Single field may be on same line:
```go
var foo = Foo{Bar: 1}
```

---

## Error Handling

### Return errors with `err != nil`
```go
x, err := foo()
if err != nil {
    return err
}

// Or inline
if err := foo(); err != nil {
    return err
}
```

### Wrap errors with context
```go
if err != nil {
    return fmt.Errorf("extracting facts from %s: %w", filename, err)
}
```

### Sentinel errors (package-level)
```go
package domain

var (
    ErrFactNotFound     = errors.New("fact not found")
    ErrInconsistentFact = errors.New("inconsistent fact detected")
)
```

### Typed errors (for error classes)
```go
type ExtractionError struct {
    File string
    Line int
}

func (e ExtractionError) Error() string {
    return fmt.Sprintf("extraction failed at %s:%d", e.File, e.Line)
}

// Caller can check type
switch err.(type) {
case ExtractionError:
    // Handle extraction errors
}
```

---

## Functions and Methods

### No named return values (unless necessary)
```go
// Good
func foo(i int) (int, error) {
    return i + 2, nil
}

// Bad
func foo(i int) (bar int, err error) {
    bar = i + 2
    return
}
```

Named returns only when:
- Needed for `defer` error handling
- Many ambiguous return values

```go
// OK - defer needs to modify err
func CopyFile(dst string, src io.Reader) (n int64, err error) {
    out, err := os.Create(dst)
    if err != nil {
        return 0, err
    }
    defer func() {
        cerr := out.Close()
        if err == nil {
            err = cerr
        }
    }()
    n, err = io.Copy(out, src)
    return
}
```

### Constructor functions
- Use `New` prefix
- Return pointer for structs with methods
- Validate inputs

```go
func NewFactService(repo FactRepository, llm LLMClient) (*FactService, error) {
    if repo == nil {
        return nil, errors.New("repository is required")
    }
    return &FactService{
        repo: repo,
        llm:  llm,
    }, nil
}
```

### Method receivers
- Pointer receivers for methods that modify state
- Short, consistent names (1-2 letters)

```go
func (s *FactService) Save(fact Fact) error {}
func (s *FactService) FindByID(id string) (Fact, error) {}
```

### Context as first parameter
```go
func (s *FactService) Save(ctx context.Context, fact Fact) error {}
```

---

## Imports

### No dot imports
```go
// Bad - pollutes namespace
import . "fmt"

// Good
import "fmt"
```

### Avoid renaming imports unless necessary

---

## Directory Structure

```
lore-core/
├── cmd/                     # Command entry points
│   └── lore/
│       └── main.go
├── internal/                # Private application code
│   ├── domain/              # Core business logic (no external deps)
│   │   ├── entities/
│   │   ├── services/
│   │   └── ports/           # Interfaces
│   ├── application/         # Use cases
│   │   └── usecases/
│   └── infrastructure/      # External implementations
│       ├── qdrant/
│       ├── claude/
│       └── openai/
├── pkg/                     # Public library code (if any)
├── testdata/                # Test fixtures
├── vendor/                  # Vendored dependencies (committed)
├── reports/
├── tasks/
└── task-sprints/
```

### Import Rules (Hexagonal)
- Domain imports nothing from infrastructure
- Application imports domain only
- Infrastructure imports domain (to implement ports)
- cmd imports everything to wire dependencies

---

## Dependencies

### Use `go mod`
```bash
go mod init           # Initialize
go mod tidy           # Clean unused
go mod vendor         # Vendor dependencies
```

- Commit `go.mod`, `go.sum`, and `vendor/` (this is an application, not a library)

### Be judicious with external dependencies
- Avoid compulsive use of frameworks
- Prefer stdlib when reasonable

---

## Code Reuse and DRY

DRY is not as strict in Go. Sometimes it's okay to copy code.

- Apply DRY when code represents the **same concept**
- Allow duplication when code is **similar but different contexts**
- Don't overdesign — wrong abstraction is worse than duplication

---

## Testing

### Use testify for assertions
```go
import (
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
    result, err := DoSomething()
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Use mockery for generating mocks
```bash
mockery --name=LLMClient --output=mocks
```

### Table-driven tests
```go
func TestFactType_IsValid(t *testing.T) {
    tests := []struct {
        name     string
        factType FactType
        want     bool
    }{
        {"valid character", FactTypeCharacter, true},
        {"valid location", FactTypeLocation, true},
        {"invalid type", FactType("invalid"), false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.factType.IsValid()
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Test data
- Put test fixtures in `testdata/` folder within the relevant package

---

## Comments and Documentation

### Package comments
```go
// Package extraction provides services for extracting facts from text
// using LLM-based analysis.
package extraction
```

### Exported function comments
```go
// Extract analyzes the given text and returns a slice of Facts.
// It uses the configured LLM to identify characters, locations,
// events, and relationships.
func (s *ExtractionService) Extract(ctx context.Context, text string) ([]entities.Fact, error) {}
```

### Comment the WHY, not the WHAT
```go
// Bad
// Increment counter by 1
counter++

// Good
// Rate limit: max 10 requests per second to Claude API
time.Sleep(100 * time.Millisecond)
```

---

## Patterns

### Options pattern for complex configuration
```go
type Option func(*Config)

func WithChunkSize(size int) Option {
    return func(c *Config) {
        c.ChunkSize = size
    }
}

func NewExtractor(opts ...Option) *Extractor {
    cfg := defaultConfig()
    for _, opt := range opts {
        opt(&cfg)
    }
    return &Extractor{config: cfg}
}
```

### Avoid init() functions
- Prefer explicit initialization in main()
- init() makes testing harder

### Zero values should be useful
```go
type Config struct {
    ChunkSize int // 0 means use default
}

func (c Config) GetChunkSize() int {
    if c.ChunkSize == 0 {
        return 2000
    }
    return c.ChunkSize
}
```

---

## Anti-Patterns to Avoid

### No mutable global variables for flags
Use local variables with closures for Cobra flags:

```go
// Bad - global mutable state
var myFlag string

func newCmd() *cobra.Command {
    cmd := &cobra.Command{RunE: runCmd}
    cmd.Flags().StringVar(&myFlag, "flag", "", "description")
    return cmd
}

// Good - local scope with closure
func newCmd() *cobra.Command {
    var myFlag string
    cmd := &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            return runCmd(cmd, args, myFlag)
        },
    }
    cmd.Flags().StringVar(&myFlag, "flag", "", "description")
    return cmd
}
```

### No magic numbers
Use named constants for numeric values:

```go
// Bad
facts, err := repo.List(ctx, 1000, 0)

// Good
const DefaultListLimit = 1000
facts, err := repo.List(ctx, DefaultListLimit, 0)
```

### No ignored errors
Always handle or explicitly acknowledge errors:

```go
// Bad - silent failure
count, _ := repo.Count(ctx)

// Good - handle error
count, err := repo.Count(ctx)
if err != nil {
    log.Printf("warning: could not get count: %v", err)
    count = 0
}

// Good - document intentional ignore
response, _ := reader.ReadString('\n') // EOF treated as empty
```

### No deep nesting
Use early returns and guard clauses (max 3 levels):

```go
// Bad - deep nesting
if condition1 {
    if condition2 {
        if condition3 {
            doSomething()
        }
    }
}

// Good - early returns
if !condition1 {
    return
}
if !condition2 {
    return
}
if !condition3 {
    return
}
doSomething()
```

### No empty interface without reason
Prefer typed structs over `interface{}` or `any`:

```go
// Bad - loses type safety
data := map[string]interface{}{}

// Good - typed struct
type Config struct {
    Name  string
    Value int
}
```

### No panics in library code
Return errors instead of panicking:

```go
// Bad
func Parse(s string) int {
    v, err := strconv.Atoi(s)
    if err != nil {
        panic(err)
    }
    return v
}

// Good
func Parse(s string) (int, error) {
    return strconv.Atoi(s)
}
```

### No context in structs
Pass context as first parameter, don't store it:

```go
// Bad
type Service struct {
    ctx context.Context
}

// Good
func (s *Service) DoWork(ctx context.Context) error {
    // use ctx here
}
```

### No API/database calls inside loops
Batch external calls instead of calling per-item:

```go
// Bad - O(n) API calls, extremely slow
for _, fact := range facts {
    result, _ := llmClient.Process(ctx, fact)      // API call per item
    similar, _ := vectorDB.Search(ctx, fact.Embedding)  // DB call per item
}

// Good - batch operations, O(1) API calls
embeddings, _ := embedder.EmbedBatch(ctx, texts)   // Single API call
err := vectorDB.SaveBatch(ctx, facts)              // Single DB call

// Good - collect first, then batch
var allSimilar []Fact
for _, fact := range facts {
    similar, _ := vectorDB.Search(ctx, fact.Embedding)  // Fast DB calls OK
    allSimilar = append(allSimilar, similar...)
}
issues, _ := llmClient.CheckBatch(ctx, facts, allSimilar)  // Single LLM call
```

**Impact:** 50 items with per-item LLM calls = ~100 seconds. Batched = ~3 seconds.

---

## Makefile Commands

```bash
make format   # Format all Go files with goimports
make lint     # Run golangci-lint
make test     # Run all tests
make vet      # Run go vet
make build    # Build binary to bin/lore
make check    # Run all checks (format, vet, lint, test)
make vendor   # Sync vendor directory (go mod tidy && go mod vendor)
make clean    # Clean build artifacts
make tools    # Install development tools (goimports, golangci-lint, mockery)
```

---

## Pre-commit Checklist

Run before providing commit message to user:
```bash
make check
```

This runs format, vet, lint, and test. Fix any issues before presenting the commit.

---

## Git Branching (GitHub Flow)

### Branch Structure

```
main ─────────────────────────────────────────►
       \                    /
        └── feature/xyz ───┘
```

- `main` — Always stable and deployable
- Feature branches — Short-lived, merged via PR

### Branch Naming

| Prefix | Purpose | Example |
|--------|---------|---------|
| `feature/` | New features | `feature/ingest-command` |
| `fix/` | Bug fixes | `fix/qdrant-connection` |
| `docs/` | Documentation | `docs/api-reference` |
| `refactor/` | Code refactoring | `refactor/extraction-service` |
| `chore/` | Maintenance tasks | `chore/update-dependencies` |

### Rules

1. Never commit directly to `main`
2. Create branch from `main` for each task
3. Keep branches short-lived (merge within days, not weeks)
4. Delete branch after merge
5. One task = one branch = one PR

### Workflow

```bash
# Start new work
git checkout main
git pull
git checkout -b feature/my-feature

# Do work, commit often
git add .
git commit -m "feat(scope): description"

# Push and create PR
git push -u origin feature/my-feature
# Create PR on GitHub, merge, delete branch
```
