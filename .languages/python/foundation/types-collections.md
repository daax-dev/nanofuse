# Python Types and Collections Comprehensive Guide

## Built-in Data Types

Python has a rich type system with built-in types that form the foundation of all Python programs.

### Numeric Types

#### Integers (int)

```python
# Integer literals
decimal = 42
binary = 0b101010    # Binary (42 in decimal)
octal = 0o52         # Octal (42 in decimal)
hexadecimal = 0x2A   # Hexadecimal (42 in decimal)

print(f"All represent: {decimal == binary == octal == hexadecimal}")

# Large integers (unlimited precision)
large_number = 123456789012345678901234567890
print(f"Large number: {large_number}")
print(f"Type: {type(large_number)}")

# Integer operations
a, b = 10, 3
print(f"Addition: {a + b}")
print(f"Subtraction: {a - b}")
print(f"Multiplication: {a * b}")
print(f"Division (float): {a / b}")
print(f"Floor division: {a // b}")
print(f"Modulo: {a % b}")
print(f"Power: {a ** b}")

# Bitwise operations
x, y = 12, 10  # 1100, 1010 in binary
print(f"Bitwise AND: {x & y}")    # 8 (1000)
print(f"Bitwise OR: {x | y}")     # 14 (1110)
print(f"Bitwise XOR: {x ^ y}")    # 6 (0110)
print(f"Bitwise NOT: {~x}")       # -13
print(f"Left shift: {x << 2}")    # 48 (110000)
print(f"Right shift: {x >> 2}")   # 3 (11)
```

#### Floating Point Numbers (float)

```python
# Float literals
simple_float = 3.14
scientific = 1.23e-4  # 0.000123
infinity = float('inf')
negative_infinity = float('-inf')
not_a_number = float('nan')

print(f"Simple: {simple_float}")
print(f"Scientific: {scientific}")
print(f"Infinity: {infinity}")
print(f"NaN: {not_a_number}")

# Float precision and representation
import sys
print(f"Float info: {sys.float_info}")

# Common float operations and pitfalls
a = 0.1 + 0.2
print(f"0.1 + 0.2 = {a}")  # Not exactly 0.3!
print(f"Is equal to 0.3? {a == 0.3}")

# Better floating point comparison
from decimal import Decimal
import math

def almost_equal(a, b, tolerance=1e-9):
    return abs(a - b) < tolerance

print(f"Almost equal to 0.3? {almost_equal(a, 0.3)}")

# Using Decimal for precise calculations
decimal_a = Decimal('0.1') + Decimal('0.2')
print(f"Decimal result: {decimal_a}")
print(f"Is exactly 0.3? {decimal_a == Decimal('0.3')}")

# Math operations
import math
x = 16
print(f"Square root: {math.sqrt(x)}")
print(f"Logarithm: {math.log(x)}")
print(f"Sin: {math.sin(math.pi/2)}")
print(f"Ceiling: {math.ceil(3.7)}")
print(f"Floor: {math.floor(3.7)}")
```

#### Complex Numbers (complex)

```python
# Complex number creation
c1 = 3 + 4j
c2 = complex(3, 4)
c3 = complex('3+4j')

print(f"All equal: {c1 == c2 == c3}")

# Complex number operations
print(f"Real part: {c1.real}")
print(f"Imaginary part: {c1.imag}")
print(f"Conjugate: {c1.conjugate()}")
print(f"Magnitude: {abs(c1)}")

# Arithmetic with complex numbers
c4 = 2 + 3j
print(f"Addition: {c1 + c4}")
print(f"Multiplication: {c1 * c4}")
print(f"Division: {c1 / c4}")

# Euler's formula: e^(iπ) + 1 = 0
import cmath
euler = cmath.exp(1j * cmath.pi) + 1
print(f"Euler's identity: {euler}")
print(f"Close to zero: {abs(euler) < 1e-15}")
```

### Boolean Type (bool)

```python
# Boolean literals
true_value = True
false_value = False

# Boolean operations
print(f"AND: {True and False}")
print(f"OR: {True or False}")
print(f"NOT: {not True}")

# Truthiness in Python
def check_truthiness(value):
    return bool(value), "truthy" if value else "falsy"

# Falsy values
falsy_values = [False, None, 0, 0.0, 0j, "", [], {}, set()]
for value in falsy_values:
    truthiness, description = check_truthiness(value)
    print(f"{repr(value):>10} -> {description}")

# Everything else is truthy
truthy_values = [True, 1, -1, 0.1, "hello", [1], {"a": 1}, {1}]
for value in truthy_values:
    truthiness, description = check_truthiness(value)
    print(f"{repr(value):>10} -> {description}")

# Boolean in numeric context
print(f"True as int: {int(True)}")   # 1
print(f"False as int: {int(False)}")  # 0
print(f"True + True: {True + True}")  # 2
```

