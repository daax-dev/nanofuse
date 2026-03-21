# C# Principles, Idioms, and Philosophy

## Language Design Philosophy

### Core Design Tenets

#### 1. **Type Safety with Flexibility**
C# prioritizes compile-time type safety while providing mechanisms for runtime flexibility when needed.

```csharp
// Strong typing prevents many runtime errors
string name = "John";
int age = 30;
// name = age; // Compile error - type safety

// But allows flexibility when needed
dynamic flexibleVar = GetSomeValue();
object boxedValue = 42; // Boxing for polymorphism
```

#### 2. **Managed Memory with Performance**
Automatic memory management through garbage collection, with escape hatches for performance-critical scenarios.

```csharp
// Automatic memory management
var list = new List<int> { 1, 2, 3 }; // GC handles cleanup

// Performance-critical scenarios
unsafe
{
    int* ptr = stackalloc int[1000]; // Stack allocation
    // Manual memory control when needed
}

// Value types for performance
struct Point { public int X, Y; } // No heap allocation
```

#### 3. **Object-Oriented with Functional Elements**
Primarily object-oriented with increasing support for functional programming paradigms.

```csharp
// Object-oriented foundation
public class BankAccount
{
    public decimal Balance { get; private set; }
    public void Deposit(decimal amount) => Balance += amount;
}

// Functional elements
var numbers = Enumerable.Range(1, 10)
    .Where(x => x % 2 == 0)
    .Select(x => x * x)
    .Sum();

// Expression-bodied members
public string FullName => $"{FirstName} {LastName}";
```

## Core Principles

### 1. **Explicitness Over Implicitness**
C# favors explicit code that clearly expresses intent, while providing convenience features judiciously.

```csharp
// Explicit intent
public class OrderProcessor : IOrderProcessor
{
    private readonly ILogger _logger;

    public OrderProcessor(ILogger logger)
    {
        _logger = logger ?? throw new ArgumentNullException(nameof(logger));
    }
}

// Convenient but still explicit
var processor = new OrderProcessor(logger); // var for obvious types
string message = $"Processing order {orderId}"; // string interpolation
```

### 2. **Null Safety and Defensive Programming**
Progressive enhancement of null safety while maintaining backward compatibility.

```csharp
// Traditional null checking
public string GetDisplayName(Person? person)
{
    if (person == null)
        return "Unknown";

    return person.Name ?? "No Name";
}

// Modern null safety (C# 8+)
public string GetDisplayName(Person? person) => person?.Name ?? "Unknown";

// Nullable reference types
#nullable enable
public class Person
{
    public string Name { get; set; } = string.Empty; // Non-nullable
    public string? NickName { get; set; } // Nullable
}
```

### 3. **Performance Through Design**
Language features designed to minimize overhead and maximize performance.

```csharp
// Struct for value semantics and performance
public readonly struct Money
{
    public decimal Amount { get; }
    public string Currency { get; }

    public Money(decimal amount, string currency)
    {
        Amount = amount;
        Currency = currency;
    }
}

// Span<T> for efficient memory operations
public void ProcessData(ReadOnlySpan<byte> data)
{
    // No allocations for slicing
    var header = data[..10];
    var body = data[10..];
}

// ref returns for avoiding copies
public ref readonly Money GetBestRate() => ref _rates[_bestIndex];
```

## Essential Idioms

### 1. **Property-Based Design**
Properties as the primary mechanism for encapsulation and data access.

```csharp
public class Product
{
    // Auto-implemented property
    public string Name { get; set; } = string.Empty;

    // Computed property
    public string DisplayName => $"{Name} ({Sku})";

    // Property with validation
    private decimal _price;
    public decimal Price
    {
        get => _price;
        set => _price = value >= 0 ? value : throw new ArgumentException("Price cannot be negative");
    }

    // Init-only property (C# 9+)
    public string Sku { get; init; } = string.Empty;
}
```

### 2. **LINQ as Query Language**
LINQ as the idiomatic way to work with collections and data.

```csharp
// Method syntax for complex transformations
var summary = orders
    .Where(o => o.Date >= startDate)
    .GroupBy(o => o.CustomerId)
    .Select(g => new CustomerSummary
    {
        CustomerId = g.Key,
        TotalOrders = g.Count(),
        TotalAmount = g.Sum(o => o.Amount),
        AverageOrderValue = g.Average(o => o.Amount)
    })
    .OrderByDescending(s => s.TotalAmount)
    .ToList();

// Query syntax for SQL-like operations
var expensiveItems = from item in inventory
                    where item.Price > 100
                    orderby item.Price descending
                    select new { item.Name, item.Price };
```

### 3. **Exception-Based Error Handling**
Structured exception handling as the primary error management strategy.

```csharp
public async Task<Order> CreateOrderAsync(OrderRequest request)
{
    try
    {
        ValidateRequest(request);
        var order = await _orderService.CreateAsync(request);
        await _eventBus.PublishAsync(new OrderCreated(order.Id));
        return order;
    }
    catch (ValidationException ex)
    {
        _logger.LogWarning(ex, "Invalid order request: {Request}", request);
        throw; // Re-throw to preserve stack trace
    }
    catch (Exception ex)
    {
        _logger.LogError(ex, "Failed to create order");
        throw new OrderCreationException("Unable to create order", ex);
    }
}
```

