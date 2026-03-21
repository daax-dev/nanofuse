# C# Keywords and Syntax Reference

## Overview
Comprehensive reference of all C# keywords, operators, and syntax constructs from C# 1.0 through C# 12.

## Reserved Keywords

### Type Keywords
```csharp
// Primitive Types
bool isActive = true;
byte age = 255;
sbyte temperature = -10;
char grade = 'A';
decimal price = 99.99m;
double pi = 3.14159;
float ratio = 0.5f;
int count = 42;
uint positive = 100u;
long bigNumber = 9223372036854775807L;
ulong hugNumber = 18446744073709551615UL;
short small = 32767;
ushort positive16 = 65535;

// Reference Types
string name = "Hello";
object obj = new object();
dynamic flexible = "anything";

// Special Types
void DoNothing() { }
var inferred = "Type inferred";
```

### Control Flow Keywords
```csharp
// Selection Statements
if (condition) { }
else if (otherCondition) { }
else { }

switch (value)
{
    case 1:
        break;
    case 2:
        goto case 1;
    default:
        return;
}

// Iteration Statements
for (int i = 0; i < 10; i++) { }
foreach (var item in collection) { }
while (condition) { }
do { } while (condition);

// Jump Statements
break;
continue;
goto label;
return value;
```

### Access Modifiers
```csharp
public class PublicClass { }
private int privateField;
protected string protectedProperty { get; set; }
internal void InternalMethod() { }
protected internal int ProtectedInternalField;
private protected string PrivateProtectedProperty { get; set; } // C# 7.2
```

### Modifiers
```csharp
abstract class AbstractClass { }
sealed class SealedClass { }
static class StaticClass { }
partial class PartialClass { }

virtual void VirtualMethod() { }
override void OverrideMethod() { }
abstract void AbstractMethod();

const int CONSTANT = 42;
static readonly DateTime StartTime = DateTime.Now;
readonly int ReadOnlyField = 100;

extern static void ExternalMethod();
volatile bool flag;
```

### Object-Oriented Keywords
```csharp
class MyClass : BaseClass, IInterface
{
    public MyClass() : base() { } // constructor with base call

    public new void HideBaseMethod() { } // method hiding

    interface INestedInterface { }

    enum Color { Red, Green, Blue }

    struct Point { public int X, Y; }

    delegate void EventHandler();

    event EventHandler SomethingHappened;
}

// Inheritance and Implementation
public class Derived : Base, IComparable
{
    public int CompareTo(object obj) => 0;
}
```

### Exception Handling
```csharp
try
{
    RiskyOperation();
}
catch (SpecificException ex) when (ex.Message.Contains("error"))
{
    // Handle specific exception with condition
}
catch (Exception ex)
{
    // Handle general exception
    throw; // Re-throw preserving stack trace
}
finally
{
    // Cleanup code
}

// Custom exceptions
throw new ArgumentException("Invalid argument");
```

### Memory and Resource Management
```csharp
// Using statements
using System;
using System.Collections.Generic;
using static System.Math; // Static using (C# 6)

// Using declarations and statements
using var stream = new FileStream("file.txt", FileMode.Open); // C# 8
using (var connection = new SqlConnection(connectionString))
{
    // Resource automatically disposed
}

// Unsafe code
unsafe
{
    int* ptr = stackalloc int[100];
    fixed (byte* p = buffer)
    {
        // Work with pointer
    }
}

// Memory management
IntPtr nativePointer;
sizeof(int); // Size of type
stackalloc int[10]; // Stack allocation
```

## Contextual Keywords

### LINQ Keywords
```csharp
var query = from item in collection
            where item.IsActive
            join other in otherCollection on item.Id equals other.ItemId
            let computed = item.Value * 2
            group item by item.Category into g
            orderby g.Key ascending
            select new { Category = g.Key, Count = g.Count() };

// Method syntax equivalent
var methodQuery = collection
    .Where(item => item.IsActive)
    .Join(otherCollection,
          item => item.Id,
          other => other.ItemId,
          (item, other) => new { item, other })
    .Select(x => new { x.item, computed = x.item.Value * 2 })
    .GroupBy(x => x.item.Category)
    .OrderBy(g => g.Key)
    .Select(g => new { Category = g.Key, Count = g.Count() });
```

