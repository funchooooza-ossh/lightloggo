# pyloggo

**pyloggo** is a structured, high-performance logging system built in Go with a Python FFI bridge. It allows fine-grained customization of log formatting, routing, and output, while exposing a clean Pythonic interface inspired by [loguru](https://github.com/Delgan/loguru).

## Why

I’ve never been satisfied with existing logging libraries:

- Some are outdated and difficult to extend  
- Others lack structured logging and proper routing  
- Many are too slow for production-scale use  

So I decided to write my own lightweight logging system — primarily for learning, performance, and control.

## Overview

- The core is implemented in Go for speed and async I/O  
- Log routing is configurable with formatters and writers  
- JSON and colorized text formats are supported  
- Log rotation and compression (e.g. `.gz`) built in  
- Exposed to Python via `ctypes` using `.so` bridge  
- Python side gives an interface similar to `loguru`

## Usage

```python
from pyloggo import (
    logger,
    FormatStyle,
    JsonFormatter,
    TextFormatter,
    FileWriter,
    StdoutWriter,
    RouteProcessor,
)

def configure():
    style = FormatStyle(color_keys=False, color_values=False)
    formatter1 = JsonFormatter(style)
    writer1 = FileWriter("test/logs/app.json", max_backups=3)
    router1 = RouteProcessor(formatter=formatter1, writer=writer1, level=10)

    style2 = FormatStyle()  # colorized
    formatter2 = TextFormatter(style2)
    writer2 = StdoutWriter()
    router2 = RouteProcessor(formatter=formatter2, writer=writer2, level=10)

    logger.configure([router1, router2])  # replace global logger
    logger.info("logger configured", stage="test")

```
Then anywhere else:
```python
from pyloggo import logger

logger.info("something happened", user="admin", request_id="abc123")
logger.error("oops", exception="db crash")
```

## Notes
- I'm not a Go developer — I just use it for fun
- This is my first time writing a full-fledged library

## License
***MIT***