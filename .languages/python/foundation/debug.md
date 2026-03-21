# Python Debugging Comprehensive Guide

Master Python debugging techniques, tools, and best practices for efficient problem-solving and code analysis.

## Built-in Debugging Tools

### Python Debugger (pdb)

```python
import pdb

# Basic debugging with pdb
def buggy_function(numbers):
    total = 0
    for i in range(len(numbers)):
        pdb.set_trace()  # Set breakpoint here
        total += numbers[i] * 2
    return total

# Using breakpoint() (Python 3.7+)
def better_debugging(numbers):
    total = 0
    for i in range(len(numbers)):
        breakpoint()  # Modern way to set breakpoint
        total += numbers[i] * 2
    return total

# Post-mortem debugging
def crash_function():
    x = 1
    y = 0
    result = x / y  # This will cause ZeroDivisionError
    return result

# Run with post-mortem debugging
try:
    crash_function()
except:
    import pdb
    pdb.post_mortem()  # Debug the exception
```

### pdb Commands Reference

```python
# Essential pdb commands
"""
Navigation:
  n (next)         - Execute next line
  s (step)         - Step into function calls
  c (continue)     - Continue execution
  r (return)       - Continue until return from current function
  u (up)           - Move up one stack frame
  d (down)         - Move down one stack frame

Inspection:
  l (list)         - Show current code
  ll (longlist)    - Show current function
  w (where)        - Show stack trace
  p <var>          - Print variable value
  pp <var>         - Pretty-print variable value
  whatis <var>     - Show type of variable
  source <obj>     - Show source code of object

Breakpoints:
  b (break)        - Set breakpoint
  b <line>         - Set breakpoint at line
  b <file>:<line>  - Set breakpoint at file:line
  cl (clear)       - Clear breakpoints
  disable <bp#>    - Disable breakpoint
  enable <bp#>     - Enable breakpoint

Execution:
  run <args>       - Restart with arguments
  q (quit)         - Quit debugger
  h (help)         - Show help
"""

# Example debugging session
def fibonacci(n):
    if n <= 1:
        return n
    else:
        return fibonacci(n-1) + fibonacci(n-2)

def debug_fibonacci():
    import pdb; pdb.set_trace()
    result = fibonacci(5)
    print(f"Result: {result}")

# Advanced pdb usage
def conditional_debugging():
    data = list(range(100))
    for i, value in enumerate(data):
        # Conditional breakpoint
        if value > 50:
            pdb.set_trace()
        process_value(value)

def process_value(value):
    return value * 2
```

### IPython Debugger (ipdb)

```bash
# Install ipdb
pip install ipdb

# Use in code
import ipdb; ipdb.set_trace()

# Or set as default debugger
export PYTHONBREAKPOINT=ipdb.set_trace
```

```python
# Enhanced debugging with ipdb
import ipdb

def enhanced_debugging_example():
    data = {'a': 1, 'b': 2, 'c': 3}

    # IPython debugger provides better interface
    ipdb.set_trace()

    # You can use IPython magic commands in ipdb
    # %timeit, %run, %load, etc.

    processed = {k: v * 2 for k, v in data.items()}
    return processed
```

## Advanced Debugging Techniques

### Remote Debugging

```python
# Remote debugging with pdb
import pdb

# Start pdb server
pdb.Pdb().set_trace(frame=None)

# For remote debugging over network
import socket
import pdb

class RemotePdb(pdb.Pdb):
    def __init__(self, host='localhost', port=4444):
        self.old_stdout = sys.stdout
        self.old_stdin = sys.stdin
        self.skt = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.skt.bind((host, port))
        self.skt.listen(1)
        (clientsocket, address) = self.skt.accept()
        handle = clientsocket.makefile('rw')
        pdb.Pdb.__init__(self, completekey='tab', stdin=handle, stdout=handle)
        sys.stdout = sys.stdin = handle

def remote_debug():
    RemotePdb().set_trace()  # Connect via telnet localhost 4444
```

### Web-based Debugging

```python
# Web debugging with web-pdb
# pip install web-pdb

import web_pdb

def web_debugging_example():
    data = [1, 2, 3, 4, 5]
    web_pdb.set_trace()  # Access via browser at localhost:5555

    processed = []
    for item in data:
        processed.append(item * 2)

    return processed
```

## IDE Debugging Integration

### Visual Studio Code Debugging

