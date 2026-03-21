# Python Memory Management Deep Dive

## Memory Model Overview

Python uses automatic memory management through reference counting and a cyclic garbage collector. Understanding how Python manages memory is crucial for writing efficient applications.

### Memory Layout

```
Python Memory Layout:
┌─────────────────────────┐
│     Code Objects        │ ← Bytecode, constants
├─────────────────────────┤
│     Global Variables    │ ← Module-level objects
├─────────────────────────┤
│     Call Stack          │ ← Function calls, local vars
├─────────────────────────┤
│     Heap                │ ← Dynamic objects
│  ┌─────────────────────┐│
│  │ Small Objects Pool  ││ ← Objects < 512 bytes
│  ├─────────────────────┤│
│  │ Large Objects       ││ ← Objects >= 512 bytes
│  └─────────────────────┘│
└─────────────────────────┘
```

## Reference Counting

### Basic Reference Counting

```python
import sys

# Check reference count
x = [1, 2, 3]
print(f"Reference count: {sys.getrefcount(x)}")  # Usually 2 (x + getrefcount param)

# Creating references increases count
y = x
print(f"After y = x: {sys.getrefcount(x)}")  # Count increases

# Deleting references decreases count
del y
print(f"After del y: {sys.getrefcount(x)}")  # Count decreases

# Function parameters create temporary references
def use_list(lst):
    print(f"Inside function: {sys.getrefcount(lst)}")

use_list(x)  # Reference count temporarily increases
print(f"After function: {sys.getrefcount(x)}")  # Back to original
```

### Reference Count Examples

```python
import sys

def analyze_references():
    # Simple object
    obj = {"key": "value"}
    print(f"Initial: {sys.getrefcount(obj)}")
    
    # Container references
    container = [obj, obj]  # Two references in list
    print(f"In container: {sys.getrefcount(obj)}")
    
    # Dictionary references
    mapping = {"a": obj, "b": obj}
    print(f"In mapping: {sys.getrefcount(obj)}")
    
    # Function closure reference
    def closure():
        return obj
    print(f"With closure: {sys.getrefcount(obj)}")
    
    # Class attribute reference
    class Container:
        attr = obj
    print(f"As class attr: {sys.getrefcount(obj)}")

analyze_references()
```

### Circular References

```python
import sys
import gc

# Circular reference example
class Node:
    def __init__(self, value):
        self.value = value
        self.parent = None
        self.children = []
    
    def add_child(self, child):
        child.parent = self  # Creates circular reference
        self.children.append(child)

# Create circular reference
root = Node("root")
child = Node("child")
root.add_child(child)

print(f"Root refs: {sys.getrefcount(root)}")
print(f"Child refs: {sys.getrefcount(child)}")

# Breaking circular references
def break_cycles(node):
    """Manually break circular references"""
    for child in node.children:
        child.parent = None
        break_cycles(child)
    node.children.clear()

# Weak references to avoid cycles
import weakref

class BetterNode:
    def __init__(self, value):
        self.value = value
        self._parent = None  # Will store weak reference
        self.children = []
    
    @property
    def parent(self):
        return self._parent() if self._parent else None
    
    @parent.setter
    def parent(self, value):
        self._parent = weakref.ref(value) if value else None
    
    def add_child(self, child):
        child.parent = self  # Now uses weak reference
        self.children.append(child)
```

## Garbage Collection

### Garbage Collection Mechanics

```python
import gc
import weakref

# Check garbage collection status
print(f"GC enabled: {gc.isenabled()}")
print(f"GC thresholds: {gc.get_threshold()}")
print(f"GC counts: {gc.get_count()}")

# Manual garbage collection
collected = gc.collect()
print(f"Objects collected: {collected}")

# Garbage collection with generations
def analyze_gc_generations():
    # Create objects in different generations
    temp_objects = []
    
    # Generation 0 objects (new)
    for i in range(1000):
        temp_objects.append([i] * 100)
    
    print(f"Before collection: {gc.get_count()}")
    
    # Force collection
    gc.collect()
    print(f"After collection: {gc.get_count()}")
    
    # Keep some objects alive to move to generation 1
    persistent = temp_objects[:100]
    del temp_objects
    
    # Create more objects
    for i in range(1000):
        temp = [i] * 50
    
    print(f"After more allocations: {gc.get_count()}")
    gc.collect()
    print(f"After second collection: {gc.get_count()}")

analyze_gc_generations()
```