### Property and Indexer Keywords
```csharp
public class PropertyExample
{
    private int _value;

    // Auto-implemented property
    public string Name { get; set; }

    // Property with backing field
    public int Value
    {
        get { return _value; }
        set { _value = value; }
    }

    // Expression-bodied property (C# 6)
    public string DisplayName => $"Value: {Value}";

    // Init-only property (C# 9)
    public string Id { get; init; }

    // Required property (C# 11)
    public required string RequiredProperty { get; set; }

    // Indexer
    public int this[int index]
    {
        get { return array[index]; }
        set { array[index] = value; }
    }

    private int[] array = new int[10];
}
```

### Modern C# Keywords and Features

#### Pattern Matching (C# 7+)
```csharp
// Switch expressions (C# 8)
string GetDescription(object obj) => obj switch
{
    int i when i > 0 => "Positive integer",
    int i when i < 0 => "Negative integer",
    int => "Zero",
    string s when !string.IsNullOrEmpty(s) => $"String: {s}",
    string => "Empty string",
    null => "Null value",
    _ => "Unknown type"
};

// Property patterns (C# 8)
public static string DescribePerson(Person person) => person switch
{
    { Age: < 18 } => "Minor",
    { Age: >= 18, Age: < 65 } => "Adult",
    { Age: >= 65 } => "Senior",
    _ => "Unknown"
};

// Relational patterns (C# 9)
string ClassifyGrade(int score) => score switch
{
    >= 90 => "A",
    >= 80 and < 90 => "B",
    >= 70 and < 80 => "C",
    < 70 => "F"
};

// Logical patterns (C# 9)
bool IsLetter(char c) => c is >= 'A' and <= 'Z' or >= 'a' and <= 'z';
```

#### Records (C# 9+)
```csharp
// Record declaration
public record Person(string FirstName, string LastName);

// Record with additional members
public record Employee(string FirstName, string LastName, decimal Salary)
{
    public string FullName => $"{FirstName} {LastName}";

    // with expression usage
    public Employee GiveRaise(decimal amount) => this with { Salary = Salary + amount };
}

// Record struct (C# 10)
public readonly record struct Point(int X, int Y);

// Record class (explicit in C# 10)
public record class Manager(string FirstName, string LastName, decimal Salary, int TeamSize)
    : Employee(FirstName, LastName, Salary);
```

#### File-scoped Types and Global Using (C# 10+)
```csharp
// File-scoped namespace (C# 10)
namespace MyProject.Models;

// Global using (in GlobalUsings.cs)
global using System;
global using System.Collections.Generic;
global using System.Linq;

// File-scoped types (C# 11)
file class FileLocalClass
{
    // Only accessible within this file
}
```

#### Raw String Literals and UTF-8 Strings (C# 11)
```csharp
// Raw string literals
string json = """
{
    "name": "John",
    "age": 30,
    "city": "New York"
}
""";

// UTF-8 string literals
ReadOnlySpan<byte> utf8Text = "Hello World"u8;
```

## Operators

### Arithmetic Operators
```csharp
int a = 10, b = 3;

int addition = a + b;        // 13
int subtraction = a - b;     // 7
int multiplication = a * b;  // 30
int division = a / b;        // 3
int remainder = a % b;       // 1

// Unary operators
int positive = +a;           // 10
int negative = -a;           // -10
int preIncrement = ++a;      // 11 (a becomes 11)
int postIncrement = a++;     // 11 (returns 11, a becomes 12)
int preDecrement = --b;      // 2 (b becomes 2)
int postDecrement = b--;     // 2 (returns 2, b becomes 1)
```

### Assignment Operators
```csharp
int x = 10;
x += 5;  // x = x + 5; (15)
x -= 3;  // x = x - 3; (12)
x *= 2;  // x = x * 2; (24)
x /= 4;  // x = x / 4; (6)
x %= 5;  // x = x % 5; (1)

int y = 8;
y &= 3;  // y = y & 3; (0)
y |= 4;  // y = y | 4; (4)
y ^= 2;  // y = y ^ 2; (6)
y <<= 1; // y = y << 1; (12)
y >>= 2; // y = y >> 2; (3)

// Null-coalescing assignment (C# 8)
string name = null;
name ??= "Default"; // Assigns "Default" if name is null
```

