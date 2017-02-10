# Whiskey

Whiskey is a pre-fork Python WSGI server written in Go. It's intended as a
proof-of-concept alternative to something like [gunicorn]. If you had a
standard WSGI application file called hello.py like this one:

```python
def application(environ, start_response):
    start_response('200 OK', [('Content-Type', 'text/plain')])
    return ['hello', ' ', 'world']
```

You could run Whiskey like so:

```
whiskey -addr 127.0.0.1:8080 hello:application
```

(Note that hello.py must be importable by Python, so make sure that the
folder containing it has been added to your PYTHONPATH when you run Whiskey.)

This is currently written for Python 2.7, with plans for Python 3 in the
future at some point.

## Caveats

This is far from complete, and isn't intended for use in anything real. It's
a toy to explore interactions between Go and Python.

[gunicorn]: http://gunicorn.org