# Python Keywords and Syntax Reference

## Reserved Keywords (Python 3.12)

Python has 35 reserved keywords that cannot be used as variable names, function names, or identifiers:

```python
import keyword
print(keyword.kwlist)
print(f"Total keywords: {len(keyword.kwlist)}")
```

### Complete Keyword List

```python
False      await      else       import     pass
None       break      except     in         raise
True       class      finally    is         return
and        continue   for        lambda     try
as         def        from       nonlocal   while
assert     del        global     not        with
async      elif       if         or         yield
```

## Keyword Categories and Usage

### 1. Logical and Comparison Keywords

#### `and`, `or`, `not`
```python
# Logical operators
result = True and False  # False
result = True or False   # True
result = not True        # False

# Short-circuit evaluation
x = 5
result = x > 0 and x < 10  # True
result = x < 0 or x > 100  # False
```

#### `is`, `in`
```python
# Identity comparison
a = [1, 2, 3]
b = a
c = [1, 2, 3]
print(a is b)      # True (same object)
print(a is c)      # False (different objects)
print(a == c)      # True (same content)

# Membership testing
print(2 in [1, 2, 3])        # True
print('hello' in 'hello world')  # True
print('x' not in 'hello')    # True
```

### 2. Conditional Keywords

#### `if`, `elif`, `else`
```python
# Basic conditional
age = 18
if age < 13:
    category = \"child\"
elif age < 20:
    category = \"teenager\"
elif age < 65:
    category = \"adult\"
else:
    category = \"senior\"

# Ternary operator (conditional expression)
status = \"adult\" if age >= 18 else \"minor\"

# Nested conditions
if age >= 18:
    if has_license:
        can_drive = True
    else:
        can_drive = False
else:
    can_drive = False
```

### 3. Loop Keywords

#### `for`, `while`
```python
# For loops
for i in range(5):
    print(i)

for item in ['a', 'b', 'c']:
    print(item)

for i, item in enumerate(['a', 'b', 'c']):
    print(f\"{i}: {item}\")

# Dictionary iteration
data = {'a': 1, 'b': 2}
for key in data:
    print(key)

for key, value in data.items():
    print(f\"{key}: {value}\")

# While loops
count = 0
while count < 5:
    print(count)
    count += 1

# Infinite loop with break
while True:
    user_input = input(\"Enter 'quit' to exit: \")
    if user_input == 'quit':
        break
    print(f\"You entered: {user_input}\")
```

#### `break`, `continue`
```python
# Break - exit loop completely
for i in range(10):
    if i == 5:
        break
    print(i)  # Prints 0, 1, 2, 3, 4

# Continue - skip current iteration
for i in range(5):
    if i == 2:
        continue
    print(i)  # Prints 0, 1, 3, 4

# Loop with else clause
for i in range(5):
    if i == 10:  # Never true
        break
    print(i)
else:
    print(\"Loop completed normally\")  # This executes
```

### 4. Function and Class Keywords

#### `def`, `return`
```python
# Basic function definition
def greet(name):
    return f\"Hello, {name}!\"

# Function with default parameters
def calculate(x, y=1, operation='add'):
    if operation == 'add':
        return x + y
    elif operation == 'multiply':
        return x * y
    else:
        return None

# Multiple return values
def get_name_parts(full_name):
    parts = full_name.split()
    return parts[0], parts[-1]  # first, last

first, last = get_name_parts(\"John Doe\")

# Generator function
def fibonacci(n):
    a, b = 0, 1
    for _ in range(n):
        yield a
        a, b = b, a + b

# Async function
async def fetch_data():
    await some_async_operation()
    return data
```

#### `class`
```python
# Basic class definition
class Person:
    def __init__(self, name, age):
        self.name = name
        self.age = age

    def greet(self):
        return f\"Hi, I'm {self.name}\"

    @property
    def is_adult(self):
        return self.age >= 18

    @staticmethod
    def species():
        return \"Homo sapiens\"

    @classmethod
    def from_string(cls, person_str):
        name, age = person_str.split(',')
        return cls(name, int(age))

# Inheritance
class Student(Person):
    def __init__(self, name, age, student_id):
        super().__init__(name, age)
        self.student_id = student_id

    def greet(self):  # Method override
        return f\"Hi, I'm {self.name}, student #{self.student_id}\"
```

#### `lambda`
```python
# Lambda expressions (anonymous functions)
square = lambda x: x ** 2
add = lambda x, y: x + y

# Common use with map, filter, sorted
numbers = [1, 2, 3, 4, 5]
squares = list(map(lambda x: x**2, numbers))
evens = list(filter(lambda x: x % 2 == 0, numbers))

# Sorting with custom key
students = [('Alice', 85), ('Bob', 90), ('Charlie', 78)]
by_grade = sorted(students, key=lambda x: x[1], reverse=True)
```

