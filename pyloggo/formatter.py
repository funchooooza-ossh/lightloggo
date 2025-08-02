from .style import FormatStyle
from typing import Union
from .c import CFormatter, CJsonFormatter, CTextFormatter


class formatter:
    _c_formatter: CFormatter

    @property
    def id(self) -> int:
        return self._c_formatter._id


class TextFormatter(formatter):
    def __init__(self, style: FormatStyle = None):
        style_id = style.id if style else 0
        self._c_formatter = CTextFormatter(style_id)


class JsonFormatter(formatter):
    def __init__(self, style: FormatStyle = None):
        style_id = style.id if style else 0
        self._c_formatter = CJsonFormatter(style_id)


Formatter = Union[TextFormatter, JsonFormatter]
