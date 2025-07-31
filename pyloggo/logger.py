from .ffi import lib
import json
import ctypes
from .route import RouteProcessor


import threading


class Logger:
    def __init__(self, routes: list["RouteProcessor"]):
        arr_type = ctypes.c_ulong * len(routes)
        route_ids = arr_type(*(r._id for r in routes))
        self._id = lib.NewLoggerWithRoutes(route_ids, len(routes))

    def _log(self, method: str, msg: str, **kwargs):
        msg_b = msg.encode()
        fields_b = json.dumps(kwargs or {}).encode()
        getattr(lib, f"Logger_{method.capitalize()}WithFields")(
            self._id, msg_b, fields_b
        )

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
        lib.Logger_Close(self._id)

    def __del__(self):
        lib.FreeLogger(self._id)


def create_default_logger() -> Logger:
    logger_id = lib.NewDefaultLogger()
    logger = Logger.__new__(Logger)
    logger._id = logger_id
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

    def configure(self, routes: list):
        with self._lock:
            self._logger.close()
            self._logger = Logger(routes=routes)

    def close(self):
        with self._lock:
            self._logger.close()

    def __del__(self):
        lib.FreeLogger(self._logger._id)
