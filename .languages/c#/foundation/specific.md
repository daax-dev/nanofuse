# C# Unique Language Features

## Overview
This guide covers the distinctive features that make C# unique among programming languages, focusing on language constructs and capabilities that set it apart from other languages.

## Properties and Automatic Properties

### Traditional Properties
```csharp
public class Person
{
    private string _name;
    private int _age;

    // Full property with validation
    public string Name
    {
        get { return _name; }
        set
        {
            if (string.IsNullOrWhiteSpace(value))
                throw new ArgumentException("Name cannot be empty");
            _name = value;
        }
    }

    // Read-only property
    public int Age
    {
        get { return _age; }
        private set { _age = value; }
    }

    // Computed property
    public string FullInfo => $"{Name} ({Age} years old)";
}
```

### Auto-Implemented Properties
```csharp
public class Product
{
    // Auto-implemented properties (C# 3.0)
    public string Name { get; set; }
    public decimal Price { get; set; }

    // Auto-implemented with private setter
    public int Id { get; private set; }

    // Auto-implemented with initializer (C# 6.0)
    public DateTime CreatedAt { get; set; } = DateTime.Now;

    // Init-only properties (C# 9.0)
    public string Category { get; init; }

    // Required properties (C# 11.0)
    public required string Sku { get; set; }
}

// Usage
var product = new Product
{
    Name = "Widget",
    Price = 9.99m,
    Category = "Tools",
    Sku = "WDG001"
};
```

### Expression-Bodied Properties
```csharp
public class Rectangle
{
    public double Width { get; set; }
    public double Height { get; set; }

    // Expression-bodied get-only property
    public double Area => Width * Height;

    // Expression-bodied property with getter and setter (C# 7.0)
    private double _perimeter;
    public double Perimeter
    {
        get => _perimeter;
        set => _perimeter = value;
    }
}
```

## Events and Delegates

### Delegates
```csharp
// Delegate declaration
public delegate void EventHandler(string message);
public delegate T Func<T>(int input);
public delegate bool Predicate<T>(T item);

public class EventPublisher
{
    // Multicast delegate
    public EventHandler OnMessageReceived;

    // Built-in delegates
    public Action<string> OnLog;
    public Func<int, string> FormatNumber;

    public void TriggerEvent(string message)
    {
        // Null-conditional delegate invocation
        OnMessageReceived?.Invoke(message);
        OnLog?.Invoke($"Event triggered: {message}");
    }

    // Delegate as parameter
    public void ProcessItems(List<int> numbers, Predicate<int> filter)
    {
        foreach (int number in numbers)
        {
            if (filter(number))
            {
                Console.WriteLine(FormatNumber?.Invoke(number) ?? number.ToString());
            }
        }
    }
}

// Usage
var publisher = new EventPublisher();
publisher.OnMessageReceived += msg => Console.WriteLine($"Received: {msg}");
publisher.OnMessageReceived += msg => File.WriteAllText("log.txt", msg);
publisher.FormatNumber = num => $"Number: {num:N0}";
```

### Events
```csharp
public class BankAccount
{
    private decimal _balance;

    // Event declaration
    public event Action<decimal> BalanceChanged;
    public event Action<string> TransactionOccurred;

    public decimal Balance
    {
        get => _balance;
        private set
        {
            if (_balance != value)
            {
                _balance = value;
                // Raise event
                BalanceChanged?.Invoke(_balance);
            }
        }
    }

    public void Deposit(decimal amount)
    {
        if (amount <= 0) throw new ArgumentException("Amount must be positive");

        Balance += amount;
        TransactionOccurred?.Invoke($"Deposited: {amount:C}");
    }

    public void Withdraw(decimal amount)
    {
        if (amount <= 0) throw new ArgumentException("Amount must be positive");
        if (amount > Balance) throw new InvalidOperationException("Insufficient funds");

        Balance -= amount;
        TransactionOccurred?.Invoke($"Withdrawn: {amount:C}");
    }
}

// Usage
var account = new BankAccount();
account.BalanceChanged += balance => Console.WriteLine($"Balance: {balance:C}");
account.TransactionOccurred += transaction => Console.WriteLine($"Transaction: {transaction}");
```

## Extension Methods