### Debugging Garbage Collection

```python
import gc
import sys

# Enable garbage collection debugging
gc.set_debug(gc.DEBUG_STATS | gc.DEBUG_LEAK)

# Track specific object types
def track_object_creation():
    # Get initial counts
    initial_objects = len(gc.get_objects())
    
    # Create some objects
    data = []
    for i in range(1000):
        data.append({"id": i, "data": list(range(i % 10))})
    
    current_objects = len(gc.get_objects())
    print(f"Objects created: {current_objects - initial_objects}")
    
    # Clear references
    data.clear()
    
    # Force garbage collection
    collected = gc.collect()
    print(f"Objects collected: {collected}")
    
    final_objects = len(gc.get_objects())
    print(f"Objects remaining: {final_objects - initial_objects}")

# Find uncollectable objects
def find_uncollectable():
    gc.collect()  # Clear any existing garbage
    
    # Create circular reference
    class Circular:
        def __init__(self):
            self.ref = self
    
    obj = Circular()
    del obj
    
    collected = gc.collect()
    print(f"Collected: {collected}")
    print(f"Uncollectable: {len(gc.garbage)}")
    
    # Examine garbage
    for item in gc.garbage:
        print(f"Uncollectable object: {type(item)}")

track_object_creation()
find_uncollectable()
```

## Memory Optimization Techniques

### Object Internment and Caching

```python
# Small integer caching
a = 256
b = 256
print(f"a is b: {a is b}")  # True - cached

a = 257
b = 257
print(f"a is b: {a is b}")  # Usually False - not cached

# String internment
s1 = "hello"
s2 = "hello"
print(f"s1 is s2: {s1 is s2}")  # True - interned

# Manual string internment
import sys
s3 = sys.intern("dynamic_string")
s4 = sys.intern("dynamic_string")
print(f"s3 is s4: {s3 is s4}")  # True

# None, True, False are singletons
a = None
b = None
print(f"a is b: {a is b}")  # Always True
```

### __slots__ for Memory Efficiency

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

# Memory comparison
regular = RegularPoint(1, 2)
slotted = SlottedPoint(1, 2)

print(f"Regular class size: {sys.getsizeof(regular)} bytes")
print(f"Regular __dict__ size: {sys.getsizeof(regular.__dict__)} bytes")
print(f"Slotted class size: {sys.getsizeof(slotted)} bytes")

# Slots with inheritance
class Point3D(SlottedPoint):
    __slots__ = ['z']
    
    def __init__(self, x, y, z):
        super().__init__(x, y)
        self.z = z

# Memory usage analysis
def analyze_memory_usage():
    import tracemalloc
    
    tracemalloc.start()
    
    # Create many regular instances
    regular_points = [RegularPoint(i, i) for i in range(10000)]
    
    current, peak = tracemalloc.get_traced_memory()
    print(f"Regular points - Current: {current / 1024 / 1024:.1f} MB")
    
    tracemalloc.reset_peak()
    
    # Create many slotted instances
    slotted_points = [SlottedPoint(i, i) for i in range(10000)]
    
    current, peak = tracemalloc.get_traced_memory()
    print(f"Slotted points - Current: {current / 1024 / 1024:.1f} MB")
    
    tracemalloc.stop()

analyze_memory_usage()
```

### Data Structure Optimization

```python
import sys
from array import array
import struct