### None Type

```python
# None - Python's null value
nothing = None
print(f"None type: {type(None)}")

# None is a singleton
another_none = None
print(f"Same object: {nothing is another_none}")

# Common None patterns
def maybe_return_value(condition):
    if condition:
        return "value"
    return None  # Explicit None return

def default_parameter(value=None):
    if value is None:
        value = []  # Avoid mutable default argument
    return value

# Checking for None
value = maybe_return_value(False)
if value is None:
    print("No value returned")

# None in boolean context
if not value:  # None is falsy
    print("Value is None or other falsy value")

# Distinguishing None from other falsy values
if value is None:
    print("Value is specifically None")
```

## String Types (str)

### String Basics

```python
# String creation
single_quotes = 'Hello'
double_quotes = "World"
triple_quotes = """Multi-line
string with embedded 'quotes' and "quotes\""""

# Raw strings (no escape processing)
raw_string = r"C:\Users\Name\Documents"
regex_pattern = r"\d+\.\d+"

# Unicode strings (default in Python 3)
unicode_string = "Hello, 世界! 🌍"
print(f"Unicode: {unicode_string}")

# String methods
text = "  Hello, World!  "
print(f"Original: '{text}'")
print(f"Strip: '{text.strip()}'")
print(f"Lower: '{text.lower()}'")
print(f"Upper: '{text.upper()}'")
print(f"Title: '{text.title()}'")
print(f"Replace: '{text.replace('World', 'Python')}'")

# String searching
sentence = "The quick brown fox jumps over the lazy dog"
print(f"Find 'fox': {sentence.find('fox')}")
print(f"Count 'o': {sentence.count('o')}")
print(f"Starts with 'The': {sentence.startswith('The')}")
print(f"Ends with 'dog': {sentence.endswith('dog')}")

# String splitting and joining
words = sentence.split()
print(f"Words: {words}")
rejoined = " ".join(words)
print(f"Rejoined: {rejoined}")
```

### String Formatting

```python
# Old-style formatting (avoid)
name = "Alice"
age = 30
old_style = "Name: %s, Age: %d" % (name, age)

# str.format() method
format_method = "Name: {}, Age: {}".format(name, age)
format_named = "Name: {name}, Age: {age}".format(name=name, age=age)
format_indexed = "Name: {0}, Age: {1}".format(name, age)

# f-strings (preferred in Python 3.6+)
f_string = f"Name: {name}, Age: {age}"

# f-string expressions
numbers = [1, 2, 3, 4, 5]
f_expression = f"Sum: {sum(numbers)}, Average: {sum(numbers)/len(numbers):.2f}"

# Format specifications
pi = 3.14159265359
print(f"Pi: {pi:.2f}")           # 2 decimal places
print(f"Pi: {pi:.5f}")           # 5 decimal places
print(f"Pi: {pi:10.3f}")         # Width 10, 3 decimals

number = 42
print(f"Binary: {number:b}")      # Binary
print(f"Octal: {number:o}")       # Octal
print(f"Hex: {number:x}")         # Hexadecimal
print(f"Padded: {number:05d}")    # Zero-padded

# String alignment
text = "hello"
print(f"Left: '{text:<10}'")      # Left-aligned
print(f"Right: '{text:>10}'")     # Right-aligned
print(f"Center: '{text:^10}'")    # Centered
print(f"Fill: '{text:*^10}'")     # Centered with fill character

# Date formatting
from datetime import datetime
now = datetime.now()
print(f"Date: {now:%Y-%m-%d %H:%M:%S}")
print(f"ISO format: {now:%Y-%m-%dT%H:%M:%S}")
```

### String Operations and Performance

