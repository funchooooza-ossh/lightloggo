from ...ffi.ffi import lib
from ...enums import LogLevel


class CRouteProcessor:
    def __init__(
        self, formatter_id: int, writer_id: int, level: LogLevel = 20
    ):  # default INFO
        self._id = lib.NewRouteProcessor(formatter_id, writer_id, level)