### Basic Extension Methods
```csharp
public static class StringExtensions
{
    // Extension method for string
    public static bool IsValidEmail(this string email)
    {
        if (string.IsNullOrWhiteSpace(email))
            return false;

        return Regex.IsMatch(email, @"^[^@\s]+@[^@\s]+\.[^@\s]+$");
    }

    public static string Truncate(this string input, int maxLength)
    {
        if (string.IsNullOrEmpty(input) || input.Length <= maxLength)
            return input;

        return input.Substring(0, maxLength) + "...";
    }

    public static string ToTitleCase(this string input)
    {
        if (string.IsNullOrEmpty(input))
            return input;

        return CultureInfo.CurrentCulture.TextInfo.ToTitleCase(input.ToLower());
    }
}

// Usage
string email = "user@example.com";
bool isValid = email.IsValidEmail();           // Extension method call
string truncated = "Long text here".Truncate(10);
```

### Generic Extension Methods
```csharp
public static class EnumerableExtensions
{
    // Extension method for IEnumerable<T>
    public static IEnumerable<T> WhereNotNull<T>(this IEnumerable<T?> source)
        where T : class
    {
        return source.Where(item => item != null)!;
    }

    public static void ForEach<T>(this IEnumerable<T> source, Action<T> action)
    {
        foreach (T item in source)
        {
            action(item);
        }
    }

    public static IEnumerable<T> Distinct<T, TKey>(this IEnumerable<T> source, Func<T, TKey> keySelector)
    {
        return source.GroupBy(keySelector).Select(g => g.First());
    }

    // Batching extension
    public static IEnumerable<IEnumerable<T>> Batch<T>(this IEnumerable<T> source, int batchSize)
    {
        using var enumerator = source.GetEnumerator();
        while (enumerator.MoveNext())
        {
            yield return GetBatch(enumerator, batchSize);
        }
    }

    private static IEnumerable<T> GetBatch<T>(IEnumerator<T> enumerator, int batchSize)
    {
        do
        {
            yield return enumerator.Current;
        }
        while (--batchSize > 0 && enumerator.MoveNext());
    }
}

// Usage
var numbers = new int?[] { 1, null, 2, null, 3 };
var validNumbers = numbers.WhereNotNull().ToList(); // { 1, 2, 3 }

var items = Enumerable.Range(1, 10);
items.ForEach(Console.WriteLine);

var batches = items.Batch(3).ToList(); // { {1,2,3}, {4,5,6}, {7,8,9}, {10} }
```

## Partial Classes and Methods

### Partial Classes
```csharp
// File: Person.Core.cs
public partial class Person
{
    public string FirstName { get; set; }
    public string LastName { get; set; }

    partial void OnNameChanged();

    public string FullName
    {
        get => $"{FirstName} {LastName}";
        set
        {
            var parts = value.Split(' ', 2);
            FirstName = parts[0];
            LastName = parts.Length > 1 ? parts[1] : "";
            OnNameChanged();
        }
    }
}

// File: Person.Validation.cs
public partial class Person
{
    partial void OnNameChanged()
    {
        ValidateName();
    }

    private void ValidateName()
    {
        if (string.IsNullOrWhiteSpace(FirstName))
            throw new ArgumentException("First name is required");
    }
}

// File: Person.Extensions.cs
public partial class Person
{
    public string GetInitials() => $"{FirstName?.FirstOrDefault()}{LastName?.FirstOrDefault()}";

    public override string ToString() => FullName;
}
```

### Partial Methods
```csharp
// Generated code file
public partial class DataContext
{
    partial void OnUserInserted(User user);
    partial void OnUserUpdated(User oldUser, User newUser);

    public void InsertUser(User user)
    {
        // Database insert logic
        SaveToDatabase(user);

        // Call partial method (may or may not be implemented)
        OnUserInserted(user);
    }
}

// User implementation file
public partial class DataContext
{
    partial void OnUserInserted(User user)
    {
        // Custom logic after user insertion
        LogUserActivity($"User {user.Name} was inserted");
        SendWelcomeEmail(user.Email);
    }
}
```

## Anonymous Types and Methods

### Anonymous Types
```csharp
public void DemonstrateAnonymousTypes()
{
    // Anonymous type creation
    var person = new { Name = "John", Age = 30, City = "New York" };
    var product = new { Id = 1, Name = "Widget", Price = 9.99m, InStock = true };

    // Anonymous types in LINQ projections
    var customerSummary = customers
        .Where(c => c.IsActive)
        .Select(c => new
        {
            c.Id,
            c.Name,
            OrderCount = c.Orders.Count(),
            TotalSpent = c.Orders.Sum(o => o.Total),
            LastOrder = c.Orders.Max(o => o.Date)
        })
        .ToList();

    // Anonymous types with computed properties
    var report = sales
        .GroupBy(s => s.Date.Year)
        .Select(g => new
        {
            Year = g.Key,
            TotalSales = g.Sum(s => s.Amount),
            AverageSale = g.Average(s => s.Amount),
            SalesCount = g.Count(),
            TopProduct = g.GroupBy(s => s.ProductName)
                          .OrderByDescending(pg => pg.Sum(s => s.Amount))
                          .First().Key
        })
        .ToList();
}
```