### Comparison Operators
```csharp
bool equal = (a == b);
bool notEqual = (a != b);
bool lessThan = (a < b);
bool lessOrEqual = (a <= b);
bool greaterThan = (a > b);
bool greaterOrEqual = (a >= b);

// Reference equality
object obj1 = new object();
object obj2 = obj1;
bool sameReference = ReferenceEquals(obj1, obj2); // true
bool sameReference2 = (obj1 is obj2); // Pattern matching (C# 7)
```

### Logical Operators
```csharp
bool a = true, b = false;

bool and = a && b;      // false (short-circuit evaluation)
bool or = a || b;       // true (short-circuit evaluation)
bool not = !a;          // false
bool xor = a ^ b;       // true

// Bitwise logical operators (non-short-circuit)
bool bitwiseAnd = a & b;    // false
bool bitwiseOr = a | b;     // true
```

### Bitwise Operators
```csharp
int x = 12;  // 1100 in binary
int y = 7;   // 0111 in binary

int bitwiseAnd = x & y;     // 4 (0100)
int bitwiseOr = x | y;      // 15 (1111)
int bitwiseXor = x ^ y;     // 11 (1011)
int bitwiseNot = ~x;        // -13 (two's complement)
int leftShift = x << 2;     // 48 (110000)
int rightShift = x >> 2;    // 3 (0011)

// Unsigned right shift (C# 11)
uint unsignedRightShift = (uint)x >>> 2; // Logical right shift
```

### Type Testing and Conversion
```csharp
object obj = "Hello";

// Type testing
bool isString = obj is string;
bool isInt = obj is int;

// Pattern matching with is
if (obj is string str && str.Length > 0)
{
    Console.WriteLine($"Non-empty string: {str}");
}

// Type casting
string text = (string)obj;           // Explicit cast
string text2 = obj as string;        // Safe cast (null if fails)

// Type information
Type type = obj.GetType();
Type stringType = typeof(string);

// sizeof operator (compile-time constant for value types)
int intSize = sizeof(int);           // 4 bytes
```

### Null-related Operators
```csharp
string text = null;

// Null-coalescing operator
string result = text ?? "default";   // "default"

// Null-coalescing assignment (C# 8)
text ??= "assigned";                 // Assigns if null

// Null-conditional operators (C# 6)
int? length = text?.Length;          // null if text is null
char? firstChar = text?[0];          // null if text is null
string upper = text?.ToUpper();      // null if text is null

// Null-forgiving operator (C# 8) - suppresses nullable warnings
string definitelyNotNull = text!;   // Tells compiler text is not null
```

### Range and Index Operators (C# 8)
```csharp
string text = "Hello World";
int[] numbers = { 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 };

// Index from end
char lastChar = text[^1];            // 'd' (last character)
char secondLast = text[^2];          // 'l'

// Ranges
string substring = text[1..5];       // "ello"
string fromStart = text[..5];        // "Hello"
string toEnd = text[6..];            // "World"
string allExceptEnds = text[1..^1];  // "ello Worl"

int[] slice = numbers[2..8];         // { 3, 4, 5, 6, 7, 8 }
int[] lastThree = numbers[^3..];     // { 8, 9, 10 }
```

### Lambda and Expression Operators
```csharp
// Lambda expressions (C# 3)
Func<int, int> square = x => x * x;
Func<int, int, int> add = (x, y) => x + y;
Action<string> print = s => Console.WriteLine(s);

// Expression-bodied members (C# 6)
public int DoubleValue => value * 2;
public void PrintValue() => Console.WriteLine(value);

// Local functions (C# 7)
int CalculateSum(int n)
{
    return Sum(n);

    int Sum(int num) => num <= 1 ? num : num + Sum(num - 1);
}

// Switch expressions (C# 8)
string GetSeason(int month) => month switch
{
    12 or 1 or 2 => "Winter",
    3 or 4 or 5 => "Spring",
    6 or 7 or 8 => "Summer",
    9 or 10 or 11 => "Fall",
    _ => throw new ArgumentException("Invalid month")
};
```

