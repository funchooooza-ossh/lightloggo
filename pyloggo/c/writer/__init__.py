from .file_writer import CFileWriter
from .stdout_writer import CStdoutWriter
from typing import Union

CWriter = Union[CFileWriter, CStdoutWriter]