### 5. Exception Handling Keywords

#### `try`, `except`, `finally`, `raise`
```python
# Basic exception handling
try:
    result = 10 / 0
except ZeroDivisionError:
    print(\"Cannot divide by zero!\")
except Exception as e:
    print(f\"An error occurred: {e}\")
finally:
    print(\"This always runs\")

# Multiple exception types
try:
    value = int(input(\"Enter a number: \"))
    result = 10 / value
except (ValueError, ZeroDivisionError) as e:
    print(f\"Invalid input: {e}\")

# Raising exceptions
def validate_age(age):
    if age < 0:
        raise ValueError(\"Age cannot be negative\")
    if age > 150:
        raise ValueError(\"Age seems unrealistic\")
    return age

# Custom exceptions
class CustomError(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

try:
    raise CustomError(\"Something went wrong\")
except CustomError as e:
    print(f\"Custom error: {e.message}\")

# Exception chaining
try:
    # Some operation that fails
    raise ValueError(\"Original error\")
except ValueError as e:
    raise RuntimeError(\"New error context\") from e
```

### 6. Context Management Keywords

#### `with`, `as`
```python
# File handling with context manager
with open('file.txt', 'r') as file:
    content = file.read()
# File is automatically closed

# Multiple context managers
with open('input.txt', 'r') as infile, open('output.txt', 'w') as outfile:
    outfile.write(infile.read().upper())

# Custom context manager
from contextlib import contextmanager

@contextmanager
def timing_context():
    import time
    start = time.time()
    try:
        yield
    finally:
        end = time.time()
        print(f\"Elapsed: {end - start:.2f} seconds\")

with timing_context():
    # Some time-consuming operation
    import time
    time.sleep(1)

# Class-based context manager
class DatabaseConnection:
    def __enter__(self):
        print(\"Connecting to database\")
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        print(\"Closing database connection\")
        if exc_type is not None:
            print(f\"Exception occurred: {exc_val}\")
        return False  # Don't suppress exceptions

with DatabaseConnection() as db:
    print(\"Doing database operations\")
```

### 7. Module and Import Keywords

#### `import`, `from`, `as`
```python
# Basic imports
import math
import os.path

# From imports
from math import sqrt, pi
from collections import defaultdict, Counter

# Aliased imports
import numpy as np
from matplotlib import pyplot as plt
import pandas as pd

# Import all (generally discouraged)
from math import *

# Conditional imports
try:
    import ujson as json  # Faster JSON library
except ImportError:
    import json  # Fallback to standard library

# Relative imports (in packages)
from . import sibling_module
from ..parent_package import other_module
from .subpackage import another_module
```

### 8. Variable Scope Keywords

#### `global`, `nonlocal`
```python
# Global keyword
counter = 0

def increment():
    global counter
    counter += 1

increment()
print(counter)  # 1

# Nonlocal keyword
def outer():
    x = 10

    def inner():
        nonlocal x
        x += 1
        return x

    return inner

func = outer()
print(func())  # 11
print(func())  # 12

# Without nonlocal (creates local variable)
def outer_wrong():
    x = 10

    def inner_wrong():
        x = x + 1  # UnboundLocalError
        return x

    return inner_wrong
```

### 9. Deletion and Memory Keywords

#### `del`
```python
# Delete variables
x = 10
del x
# print(x)  # NameError: name 'x' is not defined

# Delete list elements
my_list = [1, 2, 3, 4, 5]
del my_list[0]     # Remove first element
del my_list[1:3]   # Remove slice
print(my_list)     # [2, 5]

# Delete dictionary items
my_dict = {'a': 1, 'b': 2, 'c': 3}
del my_dict['b']
print(my_dict)     # {'a': 1, 'c': 3}

# Delete attributes
class MyClass:
    def __init__(self):
        self.attr = \"value\"

obj = MyClass()
del obj.attr
# print(obj.attr)  # AttributeError
```

### 10. Control Flow Keywords

#### `pass`
```python
# Placeholder for empty code blocks
def todo_function():
    pass  # Will implement later

class EmptyClass:
    pass

# In conditionals
for i in range(10):
    if i % 2 == 0:
        pass  # Do nothing for even numbers
    else:
        print(f\"Odd: {i}\")

# Exception handling placeholder
try:
    risky_operation()
except SpecificError:
    pass  # Ignore this specific error
```

### 11. Literal Keywords