```python
import timeit

# String concatenation performance
def inefficient_concatenation(words):
    result = ""
    for word in words:
        result += word
    return result

def efficient_concatenation(words):
    return "".join(words)

words = ["word"] * 10000

# Time both methods
inefficient_time = timeit.timeit(
    lambda: inefficient_concatenation(words[:1000]),
    number=100
)
efficient_time = timeit.timeit(
    lambda: efficient_concatenation(words),
    number=100
)

print(f"Inefficient: {inefficient_time:.4f}s")
print(f"Efficient: {efficient_time:.4f}s")
print(f"Speedup: {inefficient_time/efficient_time:.2f}x")

# String membership testing
large_text = "word " * 100000
test_word = "word"

# Different membership tests
print(f"'word' in text: {'word' in large_text}")
print(f"text.find('word'): {large_text.find('word')}")
print(f"text.index('word'): {large_text.index('word')}")
```

## Collections

### Lists (list)

```python
# List creation and initialization
empty_list = []
numbers = [1, 2, 3, 4, 5]
mixed_types = [1, "hello", 3.14, True, None]
nested_lists = [[1, 2], [3, 4], [5, 6]]

# List comprehensions
squares = [x**2 for x in range(10)]
even_squares = [x**2 for x in range(10) if x % 2 == 0]
matrix = [[i*j for j in range(3)] for i in range(3)]

print(f"Squares: {squares}")
print(f"Even squares: {even_squares}")
print(f"Matrix: {matrix}")

# List operations
fruits = ["apple", "banana", "cherry"]

# Adding elements
fruits.append("date")                    # Add to end
fruits.insert(1, "blueberry")          # Insert at position
fruits.extend(["elderberry", "fig"])    # Add multiple

print(f"After additions: {fruits}")

# Removing elements
fruits.remove("banana")       # Remove first occurrence
popped = fruits.pop()         # Remove and return last
popped_index = fruits.pop(0)  # Remove and return at index
del fruits[1]                 # Delete by index

print(f"After removals: {fruits}")
print(f"Popped: {popped}, Popped at index: {popped_index}")

# List methods
numbers = [3, 1, 4, 1, 5, 9, 2, 6]
print(f"Original: {numbers}")
print(f"Count of 1: {numbers.count(1)}")
print(f"Index of 4: {numbers.index(4)}")

# Sorting
numbers_copy = numbers.copy()
numbers_copy.sort()                    # In-place sort
print(f"Sorted (in-place): {numbers_copy}")

sorted_numbers = sorted(numbers)       # Returns new list
print(f"Sorted (new list): {sorted_numbers}")

# Reverse sorting
numbers_copy.sort(reverse=True)
print(f"Reverse sorted: {numbers_copy}")

# Custom sorting
words = ["apple", "pie", "Washington", "book"]
sorted_by_length = sorted(words, key=len)
sorted_case_insensitive = sorted(words, key=str.lower)

print(f"By length: {sorted_by_length}")
print(f"Case insensitive: {sorted_case_insensitive}")

# List slicing
numbers = list(range(10))  # [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
print(f"Original: {numbers}")
print(f"First 3: {numbers[:3]}")
print(f"Last 3: {numbers[-3:]}")
print(f"Middle: {numbers[2:8]}")
print(f"Every 2nd: {numbers[::2]}")
print(f"Reversed: {numbers[::-1]}")

# Slice assignment
numbers[2:5] = [20, 30, 40]
print(f"After slice assignment: {numbers}")
```

### Tuples (tuple)

```python
# Tuple creation
empty_tuple = ()
single_element = (42,)  # Comma needed for single element
coordinates = (10, 20)
rgb_color = (255, 128, 0)

# Tuple unpacking
x, y = coordinates
r, g, b = rgb_color
print(f"Coordinates: x={x}, y={y}")
print(f"RGB: red={r}, green={g}, blue={b}")

# Multiple assignment using tuples
a, b, c = 1, 2, 3  # Same as (1, 2, 3)

# Swapping variables
a, b = b, a
print(f"After swap: a={a}, b={b}")

# Extended unpacking (Python 3+)
numbers = (1, 2, 3, 4, 5, 6)
first, *middle, last = numbers
print(f"First: {first}, Middle: {middle}, Last: {last}")

# Named tuples
from collections import namedtuple

Point = namedtuple('Point', ['x', 'y'])
Color = namedtuple('Color', 'r g b')  # String also works

point = Point(10, 20)
color = Color(255, 128, 0)

print(f"Point: {point}")
print(f"Point x: {point.x}, y: {point.y}")
print(f"Color: {color}")
print(f"Color as dict: {color._asdict()}")

# Named tuple methods
new_point = point._replace(x=30)
print(f"New point: {new_point}")

# Tuple methods (limited)
values = (1, 2, 2, 3, 2, 4)
print(f"Count of 2: {values.count(2)}")
print(f"Index of 3: {values.index(3)}")

# Tuple vs List performance
import sys
import timeit

list_data = [1, 2, 3, 4, 5]
tuple_data = (1, 2, 3, 4, 5)

print(f"List size: {sys.getsizeof(list_data)} bytes")
print(f"Tuple size: {sys.getsizeof(tuple_data)} bytes")

# Access time comparison
list_time = timeit.timeit('data[2]', setup='data=[1,2,3,4,5]', number=1000000)
tuple_time = timeit.timeit('data[2]', setup='data=(1,2,3,4,5)', number=1000000)

print(f"List access time: {list_time:.6f}s")
print(f"Tuple access time: {tuple_time:.6f}s")
```

