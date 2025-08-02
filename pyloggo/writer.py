from typing import Union
from .c import CFileWriter, CStdoutWriter, CWriter


class writer:
    _c_writer: CWriter

    @property
    def id(self) -> int:
        return self._c_writer._id


class FileWriter(writer):
    def __init__(
        self,
        path: str,
        max_size_mb: int = 10,
        max_backups: int = 5,
        interval: str = "day",  # "day", "week", "month"
        compress: str = "gz",  # "gz" or ""
    ):
        self._c_writer = CFileWriter(
            path=path,
            max_backups=max_backups,
            max_size_mb=max_size_mb,
            interval=interval,
            compress=compress,
        )


class StdoutWriter(writer):
    def __init__(self):
        self._c_writer = CStdoutWriter()


Writer = Union[FileWriter, StdoutWriter]
