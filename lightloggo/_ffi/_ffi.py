"""
Internal FFI (Foreign Function Interface) Bridge.

This module is the sole entry point for interacting with the compiled Go
core library (`lightloggo.so`, `.dylib`, or `.dll`). It is responsible for:
1.  Locating and loading the shared library based on the operating system.
2.  Defining C-compatible data types that map to Go/C types.
3.  Binding to the exported C functions from the Go library, setting their
    argument types (`argtypes`) and return types (`restype`) for correctness
    and marshalling.
4.  Providing low-level Python wrappers (`log_call`) that handle the
    conversion of Python types to C types before calling the native functions.

This module is considered a private, internal implementation detail of the
lightloggo package and should not be used directly by end-users. Its stability
is not guaranteed across versions.
"""

from __future__ import annotations

import ctypes as C
import os
import sys
from typing import Callable, Union

try:
    from typing import Literal
except ImportError:
    # Maintain compatibility with Python versions older than 3.8.
    from typing_extensions import Literal  # type: ignore

# --- Type Aliases for internal consistency ---
BytesLike = Union[bytes, bytearray, memoryview]
StrOrBytesLike = Union[str, BytesLike]


def _lib_filename() -> str:
    """Determines the appropriate shared library filename for the current OS.

    Returns:
        str: The name of the library file (e.g., 'lightloggo.dll').
    """

    # Windows uses .dll for shared libraries.
    if sys.platform.startswith("win"):
        return "lightloggo.dll"
    # macOS uses .dylib.
    elif sys.platform == "darwin":
        return "lightloggo.dylib"
    # Linux and other Unix-like systems use .so (shared object).
    else:
        return "lightloggo.so"


def _candidate_paths() -> list[str]:
    """Generates a list of potential paths to find the shared library.

    The primary search location is the same directory as this FFI module.

    Returns:
        list[str]: A list of absolute paths to check for the library file.
    """
    # The library is expected to be co-located with this Python module.
    # This is the standard location when packaged in a wheel.
    here = os.path.dirname(__file__)
    return [os.path.join(here, _lib_filename())]


def _load_lib() -> C.CDLL:
    """Loads and returns the Go shared library object.

    Iterates through candidate paths and attempts to load the library.
    If loading fails for all candidates, it raises a comprehensive OSError.

    Returns:
        ctypes.CDLL: The loaded library object.

    Raises:
        OSError: If the shared library cannot be found or loaded.
    """
    last_exc: Exception | None = None
    for path in _candidate_paths():
        try:
            # Load the library using its absolute path to avoid ambiguity.
            return C.CDLL(os.path.abspath(path))
        except OSError as e:
            # Store the last exception to provide more context on failure.
            last_exc = e

    # If the loop completes without returning, the library was not found.
    tried = "\n  - ".join(_candidate_paths())
    raise OSError(
        "Failed to load the native lightloggo library for platform "
        f"{sys.platform!r}. Tried paths:\n  - {tried}\n"
        f"Current working directory: {os.getcwd()}\n"
        f"Last error: {last_exc}"
    )


# --- Global Library Handle ---
# This handle provides access to all exported functions from the Go core.
lib: C.CDLL = _load_lib()

# --- Type Definitions & Mappings ---
# In Go/C, handles and IDs are represented as `uintptr_t`.
# In Python's ctypes, `ctypes.c_size_t` is the correct, platform-agnostic
# equivalent for memory addresses and handles.
ID_T = C.c_size_t

# =========================================================================
# ==                 C FUNCTION BINDINGS & DEFINITIONS                   ==
# =========================================================================
# This section defines the function signatures for every C function exported
# from the Go library. Setting `argtypes` and `restype` is crucial for
# `ctypes` to perform type checking and proper data marshalling.

# ---- Constructor and Destructor Bindings ----

# func NewLoggerWithSingleRoute(routeID uintptr_t) uintptr_t
lib.NewLoggerWithSingleRoute.argtypes = [ID_T]
lib.NewLoggerWithSingleRoute.restype = ID_T

# func NewLoggerWithRoutes(routeIDs []uintptr_t, count int) uintptr_t
# Note: A C array `T*` is represented by `POINTER(T)` in ctypes.
lib.NewLoggerWithRoutes.argtypes = [C.POINTER(ID_T), C.c_int]
lib.NewLoggerWithRoutes.restype = ID_T

# func FreeLogger(loggerID uintptr_t)
lib.FreeLogger.argtypes = [ID_T]
lib.FreeLogger.restype = None  # Represents a 'void' return type.

