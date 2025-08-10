from .formatter import Formatter
from .writer import Writer
from .enums import LogLevel
from .c import CRouter, CRouteProcessor
from .formatter import TextFormatter
from .writer import StdoutWriter


class router:
    _c_router: CRouter

    @property
    def id(self) -> int:
        return self._c_router._id


class RouteProcessor(router):
    def __init__(
        self,
        formatter: Formatter = TextFormatter(),
        writer: Writer = StdoutWriter(),
        level: LogLevel = 20,
    ):  # default INFO
        self._c_router = CRouteProcessor(formatter.id, writer.id, level)
