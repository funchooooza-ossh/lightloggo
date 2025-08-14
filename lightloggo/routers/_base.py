"""
Defines the base class for all public router objects.

This module provides the abstract `router` class, establishing a common
interface for all routing implementations.
"""

from .._c import CRouter


class router:
    """
    The abstract base class for all log routing objects.

    This class should not be instantiated directly. A router represents a
    single, complete processing pipeline for a log message.

    Its primary role is to provide a common interface, including the `.id`
    property, which is used internally to link the public router object
    to its corresponding resource in the native Go core.
    """

    # This class attribute annotation declares the contract that all subclasses
    # must have a `_c_router` instance attribute, which holds the handle
    # to the low-level C-level resource.
    _c_router: CRouter

    @property
    def id(self) -> int:
        """
        The unique, low-level identifier for the native route resource.

        This read-only property provides access to the integer handle that
        uniquely identifies this route within the Go core. This ID is used
        by a logger instance to register this complete processing pipeline.

        Returns:
            int: The internal resource ID.
        """
        return self._c_router._id
