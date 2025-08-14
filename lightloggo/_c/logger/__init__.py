"""
Internal C-level logger package.

This package contains the primary C-level wrapper for the logger resource
managed by the Go core library.
"""

from .base import CLogger

# Explicitly define the internal API of this package.
__all__ = ["CLogger"]
