# Contributing to FlowForge

Thank you for your interest in contributing to FlowForge! This document provides guidelines for contributing to the project.

---

## 🤝 Code of Conduct

Please be respectful and constructive in all interactions with other community members. We're committed to creating a welcoming environment for all contributors.

---

## 🚀 Getting Started

### 1. Fork & Clone

```bash
git clone https://github.com/your-username/flowforge.git
cd flowforge
```

### 2. Setup Development Environment

```bash
make setup-dev
make dev-up
```

### 3. Create a Branch

```bash
git checkout -b feature/my-feature
# or
git checkout -b fix/issue-123
```

---

## 📋 Before You Start

### Check Module Ownership

See [MODULE_BOUNDARIES.md](MODULE_BOUNDARIES.md) to understand:
- **Module responsibilities** (what each module does)
- **Ownership** (which team maintains each module)
- **Dependencies** (what each module depends on)
- **Public interfaces** (which interfaces to implement)

**Key Principle**: Each module has clear ownership. Respect ownership boundaries and discuss cross-module changes.

### Understanding the Monorepo

FlowForge is organized as a **monorepo** with these modules:

```
ir/             → Core IR specification & validation
compiler/       → Parsing, compilation, optimization
runtime/        → Local execution engine
executors/      → Argo, Airflow, Local drivers
storage/        → PostgreSQL, Redis persistence
lineage/        → Data provenance tracking
observability/  → Metrics, logging, tracing
api/            → gRPC + REST server
sdk/            → Python SDK + CLI
ui/             → React dashboard & builder
deployment/     → Terraform, Helm, Docker
tests/          → E2E integration tests
```

**Do NOT**:
- Create circular dependencies (check [MODULE_BOUNDARIES.md](MODULE_BOUNDARIES.md))
- Import `internal/` packages from other modules (use public API in `pkg/`)
- Break existing module contracts/interfaces

---

## 🎯 Contribution Types

### 1. Bug Fix

```bash
# Create issue first (if not exists)
# Then create branch:
git checkout -b fix/issue-123-description

# Make changes
# Write tests
make test

# Commit
git add .
git commit -m "Fix: brief description (fixes #123)"

# Push & create PR
git push origin fix/issue-123-description
```

### 2. Feature in Existing Module

**Example**: Add new optimization pass to compiler

```bash
git checkout -b feature/compiler-new-pass

# Read compiler README & interfaces
cat compiler/README.md

# Implement in appropriate place
# Example: compiler/internal/optimizer/my_pass.go

# Write tests
# Run tests
make test-unit

# Follow naming conventions
# File naming: snake_case.go
# Function naming: PascalCase

git add .
git commit -m "Feature: add my optimization pass"

# Push & create PR
git push origin feature/compiler-new-pass
```

### 3. New Module (Rare)

**Only with architecture team approval**. See [ARCHITECTURE.md](ARCHITECTURE.md) for design principles.

### 4. Documentation

```bash
git checkout -b docs/improve-sdk-docs

# Edit docs/ files
# Test links & formatting

git add docs/
git commit -m "Docs: improve SDK documentation"

git push origin docs/improve-sdk-docs
```

### 5. Infrastructure/Deployment

```bash
git checkout -b infra/upgrade-postgres

# Edit deployment/ files (Terraform, Helm, Docker)
# Test locally:
make dev-up

git add deployment/
git commit -m "Infra: upgrade PostgreSQL to 16"

git push origin infra/upgrade-postgres
```

---

## ✍️ Coding Standards

### Go

- **Format**: `gofmt` (enforced by CI)
- **Lint**: `golangci-lint` (enforced by CI)
- **Tests**: Write unit + integration tests
- **Interfaces**: Define in `pkg/` (public API)
- **Implementations**: Put in `internal/` (private)
- **Naming**: PascalCase for exported, camelCase for unexported
- **Errors**: Return custom error types with context

Example:
```go
// In pkg/compiler.go (public interface)
type Compiler interface {
    Compile(ctx context.Context, spec PipelineSpec) (interface{}, error)
}

// In internal/compiler/compiler.go (implementation)
type compilerImpl struct {
    parsers map[string]Parser
}

func (c *compilerImpl) Compile(ctx context.Context, spec PipelineSpec) (interface{}, error) {
    // implementation
}
```

