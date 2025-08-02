from typing import Any


def _escape_string(s: str) -> str:
    return (
        '"'
        + s.replace("\\", "\\\\")
        .replace('"', '\\"')
        .replace("\n", "\\n")
        .replace("\r", "\\r")
        .replace("\t", "\\t")
        + '"'
    )


def _serialize_value(val: Any) -> str:
    if isinstance(val, str):
        return _escape_string(val)
    elif isinstance(val, bool):
        return "true" if val else "false"
    elif val is None:
        return "null"
    elif isinstance(val, (int, float)):
        return str(val)
    elif isinstance(val, dict):
        items = ", ".join(
            f"{_escape_string(str(k))}: {_serialize_value(v)}" for k, v in val.items()
        )
        return f"{{{items}}}"
    elif isinstance(val, list):
        items = ", ".join(_serialize_value(v) for v in val)
        return f"[{items}]"
    else:
        return _escape_string(repr(val))  # fallback


def _serialize_fields(fields: dict[str, Any]) -> bytes:
    items = ", ".join(
        f"{_escape_string(str(k))}: {_serialize_value(v)}" for k, v in fields.items()
    )
    return f"{{{items}}}".encode("utf-8")
