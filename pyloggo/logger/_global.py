import threading
from .logger import create_default_logger, _Logger
from ..route import RouteProcessor

class Logger:
    def __init__(self):
        self._lock = threading.Lock()
        self._logger = self._create_default_logger()

    def _create_default_logger(self) -> _Logger:
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
            self._logger = _Logger(routes=[route])

    def remove(self):
        with self._lock:
            self._logger.close()
            self._logger = self._create_default_logger()

    def configure(self, routes: list,         
        tb: bool = False,
        tb_max_depth: int = 10,
        tb_level: int = 50,
        scope: bool = True):
        with self._lock:
            self._logger.close()
            self._logger = _Logger(routes=list(routes), tb=tb, tb_max_depth=tb_max_depth, tb_level=tb_level, scope=scope)

    def close(self):
        with self._lock:
            self._logger.close()

    def __del__(self):
        try:
            self._logger.close()
        except Exception:
            pass