#### `True`, `False`, `None`
```python
# Boolean literals
is_active = True
is_complete = False

# None (null equivalent)
result = None
optional_value = None

def might_return_value(condition):
    if condition:
        return \"some value\"
    return None  # Explicit None return

# Checking for None
if result is None:
    print(\"No result\")

if result is not None:
    print(f\"Result: {result}\")

# Default parameter values
def greet(name=None):
    if name is None:
        name = \"World\"
    return f\"Hello, {name}!\"
```

### 12. Assertion Keyword

#### `assert`
```python
# Basic assertions
x = 10
assert x > 5  # Passes silently

try:
    assert x < 5  # Raises AssertionError
except AssertionError:
    print(\"Assertion failed\")

# Assertions with messages
def divide(a, b):
    assert b != 0, \"Division by zero not allowed\"
    return a / b

# Debugging assertions
def process_list(items):
    assert isinstance(items, list), f\"Expected list, got {type(items)}\"
    assert len(items) > 0, \"List cannot be empty\"

    # Process items...
    return processed_items

# Note: Assertions can be disabled with python -O
```

### 13. Async/Await Keywords (Python 3.5+)

#### `async`, `await`
```python
import asyncio

# Async function definition
async def fetch_data(url):
    # Simulate network request
    await asyncio.sleep(1)
    return f\"Data from {url}\"

# Async context manager
async def async_context_example():
    async with aiofiles.open('file.txt', 'r') as f:
        content = await f.read()
    return content

# Async iteration
async def async_generator():
    for i in range(5):
        await asyncio.sleep(0.1)
        yield i

async def consume_async_generator():
    async for item in async_generator():
        print(item)

# Running async code
async def main():
    result = await fetch_data(\"https://example.com\")
    print(result)

    await consume_async_generator()

# Event loop
asyncio.run(main())
```

### 14. Yield Keyword

#### `yield`
```python
# Basic generator
def count_up_to(max_count):
    count = 1
    while count <= max_count:
        yield count
        count += 1

# Using the generator
for num in count_up_to(5):
    print(num)  # Prints 1, 2, 3, 4, 5

# Yield expressions (receiving values)
def echo():
    while True:
        value = yield
        if value is not None:
            print(f\"Received: {value}\")

gen = echo()
next(gen)  # Prime the generator
gen.send(\"Hello\")  # Prints: Received: Hello

# Yield from (delegating to subgenerator)
def sub_generator():
    yield 1
    yield 2

def main_generator():
    yield from sub_generator()
    yield 3

list(main_generator())  # [1, 2, 3]
```

## Operators and Special Characters

### Arithmetic Operators
```python
# Basic arithmetic
a, b = 10, 3
print(a + b)   # Addition: 13
print(a - b)   # Subtraction: 7
print(a * b)   # Multiplication: 30
print(a / b)   # Division: 3.333...
print(a // b)  # Floor division: 3
print(a % b)   # Modulo: 1
print(a ** b)  # Exponentiation: 1000

# Augmented assignment
a += 1   # a = a + 1
a -= 1   # a = a - 1
a *= 2   # a = a * 2
a /= 2   # a = a / 2
a //= 2  # a = a // 2
a %= 3   # a = a % 3
a **= 2  # a = a ** 2
```

### Comparison Operators
```python
a, b = 5, 3
print(a == b)  # Equal: False
print(a != b)  # Not equal: True
print(a < b)   # Less than: False
print(a > b)   # Greater than: True
print(a <= b)  # Less than or equal: False
print(a >= b)  # Greater than or equal: True

# Chained comparisons
x = 5
print(1 < x < 10)     # True
print(x < 3 or x > 7) # False
```

### Bitwise Operators
```python
a, b = 12, 10  # 1100, 1010 in binary
print(a & b)   # AND: 8 (1000)
print(a | b)   # OR: 14 (1110)
print(a ^ b)   # XOR: 6 (0110)
print(~a)      # NOT: -13
print(a << 2)  # Left shift: 48
print(a >> 2)  # Right shift: 3

# Augmented bitwise assignment
a &= b   # a = a & b
a |= b   # a = a | b
a ^= b   # a = a ^ b
a <<= 2  # a = a << 2
a >>= 2  # a = a >> 2
```

## Syntax Constructs

### List Comprehensions
```python
# Basic list comprehension
squares = [x**2 for x in range(10)]

# With condition
evens = [x for x in range(20) if x % 2 == 0]

# Nested comprehension
matrix = [[i*j for j in range(3)] for i in range(3)]

# Multiple iterables
pairs = [(x, y) for x in range(3) for y in range(3)]
```

### Dictionary Comprehensions
```python
# Basic dict comprehension
squares_dict = {x: x**2 for x in range(5)}

# From two lists
keys = ['a', 'b', 'c']
values = [1, 2, 3]
combined = {k: v for k, v in zip(keys, values)}

# With condition
filtered = {k: v for k, v in data.items() if v > 10}
```