### Dictionaries (dict)

```python
# Dictionary creation
empty_dict = {}
person = {"name": "Alice", "age": 30, "city": "New York"}
numbers = dict(one=1, two=2, three=3)
from_pairs = dict([("a", 1), ("b", 2), ("c", 3)])

# Dictionary comprehensions
squares_dict = {x: x**2 for x in range(5)}
filtered_dict = {k: v for k, v in person.items() if isinstance(v, str)}

print(f"Squares: {squares_dict}")
print(f"String values only: {filtered_dict}")

# Dictionary operations
person = {"name": "Alice", "age": 30}

# Adding/updating
person["city"] = "New York"      # Add new key
person["age"] = 31               # Update existing key
person.update({"job": "Engineer", "salary": 80000})

print(f"Updated person: {person}")

# Accessing values
print(f"Name: {person['name']}")                    # KeyError if missing
print(f"Age: {person.get('age')}")                  # None if missing
print(f"Country: {person.get('country', 'USA')}")   # Default value

# Safe access patterns
country = person.setdefault('country', 'USA')  # Add if missing
print(f"Country (setdefault): {country}")

# Dictionary methods
keys = person.keys()
values = person.values()
items = person.items()

print(f"Keys: {list(keys)}")
print(f"Values: {list(values)}")
print(f"Items: {list(items)}")

# Dictionary iteration
for key in person:
    print(f"Key: {key}")

for key, value in person.items():
    print(f"{key}: {value}")

# Removing items
del person["salary"]              # Delete specific key
job = person.pop("job")           # Remove and return value
last_item = person.popitem()      # Remove and return last item (Python 3.7+)

print(f"After removals: {person}")
print(f"Popped job: {job}")
print(f"Last item: {last_item}")

# Dictionary merging (Python 3.9+)
dict1 = {"a": 1, "b": 2}
dict2 = {"c": 3, "d": 4}
dict3 = {"a": 10, "e": 5}  # Note: 'a' will override

# Union operator
merged = dict1 | dict2
print(f"Merged: {merged}")

# Update operator
dict1_copy = dict1.copy()
dict1_copy |= dict3  # In-place update
print(f"Updated: {dict1_copy}")

# Nested dictionaries
nested = {
    "users": {
        "alice": {"age": 30, "city": "NY"},
        "bob": {"age": 25, "city": "LA"}
    },
    "settings": {
        "theme": "dark",
        "language": "en"
    }
}

print(f"Alice's age: {nested['users']['alice']['age']}")
print(f"Theme: {nested['settings']['theme']}")
```

### Sets (set)

