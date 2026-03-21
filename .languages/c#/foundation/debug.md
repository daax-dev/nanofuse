# C# Debugging Techniques and Tools

## Overview
Comprehensive guide to debugging C# applications across different environments, tools, and scenarios, from basic breakpoint debugging to advanced production diagnostics.

## Visual Studio Debugging

### Basic Debugging Features
```csharp
public class DebuggingExample
{
    public static void Main()
    {
        int result = CalculateSum(1, 2, 3, 4, 5);
        Console.WriteLine($"Result: {result}");

        // Set breakpoint here (F9)
        ProcessArray(new[] { 1, 2, 3, null, 5 });
    }

    private static int CalculateSum(params int[] numbers)
    {
        int sum = 0;
        foreach (int number in numbers)
        {
            sum += number; // Set conditional breakpoint: number > 3
        }
        return sum;
    }

    private static void ProcessArray(int?[] numbers)
    {
        for (int i = 0; i < numbers.Length; i++)
        {
            // Use DataTips to inspect values
            if (numbers[i].HasValue)
            {
                Console.WriteLine($"Value at {i}: {numbers[i].Value}");
            }
        }
    }
}
```

### Advanced Breakpoint Types
```csharp
public class AdvancedBreakpoints
{
    private List<string> _items = new List<string>();

    public void AddItem(string item)
    {
        // Conditional breakpoint: item.StartsWith("Error")
        if (string.IsNullOrEmpty(item))
            throw new ArgumentException("Item cannot be null or empty");

        _items.Add(item);

        // Hit count breakpoint: Break when hit 5 times
        Console.WriteLine($"Added item: {item}");
    }

    public void ProcessItems()
    {
        foreach (string item in _items)
        {
            // Tracepoint: Print "Processing: {item}" to output
            ProcessSingleItem(item);
        }
    }

    private void ProcessSingleItem(string item)
    {
        // Function breakpoint on method entry
        if (item.Contains("exception"))
        {
            throw new InvalidOperationException($"Error processing: {item}");
        }
    }
}
```

### Watch Windows and Expressions
```csharp
public class WatchExamples
{
    private Dictionary<string, int> _scores = new Dictionary<string, int>();

    public void AddScore(string player, int score)
    {
        _scores[player] = score;

        // Watch expressions to add:
        // _scores.Count
        // _scores.Keys.ToArray()
        // _scores.Values.Max()
        // _scores.Where(kvp => kvp.Value > 50)
    }

    public void AnalyzeScores()
    {
        var averageScore = _scores.Values.Average();
        var topPlayer = _scores.OrderByDescending(kvp => kvp.Value).First();

        // Immediate Window commands to try:
        // _scores.Add("TestPlayer", 100)
        // _scores.ContainsKey("TestPlayer")
        // topPlayer.Key + ": " + topPlayer.Value
    }
}
```

### Call Stack and Threads Debugging
```csharp
public class CallStackExample
{
    public async Task ProcessDataAsync()
    {
        await Task.Run(() => Level1());
    }

    private void Level1()
    {
        Level2("Parameter from Level1");
    }

    private void Level2(string parameter)
    {
        Level3(parameter.Length);
    }

    private void Level3(int value)
    {
        // Set breakpoint here
        // Use Call Stack window to navigate up the stack
        // Use Locals window to see variables at each level
        if (value > 10)
        {
            throw new InvalidOperationException("Value too large");
        }
    }
}
```

## VS Code Debugging

### Launch Configuration
```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": ".NET Core Launch (console)",
            "type": "coreclr",
            "request": "launch",
            "preLaunchTask": "build",
            "program": "${workspaceFolder}/bin/Debug/net8.0/MyApp.dll",
            "args": ["arg1", "arg2"],
            "cwd": "${workspaceFolder}",
            "console": "integratedTerminal",
            "stopAtEntry": false
        },
        {
            "name": ".NET Core Attach",
            "type": "coreclr",
            "request": "attach",
            "processId": "${command:pickProcess}"
        },
        {
            "name": "Debug Tests",
            "type": "coreclr",
            "request": "launch",
            "program": "dotnet",
            "args": ["test", "--logger", "console;verbosity=detailed"],
            "cwd": "${workspaceFolder}",
            "console": "integratedTerminal"
        }
    ]
}
```

