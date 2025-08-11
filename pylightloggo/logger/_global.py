import contextlib
import threading
from typing import Any, List

from ..routers import RouteProcessor
from .logger import _Logger, create_default_logger


class Logger:
    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._logger = self._create_default_logger()

    def _create_default_logger(self) -> _Logger:
        return create_default_logger()

    def info(self, msg: str, **kwargs: Any) -> None:
        self._logger.info(msg, **kwargs)

    def debug(self, msg: str, **kwargs: Any) -> None:
        self._logger.debug(msg, **kwargs)

    def warning(self, msg: str, **kwargs: Any) -> None:
        self._logger.warning(msg, **kwargs)

    def error(self, msg: str, **kwargs: Any) -> None:
        self._logger.error(msg, **kwargs)

    def exception(self, msg: str, **kwargs: Any) -> None:
        self._logger.exception(msg, **kwargs)

    def trace(self, msg: str, **kwargs: Any) -> None:
        self._logger.trace(msg, **kwargs)

    def add(self, route: RouteProcessor) -> None:
        with self._lock:
            # пересоздаём logger с новым роутом
            self._logger.close()
            self._logger = _Logger(routes=[route])

    def remove(self) -> None:
        with self._lock:
            self._logger.close()
            self._logger = self._create_default_logger()

    def configure(
        self,
        routes: List[RouteProcessor],
        tb: bool = False,
        tb_max_depth: int = 10,
        tb_level: int = 50,
        scope: bool = True,
    ) -> None:
        with self._lock:
            self._logger.close()
            self._logger = _Logger(
                routes=list(routes),
                tb=tb,
                tb_max_depth=tb_max_depth,
                tb_level=tb_level,
                scope=scope,
            )

    def close(self) -> None:
        with self._lock:
            self._logger.close()

    def __del__(self) -> None:
        with contextlib.suppress(Exception):
            self._logger.close()
