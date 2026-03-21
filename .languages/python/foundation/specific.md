# Python-Specific Features and Unique Capabilities

Python has several distinctive features that set it apart from other programming languages. These features contribute to Python's philosophy of being readable, expressive, and powerful.

## Decorators

Decorators provide a clean way to modify or extend the behavior of functions and classes without permanently modifying their code.

### Function Decorators

```python
import time
from functools import wraps
import logging

# Basic decorator
def simple_timer(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        start_time = time.time()
        result = func(*args, **kwargs)
        end_time = time.time()
        print(f"{func.__name__} took {end_time - start_time:.4f} seconds")
        return result
    return wrapper

@simple_timer
def slow_function():
    time.sleep(1)
    return "Done"

# Decorator with parameters
def retry(max_attempts=3, delay=1):
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            for attempt in range(max_attempts):
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    if attempt == max_attempts - 1:
                        raise
                    print(f"Attempt {attempt + 1} failed: {e}")
                    time.sleep(delay)
        return wrapper
    return decorator

@retry(max_attempts=3, delay=0.5)
def unreliable_function():
    import random
    if random.random() < 0.7:
        raise Exception("Random failure")
    return "Success!"

# Multiple decorators (applied bottom-up)
def log_calls(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        print(f"Calling {func.__name__} with args={args}, kwargs={kwargs}")
        result = func(*args, **kwargs)
        print(f"{func.__name__} returned {result}")
        return result
    return wrapper

def validate_positive(func):
    @wraps(func)
    def wrapper(x):
        if x <= 0:
            raise ValueError("Argument must be positive")
        return func(x)
    return wrapper

@log_calls
@validate_positive
def square_root(x):
    return x ** 0.5

# result = square_root(16)  # Works
# result = square_root(-4)  # Raises ValueError

# Class-based decorators
class CountCalls:
    def __init__(self, func):
        self.func = func
        self.count = 0

    def __call__(self, *args, **kwargs):
        self.count += 1
        print(f"{self.func.__name__} has been called {self.count} times")
        return self.func(*args, **kwargs)

@CountCalls
def greet(name):
    return f"Hello, {name}!"

greet("Alice")
greet("Bob")

# Property decorators
class Circle:
    def __init__(self, radius):
        self._radius = radius

    @property
    def radius(self):
        return self._radius

    @radius.setter
    def radius(self, value):
        if value <= 0:
            raise ValueError("Radius must be positive")
        self._radius = value

    @property
    def area(self):
        return 3.14159 * self._radius ** 2

    @property
    def circumference(self):
        return 2 * 3.14159 * self._radius

circle = Circle(5)
print(f"Area: {circle.area}")
circle.radius = 10
print(f"New area: {circle.area}")
```

### Class Decorators

```python
from dataclasses import dataclass, field
from typing import List

# Dataclass decorator
@dataclass
class Person:
    name: str
    age: int
    hobbies: List[str] = field(default_factory=list)

    def __post_init__(self):
        if self.age < 0:
            raise ValueError("Age cannot be negative")

# Custom class decorator
def singleton(cls):
    instances = {}
    def get_instance(*args, **kwargs):
        if cls not in instances:
            instances[cls] = cls(*args, **kwargs)
        return instances[cls]
    return get_instance

@singleton
class Database:
    def __init__(self):
        print("Creating database connection")
        self.connection = "db_connection"

db1 = Database()
db2 = Database()
print(f"Same instance: {db1 is db2}")

# Class decorator for adding methods
def add_repr(cls):
    def __repr__(self):
        class_name = self.__class__.__name__
        attrs = ', '.join(f"{k}={v!r}" for k, v in self.__dict__.items())
        return f"{class_name}({attrs})"
    cls.__repr__ = __repr__
    return cls

@add_repr
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

point = Point(10, 20)
print(point)  # Point(x=10, y=20)
```

## Context Managers

Context managers provide a clean way to handle resource management and setup/teardown operations.

### Built-in Context Managers

```python
# File handling
with open('example.txt', 'w') as f:
    f.write("Hello, World!")
# File is automatically closed

# Multiple context managers
with open('input.txt', 'r') as infile, open('output.txt', 'w') as outfile:
    outfile.write(infile.read().upper())

# Thread locks
import threading
lock = threading.Lock()

with lock:
    # Critical section
    print("Thread-safe operation")
# Lock is automatically released

# Suppressing exceptions
from contextlib import suppress

with suppress(FileNotFoundError):
    with open('nonexistent.txt', 'r') as f:
        content = f.read()
# No exception raised if file doesn't exist
```

