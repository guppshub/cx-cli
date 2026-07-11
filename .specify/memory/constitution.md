<!--
SYNC IMPACT REPORT
==================
Version Change: None -> 1.0.0
Ratification Date: 2026-07-10
Last Amended Date: 2026-07-10

Modified Principles:
- Initialized all 14 core principles replacing template placeholders.

Added Sections:
- Engineering Principles
- Decision Framework
- Governance

Removed Sections:
- None

Templates Status:
- .specify/templates/plan-template.md: ✅ No changes needed
- .specify/templates/spec-template.md: ✅ No changes needed
- .specify/templates/tasks-template.md: ✅ No changes needed

Follow-up TODOs:
- None
-->

# cx-cli Constitution

This constitution defines the engineering principles that guide the design, implementation, and evolution of cx-cli. These principles apply equally to human contributors and AI agents. Every architectural and implementation decision MUST align with this document. (cx-cli is short for cloudx-cli).

## Core Principles

### 1. Specification Before Implementation
Requirements MUST be defined before implementation begins.

Every feature starts with a clear specification that includes:
- Problem statement
- Functional requirements
- Acceptance criteria
- Edge cases
- Success criteria

Implementation MUST never drive requirements. When requirements are unclear or incomplete, seek clarification before writing code.

### 2. Simplicity Over Cleverness
Favor simple, readable solutions over clever or highly abstract implementations.

Every line of code MUST have a clear purpose.

Avoid unnecessary abstractions, frameworks, design patterns, or optimizations that do not solve a current problem.

Code MUST be understandable by a new contributor with minimal context.

### 3. Extensibility Without Overengineering
cx-cli MUST be easy to extend without introducing unnecessary complexity.

Design packages with clear responsibilities and well-defined boundaries.

Prefer composition over inheritance and small interfaces over large abstractions.

Do NOT introduce plugin systems, factories, generic frameworks, or excessive indirection until there is a demonstrated need.

Abstractions MUST emerge from working implementations, not anticipated future needs. Proving the design with one real provider and one real workflow before generalizing interfaces avoids speculative over-design.

Design for future growth without paying today's complexity cost.

### 4. Go Idioms First
cx-cli follows established Go conventions and community best practices.

All code MUST align with:
- Effective Go
- Go Code Review Comments
- Standard Go project layout where appropriate
- Idiomatic package organization
- Composition over inheritance
- Small focused packages
- Clear naming
- Explicit error handling
- Minimal interfaces

When practices from other languages conflict with Go conventions, Go conventions MUST take precedence.

### 5. Standard Library First
Prefer the Go standard library whenever practical.

Third-party dependencies MUST only be introduced when they provide substantial value that cannot reasonably be achieved using the standard library.

Every dependency MUST have a clear technical justification.

Minimize dependency footprint to improve maintainability, portability, and long-term stability.

### 6. Production-Quality Code
Every contribution MUST be production-ready.

Code MUST prioritize:
- Readability
- Maintainability
- Reliability
- Consistency
- Observability

Temporary workarounds, commented-out code, dead code, and unfinished implementations MUST never be merged.

### 7. Explicit Error Handling
Errors are part of the API.

Errors MUST:
- Include meaningful context
- Be wrapped appropriately
- Never be silently ignored
- Never expose secrets or sensitive information

Panics are reserved only for unrecoverable initialization failures.

### 8. Testability by Design
Code MUST naturally support testing.

Business logic MUST remain independent from infrastructure and user interfaces.

Use network socket boundaries (e.g. net.Conn interfaces) to mock or encapsulate subprocesses or platform-specific connection tools rather than coupling the core application logic to them.

Prefer deterministic, isolated unit tests.

Critical workflows SHOULD include integration tests where appropriate.

Design code that is easy to test rather than relying heavily on mocks.

### 9. Backward Compatibility
cx-cli is intended to become a widely used open-source CLI.

Breaking changes to:
- CLI commands
- Configuration files
- Public APIs
- Output formats
- File structures

MUST be deliberate, documented, justified, and accompanied by a migration strategy whenever practical.

### 10. Security by Default
User trust is non-negotiable.

cx-cli MUST:
- Never log secrets, credentials, session tokens, or private keys
- Validate all external input
- Minimize shell execution
- Protect against command injection
- Follow the principle of least privilege

Providers must be headless library packages. Any user interaction (such as password/MFA prompts) must be driven via functional callbacks provided by the caller, avoiding internal stdin/stdout prompts within provider packages.

Security MUST be considered during design rather than added later.

### 11. Documentation is Part of the Feature
A feature is not complete until its documentation is complete.

User-facing functionality MUST include appropriate documentation such as:
- Command help
- Usage examples
- README updates where applicable
- Migration notes for breaking changes

Documentation MUST evolve alongside the codebase.

### 12. AI Contributor Guidelines
AI-generated code MUST meet the same quality standards as human-written code.

AI contributors MUST:
- Prefer modifying existing code over rewriting it
- Avoid unnecessary refactoring
- Preserve existing behavior unless requirements specify otherwise
- Explain significant architectural decisions
- Never invent APIs, libraries, or behavior
- Ask for clarification when requirements are ambiguous
- Produce code that follows this constitution without exception

### 13. Contributor Experience
cx-cli is designed to be approachable for new contributors.

The codebase MUST emphasize:
- Predictable package structure
- Consistent coding patterns
- Self-documenting code
- Descriptive naming
- Low cognitive overhead

A contributor MUST be able to understand a package's purpose without reading the entire repository.

Maintainability is a feature.

### 14. Graceful Dependency Management
cx-cli depends on external tools such as AWS CLI, Session Manager Plugin, tmux, and other platform-specific utilities.

The application MUST:
- Detect missing dependencies before they are required
- Provide actionable error messages
- Offer installation guidance whenever possible
- Fail fast with clear explanations
- Isolate platform-specific behavior behind well-defined interfaces

Dependency management MUST improve the user experience rather than create friction.

## Engineering Principles
The following principles guide everyday engineering decisions:
- Specification before implementation.
- Simplicity over cleverness.
- Design for extension, not speculation.
- Follow idiomatic Go.
- Prefer the standard library.
- Production-ready over prototype-ready.
- Security by default.
- Scalability MUST be considered during design, but avoid optimizing for hypothetical scale.
- Performance optimizations MUST be driven by measurement, not assumptions.
- Every dependency MUST earn its place.
- Documentation is part of the implementation.
- Code MUST be written for humans first and machines second.

## Decision Framework
When multiple solutions are possible, prefer the one that is:
1. Correct
2. Simple
3. Idiomatic Go
4. Maintainable
5. Extensible
6. Well-tested
7. Secure
8. Performant enough for current requirements
9. Easy for future contributors to understand

### Final Principle
cx-cli aims to become a high-quality open-source project.

Every contribution MUST leave the codebase in a better state than it was found.

When in doubt, choose the solution that maximizes clarity, maintainability, and long-term sustainability.

## Governance
- **Compliance**: All contributions (including code, specifications, and plans) MUST comply with this Constitution. Pull requests and AI-generated code will be reviewed against these principles.
- **Amendments**: Proposed changes to these principles or governance rules MUST be documented, justified, and ratified. Versioning follows semantic rules (MAJOR for removals/redefinitions, MINOR for additions, PATCH for clarifications).
- **Supersedence**: This Constitution is the authoritative reference for cx-cli engineering standards and supersedes any conflicting local or personal development preferences.

**Version**: 1.0.0 | **Ratified**: 2026-07-10 | **Last Amended**: 2026-07-10
