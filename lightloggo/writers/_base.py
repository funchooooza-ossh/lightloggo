"""
Defines the base class for all public writer objects.

This module provides the abstract `writer` class, establishing a common
interface for all writers, which are responsible for the final output
destination of log records.
"""

from .._c import CWriter


class writer:
    """
    The abstract base class for all log record writers.

    This class should not be instantiated directly. Instead, use one of its
    concrete subclasses like `StdoutWriter` or `FileWriter`.

    Its primary role is to provide a common interface, including the `.id`
    property, which is used internally to link the public writer object
    to its corresponding resource in the native Go core.
    """

    # This class attribute annotation declares the contract that all subclasses
    # must have a `_c_writer` instance attribute, which holds the handle
    # to the low-level C-level resource.
    _c_writer: CWriter

    @property
    def id(self) -> int:
        """
        The unique, low-level identifier for the native writer resource.

        This read-only property provides access to the integer handle that
        uniquely identifies this writer within the Go core. This ID is used
        internally by a Router to link this writer to a specific log
        processing pipeline.

        Returns:
            int: The internal resource ID.
        """
        return self._c_writer._id