```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Python: Current File",
            "type": "python",
            "request": "launch",
            "program": "${file}",
            "console": "integratedTerminal",
            "cwd": "${workspaceFolder}",
            "env": {"PYTHONPATH": "${workspaceFolder}"}
        },
        {
            "name": "Python: Django",
            "type": "python",
            "request": "launch",
            "program": "${workspaceFolder}/manage.py",
            "args": ["runserver", "0.0.0.0:8000"],
            "django": true,
            "env": {"DEBUG": "True"}
        },
        {
            "name": "Python: Flask",
            "type": "python",
            "request": "launch",
            "program": "${workspaceFolder}/app.py",
            "env": {
                "FLASK_APP": "app.py",
                "FLASK_ENV": "development"
            },
            "jinja": true
        },
        {
            "name": "Python: pytest",
            "type": "python",
            "request": "launch",
            "module": "pytest",
            "args": ["${workspaceFolder}/tests", "-v"],
            "console": "integratedTerminal"
        },
        {
            "name": "Python: Attach to Process",
            "type": "python",
            "request": "attach",
            "processId": "${command:pickProcess}",
            "justMyCode": false
        }
    ]
}
```

### PyCharm Debugging

```python
# PyCharm debugging features:
# - Visual breakpoints
# - Conditional breakpoints
# - Exception breakpoints
# - Smart step into
# - Evaluate expressions
# - Watches
# - Debug console

# Example with PyCharm debugging
def complex_calculation(data):
    # Set breakpoint here in PyCharm
    result = []
    for item in data:
        # Evaluate expressions in debug console
        intermediate = item ** 2
        if intermediate > 100:  # Add watch for this condition
            intermediate = intermediate / 2
        result.append(intermediate)
    return result

# Debug configuration in PyCharm:
# Run -> Edit Configurations
# Add Python configuration with:
# - Script path or module
# - Parameters
# - Environment variables
# - Working directory
```

## Logging for Debugging

### Structured Logging

```python
import logging
import json
from datetime import datetime

# Configure logging for debugging
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('debug.log'),
        logging.StreamHandler()
    ]
)

logger = logging.getLogger(__name__)

def debug_with_logging():
    logger.debug("Starting debug_with_logging function")

    data = {'users': [{'id': 1, 'name': 'Alice'}, {'id': 2, 'name': 'Bob'}]}
    logger.debug(f"Initial data: {json.dumps(data, indent=2)}")

    try:
        for user in data['users']:
            logger.debug(f"Processing user: {user}")
            process_user(user)
    except Exception as e:
        logger.exception("Exception occurred while processing users")
        raise

    logger.debug("Completed debug_with_logging function")

def process_user(user):
    logger.debug(f"process_user called with: {user}")
    # Processing logic here
    result = user['name'].upper()
    logger.debug(f"process_user returning: {result}")
    return result

# Advanced logging with context
import contextlib
import threading

class DebugContext:
    def __init__(self):
        self.data = threading.local()

    def set_context(self, **kwargs):
        for key, value in kwargs.items():
            setattr(self.data, key, value)

    def get_context(self):
        return {k: v for k, v in self.data.__dict__.items()
                if not k.startswith('_')}

    @contextlib.contextmanager
    def context(self, **kwargs):
        old_context = getattr(self.data, 'context', {})
        new_context = {**old_context, **kwargs}
        self.data.context = new_context
        try:
            yield
        finally:
            self.data.context = old_context

debug_context = DebugContext()

class ContextFilter(logging.Filter):
    def filter(self, record):
        record.debug_context = debug_context.get_context()
        return True

# Add context filter to logger
logger.addFilter(ContextFilter())

def contextual_debugging():
    with debug_context.context(user_id=123, session_id='abc'):
        logger.info("Processing user request")
        # Context automatically included in log messages
```

### Debugging Decorators

```python
from functools import wraps
import time
import traceback

def debug_calls(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        logger.debug(f"Calling {func.__name__}")
        logger.debug(f"Args: {args}")
        logger.debug(f"Kwargs: {kwargs}")

        start_time = time.time()
        try:
            result = func(*args, **kwargs)
            end_time = time.time()
            logger.debug(f"{func.__name__} completed in {end_time - start_time:.4f}s")
            logger.debug(f"Returned: {result}")
            return result
        except Exception as e:
            logger.exception(f"Exception in {func.__name__}: {e}")
            raise
    return wrapper

def trace_calls(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        stack = traceback.extract_stack()
        caller = stack[-2]
        logger.debug(f"Called {func.__name__} from {caller.filename}:{caller.lineno}")
        return func(*args, **kwargs)
    return wrapper

# Usage examples
@debug_calls
@trace_calls
def calculate_fibonacci(n):
    if n <= 1:
        return n
    return calculate_fibonacci(n-1) + calculate_fibonacci(n-2)

# result = calculate_fibonacci(5)
```