```python
# Set creation
empty_set = set()  # Note: {} creates an empty dict
numbers_set = {1, 2, 3, 4, 5}
from_list = set([1, 2, 2, 3, 3, 4])  # Duplicates removed

print(f"From list with duplicates: {from_list}")

# Set comprehensions
even_squares = {x**2 for x in range(10) if x % 2 == 0}
print(f"Even squares: {even_squares}")

# Set operations
set1 = {1, 2, 3, 4, 5}
set2 = {4, 5, 6, 7, 8}

# Union (elements in either set)
union = set1 | set2
union_method = set1.union(set2)
print(f"Union: {union}")

# Intersection (elements in both sets)
intersection = set1 & set2
intersection_method = set1.intersection(set2)
print(f"Intersection: {intersection}")

# Difference (elements in set1 but not set2)
difference = set1 - set2
difference_method = set1.difference(set2)
print(f"Difference: {difference}")

# Symmetric difference (elements in either set but not both)
sym_diff = set1 ^ set2
sym_diff_method = set1.symmetric_difference(set2)
print(f"Symmetric difference: {sym_diff}")

# Set relationships
set_a = {1, 2, 3}
set_b = {1, 2, 3, 4, 5}
set_c = {4, 5, 6}

print(f"A subset of B: {set_a.issubset(set_b)}")
print(f"B superset of A: {set_b.issuperset(set_a)}")
print(f"A and C disjoint: {set_a.isdisjoint(set_c)}")

# Modifying sets
mutable_set = {1, 2, 3}
mutable_set.add(4)                    # Add single element
mutable_set.update([5, 6, 7])         # Add multiple elements
mutable_set.discard(2)                # Remove element (no error if missing)
mutable_set.remove(3)                 # Remove element (KeyError if missing)

print(f"Modified set: {mutable_set}")

# Frozen sets (immutable)
frozen = frozenset([1, 2, 3, 4])
# frozen.add(5)  # This would raise AttributeError

# Sets for membership testing
large_list = list(range(100000))
large_set = set(large_list)

import timeit

list_time = timeit.timeit('50000 in data', setup=f'data={large_list}', number=1000)
set_time = timeit.timeit('50000 in data', setup=f'data={large_set}', number=1000)

print(f"List membership time: {list_time:.6f}s")
print(f"Set membership time: {set_time:.6f}s")
print(f"Set speedup: {list_time/set_time:.1f}x")
```

## Advanced Collections (collections module)

### defaultdict

```python
from collections import defaultdict

# Regular dict with manual checking
regular_dict = {}
words = ["apple", "banana", "apple", "cherry", "banana", "banana"]

for word in words:
    if word in regular_dict:
        regular_dict[word] += 1
    else:
        regular_dict[word] = 1

print(f"Regular dict count: {regular_dict}")

# defaultdict with automatic default values
count_dict = defaultdict(int)  # Default value is 0
for word in words:
    count_dict[word] += 1

print(f"defaultdict count: {dict(count_dict)}")

# Different default factories
list_dict = defaultdict(list)
set_dict = defaultdict(set)
lambda_dict = defaultdict(lambda: "N/A")

# Group items by first letter
words = ["apple", "apricot", "banana", "blueberry", "cherry"]
for word in words:
    list_dict[word[0]].append(word)

print(f"Grouped by first letter: {dict(list_dict)}")

# Nested defaultdict
nested_default = defaultdict(lambda: defaultdict(int))
data = [("fruit", "apple", 5), ("fruit", "banana", 3), ("veggie", "carrot", 2)]

for category, item, count in data:
    nested_default[category][item] = count

print(f"Nested structure: {dict(nested_default)}")
```

### Counter

```python
from collections import Counter

# Counting elements
text = "hello world"
char_count = Counter(text)
print(f"Character count: {char_count}")

# Most common elements
words = ["apple", "banana", "apple", "cherry", "banana", "banana"]
word_count = Counter(words)
print(f"Word count: {word_count}")
print(f"Most common 2: {word_count.most_common(2)}")

# Counter arithmetic
counter1 = Counter(['a', 'b', 'c', 'a', 'b', 'b'])
counter2 = Counter(['a', 'b', 'b', 'c', 'c', 'c'])

print(f"Counter 1: {counter1}")
print(f"Counter 2: {counter2}")
print(f"Addition: {counter1 + counter2}")
print(f"Subtraction: {counter1 - counter2}")
print(f"Intersection: {counter1 & counter2}")
print(f"Union: {counter1 | counter2}")

# Updating counters
counter1.update(['d', 'e', 'd'])
print(f"After update: {counter1}")

# Counter methods
print(f"Total count: {sum(counter1.values())}")
print(f"Elements: {list(counter1.elements())}")

# Missing elements have count 0
print(f"Count of 'z': {counter1['z']}")
```

### deque (Double-ended queue)