### Anonymous Methods
```csharp
public void DemonstrateAnonymousMethods()
{
    // Anonymous method (C# 2.0)
    Button.Click += delegate(object sender, EventArgs e)
    {
        MessageBox.Show("Button clicked!");
    };

    // Lambda expressions (C# 3.0) - more concise
    Button.Click += (sender, e) => MessageBox.Show("Button clicked!");

    // Anonymous methods with captures
    string message = "Hello World";
    Button.Click += delegate { MessageBox.Show(message); };

    // Complex anonymous methods
    Func<int, int, int> calculator = delegate(int x, int y)
    {
        if (x < 0 || y < 0)
            throw new ArgumentException("Negative numbers not allowed");
        return x + y;
    };

    // Equivalent lambda
    Func<int, int, int> calculatorLambda = (x, y) =>
    {
        if (x < 0 || y < 0)
            throw new ArgumentException("Negative numbers not allowed");
        return x + y;
    };
}
```

## Iterators and Yield

### Basic Iterators
```csharp
public static class NumberGenerators
{
    // Simple iterator
    public static IEnumerable<int> GetNumbers(int start, int count)
    {
        for (int i = 0; i < count; i++)
        {
            yield return start + i;
        }
    }

    // Infinite iterator
    public static IEnumerable<int> GetFibonacci()
    {
        int a = 0, b = 1;
        while (true)
        {
            yield return a;
            (a, b) = (b, a + b);
        }
    }

    // Iterator with conditions
    public static IEnumerable<T> Filter<T>(IEnumerable<T> source, Predicate<T> predicate)
    {
        foreach (T item in source)
        {
            if (predicate(item))
                yield return item;
        }
    }

    // Iterator that can yield break
    public static IEnumerable<string> ReadLines(string filePath)
    {
        if (!File.Exists(filePath))
            yield break;

        using var reader = new StreamReader(filePath);
        string line;
        while ((line = reader.ReadLine()) != null)
        {
            if (line.StartsWith("#")) // Skip comments
                continue;
            yield return line;
        }
    }
}

// Usage
foreach (int number in NumberGenerators.GetNumbers(10, 5))
{
    Console.WriteLine(number); // 10, 11, 12, 13, 14
}

var fibonacci = NumberGenerators.GetFibonacci()
    .Take(10)
    .ToList(); // First 10 Fibonacci numbers
```

### Advanced Iterator Patterns
```csharp
public class TreeNode<T>
{
    public T Value { get; set; }
    public List<TreeNode<T>> Children { get; set; } = new List<TreeNode<T>>();

    // Depth-first traversal iterator
    public IEnumerable<T> DepthFirst()
    {
        yield return Value;

        foreach (var child in Children)
        {
            foreach (var descendant in child.DepthFirst())
            {
                yield return descendant;
            }
        }
    }

    // Breadth-first traversal iterator
    public IEnumerable<T> BreadthFirst()
    {
        var queue = new Queue<TreeNode<T>>();
        queue.Enqueue(this);

        while (queue.Count > 0)
        {
            var node = queue.Dequeue();
            yield return node.Value;

            foreach (var child in node.Children)
            {
                queue.Enqueue(child);
            }
        }
    }
}
```

## Pattern Matching and Switch Expressions

