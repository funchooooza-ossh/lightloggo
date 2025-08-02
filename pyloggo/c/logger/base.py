from ...ffi.ffi import lib
import ctypes


class CLogger:
    def __init__(self, routes: list[int]):
        arr_type = ctypes.c_ulong * len(routes)
        route_ids = arr_type(*routes)
        self._id = lib.NewLoggerWithRoutes(route_ids, len(routes))

    def close(self):
        if self._id:
            lib.Logger_Close(self._id)
            lib.FreeLogger(self._id)
            self._id = 0  # защитное обнуление

    @classmethod
    def from_id(cls, id_: int):
        obj = cls.__new__(cls)
        obj._id = id_
        return obj

    def __del__(self):
        try:
            self.close()
        except Exception:
            pass
