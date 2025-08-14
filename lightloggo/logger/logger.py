"""
Provides the internal, core implementation of the logger object.

This module contains the `_Logger` class, which serves as the primary
workhorse of the library. It handles the low-level logic of preparing log
data, adding context like scope or tracebacks, and dispatching the final
payload to the FFI layer.
"""

from __future__ import annotations

import contextlib
import linecache
import os
import sys
from typing import Any, Dict, List

try:
    from typing import Literal
except ImportError:
    # Maintain compatibility with Python versions older than 3.8.
    from typing_extensions import Literal  # type: ignore

from .._c import CLogger
from .._ffi._ffi import _as_bytes, log_call
from .._help import _serialize_fields
from ..enums import LogLevel
from ..routers import RouteProcessor


class _Logger:
    """
    The internal workhorse class for all logging operations.

    This class is not intended for direct user instantiation. It is managed
    by the public-facing, thread-safe `Logger` class.

    Its responsibilities include:
    - Managing the lifecycle of a `CLogger` resource handle.
    - Preparing log messages and fields for serialization.
    - Conditionally injecting scope or traceback information.
    - Dispatching the final serialized data to the Go core via the FFI.
    """

    def __init__(
        self,
        routes: List[RouteProcessor],
        tb: bool = False,
        tb_max_depth: int = 10,
        tb_level: int = LogLevel.Error,
        scope: bool = True,
    ) -> None:
        """
        Initializes the internal logger instance.

        Args:
            routes (List[RouteProcessor]): A list of configured route objects
                that define the logging pipelines.
            tb (bool, optional): If True, enables automatic traceback
                injection for logs at or above `tb_level`. Defaults to False.
            tb_max_depth (int, optional): The maximum number of frames to
                include in an automatic traceback. Defaults to 10.
            tb_level (int, optional): The minimum `LogLevel` required to
                trigger automatic traceback injection. Defaults to `LogLevel.Error`.
            scope (bool, optional): If True, adds caller scope information
                (file, line, function) to logs that do not have a traceback.
                Defaults to True.
        """
        # Extract the low-level resource IDs from the public route objects.
        route_ids = [r.id for r in routes]
        # Create the underlying C-level logger resource with these routes.
        self._c_logger = CLogger(route_ids)

        # Store the configuration for this logger instance.
        self._routes = routes
        self._tb = tb
        self._tb_max_depth = tb_max_depth
        self._scope = scope
        self._tb_level = tb_level

    @property
    def id(self) -> int:
        """The unique, low-level identifier for the native logger resource."""
        return self._c_logger._id

    def _log(
        self,
        method: Literal["trace", "debug", "info", "warning", "error", "exception"],
        msg: str,
        **kwargs: Any,
    ) -> None:
        """
        The central internal method for processing and dispatching a log call.

        Args:
            method: The string name of the log level.
            msg: The primary log message.
            **kwargs: Additional key-value fields to include in the log.
        """
        # Avoid processing if there is no content.
        if not msg and not kwargs:
            return

        # Resolve the numeric log level from the method name.
        level = getattr(LogLevel, method.capitalize())
        msg_b = _as_bytes(msg)

        # Prepare and serialize the fields.
        if not kwargs:
            # Use a placeholder if no fields are provided. The Go core expects
            # a non-empty fields payload. "0" is a minimal indicator.
            fields_b = b"0"
        else:
            # Add automatic context (scope/traceback) and serialize.
            final_fields = self._resolve_fields(level, kwargs)
            fields_b = _serialize_fields(final_fields)

        # Dispatch the prepared, serialized data to the low-level FFI function.
        log_call(method, self.id, msg_b, fields_b)

    def _resolve_fields(
        self,
        level: LogLevel,
        fields: Dict[str, Any],
    ) -> Dict[str, Any]:
        """
        Conditionally adds automatic scope or traceback info to the fields.

        Traceback injection takes precedence over scope injection.

        Args:
            level: The numeric level of the current log message.
            fields: The original dictionary of user-provided fields.

        Returns:
            A new dictionary of fields, potentially with 'tb' or 'scope' added.
        """
        fields_cp = dict(fields)
        has_added_context = False

        # Add traceback if enabled and the log level is high enough.
        if self._tb and self._tb_level <= level:
            fields_cp["tb"] = self._add_traceback(max_depth=self._tb_max_depth)
            has_added_context = True

        # Add scope only if traceback was not added and scope is enabled.
        if not has_added_context and self._scope:
            fields_cp["scope"] = self._add_scope()

        return fields_cp

    @staticmethod
    def _add_scope(frame_depth: int = 5) -> str:
        """
        Captures the caller's scope (file, line, function).

        Note: This relies on `sys._getframe`, a CPython implementation detail.

        Args:
            frame_depth: How many frames to go up the stack to find the
                user's calling frame, bypassing internal logger calls.

        Returns:
            A formatted string representing the caller's scope.
        """
        try:
            # Go up the stack to find the frame outside the logger's own calls.
            frame = sys._getframe(frame_depth)
            filename = os.path.basename(frame.f_code.co_filename)
            lineno = frame.f_lineno
            func = frame.f_code.co_name
            return f"{filename}:{lineno} in {func}()"
        except (ValueError, AttributeError):
            # Fallback if the frame is unavailable (e.g., in some execution contexts).
            return "<scope unavailable>"

    @staticmethod
    def _add_traceback(max_depth: int = 10, skip: int = 5) -> str:
        """
        Captures and formats a multi-line traceback from the current call stack.

        Args:
            max_depth: The maximum number of frames to capture.
            skip: The number of initial frames to skip to exclude logger internals.

        Returns:
            A formatted, multi-line string representing the traceback.
        """
        lines = []
        try:
            # Start from a frame high enough up the stack to be in user code.
            frame = sys._getframe(skip)

            for _ in range(max_depth):
                if frame is None:
                    break

                filename_full = frame.f_code.co_filename
                filename = os.path.basename(filename_full)
                lineno = frame.f_lineno
                func = frame.f_code.co_name

                # Use linecache to efficiently retrieve the source code line.
                code_line = linecache.getline(filename_full, lineno).strip()

                lines.append(
                    f'  File "{filename}", line {lineno}, in {func}()\n    {code_line}\n'
                )

                frame = frame.f_back
            # Reverse the lines to show the call order from top to bottom.
            return "".join(reversed(lines))
        except (ValueError, AttributeError):
            return "<traceback unavailable>"

    def trace(self, msg: str, **kwargs: Any) -> None:
        """Logs a message with the TRACE level."""
        self._log("trace", msg, **kwargs)

    def debug(self, msg: str, **kwargs: Any) -> None:
        """Logs a message with the DEBUG level."""
        self._log("debug", msg, **kwargs)

    def info(self, msg: str, **kwargs: Any) -> None:
        """
        Logs a message with the INFO level.

        Example:
            >>> logger.info("User authenticated successfully", user_id=123)
        """
        self._log("info", msg, **kwargs)

    def warning(self, msg: str, **kwargs: Any) -> None:
        """Logs a message with the WARNING level."""
        self._log("warning", msg, **kwargs)

    def error(self, msg: str, **kwargs: Any) -> None:
        """Logs a message with the ERROR level."""
        self._log("error", msg, **kwargs)

    def exception(self, msg: str, **kwargs: Any) -> None:
        """Logs a message with the EXCEPTION level."""
        self._log("exception", msg, **kwargs)

    def close(self) -> None:
        """Closes and frees the underlying native logger resource."""
        self._c_logger.close()

    def __del__(self) -> None:
        """A fail-safe finalizer to ensure resources are released."""
        with contextlib.suppress(Exception):
            self.close()


def create_default_logger() -> _Logger:
    """An internal factory to create a default logger instance."""
    # A default logger has one route: INFO level to stdout with a text formatter.
    router = RouteProcessor()
    return _Logger([router])


# Defines the internal API of this module.
__all__ = [
    "_Logger",
    "create_default_logger",
]
