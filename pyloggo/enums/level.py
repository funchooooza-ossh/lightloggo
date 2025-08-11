from enum import IntEnum

try:
    from enum import StrEnum
except ImportError:
    import enum

    class StrEnum(str, enum.Enum):
        def __str__(self) -> str:
            return str(self.value)

        def __repr__(self) -> str:
            return f"{self.__class__.__name__}.{self.name}"


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
        if level == Level.Trace:
            return cls.Trace
        elif level == Level.Debug:
            return cls.Debug
        elif level == Level.Info:
            return cls.Info
        elif level == Level.Warning:
            return cls.Warning
        elif level == Level.Error:
            return cls.Error
        elif level == Level.Exception:
            return cls.Exception
