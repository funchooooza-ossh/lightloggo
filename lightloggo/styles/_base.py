"""
Defines public classes for controlling the visual style of log output.

This module provides the `FormatStyle` class, which allows for detailed
configuration of colors and other styling options applied by a formatter.
"""

from .._c import CFormatStyle, CStyle


class BaseStyle:
    """
    The abstract base class for all style configuration objects.

    This class should not be instantiated directly. It provides a common
    interface for all style implementations, most importantly the `.id`
    property for communicating with the library's core.
    """

    # This class attribute annotation establishes the contract that all
    # subclasses must define a `_c_style` instance attribute, which holds
    # the handle to the corresponding low-level C-wrapper.
    _c_style: CStyle

    @property
    def id(self) -> int:
        """
        The unique, low-level identifier for the native style resource.

        This read-only property provides access to the integer handle that
        uniquely identifies this style configuration within the Go core. This
        ID is used internally to link this style to a formatter.

        Returns:
            int: The internal resource ID.

        Raises:
            RuntimeError: If the internal C-style object has not been
                initialized by a subclass.
        """
        if not hasattr(self, "_c_style"):
            raise RuntimeError(
                "C-level style object has not been initialized. "
                "Ensure the subclass constructor assigns to self._c_style."
            )
        return self._c_style._id


class FormatStyle(BaseStyle):
    """
    Provides a detailed configuration for text-based styling.

    This class allows for fine-grained control over the use of colors in
    the log output, typically via ANSI escape codes. An instance of this
    class can be passed to a formatter to apply the specified styles.

    Example:
        >>> from lightloggo.styles import FormatStyle
        >>> from lightloggo.formatters import TextFormatter
        >>>
        >>> # Define a style with yellow keys and no color for values
        >>> my_style = FormatStyle(key_color="\\033[33m", color_values=False)
        >>>
        >>> # Apply the style to a formatter
        >>> my_formatter = TextFormatter(style=my_style)
    """

    def __init__(
        self,
        color_keys: bool = True,
        color_values: bool = True,
        color_level: bool = False,
        key_color: str = "\033[34m",
        value_color: str = "\033[33m",
        reset: str = "\033[0m",
    ) -> None:
        """
        Initializes the FormatStyle configuration.

        Args:
            color_keys (bool, optional): If True, applies color to keys in
                log fields. Defaults to True.
            color_values (bool, optional): If True, applies color to values
                in log fields. Defaults to True.
            color_level (bool, optional): If True, applies color to the log
                level name (e.g., 'INFO'). Defaults to False.
            key_color (str, optional): The ANSI escape code for key color.
                Defaults to blue.
            value_color (str, optional): The ANSI escape code for value color.
                Defaults to yellow.
            reset (str, optional): The ANSI escape code to reset all
                color attributes. Defaults to the standard reset code.
        """
        # Instantiate the internal C-level style object, passing all
        # configuration options. This call crosses the FFI boundary to
        # create the corresponding resource in the Go core.
        self._c_style = CFormatStyle(
            color_keys=color_keys,
            color_values=color_values,
            color_level=color_level,
            key_color=key_color,
            value_color=value_color,
            reset=reset,
        )