### Pattern Matching Evolution
```csharp
public class PatternMatchingExamples
{
    // Traditional approach
    public string DescribeObjectOld(object obj)
    {
        if (obj is int)
        {
            return "Integer: " + obj;
        }
        else if (obj is string s && s.Length > 0)
        {
            return "Non-empty string: " + s;
        }
        else if (obj is List<int> list)
        {
            return $"Integer list with {list.Count} items";
        }
        return "Unknown";
    }

    // Modern pattern matching (C# 8+)
    public string DescribeObject(object obj) => obj switch
    {
        int i when i > 0 => $"Positive integer: {i}",
        int i when i < 0 => $"Negative integer: {i}",
        int => "Zero",
        string s when !string.IsNullOrEmpty(s) => $"Non-empty string: {s}",
        string => "Empty string",
        List<int> list => $"Integer list with {list.Count} items",
        null => "Null value",
        _ => "Unknown type"
    };

    // Property patterns
    public string ClassifyPerson(Person person) => person switch
    {
        { Age: < 13 } => "Child",
        { Age: >= 13, Age: < 20 } => "Teenager",
        { Age: >= 20, Age: < 65 } => "Adult",
        { Age: >= 65 } => "Senior",
        _ => "Unknown age"
    };

    // Relational patterns (C# 9)
    public string GetGrade(int score) => score switch
    {
        >= 90 => "A",
        >= 80 and < 90 => "B",
        >= 70 and < 80 => "C",
        >= 60 and < 70 => "D",
        < 60 => "F"
    };

    // Logical patterns (C# 9)
    public bool IsValidAge(int age) => age is >= 0 and <= 150;

    public bool IsVowel(char c) => c is 'a' or 'e' or 'i' or 'o' or 'u' or
                                           'A' or 'E' or 'I' or 'O' or 'U';
}
```

### Advanced Pattern Matching
```csharp
public record Point(int X, int Y);
public record Circle(Point Center, int Radius);
public record Rectangle(Point TopLeft, int Width, int Height);

public class GeometryPatterns
{
    // Deconstruction patterns
    public string DescribePoint(Point point) => point switch
    {
        (0, 0) => "Origin",
        (var x, 0) => $"On X-axis at {x}",
        (0, var y) => $"On Y-axis at {y}",
        (var x, var y) when x == y => $"On diagonal at ({x}, {y})",
        (var x, var y) => $"Point at ({x}, {y})"
    };

    // Nested patterns
    public double CalculateArea(object shape) => shape switch
    {
        Circle { Radius: var r } => Math.PI * r * r,
        Rectangle { Width: var w, Height: var h } => w * h,
        Circle { Center: { X: 0, Y: 0 }, Radius: var r } => Math.PI * r * r, // Nested
        _ => 0
    };

    // List patterns (C# 11)
    public string DescribeList<T>(IList<T> list) => list switch
    {
        [] => "Empty list",
        [var single] => $"Single item: {single}",
        [var first, var second] => $"Two items: {first}, {second}",
        [var first, .., var last] => $"Multiple items, first: {first}, last: {last}",
        _ => "Unknown pattern"
    };
}
```

## Records and Init-Only Properties

### Record Types
```csharp
// Basic record (C# 9)
public record Person(string FirstName, string LastName);

// Record with additional members
public record Employee(string FirstName, string LastName, decimal Salary)
{
    public string FullName => $"{FirstName} {LastName}";
    public string Department { get; init; } = "Unassigned";

    // Custom ToString override
    public override string ToString() => $"{FullName} ({Department}) - {Salary:C}";
}

// Record inheritance
public record Manager(string FirstName, string LastName, decimal Salary, int TeamSize)
    : Employee(FirstName, LastName, Salary)
{
    public string Department { get; init; } = "Management";
}

// Usage
var employee = new Employee("John", "Doe", 50000) { Department = "IT" };
var manager = employee with { Salary = 75000, TeamSize = 5 }; // Creates new instance

// Record equality
var emp1 = new Employee("Jane", "Smith", 60000);
var emp2 = new Employee("Jane", "Smith", 60000);
Console.WriteLine(emp1 == emp2); // True - value equality

// Deconstruction
var (first, last, salary) = employee;
```

### Class Records and Struct Records
```csharp
// Record class (explicit)
public record class PersonClass(string Name, int Age);

// Record struct (C# 10)
public readonly record struct Point(int X, int Y)
{
    public double DistanceFromOrigin => Math.Sqrt(X * X + Y * Y);
}

// Mutable record struct
public record struct MutablePoint(int X, int Y)
{
    public void Move(int deltaX, int deltaY)
    {
        X += deltaX;
        Y += deltaY;
    }
}
```

## Local Functions

### Basic Local Functions
```csharp
public int CalculateFactorial(int n)
{
    if (n < 0)
        throw new ArgumentException("Negative numbers not supported");

    return Factorial(n);

    // Local function
    int Factorial(int num)
    {
        return num <= 1 ? 1 : num * Factorial(num - 1);
    }
}

public async Task<string> ProcessDataAsync(string input)
{
    if (string.IsNullOrEmpty(input))
        throw new ArgumentException("Input cannot be empty");

    var processed = await ProcessInternalAsync(input);
    return FormatResult(processed);

    // Async local function
    async Task<string> ProcessInternalAsync(string data)
    {
        await Task.Delay(100); // Simulate async work
        return data.ToUpper();
    }

    // Regular local function
    string FormatResult(string data)
    {
        return $"Processed: {data}";
    }
}
```

