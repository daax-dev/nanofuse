# C# Type System and Collections

## Overview
Comprehensive guide to C#'s type system, built-in collections, and advanced collection patterns for efficient data management and manipulation.

## C# Type System Fundamentals

### Value Types vs Reference Types

#### Value Types
```csharp
// Built-in value types
int number = 42;                    // Stack allocated
bool flag = true;                   // Stack allocated
DateTime now = DateTime.Now;        // Stack allocated
decimal price = 99.99m;             // Stack allocated

// Custom value types (structs)
public struct Point
{
    public int X { get; set; }
    public int Y { get; set; }

    public Point(int x, int y)
    {
        X = x;
        Y = y;
    }
}

Point p1 = new Point(10, 20);       // Stack allocated
Point p2 = p1;                      // Value copied
p2.X = 30;                          // p1.X remains 10
```

#### Reference Types
```csharp
// Built-in reference types
string text = "Hello";              // Heap allocated
object obj = new object();          // Heap allocated
int[] numbers = { 1, 2, 3 };        // Heap allocated

// Custom reference types (classes)
public class Person
{
    public string Name { get; set; }
    public int Age { get; set; }
}

Person p1 = new Person { Name = "John", Age = 30 };
Person p2 = p1;                     // Reference copied
p2.Age = 31;                        // p1.Age also becomes 31
```

### Nullable Types

#### Nullable Value Types
```csharp
// Traditional nullable types (C# 2.0)
int? nullableInt = null;
bool? nullableBool = true;
DateTime? nullableDate = DateTime.Now;

// Checking for null
if (nullableInt.HasValue)
{
    int value = nullableInt.Value;
}

// Null-coalescing operator
int actualValue = nullableInt ?? 0;

// Nullable operators
int? a = 5;
int? b = null;
int? sum = a + b;                   // Result is null
```

#### Nullable Reference Types (C# 8+)
```csharp
#nullable enable

public class PersonService
{
    public string? FindPersonName(int id)
    {
        // May return null
        return id > 0 ? "John Doe" : null;
    }

    public string GetDisplayName(Person person)
    {
        // person cannot be null in this context
        return person.Name ?? "Unknown";
    }

    public void ProcessPerson(Person? person)
    {
        if (person is not null)
        {
            // Safe to use person here
            Console.WriteLine(person.Name);
        }
    }
}

#nullable restore
```

## Built-in Primitive Types

### Numeric Types
```csharp
// Integral types
byte byteValue = 255;               // 0 to 255
sbyte sbyteValue = -128;            // -128 to 127
short shortValue = -32768;          // -32,768 to 32,767
ushort ushortValue = 65535;         // 0 to 65,535
int intValue = -2147483648;         // -2,147,483,648 to 2,147,483,647
uint uintValue = 4294967295;        // 0 to 4,294,967,295
long longValue = -9223372036854775808L;
ulong ulongValue = 18446744073709551615UL;

// Floating-point types
float floatValue = 3.14f;           // ~7 digits precision
double doubleValue = 3.14159265359; // ~15-17 digits precision
decimal decimalValue = 99.99m;      // 28-29 digits precision (financial)

// Character and string types
char charValue = 'A';               // Unicode character
string stringValue = "Hello World"; // Immutable string

// Boolean type
bool boolValue = true;              // true or false

// Type information
Console.WriteLine($"int size: {sizeof(int)} bytes");
Console.WriteLine($"double range: {double.MinValue} to {double.MaxValue}");
```

### Special Types
```csharp
// Object type (base of all types)
object obj = 42;                    // Boxing
object obj2 = "Hello";              // Reference assignment
int unboxed = (int)obj;             // Unboxing

// Dynamic type (runtime binding)
dynamic dyn = "Hello";
Console.WriteLine(dyn.Length);      // Resolved at runtime
dyn = 42;
Console.WriteLine(dyn + 10);        // Resolved at runtime

// Void type (no return value)
void DoSomething() { }

// Var (implicit typing)
var inferredInt = 42;               // Inferred as int
var inferredString = "Hello";       // Inferred as string
var inferredArray = new[] { 1, 2, 3 }; // Inferred as int[]
```

## Generic Types and Constraints

