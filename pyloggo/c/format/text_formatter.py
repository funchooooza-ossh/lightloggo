from ...ffi.ffi import lib


class CTextFormatter:
    def __init__(self, style_id: int = 0):
        self._id = lib.NewTextFormatter(style_id)
