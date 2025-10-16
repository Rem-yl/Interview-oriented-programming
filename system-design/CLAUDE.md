# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a **12-week System Design Deep Learning Course** repository focused on teaching distributed systems principles through hands-on projects. Unlike interview-oriented materials, this course emphasizes deep understanding of system design principles, source code reading, and production-grade implementations.

**Key Philosophy**: Learn by doing - every concept must be implemented in code (Go or Python), with emphasis on understanding "why" over "what".

## Repository Structure

```
system-design/
├── README.md                    # Main course curriculum (12 weeks, 4 phases)
├── DAILY_CHECKLIST.md          # Daily learning tasks and progress tracking
├── notes/                       # Learning notes and summaries
│   ├── README.md               # Note-taking guidance with detailed templates
│   ├── week{N}/                # Weekly notes
│   ├── phase{N}-summary.md     # Phase summaries
│   └── final-summary.md        # Course completion summary
└── projects/                    # Hands-on implementation projects
    ├── README.md               # Project guidelines
    └── week{N}/                # Weekly projects
        └── {project-name}/     # Individual projects with READMEs
```

## Learning Approach

### Course Structure (12 Weeks)

**Phase 1 (Weeks 1-3)**: Foundation - Scalability, storage systems, network communication
**Phase 2 (Weeks 4-7)**: Advanced patterns - Caching, databases, event-driven, microservices
**Phase 3 (Weeks 8-10)**: Distributed systems core - Consensus algorithms, observability, optimization
**Phase 4 (Weeks 11-12)**: Real-world systems - Case studies and capstone project

### Weekly Pattern

Each week is organized into **modules** (not days), with each module containing:
- **Deep dive into principles**: Not just what, but why and how
- **Source code reading**: Study real implementations (PostgreSQL, Redis, etcd, etc.)
- **Production-grade projects**: Complete implementations with tests and performance benchmarks
- **Case studies**: Analyze real-world system architectures

### Technology Stack

**Go Projects** (preferred for high-performance services):
- Build: `go build`
- Run: `go run main.go`
- Test: `go test ./...` or `go test -v ./pkg`
- Benchmark: `go test -bench=. -benchmem`
- Init: `go mod init github.com/username/project-name`

**Python Projects** (for rapid prototyping):
- Virtual env: `python -m venv venv && source venv/bin/activate`
- Install deps: `pip install -r requirements.txt`
- Run: `python main.py`
- Test: `pytest` or `pytest -v tests/`

## Project Guidelines

### Project Structure Standards

**Go Project**:
```
project-name/
├── main.go              # Entry point
├── go.mod, go.sum       # Dependencies
├── README.md            # Project documentation
├── pkg/                 # Public packages
│   └── *.go
├── internal/            # Private packages
│   └── */
└── tests/               # Test files
    ├── *_test.go
    └── benchmark_test.go
```

**Python Project**:
```
project-name/
├── main.py              # Entry point
├── requirements.txt     # Dependencies
├── README.md
├── src/                 # Source code
│   └── *.py
└── tests/               # Test files
    └── test_*.py
```

### Code Quality Expectations

1. **Production-grade code**: Not just "works", but performant, tested, and documented
2. **Performance testing**: Include benchmarks showing O(1) complexity where claimed
3. **Concurrency safety**: All shared data structures must be goroutine-safe (Go) or thread-safe (Python)
4. **Comprehensive tests**: Unit tests + integration tests + performance benchmarks
5. **Deep comments**: Explain the "why", especially for complex algorithms

### Example Projects

**Week 1 - Load Balancer (Go)**:
- Implements Round Robin, Weighted Round Robin, Least Connections algorithms
- Health checking with automatic failover
- Performance metrics collection
- Key files: `pkg/algorithms/*.go`, `internal/healthcheck/checker.go`

**Week 4 - LRU Cache (Go)**:
- O(1) get/put operations using HashMap + Doubly Linked List
- TTL support with expiration
- Thread-safe with `sync.RWMutex`
- Statistics tracking (hit rate, evictions)
- Benchmark tests comparing performance

## Note-Taking System

### Templates Provided

Three comprehensive templates in `notes/README.md`:
1. **Daily notes**: Concepts, source code reading, implementation notes, reflections
2. **Weekly summary**: Knowledge synthesis, project retrospectives, capability assessment
3. **Phase summary**: Knowledge tree, breakthrough moments, ability evaluation

### Note Organization

- **Each module** gets detailed notes (not just each day)
- **Emphasis on**: Understanding principles, not memorizing facts
- **Include**: Architecture diagrams, code snippets with detailed comments, trade-off analysis
- **Track**: Questions, thoughts, eureka moments, resources discovered

