from typing import Union

from .json import JsonFormatter
from .text import TextFormatter

Formatter = Union[TextFormatter, JsonFormatter]