### Basic Generics
```csharp
// Generic class
public class Container<T>
{
    private T _item;

    public void Set(T item) => _item = item;
    public T Get() => _item;
}

// Usage
Container<int> intContainer = new Container<int>();
Container<string> stringContainer = new Container<string>();

// Generic methods
public static T GetDefault<T>() => default(T);
public static void Swap<T>(ref T a, ref T b)
{
    T temp = a;
    a = b;
    b = temp;
}
```

### Generic Constraints
```csharp
// Class constraint
public class Repository<T> where T : class
{
    public T Find(int id) => default(T);
}

// Struct constraint
public class ValueProcessor<T> where T : struct
{
    public T Process(T value) => value;
}

// Interface constraint
public class Sorter<T> where T : IComparable<T>
{
    public void Sort(T[] array) => Array.Sort(array);
}

// Multiple constraints
public class EntityRepository<T>
    where T : class, IEntity, new()
{
    public T CreateNew() => new T();
}

// Type parameter constraints
public class Converter<TInput, TOutput>
    where TInput : class
    where TOutput : struct
{
    public TOutput Convert(TInput input) => default(TOutput);
}
```

### Covariance and Contravariance
```csharp
// Covariance (out keyword)
public interface IProducer<out T>
{
    T Produce();
}

IProducer<string> stringProducer = new StringProducer();
IProducer<object> objectProducer = stringProducer; // Valid

// Contravariance (in keyword)
public interface IConsumer<in T>
{
    void Consume(T item);
}

IConsumer<object> objectConsumer = new ObjectConsumer();
IConsumer<string> stringConsumer = objectConsumer; // Valid

// Invariance (no keyword)
public interface IContainer<T>
{
    T Get();
    void Set(T item);
}
```

## Arrays

### Single-Dimensional Arrays
```csharp
// Array declaration and initialization
int[] numbers = new int[5];         // Default values (0)
int[] values = { 1, 2, 3, 4, 5 };   // Literal initialization
string[] names = new string[] { "Alice", "Bob", "Charlie" };

// Array operations
Console.WriteLine($"Length: {values.Length}");
Console.WriteLine($"First: {values[0]}");
Console.WriteLine($"Last: {values[^1]}");    // Index from end

// Array slicing (C# 8+)
int[] slice = values[1..4];         // { 2, 3, 4 }
int[] fromStart = values[..3];      // { 1, 2, 3 }
int[] toEnd = values[2..];          // { 3, 4, 5 }

// Array iteration
foreach (int value in values)
{
    Console.WriteLine(value);
}

for (int i = 0; i < values.Length; i++)
{
    Console.WriteLine($"Index {i}: {values[i]}");
}
```

### Multi-Dimensional Arrays
```csharp
// Rectangular arrays
int[,] matrix = new int[3, 4];      // 3 rows, 4 columns
int[,] grid = {
    { 1, 2, 3, 4 },
    { 5, 6, 7, 8 },
    { 9, 10, 11, 12 }
};

// Access elements
matrix[0, 0] = 1;
int value = grid[1, 2];             // 7

// Dimensions
Console.WriteLine($"Rank: {grid.Rank}");           // 2
Console.WriteLine($"Rows: {grid.GetLength(0)}");   // 3
Console.WriteLine($"Cols: {grid.GetLength(1)}");   // 4

// Iteration
for (int i = 0; i < grid.GetLength(0); i++)
{
    for (int j = 0; j < grid.GetLength(1); j++)
    {
        Console.Write($"{grid[i, j]} ");
    }
    Console.WriteLine();
}
```

### Jagged Arrays
```csharp
// Jagged array declaration
int[][] jaggedArray = new int[3][];
jaggedArray[0] = new int[4];        // 4 elements
jaggedArray[1] = new int[2];        // 2 elements
jaggedArray[2] = new int[3];        // 3 elements

// Literal initialization
int[][] numbers = {
    new int[] { 1, 2, 3, 4 },
    new int[] { 5, 6 },
    new int[] { 7, 8, 9 }
};

// Access and iteration
numbers[0][2] = 100;                // Modify element

for (int i = 0; i < numbers.Length; i++)
{
    for (int j = 0; j < numbers[i].Length; j++)
    {
        Console.Write($"{numbers[i][j]} ");
    }
    Console.WriteLine();
}

// Using foreach
foreach (int[] row in numbers)
{
    foreach (int value in row)
    {
        Console.Write($"{value} ");
    }
    Console.WriteLine();
}
```

