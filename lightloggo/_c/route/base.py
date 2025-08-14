"""
Internal C-level wrapper for the Route Processor resource.

This module provides the CRouteProcessor class, which interfaces with the FFI
layer. It creates and manages a "route" instance—a fundamental routing rule
that links a minimum log level to a specific formatter and writer pair.
"""

from ..._ffi._ffi import lib
from ...enums import LogLevel


class CRouteProcessor:
    """
    A low-level handle to a route processor resource in the Go core.

    This class is an internal component and not intended for direct use.
    It represents a single, complete processing pipeline for a log message:
    if a message's level is at or above the specified threshold, it will be
    processed by the given formatter and sent to the given writer.

    Multiple CRouteProcessor instances can be combined to create a complex
    logger with multiple outputs (e.g., INFO to stdout, ERROR to a file).

    Attributes:
        _id (int): The unique identifier/handle for the native route resource.
    """

    def __init__(
        self, formatter_id: int, writer_id: int, level: LogLevel = 20
    ) -> None:  # The default level of 20 corresponds to INFO.
        """
        Initializes CRouteProcessor by creating a new route resource via FFI.

        Args:
            formatter_id (int): The ID handle of a pre-configured CFormatter
                resource.
            writer_id (int): The ID handle of a pre-configured CWriter
                resource.
            level (LogLevel): The minimum log level required for this route
                to be activated. Defaults to 20 (INFO).
        """
        # Call the FFI function to create a new route processor resource
        # in the Go core. This resource internally links the three components
        # (level, formatter, writer) by their respective IDs.
        self._id = lib.NewRouteProcessor(formatter_id, writer_id, level)
