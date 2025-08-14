"""
Provides public classes for defining logging routes.

A "route" is the core concept that connects all the pieces of the logging
pipeline. It defines a rule that states: "log messages at or above a certain
`level` should be processed by a specific `formatter` and sent to a
particular `writer`."

By combining multiple routes, a user can create sophisticated logging
configurations (e.g., sending errors to a file and info-level messages
to the console).
"""

from .._c import CRouteProcessor
from ..enums import LogLevel
from ..formatters import Formatter, TextFormatter
from ..writers import StdoutWriter, Writer
from ._base import router


class RouteProcessor(router):
    """
    Defines a single, complete pipeline for processing log messages.

    This class is the standard implementation of a route. It takes a writer,
    a formatter, and a minimum log level, and combines them into a single
    routing rule that can be added to a logger.

    Thanks to its sensible defaults, it can be instantiated without arguments
    to create a simple console logger.

    Example:
        >>> # A simple route to log INFO messages to the console.
        >>> console_route = RouteProcessor()
        >>>
        >>> # A more complex route to log WARNING and above to a JSON file.
        >>> from lightloggo.enums import LogLevel
        >>> from lightloggo.writers import FileWriter
        >>> from lightloggo.formatters import JsonFormatter
        >>>
        >>> file_route = RouteProcessor(
        ...     writer=FileWriter("app_errors.log"),
        ...     formatter=JsonFormatter(),
        ...     level=LogLevel.Warning
        ... )
        >>>
        >>> # logger.add_route(console_route)
        >>> # logger.add_route(file_route)
    """

    def __init__(
        self,
        formatter: Formatter = None,
        writer: Writer = None,
        level: LogLevel = LogLevel.Info,
    ) -> None:
        """
        Initializes the RouteProcessor.

        Args:
            formatter (Formatter, optional): The formatter instance to apply
                to log records. Defaults to a new `TextFormatter()`.
            writer (Writer, optional): The writer instance that determines the
                log output destination. Defaults to a new `StdoutWriter()`.
            level (LogLevel, optional): The minimum log level required for
                this route to be activated. Defaults to `LogLevel.Info`.
        """
        # Provide user-friendly defaults if none are specified.
        if formatter is None:
            formatter = TextFormatter()
        if writer is None:
            writer = StdoutWriter()

        # Instantiate the internal C-level router. This call crosses the FFI
        # boundary, passing the unique IDs from the public formatter and
        # writer objects to create the native resource in the Go core.
        self._c_router = CRouteProcessor(formatter.id, writer.id, level)


# Defines the public API of the routers package.
__all__ = [
    "RouteProcessor",
    "router",
]