## Collection Interfaces

### Core Collection Interfaces
```csharp
// IEnumerable<T> - Basic iteration
public interface IEnumerable<out T> : IEnumerable
{
    IEnumerator<T> GetEnumerator();
}

// ICollection<T> - Size and modification
public interface ICollection<T> : IEnumerable<T>
{
    int Count { get; }
    bool IsReadOnly { get; }
    void Add(T item);
    bool Remove(T item);
    bool Contains(T item);
    void Clear();
    void CopyTo(T[] array, int arrayIndex);
}

// IList<T> - Indexed access
public interface IList<T> : ICollection<T>
{
    T this[int index] { get; set; }
    int IndexOf(T item);
    void Insert(int index, T item);
    void RemoveAt(int index);
}

// IDictionary<TKey, TValue> - Key-value pairs
public interface IDictionary<TKey, TValue> : ICollection<KeyValuePair<TKey, TValue>>
{
    TValue this[TKey key] { get; set; }
    ICollection<TKey> Keys { get; }
    ICollection<TValue> Values { get; }
    bool ContainsKey(TKey key);
    bool TryGetValue(TKey key, out TValue value);
}
```

### Custom Collection Example
```csharp
public class ObservableList<T> : IList<T>, INotifyCollectionChanged
{
    private readonly List<T> _items = new List<T>();

    public event NotifyCollectionChangedEventHandler CollectionChanged;

    public T this[int index]
    {
        get => _items[index];
        set
        {
            T oldValue = _items[index];
            _items[index] = value;
            OnCollectionChanged(new NotifyCollectionChangedEventArgs(
                NotifyCollectionChangedAction.Replace, value, oldValue, index));
        }
    }

    public void Add(T item)
    {
        _items.Add(item);
        OnCollectionChanged(new NotifyCollectionChangedEventArgs(
            NotifyCollectionChangedAction.Add, item, _items.Count - 1));
    }

    // Implement other IList<T> members...

    protected virtual void OnCollectionChanged(NotifyCollectionChangedEventArgs e)
    {
        CollectionChanged?.Invoke(this, e);
    }
}
```

## Standard Collections

### List<T>
```csharp
// Creation and initialization
List<int> numbers = new List<int>();
List<string> names = new List<string> { "Alice", "Bob", "Charlie" };
List<int> range = Enumerable.Range(1, 10).ToList();

// Common operations
numbers.Add(42);                    // Add single item
numbers.AddRange(new[] { 1, 2, 3 });   // Add multiple items
numbers.Insert(0, 100);             // Insert at index
bool contains = numbers.Contains(42);   // Check existence
int index = numbers.IndexOf(42);    // Find index
numbers.Remove(42);                 // Remove first occurrence
numbers.RemoveAt(0);                // Remove by index
numbers.Clear();                    // Remove all items

// Searching and filtering
List<int> evens = numbers.Where(n => n % 2 == 0).ToList();
int? firstEven = numbers.FirstOrDefault(n => n % 2 == 0);
bool anyPositive = numbers.Any(n => n > 0);
bool allPositive = numbers.All(n => n > 0);

// Sorting
numbers.Sort();                     // Natural order
numbers.Sort((a, b) => b.CompareTo(a)); // Descending
numbers.Reverse();                  // Reverse order

// Capacity management
Console.WriteLine($"Count: {numbers.Count}, Capacity: {numbers.Capacity}");
numbers.TrimExcess();               // Reduce capacity to count
```

