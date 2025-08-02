from ...ffi.ffi import lib


class CJsonFormatter:
    def __init__(self, style_id: int = 0):
        self._id = lib.NewJsonFormatter(style_id)
