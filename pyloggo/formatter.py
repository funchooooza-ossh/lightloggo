from .ffi import lib
from .style import FormatStyle
from typing import Union


class TextFormatter:
    def __init__(self, style: FormatStyle = None):
        style_id = style._id if style else 0
        self._id = lib.NewTextFormatter(style_id)


class JsonFormatter:
    def __init__(self, style: FormatStyle = None):
        style_id = style._id if style else 0
        self._id = lib.NewJsonFormatter(style_id)


Formatter = Union[TextFormatter, JsonFormatter]