```python
from collections import deque
import timeit

# Creating deques
empty_deque = deque()
from_iterable = deque([1, 2, 3, 4, 5])
max_length = deque([1, 2, 3], maxlen=3)  # Fixed size

print(f"From iterable: {from_iterable}")
print(f"Max length: {max_length}")

# Adding elements
d = deque([2, 3, 4])
d.appendleft(1)      # Add to left
d.append(5)          # Add to right
d.extendleft([0])    # Extend left (note: order reverses)
d.extend([6, 7])     # Extend right

print(f"After additions: {d}")

# Removing elements
left_element = d.popleft()   # Remove from left
right_element = d.pop()      # Remove from right

print(f"After removals: {d}")
print(f"Removed left: {left_element}, right: {right_element}")

# Rotation
d = deque([1, 2, 3, 4, 5])
print(f"Original: {d}")

d.rotate(2)    # Rotate right
print(f"Rotate right 2: {d}")

d.rotate(-2)   # Rotate left
print(f"Rotate left 2: {d}")

# Performance comparison: deque vs list for left operations
def list_left_operations(data, n):
    for _ in range(n):
        data.insert(0, 0)

def deque_left_operations(data, n):
    for _ in range(n):
        data.appendleft(0)

# Time comparisons
n = 10000
list_data = [1, 2, 3, 4, 5]
deque_data = deque([1, 2, 3, 4, 5])

list_time = timeit.timeit(
    lambda: list_left_operations(list_data.copy(), 1000),
    number=10
)
deque_time = timeit.timeit(
    lambda: deque_left_operations(deque(deque_data), 1000),
    number=10
)

print(f"List left operations: {list_time:.6f}s")
print(f"Deque left operations: {deque_time:.6f}s")
print(f"Deque speedup: {list_time/deque_time:.1f}x")

# Circular buffer using maxlen
circular_buffer = deque(maxlen=5)
for i in range(10):
    circular_buffer.append(i)
    print(f"Buffer after adding {i}: {circular_buffer}")
```

### OrderedDict

```python
from collections import OrderedDict

# Note: Regular dicts maintain insertion order in Python 3.7+
# OrderedDict still useful for explicit ordering guarantees

# Creating OrderedDict
ordered = OrderedDict([('first', 1), ('second', 2), ('third', 3)])
print(f"Ordered dict: {ordered}")

# Order-sensitive operations
ordered1 = OrderedDict([('a', 1), ('b', 2)])
ordered2 = OrderedDict([('b', 2), ('a', 1)])
regular1 = {'a': 1, 'b': 2}
regular2 = {'b': 2, 'a': 1}

print(f"OrderedDict equality: {ordered1 == ordered2}")  # False
print(f"Regular dict equality: {regular1 == regular2}")   # True

# Moving items to end
ordered = OrderedDict([('a', 1), ('b', 2), ('c', 3)])
ordered.move_to_end('a')  # Move to end
print(f"After move_to_end('a'): {ordered}")

ordered.move_to_end('c', last=False)  # Move to beginning
print(f"After move_to_end('c', last=False): {ordered}")

# LIFO (Last In, First Out) behavior
ordered.popitem(last=True)   # Remove from end
print(f"After LIFO pop: {ordered}")

# FIFO (First In, First Out) behavior
ordered.popitem(last=False)  # Remove from beginning
print(f"After FIFO pop: {ordered}")
```

### ChainMap

```python
from collections import ChainMap

# Combining multiple dictionaries
defaults = {'color': 'blue', 'size': 'medium'}
user_prefs = {'color': 'red'}
cli_args = {'size': 'large'}

# Chain maps (first dictionary has precedence)
combined = ChainMap(cli_args, user_prefs, defaults)
print(f"Combined: {dict(combined)}")
print(f"Color: {combined['color']}")  # From user_prefs
print(f"Size: {combined['size']}")    # From cli_args

# Modifying ChainMap
combined['theme'] = 'dark'  # Added to first map
print(f"After addition: {dict(combined)}")
print(f"First map: {combined.maps[0]}")

# New child (prepends new map)
child = combined.new_child({'font': 'arial'})
print(f"Child: {dict(child)}")

# Context management example
class ConfigManager:
    def __init__(self, defaults):
        self.defaults = defaults
        self.config = ChainMap({}, defaults)

    def load_config(self, config_dict):
        self.config = self.config.new_child(config_dict)

    def get(self, key):
        return self.config.get(key)

    def set(self, key, value):
        self.config[key] = value

config = ConfigManager({'debug': False, 'port': 8000})
config.load_config({'debug': True, 'host': 'localhost'})
config.set('version', '1.0')

print(f"Debug: {config.get('debug')}")
print(f"Port: {config.get('port')}")
print(f"Host: {config.get('host')}")
```

## Type Annotations and Hints

### Basic Type Annotations