# Compare memory usage of different data structures
def compare_data_structures():
    size = 10000
    
    # Python list
    py_list = list(range(size))
    print(f"Python list: {sys.getsizeof(py_list)} bytes")
    
    # Array module
    int_array = array('i', range(size))
    print(f"Array module: {sys.getsizeof(int_array)} bytes")
    
    # Bytes object
    byte_data = struct.pack(f'{size}i', *range(size))
    print(f"Bytes object: {sys.getsizeof(byte_data)} bytes")
    
    # Memory view
    mv = memoryview(byte_data)
    print(f"Memory view: {sys.getsizeof(mv)} bytes")

compare_data_structures()

# Efficient string operations
def efficient_string_building():
    import io
    
    # Inefficient - creates many intermediate strings
    def inefficient_join(words):
        result = ""
        for word in words:
            result += word + " "
        return result
    
    # Efficient - uses join
    def efficient_join(words):
        return " ".join(words) + " "
    
    # Very efficient - uses StringIO
    def stringio_join(words):
        buffer = io.StringIO()
        for word in words:
            buffer.write(word)
            buffer.write(" ")
        return buffer.getvalue()
    
    words = ["word"] * 10000
    
    import time
    
    # Time inefficient method
    start = time.time()
    result1 = inefficient_join(words[:1000])  # Smaller set for timing
    inefficient_time = time.time() - start
    
    # Time efficient method
    start = time.time()
    result2 = efficient_join(words)
    efficient_time = time.time() - start
    
    # Time StringIO method
    start = time.time()
    result3 = stringio_join(words)
    stringio_time = time.time() - start
    
    print(f"Inefficient: {inefficient_time:.4f}s")
    print(f"Efficient: {efficient_time:.4f}s")
    print(f"StringIO: {stringio_time:.4f}s")

efficient_string_building()
```

## Memory Profiling and Monitoring

### Using tracemalloc

```python
import tracemalloc
import linecache

def memory_profiling_example():
    # Start tracing
    tracemalloc.start()
    
    # Take snapshot before allocation
    snapshot1 = tracemalloc.take_snapshot()
    
    # Allocate memory
    data = []
    for i in range(10000):
        data.append([i] * 100)
    
    # Take snapshot after allocation
    snapshot2 = tracemalloc.take_snapshot()
    
    # Compare snapshots
    top_stats = snapshot2.compare_to(snapshot1, 'lineno')
    
    print("Top 10 memory allocations:")
    for index, stat in enumerate(top_stats[:10], 1):
        print(f"{index}. {stat}")
        
        # Show source code
        frame = stat.traceback.frames[0]
        print(f"   {linecache.getline(frame.filename, frame.lineno).strip()}")
    
    # Current memory usage
    current, peak = tracemalloc.get_traced_memory()
    print(f"Current memory usage: {current / 1024 / 1024:.1f} MB")
    print(f"Peak memory usage: {peak / 1024 / 1024:.1f} MB")
    
    tracemalloc.stop()

memory_profiling_example()
```

### Memory Monitoring Tools

```python
import psutil
import os

def monitor_process_memory():
    """Monitor current process memory usage"""
    process = psutil.Process(os.getpid())
    
    # Memory info
    memory_info = process.memory_info()
    print(f"RSS (Resident Set Size): {memory_info.rss / 1024 / 1024:.1f} MB")
    print(f"VMS (Virtual Memory Size): {memory_info.vms / 1024 / 1024:.1f} MB")
    
    # Memory percentage
    memory_percent = process.memory_percent()
    print(f"Memory percentage: {memory_percent:.2f}%")
    
    # Detailed memory info (Linux/macOS)
    try:
        memory_full = process.memory_full_info()
        print(f"USS (Unique Set Size): {memory_full.uss / 1024 / 1024:.1f} MB")
        print(f"PSS (Proportional Set Size): {memory_full.pss / 1024 / 1024:.1f} MB")
    except AttributeError:
        print("Extended memory info not available on this platform")

monitor_process_memory()

