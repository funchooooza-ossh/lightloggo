from ..c import CJsonFormatter
from ..styles import FormatStyle
from ._base import formatter


class JsonFormatter(formatter):
    def __init__(self, style: FormatStyle = None, max_depth: int = 3) -> None:
        style_id = style.id if style else 0
        self._c_formatter = CJsonFormatter(style_id=style_id, max_depth=max_depth)
