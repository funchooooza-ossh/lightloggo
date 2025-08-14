"""
Internal C-level routing package.

This package contains the low-level Python wrappers that directly correspond
to routing resources managed by the Go core library.

It also provides the internal-facing type alias, `CRouter`.
"""

from typing import Union

from .base import CRouteProcessor

# A type alias representing any available low-level C-wrapper for a routing
# resource. This is used for internal type hinting. The Union is used to
# accommodate potential future routing classes.
CRouter = Union[CRouteProcessor]
