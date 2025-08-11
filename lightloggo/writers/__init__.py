from typing import Union

from .file import FileWriter
from .stdout import StdoutWriter

Writer = Union[FileWriter, StdoutWriter]