### Advanced Local Function Patterns
```csharp
public IEnumerable<T> Filter<T>(IEnumerable<T> source, Func<T, bool> predicate)
{
    if (source == null) throw new ArgumentNullException(nameof(source));
    if (predicate == null) throw new ArgumentNullException(nameof(predicate));

    return FilterIterator();

    // Local iterator function
    IEnumerable<T> FilterIterator()
    {
        foreach (T item in source)
        {
            if (predicate(item))
                yield return item;
        }
    }
}

public void ProcessWithLocalFunctions()
{
    var numbers = new[] { 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 };

    // Local function with closure
    bool IsInRange(int value) => value >= minValue && value <= maxValue;
    int Square(int value) => value * value;

    var minValue = 3;
    var maxValue = 8;

    var result = numbers
        .Where(IsInRange)
        .Select(Square)
        .ToArray();

    Console.WriteLine(string.Join(", ", result)); // 9, 16, 25, 36, 49, 64
}
```

## Expression Trees

### Basic Expression Trees
```csharp
using System.Linq.Expressions;

public class ExpressionTreeExamples
{
    public void BasicExpressionTrees()
    {
        // Lambda to expression tree
        Expression<Func<int, bool>> isEven = x => x % 2 == 0;

        // Compile and execute
        Func<int, bool> compiled = isEven.Compile();
        bool result = compiled(4); // true

        // Examine expression structure
        Console.WriteLine(isEven.Body);          // ((x % 2) == 0)
        Console.WriteLine(isEven.Parameters[0]); // x
    }

    public void BuildExpressionManually()
    {
        // Build expression: x => x * x + 1
        ParameterExpression x = Expression.Parameter(typeof(int), "x");
        Expression multiply = Expression.Multiply(x, x);
        Expression constant = Expression.Constant(1);
        Expression add = Expression.Add(multiply, constant);
        Expression<Func<int, int>> lambda = Expression.Lambda<Func<int, int>>(add, x);

        Func<int, int> compiled = lambda.Compile();
        int result = compiled(3); // 10
    }

    // Dynamic LINQ example
    public IQueryable<T> ApplyDynamicFilter<T>(IQueryable<T> query, string propertyName, object value)
    {
        ParameterExpression parameter = Expression.Parameter(typeof(T), "x");
        MemberExpression property = Expression.Property(parameter, propertyName);
        ConstantExpression constant = Expression.Constant(value);
        BinaryExpression equal = Expression.Equal(property, constant);
        Expression<Func<T, bool>> lambda = Expression.Lambda<Func<T, bool>>(equal, parameter);

        return query.Where(lambda);
    }
}
```

## Attributes and Reflection

### Custom Attributes
```csharp
// Attribute definition
[AttributeUsage(AttributeTargets.Property | AttributeTargets.Field,
                AllowMultiple = false, Inherited = true)]
public class ValidationAttribute : Attribute
{
    public string ErrorMessage { get; set; }
    public bool IsRequired { get; set; }

    public ValidationAttribute(string errorMessage = null)
    {
        ErrorMessage = errorMessage;
    }
}

[AttributeUsage(AttributeTargets.Class)]
public class TableAttribute : Attribute
{
    public string Name { get; }
    public string Schema { get; set; } = "dbo";

    public TableAttribute(string name)
    {
        Name = name;
    }
}

// Usage
[Table("Users", Schema = "Identity")]
public class User
{
    [Validation("User ID is required", IsRequired = true)]
    public int Id { get; set; }

    [Validation("Name cannot be empty", IsRequired = true)]
    public string Name { get; set; }

    [Validation]
    public string Email { get; set; }
}
```

### Reflection Usage
```csharp
public class ReflectionValidator
{
    public List<string> Validate<T>(T instance)
    {
        var errors = new List<string>();
        var type = typeof(T);

        // Get table attribute
        var tableAttribute = type.GetCustomAttribute<TableAttribute>();
        if (tableAttribute != null)
        {
            Console.WriteLine($"Validating table: {tableAttribute.Schema}.{tableAttribute.Name}");
        }

        // Validate properties
        foreach (var property in type.GetProperties())
        {
            var validationAttribute = property.GetCustomAttribute<ValidationAttribute>();
            if (validationAttribute != null)
            {
                var value = property.GetValue(instance);

                if (validationAttribute.IsRequired && (value == null ||
                    (value is string str && string.IsNullOrWhiteSpace(str))))
                {
                    errors.Add(validationAttribute.ErrorMessage ??
                              $"{property.Name} is required");
                }
            }
        }

        return errors;
    }
}

// Usage
var user = new User { Id = 1, Name = "", Email = "test@example.com" };
var validator = new ReflectionValidator();
var errors = validator.Validate(user); // ["Name cannot be empty"]
```

