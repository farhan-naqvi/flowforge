"""Fan-out / fan-in: one source feeds two branches that join at a sink.

Demonstrates non-linear DAGs and parallel branches in the capability report.
"""

from flowforge import pipeline, task


@task
def source():
    return list(range(10))


@task
def branch_even(source):
    return [x for x in source if x % 2 == 0]


@task
def branch_odd(source):
    return [x for x in source if x % 2 == 1]


@task
def join(branch_even, branch_odd):
    print(f"evens={len(branch_even)} odds={len(branch_odd)}")
    return {"evens": branch_even, "odds": branch_odd}


analytics = pipeline(
    "fan-out-fan-in",
    tasks=[source, branch_even, branch_odd, join],
    owner="analytics",
    version="1.0.0",
)
