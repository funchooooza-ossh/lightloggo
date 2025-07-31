from .ffi import lib
import ctypes
from typing import Union


class FileWriter:
    def __init__(
        self,
        path: str,
        max_size_mb: int = 10,
        max_backups: int = 5,
        interval: str = "day",  # "day", "week", "month"
        compress: str = "gz",  # "gz" or ""
    ):
        self._path = path.encode()
        self._interval = interval.encode()
        self._compress = compress.encode()

        self._id = lib.NewFileWriter(
            ctypes.c_char_p(self._path),
            ctypes.c_long(max_size_mb),
            ctypes.c_int(max_backups),
            ctypes.c_char_p(self._interval),
            ctypes.c_char_p(self._compress),
        )

        if not self._id:
            raise RuntimeError("Failed to create FileWriter")


class StdoutWriter:
    def __init__(self):
        self._id = lib.NewStdoutWriter()


Writer = Union[FileWriter, StdoutWriter]
