# Happy Study 🎉

A Go learning and development project.

## Getting Started

First, initialize dependencies:

```bash
go mod tidy
```

Then build and run:

```bash
# Build
make build

# Run
make run

# Test
make test
```

## Project Structure

```
happy-study/
├── cmd/          # Entry points
│   └── app/      # Main application
├── internal/     # Internal packages
│   ├── agent/    # AI Agent definitions
│   │   ├── interviewer/  # Interviewer Agent
│   │   ├── teacher/      # Teacher Agent
│   │   └── workflow/     # Eino Graph orchestration
│   ├── handler/  # HTTP handlers
│   ├── service/  # Business logic
│   └── repository/ # Data access
├── pkg/          # Shared packages
├── configs/      # Configuration files
├── docs/         # Documentation
│   ├── design/           # Product design
│   └── architecture/     # Tech stack & architecture
└── scripts/      # Build scripts
```

## Prerequisites

- Go 1.26+
- Make

## Quick Start

```bash
# Run the diagnostic interview and learning demo
go run ./cmd/app --topic "Go 并发编程"

# Diagnosis only (skip curriculum generation)
go run ./cmd/app --topic "Go 并发编程" --diagnosis-only
```
