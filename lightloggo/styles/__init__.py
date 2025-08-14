"""
Provides public classes for style configuration.

This package allows users to define and manage the visual styling applied
to log records by the formatters.
"""

from ._base import BaseStyle, FormatStyle

# Defines the public API of the styles package.
__all__ = [
    "BaseStyle",
    "FormatStyle",
]
