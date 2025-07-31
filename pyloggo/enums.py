from enum import IntEnum


class LogLevel(IntEnum):
    Trace = 0
    Debug = 10
    Info = 20
    Warning = 30
    Error = 40
    Exception = 50
