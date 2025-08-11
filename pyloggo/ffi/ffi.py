import ctypes as C
import os
from typing import Callable, Union

try:
    from typing import Literal
except ImportError:
    from typing_extensions import Literal  # type: ignore

BytesLike = Union[bytes, bytearray, memoryview]
StrOrBytesLike = Union[str, BytesLike]

lib_path = os.path.join(os.path.dirname(__file__), "loggo.so")
lib = C.CDLL(lib_path)

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