### Dictionary<TKey, TValue>
```csharp
// Creation and initialization
Dictionary<string, int> ages = new Dictionary<string, int>();
Dictionary<string, string> capitals = new Dictionary<string, string>
{
    ["USA"] = "Washington, D.C.",
    ["UK"] = "London",
    ["France"] = "Paris"
};

// Basic operations
ages["Alice"] = 30;                 // Add or update
ages.Add("Bob", 25);                // Add (throws if exists)
bool hasAlice = ages.ContainsKey("Alice");
bool hasAge30 = ages.ContainsValue(30);

// Safe retrieval
if (ages.TryGetValue("Alice", out int aliceAge))
{
    Console.WriteLine($"Alice is {aliceAge} years old");
}

// Iteration
foreach (KeyValuePair<string, int> kvp in ages)
{
    Console.WriteLine($"{kvp.Key}: {kvp.Value}");
}

foreach (string name in ages.Keys)
{
    Console.WriteLine($"Name: {name}");
}

foreach (int age in ages.Values)
{
    Console.WriteLine($"Age: {age}");
}

// Removal
ages.Remove("Bob");                 // Remove by key
ages.Clear();                       // Remove all items
```

### HashSet<T>
```csharp
// Creation and initialization
HashSet<int> uniqueNumbers = new HashSet<int>();
HashSet<string> uniqueNames = new HashSet<string> { "Alice", "Bob", "Alice" }; // Only one Alice

// Basic operations
uniqueNumbers.Add(1);               // Returns true if added
uniqueNumbers.Add(1);               // Returns false (already exists)
bool contains = uniqueNumbers.Contains(1);
bool removed = uniqueNumbers.Remove(1);

// Set operations
HashSet<int> set1 = new HashSet<int> { 1, 2, 3, 4 };
HashSet<int> set2 = new HashSet<int> { 3, 4, 5, 6 };

set1.UnionWith(set2);               // Union: { 1, 2, 3, 4, 5, 6 }
set1.IntersectWith(set2);           // Intersection: { 3, 4 }
set1.ExceptWith(set2);              // Difference: { 1, 2 }
set1.SymmetricExceptWith(set2);     // Symmetric difference

// Set comparisons
bool isSubset = set1.IsSubsetOf(set2);
bool isSuperset = set1.IsSupersetOf(set2);
bool overlaps = set1.Overlaps(set2);
```

### Queue<T> and Stack<T>
```csharp
// Queue (FIFO - First In, First Out)
Queue<string> queue = new Queue<string>();
queue.Enqueue("First");
queue.Enqueue("Second");
queue.Enqueue("Third");

string first = queue.Peek();        // "First" (doesn't remove)
string dequeued = queue.Dequeue();  // "First" (removes)
int count = queue.Count;            // 2

// Stack (LIFO - Last In, First Out)
Stack<int> stack = new Stack<int>();
stack.Push(1);
stack.Push(2);
stack.Push(3);

int top = stack.Peek();             // 3 (doesn't remove)
int popped = stack.Pop();           // 3 (removes)
bool hasItems = stack.Count > 0;

// Use cases
// Queue: Task scheduling, breadth-first search
// Stack: Undo operations, depth-first search, expression evaluation
```

## Specialized Collections

### Concurrent Collections
```csharp
// Thread-safe collections
ConcurrentDictionary<string, int> concurrentDict = new ConcurrentDictionary<string, int>();
ConcurrentQueue<string> concurrentQueue = new ConcurrentQueue<string>();
ConcurrentStack<int> concurrentStack = new ConcurrentStack<int>();
ConcurrentBag<double> concurrentBag = new ConcurrentBag<double>();

// ConcurrentDictionary operations
concurrentDict.TryAdd("key1", 100);
concurrentDict.AddOrUpdate("key2", 200, (key, oldValue) => oldValue + 10);
int value = concurrentDict.GetOrAdd("key3", 300);

if (concurrentDict.TryGetValue("key1", out int result))
{
    Console.WriteLine($"Value: {result}");
}

// ConcurrentQueue operations
concurrentQueue.Enqueue("item1");
if (concurrentQueue.TryDequeue(out string item))
{
    Console.WriteLine($"Dequeued: {item}");
}

// Producer-consumer pattern
BlockingCollection<int> blockingCollection = new BlockingCollection<int>();

// Producer
Task.Run(() =>
{
    for (int i = 0; i < 10; i++)
    {
        blockingCollection.Add(i);
        Thread.Sleep(100);
    }
    blockingCollection.CompleteAdding();
});

// Consumer
Task.Run(() =>
{
    foreach (int item in blockingCollection.GetConsumingEnumerable())
    {
        Console.WriteLine($"Processing: {item}");
    }
});
```