## Profiling and Performance Debugging

### cProfile and pstats

```python
import cProfile
import pstats
from io import StringIO

def profile_function():
    # Profile a specific function
    pr = cProfile.Profile()
    pr.enable()

    # Your code here
    result = expensive_calculation()

    pr.disable()

    # Analyze results
    s = StringIO()
    ps = pstats.Stats(pr, stream=s)
    ps.sort_stats('cumulative')
    ps.print_stats(10)  # Top 10 functions

    print(s.getvalue())
    return result

def expensive_calculation():
    total = 0
    for i in range(1000000):
        total += i ** 2
    return total

# Command line profiling
# python -m cProfile -o profile.stats your_script.py

# Analyzing profile results
def analyze_profile():
    stats = pstats.Stats('profile.stats')

    # Sort by cumulative time
    stats.sort_stats('cumulative')
    stats.print_stats(10)

    # Filter specific functions
    stats.print_stats('your_function_name')

    # Callers and callees
    stats.print_callers('expensive_function')
    stats.print_callees('expensive_function')
```

### Line Profiler

```bash
# Install line_profiler
pip install line_profiler

# Usage
kernprof -l -v script.py
```

```python
# Line-by-line profiling
@profile
def line_by_line_analysis():
    data = list(range(1000))

    # These lines will be profiled
    result = []
    for item in data:
        if item % 2 == 0:
            result.append(item ** 2)
        else:
            result.append(item ** 3)

    return result
```

### Memory Profiling

```bash
# Install memory_profiler
pip install memory_profiler psutil
```

```python
from memory_profiler import profile
import psutil
import os

@profile
def memory_intensive_function():
    # Large list creation
    big_list = [i for i in range(1000000)]

    # Dictionary comprehension
    big_dict = {i: i**2 for i in range(100000)}

    # String concatenation (memory intensive)
    big_string = ""
    for i in range(10000):
        big_string += str(i)

    return len(big_list), len(big_dict), len(big_string)

# Monitor memory usage
def monitor_memory():
    process = psutil.Process(os.getpid())

    print(f"Memory before: {process.memory_info().rss / 1024 / 1024:.1f} MB")
    result = memory_intensive_function()
    print(f"Memory after: {process.memory_info().rss / 1024 / 1024:.1f} MB")

    return result

# Memory profiling with tracemalloc (Python 3.4+)
import tracemalloc

def trace_memory_usage():
    tracemalloc.start()

    # Take snapshot before
    snapshot1 = tracemalloc.take_snapshot()

    # Your code here
    data = [i**2 for i in range(100000)]

    # Take snapshot after
    snapshot2 = tracemalloc.take_snapshot()

    # Compare snapshots
    top_stats = snapshot2.compare_to(snapshot1, 'lineno')

    print("Top 10 memory allocations:")
    for stat in top_stats[:10]:
        print(stat)
```

## Debugging Specific Scenarios

### Debugging Async Code

```python
import asyncio
import logging

# Configure async debugging
logging.basicConfig(level=logging.DEBUG)
asyncio_logger = logging.getLogger('asyncio')

async def debug_async_function():
    logger.debug("Starting async function")

    # Simulate async work
    await asyncio.sleep(1)

    logger.debug("Async work completed")
    return "result"

async def debug_concurrent_tasks():
    tasks = [
        asyncio.create_task(debug_async_function(), name=f"task-{i}")
        for i in range(3)
    ]

    # Wait for all tasks with debugging
    results = await asyncio.gather(*tasks, return_exceptions=True)

    for i, result in enumerate(results):
        if isinstance(result, Exception):
            logger.error(f"Task {i} failed: {result}")
        else:
            logger.debug(f"Task {i} result: {result}")

    return results

# Debug with event loop inspection
def debug_event_loop():
    loop = asyncio.get_event_loop()

    # Enable debug mode
    loop.set_debug(True)

    # Run with debugging
    result = loop.run_until_complete(debug_concurrent_tasks())

    return result

# Debugging hanging async code
import signal

def timeout_handler(signum, frame):
    import traceback
    print("\\n".join(traceback.format_stack()))
    raise TimeoutError("Function call timed out")

async def potentially_hanging_function():
    # Set timeout for debugging
    signal.signal(signal.SIGALRM, timeout_handler)
    signal.alarm(5)  # 5 second timeout

    try:
        await asyncio.sleep(10)  # This will timeout
    finally:
        signal.alarm(0)  # Cancel timeout
```

