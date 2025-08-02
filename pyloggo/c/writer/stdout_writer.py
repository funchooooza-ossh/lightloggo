from ...ffi.ffi import lib


class CStdoutWriter:
    def __init__(self):
        self._id = lib.NewStdoutWriter()
