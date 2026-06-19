"""FlowForge IR graph operations for Python."""

from typing import List, Dict, Optional, Set
from collections import defaultdict, deque
from .spec import PipelineSpec


class DAGGraph:
    """Represents the pipeline as a directed acyclic graph."""

    def __init__(self, spec: PipelineSpec):
        """Initialize graph from pipeline spec."""
        self.spec = spec
        self.adj: Dict[str, List[str]] = defaultdict(list)
        self.radj: Dict[str, List[str]] = defaultdict(list)

        # Initialize vertices
        for task_id in spec.tasks:
            self.adj[task_id] = []
            self.radj[task_id] = []

        # Add edges
        for edge in spec.edges:
            self.adj[edge.from_task.task].append(edge.to_task.task)
            self.radj[edge.to_task.task].append(edge.from_task.task)

        # Sort for consistency
        for neighbors in self.adj.values():
            neighbors.sort()
        for neighbors in self.radj.values():
            neighbors.sort()

    def nodes(self) -> List[str]:
        """Return all task IDs."""
        return sorted(self.spec.tasks.keys())

    def successors(self, task_id: str) -> List[str]:
        """Return tasks that this task feeds into."""
        return self.adj.get(task_id, [])

    def predecessors(self, task_id: str) -> List[str]:
        """Return tasks that feed into this task."""
        return self.radj.get(task_id, [])

    def topological_sort(self) -> tuple[List[str], Optional[str]]:
        """Return tasks in topological order. Returns (sorted_list, error_msg)."""
        # Kahn's algorithm
        in_degree = {task_id: 0 for task_id in self.spec.tasks}

        for task_id, successors in self.adj.items():
            for succ in successors:
                in_degree[succ] += 1

        queue = deque([task_id for task_id in self.spec.tasks if in_degree[task_id] == 0])
        queue = deque(sorted(queue))

        result = []
        while queue:
            task_id = queue.popleft()
            result.append(task_id)

            for succ in sorted(self.adj[task_id]):
                in_degree[succ] -= 1
                if in_degree[succ] == 0:
                    queue.append(succ)

            queue = deque(sorted(queue))

        if len(result) != len(self.spec.tasks):
            cycle = self.get_cycle()
            return [], f"Cycle detected: {cycle}"

        return result, None

    def has_cycle(self) -> bool:
        """Check if graph has cycles."""
        _, error = self.topological_sort()
        return error is not None

    def get_cycle(self) -> Optional[List[str]]:
        """Return a cycle if one exists."""
        visited: Set[str] = set()
        rec_stack: Set[str] = set()
        path: List[str] = []

        def dfs(node: str) -> Optional[List[str]]:
            visited.add(node)
            rec_stack.add(node)
            path.append(node)

            for neighbor in self.adj[node]:
                if neighbor not in visited:
                    result = dfs(neighbor)
                    if result:
                        return result
                elif neighbor in rec_stack:
                    # Found cycle
                    for i, n in enumerate(path):
                        if n == neighbor:
                            return path[i:] + [neighbor]
                    return None

            rec_stack.discard(node)
            path.pop()
            return None

        for task_id in self.spec.tasks:
            if task_id not in visited:
                result = dfs(task_id)
                if result:
                    return result

        return None

    def __str__(self) -> str:
        """String representation."""
        s = f"DAG with {len(self.spec.tasks)} nodes and {len(self.spec.edges)} edges:\n"
        for edge in self.spec.edges:
            s += f"  {edge.from_task.task}.{edge.from_task.port} -> {edge.to_task.task}.{edge.to_task.port}\n"
        return s
