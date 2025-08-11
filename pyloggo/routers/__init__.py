from ..c import CRouteProcessor
from ..enums import LogLevel
from ..formatters import Formatter, TextFormatter
from ..writers import StdoutWriter, Writer
from ._base import router


class RouteProcessor(router):
    def __init__(
        self,
        formatter: Formatter = TextFormatter(),
        writer: Writer = StdoutWriter(),
        level: LogLevel = 20,
    ) -> None:  # default INFO
        self._c_router = CRouteProcessor(formatter.id, writer.id, level)
