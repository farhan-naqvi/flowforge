"""Load a live :class:`~flowforge.authoring.Pipeline` from a user module.

Used only by ``ff run`` (which must execute real callables). ``ff compile`` and
``ff validate`` use the static AST parser instead and never import user code.
"""

from __future__ import annotations

import importlib.util
import os
from typing import Optional

from .authoring import Pipeline


class LoadError(RuntimeError):
    pass


def load_pipeline(path: str, name: Optional[str] = None) -> Pipeline:
    if not os.path.isfile(path):
        raise LoadError(f"file not found: {path}")
    spec = importlib.util.spec_from_file_location("_flowforge_user_module", path)
    if spec is None or spec.loader is None:
        raise LoadError(f"cannot import {path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)  # NOTE: executes user code (see runner trust boundary)

    candidates = {
        k: v for k, v in vars(module).items() if isinstance(v, Pipeline)
    }
    if not candidates:
        raise LoadError(f"no flowforge.Pipeline object found in {path}")
    if name:
        if name not in candidates:
            raise LoadError(f"pipeline variable {name!r} not found in {path}")
        return candidates[name]
    if len(candidates) == 1:
        return next(iter(candidates.values()))
    raise LoadError(
        f"multiple pipelines in {path}: {', '.join(candidates)}; choose one with --name"
    )
