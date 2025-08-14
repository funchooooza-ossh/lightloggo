"""
Internal C-level wrapper for the JSON Formatter resource.

This module provides the CJsonFormatter class, which directly interfaces
with the FFI layer to create and manage a JSON formatter instance within
the Go core library.
"""

from ..._ffi._ffi import lib


class CJsonFormatter:
    """
    A low-level handle to a JSON formatter resource in the Go core.

    This class is not intended for direct use by end-users. It serves as an
    internal container for the ID (handle) of a formatter resource created
    by the native library. Higher-level classes in the public API will use
    this object to perform formatting operations.

    The instance's primary attribute, `_id`, holds the memory address or
    handle that uniquely identifies the formatter in the Go runtime.

    Attributes:
        _id (int): The unique identifier/handle for the native resource.
    """

    def __init__(self, style_id: int = 0, max_depth: int = 3) -> None:
        """
        Initializes the CJsonFormatter by creating a new resource via FFI.

        Args:
            style_id (int, optional): The ID of a pre-configured CFormatStyle
                resource. Defaults to 0, indicating a default style.
            max_depth (int, optional): The maximum depth for serializing
                complex data structures. Defaults to 3.
        """
        # Call the FFI function to create a new JSON formatter instance
        # in the Go core. The returned value is a handle (ID) to that resource,
        # which we store for all subsequent interactions.
        self._id = lib.NewJsonFormatter(style_id, max_depth)
