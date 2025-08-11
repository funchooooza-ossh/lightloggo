from ..c import CWriter


class writer:
    _c_writer: CWriter

    @property
    def id(self) -> int:
        return self._c_writer._id