### Creating Custom Context Managers

```python
from contextlib import contextmanager
import time

# Function-based context manager
@contextmanager
def timer_context(name):
    start_time = time.time()
    print(f"Starting {name}")
    try:
        yield
    finally:
        end_time = time.time()
        print(f"{name} completed in {end_time - start_time:.4f} seconds")

with timer_context("database operation"):
    time.sleep(1)  # Simulate database work

# Class-based context manager
class DatabaseTransaction:
    def __init__(self, connection):
        self.connection = connection
        self.transaction = None

    def __enter__(self):
        print("Beginning transaction")
        self.transaction = self.connection.begin()
        return self.transaction

    def __exit__(self, exc_type, exc_val, exc_tb):
        if exc_type is None:
            print("Committing transaction")
            self.transaction.commit()
        else:
            print(f"Rolling back transaction due to {exc_type.__name__}")
            self.transaction.rollback()
        return False  # Don't suppress exceptions

# Mock connection for demonstration
class MockConnection:
    def begin(self):
        return MockTransaction()

class MockTransaction:
    def commit(self):
        print("Transaction committed")

    def rollback(self):
        print("Transaction rolled back")

connection = MockConnection()
with DatabaseTransaction(connection) as tx:
    # Database operations here
    pass

# Context manager for temporary changes
@contextmanager
def temporary_attribute(obj, attr, temp_value):
    original_value = getattr(obj, attr)
    setattr(obj, attr, temp_value)
    try:
        yield obj
    finally:
        setattr(obj, attr, original_value)

class Config:
    debug = False

config = Config()
print(f"Debug mode: {config.debug}")

with temporary_attribute(config, 'debug', True):
    print(f"Debug mode: {config.debug}")

print(f"Debug mode: {config.debug}")
```

## Generators and Iterators

Generators provide memory-efficient iteration and lazy evaluation.

### Generator Functions

```python
# Basic generator
def count_up_to(max_count):
    count = 1
    while count <= max_count:
        yield count
        count += 1

# Using the generator
for num in count_up_to(5):
    print(num)

# Generator expressions
squares = (x**2 for x in range(10))
print(f"Generator object: {squares}")
print(f"First few squares: {list(squares)[:5]}")

# Infinite generators
def fibonacci():
    a, b = 0, 1
    while True:
        yield a
        a, b = b, a + b

# Take first 10 Fibonacci numbers
fib_gen = fibonacci()
first_10_fib = [next(fib_gen) for _ in range(10)]
print(f"First 10 Fibonacci: {first_10_fib}")

# Generator with state
def stateful_generator():
    state = {"count": 0}

    def inner():
        while True:
            state["count"] += 1
            yield f"Item {state['count']}"

    return inner()

gen = stateful_generator()
print(next(gen))  # Item 1
print(next(gen))  # Item 2

# Generator delegation with yield from
def flatten(nested_list):
    for item in nested_list:
        if isinstance(item, list):
            yield from flatten(item)  # Recursive flattening
        else:
            yield item

nested = [1, [2, 3], [4, [5, 6]], 7]
flattened = list(flatten(nested))
print(f"Flattened: {flattened}")

# Generator pipeline
def read_lines(filename):
    with open(filename, 'r') as f:
        for line in f:
            yield line.strip()

def filter_comments(lines):
    for line in lines:
        if not line.startswith('#'):
            yield line

def process_line(line):
    return line.upper()

def process_file(filename):
    lines = read_lines(filename)
    filtered = filter_comments(lines)
    for line in filtered:
        yield process_line(line)

# Memory-efficient file processing
# for processed_line in process_file('config.txt'):
#     print(processed_line)
```

### Custom Iterators

