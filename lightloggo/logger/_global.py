"""
Provides the primary, public-facing, thread-safe Logger class.

This module contains the main `Logger` object that users will typically
import and interact with throughout their application.
"""

import contextlib
import threading
from typing import Any, List

from ..routers import RouteProcessor
from .logger import _Logger, create_default_logger


class Logger:
    """
    The main public-facing, thread-safe logger object.

    This class acts as a high-level manager for the logging configuration.
    It ensures that all operations, especially configuration changes, are
    performed in a thread-safe manner.

    When the configuration is changed (e.g., via `.configure()`), this class
    handles the safe shutdown of the old logging engine and the creation of
    a new one with the updated settings.

    Example:
        >>> from lightloggo import Logger
        >>> from lightloggo.routers import RouteProcessor
        >>> from lightloggo.writers import FileWriter
        >>>
        >>> # Get the global logger instance
        >>> log = Logger()
        >>>
        >>> # Log a message using the default configuration (logs to console)
        >>> log.info("Application starting up.")
        >>>
        >>> # Reconfigure the logger to write to a file
        >>> file_route = RouteProcessor(writer=FileWriter("app.log"))
        >>> log.configure(routes=[file_route])
        >>>
        >>> log.warning("Configuration updated to file logging.")
    """

    def __init__(self) -> None:
        """Initializes the thread-safe logger manager."""
        self._lock = threading.Lock()
        # The managed logger instance. It's replaced during reconfiguration.
        self._logger = self._create_default_logger()

    def _create_default_logger(self) -> _Logger:
        """Creates a logger with a single, default route."""
        return create_default_logger()

    # --- Public Logging Methods ---

    def info(self, msg: str, **kwargs: Any) -> None:
        """Logs a message with the INFO level."""
        self._logger.info(msg, **kwargs)

    def debug(self, msg: str, **kwargs: Any) -> None:
        """Logs a message with the DEBUG level."""
        self._logger.debug(msg, **kwargs)

    def warning(self, msg: str, **kwargs: Any) -> None:
        """Logs a message with the WARNING level."""
        self._logger.warning(msg, **kwargs)

    def error(self, msg: str, **kwargs: Any) -> None:
        """Logs a message with the ERROR level."""
        self._logger.error(msg, **kwargs)

    def exception(self, msg: str, **kwargs: Any) -> None:
        """Logs a message with the EXCEPTION level."""
        self._logger.exception(msg, **kwargs)

    def trace(self, msg: str, **kwargs: Any) -> None:
        """Logs a message with the TRACE level."""
        self._logger.trace(msg, **kwargs)

    # --- Configuration Methods ---

    def add(self, route: RouteProcessor) -> None:
        """
        Reconfigures the logger to use only a single new route.

        Warning: This method is destructive. It will close and replace the
        entire existing logger configuration with a new one containing
        only the provided route. For adding routes non-destructively,
        it is recommended to manage a list of routes externally and use
        the `.configure()` method.

        Args:
            route: The single `RouteProcessor` to use for the new configuration.
        """
        with self._lock:
            # This is a "close and replace" operation.
            self._logger.close()
            self._logger = _Logger(routes=[route])

    def remove(self) -> None:
        """
        Resets the logger to its default initial configuration.

        This is a destructive operation that closes the current logger
        and replaces it with a new one that logs INFO-level messages and
        above to the console.
        """
        with self._lock:
            self._logger.close()
            self._logger = self._create_default_logger()

    def configure(
        self,
        routes: List[RouteProcessor],
        tb: bool = False,
        tb_max_depth: int = 10,
        tb_level: int = 50,
        scope: bool = True,
    ) -> None:
        """
        Applies a new, complete configuration to the logger.

        This is the primary method for setting up a custom logging
        configuration. It is a destructive operation that safely closes the
        existing logger and initializes a new one with the specified settings.

        Args:
            routes: A list of `RouteProcessor` objects defining all desired
                logging pipelines.
            tb (bool, optional): Enable automatic traceback injection.
            tb_max_depth (int, optional): Max frames in the traceback.
            tb_level (int, optional): Min `LogLevel` for traceback injection.
            scope (bool, optional): Enable automatic scope injection.
        """
        with self._lock:
            # The core "close and replace" reconfiguration pattern.
            self._logger.close()
            self._logger = _Logger(
                routes=list(routes),
                tb=tb,
                tb_max_depth=tb_max_depth,
                tb_level=tb_level,
                scope=scope,
            )

    def close(self) -> None:
        """Safely closes the currently active logger instance."""
        with self._lock:
            self._logger.close()

    def __del__(self) -> None:
        """A fail-safe finalizer to ensure resources are released."""
        with contextlib.suppress(Exception):
            self._logger.close()
