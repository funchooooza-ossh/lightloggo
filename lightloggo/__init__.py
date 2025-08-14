"""
lightloggo: A high-performance, structured logger for Python with a Go core.

This library provides a fast, flexible, and thread-safe logging solution designed
for modern applications.

Key Features:
- **High Performance:** Core logging operations are handled by a compiled Go
  backend for minimal overhead.
- **Structured Logging:** Log messages with rich, key-value context, ideal for
  machine-readable formats like JSON.
- **Flexible Routing:** Easily direct logs to different destinations (console,
  files) with different formats and level requirements using a powerful
  routing system.
- **Automatic Context:** Automatically inject caller scope (file, line, function)
  or full tracebacks for easy debugging.
- **Thread-Safe:** The global logger instance is designed for safe use in
  multi-threaded applications.

Basic Usage:
    A pre-configured global logger is available for immediate use.

    >>> import lightloggo
    >>>
    >>> lightloggo.logger.info("Application has started.")
    >>> lightloggo.logger.warning("Cache is nearing capacity.", usage_percent=95)

Configuration Example:
    Easily reconfigure the logger by importing component classes.

    >>> from lightloggo import logger, RouteProcessor, FileWriter, JsonFormatter
    >>>
    >>> # Define a route to write JSON logs to a file for WARNING level and up
    >>> file_route = RouteProcessor(
    ...     writer=FileWriter("app.log"),
    ...     formatter=JsonFormatter(),
    ...     level=lightloggo.LogLevel.Warning
    ... )
    >>>
    >>> # Apply the new configuration
    >>> logger.configure(routes=[file_route])
    >>>
    >>> logger.error("This will be written to app.log in JSON format.")
"""

from .enums import LogLevel
from .formatters import JsonFormatter, TextFormatter
from .logger._global import Logger
from .routers import RouteProcessor
from .styles import FormatStyle
from .writers import FileWriter, StdoutWriter

# The default, pre-configured, global logger instance.
# This instance is thread-safe and ready to use immediately upon import.
# It initially logs INFO-level messages and higher to the console.
logger: Logger = Logger()


# Defines the complete public API of the lightloggo library.
# These are the names accessible via `from lightloggo import ...`.
__all__ = [  # noqa: RUF022
    # --- Component Classes ---
    "FileWriter",
    "FormatStyle",
    "JsonFormatter",
    "RouteProcessor",
    "StdoutWriter",
    "TextFormatter",
    # --- Enumerations & Type Aliases ---
    "LogLevel",
    # --- Logger Objects ---
    "logger",  # The pre-configured, global logger instance.
]
