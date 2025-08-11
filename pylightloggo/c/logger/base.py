from __future__ import annotations

import contextlib
import ctypes
from typing import List

from ...ffi.ffi import lib


class CLogger:
    def __init__(self, routes: List[int]) -> None:
        arr_type = ctypes.c_ulong * len(routes)
        route_ids = arr_type(*routes)
        self._id = lib.NewLoggerWithRoutes(route_ids, len(routes))

    def close(self) -> None:
        if self._id:
            lib.Logger_Close(self._id)
            lib.FreeLogger(self._id)
            self._id = 0  # защитное обнуление

    @classmethod
    def from_id(cls, id_: int) -> CLogger:
        obj = cls.__new__(cls)
        obj._id = id_
        return obj

    def __del__(self) ->  None:
        with contextlib.suppress(Exception):
            self.close()
