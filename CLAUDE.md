# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the `tools/AI/claude` directory within the "面向面试编程" (Interview-oriented Programming) repository. The main repository is focused on interview preparation through algorithmic problem solving, data structures, and common CRUD components.

## Architecture

The repository is organized into several main directories:
- `network-program/` - Network programming examples including Go concurrency patterns, HTTP demos, socket programming
- `program-language/` - Language-specific learning materials  
- `leetcode/` - Algorithm and data structure practice
- `tools/` - Development tools and utilities

### Key Go Projects Structure
- Each Go project has its own `go.mod` file using Go 1.23.0
- Common pattern: logger packages with corresponding test files
- Main applications typically in project root directories
- Examples include:
  - `go-concurrency/` - Concurrency patterns, port scanning, pprof analysis
  - `http-demo/` - HTTP servers and uploaders
  - `socket-demo/` - Socket programming examples

## Common Commands

### Go Development
```bash
# Build and run Go programs
go run main.go
go run <file>.go

# Run tests
go test ./...
go test -v ./logger

# Build executables
go build

# Get dependencies
go mod tidy
```

### Debug Configuration
VSCode is configured with Go debugging support in `.vscode/launch.json` with "Launch file" configuration that debugs the currently open file.

## Development Workflow

1. Navigate to specific project directories (e.g., `network-program/go-concurrency/`)
2. Each Go module is self-contained with its own dependencies
3. Use `go run` for quick testing and iteration
4. Logger packages are commonly used across projects with corresponding test files

## Project Goals
The repository focuses on:
- Algorithm and data structure practice for interviews
- Go and Python programming (Go preferred for leetcode problems)
- Network programming concepts and implementations
- Common backend component understanding