```python
from typing import List, Dict, Set, Tuple, Optional, Union, Any

# Basic type annotations
def greet(name: str) -> str:
    return f"Hello, {name}!"

def add_numbers(a: int, b: int) -> int:
    return a + b

def divide(a: float, b: float) -> float:
    return a / b

# Variable annotations
age: int = 30
height: float = 5.9
is_student: bool = False
name: str = "Alice"

# Collection annotations
numbers: List[int] = [1, 2, 3, 4, 5]
scores: Dict[str, int] = {"Alice": 95, "Bob": 87}
unique_ids: Set[str] = {"id1", "id2", "id3"}
coordinates: Tuple[float, float] = (10.5, 20.7)

# Optional and Union types
def find_user(user_id: int) -> Optional[str]:
    users = {1: "Alice", 2: "Bob"}
    return users.get(user_id)

def process_id(id_value: Union[int, str]) -> str:
    return str(id_value)

# Any type (avoid when possible)
def process_data(data: Any) -> Any:
    return data

print(f"Greeting: {greet('World')}")
print(f"Addition: {add_numbers(5, 3)}")
print(f"User: {find_user(1)}")
```

### Advanced Type Annotations

```python
from typing import (
    List, Dict, Set, Tuple, Optional, Union, Any, Callable,
    TypeVar, Generic, Protocol, Literal, Final, ClassVar
)
from typing_extensions import NotRequired, Required  # Python 3.11+

# Generic types
T = TypeVar('T')

class Stack(Generic[T]):
    def __init__(self) -> None:
        self._items: List[T] = []

    def push(self, item: T) -> None:
        self._items.append(item)

    def pop(self) -> T:
        return self._items.pop()

    def peek(self) -> Optional[T]:
        return self._items[-1] if self._items else None

# Usage
int_stack = Stack[int]()
str_stack = Stack[str]()

# Callable types
def apply_function(func: Callable[[int, int], int], a: int, b: int) -> int:
    return func(a, b)

def multiply(x: int, y: int) -> int:
    return x * y

result = apply_function(multiply, 5, 3)

# Protocol (structural subtyping)
class Drawable(Protocol):
    def draw(self) -> None: ...

class Circle:
    def draw(self) -> None:
        print("Drawing a circle")

class Square:
    def draw(self) -> None:
        print("Drawing a square")

def render(shape: Drawable) -> None:
    shape.draw()

# Literal types
Color = Literal["red", "green", "blue"]

def set_color(color: Color) -> None:
    print(f"Setting color to {color}")

# Final and ClassVar
class Config:
    API_VERSION: Final[str] = "v1"
    instance_count: ClassVar[int] = 0

    def __init__(self, name: str) -> None:
        self.name: Final[str] = name
        Config.instance_count += 1

# TypedDict (structured dictionaries)
from typing_extensions import TypedDict

class PersonDict(TypedDict):
    name: str
    age: int
    email: NotRequired[str]  # Optional field

def process_person(person: PersonDict) -> str:
    return f"{person['name']} is {person['age']} years old"

person: PersonDict = {"name": "Alice", "age": 30}
```

### Type Checking and Validation

```python
import typing
from typing import get_type_hints, get_origin, get_args

def runtime_type_check(func):
    """Decorator for runtime type checking"""
    hints = get_type_hints(func)

    def wrapper(*args, **kwargs):
        # Check argument types
        for i, (arg_name, expected_type) in enumerate(hints.items()):
            if arg_name == 'return':
                continue

            if i < len(args):
                arg_value = args[i]
                if not isinstance(arg_value, expected_type):
                    raise TypeError(
                        f"Argument '{arg_name}' expected {expected_type}, "
                        f"got {type(arg_value)}"
                    )

        result = func(*args, **kwargs)

        # Check return type
        if 'return' in hints:
            expected_return = hints['return']
            if not isinstance(result, expected_return):
                raise TypeError(
                    f"Return value expected {expected_return}, "
                    f"got {type(result)}"
                )

        return result

    return wrapper

# Usage
@runtime_type_check
def add_strings(a: str, b: str) -> str:
    return a + b

# This works
result = add_strings("hello", "world")

# This would raise TypeError
# result = add_strings("hello", 123)

# Type introspection
def analyze_type_hints(func):
    hints = get_type_hints(func)
    for name, hint in hints.items():
        origin = get_origin(hint)
        args = get_args(hint)
        print(f"{name}: {hint} (origin: {origin}, args: {args})")

def example_func(data: List[Dict[str, int]]) -> Optional[str]:
    pass

analyze_type_hints(example_func)
```

## Performance Considerations

### Memory Usage Comparison

