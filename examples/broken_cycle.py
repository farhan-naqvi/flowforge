"""A deliberately broken pipeline used to demonstrate validation.

`a` depends on `b` and `b` depends on `a`, forming a cycle. `ff validate`
reports it and `ff compile` refuses to emit an artifact.
"""

from flowforge import pipeline, task


@task
def a(b):
    return b + 1


@task
def b(a):
    return a + 1


broken = pipeline("broken-cycle", tasks=[a, b], owner="qa")
