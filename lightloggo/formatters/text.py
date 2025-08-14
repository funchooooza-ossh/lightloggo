"""
Provides the TextFormatter for human-readable log output.
"""

from .._c import CTextFormatter
from ..styles import FormatStyle
from ._base import formatter


class TextFormatter(formatter):
    """
    Formats log records into a human-readable, plain-text line.

    This is a standard formatter suitable for logging directly to a console
    or a simple text file during development or for simple applications.

    Example:
        >>> style = FormatStyle(color_keys=True)
        >>> text_fmt = TextFormatter(style=style)
        >>> logger.add_route(writer=StdoutWriter(), formatter=text_fmt)
    """

    def __init__(self, style: FormatStyle = None, max_depth: int = 3) -> None:
        """
        Initializes the TextFormatter.

        Args:
            style (FormatStyle, optional): A `FormatStyle` object to customize
                output colors and appearance. If None, a default style is used.
                Defaults to None.
            max_depth (int, optional): The maximum depth for serializing
                complex data structures (e.g., nested dicts) in the log fields.
                Defaults to 3.
        """
        # Retrieve the internal resource ID from the public style object.
        # If no style is provided, default to 0, which the Go core
        # interprets as the default style configuration.
        style_id = style.id if style else 0

        # Instantiate the internal C-level formatter, which calls the FFI
        # to create the corresponding resource in the Go core.
        self._c_formatter = CTextFormatter(style_id=style_id, max_depth=max_depth)
