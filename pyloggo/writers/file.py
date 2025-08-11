from ..c import CFileWriter
from ._base import writer


class FileWriter(writer):
    def __init__(
        self,
        path: str,
        max_size_mb: int = 10,
        max_backups: int = 5,
        interval: str = "day",  # "day", "week", "month"
        compress: str = "gz",  # "gz" or ""
    ) -> None:
        self._c_writer = CFileWriter(
            path=path,
            max_backups=max_backups,
            max_size_mb=max_size_mb,
            interval=interval,
            compress=compress,
        )
