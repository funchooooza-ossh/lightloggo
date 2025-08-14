"""
Internal C-level style configuration package.

This package contains the low-level Python wrappers that directly correspond
to styling resources managed by the Go core library.

It also provides the internal-facing type alias, `CStyle`.
"""

from typing import Union

from .base import CFormatStyle

# A type alias representing the low-level C-wrapper for a style resource.
# This is used for internal type hinting to ensure that higher-level
# components are passed a valid C-level handle. The Union is used to
# accommodate potential future style classes.
CStyle = Union[CFormatStyle]
