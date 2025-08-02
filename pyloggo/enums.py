from enum import IntEnum, StrEnum


class Level(StrEnum):
    Trace = "Trace"
    Debug = "Debug"
    Info = "Info"
    Warning = "Warning"
    Error = "Error"
    Exception = "Exception"


class LogLevel(IntEnum):
    Trace = 0
    Debug = 10
    Info = 20
    Warning = 30
    Error = 40
    Exception = 50

    @classmethod
    def _from_string(cls, level: Level) -> "LogLevel":
        match level:
            case Level.Trace:
                return cls.Trace
            case Level.Debug:
                return cls.Debug
            case Level.Info:
                return cls.Info
            case Level.Warning:
                return cls.Warning
            case Level.Error:
                return cls.Error
            case Level.Exception:
                return cls.Exception
