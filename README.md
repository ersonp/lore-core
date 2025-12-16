# Lore-Core

A factual consistency database for fictional worlds. Extract, store, and query facts from stories using vector search and LLM-powered analysis.

## Overview

Lore-Core helps writers and world-builders maintain consistency in their fictional universes by:

- **Extracting** facts (characters, locations, events, relationships) from existing stories
- **Storing** facts in a vector database for semantic search
- **Querying** related facts using natural language
- **Checking** new content for inconsistencies with established lore

## Features

- LLM-powered fact extraction (Claude/OpenAI)
- Semantic search via Qdrant vector database
- CLI interface for easy integration into writing workflows
- Consistency checking against established facts

## Installation

```bash
go install github.com/ersonp/lore-core/cmd/lore@latest
```

Or build from source:

```bash
git clone https://github.com/ersonp/lore-core.git
cd lore-core
make build
```

## Quick Start

```bash
# Initialize a new lore database
lore init

# Ingest a story
lore ingest story.txt

# Query facts
lore query "What color are Frodo's eyes?"

# Check new content for inconsistencies
lore check new-chapter.txt
```

## Configuration

Create `.lore/config.yaml` in your project:

```yaml
llm:
  provider: claude
  model: claude-sonnet-4-20250514
  api_key_env: ANTHROPIC_API_KEY

embeddings:
  provider: openai
  model: text-embedding-3-small
  api_key_env: OPENAI_API_KEY

qdrant:
  host: localhost
  port: 6334
  collection: lore_facts
```

## Requirements

- Go 1.21+
- Qdrant (local or cloud)
- API keys for Claude and/or OpenAI

### Running Qdrant locally

```bash
docker run -p 6333:6333 -p 6334:6334 qdrant/qdrant
```

## Architecture

Lore-Core follows hexagonal architecture:

```
internal/
├── domain/           # Core business logic
│   ├── entities/     # Fact, Character, Location, Event
│   ├── services/     # Extraction, Query, Consistency
│   └── ports/        # Interfaces for external services
├── application/      # Use cases
└── infrastructure/   # Qdrant, Claude, OpenAI adapters
```

## Lore-* Ecosystem

Lore-Core is the foundation for a suite of tools:

| Module | Purpose | Status |
|--------|---------|--------|
| `lore-core` | Factual database and consistency | In Development |
| `lore-viz` | Consistent image generation | Planned |
| `lore-timeline` | Visual timeline generation | Planned |
| `lore-graph` | Relationship visualization | Planned |

## Development

```bash
# Format code
make format

# Run tests
make test

# Run linter
make lint

# Run all checks
make check
```

## Contributing

See [CLAUDE.md](CLAUDE.md) for coding guidelines.

## License

Apache 2.0 - see [LICENSE](LICENSE)
