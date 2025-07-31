from .ffi import lib
from .formatter import Formatter
from .writer import Writer
from .enums import LogLevel


class RouteProcessor:
    def __init__(
        self, formatter: Formatter, writer: Writer, level: LogLevel = 20
    ):  # default INFO
        self._id = lib.NewRouteProcessor(formatter._id, writer._id, level)
