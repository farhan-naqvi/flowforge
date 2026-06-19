"""Pytest configuration."""

import sys
from pathlib import Path

# Add SDK to path
sdk_path = Path(__file__).parent
sys.path.insert(0, str(sdk_path))
