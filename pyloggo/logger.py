from .ffi.ffi import lib
import json
from .route import RouteProcessor
from .c import CLogger
import sys
import linecache
import os
from .enums import LogLevel

import threading


class Logger:
    def __init__(self, routes: list["RouteProcessor"], tb_level: int = 50):
        route_ids = [r.id for r in routes]
        self._c_logger = CLogger(route_ids)
        self._tb_level = tb_level

    @property
    def id(self) -> int:
        return self._c_logger._id

    def _log(self, method: str, msg: str, **kwargs):
        msg_b = msg.encode()

        if LogLevel._from_string(method.capitalize()) >= self._tb_level:
            kwargs["tb"] = self._add_traceback()
        else:
            kwargs["scope"] = self._add_scope()

        fields_b = json.dumps(kwargs or {}).encode()
        getattr(lib, f"Logger_{method.capitalize()}WithFields")(
            self.id, msg_b, fields_b
        )

    def _add_scope(self, frame_depth: int = 4) -> str:
        try:
            frame = sys._getframe(frame_depth)
            filename = os.path.basename(frame.f_code.co_filename)
            lineno = frame.f_lineno
            func = frame.f_code.co_name
            return f"{filename}:{lineno} in {func}()"
        except Exception:
            return "<scope unavailable>"

    def _add_traceback(self, max_depth: int = 10, skip: int = 4) -> str:
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


def create_default_logger() -> Logger:
    logger_id = lib.NewDefaultLogger()
    logger = Logger.__new__(Logger)
    logger._c_logger = CLogger.from_id(logger_id)
    return logger


class GlobalLogger:
    def __init__(self):
        self._lock = threading.Lock()
        self._logger = self._create_default_logger()

    def _create_default_logger(self) -> Logger:
        return create_default_logger()

    def info(self, msg, **kwargs):
        self._logger.info(msg, **kwargs)

    def debug(self, msg, **kwargs):
        self._logger.debug(msg, **kwargs)

    def warning(self, msg, **kwargs):
        self._logger.warning(msg, **kwargs)

    def error(self, msg, **kwargs):
        self._logger.error(msg, **kwargs)

    def exception(self, msg, **kwargs):
        self._logger.exception(msg, **kwargs)

    def trace(self, msg, **kwargs):
        self._logger.trace(msg, **kwargs)

    def add(self, route: RouteProcessor):
        with self._lock:
            # пересоздаём logger с новым роутом
            self._logger.close()
            self._logger = Logger(routes=[route])

    def remove(self):
        with self._lock:
            self._logger.close()
            self._logger = self._create_default_logger()

    def configure(self, routes: list, tb_level: LogLevel = 50):
        with self._lock:
            self._logger.close()
            self._logger = Logger(routes=list(routes), tb_level=tb_level)

    def close(self):
        with self._lock:
            self._logger.close()

    def __del__(self):
        try:
            self._logger.close()
        except Exception:
            pass
