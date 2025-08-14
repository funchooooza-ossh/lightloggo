"""
Internal C-level wrapper for the Stdout Writer resource.

This module provides the CStdoutWriter class, which interfaces with the FFI
layer to create a resource that writes log messages to standard output.
"""

from ..._ffi._ffi import lib


class CStdoutWriter:
    """
    A low-level handle to a stdout writer resource in the Go core.

    This is a simple internal wrapper that calls the FFI function to get a
    handle to the native stdout writer. It is used by higher-level API
    classes.

    Attributes:
        _id (int): The unique identifier/handle for the native resource.
    """

    def __init__(self) -> None:
        """Initializes CStdoutWriter by creating a new resource via FFI."""
        # Call the FFI function to get the handle for the singleton-like
        # stdout writer instance within the Go core.
        self._id = lib.NewStdoutWriter()