# Memory context manager
class MemoryMonitor:
    def __init__(self, description=""):
        self.description = description
        self.process = psutil.Process(os.getpid())
    
    def __enter__(self):
        self.start_memory = self.process.memory_info().rss
        print(f"Memory before {self.description}: {self.start_memory / 1024 / 1024:.1f} MB")
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        end_memory = self.process.memory_info().rss
        diff = end_memory - self.start_memory
        print(f"Memory after {self.description}: {end_memory / 1024 / 1024:.1f} MB")
        print(f"Memory difference: {diff / 1024 / 1024:.1f} MB")

# Usage
with MemoryMonitor("large allocation"):
    large_list = [i for i in range(1000000)]
```

### Memory Leaks Detection

```python
import gc
import weakref
from collections import defaultdict

class MemoryLeakDetector:
    """Simple memory leak detector"""
    
    def __init__(self):
        self.snapshots = []
    
    def take_snapshot(self, name=""):
        """Take a snapshot of current objects"""
        gc.collect()  # Force garbage collection first
        
        objects_by_type = defaultdict(int)
        for obj in gc.get_objects():
            obj_type = type(obj).__name__
            objects_by_type[obj_type] += 1
        
        snapshot = {
            'name': name,
            'objects': dict(objects_by_type),
            'total': len(gc.get_objects())
        }
        
        self.snapshots.append(snapshot)
        return snapshot
    
    def compare_snapshots(self, index1=-2, index2=-1):
        """Compare two snapshots to find potential leaks"""
        if len(self.snapshots) < 2:
            print("Need at least 2 snapshots to compare")
            return
        
        snap1 = self.snapshots[index1]
        snap2 = self.snapshots[index2]
        
        print(f"Comparing '{snap1['name']}' to '{snap2['name']}'")
        print(f"Total objects: {snap1['total']} → {snap2['total']} ({snap2['total'] - snap1['total']:+d})")
        
        # Find object type differences
        all_types = set(snap1['objects'].keys()) | set(snap2['objects'].keys())
        
        differences = []
        for obj_type in all_types:
            count1 = snap1['objects'].get(obj_type, 0)
            count2 = snap2['objects'].get(obj_type, 0)
            diff = count2 - count1
            
            if diff != 0:
                differences.append((obj_type, count1, count2, diff))
        
        # Sort by difference
        differences.sort(key=lambda x: abs(x[3]), reverse=True)
        
        print("\nTop object type changes:")
        for obj_type, count1, count2, diff in differences[:10]:
            print(f"  {obj_type}: {count1} → {count2} ({diff:+d})")

# Usage example
detector = MemoryLeakDetector()

# Initial snapshot
detector.take_snapshot("initial")

# Simulate memory usage
data_store = []
for i in range(1000):
    data_store.append([i] * 100)

detector.take_snapshot("after allocation")

# Simulate partial cleanup
del data_store[::2]  # Delete every other item

detector.take_snapshot("after partial cleanup")

# Compare snapshots
detector.compare_snapshots(0, 1)  # initial vs after allocation
detector.compare_snapshots(1, 2)  # after allocation vs after cleanup
```

## Advanced Memory Techniques

### Memory Mapping

```python
import mmap
import os

def memory_mapping_example():
    # Create a test file
    filename = "test_mmap.dat"
    with open(filename, "wb") as f:
        f.write(b"\x00" * 1024 * 1024)  # 1MB of zeros
    
    # Memory map the file
    with open(filename, "r+b") as f:
        with mmap.mmap(f.fileno(), 0) as mm:
            # Write to memory map
            mm[0:5] = b"Hello"
            mm[100:105] = b"World"
            
            # Read from memory map
            print(f"First 5 bytes: {mm[0:5]}")
            print(f"Bytes 100-105: {mm[100:105]}")
            
            # Find data
            pos = mm.find(b"World")
            print(f"Found 'World' at position: {pos}")
            
            # Memory map behaves like bytes
            print(f"Total size: {len(mm)} bytes")
    
    # Clean up
    os.unlink(filename)