### Debugging Multi-threading

```python
import threading
import time
import logging
from concurrent.futures import ThreadPoolExecutor

# Thread-aware logging
def configure_thread_logging():
    formatter = logging.Formatter(
        '%(asctime)s - %(threadName)s - %(name)s - %(levelname)s - %(message)s'
    )

    handler = logging.StreamHandler()
    handler.setFormatter(formatter)

    logger = logging.getLogger()
    logger.addHandler(handler)
    logger.setLevel(logging.DEBUG)

def debug_threading():
    configure_thread_logging()

    def worker(name, duration):
        logger = logging.getLogger(__name__)
        logger.debug(f"Worker {name} starting")

        time.sleep(duration)

        logger.debug(f"Worker {name} completed")
        return f"Result from {name}"

    with ThreadPoolExecutor(max_workers=3) as executor:
        futures = [
            executor.submit(worker, f"thread-{i}", i)
            for i in range(5)
        ]

        for future in futures:
            result = future.result()
            logger.debug(f"Got result: {result}")

# Debugging race conditions
import threading

class ThreadSafeCounter:
    def __init__(self):
        self._value = 0
        self._lock = threading.Lock()
        self._debug = True

    def increment(self):
        thread_name = threading.current_thread().name

        if self._debug:
            logger.debug(f"{thread_name} attempting to acquire lock")

        with self._lock:
            if self._debug:
                logger.debug(f"{thread_name} acquired lock, current value: {self._value}")

            old_value = self._value
            time.sleep(0.001)  # Simulate race condition
            self._value = old_value + 1

            if self._debug:
                logger.debug(f"{thread_name} incremented value to: {self._value}")

    @property
    def value(self):
        return self._value

def test_race_condition():
    counter = ThreadSafeCounter()

    def worker():
        for _ in range(100):
            counter.increment()

    threads = [threading.Thread(target=worker) for _ in range(10)]

    for thread in threads:
        thread.start()

    for thread in threads:
        thread.join()

    print(f"Final counter value: {counter.value}")
```

### Debugging Web Applications

```python
# Django debugging
import django
from django.conf import settings

# Flask debugging
from flask import Flask
import logging

def setup_flask_debugging():
    app = Flask(__name__)
    app.debug = True

    # Add request logging
    @app.before_request
    def log_request_info():
        app.logger.debug('Headers: %s', request.headers)
        app.logger.debug('Body: %s', request.get_data())

    @app.after_request
    def log_response_info(response):
        app.logger.debug('Response Status: %s', response.status)
        return response

    return app

# FastAPI debugging
from fastapi import FastAPI, Request
import logging

def setup_fastapi_debugging():
    app = FastAPI()

    @app.middleware("http")
    async def log_requests(request: Request, call_next):
        logger.debug(f"Request: {request.method} {request.url}")
        response = await call_next(request)
        logger.debug(f"Response: {response.status_code}")
        return response

    return app
```

## Advanced Debugging Tools

### Python Tracing

```python
import sys
import linecache

def trace_calls(frame, event, arg):
    if event == 'call':
        filename = frame.f_code.co_filename
        lineno = frame.f_lineno
        line = linecache.getline(filename, lineno)
        print(f"Calling: {filename}:{lineno} {line.strip()}")
    elif event == 'return':
        print(f"Returning: {arg}")
    return trace_calls

def trace_execution():
    sys.settrace(trace_calls)

    # Your code here
    result = fibonacci(5)

    sys.settrace(None)
    return result

# Custom tracer for specific functions
class FunctionTracer:
    def __init__(self, target_functions):
        self.target_functions = target_functions
        self.call_stack = []

    def trace_calls(self, frame, event, arg):
        if event == 'call':
            func_name = frame.f_code.co_name
            if func_name in self.target_functions:
                self.call_stack.append(func_name)
                indent = "  " * len(self.call_stack)
                print(f"{indent}-> {func_name}()")

        elif event == 'return':
            func_name = frame.f_code.co_name
            if func_name in self.target_functions and self.call_stack:
                indent = "  " * len(self.call_stack)
                print(f"{indent}<- {func_name}() = {arg}")
                self.call_stack.pop()

        return self.trace_calls

def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

# Usage
tracer = FunctionTracer(['fibonacci'])
sys.settrace(tracer.trace_calls)
result = fibonacci(4)
sys.settrace(None)
```

