"""
Provides the FileWriter for directing log output to a file.
"""

from .._c import CFileWriter
from ._base import writer


class FileWriter(writer):
    """
    Writes log records to a file with built-in rotation capabilities.

    This writer is ideal for persistent logging in production environments.
    It automatically handles log rotation based on file size and time
    intervals, and can compress archived log files to save space.

    Example:
        >>> # Log to a file, rotating daily or when it reaches 20MB.
        >>> file_writer = FileWriter(
        ...     path="/var/log/my_app.log",
        ...     max_size_mb=20,
        ...     interval="day"
        ... )
        >>> logger.add_route(writer=file_writer, formatter=JsonFormatter())
    """

    def __init__(
        self,
        path: str,
        max_size_mb: int = 10,
        max_backups: int = 5,
        interval: str = "day",
        compress: str = "gz",
    ) -> None:
        """
        Initializes the FileWriter.

        Args:
            path (str): The path to the log file.
            max_size_mb (int, optional): The maximum size in megabytes before
                the log file is rotated. Defaults to 10.
            max_backups (int, optional): The maximum number of old, rotated
                log files to retain. Defaults to 5.
            interval (str, optional): The time-based rotation interval.
                Valid options: "day", "week", "month". Defaults to "day".
            compress (str, optional): The compression method for rotated logs.
                Valid options: "gz" or "". Defaults to "gz".
        """
        # Instantiate the internal C-level writer, which calls the FFI
        # to create the corresponding resource with all its configuration
        # in the Go core.
        self._c_writer = CFileWriter(
            path=path,
            max_backups=max_backups,
            max_size_mb=max_size_mb,
            interval=interval,
            compress=compress,
        )