# func NewFormatStyle(colorLevel, colorMsg, colorFields int, timeFmt, levelFmt, fieldsFmt *C.char) uintptr_t
lib.NewFormatStyle.argtypes = [
    C.c_int,
    C.c_int,
    C.c_int,
    C.c_char_p,
    C.c_char_p,
    C.c_char_p,
]
lib.NewFormatStyle.restype = ID_T

# func NewTextFormatter(styleID uintptr_t, withColor int) uintptr_t
lib.NewTextFormatter.argtypes = [ID_T, C.c_int]
lib.NewTextFormatter.restype = ID_T

# func NewJsonFormatter(styleID uintptr_t, withColor int) uintptr_t
lib.NewJsonFormatter.argtypes = [ID_T, C.c_int]
lib.NewJsonFormatter.restype = ID_T

# func NewStdoutWriter() uintptr_t
lib.NewStdoutWriter.argtypes = []
lib.NewStdoutWriter.restype = ID_T

# func NewFileWriter(path *C.char, maxSizeMB int64, maxBackups int, interval, compression *C.char) uintptr_t
lib.NewFileWriter.argtypes = [
    C.c_char_p,  # path
    C.c_long,  # maxSizeMB
    C.c_int,  # maxBackups
    C.c_char_p,  # interval
    C.c_char_p,
]
lib.NewFileWriter.restype = ID_T


# ---- Log Function Bindings ----
def _bind5(name: str) -> Callable[[int, bytes, int, bytes, int], None]:
    """A factory to bind a standard 5-argument log function from the library.

    This reduces boilerplate code for binding the various log level functions
    (Trace, Debug, Info, etc.), which all share the same C signature.

    The C signature is:
    `void Logger_METHOD(loggerID uintptr_t, msg *C.char, msgLen size_t, fields *C.char, fieldsLen size_t)`

    Args:
        name (str): The name of the function to bind (e.g., "Logger_Info").

    Returns:
        Callable: The bound and type-hinted Python callable.
    """
    fn = getattr(lib, name)
    fn.argtypes = [ID_T, C.c_char_p, C.c_size_t, C.c_char_p, C.c_size_t]
    fn.restype = None  # All log functions return void.
    return fn


# A dictionary mapping log level names to their bound C functions.
# This allows for dynamic dispatch in the `log_call` function.
LOG_FUNS = {
    "trace": _bind5("Logger_Trace"),
    "debug": _bind5("Logger_Debug"),
    "info": _bind5("Logger_Info"),
    "warning": _bind5("Logger_Warning"),
    "error": _bind5("Logger_Error"),
    "exception": _bind5("Logger_Exception"),
}


# =========================================================================
# ==                        FFI UTILITY FUNCTIONS                        ==
# =========================================================================
def _as_bytes(x: StrOrBytesLike) -> bytes:
    """Encodes a string to UTF-8 bytes if it is not already bytes.

    This is a helper to ensure all string data passed to the C layer
    is in a consistent byte representation.

    Args:
        x: The input string or bytes-like object.

    Returns:
        bytes: The UTF-8 encoded bytes.
    """
    if isinstance(x, (bytes, bytearray)):
        return x
    # For zero-copy slicing and conversion, we create a memoryview first.
    return memoryview(x.encode("utf-8")).tobytes()


def log_call(
    method: Literal["trace", "debug", "info", "warning", "error", "exception"],
    logger_id: int,
    msg_b: bytes,
    fields_b: bytes,
) -> None:
    """
    Performs the low-level FFI call to a specific Go logging function.

    This function acts as the final dispatch point before crossing the
    Python-to-C boundary. It selects the correct bound C function from
    LOG_FUNS and invokes it with the provided arguments, which are
    already expected to be in the correct format (ID and bytes).

    Args:
        method: The name of the logging method to call.
        logger_id: The ID handle of the target logger instance.
        msg_b: The primary log message, encoded as bytes.
        fields_b: The structured fields, serialized (e.g., as JSON) and
                  encoded as bytes.
    """
    # Retrieve the pre-bound C function for the requested method.
    fn = LOG_FUNS[method]

    # Invoke the C function, passing all arguments.
    # `ctypes` handles the conversion from Python types (int, bytes) to the
    # C types defined in `argtypes` (c_size_t, c_char_p, c_size_t).
    fn(
        ID_T(logger_id),
        C.c_char_p(msg_b),
        len(msg_b),
        C.c_char_p(fields_b),
        len(fields_b),
    )
