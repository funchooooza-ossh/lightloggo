from ..c import CTextFormatter
from ..styles import FormatStyle
from ._base import formatter


class TextFormatter(formatter):
    def __init__(self, style: FormatStyle = None, max_depth: int = 3) -> None:
        style_id = style.id if style else 0
        self._c_formatter = CTextFormatter(style_id=style_id, max_depth=max_depth)
