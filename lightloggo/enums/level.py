"""
Defines the logging level enumerations for the lightloggo library.

This module provides two distinct enumerations for handling log levels:
1.  `Level (StrEnum)`: User-facing, string-based constants (e.g., "Info").
2.  `LogLevel (IntEnum)`: Integer-based constants used for efficient numeric
    comparison, compatible with Python's standard logging conventions.
"""

from enum import IntEnum

# This block provides a fallback implementation of `StrEnum` for Python
# versions older than 3.11, ensuring compatibility. StrEnum allows enum
# members to behave like their string values.
try:
    from enum import StrEnum
except ImportError:
    import enum

    class StrEnum(str, enum.Enum):
        """A string-based enumeration for compatibility with Python < 3.11."""

        def __str__(self) -> str:
            return str(self.value)

        def __repr__(self) -> str:
            return f"{self.__class__.__name__}.{self.name}"


class Level(StrEnum):
    """
    A user-facing, string-based enumeration of available log levels.

    These members are intended for use in configuration, providing clear and
    readable level names. They can be used interchangeably with their string
    literals.

    Example:
        >>> logger.set_level(Level.Info)
        >>> logger.set_level("Info")  # Equivalent
    """

    Trace = "Trace"
    Debug = "Debug"
    Info = "Info"
    Warning = "Warning"
    Error = "Error"
    Exception = "Exception"


class LogLevel(IntEnum):
    """
    An integer-based enumeration of log levels for efficient filtering.

    These values align with the conventions of Python's standard `logging`
    module. They are used internally for all numeric comparisons and are
    passed to the native Go core to determine if a log message should be
    processed.

    - TRACE: 0
    - DEBUG: 10
    - INFO: 20
    - WARNING: 30
    - ERROR: 40
    - EXCEPTION: 50
    """

    Trace = 0
    Debug = 10
    Info = 20
    Warning = 30
    Error = 40
    Exception = 50

    @classmethod
    def _from_string(cls, level: Level) -> "LogLevel":
        """
        An internal utility to convert a string-based Level to a numeric LogLevel.

        Args:
            level (Level): The string-based Level enum member.

        Returns:
            LogLevel: The corresponding integer-based LogLevel enum member.
        """
        # This direct mapping converts from the user-facing string enum
        # to the internal integer enum used by the core.
        if level == Level.Trace:
            return cls.Trace
        elif level == Level.Debug:
            return cls.Debug
        elif level == Level.Info:
            return cls.Info
        elif level == Level.Warning:
            return cls.Warning
        elif level == Level.Error:
            return cls.Error
        elif level == Level.Exception:
            return cls.Exception
        # A direct mapping is used, so an else/default case is not strictly
        # necessary if the input is always a valid Level member.
