from __future__ import annotations

import ctypes as C
import os
import sys
from typing import Callable, Union

try:
    from typing import Literal
except ImportError:
    from typing_extensions import Literal  # type: ignore

BytesLike = Union[bytes, bytearray, memoryview]
StrOrBytesLike = Union[str, BytesLike]


def _lib_filename() -> str:
    # Windows: .dll, macOS: .dylib, Linux/Unix: .so
    if sys.platform.startswith("win"):
        return "loggo.dll"
    elif sys.platform == "darwin":
        return "loggo.dylib"
    else:
        return "loggo.so"


def _candidate_paths() -> list[str]:
    here = os.path.dirname(__file__)
    return [os.path.join(here, _lib_filename())]


def _load_lib() -> C.CDLL:
    last_exc: Exception | None = None
    for path in _candidate_paths():
        try:
            return C.CDLL(os.path.abspath(path))
        except OSError as e:
            last_exc = e
    tried = "\n  - ".join(_candidate_paths())
    raise OSError(
        "Failed to load the native loggo library for platform "
        f"{sys.platform!r}. Tried paths:\n  - {tried}\n"
        f"Current working directory: {os.getcwd()}\n"
        f"Last error: {last_exc}"
    )


lib: C.CDLL = _load_lib()

# uintptr_t на стороне C/Go → используем c_size_t
ID_T = C.c_size_t

# ---- bind конструкторов/утилит ----
lib.NewLoggerWithSingleRoute.argtypes = [ID_T]
lib.NewLoggerWithSingleRoute.restype = ID_T

lib.NewLoggerWithRoutes.argtypes = [C.POINTER(ID_T), C.c_int]
lib.NewLoggerWithRoutes.restype = ID_T

lib.FreeLogger.argtypes = [ID_T]
lib.FreeLogger.restype = None

lib.NewFormatStyle.argtypes = [
    C.c_int,
    C.c_int,
    C.c_int,
    C.c_char_p,
    C.c_char_p,
    C.c_char_p,
]
lib.NewFormatStyle.restype = ID_T

lib.NewTextFormatter.argtypes = [ID_T, C.c_int]
lib.NewTextFormatter.restype = ID_T

lib.NewJsonFormatter.argtypes = [ID_T, C.c_int]
lib.NewJsonFormatter.restype = ID_T

lib.NewStdoutWriter.argtypes = []
lib.NewStdoutWriter.restype = ID_T

lib.NewFileWriter.argtypes = [
    C.c_char_p,  # path
    C.c_long,  # maxSizeMB
    C.c_int,  # maxBackups
    C.c_char_p,  # interval
    C.c_char_p,
]
lib.NewFileWriter.restype = ID_T


def _bind5(name: str) -> Callable[[int, bytes, int, bytes, int], None]:
    fn = getattr(lib, name)
    fn.argtypes = [ID_T, C.c_char_p, C.c_size_t, C.c_char_p, C.c_size_t]
    fn.restype = None
    return fn


LOG_FUNS = {
    "trace": _bind5("Logger_Trace"),
    "debug": _bind5("Logger_Debug"),
    "info": _bind5("Logger_Info"),
    "warning": _bind5("Logger_Warning"),
    "error": _bind5("Logger_Error"),
    "exception": _bind5("Logger_Exception"),
}


# ---- утилиты ----
def _as_bytes(x: StrOrBytesLike) -> bytes:
    return (
        x
        if isinstance(x, (bytes, bytearray))
        else memoryview(x.encode("utf-8")).tobytes()
    )


def log_call(
    method: Literal["trace", "debug", "info", "warning", "error", "exception"],
    logger_id: int,
    msg_b: bytes,
    fields_b: bytes,
) -> None:
    fn = LOG_FUNS[method]
    fn(
        ID_T(logger_id),
        C.c_char_p(msg_b),
        len(msg_b),
        C.c_char_p(fields_b),
        len(fields_b),
    )