### Key Practices

- Use Mermaid diagrams for architecture/flows (embedded in markdown)
- Record source code reading insights (file, line numbers, what you learned)
- Document trade-offs and design decisions
- Connect new concepts to previously learned material

## Common Operations

### Starting a New Week

1. Read week overview in `README.md`
2. Create `notes/week{N}/` directory
3. For each project:
   ```bash
   cd projects/week{N}/{project-name}
   # For Go:
   go mod init github.com/username/project-name
   # For Python:
   python -m venv venv && source venv/bin/activate
   ```

### Daily Learning Flow

1. **Theory** (1-1.5h): Read materials, watch videos, take notes
2. **Practice** (1-1.5h): Implement project, write tests, optimize
3. **Reflect** (30min): Update notes, review progress, plan next steps

### Running Tests

```bash
# Go - run all tests
cd projects/week{N}/{project}
go test ./...

# Go - run with coverage
go test -cover ./...

# Go - run specific test
go test -v -run TestLRUCache

# Go - benchmarks
go test -bench=. -benchmem

# Python - run all tests
pytest

# Python - with coverage
pytest --cov=src tests/
```

### Performance Analysis

```bash
# Go - CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Go - memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Python - profiling
python -m cProfile -o output.prof main.py
python -m pstats output.prof
```

## Source Code Reading

The course emphasizes reading real-world implementations. When studying source code:

1. **Start with architecture overview**: Understand the big picture first
2. **Focus on core algorithms**: Not every line, but key implementations
3. **Document insights**: In notes, record file paths, line numbers, and learnings
4. **Compare with your implementation**: What did they do differently? Why?

### Recommended Projects to Study

- **etcd**: Raft consensus algorithm implementation (Go)
- **Redis**: Data structures (C, but well-commented)
- **PostgreSQL**: B-Tree indexes, MVCC (C)
- **RocksDB**: LSM-Tree storage engine (C++)
- **Nginx**: Load balancing and event loop (C)

## Design Principles

### When Implementing Projects

1. **Principle over pattern**: Understand why, not just how
2. **Trade-offs matter**: Document performance vs complexity, consistency vs availability
3. **Iterative refinement**: Start simple, measure, optimize based on data
4. **Production mindset**: Code should handle errors, edge cases, concurrent access

### When Taking Notes

1. **Explain like teaching**: Use your own words, as if teaching someone
2. **Visual thinking**: Draw diagrams for complex systems
3. **Connect concepts**: Link to previous learnings
4. **Question everything**: Write down "why" questions even if unanswered

### When Analyzing Systems

1. **Requirements first**: Functional and non-functional needs
2. **Constraints matter**: Scale, consistency, latency requirements
3. **No silver bullet**: Every design has trade-offs
4. **Evolution over perfection**: Real systems evolve; design for change

## Weekly Project Expectations

Each week builds on previous knowledge. Projects should demonstrate:

- **Week 1-3**: Foundation implementations (load balancer, B-Tree, RPC framework)
- **Week 4-7**: Advanced patterns (LRU cache, sharding middleware, Saga coordinator)
- **Week 8-10**: Distributed algorithms (Raft implementation, distributed locks, observability)
- **Week 11**: System design case studies (URL shortener, Twitter, ride-sharing)
- **Week 12**: Capstone project (complete distributed system with monitoring)

## Course Completion Goals

By the end of 12 weeks, learners should be able to:

1. **Understand deeply**: Not just use Redis, but understand how skiplist works
2. **Implement independently**: Build production-grade Raft from scratch
3. **Design systems**: From requirements to architecture with justified trade-offs
4. **Read source code**: Navigate large codebases (etcd, PostgreSQL) effectively
5. **Optimize performance**: Identify bottlenecks, measure, and improve
6. **Build observability**: Integrate monitoring, logging, tracing into systems

## Resources Referenced in Course

- **Books**: "Designing Data-Intensive Applications" (DDIA), "Redis Design and Implementation"
- **Papers**: GFS, BigTable, Raft, Dynamo (all in course reading lists)
- **Courses**: MIT 6.824 Distributed Systems
- **Blogs**: Company engineering blogs (Uber, Netflix, Meta)

## Important Notes

- This is a **learning course**, not interview prep - depth over breadth
- **Source code reading** is mandatory, not optional
- **Projects must be completed** - no skipping to next week without finishing
- **Notes are personal** - use templates as guides, not rigid structures
- The **12-week timeline** is a guide; adjust based on your pace and depth of exploration