## Special Syntax Constructs

### Generics
```csharp
// Generic classes
public class Container<T> where T : class, new()
{
    private T item = new T();

    public void SetItem(T newItem) => item = newItem;
    public T GetItem() => item;
}

// Generic methods
public T CreateInstance<T>() where T : new() => new T();

// Generic constraints
public class Repository<T> where T : class, IEntity, new()
{
    // Implementation
}

// Covariance and contravariance (C# 4)
public interface ICovariant<out T> { T GetItem(); }
public interface IContravariant<in T> { void SetItem(T item); }
```

### Async/Await (C# 5)
```csharp
// Async method
public async Task<string> GetDataAsync()
{
    using var client = new HttpClient();
    string result = await client.GetStringAsync("https://api.example.com/data");
    return result;
}

// Async enumerable (C# 8)
public async IAsyncEnumerable<int> GenerateNumbersAsync()
{
    for (int i = 0; i < 10; i++)
    {
        await Task.Delay(100);
        yield return i;
    }
}

// Await foreach (C# 8)
await foreach (var number in GenerateNumbersAsync())
{
    Console.WriteLine(number);
}
```

### Anonymous Types and Methods
```csharp
// Anonymous types (C# 3)
var person = new { Name = "John", Age = 30, City = "New York" };

// Anonymous methods (C# 2)
Button.Click += delegate(object sender, EventArgs e)
{
    MessageBox.Show("Clicked!");
};

// Improved with lambdas (C# 3)
Button.Click += (sender, e) => MessageBox.Show("Clicked!");
```

### Tuple Types (C# 7)
```csharp
// Tuple literal
var point = (X: 10, Y: 20);

// Tuple return type
public (string First, string Last) GetName()
{
    return ("John", "Doe");
}

// Tuple deconstruction
var (first, last) = GetName();

// Discards in deconstruction
var (firstName, _) = GetName(); // Ignore last name
```

### Local Functions (C# 7)
```csharp
public int CalculateFactorial(int n)
{
    if (n < 0) throw new ArgumentException("Negative numbers not allowed");

    return LocalFactorial(n);

    int LocalFactorial(int num)
    {
        return num <= 1 ? 1 : num * LocalFactorial(num - 1);
    }
}
```

### Throw Expressions (C# 7)
```csharp
// Throw as expression
string name = input ?? throw new ArgumentNullException(nameof(input));

// In conditional operator
string result = condition ? "valid" : throw new InvalidOperationException();

// In expression-bodied members
public string Name => _name ?? throw new InvalidOperationException("Name not set");
```

### Default Expressions (C# 7.1)
```csharp
// Default literal
int number = default;           // 0
string text = default;          // null
bool flag = default;            // false

// In method calls
ProcessValue(default);          // Type inferred from method signature

// Generic default
T GetDefault<T>() => default(T);
```

### Ref Returns and Locals (C# 7)
```csharp
public class RefExample
{
    private int[] numbers = { 1, 2, 3, 4, 5 };

    // Ref return
    public ref int GetElement(int index)
    {
        return ref numbers[index];
    }

    public void UseRefReturn()
    {
        ref int element = ref GetElement(2);
        element = 100; // Modifies the original array element
    }
}
```

### In Parameters (C# 7.2)
```csharp
// Read-only reference parameter
public void ProcessLargeStruct(in LargeStruct data)
{
    // data is passed by reference but cannot be modified
    Console.WriteLine(data.SomeProperty);
}
```

### Stackalloc in More Places (C# 7.2)
```csharp
// Stackalloc in safe context
Span<int> numbers = stackalloc int[] { 1, 2, 3, 4, 5 };

// Stackalloc with var
var moreNumbers = stackalloc int[] { 6, 7, 8, 9, 10 };
```

This comprehensive reference covers all major C# keywords and syntax constructs, providing a complete foundation for understanding and using the C# programming language.