"""
Internal C-level wrapper for the Format Style resource.

This module provides the CFormatStyle class, which interfaces with the FFI
layer to create and manage a styling configuration instance within the Go
core library. This configuration is then used by formatters.
"""

from ..._ffi._ffi import lib


class CFormatStyle:
    """
    A low-level handle to a formatting style resource in the Go core.

    This class is an internal component and is not intended for direct use
    by end-users. It encapsulates all styling options—primarily colors—into a
    single configuration resource within the native library.

    The resulting resource ID can be passed to C-level formatters (e.g.,
    CJsonFormatter) to apply a consistent style.

    Attributes:
        _id (int): The unique identifier/handle for the native style resource.
    """

    def __init__(
        self,
        color_keys: bool = True,
        color_values: bool = True,
        color_level: bool = False,
        # Default to standard ANSI escape codes for blue keys and yellow values.
        key_color: str = "\033[34m",
        value_color: str = "\033[33m",
        reset: str = "\033[0m",
    ) -> None:
        """
        Initializes CFormatStyle by creating a new style resource via FFI.

        Args:
            color_keys (bool, optional): If True, apply color to keys in
                log fields. Defaults to True.
            color_values (bool, optional): If True, apply color to values
                in log fields. Defaults to True.
            color_level (bool, optional): If True, apply color to the log
                level name (e.g., 'INFO', 'WARN'). Defaults to False.
            key_color (str, optional): The ANSI escape code for key color.
                Defaults to blue.
            value_color (str, optional): The ANSI escape code for value color.
                Defaults to yellow.
            reset (str, optional): The ANSI escape code to reset all
                color attributes. Defaults to standard reset.
        """
        # Call the FFI function to create a new style configuration object
        # in the Go core. The boolean flags are converted to integers (0 or 1)
        # and strings are encoded to bytes, as required by the C interface.
        self._id = lib.NewFormatStyle(
            int(color_keys),
            int(color_values),
            int(color_level),
            key_color.encode(),
            value_color.encode(),
            reset.encode(),
        )
