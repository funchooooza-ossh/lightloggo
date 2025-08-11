from ..c import CFormatter


class formatter:
    _c_formatter: CFormatter

    @property
    def id(self) -> int:
        return self._c_formatter._id