## Operator Overloading

### Basic Operator Overloading
```csharp
public struct Complex
{
    public double Real { get; }
    public double Imaginary { get; }

    public Complex(double real, double imaginary = 0)
    {
        Real = real;
        Imaginary = imaginary;
    }

    // Arithmetic operators
    public static Complex operator +(Complex a, Complex b)
        => new Complex(a.Real + b.Real, a.Imaginary + b.Imaginary);

    public static Complex operator -(Complex a, Complex b)
        => new Complex(a.Real - b.Real, a.Imaginary - b.Imaginary);

    public static Complex operator *(Complex a, Complex b)
        => new Complex(
            a.Real * b.Real - a.Imaginary * b.Imaginary,
            a.Real * b.Imaginary + a.Imaginary * b.Real);

    // Unary operators
    public static Complex operator -(Complex a)
        => new Complex(-a.Real, -a.Imaginary);

    public static Complex operator +(Complex a) => a;

    // Comparison operators
    public static bool operator ==(Complex a, Complex b)
        => a.Real == b.Real && a.Imaginary == b.Imaginary;

    public static bool operator !=(Complex a, Complex b) => !(a == b);

    // Implicit conversion
    public static implicit operator Complex(double real) => new Complex(real);

    // Explicit conversion
    public static explicit operator double(Complex complex) => complex.Real;

    public override string ToString()
        => Imaginary >= 0 ? $"{Real} + {Imaginary}i" : $"{Real} - {-Imaginary}i";

    public override bool Equals(object obj) => obj is Complex c && this == c;
    public override int GetHashCode() => HashCode.Combine(Real, Imaginary);
}

// Usage
Complex a = 3;           // Implicit conversion
Complex b = new Complex(2, 4);
Complex sum = a + b;     // 5 + 4i
double real = (double)sum; // Explicit conversion to 5
```

## Indexers

### Basic and Advanced Indexers
```csharp
public class Matrix<T>
{
    private readonly T[,] _data;

    public int Rows { get; }
    public int Columns { get; }

    public Matrix(int rows, int columns)
    {
        Rows = rows;
        Columns = columns;
        _data = new T[rows, columns];
    }

    // Basic indexer
    public T this[int row, int column]
    {
        get
        {
            ValidateIndices(row, column);
            return _data[row, column];
        }
        set
        {
            ValidateIndices(row, column);
            _data[row, column] = value;
        }
    }

    // String indexer for named access
    public T this[string position]
    {
        get
        {
            var (row, col) = ParsePosition(position);
            return this[row, col];
        }
        set
        {
            var (row, col) = ParsePosition(position);
            this[row, col] = value;
        }
    }

    // Range indexer (C# 8+)
    public T[] this[int row, Range columnRange]
    {
        get
        {
            ValidateRow(row);
            var (start, length) = columnRange.GetOffsetAndLength(Columns);
            var result = new T[length];
            for (int i = 0; i < length; i++)
            {
                result[i] = _data[row, start + i];
            }
            return result;
        }
    }

    private void ValidateIndices(int row, int column)
    {
        if (row < 0 || row >= Rows)
            throw new IndexOutOfRangeException($"Row {row} out of range");
        if (column < 0 || column >= Columns)
            throw new IndexOutOfRangeException($"Column {column} out of range");
    }

    private (int row, int col) ParsePosition(string position)
    {
        // Parse "A1", "B2" style positions
        if (position.Length < 2)
            throw new ArgumentException("Invalid position format");

        int col = char.ToUpper(position[0]) - 'A';
        int row = int.Parse(position.Substring(1)) - 1;

        return (row, col);
    }
}

// Usage
var matrix = new Matrix<int>(3, 3);
matrix[0, 0] = 1;           // Basic indexer
matrix["A1"] = 1;           // String indexer
var row = matrix[0, 1..3];  // Range indexer
```

This comprehensive guide covers the unique features that distinguish C# from other programming languages, providing both theoretical understanding and practical implementation examples.