### Python (SDK)

- **Format**: `black` (enforced by CI)
- **Lint**: `pylint` (enforced by CI)
- **Type hints**: Always use type annotations
- **Tests**: pytest with fixtures
- **Docstrings**: Google-style docstrings

Example:
```python
# In flowforge/pipeline.py
class Pipeline:
    """Fluent API for building pipelines."""
    
    def __init__(self, name: str) -> None:
        """Initialize pipeline.
        
        Args:
            name: Pipeline name
            
        Raises:
            ValueError: If name is empty
        """
        if not name:
            raise ValueError("Pipeline name cannot be empty")
        self.name = name
    
    def add_task(self, task_id: str, image: str) -> "Pipeline":
        """Add a task to the pipeline.
        
        Args:
            task_id: Unique task identifier
            image: Docker image URI
            
        Returns:
            Self for method chaining
        """
        # implementation
        return self
```

### TypeScript/React (UI)

- **Format**: `prettier` (enforced by CI)
- **Lint**: `eslint` (enforced by CI)
- **Type safety**: Strict mode enabled
- **Testing**: Jest + React Testing Library
- **Components**: Functional components + hooks

Example:
```typescript
// In src/components/Builder/Canvas.tsx
interface CanvasProps {
    pipeline: PipelineSpec;
    onTaskMove: (taskId: string, x: number, y: number) => void;
}

export const Canvas: React.FC<CanvasProps> = ({ pipeline, onTaskMove }) => {
    const [selectedTask, setSelectedTask] = React.useState<string | null>(null);
    
    return (
        <div className="canvas">
            {/* render canvas */}
        </div>
    );
};
```

---

## 🧪 Testing Requirements

**All contributions must include tests**.

### Go Modules

```bash
# Unit tests
go test -short ./...

# Integration tests
go test ./...

# Coverage
go test -cover ./...

# Run specific test
go test -run TestMyFeature -v ./internal/compiler
```

### Python SDK

```bash
# Unit tests
cd sdk && pytest tests/unit -v

# Integration tests
cd sdk && pytest tests/integration -v

# Coverage
cd sdk && pytest --cov=flowforge tests/
```

### E2E Tests

```bash
# Full stack test
cd tests && pytest e2e/test_yaml_to_argo.py -v
```

**Coverage Targets**:
- New code: >= 80% line coverage
- Existing code: maintain or improve coverage
- E2E: add happy path + error cases

---

## 📝 Commit Messages

Follow conventional commits format:

```
type(scope): subject

body

footer
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code formatting (no logic change)
- `refactor`: Code restructuring (no logic change)
- `test`: Test additions/updates
- `chore`: Build, CI/CD, dependency updates
- `perf`: Performance improvements
- `infra`: Infrastructure/deployment changes

**Scope**:
- Module name: `compiler`, `sdk`, `api`, etc.
- Or subsystem: `lineage-engine`, `argo-driver`, etc.

**Examples**:
```
feat(compiler): add optimization pass for task merging

Implement sequential task merging to reduce DAG depth.
Improves Argo workflow submission time by ~10%.

Closes #456
```

```
fix(sdk): handle empty pipeline name

Validate pipeline name in __init__ to catch errors early.
```

```
docs(api): update gRPC service documentation
```

---

## 🔄 Pull Request Process

### 1. Before Submitting

- [ ] Run all tests: `make test`
- [ ] Lint code: `make lint`
- [ ] Format code: `make fmt`
- [ ] Update docs if needed
- [ ] Commit messages follow convention
- [ ] No unrelated changes

### 2. PR Description

```markdown
## Description
Brief description of changes

## Motivation
Why are these changes needed?

## Changes
- What was added/modified/removed

## Testing
- How was this tested?
- Add new test cases? Yes/No
- Test coverage: X%

## Module Impact
- Which module(s) are affected?
- Any breaking changes?
- Any dependency changes?

## Screenshots
(if UI changes)