memory_mapping_example()
```

### Custom Memory Allocators

```python
import sys
from pymalloc import PyMalloc  # Hypothetical custom allocator

# Object pooling for memory efficiency
class ObjectPool:
    """Simple object pool to reduce allocation overhead"""
    
    def __init__(self, factory, initial_size=10):
        self.factory = factory
        self.pool = [factory() for _ in range(initial_size)]
        self.active = set()
    
    def acquire(self):
        if self.pool:
            obj = self.pool.pop()
        else:
            obj = self.factory()
        
        self.active.add(id(obj))
        return obj
    
    def release(self, obj):
        obj_id = id(obj)
        if obj_id in self.active:
            self.active.remove(obj_id)
            # Reset object state if needed
            if hasattr(obj, 'reset'):
                obj.reset()
            self.pool.append(obj)
    
    def stats(self):
        return {
            'pool_size': len(self.pool),
            'active_objects': len(self.active)
        }

# Example usage
class ExpensiveObject:
    def __init__(self):
        self.data = [0] * 1000
        self.processed = False
    
    def reset(self):
        self.data = [0] * 1000
        self.processed = False
    
    def process(self):
        # Simulate expensive processing
        self.data = [i ** 2 for i in self.data]
        self.processed = True

# Use object pool
pool = ObjectPool(ExpensiveObject, initial_size=5)

# Acquire and use objects
objs = []
for i in range(10):
    obj = pool.acquire()
    obj.process()
    objs.append(obj)

print(f"Pool stats after acquisition: {pool.stats()}")

# Release objects back to pool
for obj in objs:
    pool.release(obj)

print(f"Pool stats after release: {pool.stats()}")
```

### Memory-Mapped Data Structures

```python
import struct
import mmap
import os

class MemoryMappedArray:
    """Memory-mapped array for large datasets"""
    
    def __init__(self, filename, dtype='i', mode='r+'):
        self.filename = filename
        self.dtype = dtype
        self.mode = mode
        
        # Get struct format and size
        self.format = {'i': 'i', 'f': 'f', 'd': 'd'}[dtype]
        self.item_size = struct.calcsize(self.format)
        
        # Open file and create memory map
        self.file = open(filename, mode + 'b')
        self.mmap = mmap.mmap(self.file.fileno(), 0)
        
        # Calculate array length
        self.length = len(self.mmap) // self.item_size
    
    def __len__(self):
        return self.length
    
    def __getitem__(self, index):
        if index >= self.length:
            raise IndexError("Index out of range")
        
        offset = index * self.item_size
        data = self.mmap[offset:offset + self.item_size]
        return struct.unpack(self.format, data)[0]
    
    def __setitem__(self, index, value):
        if index >= self.length:
            raise IndexError("Index out of range")
        
        offset = index * self.item_size
        packed = struct.pack(self.format, value)
        self.mmap[offset:offset + self.item_size] = packed
    
    def close(self):
        if hasattr(self, 'mmap'):
            self.mmap.close()
        if hasattr(self, 'file'):
            self.file.close()
    
    def __del__(self):
        self.close()

# Create a memory-mapped array
filename = "large_array.dat"
with open(filename, "wb") as f:
    # Write 1 million integers
    for i in range(1000000):
        f.write(struct.pack('i', i))

# Use memory-mapped array
array = MemoryMappedArray(filename, dtype='i')
print(f"Array length: {len(array)}")
print(f"First 10 elements: {[array[i] for i in range(10)]}")
print(f"Last 10 elements: {[array[i] for i in range(len(array)-10, len(array))]}")

# Modify elements
array[0] = 999
array[1] = 888

print(f"Modified first 2: {array[0]}, {array[1]}")

array.close()
os.unlink(filename)
```

This comprehensive guide covers Python's memory management from basic concepts like reference counting to advanced techniques like memory mapping and object pooling, providing the knowledge needed to write memory-efficient Python applications.