```python
import sys
import array
from collections import deque

def compare_memory_usage():
    # Compare memory usage of different collection types
    size = 10000

    # List
    int_list = list(range(size))

    # Tuple
    int_tuple = tuple(range(size))

    # Array
    int_array = array.array('i', range(size))

    # Deque
    int_deque = deque(range(size))

    # Set
    int_set = set(range(size))

    # Dict
    int_dict = {i: i for i in range(size)}

    print("Memory Usage Comparison (10,000 integers):")
    print(f"List:  {sys.getsizeof(int_list):>8} bytes")
    print(f"Tuple: {sys.getsizeof(int_tuple):>8} bytes")
    print(f"Array: {sys.getsizeof(int_array):>8} bytes")
    print(f"Deque: {sys.getsizeof(int_deque):>8} bytes")
    print(f"Set:   {sys.getsizeof(int_set):>8} bytes")
    print(f"Dict:  {sys.getsizeof(int_dict):>8} bytes")

compare_memory_usage()
```

### Performance Benchmarks

```python
import timeit
from collections import deque, defaultdict

def performance_comparison():
    # List vs Deque for different operations
    setup_code = """
from collections import deque
import random

data_list = list(range(10000))
data_deque = deque(range(10000))
    """

    # Append operations
    list_append = timeit.timeit(
        'data_list.append(1)',
        setup=setup_code,
        number=100000
    )

    deque_append = timeit.timeit(
        'data_deque.append(1)',
        setup=setup_code,
        number=100000
    )

    # Prepend operations
    list_prepend = timeit.timeit(
        'data_list.insert(0, 1)',
        setup=setup_code,
        number=10000
    )

    deque_prepend = timeit.timeit(
        'data_deque.appendleft(1)',
        setup=setup_code,
        number=100000
    )

    print("Performance Comparison:")
    print(f"List append:    {list_append:.6f}s")
    print(f"Deque append:   {deque_append:.6f}s")
    print(f"List prepend:   {list_prepend:.6f}s")
    print(f"Deque prepend:  {deque_prepend:.6f}s")
    print(f"Prepend speedup: {list_prepend/deque_prepend:.1f}x")

performance_comparison()
```

## Best Practices and Guidelines

### Choosing the Right Collection Type

```python
# Decision matrix for collection types
COLLECTION_GUIDE = {
    "Need ordered, mutable sequence": "list",
    "Need ordered, immutable sequence": "tuple",
    "Need unique elements, fast membership testing": "set",
    "Need key-value mapping": "dict",
    "Need fast insertion/deletion at both ends": "deque",
    "Need automatic counting": "Counter",
    "Need default values for missing keys": "defaultdict",
    "Need to maintain insertion order explicitly": "OrderedDict",
    "Need to chain multiple mappings": "ChainMap",
    "Need memory-efficient numeric arrays": "array.array",
    "Need named access to tuple elements": "namedtuple"
}

def recommend_collection(requirement):
    return COLLECTION_GUIDE.get(requirement, "Consider your specific needs")

# Example usage
for requirement, recommendation in COLLECTION_GUIDE.items():
    print(f"{requirement}: {recommendation}")
```

### Common Pitfalls and Solutions

```python
# Pitfall 1: Mutable default arguments
def bad_function(items=[]):  # DON'T DO THIS
    items.append("new item")
    return items

def good_function(items=None):  # DO THIS
    if items is None:
        items = []
    items.append("new item")
    return items

# Pitfall 2: Modifying list while iterating
numbers = [1, 2, 3, 4, 5]

# Bad - modifies list during iteration
# for i, num in enumerate(numbers):
#     if num % 2 == 0:
#         numbers.remove(num)  # Can skip elements or raise errors

# Good - iterate over copy or use list comprehension
numbers = [num for num in numbers if num % 2 != 0]

# Pitfall 3: Using mutable objects as dictionary keys
# This won't work:
# bad_dict = {[1, 2]: "value"}  # TypeError

# Use immutable types as keys:
good_dict = {(1, 2): "value"}

# Pitfall 4: Inefficient string concatenation
# Bad for large amounts of data
def inefficient_join(words):
    result = ""
    for word in words:
        result += word
    return result

# Good
def efficient_join(words):
    return "".join(words)

# Pitfall 5: Not using appropriate collection for the task
# Using list for membership testing on large data
large_list = list(range(100000))
large_set = set(large_list)

# Bad: O(n) complexity
# if 50000 in large_list:
#     pass

# Good: O(1) complexity
if 50000 in large_set:
    pass

print("Common pitfalls and solutions covered")
```

This comprehensive guide covers Python's type system and collections, providing the foundation for understanding how to effectively use Python's built-in and extended data structures for different programming scenarios.