```python
# Class-based iterator
class Countdown:
    def __init__(self, start):
        self.start = start

    def __iter__(self):
        return self

    def __next__(self):
        if self.start <= 0:
            raise StopIteration
        self.start -= 1
        return self.start + 1

countdown = Countdown(5)
for num in countdown:
    print(num)

# Iterator with __iter__ and __next__
class SquareNumbers:
    def __init__(self, max_value):
        self.max_value = max_value
        self.current = 0

    def __iter__(self):
        return self

    def __next__(self):
        if self.current >= self.max_value:
            raise StopIteration
        self.current += 1
        return self.current ** 2

squares = SquareNumbers(5)
print(f"Squares: {list(squares)}")

# Iterator tools
import itertools

# Infinite iterators
counter = itertools.count(10, 2)  # Start at 10, step by 2
first_5 = [next(counter) for _ in range(5)]
print(f"Count by 2s: {first_5}")

# Cycling through values
colors = itertools.cycle(['red', 'green', 'blue'])
first_10_colors = [next(colors) for _ in range(10)]
print(f"Cycling colors: {first_10_colors}")

# Chaining iterators
list1 = [1, 2, 3]
list2 = [4, 5, 6]
chained = itertools.chain(list1, list2)
print(f"Chained: {list(chained)}")

# Grouping
data = [('a', 1), ('a', 2), ('b', 3), ('b', 4), ('c', 5)]
grouped = itertools.groupby(data, key=lambda x: x[0])
for key, group in grouped:
    print(f"{key}: {list(group)}")
```

## Comprehensions

Python's comprehensions provide concise ways to create collections.

### List Comprehensions

```python
# Basic list comprehension
numbers = [1, 2, 3, 4, 5]
squares = [x**2 for x in numbers]
print(f"Squares: {squares}")

# With condition
even_squares = [x**2 for x in numbers if x % 2 == 0]
print(f"Even squares: {even_squares}")

# Nested loops
matrix = [[i*j for j in range(1, 4)] for i in range(1, 4)]
print(f"Multiplication table: {matrix}")

# Flattening with comprehension
matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
flattened = [item for row in matrix for item in row]
print(f"Flattened: {flattened}")

# Complex transformations
words = ["hello", "world", "python", "programming"]
title_words = [word.title() for word in words if len(word) > 5]
print(f"Long title words: {title_words}")

# Using functions in comprehensions
def is_prime(n):
    if n < 2:
        return False
    for i in range(2, int(n**0.5) + 1):
        if n % i == 0:
            return False
    return True

primes = [x for x in range(2, 50) if is_prime(x)]
print(f"Primes under 50: {primes}")
```

### Dictionary Comprehensions

```python
# Basic dictionary comprehension
words = ["apple", "banana", "cherry"]
word_lengths = {word: len(word) for word in words}
print(f"Word lengths: {word_lengths}")

# Transforming existing dictionary
original = {"a": 1, "b": 2, "c": 3}
squared = {k: v**2 for k, v in original.items()}
print(f"Squared values: {squared}")

# Filtering dictionary
filtered = {k: v for k, v in original.items() if v > 1}
print(f"Filtered: {filtered}")

# Swapping keys and values
swapped = {v: k for k, v in original.items()}
print(f"Swapped: {swapped}")

# Complex dictionary comprehension
text = "hello world"
char_positions = {char: [i for i, c in enumerate(text) if c == char]
                  for char in set(text) if char != ' '}
print(f"Character positions: {char_positions}")
```

### Set Comprehensions

```python
# Basic set comprehension
numbers = [1, 2, 2, 3, 3, 4, 5]
unique_squares = {x**2 for x in numbers}
print(f"Unique squares: {unique_squares}")

# Set comprehension with condition
text = "Hello World Programming"
vowels = {char.lower() for char in text if char.lower() in 'aeiou'}
print(f"Unique vowels: {vowels}")

# Set operations with comprehensions
set1 = {x for x in range(10) if x % 2 == 0}
set2 = {x for x in range(5, 15) if x % 3 == 0}
print(f"Even numbers 0-9: {set1}")
print(f"Multiples of 3, 5-14: {set2}")
print(f"Intersection: {set1 & set2}")
```

## Metaclasses

Metaclasses are classes whose instances are classes themselves.

### Understanding Metaclasses