### Tasks Configuration
```json
// .vscode/tasks.json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "build",
            "command": "dotnet",
            "type": "process",
            "args": ["build", "${workspaceFolder}/MyApp.csproj"],
            "problemMatcher": "$msCompile"
        },
        {
            "label": "watch",
            "command": "dotnet",
            "type": "process",
            "args": ["watch", "run"],
            "isBackground": true
        }
    ]
}
```

## Command-Line Debugging

### dotnet-dump Tool
```bash
# Install the tool
dotnet tool install -g dotnet-dump

# Create a dump of a running process
dotnet-dump collect -p <process-id> -o mydump.dmp

# Analyze the dump
dotnet-dump analyze mydump.dmp

# Commands within the analyzer:
# > threads - List all threads
# > clrstack - Show managed call stack
# > pe -lines - Show exception with line numbers
# > dumpheap -stat - Show heap statistics
# > gcroot <object-address> - Show GC roots for object
```

### dotnet-trace Tool
```bash
# Install the tool
dotnet tool install -g dotnet-trace

# Collect a trace
dotnet-trace collect -p <process-id> -o trace.nettrace

# Collect specific events
dotnet-trace collect -p <process-id> \
  --providers Microsoft-DotNETCore-SampleProfiler \
  --providers "Microsoft-AspNetCore-Server-Kestrel:0x1:Verbose"

# Convert to other formats
dotnet-trace convert trace.nettrace --format chromium
```

### dotnet-counters Tool
```bash
# Install the tool
dotnet tool install -g dotnet-counters

# Monitor performance counters
dotnet-counters monitor -p <process-id>

# Monitor specific counters
dotnet-counters monitor -p <process-id> \
  --counters System.Runtime,Microsoft.AspNetCore.Hosting

# Export counter data
dotnet-counters collect -p <process-id> -o counters.csv --format csv
```

## Remote and Production Debugging

### Application Insights Integration
```csharp
public class ProductionDebugging
{
    private readonly ILogger<ProductionDebugging> _logger;
    private readonly TelemetryClient _telemetryClient;

    public ProductionDebugging(ILogger<ProductionDebugging> logger,
                              TelemetryClient telemetryClient)
    {
        _logger = logger;
        _telemetryClient = telemetryClient;
    }

    public async Task ProcessOrderAsync(Order order)
    {
        using var activity = Activity.StartActivity("ProcessOrder");
        activity?.SetTag("order.id", order.Id.ToString());

        try
        {
            _logger.LogInformation("Processing order {OrderId}", order.Id);

            // Track custom metrics
            _telemetryClient.TrackMetric("OrderProcessing.Started", 1);

            var stopwatch = Stopwatch.StartNew();

            await ValidateOrder(order);
            await ProcessPayment(order);
            await FulfillOrder(order);

            stopwatch.Stop();

            // Track timing
            _telemetryClient.TrackMetric("OrderProcessing.Duration",
                                       stopwatch.ElapsedMilliseconds);

            _logger.LogInformation("Successfully processed order {OrderId} in {Duration}ms",
                                 order.Id, stopwatch.ElapsedMilliseconds);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to process order {OrderId}", order.Id);
            _telemetryClient.TrackException(ex);
            throw;
        }
    }

    private async Task ValidateOrder(Order order)
    {
        if (order.Items?.Any() != true)
        {
            var ex = new InvalidOperationException("Order has no items");
            ex.Data["OrderId"] = order.Id;
            ex.Data["CustomerId"] = order.CustomerId;
            throw ex;
        }
    }
}
```

