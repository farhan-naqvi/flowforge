# Contributing to FlowForge

Thanks for your interest. FlowForge is an early, honest project — contributions
that keep it that way (real code, real tests, no overstated claims) are welcome.

## Ground rules

- The product is the top-level [`flowforge/`](flowforge) Python package. Anything
  under [`experimental/`](experimental/README.md) is quarantined and out of scope
  unless a change explicitly promotes it (with tests) into the product.
- Every change to product code needs a test. We do not claim a coverage percentage
  we do not measure in CI.
- Keep the canonical path single: all surfaces (CLI, API, UI) call the same
  functions in `flowforge/`. Don't add a second parser/compiler.
- Read [docs/architecture/DECISION_LOG.md](docs/architecture/DECISION_LOG.md)
  before proposing architectural changes.

## Setup

Requires Python 3.11+ (tested 3.11–3.13) and, for the Go schema-of-record, Go 1.22+.

```bash
pip install -e ".[dev]"     # or: make setup
make test                   # Python test suite (pytest)
make test-go                # Go IR (ir/) build + tests
make smoke                  # validate + compile the examples
make demo                   # guided end-to-end walkthrough
```

## Workflow

1. Branch from `main` (e.g. `feature/…` or `fix/…`).
2. Make the change with tests. Run `make test && make test-go && make smoke`.
3. Keep commits small and descriptive.
4. Open a PR describing what changed and how you verified it. CI runs the Python
   matrix, the Go IR tests, and security scans (see
   [.github/workflows/ci.yml](.github/workflows/ci.yml)).

## Adding a compile target

A new target is a data entry plus one compiler — not a rewrite:

1. Add a capability table for the target in [`flowforge/capability.py`](flowforge/capability.py)
   (`supported` / `lossy` / `unsupported` per feature, with a note).
2. Add `flowforge/compilers/<target>.py` with a `compile(spec) -> str` that emits
   through a **safe serializer** (never f-string interpolation of user values —
   see the Argo/Airflow compilers and [the threat model](docs/security/THREAT_MODEL.md)).
3. Register it in [`flowforge/compilers/__init__.py`](flowforge/compilers/__init__.py).
4. Add tests: valid output, determinism, and no-injection.

## Security

Report suspected vulnerabilities privately rather than in a public issue. See
[docs/security/THREAT_MODEL.md](docs/security/THREAT_MODEL.md) for the trust model
and what the project must never claim.

## License

By contributing you agree your contributions are licensed under the
[Apache License 2.0](LICENSE).
