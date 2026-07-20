import os
import sys

# ensure the repository root is on sys.path so `import src.*` works during pytest collection
ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
if ROOT not in sys.path:
    sys.path.insert(0, ROOT)
