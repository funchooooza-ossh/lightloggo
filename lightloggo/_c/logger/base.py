"""
Internal C-level wrapper for the core Logger resource.

This module provides the CLogger class, which is the primary C-level handle
for a fully configured logger instance in the Go core. It aggregates one or
more route processors to create a complete logging pipeline.
"""

from __future__ import annotations

import contextlib
import ctypes
from typing import List

from ..._ffi._ffi import lib


class CLogger:
    """
    A low-level handle to a native logger instance in the Go core.

    This class is a core internal component and is not for direct use. It
    represents a complete logger object constructed from a set of routing
    rules (`CRouteProcessor` IDs). It is the primary entity against which
    log write operations are ultimately dispatched.

    This class also manages the lifecycle of the native logger resource,
    ensuring it is properly shut down and freed when no longer in use.

    Attributes:
        _id (int): The unique identifier/handle for the native logger resource.
                   A value of 0 indicates the resource has been freed.
    """

    def __init__(self, routes: List[int]) -> None:
        """
        Initializes CLogger by creating a new logger resource from routes.

        This constructor takes a list of CRouteProcessor IDs, converts them
        into a C-compatible array, and passes them to the FFI to construct
        the new logger instance in the Go core.

        Args:
            routes (List[int]): A list of integer IDs from pre-configured
                CRouteProcessor instances.
        """
        # Create a ctypes array type that is correctly sized for the list of
        # route IDs. `ctypes.c_ulong` is used to match the `uintptr_t`
        # type used for IDs in the Go/C layer.
        arr_type = ctypes.c_ulong * len(routes)

        # Instantiate the array type with the values from the Python list.
        # The `*` operator unpacks the list into arguments for the constructor.
        route_ids = arr_type(*routes)

        # Call the FFI function with a pointer to the C array and its length
        # to create the native logger resource.
        self._id = lib.NewLoggerWithRoutes(route_ids, len(routes))

    def close(self) -> None:
        """
        Gracefully shuts down and frees the native logger resource.

        This method should be called to ensure that all buffered log messages
        are flushed and all memory allocated by the Go core for this logger
        is released. It makes the logger instance unusable thereafter.
        """
        # Ensure the resource has not already been freed.
        if self._id:
            # First, call the graceful shutdown function. This is assumed to
            # handle tasks like flushing file writers. (Note: A Logger_Close
            # function must be bound in the FFI layer for this to work).
            if hasattr(lib, "Logger_Close"):
                lib.Logger_Close(self._id)

            # Second, free the memory associated with the logger resource.
            lib.FreeLogger(self._id)

            # Nullify the ID to prevent double-freeing or use-after-free bugs.
            # This defensively marks the Python object as closed.
            self._id = 0

    def __del__(self) -> None:
        """
        A fail-safe finalizer to ensure native resources are released.

        This method is called by the Python garbage collector when the CLogger
        object is about to be destroyed. It attempts to call `close()` to
        prevent memory leaks in the Go core.
        """
        # The __del__ method must never raise an exception. We suppress all
        # potential errors (e.g., if the library was already unloaded during
        # interpreter shutdown) to ensure a safe and clean exit.
        with contextlib.suppress(Exception):
            self.close()
