# Contributing to Lore-Core

## Getting Started

1. Fork the repository
2. Clone your fork
3. Install dependencies: `make tools`
4. Run checks: `make check`

## Development Workflow

### Branch Naming

| Prefix | Purpose | Example |
|--------|---------|---------|
| `feature/` | New features | `feature/import-command` |
| `fix/` | Bug fixes | `fix/embedding-error` |
| `docs/` | Documentation | `docs/api-reference` |
| `refactor/` | Code refactoring | `refactor/extraction-service` |
| `chore/` | Maintenance | `chore/update-dependencies` |

### Before Coding

1. Create a branch from `main`
2. Document your task in `tasks/YYYY-MM-DD-description.md`

### Commit Messages

Use conventional commits:

```
<type>(<scope>): <description>

[body - what and why]
```

**Types:** `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `style`, `perf`

**Rules:**
- Header: max 50 characters (hard limit: 72)
- Body: wrap at 72 characters

### Before Submitting

Run all checks:

```bash
make check
```

This runs format, vet, lint, and test.

## Pull Requests

### Title
Same as commit header or descriptive title.

### Changes
Bullet points summarizing what changed.

### Testing
Include ALL steps to test:
- Prerequisites (Docker, build commands)
- Configuration (API keys, env vars)
- Data setup
- Test commands (each in its own code block)

## Code Standards

See [CLAUDE.md](CLAUDE.md) for:
- Go formatting and naming conventions
- Error handling patterns
- Directory structure
- Architecture guidelines (hexagonal)
- Anti-patterns to avoid

## Architecture

```
internal/
├── domain/           # Core business logic (no external deps)
│   ├── entities/
│   ├── services/
│   └── ports/        # Interfaces
├── application/      # Use cases, handlers
└── infrastructure/   # External implementations (qdrant, openai, etc.)
```

**Key rule:** Dependencies point inward. Domain has no external dependencies.

## Questions?

Open an issue for questions or discussions.
