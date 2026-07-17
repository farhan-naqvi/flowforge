# FlowForge Python SDK Examples

This directory contains example pipelines demonstrating FlowForge SDK capabilities.

## Quick Start

### Run Simple ETL

```bash
python simple_etl.py
```

This demonstrates:
- Pipeline definition with decorators
- Task composition
- IR export
- Local execution

### Run Fan-Out/Fan-In

```bash
python fan_out_fan_in.py
```

This demonstrates:
- Parallel processing patterns
- Multiple input/output paths
- Task merging

### Run Conditional Pipeline

```bash
python conditional_pipeline.py
```

This demonstrates:
- Conditional branching (basic)
- Data validation
- Multiple output handling

## CLI Usage

### Validate Pipeline

```bash
flowforge validate simple_etl.py
```

### Compile to IR

```bash
flowforge compile simple_etl.py -o pipeline.json
```

### Inspect Pipeline

```bash
flowforge inspect simple_etl.py
```

### Run Locally

```bash
flowforge run-local simple_etl.py
```

## Pipeline Patterns

### Linear (ETL)
```
Extract → Transform → Load
```

### Fan-Out/Fan-In (Parallel)
```
Source → [ProcessA, ProcessB] → Merge
```

### Multi-Branch
```
Extract → {Branch1, Branch2, Branch3} → Combine
```

## Next Steps

- Create custom decorators
- Integrate with external systems
- Define schemas explicitly
- Add monitoring and observability
