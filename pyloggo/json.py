from typing import Any

_TABLE = str.maketrans({
    "\\": "\\\\",
    '"':  '\\"',
    "\n": "\\n",
    "\r": "\\r",
    "\t": "\\t",
})
def _escape_string(s: str) -> str:
    return '"' + s.translate(_TABLE) + '"'




def _serialize_value(val: Any, out: list[str]) -> None:
    if isinstance(val, str):
        out.append(_escape_string(val))
    elif isinstance(val, bool):
        out.append("true" if val else "false")
    elif val is None:
        out.append("null")
    elif isinstance(val, (int, float)):
        out.append(str(val))
    elif isinstance(val, dict):
        out.append("{")
        first = True
        for k, v in val.items():
            if not first:
                out.append(", ")
            first = False
            out.append(_escape_string(str(k))); out.append(": ")
            _serialize_value(v, out)
        out.append("}")
    elif isinstance(val, list):
        out.append("[")
        first = True
        for v in val:
            if not first:
                out.append(", ")
            first = False
            _serialize_value(v, out)
        out.append("]")
    else:
        out.append(_escape_string(repr(val)))

def _serialize_fields(fields: dict[str, Any]) -> bytes:
    out: list[str] = ["{"]
    first = True
    for k, v in fields.items():
        if not first:
            out.append(", ")
        first = False
        out.append(_escape_string(str(k))); out.append(": ")
        _serialize_value(v, out)
    out.append("}")
    return "".join(out).encode("utf-8", "replace")