```python
# Everything is an object in Python
class MyClass:
    pass

obj = MyClass()
print(f"obj type: {type(obj)}")           # <class '__main__.MyClass'>
print(f"MyClass type: {type(MyClass)}")   # <class 'type'>
print(f"type type: {type(type)}")         # <class 'type'>

# Creating classes dynamically
def init_method(self, value):
    self.value = value

def str_method(self):
    return f"DynamicClass(value={self.value})"

# Create class using type()
DynamicClass = type('DynamicClass', (), {
    '__init__': init_method,
    '__str__': str_method
})

dynamic_obj = DynamicClass(42)
print(dynamic_obj)

# Custom metaclass
class SingletonMeta(type):
    _instances = {}

    def __call__(cls, *args, **kwargs):
        if cls not in cls._instances:
            cls._instances[cls] = super().__call__(*args, **kwargs)
        return cls._instances[cls]

class Singleton(metaclass=SingletonMeta):
    def __init__(self, value):
        self.value = value

# All instances are the same object
s1 = Singleton(1)
s2 = Singleton(2)
print(f"Same instance: {s1 is s2}")
print(f"Value: {s1.value}")  # Still 1, not 2

# Metaclass for validation
class ValidatedMeta(type):
    def __new__(mcs, name, bases, attrs):
        # Validate that all methods have docstrings
        for key, value in attrs.items():
            if callable(value) and not key.startswith('_'):
                if not hasattr(value, '__doc__') or not value.__doc__:
                    raise ValueError(f"Method {key} must have a docstring")
        return super().__new__(mcs, name, bases, attrs)

class ValidatedClass(metaclass=ValidatedMeta):
    def process(self):
        """Process the data."""
        return "processed"

    # This would raise ValueError:
    # def invalid_method(self):
    #     return "no docstring"

# Metaclass for automatic registration
class RegisteredMeta(type):
    registry = {}

    def __new__(mcs, name, bases, attrs):
        cls = super().__new__(mcs, name, bases, attrs)
        if name != 'RegisteredBase':  # Don't register base class
            mcs.registry[name] = cls
        return cls

class RegisteredBase(metaclass=RegisteredMeta):
    pass

class Plugin1(RegisteredBase):
    pass

class Plugin2(RegisteredBase):
    pass

print(f"Registered classes: {RegisteredMeta.registry}")
```

## Descriptors

Descriptors define how attribute access is handled for class attributes.

### Descriptor Protocol

```python
# Basic descriptor
class LoggedAttribute:
    def __init__(self, name=None):
        self.name = name

    def __set_name__(self, owner, name):
        self.name = name

    def __get__(self, obj, objtype=None):
        if obj is None:
            return self
        value = obj.__dict__.get(self.name)
        print(f"Getting {self.name} = {value}")
        return value

    def __set__(self, obj, value):
        print(f"Setting {self.name} = {value}")
        obj.__dict__[self.name] = value

    def __delete__(self, obj):
        print(f"Deleting {self.name}")
        del obj.__dict__[self.name]

class MyClass:
    attr = LoggedAttribute()

    def __init__(self, value):
        self.attr = value

obj = MyClass(42)
print(obj.attr)  # Logs getting
obj.attr = 100   # Logs setting

# Validation descriptor
class PositiveNumber:
    def __init__(self, name=None):
        self.name = name

    def __set_name__(self, owner, name):
        self.name = name
        self.private_name = f'_{name}'

    def __get__(self, obj, objtype=None):
        if obj is None:
            return self
        return getattr(obj, self.private_name)

    def __set__(self, obj, value):
        if not isinstance(value, (int, float)):
            raise TypeError(f"{self.name} must be a number")
        if value <= 0:
            raise ValueError(f"{self.name} must be positive")
        setattr(obj, self.private_name, value)

class Circle:
    radius = PositiveNumber()

    def __init__(self, radius):
        self.radius = radius

    @property
    def area(self):
        return 3.14159 * self.radius ** 2

circle = Circle(5)
print(f"Area: {circle.area}")
# circle.radius = -1  # Would raise ValueError

# Cached property descriptor
class CachedProperty:
    def __init__(self, func):
        self.func = func
        self.attrname = None
        self.__doc__ = func.__doc__

    def __set_name__(self, owner, name):
        if self.attrname is None:
            self.attrname = name
        elif name != self.attrname:
            raise RuntimeError(
                f"Cannot assign the same cached_property to two different names "
                f"({self.attrname!r} and {name!r})."
            )

    def __get__(self, instance, owner=None):
        if instance is None:
            return self
        if self.attrname is None:
            raise TypeError(
                "Cannot use cached_property instance without calling __set_name__"
            )
        try:
            cache = instance.__dict__
        except AttributeError:
            msg = (
                f"No '__dict__' attribute on {type(instance).__name__!r} "
                f"instance to cache {self.attrname!r} property."
            )
            raise TypeError(msg) from None
        val = cache.get(self.attrname, None)
        if val is None:
            print(f"Computing {self.attrname}")
            val = self.func(instance)
            cache[self.attrname] = val
        return val

class ExpensiveComputation:
    @CachedProperty
    def expensive_result(self):
        """Simulate expensive computation."""
        import time
        time.sleep(1)  # Simulate work
        return sum(range(1000000))

comp = ExpensiveComputation()
print(comp.expensive_result)  # Takes time, computed once
print(comp.expensive_result)  # Returns cached value immediately
```

