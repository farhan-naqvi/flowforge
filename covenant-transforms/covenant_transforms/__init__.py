"""covenant-transforms — the versioned primitive library that is Covenant's core.

Primitives are declarative, schema-inferring, Ibis-lowered transforms. This
package is designed to live in its own semver-versioned repository; contracts
pin a version for reproducibility. It is vendored here as a self-contained
package during early development.
"""

from __future__ import annotations

from . import primitives  # noqa: F401 - registers the primitive set
from .primitive import Primitive, all_ids, get, register
from .schema import DTYPES, Field, Schema

__version__ = "0.1.0"

__all__ = ["Primitive", "get", "register", "all_ids", "Schema", "Field", "DTYPES", "__version__"]