### Structured Logging for Debugging
```csharp
public class StructuredLoggingExample
{
    private readonly ILogger<StructuredLoggingExample> _logger;

    public StructuredLoggingExample(ILogger<StructuredLoggingExample> logger)
    {
        _logger = logger;
    }

    public async Task ProcessUserRegistration(UserRegistrationRequest request)
    {
        using (_logger.BeginScope("UserRegistration {Email}", request.Email))
        {
            try
            {
                _logger.LogInformation("Starting user registration for {Email} with role {Role}",
                                     request.Email, request.Role);

                var validationResult = ValidateRequest(request);
                if (!validationResult.IsValid)
                {
                    _logger.LogWarning("Validation failed for {Email}: {Errors}",
                                     request.Email, validationResult.Errors);
                    return;
                }

                var user = await CreateUser(request);
                _logger.LogInformation("Created user {UserId} for {Email}",
                                     user.Id, request.Email);

                await SendWelcomeEmail(user);
                _logger.LogInformation("Sent welcome email to {Email}", request.Email);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to register user {Email}", request.Email);
                throw;
            }
        }
    }

    // Logging configuration in Program.cs or Startup.cs
    public static void ConfigureLogging(IServiceCollection services)
    {
        services.AddLogging(builder =>
        {
            builder.AddConsole(options =>
            {
                options.IncludeScopes = true;
                options.TimestampFormat = "yyyy-MM-dd HH:mm:ss ";
            });

            builder.AddDebug();

            // Add Serilog for structured logging
            builder.AddSerilog(new LoggerConfiguration()
                .WriteTo.Console(outputTemplate:
                    "[{Timestamp:HH:mm:ss} {Level:u3}] {Message:lj} {Properties:j}{NewLine}{Exception}")
                .WriteTo.File("logs/app-.txt", rollingInterval: RollingInterval.Day)
                .CreateLogger());
        });
    }
}
```

## Memory Debugging and Profiling

### Detecting Memory Leaks
```csharp
public class MemoryLeakDetection
{
    // Example of a memory leak - event handler not unsubscribed
    public class Publisher
    {
        public event Action<string> MessagePublished;

        public void PublishMessage(string message)
        {
            MessagePublished?.Invoke(message);
        }
    }

    public class Subscriber : IDisposable
    {
        private readonly Publisher _publisher;
        private bool _disposed;

        public Subscriber(Publisher publisher)
        {
            _publisher = publisher;
            _publisher.MessagePublished += HandleMessage;
        }

        private void HandleMessage(string message)
        {
            Console.WriteLine($"Received: {message}");
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                // Fix: Unsubscribe from event to prevent memory leak
                _publisher.MessagePublished -= HandleMessage;
                _disposed = true;
            }
        }
    }

    // Memory diagnostic helper
    public static void LogMemoryUsage(string context)
    {
        var process = Process.GetCurrentProcess();
        var workingSet = process.WorkingSet64;
        var privateMemory = process.PrivateMemorySize64;

        Console.WriteLine($"[{context}] Working Set: {workingSet / 1024 / 1024} MB, " +
                         $"Private Memory: {privateMemory / 1024 / 1024} MB");

        // Force GC and log again
        GC.Collect();
        GC.WaitForPendingFinalizers();
        GC.Collect();

        var afterGC = Process.GetCurrentProcess().WorkingSet64;
        Console.WriteLine($"[{context}] After GC: {afterGC / 1024 / 1024} MB");
    }
}
```

### Performance Debugging with BenchmarkDotNet
```csharp
[MemoryDiagnoser]
[SimpleJob(RuntimeMoniker.Net80)]
public class PerformanceDebugging
{
    private readonly List<int> _smallList = Enumerable.Range(1, 100).ToList();
    private readonly List<int> _largeList = Enumerable.Range(1, 10000).ToList();

    [Benchmark]
    [Arguments(100)]
    [Arguments(10000)]
    public int SumWithFor(int count)
    {
        var list = count == 100 ? _smallList : _largeList;
        int sum = 0;
        for (int i = 0; i < list.Count; i++)
        {
            sum += list[i];
        }
        return sum;
    }

    [Benchmark]
    [Arguments(100)]
    [Arguments(10000)]
    public int SumWithLinq(int count)
    {
        var list = count == 100 ? _smallList : _largeList;
        return list.Sum();
    }

    [Benchmark]
    [Arguments(100)]
    [Arguments(10000)]
    public int SumWithForeach(int count)
    {
        var list = count == 100 ? _smallList : _largeList;
        int sum = 0;
        foreach (int value in list)
        {
            sum += value;
        }
        return sum;
    }
}

// Run benchmark from Program.cs
// BenchmarkRunner.Run<PerformanceDebugging>();
```

## ASP.NET Core Debugging

