from .json_formatter import CJsonFormatter
from .text_formatter import CTextFormatter
from typing import Union

CFormatter = Union[CJsonFormatter, CTextFormatter]
