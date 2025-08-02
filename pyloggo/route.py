from .formatter import Formatter
from .writer import Writer
from .enums import LogLevel
from .c import CRouter, CRouteProcessor


class router:
    _c_router: CRouter

    @property
    def id(self) -> int:
        return self._c_router._id


class RouteProcessor(router):
    def __init__(
        self, formatter: Formatter, writer: Writer, level: LogLevel = 20
    ):  # default INFO
        self._c_router = CRouteProcessor(formatter.id, writer.id, level)