## Checklist
- [ ] Tests pass locally
- [ ] Code formatted & linted
- [ ] Documentation updated
- [ ] Commit messages follow convention
- [ ] No circular dependencies introduced
```

### 3. Review Process

- **Minimum reviewers**: 1 (from module ownership team)
- **Response time**: 24-48 hours
- **Approval**: At least 1 ✓
- **CI must pass**: All checks green
- **No unresolved conversations**: All discussions resolved

### 4. Merge

- Rebase on main (if conflicts)
- Squash if multiple fixup commits
- Merge via "Squash and merge" (clean history)

---

## 🔍 Code Review Guidelines

### As a Reviewer

**Check**:
1. Does it follow module boundaries? (see [MODULE_BOUNDARIES.md](MODULE_BOUNDARIES.md))
2. Are interfaces well-designed?
3. Are tests adequate?
4. Is documentation clear?
5. Performance implications?
6. Security considerations?

**Approve if**:
- Code quality high
- Tests comprehensive
- Follows project standards
- No module boundary violations

### As an Author

**Respond to feedback**:
- Address all comments (don't ignore)
- Ask clarifying questions if confused
- Update code or explain why you disagree
- Push updates & re-request review

---

## 🐛 Reporting Bugs

Use [GitHub Issues](https://github.com/your-org/flowforge/issues/new/choose):

1. **Title**: Clear, specific
   - ❌ "Bug in compiler"
   - ✓ "Compiler crashes on cyclic DAG"

2. **Description**:
   - What did you do?
   - What did you expect?
   - What actually happened?
   - Error messages/logs

3. **Environment**:
   - OS
   - Go/Python/Node version
   - Module version

4. **Reproduction**:
   - Minimal example
   - Steps to reproduce

---

## 💡 Feature Requests

Use [GitHub Discussions](https://github.com/your-org/flowforge/discussions):

1. **Title**: Clear, descriptive
2. **Use Case**: Why do you need this?
3. **Proposed Solution**: How should it work?
4. **Alternatives**: What else could work?

---

## 📚 Documentation Guidelines

### Adding Documentation

- **Location**: Appropriate folder in `docs/`
- **Format**: Markdown
- **Links**: Relative paths, not absolute URLs
- **Code Examples**: Runnable, tested

### Example Documentation

```markdown
# Feature Name

Brief overview.

## Use Cases

When would you use this?

## Example

\`\`\`bash
# Command example
ff submit pipeline.yaml
\`\`\`

## See Also

- [Related Feature](./related.md)
- [API Reference](../api/reference.md)
```

---

## ⚙️ Development Tips

### Module-Specific Development

```bash
# Work on compiler only
cd compiler
go test ./...
make lint
make fmt

# Work on SDK only
cd sdk
pytest tests/ -v
black flowforge
pylint flowforge

# Work on UI only
cd ui
npm test
npm run lint
npm run format
```

### Local Testing

```bash
# Test specific feature
go test -run TestFeatureName -v ./internal/module

# Test with debug output
GODEBUG=... go test -v ./...

# Test with coverage
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Docker Testing

```bash
# Build and test in Docker
docker-compose up -d
docker-compose exec api go test ./...
docker-compose down
```

---

## 🚨 Important: Module Boundaries

**Forbidden**:
```go
// ❌ Don't import internal packages from other modules
import "flowforge/executors/internal/argo"  // NO!
import "flowforge/compiler/internal/codegen" // NO!

// ✓ Import public API only
import "flowforge/executors/pkg"
import "flowforge/compiler/pkg"
```

**Allowed**:
```go
// ✓ Public API imports
driver := executors.GetDriver("argo")
config, err := compiler.Compile(spec)

// ✓ Shared interfaces
func (c *CustomPass) Optimize(ctx context.Context, pipeline ir.PipelineSpec) (ir.PipelineSpec, error) {
    // implementation
}
```

---

## 📞 Questions?

- **Docs**: Check [docs/development/](docs/development/)
- **Issues**: Search [GitHub Issues](https://github.com/your-org/flowforge/issues)
- **Discussions**: Post in [GitHub Discussions](https://github.com/your-org/flowforge/discussions)
- **Slack**: Join [Community Slack](https://flowforge-community.slack.com)

---

## 🎉 Thank You!

Your contributions help make FlowForge better. We appreciate your effort!

