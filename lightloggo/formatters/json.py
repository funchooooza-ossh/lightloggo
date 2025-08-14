"""
Provides the JsonFormatter for machine-readable, structured log output.
"""

from .._c import CJsonFormatter
from ..styles import FormatStyle
from ._base import formatter


class JsonFormatter(formatter):
    """
    Formats log records into a structured JSON string.

    This formatter is ideal for production environments where logs are sent
    to aggregation systems like Elasticsearch, Splunk, or Datadog, as it
    preserves the structure of the logged data.

    Example:
        >>> json_fmt = JsonFormatter()
        >>> file_writer = FileWriter(path="/var/log/app.log")
        >>> logger.add_route(writer=file_writer, formatter=json_fmt)
    """

    def __init__(self, style: FormatStyle = None, max_depth: int = 3) -> None:
        """
        Initializes the JsonFormatter.

        Args:
            style (FormatStyle, optional): A `FormatStyle` object to customize
                output. For JSON, this typically controls whether ANSI color
                codes are embedded in the output. If None, a default style
                is used. Defaults to None.
            max_depth (int, optional): The maximum depth for serializing
                complex data structures into the final JSON object.
                Defaults to 3.
        """
        # Retrieve the internal resource ID from the public style object.
        # If no style is provided, default to 0.
        style_id = style.id if style else 0

        # Instantiate the internal C-level formatter, creating the
        # corresponding resource in the Go core via the FFI.
        self._c_formatter = CJsonFormatter(style_id=style_id, max_depth=max_depth)