### Debugging Middleware Pipeline
```csharp
public class DebuggingMiddleware
{
    private readonly RequestDelegate _next;
    private readonly ILogger<DebuggingMiddleware> _logger;

    public DebuggingMiddleware(RequestDelegate next, ILogger<DebuggingMiddleware> logger)
    {
        _next = next;
        _logger = logger;
    }

    public async Task InvokeAsync(HttpContext context)
    {
        var requestId = Guid.NewGuid().ToString();
        context.Items["RequestId"] = requestId;

        _logger.LogInformation("Request {RequestId} started: {Method} {Path}",
                             requestId, context.Request.Method, context.Request.Path);

        var stopwatch = Stopwatch.StartNew();

        try
        {
            await _next(context);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Request {RequestId} failed", requestId);
            throw;
        }
        finally
        {
            stopwatch.Stop();
            _logger.LogInformation("Request {RequestId} completed in {Duration}ms with status {StatusCode}",
                                 requestId, stopwatch.ElapsedMilliseconds, context.Response.StatusCode);
        }
    }
}

// Register in Program.cs
public static void Main(string[] args)
{
    var builder = WebApplication.CreateBuilder(args);
    var app = builder.Build();

    app.UseMiddleware<DebuggingMiddleware>();
    app.UseRouting();
    app.UseEndpoints(endpoints => endpoints.MapControllers());

    app.Run();
}
```

### API Debugging Tools
```csharp
[ApiController]
[Route("api/[controller]")]
public class DebuggingController : ControllerBase
{
    private readonly ILogger<DebuggingController> _logger;

    public DebuggingController(ILogger<DebuggingController> logger)
    {
        _logger = logger;
    }

    [HttpPost]
    public async Task<IActionResult> ProcessData([FromBody] ProcessDataRequest request)
    {
        // Log request details for debugging
        _logger.LogInformation("Processing data request: {@Request}", request);

        // Add request/response logging
        using (_logger.BeginScope("ProcessData {CorrelationId}", request.CorrelationId))
        {
            try
            {
                // Simulate processing with debugging info
                var result = await ProcessWithDebugging(request);

                _logger.LogInformation("Successfully processed data: {@Result}", result);
                return Ok(result);
            }
            catch (ValidationException ex)
            {
                _logger.LogWarning(ex, "Validation failed for request {CorrelationId}",
                                 request.CorrelationId);
                return BadRequest(ex.Message);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Unexpected error processing request {CorrelationId}",
                               request.CorrelationId);
                return StatusCode(500, "Internal server error");
            }
        }
    }

    private async Task<ProcessDataResponse> ProcessWithDebugging(ProcessDataRequest request)
    {
        // Add debug headers in development
        if (HttpContext.RequestServices.GetService<IWebHostEnvironment>().IsDevelopment())
        {
            HttpContext.Response.Headers.Add("X-Debug-CorrelationId", request.CorrelationId);
            HttpContext.Response.Headers.Add("X-Debug-ProcessedAt", DateTime.UtcNow.ToString("O"));
        }

        // Simulate async processing
        await Task.Delay(100);

        return new ProcessDataResponse
        {
            Id = Guid.NewGuid(),
            ProcessedAt = DateTime.UtcNow,
            Status = "Completed"
        };
    }
}
```

## Debugging Async/Await Issues

