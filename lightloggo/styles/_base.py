from ..c import CFormatStyle, CStyle


class BaseStyle:
    _c_style: CStyle

    @property
    def id(self) -> int:
        if not hasattr(self, "_c_style"):
            raise RuntimeError("C style not initialized")
        return self._c_style._id


class FormatStyle(BaseStyle):
    def __init__(
        self,
        color_keys: bool = True,
        color_values: bool = True,
        color_level: bool = False,
        key_color: str = "\033[34m",
        value_color: str = "\033[33m",
        reset: str = "\033[0m",
    ) -> None:
        self._c_style = CFormatStyle(
            color_keys=color_keys,
            color_values=color_values,
            color_level=color_level,
            key_color=key_color,
            value_color=value_color,
            reset=reset,
        )
