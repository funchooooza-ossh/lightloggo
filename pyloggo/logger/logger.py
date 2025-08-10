from ..ffi.ffi import log_call, _as_bytes
from ..json import _serialize_fields
from ..route import RouteProcessor
from ..c import CLogger
import sys
import linecache
import os
from ..enums import LogLevel
from typing import Any 


class _Logger:
    def __init__(self, routes: list[RouteProcessor],
        tb: bool = False,
        tb_max_depth: int = 10,
        tb_level: int = 50,
        scope: bool = True) -> None:
        route_ids = [r.id for r in routes]
        self._c_logger = CLogger(route_ids)
        self._routes = routes
        self._tb = tb
        self._tb_max_depth=tb_max_depth
        self._scope = scope
        self._tb_level=tb_level

    @property
    def id(self) -> int:
        return self._c_logger._id

    def _log(self, method: str, msg: str, **kwargs) -> None:
        level = getattr(LogLevel, method.capitalize()) 
        msg_b = _as_bytes(msg)
        if not kwargs:
            fields_b = b"0"
        else:
            fields_b = _serialize_fields(self._resolve_fields(level, kwargs))
        log_call(method, self.id, msg_b, fields_b)

    def _resolve_fields(
        self,
        level: LogLevel,
        fields: dict[str, Any],
    ) -> dict[str, Any]:
        fields_cp = dict(fields)
        tb = False
        if self._tb and self._tb_level <= level:
            fields_cp["tb"] = self._add_traceback(max_depth=self._tb_max_depth)
            tb = True
        if not tb and self._scope:
            fields_cp["scope"] = self._add_scope()

        return fields_cp

    @staticmethod
    def _add_scope(frame_depth: int = 5) -> str:
        try:
            frame = sys._getframe(frame_depth)
            filename = os.path.basename(frame.f_code.co_filename)
            lineno = frame.f_lineno
            func = frame.f_code.co_name
            return f"{filename}:{lineno} in {func}()"
        except Exception:
            return "<scope unavailable>"

    @staticmethod
    def _add_traceback(max_depth: int = 10, skip: int = 5) -> str:
        lines = []
        frame = sys._getframe(skip)

        for _ in range(max_depth):
            if frame is None:
                break

            filename_full = frame.f_code.co_filename
            filename = os.path.basename(filename_full)
            lineno = frame.f_lineno
            func = frame.f_code.co_name

            code_line = linecache.getline(filename_full, lineno).strip()

            lines.append(
                f'  File "{filename}", line {lineno}, in {func}()\n    {code_line}\n'
            )

            frame = frame.f_back

        return "".join(lines)

    def trace(self, msg: str, **kwargs):
        self._log("trace", msg, **kwargs)

    def debug(self, msg: str, **kwargs):
        self._log("debug", msg, **kwargs)

    def info(self, msg: str, **kwargs):
        self._log("info", msg, **kwargs)

    def warning(self, msg: str, **kwargs):
        self._log("warning", msg, **kwargs)

    def error(self, msg: str, **kwargs):
        self._log("error", msg, **kwargs)

    def exception(self, msg: str, **kwargs):
        self._log("exception", msg, **kwargs)

    def close(self):
        self._c_logger.close()

    def __del__(self):
        try:
            self.close()
        except Exception:
            pass


def create_default_logger() -> _Logger:
    router = RouteProcessor()
    return _Logger([router])


