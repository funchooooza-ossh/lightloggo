"""
Public enumerations for the lightloggo library.

This package exposes enumeration classes used for configuration and control
of the logging system, such as log levels.
"""

from .level import Level, LogLevel

# Defines the public API of the enums package, making Level and LogLevel
# directly available for import via `from lightloggo.enums import ...`.
__all__ = [
    "Level",
    "LogLevel",
]
