"""
Internal C-level wrapper for the File Writer resource.

This module provides the CFileWriter class, which directly interfaces
with the FFI layer to create and manage a file writer instance within
the Go core library. This instance handles log rotation, sizing, and
compression.
"""

import ctypes

from ..._ffi._ffi import lib


class CFileWriter:
    """
    A low-level handle to a file writer resource in the Go core.

    This class is an internal component and not intended for direct use. It
    manages the lifecycle of a native file writer, which includes features
    like log rotation based on size, time, and automatic compression.

    The instance holds the ID returned by the FFI call, which is used to
    identify the writer resource in all subsequent operations.

    Attributes:
        _id (int): The unique identifier/handle for the native resource.
        _path (bytes): The encoded path to the log file.
        _interval (bytes): The encoded rotation interval string.
        _compress (bytes): The encoded compression type string.
    """

    def __init__(
        self,
        path: str,
        max_size_mb: int = 10,
        max_backups: int = 5,
        interval: str = "day",  # "day", "week", "month"
        compress: str = "gz",  # "gz" or ""
    ) -> None:
        """
        Initializes CFileWriter by creating a new file writer resource via FFI.

        Args:
            path (str): The path to the log file.
            max_size_mb (int, optional): The maximum size in megabytes before
                the log file is rotated. Defaults to 10.
            max_backups (int, optional): The maximum number of old log files
                to retain. Defaults to 5.
            interval (str, optional): The time-based rotation interval.
                Valid options are "day", "week", "month". Defaults to "day".
            compress (str, optional): The compression method for rotated logs.
                Valid options are "gz" or "". Defaults to "gz".
        """
        # Encode string arguments to bytes for the C interface.
        # Storing them as instance attributes prevents them from being
        # garbage-collected before the FFI call completes.
        self._path = path.encode()
        self._interval = interval.encode()
        self._compress = compress.encode()

        # Call the FFI constructor. We explicitly cast arguments to their
        # C types to ensure correctness and prevent ambiguity.
        self._id = lib.NewFileWriter(
            ctypes.c_char_p(self._path),
            ctypes.c_long(max_size_mb),
            ctypes.c_int(max_backups),
            ctypes.c_char_p(self._interval),
            ctypes.c_char_p(self._compress),
        )

        # A returned ID of 0 from the Go core indicates a failure
        # during resource creation (e.g., invalid path, permissions error).
        if not self._id:
            raise RuntimeError(f"Failed to create native FileWriter for path: {path}")
