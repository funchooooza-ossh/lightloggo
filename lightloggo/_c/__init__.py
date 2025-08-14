"""
Internal C-Level Abstraction Package.

This package, conventionally named `_c` (for "core" or "C-level"),
encapsulates all direct interactions with the native Go shared library,
building upon the low-level FFI bindings.

The primary responsibility of this layer is to provide a one-to-one mapping
of Python objects to resources (loggers, writers, formatters, etc.)
managed by the Go core. Each class within this package typically acts as a
"handle manager"—its main purpose is to acquire a unique resource ID from
the native library upon instantiation and hold it for its lifetime.

This entire package is a private, internal implementation detail of the
lightloggo library. Its structure and components are subject to change and
are NOT part of the public API.
"""

# Import the C-level wrapper classes and their corresponding type aliases
# from their respective submodules.
from .format import CFormatter, CJsonFormatter, CTextFormatter
from .logger import CLogger
from .route import CRouteProcessor, CRouter
from .style import CFormatStyle, CStyle
from .writer import CFileWriter, CStdoutWriter, CWriter

# The __all__ list explicitly defines the internal API of this package.
# It aggregates all C-level wrapper classes, making them accessible to other
# internal parts of the lightloggo library (e.g., the public API layer)
# through a single, consistent import point: `from lightloggo._c import ...`.
__all__ = [  # noqa: RUF022
    # Writer classes
    "CFileWriter",
    "CStdoutWriter",
    # Style class
    "CFormatStyle",
    # Formatter classes
    "CJsonFormatter",
    "CTextFormatter",
    # Logger and Routing classes
    "CLogger",
    "CRouteProcessor",
    "CRouter",
    # Union type aliases for internal type hinting
    "CFormatter",
    "CStyle",
    "CWriter",
]