### Immutable Collections
```csharp
using System.Collections.Immutable;

// Immutable collections
ImmutableList<int> immutableList = ImmutableList.Create(1, 2, 3);
ImmutableDictionary<string, int> immutableDict = ImmutableDictionary.Create<string, int>();
ImmutableHashSet<string> immutableSet = ImmutableHashSet.Create("a", "b", "c");

// Operations return new instances
ImmutableList<int> newList = immutableList.Add(4);     // Original unchanged
ImmutableDictionary<string, int> newDict = immutableDict.Add("key", 100);

// Builder pattern for efficiency
ImmutableList<int>.Builder builder = ImmutableList.CreateBuilder<int>();
for (int i = 0; i < 1000; i++)
{
    builder.Add(i);
}
ImmutableList<int> finalList = builder.ToImmutable();

// Benefits: Thread-safety, functional programming, undo/redo
```

### Read-Only Collections
```csharp
// ReadOnlyCollection<T>
List<int> mutableList = new List<int> { 1, 2, 3, 4, 5 };
ReadOnlyCollection<int> readOnlyList = new ReadOnlyCollection<int>(mutableList);

// IReadOnlyCollection, IReadOnlyList, IReadOnlyDictionary
IReadOnlyList<int> readOnlyInterface = mutableList.AsReadOnly();
IReadOnlyDictionary<string, int> readOnlyDict = new Dictionary<string, int>
{
    ["key1"] = 100,
    ["key2"] = 200
};

// Collection expressions (C# 12)
ReadOnlySpan<int> span = [1, 2, 3, 4, 5];
ImmutableArray<string> array = ["a", "b", "c"];
```

## Custom Collection Implementation

### Generic Collection Example
```csharp
public class CircularBuffer<T> : ICollection<T>
{
    private readonly T[] _buffer;
    private int _start;
    private int _end;
    private int _count;

    public CircularBuffer(int capacity)
    {
        if (capacity <= 0) throw new ArgumentException("Capacity must be positive");
        _buffer = new T[capacity];
    }

    public int Count => _count;
    public bool IsReadOnly => false;
    public int Capacity => _buffer.Length;

    public void Add(T item)
    {
        _buffer[_end] = item;
        _end = (_end + 1) % _buffer.Length;

        if (_count < _buffer.Length)
        {
            _count++;
        }
        else
        {
            _start = (_start + 1) % _buffer.Length;
        }
    }

    public bool Remove(T item)
    {
        throw new NotSupportedException("Remove not supported on circular buffer");
    }

    public void Clear()
    {
        Array.Clear(_buffer, 0, _buffer.Length);
        _start = _end = _count = 0;
    }

    public bool Contains(T item)
    {
        return this.Any(x => EqualityComparer<T>.Default.Equals(x, item));
    }

    public void CopyTo(T[] array, int arrayIndex)
    {
        if (array == null) throw new ArgumentNullException(nameof(array));
        if (arrayIndex < 0) throw new ArgumentOutOfRangeException(nameof(arrayIndex));
        if (array.Length - arrayIndex < _count) throw new ArgumentException("Insufficient space");

        foreach (T item in this)
        {
            array[arrayIndex++] = item;
        }
    }

    public IEnumerator<T> GetEnumerator()
    {
        int current = _start;
        for (int i = 0; i < _count; i++)
        {
            yield return _buffer[current];
            current = (current + 1) % _buffer.Length;
        }
    }

    IEnumerator IEnumerable.GetEnumerator() => GetEnumerator();
}
```

## Performance Considerations