## Magic Methods (Dunder Methods)

Magic methods enable operator overloading and special behavior.

### Arithmetic and Comparison

```python
class Vector:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def __repr__(self):
        return f"Vector({self.x}, {self.y})"

    def __str__(self):
        return f"({self.x}, {self.y})"

    # Arithmetic operations
    def __add__(self, other):
        if isinstance(other, Vector):
            return Vector(self.x + other.x, self.y + other.y)
        return NotImplemented

    def __sub__(self, other):
        if isinstance(other, Vector):
            return Vector(self.x - other.x, self.y - other.y)
        return NotImplemented

    def __mul__(self, scalar):
        if isinstance(scalar, (int, float)):
            return Vector(self.x * scalar, self.y * scalar)
        return NotImplemented

    def __rmul__(self, scalar):
        return self.__mul__(scalar)

    def __truediv__(self, scalar):
        if isinstance(scalar, (int, float)) and scalar != 0:
            return Vector(self.x / scalar, self.y / scalar)
        return NotImplemented

    # Comparison operations
    def __eq__(self, other):
        if isinstance(other, Vector):
            return self.x == other.x and self.y == other.y
        return False

    def __lt__(self, other):
        if isinstance(other, Vector):
            return self.magnitude() < other.magnitude()
        return NotImplemented

    def __le__(self, other):
        return self < other or self == other

    # Magnitude method
    def magnitude(self):
        return (self.x**2 + self.y**2)**0.5

    # Boolean context
    def __bool__(self):
        return self.magnitude() != 0

# Using the Vector class
v1 = Vector(3, 4)
v2 = Vector(1, 2)

print(f"v1: {v1}")
print(f"v2: {v2}")
print(f"v1 + v2: {v1 + v2}")
print(f"v1 - v2: {v1 - v2}")
print(f"v1 * 2: {v1 * 2}")
print(f"3 * v1: {3 * v1}")
print(f"v1 / 2: {v1 / 2}")
print(f"v1 == v2: {v1 == v2}")
print(f"v1 > v2: {v1 > v2}")
print(f"bool(v1): {bool(v1)}")
print(f"bool(Vector(0, 0)): {bool(Vector(0, 0))}")
```

### Container Methods

```python
class CustomList:
    def __init__(self, items=None):
        self._items = list(items) if items else []

    def __repr__(self):
        return f"CustomList({self._items})"

    # Container methods
    def __len__(self):
        return len(self._items)

    def __getitem__(self, index):
        return self._items[index]

    def __setitem__(self, index, value):
        self._items[index] = value

    def __delitem__(self, index):
        del self._items[index]

    def __contains__(self, item):
        return item in self._items

    def __iter__(self):
        return iter(self._items)

    def __reversed__(self):
        return reversed(self._items)

    # Append method
    def append(self, item):
        self._items.append(item)

# Using CustomList
custom_list = CustomList([1, 2, 3, 4, 5])
print(f"Length: {len(custom_list)}")
print(f"Item at index 2: {custom_list[2]}")
print(f"Contains 3: {3 in custom_list}")

custom_list[1] = 20
print(f"After modification: {custom_list}")

for item in custom_list:
    print(f"Item: {item}")

for item in reversed(custom_list):
    print(f"Reversed item: {item}")
```

### Context Manager Methods