### 4. **Async/Await Pattern**
Asynchronous programming as a first-class language feature.

```csharp
public async Task<ApiResponse<T>> GetDataAsync<T>(string endpoint)
{
    using var client = _httpClientFactory.CreateClient();

    try
    {
        var response = await client.GetAsync(endpoint);
        response.EnsureSuccessStatusCode();

        var content = await response.Content.ReadAsStringAsync();
        var data = JsonSerializer.Deserialize<T>(content);

        return ApiResponse.Success(data);
    }
    catch (HttpRequestException ex)
    {
        return ApiResponse.Failure<T>($"HTTP error: {ex.Message}");
    }
    catch (TaskCanceledException ex)
    {
        return ApiResponse.Failure<T>("Request timeout");
    }
}
```

## SOLID Principles in C#

### 1. **Single Responsibility Principle**
Each class should have one reason to change.

```csharp
// Good: Single responsibility
public class EmailSender
{
    public async Task SendAsync(string to, string subject, string body)
    {
        // Only responsible for sending emails
    }
}

public class OrderNotificationService
{
    private readonly EmailSender _emailSender;

    public async Task NotifyOrderCreatedAsync(Order order)
    {
        var email = ComposeOrderEmail(order);
        await _emailSender.SendAsync(order.CustomerEmail, email.Subject, email.Body);
    }
}
```

### 2. **Open/Closed Principle**
Open for extension, closed for modification.

```csharp
public abstract class PaymentProcessor
{
    public async Task<PaymentResult> ProcessAsync(Payment payment)
    {
        var validation = ValidatePayment(payment);
        if (!validation.IsValid)
            return PaymentResult.Failed(validation.Errors);

        return await ProcessPaymentCore(payment);
    }

    protected virtual PaymentValidation ValidatePayment(Payment payment) => PaymentValidation.Valid;
    protected abstract Task<PaymentResult> ProcessPaymentCore(Payment payment);
}

public class CreditCardProcessor : PaymentProcessor
{
    protected override async Task<PaymentResult> ProcessPaymentCore(Payment payment)
    {
        // Credit card specific implementation
    }
}
```

### 3. **Liskov Substitution Principle**
Derived classes must be substitutable for their base classes.

```csharp
public abstract class Shape
{
    public abstract double CalculateArea();
    public virtual string GetDescription() => $"Area: {CalculateArea():F2}";
}

public class Rectangle : Shape
{
    public double Width { get; set; }
    public double Height { get; set; }

    public override double CalculateArea() => Width * Height;
}

public class Circle : Shape
{
    public double Radius { get; set; }

    public override double CalculateArea() => Math.PI * Radius * Radius;
}

// Any Shape can be used interchangeably
public void PrintShapeInfo(Shape shape) => Console.WriteLine(shape.GetDescription());
```

### 4. **Interface Segregation Principle**
Clients should not depend on interfaces they don't use.

```csharp
// Bad: Fat interface
public interface IWorker
{
    void Work();
    void Eat();
    void Sleep();
}

// Good: Segregated interfaces
public interface IWorkable
{
    void Work();
}

public interface IFeedable
{
    void Eat();
}

public interface ISleepable
{
    void Sleep();
}

public class Human : IWorkable, IFeedable, ISleepable
{
    public void Work() { /* Implementation */ }
    public void Eat() { /* Implementation */ }
    public void Sleep() { /* Implementation */ }
}

public class Robot : IWorkable
{
    public void Work() { /* Implementation */ }
    // Robots don't eat or sleep
}
```

### 5. **Dependency Inversion Principle**
Depend on abstractions, not concretions.

```csharp
// Abstraction
public interface IRepository<T>
{
    Task<T?> GetByIdAsync(int id);
    Task<IEnumerable<T>> GetAllAsync();
    Task AddAsync(T entity);
    Task UpdateAsync(T entity);
    Task DeleteAsync(int id);
}

// High-level module depends on abstraction
public class OrderService
{
    private readonly IRepository<Order> _orderRepository;
    private readonly IEventPublisher _eventPublisher;

    public OrderService(IRepository<Order> orderRepository, IEventPublisher eventPublisher)
    {
        _orderRepository = orderRepository;
        _eventPublisher = eventPublisher;
    }

    public async Task<Order> CreateOrderAsync(CreateOrderRequest request)
    {
        var order = new Order(request);
        await _orderRepository.AddAsync(order);
        await _eventPublisher.PublishAsync(new OrderCreated(order));
        return order;
    }
}
```

## Performance Principles

### 1. **Allocation Awareness**
Understanding and minimizing object allocations.

