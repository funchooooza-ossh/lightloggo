"""
Provides the StdoutWriter for directing log output to the console.
"""

from .._c import CStdoutWriter
from ._base import writer


class StdoutWriter(writer):
    """
    Writes log records to the standard output stream (`sys.stdout`).

    This is the simplest writer, useful for development, debugging, or for
    applications that run as simple command-line scripts.

    Example:
        >>> from lightloggo.writers import StdoutWriter
        >>>
        >>> console_writer = StdoutWriter()
        >>> logger.add_route(writer=console_writer, formatter=TextFormatter())
    """

    def __init__(self) -> None:
        """Initializes the StdoutWriter."""
        # Instantiate the internal C-level writer, which gets a handle
        # to the Go core's standard output writer.
        self._c_writer = CStdoutWriter()
