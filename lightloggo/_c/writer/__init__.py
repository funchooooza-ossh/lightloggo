"""
Internal C-level writers package.

This package contains the low-level Python wrappers that directly correspond
to writer resources (e.g., file, stdout) managed by the Go core library.

It also provides the internal-facing type alias, `CWriter`.
"""

from typing import Union

from .file import CFileWriter
from .stdout import CStdoutWriter

# A type alias representing any of the available low-level C-wrapper
# writer classes. This is used for internal type checking to ensure that
# higher-level components are interacting with a valid C-level handle.
CWriter = Union[CFileWriter, CStdoutWriter]
