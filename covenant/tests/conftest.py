import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
sys.path.insert(0, ROOT)                                  # covenant/
sys.path.insert(0, os.path.join(ROOT, "covenant-transforms"))

BASE_DIR = ROOT
