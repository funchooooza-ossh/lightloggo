from .ffi import lib


class FormatStyle:
    def __init__(
        self,
        color_keys=True,
        color_values=True,
        color_level=False,
        key_color="\033[34m",
        value_color="\033[33m",
        reset="\033[0m",
    ):
        self._id = lib.NewFormatStyle(
            int(color_keys),
            int(color_values),
            int(color_level),
            key_color.encode(),
            value_color.encode(),
            reset.encode(),
        )
