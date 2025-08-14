"""
Internal C-level formatters package.

This package contains the low-level Python wrappers that directly correspond
to formatter resources managed by the Go core library.

It also provides an internal-facing type alias, `CFormatter`, for use in
type hinting within the lightloggo package.
"""

from typing import Union

from .json import CJsonFormatter
from .text import CTextFormatter

# A type alias representing any of the available low-level C-wrapper
# formatter classes. This is used for internal type checking to ensure
# that higher-level components are interacting with a valid C-level handle.
CFormatter = Union[CJsonFormatter, CTextFormatter]