### Common Async Debugging Scenarios
```csharp
public class AsyncDebuggingExamples
{
    // Deadlock debugging example
    public class DeadlockExample
    {
        public string GetDataSync()
        {
            // This can cause deadlock in certain contexts (like ASP.NET Framework)
            return GetDataAsync().Result;
        }

        public async Task<string> GetDataAsync()
        {
            await Task.Delay(1000);
            return "Data";
        }

        // Better approach - use ConfigureAwait(false)
        public async Task<string> GetDataSafeAsync()
        {
            await Task.Delay(1000).ConfigureAwait(false);
            return "Data";
        }
    }

    // Exception handling in async methods
    public class AsyncExceptionHandling
    {
        public async Task ProcessMultipleItemsAsync(IEnumerable<string> items)
        {
            var tasks = items.Select(ProcessSingleItemAsync).ToArray();

            try
            {
                // This only catches the first exception
                await Task.WhenAll(tasks);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"First exception: {ex.Message}");

                // To get all exceptions:
                foreach (var task in tasks)
                {
                    if (task.IsFaulted)
                    {
                        Console.WriteLine($"Task exception: {task.Exception?.InnerException?.Message}");
                    }
                }
            }
        }

        private async Task ProcessSingleItemAsync(string item)
        {
            await Task.Delay(100);

            if (item.Contains("error"))
            {
                throw new InvalidOperationException($"Error processing {item}");
            }
        }
    }

    // Debugging async state machines
    public async Task<int> ComplexAsyncMethod(int value)
    {
        Console.WriteLine($"Method start: {value}");

        try
        {
            var result1 = await FirstAsyncOperation(value);
            Console.WriteLine($"After first operation: {result1}");

            var result2 = await SecondAsyncOperation(result1);
            Console.WriteLine($"After second operation: {result2}");

            return result2;
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Exception in async method: {ex}");
            throw;
        }
    }

    private async Task<int> FirstAsyncOperation(int value)
    {
        await Task.Delay(100);
        return value * 2;
    }

    private async Task<int> SecondAsyncOperation(int value)
    {
        await Task.Delay(100);
        if (value > 100)
            throw new ArgumentException("Value too large");
        return value + 10;
    }
}
```

## Advanced Debugging Scenarios

### Custom Debug Attributes
```csharp
[DebuggerDisplay("{GetDebuggerDisplay(),nq}")]
[DebuggerTypeProxy(typeof(PersonDebugView))]
public class Person
{
    public string FirstName { get; set; }
    public string LastName { get; set; }
    public int Age { get; set; }
    public List<string> Hobbies { get; set; } = new List<string>();

    private string GetDebuggerDisplay()
    {
        return $"{FirstName} {LastName} (Age: {Age})";
    }
}

// Custom debugger type proxy
internal class PersonDebugView
{
    private readonly Person _person;

    public PersonDebugView(Person person)
    {
        _person = person;
    }

    public string FullName => $"{_person.FirstName} {_person.LastName}";
    public int Age => _person.Age;
    public string[] Hobbies => _person.Hobbies.ToArray();
    public bool IsAdult => _person.Age >= 18;
}
```

### Conditional Compilation for Debugging
```csharp
public class ConditionalDebugging
{
    [Conditional("DEBUG")]
    public static void DebugLog(string message)
    {
        Console.WriteLine($"[DEBUG] {DateTime.Now:HH:mm:ss.fff}: {message}");
    }

    [Conditional("TRACE")]
    public static void TraceLog(string message)
    {
        Trace.WriteLine($"[TRACE] {DateTime.Now:HH:mm:ss.fff}: {message}");
    }

    public void ProcessData(string data)
    {
        DebugLog($"Starting to process: {data}");

        #if DEBUG
        var stopwatch = Stopwatch.StartNew();
        #endif

        // Processing logic here
        var result = data.ToUpper();

        #if DEBUG
        stopwatch.Stop();
        DebugLog($"Processing completed in {stopwatch.ElapsedMilliseconds}ms");
        #endif

        TraceLog($"Processed data: {result}");
    }
}
```

## Testing and Debugging Integration

### Debug Unit Tests
```csharp
[TestClass]
public class DebuggingTests
{
    [TestMethod]
    public void TestWithDebugging()
    {
        // Arrange
        var service = new DataService();
        var input = "test data";

        // Act
        var result = service.ProcessData(input);

        // Debug output
        TestContext.WriteLine($"Input: {input}");
        TestContext.WriteLine($"Result: {result}");

        // Assert
        Assert.IsNotNull(result);
        Assert.AreEqual("TEST DATA", result);
    }

    [TestMethod]
    [DataRow("hello", "HELLO")]
    [DataRow("world", "WORLD")]
    [DataRow("", "")]
    public void TestMultipleInputs(string input, string expected)
    {
        var service = new DataService();
        var result = service.ProcessData(input);

        TestContext.WriteLine($"Testing: '{input}' -> '{result}' (expected: '{expected}')");

        Assert.AreEqual(expected, result);
    }

    public TestContext TestContext { get; set; }
}
```

This comprehensive debugging guide covers the essential tools and techniques needed for effective C# debugging across different environments and scenarios.