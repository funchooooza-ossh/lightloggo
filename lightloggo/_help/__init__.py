"""
Internal helper utilities package.

This package provides low-level, internal helper functions for the lightloggo
library. These functions are typically performance-oriented and support the
process of preparing Python data for transit across the FFI boundary to the
Go core.

This module is a private implementation detail and is not part of the public API.
"""


def _serialize_fields(fields: dict) -> bytes:
    """
    Serializes a dictionary of fields into a null-terminated byte string.

    This function implements the specific serialization format expected by the
    Go core library for structured logging fields. It converts a Python
    dictionary into a single byte string where keys and values are separated
    by a null byte (`\x00`).

    The resulting format is: `key1\x00value1\x00key2\x00value2\x00`

    Args:
        fields (dict): A dictionary of key-value pairs to serialize.

    Returns:
        bytes: A byte string representing the serialized dictionary.
    """
    # Use a mutable bytearray for efficient concatenation in a loop.
    buf = bytearray()

    for k, v in fields.items():
        # Keys and values are converted to strings to ensure they can be
        # safely encoded.
        buf.extend(str(k).encode("utf-8"))
        # A null byte (`\x00`) is used as the delimiter between a key and its value.
        buf.append(0)

        buf.extend(str(v).encode("utf-8"))
        # A null byte is also used as the delimiter between one entry and the next key.
        buf.append(0)

    # Return an immutable bytes object.
    return bytes(buf)