### Set Comprehensions
```python
# Basic set comprehension
unique_squares = {x**2 for x in range(-5, 6)}

# With condition
vowels = {char.lower() for char in \"Hello World\" if char.lower() in 'aeiou'}
```

### Generator Expressions
```python
# Generator expression
squares_gen = (x**2 for x in range(10))

# Memory efficient for large datasets
sum_of_squares = sum(x**2 for x in range(1000000))
```

### Pattern Matching (Python 3.10+)

#### `match`, `case`
```python
# Basic pattern matching
def handle_value(value):
    match value:
        case 0:
            return \"zero\"
        case 1 | 2 | 3:
            return \"small number\"
        case x if x > 100:
            return \"large number\"
        case _:
            return \"other\"

# Pattern matching with data structures
def process_data(data):
    match data:
        case []:
            return \"empty list\"
        case [x]:
            return f\"single item: {x}\"
        case [x, y]:
            return f\"two items: {x}, {y}\"
        case [x, *rest]:
            return f\"first: {x}, rest: {rest}\"

# Pattern matching with dictionaries
def handle_request(request):
    match request:
        case {\"action\": \"get\", \"resource\": resource}:
            return f\"Getting {resource}\"
        case {\"action\": \"post\", \"resource\": resource, \"data\": data}:
            return f\"Posting to {resource}: {data}\"
        case _:
            return \"Unknown request format\"

# Class pattern matching
class Point:
    def __init__(self, x, y):
        self.x, self.y = x, y

def describe_point(point):
    match point:
        case Point(x=0, y=0):
            return \"Origin\"
        case Point(x=0, y=y):
            return f\"On Y-axis at {y}\"
        case Point(x=x, y=0):
            return f\"On X-axis at {x}\"
        case Point(x=x, y=y):
            return f\"Point at ({x}, {y})\"
```

## String Formatting and Literals

### String Literals
```python
# Single and double quotes
single = 'Hello'
double = \"World\"
mixed = \"I'm learning Python\"

# Triple quotes for multiline
multiline = \"\"\"
This is a
multiline string
\"\"\"

# Raw strings (no escape processing)
raw = r\"C:\\Users\\Name\\file.txt\"
regex_pattern = r\"\\d+\\.\\d+\"

# Formatted strings (f-strings)
name = \"Alice\"
age = 30
greeting = f\"Hello, {name}! You are {age} years old.\"

# F-string with expressions
price = 19.99
tax_rate = 0.08
total = f\"Total: ${price * (1 + tax_rate):.2f}\"

# F-string with format specifiers
number = 42
binary = f\"Binary: {number:b}\"
hex_num = f\"Hex: {number:x}\"
percentage = f\"Percentage: {0.85:.1%}\"
```

### Bytes and Unicode
```python
# Bytes literal
byte_string = b\"Hello\"
byte_array = bytearray(b\"Hello\")

# Unicode strings (default in Python 3)
unicode_string = \"Hello, 世界\"
unicode_literal = \"\\u4e16\\u754c\"  # 世界

# String encoding/decoding
text = \"Hello, 世界\"
encoded = text.encode('utf-8')
decoded = encoded.decode('utf-8')
```

## Indentation and Code Structure

### Indentation Rules
```python
# Python uses indentation to define code blocks
# Standard is 4 spaces per indentation level

if True:
    print(\"This is indented\")
    if True:
        print(\"This is nested indentation\")
    print(\"Back to first level\")

# Function definitions
def my_function():
    \"\"\"Docstring here\"\"\"
    return \"Hello\"

# Class definitions
class MyClass:
    def __init__(self):
        self.value = 0

    def method(self):
        return self.value
```

### Line Continuation
```python
# Implicit line continuation with parentheses
result = (1 + 2 + 3 + 4 +
         5 + 6 + 7 + 8)

# Explicit line continuation with backslash
total = 1 + 2 + 3 + \\
        4 + 5 + 6

# Function calls across lines
result = some_function(argument1,
                      argument2,
                      argument3)

# Dictionary/list spanning lines
my_dict = {
    'key1': 'value1',
    'key2': 'value2',
    'key3': 'value3'
}
```

### Comments and Documentation
```python
# Single-line comment

\"\"\"
Multi-line comment
or docstring
\"\"\"

def function_with_docstring():
    \"\"\"
    This function demonstrates docstring format.

    Args:
        param1: Description of parameter

    Returns:
        Description of return value

    Raises:
        ValueError: When something goes wrong
    \"\"\"
    pass

# Inline comment
x = 5  # This is an inline comment
```

This comprehensive reference covers all Python keywords and essential syntax constructs, providing both explanation and practical examples for each element of the language.