### Debug Utilities

```python
import inspect
import pprint
from collections import defaultdict

class DebugUtils:
    @staticmethod
    def inspect_object(obj):
        print(f"Object: {obj}")
        print(f"Type: {type(obj)}")
        print(f"ID: {id(obj)}")
        print(f"Dir: {dir(obj)}")

        if hasattr(obj, '__dict__'):
            print("Attributes:")
            pprint.pprint(obj.__dict__)

    @staticmethod
    def get_caller_info():
        frame = inspect.currentframe().f_back
        return {
            'filename': frame.f_code.co_filename,
            'function': frame.f_code.co_name,
            'lineno': frame.f_lineno
        }

    @staticmethod
    def dump_locals():
        frame = inspect.currentframe().f_back
        locals_dict = frame.f_locals
        print("Local variables:")
        pprint.pprint(locals_dict)

    @staticmethod
    def call_graph():
        stack = inspect.stack()
        print("Call stack:")
        for i, frame_info in enumerate(stack):
            indent = "  " * i
            print(f"{indent}{frame_info.filename}:{frame_info.lineno} in {frame_info.function}")

# Usage examples
def example_function(x, y):
    z = x + y
    DebugUtils.dump_locals()  # Shows x, y, z
    DebugUtils.call_graph()   # Shows call stack
    return z

# Debug context manager
class DebugContext:
    def __init__(self, name):
        self.name = name

    def __enter__(self):
        print(f"Entering debug context: {self.name}")
        caller = DebugUtils.get_caller_info()
        print(f"Called from: {caller}")
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if exc_type:
            print(f"Exception in {self.name}: {exc_type.__name__}: {exc_val}")
        print(f"Exiting debug context: {self.name}")

# Usage
def function_with_debug_context():
    with DebugContext("data_processing"):
        data = [1, 2, 3, 4, 5]
        result = sum(data)
        return result
```

## Best Practices for Debugging

### Debugging Checklist

```python
# Pre-debugging checklist
"""
1. Reproduce the issue consistently
2. Understand the expected vs actual behavior
3. Check logs first
4. Isolate the problem area
5. Use version control to identify when issue was introduced
6. Check dependencies and environment
"""

# Debugging strategies
class DebuggingStrategies:
    @staticmethod
    def binary_search_debugging():
        """
        Use binary search to find the exact line causing issues
        Comment out half the code, see if issue persists
        Continue narrowing down until you find the problematic code
        """
        pass

    @staticmethod
    def rubber_duck_debugging():
        """
        Explain the code line by line to someone (or rubber duck)
        Often reveals logical errors
        """
        pass

    @staticmethod
    def add_debug_prints():
        """
        Systematically add print statements to trace execution flow
        Remove after debugging is complete
        """
        pass

    @staticmethod
    def use_assertions():
        """
        Add assertions to verify assumptions
        Help catch issues early in development
        """
        pass

# Debugging anti-patterns to avoid
"""
1. Don't debug in production without proper safeguards
2. Don't leave debug code in production
3. Don't ignore warnings and deprecation messages
4. Don't debug without understanding the codebase
5. Don't make multiple changes while debugging
"""

# Clean debugging practices
def clean_debug_function():
    # Use meaningful variable names
    user_data = get_user_data()

    # Add type hints for clarity
    processed_data: dict = process_user_data(user_data)

    # Use early returns to reduce nesting
    if not processed_data:
        logger.warning("No processed data available")
        return None

    # Validate assumptions
    assert isinstance(processed_data, dict), "Expected dict type"

    # Clear error messages
    try:
        result = complex_calculation(processed_data)
    except ValueError as e:
        logger.error(f"Invalid data for calculation: {e}")
        raise
    except Exception as e:
        logger.exception("Unexpected error in complex_calculation")
        raise

    return result
```

This comprehensive debugging guide covers all essential techniques and tools for effective Python debugging, from basic pdb usage to advanced profiling and tracing techniques.