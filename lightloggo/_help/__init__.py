def _serialize_fields(fields: dict) -> bytes:
    buf = bytearray()
    for k, v in fields.items():
        buf.extend(str(k).encode("utf-8"))
        buf.append(0)  # разделитель
        buf.extend(str(v).encode("utf-8"))
        buf.append(0)
    return bytes(buf)