```python
class ManagedResource:
    def __init__(self, name):
        self.name = name
        self.resource = None

    def __enter__(self):
        print(f"Acquiring {self.name}")
        self.resource = f"resource_{self.name}"
        return self.resource

    def __exit__(self, exc_type, exc_val, exc_tb):
        print(f"Releasing {self.name}")
        if exc_type is not None:
            print(f"Exception occurred: {exc_type.__name__}: {exc_val}")
        self.resource = None
        return False  # Don't suppress exceptions

with ManagedResource("database") as db:
    print(f"Using {db}")
    # Resource is automatically released
```

## Advanced Python Features

### Monkey Patching

```python
# Adding methods to existing classes
class Calculator:
    def add(self, a, b):
        return a + b

calc = Calculator()
print(f"Add: {calc.add(5, 3)}")

# Add a new method
def multiply(self, a, b):
    return a * b

Calculator.multiply = multiply
print(f"Multiply: {calc.multiply(5, 3)}")

# Patching built-in classes (use with caution)
def is_palindrome(self):
    return str(self) == str(self)[::-1]

# Don't actually do this in production code!
# int.is_palindrome = is_palindrome
# print(f"121 is palindrome: {(121).is_palindrome()}")
```

### Dynamic Attribute Access

```python
class DynamicClass:
    def __init__(self):
        self.data = {}

    def __getattr__(self, name):
        if name in self.data:
            return self.data[name]
        raise AttributeError(f"'{type(self).__name__}' object has no attribute '{name}'")

    def __setattr__(self, name, value):
        if name == 'data':
            super().__setattr__(name, value)
        else:
            self.data[name] = value

    def __delattr__(self, name):
        if name in self.data:
            del self.data[name]
        else:
            raise AttributeError(f"'{type(self).__name__}' object has no attribute '{name}'")

obj = DynamicClass()
obj.x = 10
obj.y = 20
print(f"x: {obj.x}, y: {obj.y}")
print(f"Data: {obj.data}")

# Using getattr, setattr, hasattr, delattr
print(f"Has attribute 'x': {hasattr(obj, 'x')}")
print(f"Get attribute 'x': {getattr(obj, 'x')}")
setattr(obj, 'z', 30)
print(f"After setting z: {obj.data}")
delattr(obj, 'z')
print(f"After deleting z: {obj.data}")
```

### Slots for Memory Optimization

```python
import sys

# Regular class
class RegularPoint:
    def __init__(self, x, y):
        self.x = x
        self.y = y

# Class with __slots__
class SlottedPoint:
    __slots__ = ['x', 'y']

    def __init__(self, x, y):
        self.x = x
        self.y = y

regular = RegularPoint(1, 2)
slotted = SlottedPoint(1, 2)

print(f"Regular point size: {sys.getsizeof(regular)} bytes")
print(f"Slotted point size: {sys.getsizeof(slotted)} bytes")

# Regular class has __dict__
print(f"Regular point __dict__: {regular.__dict__}")

# Slotted class doesn't have __dict__
try:
    print(slotted.__dict__)
except AttributeError:
    print("Slotted point has no __dict__")

# Can't add new attributes to slotted class
try:
    slotted.z = 3
except AttributeError as e:
    print(f"Cannot add attribute: {e}")
```

### Abstract Base Classes

```python
from abc import ABC, abstractmethod

class Shape(ABC):
    @abstractmethod
    def area(self):
        pass

    @abstractmethod
    def perimeter(self):
        pass

    # Concrete method
    def description(self):
        return f"This is a {self.__class__.__name__} with area {self.area()}"

class Rectangle(Shape):
    def __init__(self, width, height):
        self.width = width
        self.height = height

    def area(self):
        return self.width * self.height

    def perimeter(self):
        return 2 * (self.width + self.height)

class Circle(Shape):
    def __init__(self, radius):
        self.radius = radius

    def area(self):
        return 3.14159 * self.radius ** 2

    def perimeter(self):
        return 2 * 3.14159 * self.radius

# Cannot instantiate abstract class
# shape = Shape()  # TypeError

rectangle = Rectangle(10, 5)
circle = Circle(3)

print(f"Rectangle: {rectangle.description()}")
print(f"Circle: {circle.description()}")

# Check if class implements interface
print(f"Rectangle is Shape: {isinstance(rectangle, Shape)}")
print(f"Circle is Shape: {isinstance(circle, Shape)}")
```

These Python-specific features make the language uniquely expressive and powerful, enabling developers to write more readable, maintainable, and efficient code. Understanding and leveraging these features is key to mastering Python development.