```csharp
// Avoid allocations in hot paths
public bool IsValidEmail(ReadOnlySpan<char> email)
{
    var atIndex = email.IndexOf('@');
    if (atIndex <= 0 || atIndex == email.Length - 1)
        return false;

    var domainPart = email[(atIndex + 1)..];
    return domainPart.IndexOf('.') > 0;
}

// Use object pooling for expensive objects
private readonly ObjectPool<StringBuilder> _stringBuilderPool;

public string FormatMessage(IEnumerable<string> parts)
{
    var sb = _stringBuilderPool.Get();
    try
    {
        foreach (var part in parts)
            sb.Append(part).Append(' ');
        return sb.ToString();
    }
    finally
    {
        _stringBuilderPool.Return(sb);
    }
}
```

### 2. **Efficient Data Structures**
Choosing appropriate collections and data structures.

```csharp
// Use appropriate collection types
private readonly Dictionary<string, Customer> _customerCache = new(); // O(1) lookup
private readonly HashSet<string> _validCodes = new(); // O(1) contains check
private readonly List<Order> _orders = new(); // Sequential access

// Consider memory layout for value types
[StructLayout(LayoutKind.Sequential, Pack = 1)]
public struct CompactPoint
{
    public short X;
    public short Y;
    // Packed for cache efficiency
}
```

### 3. **Lazy Initialization**
Deferring expensive operations until needed.

```csharp
public class ExpensiveResource
{
    private readonly Lazy<HeavyObject> _heavyObject = new(() => new HeavyObject());
    private readonly Lazy<Task<DatabaseConnection>> _connection = new(CreateConnectionAsync);

    public HeavyObject Heavy => _heavyObject.Value;
    public Task<DatabaseConnection> ConnectionAsync => _connection.Value;

    private static async Task<DatabaseConnection> CreateConnectionAsync()
    {
        var connection = new DatabaseConnection();
        await connection.OpenAsync();
        return connection;
    }
}
```

## Modern C# Idioms

### 1. **Pattern Matching**
Declarative matching for control flow and data extraction.

```csharp
public decimal CalculateDiscount(Customer customer, Order order) => customer switch
{
    { Type: CustomerType.Premium, YearsActive: > 5 } => order.Total * 0.15m,
    { Type: CustomerType.Premium } => order.Total * 0.10m,
    { Type: CustomerType.Regular, YearsActive: > 3 } => order.Total * 0.05m,
    _ => 0m
};

public string FormatValue(object value) => value switch
{
    null => "null",
    string s when string.IsNullOrEmpty(s) => "empty",
    string s => $"\"{s}\"",
    int i => i.ToString(),
    decimal d => d.ToString("C"),
    IEnumerable<object> items => $"[{string.Join(", ", items.Select(FormatValue))}]",
    _ => value.ToString() ?? "unknown"
};
```

### 2. **Records for Data**
Immutable data carriers with value semantics.

```csharp
public record Customer(int Id, string Name, string Email)
{
    public string DisplayName => $"{Name} <{Email}>";
}

public record OrderLine(string ProductId, int Quantity, decimal UnitPrice)
{
    public decimal Total => Quantity * UnitPrice;
}

public record Order(int Id, Customer Customer, IReadOnlyList<OrderLine> Lines)
{
    public decimal Total => Lines.Sum(line => line.Total);

    // With expressions for immutable updates
    public Order AddLine(OrderLine line) => this with
    {
        Lines = Lines.Append(line).ToList()
    };
}
```

### 3. **Local Functions**
Helper functions scoped to their usage context.

```csharp
public async Task<ProcessingResult> ProcessOrdersAsync(IEnumerable<Order> orders)
{
    var results = new List<OrderResult>();
    var errors = new List<string>();

    await foreach (var batch in CreateBatches(orders, batchSize: 10))
    {
        var batchResults = await ProcessBatch(batch);
        results.AddRange(batchResults.Successful);
        errors.AddRange(batchResults.Errors);
    }

    return new ProcessingResult(results, errors);

    // Local async iterator
    async IAsyncEnumerable<IEnumerable<Order>> CreateBatches(
        IEnumerable<Order> items,
        int batchSize)
    {
        var batch = new List<Order>(batchSize);
        foreach (var item in items)
        {
            batch.Add(item);
            if (batch.Count == batchSize)
            {
                yield return batch;
                batch = new List<Order>(batchSize);
                await Task.Yield(); // Allow other work
            }
        }
        if (batch.Count > 0)
            yield return batch;
    }

    // Local async method
    async Task<(List<OrderResult> Successful, List<string> Errors)> ProcessBatch(IEnumerable<Order> batch)
    {
        var successful = new List<OrderResult>();
        var batchErrors = new List<string>();

        await Task.WhenAll(batch.Select(async order =>
        {
            try
            {
                var result = await ProcessSingleOrder(order);
                lock (successful) successful.Add(result);
            }
            catch (Exception ex)
            {
                lock (batchErrors) batchErrors.Add($"Order {order.Id}: {ex.Message}");
            }
        }));

        return (successful, batchErrors);
    }
}
```

This comprehensive overview covers the fundamental principles, idioms, and philosophy that guide effective C# development, from language design concepts to practical coding patterns and performance considerations.