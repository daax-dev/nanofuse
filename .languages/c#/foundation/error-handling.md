# C# Error Handling - Comprehensive Guide

## Table of Contents

1. [Exception Handling Fundamentals](#exception-handling-fundamentals)
2. [Exception Hierarchy and Types](#exception-hierarchy-and-types)
3. [Try-Catch-Finally Patterns](#try-catch-finally-patterns)
4. [Custom Exceptions](#custom-exceptions)
5. [Exception Best Practices](#exception-best-practices)
6. [Result Pattern and Functional Error Handling](#result-pattern-and-functional-error-handling)
7. [Global Error Handling](#global-error-handling)
8. [Logging and Error Tracking](#logging-and-error-tracking)
9. [Performance Considerations](#performance-considerations)
10. [Modern Error Handling Patterns](#modern-error-handling-patterns)
11. [Enterprise Patterns](#enterprise-patterns)

---

## Exception Handling Fundamentals

### What Are Exceptions?

Exceptions are runtime errors that occur during program execution. They represent exceptional circumstances that disrupt the normal flow of execution.

```csharp
// Basic exception throwing
public void Divide(int a, int b)
{
    if (b == 0)
        throw new DivideByZeroException("Division by zero is not allowed");

    int result = a / b;
    Console.WriteLine($"Result: {result}");
}
```

### Exception vs Return Codes

```csharp
// Traditional error handling with return codes
public bool TryParseInteger(string input, out int result)
{
    result = 0;
    if (string.IsNullOrEmpty(input))
        return false;

    return int.TryParse(input, out result);
}

// Exception-based error handling
public int ParseInteger(string input)
{
    if (string.IsNullOrEmpty(input))
        throw new ArgumentNullException(nameof(input));

    if (!int.TryParse(input, out int result))
        throw new FormatException($"'{input}' is not a valid integer");

    return result;
}
```

### When to Use Exceptions

**Use exceptions for:**
- Truly exceptional conditions
- Unrecoverable errors
- Violations of contracts/preconditions
- Resource failures

**Don't use exceptions for:**
- Normal control flow
- Expected conditions
- Performance-critical paths
- Validation errors in user input

---

## Exception Hierarchy and Types

### .NET Exception Hierarchy

```
System.Object
    System.Exception
        System.ApplicationException (deprecated)
        System.SystemException
            System.ArgumentException
                System.ArgumentNullException
                System.ArgumentOutOfRangeException
            System.InvalidOperationException
            System.NotImplementedException
            System.NotSupportedException
            System.NullReferenceException
            System.IndexOutOfRangeException
            System.IO.IOException
                System.IO.FileNotFoundException
                System.IO.DirectoryNotFoundException
            System.Net.NetworkException
```

### Common Exception Types

```csharp
// Argument exceptions
public void ProcessUser(string name, int age)
{
    if (name == null)
        throw new ArgumentNullException(nameof(name));

    if (string.IsNullOrWhiteSpace(name))
        throw new ArgumentException("Name cannot be empty", nameof(name));

    if (age < 0 || age > 150)
        throw new ArgumentOutOfRangeException(nameof(age), age, "Age must be between 0 and 150");
}

// State exceptions
public class BankAccount
{
    private decimal _balance;
    private bool _isClosed;

    public void Withdraw(decimal amount)
    {
        if (_isClosed)
            throw new InvalidOperationException("Cannot withdraw from closed account");

        if (amount > _balance)
            throw new InvalidOperationException("Insufficient funds");

        _balance -= amount;
    }
}

// Not implemented/supported exceptions
public abstract class Shape
{
    public virtual double Area => throw new NotImplementedException("Subclasses must implement Area");
    public virtual void Draw() => throw new NotSupportedException("Drawing not supported for this shape");
}
```

### Exception Properties

```csharp
public void DemonstrateExceptionProperties()
{
    try
    {
        ThrowDetailedException();
    }
    catch (Exception ex)
    {
        Console.WriteLine($"Message: {ex.Message}");
        Console.WriteLine($"Stack Trace: {ex.StackTrace}");
        Console.WriteLine($"Source: {ex.Source}");
        Console.WriteLine($"Target Site: {ex.TargetSite}");
        Console.WriteLine($"Help Link: {ex.HelpLink}");

        // Inner exceptions
        Exception innerEx = ex.InnerException;
        while (innerEx != null)
        {
            Console.WriteLine($"Inner Exception: {innerEx.Message}");
            innerEx = innerEx.InnerException;
        }

        // Data dictionary
        foreach (DictionaryEntry entry in ex.Data)
        {
            Console.WriteLine($"Data: {entry.Key} = {entry.Value}");
        }
    }
}

private void ThrowDetailedException()
{
    try
    {
        throw new InvalidOperationException("Inner exception occurred");
    }
    catch (Exception innerEx)
    {
        var outerEx = new ApplicationException("Outer exception occurred", innerEx);
        outerEx.Data.Add("UserId", 12345);
        outerEx.Data.Add("Timestamp", DateTime.UtcNow);
        outerEx.HelpLink = "https://docs.mycompany.com/errors/app001";
        throw outerEx;
    }
}
```

---

## Try-Catch-Finally Patterns

### Basic Try-Catch

```csharp
public void BasicTryCatch()
{
    try
    {
        // Code that might throw an exception
        int result = 10 / 0;
    }
    catch (DivideByZeroException ex)
    {
        Console.WriteLine($"Division error: {ex.Message}");
    }
    catch (Exception ex)
    {
        Console.WriteLine($"General error: {ex.Message}");
    }
}
```

### Multiple Catch Blocks

```csharp
public void MultipleCatchBlocks()
{
    try
    {
        ProcessFile("data.txt");
    }
    catch (FileNotFoundException ex)
    {
        Console.WriteLine($"File not found: {ex.FileName}");
        // Log specific file error
    }
    catch (UnauthorizedAccessException ex)
    {
        Console.WriteLine($"Access denied: {ex.Message}");
        // Handle permission issues
    }
    catch (IOException ex)
    {
        Console.WriteLine($"IO error: {ex.Message}");
        // Handle general IO problems
    }
    catch (Exception ex)
    {
        Console.WriteLine($"Unexpected error: {ex.Message}");
        // Handle any other exceptions
        throw; // Re-throw if you can't handle it
    }
}
```

### Exception Filters (C# 6+)

```csharp
public void ExceptionFilters()
{
    try
    {
        ProcessData();
    }
    catch (HttpRequestException ex) when (ex.Message.Contains("404"))
    {
        Console.WriteLine("Resource not found");
    }
    catch (HttpRequestException ex) when (ex.Message.Contains("401"))
    {
        Console.WriteLine("Unauthorized access");
    }
    catch (Exception ex) when (LogException(ex))
    {
        // This catch block will never execute because LogException returns false
        // but the exception will be logged
    }
}

private bool LogException(Exception ex)
{
    // Log the exception and return false to not handle it here
    Console.WriteLine($"Logged: {ex.Message}");
    return false;
}
```

### Finally Block

```csharp
public void FinallyBlockDemo()
{
    FileStream file = null;
    try
    {
        file = new FileStream("data.txt", FileMode.Open);
        // Process file
    }
    catch (IOException ex)
    {
        Console.WriteLine($"File error: {ex.Message}");
    }
    finally
    {
        // This always executes
        file?.Dispose();
        Console.WriteLine("Cleanup completed");
    }
}
```

### Using Statement (Automatic Cleanup)

```csharp
public void UsingStatementDemo()
{
    // Traditional using
    using (var file = new FileStream("data.txt", FileMode.Open))
    {
        // Process file - automatically disposed even if exception occurs
    }

    // Using declaration (C# 8+)
    using var connection = new SqlConnection(connectionString);
    connection.Open();
    // Connection automatically disposed at end of scope
}
```

### Try-Catch-Finally vs Using

```csharp
// Equivalent implementations
public void ManualCleanup()
{
    SqlConnection connection = null;
    try
    {
        connection = new SqlConnection(connectionString);
        connection.Open();
        // Use connection
    }
    catch (SqlException ex)
    {
        // Handle database errors
        throw;
    }
    finally
    {
        connection?.Dispose();
    }
}

public void AutomaticCleanup()
{
    try
    {
        using var connection = new SqlConnection(connectionString);
        connection.Open();
        // Use connection - automatically disposed
    }
    catch (SqlException ex)
    {
        // Handle database errors
        throw;
    }
}
```

---

## Custom Exceptions

### Creating Custom Exceptions

```csharp
// Basic custom exception
public class BusinessRuleException : Exception
{
    public BusinessRuleException() { }

    public BusinessRuleException(string message) : base(message) { }

    public BusinessRuleException(string message, Exception innerException)
        : base(message, innerException) { }

    // For serialization (if needed)
    protected BusinessRuleException(SerializationInfo info, StreamingContext context)
        : base(info, context) { }
}
```

### Rich Custom Exceptions

```csharp
public class ValidationException : Exception
{
    public IReadOnlyList<ValidationError> Errors { get; }
    public string PropertyName { get; }

    public ValidationException(string propertyName, string message)
        : base(message)
    {
        PropertyName = propertyName;
        Errors = new List<ValidationError> { new ValidationError(propertyName, message) };
    }

    public ValidationException(IEnumerable<ValidationError> errors)
        : base("One or more validation errors occurred")
    {
        Errors = errors.ToList().AsReadOnly();
    }

    public ValidationException(string message, Exception innerException)
        : base(message, innerException) { }

    protected ValidationException(SerializationInfo info, StreamingContext context)
        : base(info, context)
    {
        PropertyName = info.GetString(nameof(PropertyName));
        Errors = (IReadOnlyList<ValidationError>)info.GetValue(nameof(Errors), typeof(IReadOnlyList<ValidationError>));
    }

    public override void GetObjectData(SerializationInfo info, StreamingContext context)
    {
        base.GetObjectData(info, context);
        info.AddValue(nameof(PropertyName), PropertyName);
        info.AddValue(nameof(Errors), Errors);
    }
}

public record ValidationError(string PropertyName, string ErrorMessage);
```

### Domain-Specific Exceptions

```csharp
// E-commerce domain exceptions
public class InsufficientInventoryException : Exception
{
    public string ProductId { get; }
    public int RequestedQuantity { get; }
    public int AvailableQuantity { get; }

    public InsufficientInventoryException(string productId, int requested, int available)
        : base($"Insufficient inventory for product {productId}. Requested: {requested}, Available: {available}")
    {
        ProductId = productId;
        RequestedQuantity = requested;
        AvailableQuantity = available;
    }
}

public class PaymentProcessingException : Exception
{
    public string TransactionId { get; }
    public string PaymentMethod { get; }
    public decimal Amount { get; }

    public PaymentProcessingException(string transactionId, string paymentMethod, decimal amount, string message)
        : base(message)
    {
        TransactionId = transactionId;
        PaymentMethod = paymentMethod;
        Amount = amount;
    }

    public PaymentProcessingException(string transactionId, string paymentMethod, decimal amount,
        string message, Exception innerException)
        : base(message, innerException)
    {
        TransactionId = transactionId;
        PaymentMethod = paymentMethod;
        Amount = amount;
    }
}
```

### Exception Factory Pattern

```csharp
public static class ExceptionFactory
{
    public static ArgumentNullException ArgumentNull(string parameterName, string message = null)
    {
        return new ArgumentNullException(parameterName, message ?? $"{parameterName} cannot be null");
    }

    public static ArgumentException ArgumentInvalid(string parameterName, object value, string reason)
    {
        return new ArgumentException($"Invalid value '{value}' for {parameterName}: {reason}", parameterName);
    }

    public static InvalidOperationException InvalidOperation(string operation, string currentState)
    {
        return new InvalidOperationException($"Cannot {operation} when in state: {currentState}");
    }

    public static ValidationException Validation(params (string Property, string Error)[] errors)
    {
        var validationErrors = errors.Select(e => new ValidationError(e.Property, e.Error));
        return new ValidationException(validationErrors);
    }
}

// Usage
public void ProcessOrder(Order order)
{
    if (order == null)
        throw ExceptionFactory.ArgumentNull(nameof(order));

    if (order.Status != OrderStatus.Pending)
        throw ExceptionFactory.InvalidOperation("process order", order.Status.ToString());
}
```

---

## Exception Best Practices

### Fail Fast Principle

```csharp
public class OrderProcessor
{
    private readonly IPaymentService _paymentService;
    private readonly IInventoryService _inventoryService;

    public OrderProcessor(IPaymentService paymentService, IInventoryService inventoryService)
    {
        // Fail fast - validate dependencies immediately
        _paymentService = paymentService ?? throw new ArgumentNullException(nameof(paymentService));
        _inventoryService = inventoryService ?? throw new ArgumentNullException(nameof(inventoryService));
    }

    public void ProcessOrder(Order order)
    {
        // Validate inputs immediately
        if (order == null)
            throw new ArgumentNullException(nameof(order));

        if (order.Items == null || !order.Items.Any())
            throw new ArgumentException("Order must contain at least one item", nameof(order));

        // Continue with processing...
    }
}
```

### Guard Clauses

```csharp
public static class Guard
{
    public static void Against<TException>(bool condition, string message)
        where TException : Exception, new()
    {
        if (condition)
        {
            throw (TException)Activator.CreateInstance(typeof(TException), message);
        }
    }

    public static void NotNull<T>(T value, string parameterName) where T : class
    {
        if (value == null)
            throw new ArgumentNullException(parameterName);
    }

    public static void NotNullOrEmpty(string value, string parameterName)
    {
        if (string.IsNullOrEmpty(value))
            throw new ArgumentException("Value cannot be null or empty", parameterName);
    }

    public static void NotNullOrWhiteSpace(string value, string parameterName)
    {
        if (string.IsNullOrWhiteSpace(value))
            throw new ArgumentException("Value cannot be null or whitespace", parameterName);
    }

    public static void InRange(int value, int min, int max, string parameterName)
    {
        if (value < min || value > max)
            throw new ArgumentOutOfRangeException(parameterName, value,
                $"Value must be between {min} and {max}");
    }
}

// Usage
public void CreateUser(string name, string email, int age)
{
    Guard.NotNullOrWhiteSpace(name, nameof(name));
    Guard.NotNullOrWhiteSpace(email, nameof(email));
    Guard.InRange(age, 0, 150, nameof(age));
    Guard.Against<ArgumentException>(!email.Contains("@"), "Email must be valid");

    // Continue with user creation...
}
```

### Exception Wrapping

```csharp
public class UserService
{
    private readonly IUserRepository _repository;

    public async Task<User> GetUserAsync(int userId)
    {
        try
        {
            return await _repository.GetByIdAsync(userId);
        }
        catch (SqlException ex)
        {
            // Wrap low-level exception with domain-specific exception
            throw new UserServiceException($"Failed to retrieve user {userId}", ex);
        }
        catch (Exception ex)
        {
            // Log unexpected exceptions and re-throw
            _logger.LogError(ex, "Unexpected error retrieving user {UserId}", userId);
            throw;
        }
    }
}
```

### Preserving Stack Traces

```csharp
public void DemonstrateStackTracePreservation()
{
    try
    {
        ThrowOriginalException();
    }
    catch (Exception ex)
    {
        // Wrong - loses original stack trace
        // throw ex;

        // Correct - preserves stack trace
        throw;

        // Also correct - adds to stack trace
        // throw new ApplicationException("Wrapper exception", ex);
    }
}

// ExceptionDispatchInfo for complex scenarios
public void AdvancedStackTracePreservation()
{
    Exception capturedException = null;

    try
    {
        DoSomethingDangerous();
    }
    catch (Exception ex)
    {
        capturedException = ex;
    }

    // Later, re-throw with preserved stack trace
    if (capturedException != null)
    {
        ExceptionDispatchInfo.Capture(capturedException).Throw();
    }
}
```

### Exception Message Guidelines

```csharp
public class ExceptionMessageExamples
{
    // Good: Specific, actionable messages
    public void GoodMessages(string fileName, int timeout)
    {
        if (string.IsNullOrEmpty(fileName))
            throw new ArgumentException("File name cannot be null or empty", nameof(fileName));

        if (timeout <= 0)
            throw new ArgumentOutOfRangeException(nameof(timeout), timeout,
                "Timeout must be greater than zero");

        if (!File.Exists(fileName))
            throw new FileNotFoundException($"The file '{fileName}' was not found", fileName);
    }

    // Bad: Vague, unhelpful messages
    public void BadMessages(string fileName, int timeout)
    {
        if (string.IsNullOrEmpty(fileName))
            throw new Exception("Error"); // Too vague

        if (timeout <= 0)
            throw new Exception("Bad timeout"); // Not descriptive

        if (!File.Exists(fileName))
            throw new Exception("File issue"); // No context
    }
}
```

---

## Result Pattern and Functional Error Handling

### Basic Result Pattern

```csharp
public class Result<T>
{
    public bool IsSuccess { get; }
    public bool IsFailure => !IsSuccess;
    public T Value { get; }
    public string Error { get; }

    private Result(T value)
    {
        IsSuccess = true;
        Value = value;
        Error = null;
    }

    private Result(string error)
    {
        IsSuccess = false;
        Value = default(T);
        Error = error;
    }

    public static Result<T> Success(T value) => new Result<T>(value);
    public static Result<T> Failure(string error) => new Result<T>(error);

    public static implicit operator Result<T>(T value) => Success(value);
    public static implicit operator Result<T>(string error) => Failure(error);
}

public class Result
{
    public bool IsSuccess { get; }
    public bool IsFailure => !IsSuccess;
    public string Error { get; }

    private Result(bool isSuccess, string error)
    {
        IsSuccess = isSuccess;
        Error = error;
    }

    public static Result Success() => new Result(true, null);
    public static Result Failure(string error) => new Result(false, error);

    public static implicit operator Result(string error) => Failure(error);
}
```

### Advanced Result Pattern

```csharp
public class Result<T, TError>
{
    public bool IsSuccess { get; }
    public bool IsFailure => !IsSuccess;
    public T Value { get; }
    public TError Error { get; }

    private Result(T value)
    {
        IsSuccess = true;
        Value = value;
        Error = default;
    }

    private Result(TError error)
    {
        IsSuccess = false;
        Value = default;
        Error = error;
    }

    public static Result<T, TError> Success(T value) => new(value);
    public static Result<T, TError> Failure(TError error) => new(error);

    // Functional operations
    public Result<TResult, TError> Map<TResult>(Func<T, TResult> func)
    {
        return IsSuccess
            ? Result<TResult, TError>.Success(func(Value))
            : Result<TResult, TError>.Failure(Error);
    }

    public async Task<Result<TResult, TError>> MapAsync<TResult>(Func<T, Task<TResult>> func)
    {
        return IsSuccess
            ? Result<TResult, TError>.Success(await func(Value))
            : Result<TResult, TError>.Failure(Error);
    }

    public Result<TResult, TError> Bind<TResult>(Func<T, Result<TResult, TError>> func)
    {
        return IsSuccess ? func(Value) : Result<TResult, TError>.Failure(Error);
    }

    public T Match(Func<T, T> onSuccess, Func<TError, T> onFailure)
    {
        return IsSuccess ? onSuccess(Value) : onFailure(Error);
    }
}
```

### Using Result Pattern

```csharp
public class UserService
{
    public Result<User> GetUser(int userId)
    {
        if (userId <= 0)
            return "Invalid user ID";

        var user = _repository.FindById(userId);
        return user ?? "User not found";
    }

    public Result<User> CreateUser(CreateUserRequest request)
    {
        return ValidateUser(request)
            .Bind(req => CheckUserExists(req.Email))
            .Bind(req => SaveUser(req))
            .Map(user => user);
    }

    private Result<CreateUserRequest> ValidateUser(CreateUserRequest request)
    {
        if (string.IsNullOrEmpty(request.Name))
            return "Name is required";

        if (string.IsNullOrEmpty(request.Email))
            return "Email is required";

        return request;
    }

    private Result<CreateUserRequest> CheckUserExists(string email)
    {
        if (_repository.ExistsByEmail(email))
            return "User already exists";

        return request;
    }

    private Result<User> SaveUser(CreateUserRequest request)
    {
        try
        {
            var user = new User(request.Name, request.Email);
            return _repository.Save(user);
        }
        catch (Exception ex)
        {
            return $"Failed to save user: {ex.Message}";
        }
    }
}

// Usage
public void HandleUserCreation()
{
    var request = new CreateUserRequest("John Doe", "john@example.com");
    var result = _userService.CreateUser(request);

    result.Match(
        user => Console.WriteLine($"Created user: {user.Name}"),
        error => Console.WriteLine($"Error: {error}")
    );
}
```

### Option Pattern for Null Handling

```csharp
public abstract class Option<T>
{
    public abstract bool HasValue { get; }
    public abstract T Value { get; }

    public abstract TResult Match<TResult>(Func<T, TResult> some, Func<TResult> none);
    public abstract Option<TResult> Map<TResult>(Func<T, TResult> mapper);
    public abstract Option<TResult> Bind<TResult>(Func<T, Option<TResult>> binder);

    public static implicit operator Option<T>(T value) =>
        value != null ? new Some<T>(value) : new None<T>();

    public static Option<T> None() => new None<T>();
    public static Option<T> Some(T value) => new Some<T>(value);
}

public class Some<T> : Option<T>
{
    public override bool HasValue => true;
    public override T Value { get; }

    public Some(T value) => Value = value ?? throw new ArgumentNullException(nameof(value));

    public override TResult Match<TResult>(Func<T, TResult> some, Func<TResult> none) => some(Value);
    public override Option<TResult> Map<TResult>(Func<T, TResult> mapper) => mapper(Value);
    public override Option<TResult> Bind<TResult>(Func<T, Option<TResult>> binder) => binder(Value);
}

public class None<T> : Option<T>
{
    public override bool HasValue => false;
    public override T Value => throw new InvalidOperationException("None has no value");

    public override TResult Match<TResult>(Func<T, TResult> some, Func<TResult> none) => none();
    public override Option<TResult> Map<TResult>(Func<T, TResult> mapper) => new None<TResult>();
    public override Option<TResult> Bind<TResult>(Func<T, Option<TResult>> binder) => new None<TResult>();
}

// Usage
public Option<User> FindUser(int id)
{
    var user = _repository.FindById(id);
    return user; // Implicit conversion
}

public void ProcessUser(int id)
{
    FindUser(id).Match(
        user => Console.WriteLine($"Found: {user.Name}"),
        () => Console.WriteLine("User not found")
    );
}
```

### Either Pattern for Multiple Error Types

```csharp
public abstract class Either<TLeft, TRight>
{
    public abstract bool IsLeft { get; }
    public abstract bool IsRight { get; }

    public abstract TResult Match<TResult>(
        Func<TLeft, TResult> left,
        Func<TRight, TResult> right);

    public static implicit operator Either<TLeft, TRight>(TLeft left) => new Left<TLeft, TRight>(left);
    public static implicit operator Either<TLeft, TRight>(TRight right) => new Right<TLeft, TRight>(right);
}

public class Left<TLeft, TRight> : Either<TLeft, TRight>
{
    public TLeft Value { get; }
    public override bool IsLeft => true;
    public override bool IsRight => false;

    public Left(TLeft value) => Value = value;

    public override TResult Match<TResult>(Func<TLeft, TResult> left, Func<TRight, TResult> right)
        => left(Value);
}

public class Right<TLeft, TRight> : Either<TLeft, TRight>
{
    public TRight Value { get; }
    public override bool IsLeft => false;
    public override bool IsRight => true;

    public Right(TRight value) => Value = value;

    public override TResult Match<TResult>(Func<TLeft, TResult> left, Func<TRight, TResult> right)
        => right(Value);
}
```

---

## Global Error Handling

### Unhandled Exception Handling

```csharp
public class Program
{
    public static void Main(string[] args)
    {
        // Set up global exception handlers
        AppDomain.CurrentDomain.UnhandledException += OnUnhandledException;
        TaskScheduler.UnobservedTaskException += OnUnobservedTaskException;

        try
        {
            var app = CreateHostBuilder(args).Build();
            app.Run();
        }
        catch (Exception ex)
        {
            // Log startup errors
            var logger = LogManager.GetCurrentClassLogger();
            logger.Fatal(ex, "Application failed to start");
            throw;
        }
    }

    private static void OnUnhandledException(object sender, UnhandledExceptionEventArgs e)
    {
        var logger = LogManager.GetCurrentClassLogger();
        logger.Fatal(e.ExceptionObject as Exception, "Unhandled exception occurred");

        // Perform cleanup if possible
        if (e.IsTerminating)
        {
            logger.Info("Application is terminating");
        }
    }

    private static void OnUnobservedTaskException(object sender, UnobservedTaskExceptionEventArgs e)
    {
        var logger = LogManager.GetCurrentClassLogger();
        logger.Error(e.Exception, "Unobserved task exception");

        // Mark as observed to prevent process termination
        e.SetObserved();
    }
}
```

### ASP.NET Core Global Exception Handling

```csharp
// Custom exception middleware
public class GlobalExceptionMiddleware
{
    private readonly RequestDelegate _next;
    private readonly ILogger<GlobalExceptionMiddleware> _logger;

    public GlobalExceptionMiddleware(RequestDelegate next, ILogger<GlobalExceptionMiddleware> logger)
    {
        _next = next;
        _logger = logger;
    }

    public async Task InvokeAsync(HttpContext context)
    {
        try
        {
            await _next(context);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "An unhandled exception occurred");
            await HandleExceptionAsync(context, ex);
        }
    }

    private async Task HandleExceptionAsync(HttpContext context, Exception exception)
    {
        context.Response.ContentType = "application/json";

        var response = exception switch
        {
            ValidationException validationEx => new
            {
                StatusCode = 400,
                Message = "Validation failed",
                Errors = validationEx.Errors.Select(e => new { e.PropertyName, e.ErrorMessage })
            },
            BusinessRuleException businessEx => new
            {
                StatusCode = 400,
                Message = businessEx.Message
            },
            UnauthorizedAccessException => new
            {
                StatusCode = 401,
                Message = "Unauthorized access"
            },
            NotFoundException => new
            {
                StatusCode = 404,
                Message = "Resource not found"
            },
            _ => new
            {
                StatusCode = 500,
                Message = "An internal server error occurred"
            }
        };

        context.Response.StatusCode = response.StatusCode;
        await context.Response.WriteAsync(JsonSerializer.Serialize(response));
    }
}

// Extension method for registration
public static class MiddlewareExtensions
{
    public static IApplicationBuilder UseGlobalExceptionHandling(this IApplicationBuilder builder)
    {
        return builder.UseMiddleware<GlobalExceptionMiddleware>();
    }
}

// In Startup.cs or Program.cs
public void Configure(IApplicationBuilder app, IWebHostEnvironment env)
{
    if (env.IsDevelopment())
    {
        app.UseDeveloperExceptionPage();
    }
    else
    {
        app.UseGlobalExceptionHandling();
        app.UseHsts();
    }

    // Other middleware...
}
```

### Exception Filter for Controllers

```csharp
public class ApiExceptionFilterAttribute : ExceptionFilterAttribute
{
    private readonly ILogger<ApiExceptionFilterAttribute> _logger;

    public ApiExceptionFilterAttribute(ILogger<ApiExceptionFilterAttribute> logger)
    {
        _logger = logger;
    }

    public override void OnException(ExceptionContext context)
    {
        var exception = context.Exception;
        _logger.LogError(exception, "API Exception occurred");

        var result = exception switch
        {
            ValidationException validationEx => new BadRequestObjectResult(new
            {
                Message = "Validation failed",
                Errors = validationEx.Errors
            }),
            BusinessRuleException businessEx => new BadRequestObjectResult(new
            {
                Message = businessEx.Message
            }),
            UnauthorizedAccessException => new UnauthorizedResult(),
            NotFoundException => new NotFoundResult(),
            _ => new ObjectResult(new { Message = "An error occurred" })
            {
                StatusCode = 500
            }
        };

        context.Result = result;
        context.ExceptionHandled = true;
    }
}

// Usage
[ApiController]
[Route("api/[controller]")]
[ServiceFilter(typeof(ApiExceptionFilterAttribute))]
public class UsersController : ControllerBase
{
    // Controller actions...
}
```

### Background Service Error Handling

```csharp
public class RobustBackgroundService : BackgroundService
{
    private readonly ILogger<RobustBackgroundService> _logger;
    private readonly IServiceProvider _serviceProvider;

    public RobustBackgroundService(ILogger<RobustBackgroundService> logger, IServiceProvider serviceProvider)
    {
        _logger = logger;
        _serviceProvider = serviceProvider;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                await ProcessWorkAsync(stoppingToken);
            }
            catch (OperationCanceledException)
            {
                _logger.LogInformation("Background service was cancelled");
                break;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error in background service");

                // Implement exponential backoff
                await Task.Delay(TimeSpan.FromMinutes(1), stoppingToken);
            }
        }
    }

    private async Task ProcessWorkAsync(CancellationToken cancellationToken)
    {
        using var scope = _serviceProvider.CreateScope();
        var workProcessor = scope.ServiceProvider.GetRequiredService<IWorkProcessor>();

        try
        {
            await workProcessor.ProcessAsync(cancellationToken);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error processing work item");
            // Don't rethrow - handle gracefully and continue
        }
    }
}
```

---

## Logging and Error Tracking

### Structured Logging with Serilog

```csharp
// Configuration
public static class LoggingConfiguration
{
    public static ILogger CreateLogger()
    {
        return new LoggerConfiguration()
            .MinimumLevel.Information()
            .MinimumLevel.Override("Microsoft", LogEventLevel.Warning)
            .MinimumLevel.Override("System", LogEventLevel.Warning)
            .Enrich.FromLogContext()
            .Enrich.WithMachineName()
            .Enrich.WithThreadId()
            .Enrich.WithEnvironmentUserName()
            .WriteTo.Console(outputTemplate: "{Timestamp:yyyy-MM-dd HH:mm:ss} [{Level:u3}] {Message:lj} {Properties:j}{NewLine}{Exception}")
            .WriteTo.File("logs/app-.log", rollingInterval: RollingInterval.Day)
            .WriteTo.Seq("http://localhost:5341") // Seq for structured logs
            .CreateLogger();
    }
}

// Usage in services
public class UserService
{
    private readonly ILogger<UserService> _logger;

    public UserService(ILogger<UserService> logger)
    {
        _logger = logger;
    }

    public async Task<User> CreateUserAsync(CreateUserRequest request)
    {
        using var activity = _logger.BeginScope("Creating user {Email}", request.Email);

        try
        {
            _logger.LogInformation("Starting user creation for {Email}", request.Email);

            var user = new User(request.Name, request.Email);
            await _repository.SaveAsync(user);

            _logger.LogInformation("Successfully created user {UserId} with email {Email}",
                user.Id, request.Email);

            return user;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to create user with email {Email}", request.Email);
            throw;
        }
    }
}
```

### Error Tracking with Custom Context

```csharp
public class ErrorContext
{
    public string CorrelationId { get; set; } = Guid.NewGuid().ToString();
    public string UserId { get; set; }
    public string SessionId { get; set; }
    public string RequestId { get; set; }
    public Dictionary<string, object> Properties { get; set; } = new();

    public static ErrorContext Current => _current.Value ??= new ErrorContext();
    private static readonly ThreadLocal<ErrorContext> _current = new();
}

public static class ErrorLogger
{
    private static readonly ILogger Logger = LogManager.GetCurrentClassLogger();

    public static void LogError(Exception exception, string message = null, params object[] args)
    {
        var context = ErrorContext.Current;

        Logger.Error(exception)
            .Message(message ?? exception.Message, args)
            .Property("CorrelationId", context.CorrelationId)
            .Property("UserId", context.UserId)
            .Property("SessionId", context.SessionId)
            .Property("RequestId", context.RequestId)
            .Properties(context.Properties)
            .Write();
    }

    public static void LogBusinessError(string category, string message, params object[] args)
    {
        var context = ErrorContext.Current;

        Logger.Warn()
            .Message($"[{category}] {message}", args)
            .Property("Category", category)
            .Property("CorrelationId", context.CorrelationId)
            .Property("UserId", context.UserId)
            .Write();
    }
}
```

### Application Insights Integration

```csharp
public class TelemetryService
{
    private readonly TelemetryClient _telemetryClient;

    public TelemetryService(TelemetryClient telemetryClient)
    {
        _telemetryClient = telemetryClient;
    }

    public void TrackException(Exception exception, Dictionary<string, string> properties = null)
    {
        var telemetryProperties = properties ?? new Dictionary<string, string>();

        // Add context information
        telemetryProperties["CorrelationId"] = ErrorContext.Current.CorrelationId;
        telemetryProperties["UserId"] = ErrorContext.Current.UserId;
        telemetryProperties["Timestamp"] = DateTimeOffset.UtcNow.ToString("O");

        _telemetryClient.TrackException(exception, telemetryProperties);
    }

    public void TrackBusinessEvent(string eventName, Dictionary<string, string> properties = null,
        Dictionary<string, double> metrics = null)
    {
        _telemetryClient.TrackEvent(eventName, properties, metrics);
    }

    public IDisposable StartOperation(string operationName)
    {
        return _telemetryClient.StartOperation<RequestTelemetry>(operationName);
    }
}

// Usage
public class OrderService
{
    private readonly TelemetryService _telemetry;

    public async Task<Order> ProcessOrderAsync(CreateOrderRequest request)
    {
        using var operation = _telemetry.StartOperation("ProcessOrder");

        try
        {
            // Process order
            var order = await ProcessOrderInternal(request);

            _telemetry.TrackBusinessEvent("OrderProcessed", new Dictionary<string, string>
            {
                ["OrderId"] = order.Id.ToString(),
                ["CustomerId"] = request.CustomerId.ToString(),
                ["Amount"] = request.TotalAmount.ToString("C")
            });

            return order;
        }
        catch (Exception ex)
        {
            _telemetry.TrackException(ex, new Dictionary<string, string>
            {
                ["CustomerId"] = request.CustomerId.ToString(),
                ["OrderValue"] = request.TotalAmount.ToString("C")
            });
            throw;
        }
    }
}
```

---

## Performance Considerations

### Exception Performance Impact

```csharp
public class PerformanceComparison
{
    // Expensive - creating exceptions
    public string SlowValidation(string input)
    {
        try
        {
            return ProcessString(input);
        }
        catch (ArgumentException)
        {
            return "Invalid input";
        }
    }

    // Faster - avoiding exceptions
    public string FastValidation(string input)
    {
        if (IsValidInput(input))
        {
            return ProcessString(input);
        }
        return "Invalid input";
    }

    // Benchmark example
    [Benchmark]
    public string UsingExceptions()
    {
        try
        {
            return int.Parse("not a number").ToString();
        }
        catch (FormatException)
        {
            return "error";
        }
    }

    [Benchmark]
    public string UsingTryParse()
    {
        return int.TryParse("not a number", out int result) ? result.ToString() : "error";
    }
}
```

### Exception Caching and Reuse

```csharp
public static class CachedExceptions
{
    // Cache common exceptions to avoid allocation overhead
    public static readonly ArgumentNullException NullArgument =
        new ArgumentNullException("Cached null argument exception");

    public static readonly InvalidOperationException InvalidState =
        new InvalidOperationException("Cached invalid operation exception");

    // For parameterized exceptions, use factory methods
    private static readonly ConcurrentDictionary<string, ArgumentNullException> _nullExceptions =
        new ConcurrentDictionary<string, ArgumentNullException>();

    public static ArgumentNullException ArgumentNull(string parameterName)
    {
        return _nullExceptions.GetOrAdd(parameterName,
            name => new ArgumentNullException(name));
    }
}

// Usage in hot paths
public void ProcessItems(IEnumerable<Item> items)
{
    if (items == null)
        throw CachedExceptions.ArgumentNull(nameof(items));

    // Process items...
}
```

### Efficient Error Handling Patterns

```csharp
// Use spans and memory for performance-critical scenarios
public ref struct ValidationResult
{
    public bool IsValid;
    public ReadOnlySpan<char> ErrorMessage;

    public ValidationResult(bool isValid, ReadOnlySpan<char> errorMessage = default)
    {
        IsValid = isValid;
        ErrorMessage = errorMessage;
    }
}

public ValidationResult ValidateEmail(ReadOnlySpan<char> email)
{
    if (email.IsEmpty)
        return new ValidationResult(false, "Email is required");

    if (!email.Contains('@'))
        return new ValidationResult(false, "Email must contain @");

    return new ValidationResult(true);
}

// Pool objects to reduce allocations
public class ValidationErrorPool
{
    private static readonly ObjectPool<List<ValidationError>> Pool =
        new DefaultObjectPool<List<ValidationError>>(new ListPooledObjectPolicy<ValidationError>());

    public static List<ValidationError> Get() => Pool.Get();
    public static void Return(List<ValidationError> errors)
    {
        errors.Clear();
        Pool.Return(errors);
    }
}
```

---

## Modern Error Handling Patterns

### Nullable Reference Types (C# 8+)

```csharp
#nullable enable

public class ModernUserService
{
    // Nullable annotations help prevent null reference exceptions
    public User? FindUser(int id)
    {
        // Compiler knows this might return null
        return _repository.FindById(id);
    }

    public void ProcessUser(User user) // Non-nullable parameter
    {
        // Compiler ensures user is not null
        Console.WriteLine(user.Name.Length); // Safe
    }

    public void SafeProcessUser(int id)
    {
        User? user = FindUser(id);

        if (user is not null) // Null check
        {
            ProcessUser(user); // Now safe to call
        }
    }

    // Null-forgiving operator when you're certain
    public void ProcessKnownUser(int id)
    {
        User user = FindUser(id)!; // ! tells compiler: trust me, it's not null
        ProcessUser(user);
    }
}

// Nullable context for existing code
public class LegacyService
{
#nullable disable
    public string GetUserName(User user)
    {
        return user.Name; // No warnings for potential null
    }
#nullable restore
}
```

### Pattern Matching for Exception Handling (C# 8+)

```csharp
public class ModernExceptionHandling
{
    public string HandleException(Exception ex) => ex switch
    {
        ArgumentNullException { ParamName: var param } => $"Null parameter: {param}",
        ArgumentException { ParamName: var param } => $"Invalid parameter: {param}",
        InvalidOperationException { Message: var msg } when msg.Contains("timeout") => "Operation timed out",
        InvalidOperationException => "Invalid operation",
        FileNotFoundException { FileName: var file } => $"File not found: {file}",
        HttpRequestException { Data: var data } when data.Contains("StatusCode") =>
            $"HTTP error: {data["StatusCode"]}",
        _ => "Unknown error occurred"
    };

    public async Task<Result<T>> SafeExecuteAsync<T>(Func<Task<T>> operation)
    {
        try
        {
            var result = await operation();
            return Result<T>.Success(result);
        }
        catch (Exception ex)
        {
            var errorMessage = ex switch
            {
                OperationCanceledException => "Operation was cancelled",
                TimeoutException => "Operation timed out",
                HttpRequestException httpEx when httpEx.Message.Contains("404") => "Resource not found",
                HttpRequestException httpEx when httpEx.Message.Contains("401") => "Unauthorized",
                HttpRequestException => "Network error occurred",
                ArgumentException argEx => $"Invalid argument: {argEx.ParamName}",
                _ => ex.Message
            };

            return Result<T>.Failure(errorMessage);
        }
    }
}
```

### Init-Only Properties and Records for Immutable Error Objects

```csharp
// Modern error representation using records
public record ValidationError(string PropertyName, string Message, object? AttemptedValue = null);

public record BusinessError(string Code, string Message, Dictionary<string, object>? Context = null);

public record ApiError
{
    public string Code { get; init; } = string.Empty;
    public string Message { get; init; } = string.Empty;
    public string[]? Details { get; init; }
    public DateTimeOffset Timestamp { get; init; } = DateTimeOffset.UtcNow;
    public string? TraceId { get; init; }
}

// Usage with modern C# features
public record CreateUserRequest
{
    public string Name { get; init; } = string.Empty;
    public string Email { get; init; } = string.Empty;
    public int Age { get; init; }
}

public class ModernValidator
{
    public IEnumerable<ValidationError> ValidateUser(CreateUserRequest request)
    {
        if (string.IsNullOrWhiteSpace(request.Name))
            yield return new ValidationError(nameof(request.Name), "Name is required", request.Name);

        if (string.IsNullOrWhiteSpace(request.Email))
            yield return new ValidationError(nameof(request.Email), "Email is required", request.Email);
        else if (!request.Email.Contains('@'))
            yield return new ValidationError(nameof(request.Email), "Invalid email format", request.Email);

        if (request.Age is < 0 or > 150)
            yield return new ValidationError(nameof(request.Age), "Age must be between 0 and 150", request.Age);
    }
}
```

### Target-Typed New Expressions and Simplified Syntax

```csharp
public class SimplifiedErrorHandling
{
    // Target-typed new (C# 9)
    private static readonly Dictionary<Type, string> ErrorMessages = new()
    {
        [typeof(ArgumentNullException)] = "Null argument provided",
        [typeof(InvalidOperationException)] = "Invalid operation",
        [typeof(NotSupportedException)] = "Operation not supported"
    };

    // Simplified exception throwing
    public void ValidateInput(string? input, int value)
    {
        ArgumentNullException.ThrowIfNull(input); // .NET 6+
        ArgumentOutOfRangeException.ThrowIfNegative(value); // .NET 8+
    }

    // Modern switch expressions
    public HttpStatusCode GetStatusCode(Exception ex) => ex switch
    {
        ArgumentException => HttpStatusCode.BadRequest,
        UnauthorizedAccessException => HttpStatusCode.Unauthorized,
        NotFoundException => HttpStatusCode.NotFound,
        TimeoutException => HttpStatusCode.RequestTimeout,
        _ => HttpStatusCode.InternalServerError
    };
}
```

### Async Enumerable Error Handling (C# 8+)

```csharp
public class AsyncEnumerableErrorHandling
{
    public async IAsyncEnumerable<Result<Item>> ProcessItemsAsync(
        IAsyncEnumerable<Item> items,
        [EnumeratorCancellation] CancellationToken cancellationToken = default)
    {
        await foreach (var item in items.WithCancellation(cancellationToken))
        {
            Result<Item> result;
            try
            {
                var processedItem = await ProcessItemAsync(item);
                result = Result<Item>.Success(processedItem);
            }
            catch (Exception ex)
            {
                result = Result<Item>.Failure(ex.Message);
            }

            yield return result;
        }
    }

    // Using the async enumerable
    public async Task ProcessAllItemsAsync()
    {
        var items = GetItemsAsync();

        await foreach (var result in ProcessItemsAsync(items))
        {
            result.Match(
                item => Console.WriteLine($"Processed: {item.Name}"),
                error => Console.WriteLine($"Error: {error}")
            );
        }
    }
}
```

---

## Enterprise Patterns

### Circuit Breaker Pattern

```csharp
public class CircuitBreaker
{
    private readonly int _failureThreshold;
    private readonly TimeSpan _timeout;
    private readonly TimeSpan _retryTimeout;
    private int _failureCount;
    private DateTime _lastFailureTime;
    private CircuitBreakerState _state = CircuitBreakerState.Closed;
    private readonly object _lock = new();

    public CircuitBreaker(int failureThreshold, TimeSpan timeout, TimeSpan retryTimeout)
    {
        _failureThreshold = failureThreshold;
        _timeout = timeout;
        _retryTimeout = retryTimeout;
    }

    public async Task<T> ExecuteAsync<T>(Func<Task<T>> operation)
    {
        if (_state == CircuitBreakerState.Open)
        {
            if (DateTime.UtcNow - _lastFailureTime < _retryTimeout)
            {
                throw new CircuitBreakerOpenException("Circuit breaker is open");
            }

            _state = CircuitBreakerState.HalfOpen;
        }

        try
        {
            using var cancellationTokenSource = new CancellationTokenSource(_timeout);
            var result = await operation().ConfigureAwait(false);

            OnSuccess();
            return result;
        }
        catch (Exception ex)
        {
            OnFailure();
            throw;
        }
    }

    private void OnSuccess()
    {
        lock (_lock)
        {
            _failureCount = 0;
            _state = CircuitBreakerState.Closed;
        }
    }

    private void OnFailure()
    {
        lock (_lock)
        {
            _failureCount++;
            _lastFailureTime = DateTime.UtcNow;

            if (_failureCount >= _failureThreshold)
            {
                _state = CircuitBreakerState.Open;
            }
        }
    }
}

public enum CircuitBreakerState
{
    Closed,
    Open,
    HalfOpen
}

public class CircuitBreakerOpenException : Exception
{
    public CircuitBreakerOpenException(string message) : base(message) { }
}
```

### Retry Pattern with Exponential Backoff

```csharp
public class RetryPolicy
{
    private readonly int _maxRetries;
    private readonly TimeSpan _baseDelay;
    private readonly Func<Exception, bool> _retryPredicate;

    public RetryPolicy(int maxRetries, TimeSpan baseDelay, Func<Exception, bool> retryPredicate = null)
    {
        _maxRetries = maxRetries;
        _baseDelay = baseDelay;
        _retryPredicate = retryPredicate ?? (_ => true);
    }

    public async Task<T> ExecuteAsync<T>(Func<Task<T>> operation)
    {
        var attempt = 0;

        while (true)
        {
            try
            {
                return await operation();
            }
            catch (Exception ex) when (attempt < _maxRetries && _retryPredicate(ex))
            {
                attempt++;
                var delay = TimeSpan.FromMilliseconds(
                    _baseDelay.TotalMilliseconds * Math.Pow(2, attempt - 1));

                await Task.Delay(delay);
            }
        }
    }
}

// Usage with Polly library (recommended for production)
public class PollyRetryExample
{
    private readonly IAsyncPolicy _retryPolicy;

    public PollyRetryExample()
    {
        _retryPolicy = Policy
            .Handle<HttpRequestException>()
            .Or<TaskCanceledException>()
            .WaitAndRetryAsync(
                retryCount: 3,
                sleepDurationProvider: retryAttempt => TimeSpan.FromSeconds(Math.Pow(2, retryAttempt)),
                onRetry: (outcome, duration, retryCount, context) =>
                {
                    Console.WriteLine($"Retry {retryCount} after {duration}ms");
                });
    }

    public async Task<string> CallExternalServiceAsync()
    {
        return await _retryPolicy.ExecuteAsync(async () =>
        {
            using var client = new HttpClient();
            var response = await client.GetAsync("https://api.example.com/data");
            response.EnsureSuccessStatusCode();
            return await response.Content.ReadAsStringAsync();
        });
    }
}
```

### Bulkhead Pattern

```csharp
public class BulkheadService
{
    private readonly SemaphoreSlim _criticalSemaphore;
    private readonly SemaphoreSlim _normalSemaphore;

    public BulkheadService()
    {
        // Separate resource pools for different priority operations
        _criticalSemaphore = new SemaphoreSlim(5, 5); // 5 slots for critical operations
        _normalSemaphore = new SemaphoreSlim(10, 10);  // 10 slots for normal operations
    }

    public async Task<T> ExecuteCriticalOperationAsync<T>(Func<Task<T>> operation)
    {
        await _criticalSemaphore.WaitAsync();
        try
        {
            return await operation();
        }
        finally
        {
            _criticalSemaphore.Release();
        }
    }

    public async Task<T> ExecuteNormalOperationAsync<T>(Func<Task<T>> operation)
    {
        await _normalSemaphore.WaitAsync();
        try
        {
            return await operation();
        }
        finally
        {
            _normalSemaphore.Release();
        }
    }
}
```

### Saga Pattern for Error Compensation

```csharp
public class SagaTransaction
{
    private readonly List<(Func<Task> action, Func<Task> compensation)> _steps = new();
    private readonly List<Func<Task>> _executedCompensations = new();

    public SagaTransaction AddStep(Func<Task> action, Func<Task> compensation)
    {
        _steps.Add((action, compensation));
        return this;
    }

    public async Task ExecuteAsync()
    {
        try
        {
            foreach (var (action, compensation) in _steps)
            {
                await action();
                _executedCompensations.Add(compensation);
            }
        }
        catch (Exception)
        {
            await CompensateAsync();
            throw;
        }
    }

    private async Task CompensateAsync()
    {
        // Execute compensations in reverse order
        _executedCompensations.Reverse();

        foreach (var compensation in _executedCompensations)
        {
            try
            {
                await compensation();
            }
            catch (Exception ex)
            {
                // Log compensation failures but don't throw
                Console.WriteLine($"Compensation failed: {ex.Message}");
            }
        }
    }
}

// Usage
public class OrderService
{
    public async Task ProcessOrderAsync(Order order)
    {
        var saga = new SagaTransaction()
            .AddStep(
                action: () => _paymentService.ChargeAsync(order.PaymentInfo),
                compensation: () => _paymentService.RefundAsync(order.PaymentInfo))
            .AddStep(
                action: () => _inventoryService.ReserveAsync(order.Items),
                compensation: () => _inventoryService.ReleaseAsync(order.Items))
            .AddStep(
                action: () => _shippingService.ScheduleAsync(order),
                compensation: () => _shippingService.CancelAsync(order));

        await saga.ExecuteAsync();
    }
}
```

### Health Check Integration

```csharp
public class ServiceHealthCheck : IHealthCheck
{
    private readonly IExternalService _externalService;
    private readonly ILogger<ServiceHealthCheck> _logger;

    public ServiceHealthCheck(IExternalService externalService, ILogger<ServiceHealthCheck> logger)
    {
        _externalService = externalService;
        _logger = logger;
    }

    public async Task<HealthCheckResult> CheckHealthAsync(
        HealthCheckContext context,
        CancellationToken cancellationToken = default)
    {
        try
        {
            await _externalService.PingAsync(cancellationToken);
            return HealthCheckResult.Healthy("External service is responsive");
        }
        catch (OperationCanceledException)
        {
            return HealthCheckResult.Unhealthy("External service health check timed out");
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Health check failed for external service");
            return HealthCheckResult.Unhealthy("External service is not responding", ex);
        }
    }
}

// Registration
public void ConfigureServices(IServiceCollection services)
{
    services.AddHealthChecks()
        .AddCheck<ServiceHealthCheck>("external-service")
        .AddDbContext<AppDbContext>()
        .AddCheck<DbContextHealthCheck<AppDbContext>>("database");
}
```

---

## Summary and Best Practices

### Key Takeaways

1. **Use exceptions for exceptional cases, not control flow**
2. **Fail fast with proper validation**
3. **Preserve stack traces when rethrowing**
4. **Use specific exception types over generic ones**
5. **Implement proper cleanup with using statements**
6. **Consider functional error handling patterns for complex scenarios**
7. **Implement comprehensive logging and monitoring**
8. **Use modern C# features for safer code**
9. **Apply enterprise patterns for resilient systems**

### Exception Hierarchy Design

```csharp
// Good exception hierarchy for a domain
public abstract class DomainException : Exception
{
    protected DomainException(string message) : base(message) { }
    protected DomainException(string message, Exception innerException) : base(message, innerException) { }
}

public class BusinessRuleViolationException : DomainException
{
    public string RuleName { get; }

    public BusinessRuleViolationException(string ruleName, string message) : base(message)
    {
        RuleName = ruleName;
    }
}

public class EntityNotFoundException : DomainException
{
    public string EntityType { get; }
    public object EntityId { get; }

    public EntityNotFoundException(string entityType, object entityId)
        : base($"{entityType} with ID {entityId} was not found")
    {
        EntityType = entityType;
        EntityId = entityId;
    }
}
```

### Final Architecture Recommendation

For enterprise applications, combine multiple approaches:

1. **Use exceptions for truly exceptional cases**
2. **Implement Result pattern for expected failures**
3. **Use nullable reference types for null safety**
4. **Apply circuit breaker and retry patterns for external services**
5. **Implement comprehensive logging and monitoring**
6. **Use global exception handlers for unhandled exceptions**
7. **Write comprehensive tests for error scenarios**

This comprehensive approach ensures robust, maintainable, and observable error handling throughout your C# applications.