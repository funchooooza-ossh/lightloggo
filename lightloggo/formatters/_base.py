"""
Defines the base class for all public formatter objects.

This module provides the abstract `formatter` class, establishing a common
interface and implementation for all formatters within the library.
"""

from .._c import CFormatter


class formatter:
    """
    The abstract base class for all log record formatters.

    This class should not be instantiated directly. Instead, use one of its
    concrete subclasses like `TextFormatter` or `JsonFormatter`.

    Its primary role is to provide a common interface, including the `.id`
    property, which is used internally to link the public formatter object
    to its corresponding resource in the native Go core.
    """

    # This class attribute annotation declares that all subclasses are
    # expected to have a `_c_formatter` instance attribute, which holds
    # the handle to the low-level C-level resource.
    _c_formatter: CFormatter

    @property
    def id(self) -> int:
        """
        The unique, low-level identifier for the native formatter resource.

        This read-only property provides access to the integer handle that
        uniquely identifies this formatter within the Go core. This ID is used
        internally by other components, such as a Router, to construct the
        complete logging pipeline.

        Returns:
            int: The internal resource ID.
        """
        return self._c_formatter._id
