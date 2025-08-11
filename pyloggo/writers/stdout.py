from ..c import CStdoutWriter
from ._base import writer


class StdoutWriter(writer):
    def __init__(self) -> None:
        self._c_writer = CStdoutWriter()
