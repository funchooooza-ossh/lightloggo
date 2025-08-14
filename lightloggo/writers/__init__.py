"""
Provides public writer classes that determine the final destination of logs.

Writers are the final step in the logging pipeline, responsible for taking
the formatted log string and writing it to a destination, such as the
console (`StdoutWriter`) or a file (`FileWriter`).
"""

from typing import Union

from .file import FileWriter
from .stdout import StdoutWriter

# A type alias for use in type hints, representing any available public
# writer class.
Writer = Union[FileWriter, StdoutWriter]

# Defines the public API of the writers package.
__all__ = [
    "FileWriter",
    "StdoutWriter",
    "Writer",
]