### Collection Selection Guide
```csharp
public static class CollectionPerformanceGuide
{
    // Use cases and performance characteristics

    // Array: Fixed size, fastest access O(1), minimal overhead
    // Best for: Known size, frequent random access

    // List<T>: Dynamic size, O(1) append, O(n) insert/remove from middle
    // Best for: Unknown size, mostly append operations

    // LinkedList<T>: O(1) insert/remove at any position, O(n) access by index
    // Best for: Frequent insert/remove operations in middle

    // Dictionary<TKey, TValue>: O(1) average lookup, insert, delete
    // Best for: Key-based lookup, unique keys

    // HashSet<T>: O(1) average insert, remove, contains
    // Best for: Unique items, set operations

    // SortedDictionary<TKey, TValue>: O(log n) operations, maintains sort order
    // Best for: Sorted key-value pairs, range queries

    // Queue<T>: O(1) enqueue/dequeue, FIFO
    // Best for: Task scheduling, breadth-first traversal

    // Stack<T>: O(1) push/pop, LIFO
    // Best for: Undo operations, depth-first traversal
}

// Benchmark example
[MemoryDiagnoser]
public class CollectionBenchmarks
{
    private const int ItemCount = 10000;

    [Benchmark]
    public void ArrayAccess()
    {
        int[] array = new int[ItemCount];
        for (int i = 0; i < ItemCount; i++)
        {
            array[i] = i;
        }
    }

    [Benchmark]
    public void ListAdd()
    {
        List<int> list = new List<int>();
        for (int i = 0; i < ItemCount; i++)
        {
            list.Add(i);
        }
    }

    [Benchmark]
    public void DictionaryLookup()
    {
        Dictionary<int, int> dict = new Dictionary<int, int>();
        for (int i = 0; i < ItemCount; i++)
        {
            dict[i] = i;
        }

        for (int i = 0; i < ItemCount; i++)
        {
            _ = dict[i];
        }
    }
}
```

### Memory-Efficient Patterns
```csharp
// Object pooling for collections
public class ListPool<T>
{
    private static readonly ConcurrentQueue<List<T>> Pool = new ConcurrentQueue<List<T>>();

    public static List<T> Rent()
    {
        if (Pool.TryDequeue(out List<T> list))
        {
            return list;
        }
        return new List<T>();
    }

    public static void Return(List<T> list)
    {
        list.Clear();
        Pool.Enqueue(list);
    }
}

// Usage with using pattern
public ref struct ListRental<T>
{
    private List<T> _list;

    public ListRental(List<T> list)
    {
        _list = list;
    }

    public List<T> List => _list;

    public void Dispose()
    {
        if (_list != null)
        {
            ListPool<T>.Return(_list);
            _list = null;
        }
    }
}

// Span<T> for stack-allocated collections
public void ProcessSmallCollection()
{
    Span<int> numbers = stackalloc int[10];
    for (int i = 0; i < numbers.Length; i++)
    {
        numbers[i] = i * i;
    }

    // Process without heap allocation
    int sum = 0;
    foreach (int number in numbers)
    {
        sum += number;
    }
}
```

### LINQ Performance Optimization
```csharp
// Efficient LINQ patterns
public class LinqOptimization
{
    public void EfficientFiltering(IEnumerable<Person> people)
    {
        // Use Where before Select to reduce processing
        var youngAdultNames = people
            .Where(p => p.Age >= 18 && p.Age < 30)    // Filter first
            .Select(p => p.Name)                      // Transform after filtering
            .ToList();

        // Use Any/All instead of Count when possible
        bool hasAdults = people.Any(p => p.Age >= 18);        // Efficient
        bool hasAdultsBad = people.Count(p => p.Age >= 18) > 0; // Inefficient

        // Use FirstOrDefault instead of Where().First()
        Person firstAdult = people.FirstOrDefault(p => p.Age >= 18);    // Efficient
        Person firstAdultBad = people.Where(p => p.Age >= 18).First();  // Inefficient

        // Pre-compile predicates for reuse
        Func<Person, bool> isAdult = p => p.Age >= 18;
        var adults = people.Where(isAdult).ToList();
        var adultCount = people.Count(isAdult);
    }

    public void AvoidMultipleEnumeration(IEnumerable<int> numbers)
    {
        // Bad: Multiple enumeration
        if (numbers.Any())
        {
            int sum = numbers.Sum();        // Enumerates again
            int count = numbers.Count();    // Enumerates again
        }

        // Good: Single enumeration
        var materializedNumbers = numbers.ToList(); // Or ToArray()
        if (materializedNumbers.Any())
        {
            int sum = materializedNumbers.Sum();
            int count = materializedNumbers.Count;
        }
    }
}
```

This comprehensive guide covers C#'s type system and collections from fundamental concepts to advanced optimization techniques, providing both theoretical understanding and practical implementation patterns.