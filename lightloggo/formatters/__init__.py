"""
Provides public formatter classes to control log record output format.

Formatters are responsible for taking a log record and converting it into a
string (or byte string) that is ultimately written by a Writer. This package
offers standard formatters like `TextFormatter` and `JsonFormatter`.
"""

from typing import Union

from .json import JsonFormatter
from .text import TextFormatter

# A type alias for use in type hints, representing any available public
# formatter class.
Formatter = Union[TextFormatter, JsonFormatter]

# Defines the public API of the formatters package.
__all__ = [
    "Formatter",
    "JsonFormatter",
    "TextFormatter",
]
