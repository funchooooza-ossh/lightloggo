from ..c import CRouter


class router:
    _c_router: CRouter

    @property
    def id(self) -> int:
        return self._c_router._id
