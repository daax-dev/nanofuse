# C# Memory Management - Comprehensive Guide

## Table of Contents
1. [.NET Memory Model and CLR Basics](#net-memory-model-and-clr-basics)
2. [Garbage Collection (GC) in Detail](#garbage-collection-gc-in-detail)
3. [Value Types vs Reference Types](#value-types-vs-reference-types)
4. [Stack vs Heap Allocation](#stack-vs-heap-allocation)
5. [Memory Optimization Techniques](#memory-optimization-techniques)
6. [Span<T> and Memory<T> for Efficient Memory Usage](#spant-and-memoryt-for-efficient-memory-usage)
7. [IDisposable and Using Patterns](#idisposable-and-using-patterns)
8. [Memory Leaks and Prevention](#memory-leaks-and-prevention)
9. [Performance Profiling and Diagnostics](#performance-profiling-and-diagnostics)
10. [Advanced Memory Topics](#advanced-memory-topics-unsafe-code-pinning-interop)

## .NET Memory Model and CLR Basics

### CLR Memory Architecture

The Common Language Runtime (CLR) manages memory through several key components:

```csharp
// The CLR divides memory into different regions
public class CLRMemoryModel
{
    // 1. Stack: Method parameters, local variables, return addresses
    public void MethodWithStackVariables()
    {
        int localInt = 42;              // Stack-allocated
        DateTime localDate = DateTime.Now; // Stack-allocated (value type)
        string localString = "Hello";    // Reference on stack, object on heap
    }

    // 2. Heap: Object instances, arrays, delegates
    private List<string> _instanceField = new List<string>(); // Heap-allocated

    // 3. Static/Global heap: Static variables, type metadata
    private static Dictionary<string, int> _staticData = new();
}
```

### Memory Regions in Detail

#### Small Object Heap (SOH)
- Objects < 85,000 bytes
- Divided into generations (Gen 0, Gen 1, Gen 2)
- Subject to standard garbage collection

#### Large Object Heap (LOH)
- Objects >= 85,000 bytes
- Only collected during Gen 2 collections
- No compaction by default (until .NET 4.5.1)

```csharp
public class HeapAllocationExamples
{
    public void DemonstrateHeapTypes()
    {
        // Small Object Heap - typical objects
        var smallObject = new StringBuilder(1000);
        var normalArray = new int[1000]; // 4000 bytes

        // Large Object Heap - large arrays/objects
        var largeArray = new byte[85_000]; // Goes to LOH
        var hugeArray = new double[20_000]; // 160,000 bytes -> LOH

        // Check if object is in LOH
        bool isInLOH = GC.GetGeneration(largeArray) >= 2;
        Console.WriteLine($"Large array in LOH: {isInLOH}");
    }
}
```

#### Pinned Object Heap (POH) - .NET 5+
- Objects that need to remain at fixed memory addresses
- Reduces fragmentation in LOH

```csharp
// .NET 5+ POH example
public unsafe class PinnedObjectHeapExample
{
    public void UsePinnedObjects()
    {
        // Objects allocated here won't move during GC
        var pinnedArray = GC.AllocateArray<byte>(1024, pinned: true);

        fixed (byte* ptr = pinnedArray)
        {
            // Can safely use pointer - object won't move
            *ptr = 42;
        }
    }
}
```

### CLR Type System

```csharp
public class CLRTypeSystem
{
    // Value types - stored directly where declared
    struct Point // Value type
    {
        public int X, Y;

        public Point(int x, int y)
        {
            X = x;
            Y = y;
        }
    }

    // Reference types - object stored on heap, reference on stack
    class Circle // Reference type
    {
        public Point Center { get; set; }
        public int Radius { get; set; }

        public Circle(Point center, int radius)
        {
            Center = center; // Value type copied
            Radius = radius;
        }
    }

    public void DemonstrateTypeSystem()
    {
        // Value type - entire struct on stack
        Point p1 = new Point(10, 20);
        Point p2 = p1; // Copy of entire struct
        p2.X = 30; // p1.X remains 10

        // Reference type - reference on stack, object on heap
        Circle c1 = new Circle(p1, 5);
        Circle c2 = c1; // Copy of reference, same object
        c2.Radius = 10; // c1.Radius is now also 10
    }
}
```

## Garbage Collection (GC) in Detail

### Generational Garbage Collection

The .NET GC uses a generational approach based on the weak generational hypothesis:
- Most objects die young
- Older objects tend to live longer

```csharp
public class GenerationalGCDemo
{
    public class ShortLivedObject
    {
        private byte[] data = new byte[1024];
        public DateTime Created = DateTime.Now;
    }

    public class LongLivedObject
    {
        private static List<LongLivedObject> _staticList = new();
        private byte[] data = new byte[1024];

        public LongLivedObject()
        {
            _staticList.Add(this); // Keeps object alive
        }
    }

    public void DemonstrateGenerations()
    {
        Console.WriteLine("Initial memory usage:");
        PrintMemoryInfo();

        // Create many short-lived objects (will be in Gen 0)
        for (int i = 0; i < 10000; i++)
        {
            var temp = new ShortLivedObject();
            // temp becomes eligible for collection immediately
        }

        Console.WriteLine("\nAfter creating short-lived objects:");
        PrintMemoryInfo();

        // Force Gen 0 collection
        GC.Collect(0, GCCollectionMode.Forced);
        Console.WriteLine("\nAfter Gen 0 collection:");
        PrintMemoryInfo();

        // Create long-lived objects
        var longLived = new List<LongLivedObject>();
        for (int i = 0; i < 1000; i++)
        {
            longLived.Add(new LongLivedObject());
        }

        // These objects will survive Gen 0 and move to Gen 1
        GC.Collect(0, GCCollectionMode.Forced);
        GC.Collect(1, GCCollectionMode.Forced);

        Console.WriteLine("\nAfter creating long-lived objects and Gen 1 collection:");
        PrintMemoryInfo();
    }

    private void PrintMemoryInfo()
    {
        Console.WriteLine($"Gen 0 collections: {GC.CollectionCount(0)}");
        Console.WriteLine($"Gen 1 collections: {GC.CollectionCount(1)}");
        Console.WriteLine($"Gen 2 collections: {GC.CollectionCount(2)}");
        Console.WriteLine($"Total memory: {GC.GetTotalMemory(false):N0} bytes");
    }
}
```

### GC Workstation vs Server Modes

```csharp
public class GCModeConfiguration
{
    public void ConfigureGC()
    {
        // Check current GC mode
        bool isServerGC = GCSettings.IsServerGC;
        Console.WriteLine($"Server GC: {isServerGC}");

        // Check latency mode
        GCLatencyMode latency = GCSettings.LatencyMode;
        Console.WriteLine($"Current latency mode: {latency}");

        // For low-latency scenarios
        if (RequiresLowLatency())
        {
            GCSettings.LatencyMode = GCLatencyMode.LowLatency;
        }

        // For sustained low-latency (use carefully)
        if (RequiresSustainedLowLatency())
        {
            GCSettings.LatencyMode = GCLatencyMode.SustainedLowLatency;
        }
    }

    // Configuration via app.config or .csproj
    /*
    <configuration>
      <runtime>
        <gcServer enabled="true" />
        <gcConcurrent enabled="true" />
      </runtime>
    </configuration>

    Or in .csproj:
    <PropertyGroup>
      <ServerGarbageCollection>true</ServerGarbageCollection>
    </PropertyGroup>
    */

    private bool RequiresLowLatency() => Environment.ProcessorCount > 1;
    private bool RequiresSustainedLowLatency() => false; // Very careful usage
}
```

### GC Notification and Monitoring

```csharp
public class GCMonitoring
{
    private static int _gen0Collections = 0;
    private static int _gen1Collections = 0;
    private static int _gen2Collections = 0;

    public void SetupGCMonitoring()
    {
        // Register for GC notifications (server GC only)
        if (GCSettings.IsServerGC)
        {
            GC.RegisterForFullGCNotification(10, 10);

            Task.Run(MonitorGC);
        }

        // Use GC.Collect notifications
        AppDomain.CurrentDomain.ProcessExit += OnProcessExit;
    }

    private async Task MonitorGC()
    {
        while (true)
        {
            GCNotificationStatus status = GC.WaitForFullGCApproach();
            if (status == GCNotificationStatus.Succeeded)
            {
                Console.WriteLine("Full GC approaching - consider reducing allocations");

                // Wait for the GC to complete
                status = GC.WaitForFullGCComplete();
                if (status == GCNotificationStatus.Succeeded)
                {
                    Console.WriteLine("Full GC completed");
                    LogGCStats();
                }
            }

            await Task.Delay(100);
        }
    }

    private void OnProcessExit(object sender, EventArgs e)
    {
        LogGCStats();
        GC.CancelFullGCNotification();
    }

    public void LogGCStats()
    {
        int currentGen0 = GC.CollectionCount(0);
        int currentGen1 = GC.CollectionCount(1);
        int currentGen2 = GC.CollectionCount(2);

        Console.WriteLine($"GC Stats since last check:");
        Console.WriteLine($"  Gen 0: {currentGen0 - _gen0Collections} collections");
        Console.WriteLine($"  Gen 1: {currentGen1 - _gen1Collections} collections");
        Console.WriteLine($"  Gen 2: {currentGen2 - _gen2Collections} collections");
        Console.WriteLine($"  Total Memory: {GC.GetTotalMemory(false):N0} bytes");

        _gen0Collections = currentGen0;
        _gen1Collections = currentGen1;
        _gen2Collections = currentGen2;
    }
}
```

## Value Types vs Reference Types

### Fundamental Differences

```csharp
public class ValueVsReferenceTypes
{
    // Value types
    public struct ValueTypeExample
    {
        public int Number;
        public string Text; // Reference to string object

        public ValueTypeExample(int number, string text)
        {
            Number = number;
            Text = text;
        }

        // Value types can have methods
        public void ModifyNumber(int newValue)
        {
            Number = newValue;
        }
    }

    // Reference types
    public class ReferenceTypeExample
    {
        public int Number { get; set; }
        public string Text { get; set; }

        public ReferenceTypeExample(int number, string text)
        {
            Number = number;
            Text = text;
        }
    }

    public void DemonstratePassingBehavior()
    {
        // Value type passing
        ValueTypeExample valueStruct = new ValueTypeExample(10, "Original");
        ModifyValueType(valueStruct);
        Console.WriteLine($"After method: {valueStruct.Number}"); // Still 10

        ModifyValueTypeByRef(ref valueStruct);
        Console.WriteLine($"After ref method: {valueStruct.Number}"); // Now 99

        // Reference type passing
        ReferenceTypeExample refClass = new ReferenceTypeExample(10, "Original");
        ModifyReferenceType(refClass);
        Console.WriteLine($"After method: {refClass.Number}"); // Now 99

        ReassignReferenceType(refClass);
        Console.WriteLine($"After reassign: {refClass.Text}"); // Still "Modified"

        ReassignReferenceTypeByRef(ref refClass);
        Console.WriteLine($"After ref reassign: {refClass.Text}"); // Now "New Object"
    }

    private void ModifyValueType(ValueTypeExample value)
    {
        value.Number = 99; // Modifies copy, not original
    }

    private void ModifyValueTypeByRef(ref ValueTypeExample value)
    {
        value.Number = 99; // Modifies original
    }

    private void ModifyReferenceType(ReferenceTypeExample reference)
    {
        reference.Number = 99; // Modifies original object
        reference.Text = "Modified";
    }

    private void ReassignReferenceType(ReferenceTypeExample reference)
    {
        reference = new ReferenceTypeExample(999, "New Object"); // Only changes local copy of reference
    }

    private void ReassignReferenceTypeByRef(ref ReferenceTypeExample reference)
    {
        reference = new ReferenceTypeExample(999, "New Object"); // Changes original reference
    }
}
```

### Boxing and Unboxing

```csharp
public class BoxingUnboxingDemo
{
    public void DemonstrateBoxing()
    {
        // Boxing - value type to reference type
        int valueInt = 42;
        object boxedInt = valueInt; // Boxing occurs - heap allocation

        // This creates a new object on the heap containing the int value
        Console.WriteLine($"Original: {valueInt}, Boxed: {boxedInt}");

        // Unboxing - reference type to value type
        int unboxedInt = (int)boxedInt; // Unboxing - must cast to exact type

        // Performance implications
        MeasureBoxingPerformance();
    }

    private void MeasureBoxingPerformance()
    {
        const int iterations = 1_000_000;

        // Scenario 1: Frequent boxing (bad performance)
        var stopwatch = System.Diagnostics.Stopwatch.StartNew();
        var list = new ArrayList(); // Non-generic, causes boxing

        for (int i = 0; i < iterations; i++)
        {
            list.Add(i); // Boxing int to object
        }
        stopwatch.Stop();
        Console.WriteLine($"Boxing time: {stopwatch.ElapsedMilliseconds} ms");

        // Scenario 2: Generic collections (no boxing)
        stopwatch.Restart();
        var genericList = new List<int>(); // Generic, no boxing

        for (int i = 0; i < iterations; i++)
        {
            genericList.Add(i); // No boxing
        }
        stopwatch.Stop();
        Console.WriteLine($"Generic time: {stopwatch.ElapsedMilliseconds} ms");

        // Memory pressure from boxing
        long memoryBefore = GC.GetTotalMemory(true);
        var boxingList = new List<object>();
        for (int i = 0; i < 10000; i++)
        {
            boxingList.Add(i); // Each int gets boxed
        }
        long memoryAfter = GC.GetTotalMemory(false);
        Console.WriteLine($"Memory from boxing: {memoryAfter - memoryBefore} bytes");
    }

    // Avoid boxing with generics
    public T ProcessValue<T>(T value) where T : struct
    {
        // No boxing occurs here
        return value;
    }

    // Boxing happens with object parameters
    public object ProcessValueWithBoxing(object value)
    {
        return value; // Boxing if value type passed
    }
}
```

### Readonly Structs and In Parameters

```csharp
public readonly struct ReadonlyStructExample
{
    public readonly int X;
    public readonly int Y;

    public ReadonlyStructExample(int x, int y)
    {
        X = x;
        Y = y;
    }

    // All members must be readonly in readonly struct
    public int CalculateDistance() => (int)Math.Sqrt(X * X + Y * Y);
}

public class InParameterDemo
{
    // Using 'in' prevents copying large structs
    public void ProcessLargeStruct(in ReadonlyStructExample largeStruct)
    {
        // largeStruct is passed by reference, but cannot be modified
        var distance = largeStruct.CalculateDistance();
        Console.WriteLine($"Distance: {distance}");

        // This would cause compiler error:
        // largeStruct.X = 10; // Cannot modify readonly struct
    }

    public void DemonstrateInParameter()
    {
        var point = new ReadonlyStructExample(3, 4);
        ProcessLargeStruct(in point); // Pass by reference, no copy

        // Can also use without 'in' - compiler may optimize
        ProcessLargeStruct(point);
    }
}
```

## Stack vs Heap Allocation

### Stack Allocation Details

```csharp
public class StackAllocationDemo
{
    public void DemonstrateStackAllocation()
    {
        // Stack allocation - automatic cleanup
        int stackInt = 42; // 4 bytes on stack
        DateTime stackDateTime = DateTime.Now; // ~8 bytes on stack
        decimal stackDecimal = 123.45m; // 16 bytes on stack

        // Local reference variables on stack, objects on heap
        string stackReference = "Hello"; // Reference ~8 bytes on stack, string object on heap
        var stackList = new List<int>(); // Reference on stack, List object on heap

        // Method call creates new stack frame
        ProcessStackData(stackInt, stackDateTime);

        // When method exits, stack frame is automatically cleaned up
        // No GC needed for stack-allocated data
    }

    private void ProcessStackData(int value, DateTime date)
    {
        // New stack frame created
        int localValue = value * 2; // Stack allocation
        string localString = date.ToString(); // Reference on stack, string on heap

        // When method returns, this stack frame is popped
        // localValue and localString reference are automatically cleaned up
    }

    // stackalloc for arrays (unsafe context not required in newer C#)
    public void DemonstrateStackalloc()
    {
        // Stack-allocated array - no GC pressure
        Span<int> stackArray = stackalloc int[1000]; // 4000 bytes on stack

        // Initialize
        for (int i = 0; i < stackArray.Length; i++)
        {
            stackArray[i] = i * i;
        }

        // Process data
        int sum = 0;
        foreach (int value in stackArray)
        {
            sum += value;
        }

        Console.WriteLine($"Sum: {sum}");

        // No cleanup needed - automatically reclaimed when method exits
        // Much faster than heap allocation for temporary arrays
    }

    // Comparing stack vs heap allocation performance
    public void CompareAllocationPerformance()
    {
        const int iterations = 100_000;

        // Heap allocation
        var sw = System.Diagnostics.Stopwatch.StartNew();
        for (int i = 0; i < iterations; i++)
        {
            var heapArray = new int[100];
            ProcessArray(heapArray);
        }
        sw.Stop();
        Console.WriteLine($"Heap allocation: {sw.ElapsedMilliseconds} ms");

        // Stack allocation
        sw.Restart();
        for (int i = 0; i < iterations; i++)
        {
            Span<int> stackArray = stackalloc int[100];
            ProcessArray(stackArray);
        }
        sw.Stop();
        Console.WriteLine($"Stack allocation: {sw.ElapsedMilliseconds} ms");
    }

    private void ProcessArray(Span<int> array)
    {
        for (int i = 0; i < array.Length; i++)
        {
            array[i] = i;
        }
    }
}
```

### Heap Allocation Patterns

```csharp
public class HeapAllocationPatterns
{
    // Different heap allocation scenarios
    public void DemonstrateHeapAllocations()
    {
        // Small object heap allocations
        var smallObject = new StringBuilder(100); // SOH
        var smallArray = new byte[1000]; // SOH

        // Large object heap allocations
        var largeArray = new byte[85_000]; // LOH - not compacted by default
        var hugeObject = new int[50_000]; // LOH - 200,000 bytes

        // String allocations
        string literalString = "Hello World"; // Interned, shared reference
        string concatenated = "Hello" + " " + "World"; // New object on heap
        string interpolated = $"Value: {42}"; // New object on heap

        // Collection allocations
        var list = new List<int>(1000); // Capacity affects initial allocation
        var dictionary = new Dictionary<string, int>(); // Internal arrays allocated

        // Analyze generations
        AnalyzeObjectGenerations(smallObject, largeArray, literalString);
    }

    private void AnalyzeObjectGenerations(params object[] objects)
    {
        foreach (var obj in objects)
        {
            int generation = GC.GetGeneration(obj);
            Console.WriteLine($"{obj.GetType().Name}: Generation {generation}");
        }
    }

    // Heap fragmentation demonstration
    public void DemonstrateFragmentation()
    {
        var objects = new List<byte[]>();

        // Create fragmentation by allocating various sizes
        Random rand = new Random();
        for (int i = 0; i < 1000; i++)
        {
            int size = rand.Next(1000, 50000);
            objects.Add(new byte[size]);
        }

        // Remove every other object to create holes
        for (int i = objects.Count - 1; i >= 0; i -= 2)
        {
            objects.RemoveAt(i);
        }

        // Force garbage collection
        GC.Collect();
        GC.WaitForPendingFinalizers();
        GC.Collect();

        Console.WriteLine("Fragmentation created - heap has holes");

        // Allocate more objects - may not fit in holes
        for (int i = 0; i < 100; i++)
        {
            objects.Add(new byte[25000]);
        }

        Console.WriteLine("Memory after additional allocations");
    }
}
```

## Memory Optimization Techniques

### Object Pooling

```csharp
// Generic object pool implementation
public class ObjectPool<T> where T : class, new()
{
    private readonly ConcurrentQueue<T> _objects = new();
    private readonly Func<T> _objectGenerator;
    private readonly Action<T> _resetAction;
    private readonly int _maxSize;

    public ObjectPool(int maxSize = 100, Func<T> objectGenerator = null, Action<T> resetAction = null)
    {
        _maxSize = maxSize;
        _objectGenerator = objectGenerator ?? (() => new T());
        _resetAction = resetAction;
    }

    public T Get()
    {
        if (_objects.TryDequeue(out T item))
        {
            return item;
        }

        return _objectGenerator();
    }

    public void Return(T item)
    {
        if (_objects.Count < _maxSize)
        {
            _resetAction?.Invoke(item);
            _objects.Enqueue(item);
        }
    }

    public int Count => _objects.Count;
}

// Example usage
public class ObjectPoolingExample
{
    private static readonly ObjectPool<StringBuilder> StringBuilderPool =
        new ObjectPool<StringBuilder>(
            maxSize: 50,
            objectGenerator: () => new StringBuilder(256),
            resetAction: sb => sb.Clear()
        );

    private static readonly ObjectPool<List<int>> ListPool =
        new ObjectPool<List<int>>(
            maxSize: 20,
            resetAction: list => list.Clear()
        );

    public string ProcessData(IEnumerable<int> data)
    {
        // Get pooled StringBuilder instead of creating new one
        StringBuilder sb = StringBuilderPool.Get();
        List<int> tempList = ListPool.Get();

        try
        {
            tempList.AddRange(data);
            tempList.Sort();

            sb.AppendLine("Sorted Data:");
            foreach (int value in tempList)
            {
                sb.AppendLine(value.ToString());
            }

            return sb.ToString();
        }
        finally
        {
            // Return objects to pool for reuse
            StringBuilderPool.Return(sb);
            ListPool.Return(tempList);
        }
    }

    // Microsoft's ObjectPool from Extensions
    public void UseBuiltInObjectPool()
    {
        var serviceProvider = new ServiceCollection()
            .AddSingleton<ObjectPoolProvider, DefaultObjectPoolProvider>()
            .BuildServiceProvider();

        var poolProvider = serviceProvider.GetService<ObjectPoolProvider>();
        var policy = new StringBuilderPooledObjectPolicy();
        var pool = poolProvider.Create(policy);

        var sb = pool.Get();
        try
        {
            sb.Append("Hello from pooled StringBuilder");
            // Use StringBuilder
        }
        finally
        {
            pool.Return(sb);
        }
    }
}

// Custom pooled object policy
public class StringBuilderPooledObjectPolicy : PooledObjectPolicy<StringBuilder>
{
    public override StringBuilder Create() => new StringBuilder(256);

    public override bool Return(StringBuilder obj)
    {
        if (obj.Capacity > 4096) // Don't pool very large StringBuilders
            return false;

        obj.Clear();
        return true;
    }
}
```

### Array Pooling

```csharp
public class ArrayPoolingExample
{
    // Use ArrayPool<T> for temporary arrays
    public void ProcessLargeDataSet(IEnumerable<int> data)
    {
        var pool = ArrayPool<int>.Shared;
        int[] rentedArray = null;

        try
        {
            // Rent array from pool instead of allocating
            int count = data.Count();
            rentedArray = pool.Rent(count); // May return larger array

            // Copy data
            int index = 0;
            foreach (int value in data)
            {
                if (index < rentedArray.Length)
                    rentedArray[index++] = value;
            }

            // Process array (only use 'count' elements, not full length)
            ProcessArray(rentedArray.AsSpan(0, count));
        }
        finally
        {
            // Always return array to pool
            if (rentedArray != null)
                pool.Return(rentedArray, clearArray: true);
        }
    }

    private void ProcessArray(Span<int> array)
    {
        // Sort using span (no additional allocation)
        array.Sort();

        // Calculate statistics
        int sum = 0;
        foreach (int value in array)
        {
            sum += value;
        }

        Console.WriteLine($"Average: {(double)sum / array.Length:F2}");
    }

    // Custom array pool for specific scenarios
    public class CustomArrayPool<T>
    {
        private readonly ConcurrentQueue<T[]>[] _buckets;
        private readonly int _maxArraysPerBucket;

        public CustomArrayPool(int maxArrayLength, int maxArraysPerBucket = 50)
        {
            int buckets = SelectBucketIndex(maxArrayLength);
            _buckets = new ConcurrentQueue<T[]>[buckets];
            _maxArraysPerBucket = maxArraysPerBucket;

            for (int i = 0; i < _buckets.Length; i++)
            {
                _buckets[i] = new ConcurrentQueue<T[]>();
            }
        }

        public T[] Rent(int minimumLength)
        {
            int bucketIndex = SelectBucketIndex(minimumLength);

            if (bucketIndex < _buckets.Length &&
                _buckets[bucketIndex].TryDequeue(out T[] array))
            {
                return array;
            }

            return new T[GetMaxSizeForBucket(bucketIndex)];
        }

        public void Return(T[] array, bool clearArray = false)
        {
            if (clearArray)
                Array.Clear(array, 0, array.Length);

            int bucketIndex = SelectBucketIndex(array.Length);
            if (bucketIndex < _buckets.Length &&
                _buckets[bucketIndex].Count < _maxArraysPerBucket)
            {
                _buckets[bucketIndex].Enqueue(array);
            }
        }

        private int SelectBucketIndex(int bufferSize)
        {
            return (int)Math.Log2(Math.Max(bufferSize, 16));
        }

        private int GetMaxSizeForBucket(int bucketIndex)
        {
            return 1 << (bucketIndex + 4); // Powers of 2 starting from 16
        }
    }
}
```

### String Optimization

```csharp
public class StringOptimization
{
    // String interning for frequently used strings
    private readonly Dictionary<string, string> _internedStrings = new();

    public void DemonstrateStringInterning()
    {
        // Literal strings are automatically interned
        string literal1 = "Hello World";
        string literal2 = "Hello World";
        bool sameReference = ReferenceEquals(literal1, literal2); // True

        // Runtime strings are not automatically interned
        string runtime1 = DateTime.Now.ToString();
        string runtime2 = DateTime.Now.ToString();
        bool differentReference = ReferenceEquals(runtime1, runtime2); // False

        // Manual interning
        string interned1 = string.Intern(runtime1);
        string interned2 = string.Intern(runtime2);
        bool internedSame = ReferenceEquals(interned1, interned2); // True if same value

        Console.WriteLine($"Literals same reference: {sameReference}");
        Console.WriteLine($"Runtime different reference: {differentReference}");
        Console.WriteLine($"Interned same reference: {internedSame}");
    }

    // String concatenation optimization
    public void OptimizedStringConcatenation()
    {
        var values = Enumerable.Range(1, 1000);

        // Bad: Multiple string concatenations
        string bad = "";
        var sw = System.Diagnostics.Stopwatch.StartNew();
        foreach (int value in values)
        {
            bad += value.ToString(); // Creates new string each time
        }
        sw.Stop();
        Console.WriteLine($"String concatenation: {sw.ElapsedMilliseconds} ms");

        // Better: StringBuilder
        sw.Restart();
        var sb = new StringBuilder(10000); // Pre-allocate capacity
        foreach (int value in values)
        {
            sb.Append(value);
        }
        string better = sb.ToString();
        sw.Stop();
        Console.WriteLine($"StringBuilder: {sw.ElapsedMilliseconds} ms");

        // Best: string.Join for simple scenarios
        sw.Restart();
        string best = string.Join("", values);
        sw.Stop();
        Console.WriteLine($"string.Join: {sw.ElapsedMilliseconds} ms");
    }

    // StringSegment for parsing without allocations
    public void UseStringSegments(string data)
    {
        // Split without creating new string objects
        var segments = new List<StringSegment>();

        int start = 0;
        for (int i = 0; i < data.Length; i++)
        {
            if (data[i] == ',')
            {
                segments.Add(new StringSegment(data, start, i - start));
                start = i + 1;
            }
        }

        // Add final segment
        if (start < data.Length)
        {
            segments.Add(new StringSegment(data, start, data.Length - start));
        }

        // Process segments without string allocation
        foreach (var segment in segments)
        {
            if (int.TryParse(segment.AsSpan(), out int value))
            {
                Console.WriteLine($"Parsed: {value}");
            }
        }
    }

    // Custom string segment implementation
    public readonly struct StringSegment
    {
        private readonly string _source;
        private readonly int _offset;
        private readonly int _length;

        public StringSegment(string source, int offset, int length)
        {
            _source = source ?? throw new ArgumentNullException(nameof(source));
            _offset = offset;
            _length = length;
        }

        public ReadOnlySpan<char> AsSpan() => _source.AsSpan(_offset, _length);
        public override string ToString() => _source.Substring(_offset, _length);

        public bool Equals(string other) => AsSpan().SequenceEqual(other.AsSpan());
    }
}
```

### Struct Optimization

```csharp
// Optimized struct design
[StructLayout(LayoutKind.Sequential, Pack = 1)] // Control memory layout
public readonly struct OptimizedPoint3D
{
    // Order fields by size (largest first) to minimize padding
    public readonly double Z;  // 8 bytes
    public readonly double Y;  // 8 bytes
    public readonly double X;  // 8 bytes
    // Total: 24 bytes (no padding needed)

    public OptimizedPoint3D(double x, double y, double z)
    {
        X = x;
        Y = y;
        Z = z;
    }

    // Implement IEquatable<T> to avoid boxing
    public bool Equals(OptimizedPoint3D other) =>
        X.Equals(other.X) && Y.Equals(other.Y) && Z.Equals(other.Z);

    public override bool Equals(object obj) =>
        obj is OptimizedPoint3D point && Equals(point);

    public override int GetHashCode() =>
        HashCode.Combine(X, Y, Z);

    // Efficient string representation
    public override string ToString() =>
        $"({X:F2}, {Y:F2}, {Z:F2})";
}

// Comparison with non-optimized struct
public struct NonOptimizedPoint3D
{
    // Poor field ordering creates padding
    public double X;  // 8 bytes
    public int Id;    // 4 bytes + 4 bytes padding
    public double Y;  // 8 bytes
    public byte Flag; // 1 byte + 7 bytes padding
    public double Z;  // 8 bytes
    // Total: 40 bytes (with padding)
}

public class StructOptimizationDemo
{
    public void CompareStructSizes()
    {
        unsafe
        {
            Console.WriteLine($"Optimized struct size: {sizeof(OptimizedPoint3D)} bytes");
            Console.WriteLine($"Non-optimized struct size: {sizeof(NonOptimizedPoint3D)} bytes");
        }

        // Memory usage comparison
        const int count = 1_000_000;

        var optimized = new OptimizedPoint3D[count];
        var nonOptimized = new NonOptimizedPoint3D[count];

        long optimizedBytes = count * Marshal.SizeOf<OptimizedPoint3D>();
        long nonOptimizedBytes = count * Marshal.SizeOf<NonOptimizedPoint3D>();

        Console.WriteLine($"Optimized array: {optimizedBytes:N0} bytes");
        Console.WriteLine($"Non-optimized array: {nonOptimizedBytes:N0} bytes");
        Console.WriteLine($"Savings: {nonOptimizedBytes - optimizedBytes:N0} bytes");
    }
}
```

## Span<T> and Memory<T> for Efficient Memory Usage

### Basic Span<T> Usage

```csharp
public class SpanBasicsDemo
{
    public void DemonstrateSpanBasics()
    {
        // Span over array
        int[] array = { 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 };
        Span<int> span = array.AsSpan();

        // Slicing without allocation
        Span<int> firstHalf = span.Slice(0, 5);    // [1, 2, 3, 4, 5]
        Span<int> secondHalf = span.Slice(5);      // [6, 7, 8, 9, 10]
        Span<int> middle = span.Slice(2, 6);       // [3, 4, 5, 6, 7, 8]

        // Direct modification
        firstHalf[0] = 99; // Modifies original array
        Console.WriteLine(array[0]); // Prints 99

        // Span over string
        string text = "Hello, World!";
        ReadOnlySpan<char> chars = text.AsSpan();
        ReadOnlySpan<char> hello = chars.Slice(0, 5); // "Hello"
        ReadOnlySpan<char> world = chars.Slice(7, 5); // "World"

        Console.WriteLine(hello.ToString());
        Console.WriteLine(world.ToString());
    }

    public void StackAllocatedSpan()
    {
        // Stack-allocated span
        Span<byte> buffer = stackalloc byte[256];

        // Initialize
        for (int i = 0; i < buffer.Length; i++)
        {
            buffer[i] = (byte)(i % 256);
        }

        // Process in chunks without allocation
        ProcessInChunks(buffer, chunkSize: 64);
    }

    private void ProcessInChunks(Span<byte> data, int chunkSize)
    {
        for (int i = 0; i < data.Length; i += chunkSize)
        {
            int length = Math.Min(chunkSize, data.Length - i);
            Span<byte> chunk = data.Slice(i, length);

            // Process chunk
            byte checksum = 0;
            foreach (byte b in chunk)
            {
                checksum ^= b;
            }

            Console.WriteLine($"Chunk {i / chunkSize}: checksum = {checksum:X2}");
        }
    }
}
```

### Memory<T> for Async Operations

```csharp
public class MemoryAsyncDemo
{
    // Memory<T> can be stored in fields and used in async methods
    private Memory<byte> _buffer;

    public MemoryAsyncDemo()
    {
        _buffer = new byte[1024];
    }

    public async Task ProcessDataAsync(Stream stream)
    {
        // Memory<T> works with async operations (Span<T> doesn't)
        int bytesRead = await stream.ReadAsync(_buffer);

        if (bytesRead > 0)
        {
            // Convert to Span for processing
            Span<byte> data = _buffer.Span.Slice(0, bytesRead);
            await ProcessBufferAsync(data);
        }
    }

    private async Task ProcessBufferAsync(ReadOnlyMemory<byte> data)
    {
        // Simulate async processing
        await Task.Delay(10);

        // Process data
        ReadOnlySpan<byte> span = data.Span;
        int sum = 0;
        foreach (byte b in span)
        {
            sum += b;
        }

        Console.WriteLine($"Processed {data.Length} bytes, sum: {sum}");
    }

    // Pooled memory for async operations
    public async Task UsePooledMemoryAsync()
    {
        using var memoryOwner = MemoryPool<byte>.Shared.Rent(4096);
        Memory<byte> memory = memoryOwner.Memory;

        // Use memory in async operations
        await FillMemoryAsync(memory);
        await ProcessMemoryAsync(memory);

        // Memory is automatically returned to pool when disposed
    }

    private async Task FillMemoryAsync(Memory<byte> memory)
    {
        Random rand = new Random();
        byte[] temp = new byte[1024];
        rand.NextBytes(temp);

        temp.AsSpan().CopyTo(memory.Span);
        await Task.Delay(1); // Simulate async work
    }

    private async Task ProcessMemoryAsync(ReadOnlyMemory<byte> memory)
    {
        // Process in chunks
        for (int offset = 0; offset < memory.Length; offset += 256)
        {
            int length = Math.Min(256, memory.Length - offset);
            ReadOnlyMemory<byte> chunk = memory.Slice(offset, length);

            // Simulate processing
            await Task.Delay(1);

            ReadOnlySpan<byte> span = chunk.Span;
            Console.WriteLine($"Chunk starting at {offset}: {span.Length} bytes");
        }
    }
}
```

### Span<T> for String Processing

```csharp
public class SpanStringProcessing
{
    // Parsing without allocations
    public List<int> ParseIntegers(ReadOnlySpan<char> input)
    {
        var result = new List<int>();

        while (!input.IsEmpty)
        {
            // Find next separator
            int separatorIndex = input.IndexOfAny(',', ';', '\t');

            ReadOnlySpan<char> token;
            if (separatorIndex >= 0)
            {
                token = input.Slice(0, separatorIndex);
                input = input.Slice(separatorIndex + 1);
            }
            else
            {
                token = input;
                input = ReadOnlySpan<char>.Empty;
            }

            // Trim whitespace (no allocation)
            token = token.Trim();

            // Parse directly from span
            if (!token.IsEmpty && int.TryParse(token, out int value))
            {
                result.Add(value);
            }
        }

        return result;
    }

    // String manipulation without allocation
    public void ProcessText(string text)
    {
        ReadOnlySpan<char> span = text.AsSpan();

        // Split lines without creating string array
        while (!span.IsEmpty)
        {
            int newlineIndex = span.IndexOf('\n');

            ReadOnlySpan<char> line;
            if (newlineIndex >= 0)
            {
                line = span.Slice(0, newlineIndex);
                span = span.Slice(newlineIndex + 1);
            }
            else
            {
                line = span;
                span = ReadOnlySpan<char>.Empty;
            }

            ProcessLine(line);
        }
    }

    private void ProcessLine(ReadOnlySpan<char> line)
    {
        // Trim without allocation
        line = line.Trim();

        if (line.IsEmpty || line[0] == '#') // Skip empty lines and comments
            return;

        // Check prefix without allocation
        if (line.StartsWith("ERROR:".AsSpan()))
        {
            var errorMessage = line.Slice(6).Trim();
            Console.WriteLine($"Found error: {errorMessage.ToString()}");
        }
        else if (line.StartsWith("WARN:".AsSpan()))
        {
            var warnMessage = line.Slice(5).Trim();
            Console.WriteLine($"Found warning: {warnMessage.ToString()}");
        }
    }

    // Format strings without allocation using interpolated string handlers
    public void EfficientStringFormatting()
    {
        const int iterations = 100_000;
        var values = Enumerable.Range(1, iterations).ToArray();

        // Traditional string interpolation (creates many strings)
        var sw = System.Diagnostics.Stopwatch.StartNew();
        for (int i = 0; i < iterations; i++)
        {
            string traditional = $"Value: {values[i]:N0}";
            // Process string...
        }
        sw.Stop();
        Console.WriteLine($"Traditional interpolation: {sw.ElapsedMilliseconds} ms");

        // Using StringBuilder with Span<char>
        sw.Restart();
        var sb = new StringBuilder(256);
        Span<char> buffer = stackalloc char[128];

        for (int i = 0; i < iterations; i++)
        {
            sb.Clear();
            sb.Append("Value: ");

            // Format directly to span
            if (values[i].TryFormat(buffer, out int charsWritten, "N0"))
            {
                sb.Append(buffer.Slice(0, charsWritten));
            }

            string efficient = sb.ToString();
            // Process string...
        }
        sw.Stop();
        Console.WriteLine($"Span-based formatting: {sw.ElapsedMilliseconds} ms");
    }
}
```

### Custom Memory Managers

```csharp
// Custom memory manager for specific scenarios
public class UnmanagedMemoryManager<T> : MemoryManager<T> where T : unmanaged
{
    private readonly unsafe T* _ptr;
    private readonly int _length;
    private bool _disposed;

    public unsafe UnmanagedMemoryManager(T* ptr, int length)
    {
        _ptr = ptr;
        _length = length;
    }

    public override Span<T> GetSpan()
    {
        if (_disposed)
            throw new ObjectDisposedException(nameof(UnmanagedMemoryManager<T>));

        unsafe
        {
            return new Span<T>(_ptr, _length);
        }
    }

    public override MemoryHandle Pin(int elementIndex = 0)
    {
        if (elementIndex < 0 || elementIndex >= _length)
            throw new ArgumentOutOfRangeException(nameof(elementIndex));

        unsafe
        {
            return new MemoryHandle(_ptr + elementIndex);
        }
    }

    public override void Unpin() { } // Memory is already pinned

    protected override void Dispose(bool disposing)
    {
        if (!_disposed)
        {
            unsafe
            {
                Marshal.FreeHGlobal((IntPtr)_ptr);
            }
            _disposed = true;
        }
    }
}

// Usage example
public class CustomMemoryManagerExample
{
    public unsafe void UseCustomMemoryManager()
    {
        const int size = 1024;

        // Allocate unmanaged memory
        T* ptr = (T*)Marshal.AllocHGlobal(size * sizeof(T));

        try
        {
            using var memoryManager = new UnmanagedMemoryManager<T>(ptr, size);
            Memory<T> memory = memoryManager.Memory;

            // Use memory normally
            Span<T> span = memory.Span;

            // Initialize data
            Random rand = new Random();
            for (int i = 0; i < span.Length; i++)
            {
                span[i] = (T)(object)rand.Next();
            }

            // Process data
            await ProcessMemoryAsync(memory);

        } // Memory manager disposes unmanaged memory
        catch
        {
            // Cleanup on exception
            Marshal.FreeHGlobal((IntPtr)ptr);
            throw;
        }
    }

    private async Task ProcessMemoryAsync<T>(ReadOnlyMemory<T> memory) where T : unmanaged
    {
        // Simulate async processing
        await Task.Delay(10);

        ReadOnlySpan<T> span = memory.Span;
        Console.WriteLine($"Processed {span.Length} elements");
    }
}
```

## IDisposable and Using Patterns

### Basic IDisposable Implementation

```csharp
public class BasicDisposableExample : IDisposable
{
    private FileStream _fileStream;
    private bool _disposed = false;

    public BasicDisposableExample(string filePath)
    {
        _fileStream = new FileStream(filePath, FileMode.Create);
    }

    public void WriteData(byte[] data)
    {
        ThrowIfDisposed();
        _fileStream.Write(data, 0, data.Length);
    }

    // Public Dispose method
    public void Dispose()
    {
        Dispose(disposing: true);
        GC.SuppressFinalize(this);
    }

    // Protected virtual method for inheritance
    protected virtual void Dispose(bool disposing)
    {
        if (!_disposed)
        {
            if (disposing)
            {
                // Dispose managed resources
                _fileStream?.Dispose();
            }

            // Dispose unmanaged resources (if any)
            // Set large fields to null
            _fileStream = null;

            _disposed = true;
        }
    }

    private void ThrowIfDisposed()
    {
        if (_disposed)
            throw new ObjectDisposedException(GetType().Name);
    }
}
```

### Advanced IDisposable with Finalizer

```csharp
public class AdvancedDisposableExample : IDisposable
{
    private IntPtr _unmanagedResource;
    private FileStream _managedResource;
    private bool _disposed = false;

    public AdvancedDisposableExample(string filePath)
    {
        _managedResource = new FileStream(filePath, FileMode.Create);
        _unmanagedResource = Marshal.AllocHGlobal(1024); // Simulate unmanaged resource
    }

    // Finalizer - only if class holds unmanaged resources
    ~AdvancedDisposableExample()
    {
        Dispose(disposing: false);
    }

    public void Dispose()
    {
        Dispose(disposing: true);
        GC.SuppressFinalize(this); // Suppress finalizer since we're disposing properly
    }

    protected virtual void Dispose(bool disposing)
    {
        if (!_disposed)
        {
            if (disposing)
            {
                // Dispose managed resources
                _managedResource?.Dispose();

                // Dispose other IDisposable fields here
            }

            // Always dispose unmanaged resources
            if (_unmanagedResource != IntPtr.Zero)
            {
                Marshal.FreeHGlobal(_unmanagedResource);
                _unmanagedResource = IntPtr.Zero;
            }

            _disposed = true;
        }
    }

    private void ThrowIfDisposed()
    {
        if (_disposed)
            throw new ObjectDisposedException(GetType().Name);
    }
}
```

### IAsyncDisposable Pattern (.NET Core 3.0+)

```csharp
public class AsyncDisposableExample : IAsyncDisposable, IDisposable
{
    private Stream _stream;
    private Timer _timer;
    private readonly SemaphoreSlim _semaphore;
    private bool _disposed = false;

    public AsyncDisposableExample()
    {
        _stream = new MemoryStream();
        _timer = new Timer(_ => { }, null, TimeSpan.Zero, TimeSpan.FromSeconds(1));
        _semaphore = new SemaphoreSlim(1, 1);
    }

    public async ValueTask DisposeAsync()
    {
        await DisposeAsyncCore();

        Dispose(disposing: false);
        GC.SuppressFinalize(this);
    }

    protected virtual async ValueTask DisposeAsyncCore()
    {
        if (_stream is not null)
        {
            await _stream.DisposeAsync().ConfigureAwait(false);
        }

        if (_timer is not null)
        {
            await _timer.DisposeAsync().ConfigureAwait(false);
        }

        _semaphore?.Dispose();
        _stream = null;
        _timer = null;
    }

    public void Dispose()
    {
        Dispose(disposing: true);
        GC.SuppressFinalize(this);
    }

    protected virtual void Dispose(bool disposing)
    {
        if (!_disposed)
        {
            if (disposing)
            {
                _stream?.Dispose();
                _timer?.Dispose();
                _semaphore?.Dispose();
            }

            _disposed = true;
        }
    }
}

// Usage of async disposable
public class AsyncDisposableUsage
{
    public async Task UseAsyncDisposableAsync()
    {
        // C# 8.0 async using
        await using var asyncResource = new AsyncDisposableExample();

        // Use resource...

        // DisposeAsync called automatically
    }

    public async Task ManualAsyncDispose()
    {
        var resource = new AsyncDisposableExample();
        try
        {
            // Use resource...
        }
        finally
        {
            await resource.DisposeAsync();
        }
    }
}
```

### Using Statements and Patterns

```csharp
public class UsingPatternExamples
{
    // Traditional using statement
    public void TraditionalUsing()
    {
        using (var fileStream = new FileStream("test.txt", FileMode.Create))
        {
            var data = System.Text.Encoding.UTF8.GetBytes("Hello World");
            fileStream.Write(data, 0, data.Length);
        } // Dispose called automatically
    }

    // C# 8.0 using declarations
    public void UsingDeclarations()
    {
        using var fileStream = new FileStream("test.txt", FileMode.Create);
        using var writer = new StreamWriter(fileStream);

        writer.WriteLine("Hello World");
        writer.WriteLine("Second line");

        // Both writer and fileStream disposed at end of method
    }

    // Multiple resources
    public void MultipleResources()
    {
        using var connection = new SqlConnection("connection string");
        using var command = new SqlCommand("SELECT * FROM Users", connection);
        using var reader = connection.ExecuteReader();

        while (reader.Read())
        {
            // Process data
        }

        // All resources disposed in reverse order
    }

    // Custom disposable struct (ref struct)
    public ref struct CustomDisposable
    {
        private Span<byte> _buffer;
        private bool _disposed;

        public CustomDisposable(Span<byte> buffer)
        {
            _buffer = buffer;
            _disposed = false;
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                // Cleanup logic
                _buffer.Clear();
                _disposed = true;
            }
        }
    }

    public void UseCustomDisposable()
    {
        Span<byte> buffer = stackalloc byte[1024];
        using var disposable = new CustomDisposable(buffer);

        // Use disposable...

        // Dispose called automatically
    }
}
```

### Resource Management Best Practices

```csharp
public class ResourceManagementBestPractices
{
    // Factory pattern with proper disposal
    public static class ResourceFactory
    {
        public static T CreateResource<T>(Func<T> factory) where T : IDisposable
        {
            try
            {
                return factory();
            }
            catch
            {
                // If factory throws, ensure any partial resources are cleaned up
                throw;
            }
        }

        public static async Task<T> CreateResourceAsync<T>(Func<Task<T>> factory)
            where T : IDisposable
        {
            try
            {
                return await factory();
            }
            catch
            {
                throw;
            }
        }
    }

    // Exception-safe resource acquisition
    public void SafeResourceAcquisition()
    {
        FileStream fileStream = null;
        StreamWriter writer = null;

        try
        {
            fileStream = new FileStream("test.txt", FileMode.Create);
            writer = new StreamWriter(fileStream);

            writer.WriteLine("Safe resource management");
        }
        catch
        {
            // Cleanup in reverse order of acquisition
            writer?.Dispose();
            fileStream?.Dispose();
            throw;
        }
        finally
        {
            writer?.Dispose();
            // fileStream disposed by StreamWriter
        }
    }

    // Better approach with using statements
    public void BetterResourceAcquisition()
    {
        using var fileStream = new FileStream("test.txt", FileMode.Create);
        using var writer = new StreamWriter(fileStream);

        writer.WriteLine("Better resource management");

        // Automatic cleanup in correct order
    }

    // Disposable wrapper for non-disposable resources
    public class DisposableWrapper<T> : IDisposable where T : class
    {
        private T _resource;
        private readonly Action<T> _cleanup;
        private bool _disposed;

        public DisposableWrapper(T resource, Action<T> cleanup)
        {
            _resource = resource ?? throw new ArgumentNullException(nameof(resource));
            _cleanup = cleanup ?? throw new ArgumentNullException(nameof(cleanup));
        }

        public T Resource
        {
            get
            {
                ThrowIfDisposed();
                return _resource;
            }
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                try
                {
                    _cleanup(_resource);
                }
                finally
                {
                    _resource = null;
                    _disposed = true;
                }
            }
        }

        private void ThrowIfDisposed()
        {
            if (_disposed)
                throw new ObjectDisposedException(GetType().Name);
        }
    }

    // Usage of disposable wrapper
    public void UseDisposableWrapper()
    {
        // Wrap a non-disposable resource that needs cleanup
        var unmanagedHandle = Marshal.AllocHGlobal(1024);

        using var wrapper = new DisposableWrapper<IntPtr>(
            unmanagedHandle,
            handle => Marshal.FreeHGlobal(handle)
        );

        // Use wrapper.Resource
        unsafe
        {
            byte* ptr = (byte*)wrapper.Resource;
            *ptr = 42;
        }

        // Automatic cleanup when disposed
    }
}
```

## Memory Leaks and Prevention

### Common Memory Leak Scenarios

```csharp
public class MemoryLeakExamples
{
    // 1. Event handler leaks
    public class EventPublisher
    {
        public event EventHandler<string> SomethingHappened;

        public void TriggerEvent(string message)
        {
            SomethingHappened?.Invoke(this, message);
        }
    }

    public class EventSubscriber : IDisposable
    {
        private EventPublisher _publisher;
        private bool _disposed;

        public EventSubscriber(EventPublisher publisher)
        {
            _publisher = publisher;
            _publisher.SomethingHappened += OnSomethingHappened; // Creates reference
        }

        private void OnSomethingHappened(object sender, string message)
        {
            Console.WriteLine($"Received: {message}");
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                // IMPORTANT: Unsubscribe to prevent leak
                _publisher.SomethingHappened -= OnSomethingHappened;
                _publisher = null;
                _disposed = true;
            }
        }
    }

    // 2. Timer leaks
    public class TimerLeakExample : IDisposable
    {
        private Timer _timer;
        private bool _disposed;

        public TimerLeakExample()
        {
            // Timer holds reference to callback, preventing GC
            _timer = new Timer(OnTimerElapsed, null, TimeSpan.Zero, TimeSpan.FromSeconds(1));
        }

        private void OnTimerElapsed(object state)
        {
            // Timer callback work
            Console.WriteLine("Timer elapsed");
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                _timer?.Dispose(); // MUST dispose timer to prevent leak
                _timer = null;
                _disposed = true;
            }
        }
    }

    // 3. Static collection leaks
    public class StaticCollectionLeak
    {
        private static readonly List<WeakReference> _instances = new List<WeakReference>();
        private byte[] _data;

        public StaticCollectionLeak()
        {
            _data = new byte[1024 * 1024]; // 1MB

            // BAD: Static collection prevents GC
            // _instances.Add(this);

            // BETTER: Use WeakReference
            _instances.Add(new WeakReference(this));

            // Cleanup dead references periodically
            CleanupDeadReferences();
        }

        private static void CleanupDeadReferences()
        {
            _instances.RemoveAll(wr => !wr.IsAlive);
        }
    }

    // 4. Circular reference leak (rare with modern GC)
    public class CircularReferenceExample
    {
        public CircularReferenceExample Parent { get; set; }
        public List<CircularReferenceExample> Children { get; set; } = new List<CircularReferenceExample>();

        public void AddChild(CircularReferenceExample child)
        {
            child.Parent = this; // Circular reference
            Children.Add(child);
        }

        // Modern GC handles circular references, but be careful with unmanaged resources
    }
}
```

### Memory Leak Detection and Prevention

```csharp
public class MemoryLeakDetection
{
    // WeakReference pattern for caches
    public class WeakCache<TKey, TValue> where TValue : class
    {
        private readonly ConcurrentDictionary<TKey, WeakReference<TValue>> _cache
            = new ConcurrentDictionary<TKey, WeakReference<TValue>>();

        public bool TryGet(TKey key, out TValue value)
        {
            value = null;

            if (_cache.TryGetValue(key, out var weakRef))
            {
                if (weakRef.TryGetTarget(out value))
                {
                    return true;
                }
                else
                {
                    // Remove dead reference
                    _cache.TryRemove(key, out _);
                }
            }

            return false;
        }

        public void Set(TKey key, TValue value)
        {
            _cache[key] = new WeakReference<TValue>(value);
        }

        public void Cleanup()
        {
            var keysToRemove = new List<TKey>();

            foreach (var kvp in _cache)
            {
                if (!kvp.Value.TryGetTarget(out _))
                {
                    keysToRemove.Add(kvp.Key);
                }
            }

            foreach (var key in keysToRemove)
            {
                _cache.TryRemove(key, out _);
            }
        }
    }

    // Memory monitoring
    public class MemoryMonitor
    {
        private readonly Timer _monitor;
        private long _lastMemoryUsage;
        private readonly List<long> _memoryHistory = new List<long>();

        public MemoryMonitor()
        {
            _monitor = new Timer(CheckMemory, null, TimeSpan.Zero, TimeSpan.FromSeconds(10));
        }

        private void CheckMemory(object state)
        {
            long currentMemory = GC.GetTotalMemory(false);
            long memoryIncrease = currentMemory - _lastMemoryUsage;

            _memoryHistory.Add(currentMemory);
            if (_memoryHistory.Count > 100) // Keep last 100 readings
            {
                _memoryHistory.RemoveAt(0);
            }

            // Detect potential memory leaks
            if (memoryIncrease > 10 * 1024 * 1024) // 10MB increase
            {
                Console.WriteLine($"WARNING: Large memory increase detected: {memoryIncrease / 1024 / 1024:F2} MB");

                // Force GC to see if it's actually a leak
                GC.Collect();
                GC.WaitForPendingFinalizers();
                GC.Collect();

                long afterGC = GC.GetTotalMemory(false);
                long actualIncrease = afterGC - _lastMemoryUsage;

                if (actualIncrease > 5 * 1024 * 1024) // Still 5MB after GC
                {
                    Console.WriteLine($"ALERT: Potential memory leak - {actualIncrease / 1024 / 1024:F2} MB not collected");
                    LogMemoryDump();
                }
            }

            _lastMemoryUsage = currentMemory;
        }

        private void LogMemoryDump()
        {
            Console.WriteLine("Memory Statistics:");
            Console.WriteLine($"Total Memory: {GC.GetTotalMemory(false) / 1024 / 1024:F2} MB");
            Console.WriteLine($"Gen 0 Collections: {GC.CollectionCount(0)}");
            Console.WriteLine($"Gen 1 Collections: {GC.CollectionCount(1)}");
            Console.WriteLine($"Gen 2 Collections: {GC.CollectionCount(2)}");

            // In production, consider writing to log file or monitoring system
        }

        public void Dispose()
        {
            _monitor?.Dispose();
        }
    }

    // Finalizer tracking for leak detection
    public class FinalizerTracker
    {
        private static readonly ConcurrentBag<string> _finalizerLog = new ConcurrentBag<string>();

        private readonly string _typeName;
        private readonly DateTime _created;

        public FinalizerTracker(string typeName)
        {
            _typeName = typeName;
            _created = DateTime.Now;
        }

        ~FinalizerTracker()
        {
            var message = $"{_typeName} finalized after {DateTime.Now - _created:c} - possible missed Dispose()";
            _finalizerLog.Add(message);
            Console.WriteLine(message);
        }

        public void Suppress()
        {
            GC.SuppressFinalize(this);
        }

        public static IEnumerable<string> GetFinalizerLog()
        {
            return _finalizerLog.ToArray();
        }
    }

    // Usage in disposable classes
    public class TrackedDisposable : IDisposable
    {
        private readonly FinalizerTracker _tracker;
        private bool _disposed;

        public TrackedDisposable()
        {
            _tracker = new FinalizerTracker(GetType().Name);
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                _tracker.Suppress(); // Suppress finalizer since we're disposing properly
                _disposed = true;
            }
        }
    }
}
```

### Automatic Memory Leak Prevention

```csharp
public class AutomaticLeakPrevention
{
    // Automatic event unsubscription
    public class AutoEventSubscription : IDisposable
    {
        private readonly EventInfo _eventInfo;
        private readonly object _source;
        private readonly Delegate _handler;

        public AutoEventSubscription(object source, string eventName, Delegate handler)
        {
            _source = source;
            _handler = handler;
            _eventInfo = source.GetType().GetEvent(eventName);

            _eventInfo.AddEventHandler(source, handler);
        }

        public void Dispose()
        {
            _eventInfo?.RemoveEventHandler(_source, _handler);
        }
    }

    // Extension method for automatic subscription
    public static class EventExtensions
    {
        public static IDisposable Subscribe<T>(this EventHandler<T> eventHandler,
            object source, Action<T> handler) where T : EventArgs
        {
            EventHandler<T> wrapper = (sender, args) => handler(args);
            eventHandler += wrapper;

            return new ActionDisposable(() => eventHandler -= wrapper);
        }
    }

    public class ActionDisposable : IDisposable
    {
        private readonly Action _disposeAction;
        private bool _disposed;

        public ActionDisposable(Action disposeAction)
        {
            _disposeAction = disposeAction ?? throw new ArgumentNullException(nameof(disposeAction));
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                _disposeAction();
                _disposed = true;
            }
        }
    }

    // Scoped resource manager
    public class ScopedResourceManager : IDisposable
    {
        private readonly Stack<IDisposable> _resources = new Stack<IDisposable>();
        private bool _disposed;

        public T AddResource<T>(T resource) where T : IDisposable
        {
            ThrowIfDisposed();
            _resources.Push(resource);
            return resource;
        }

        public T AddResource<T>(Func<T> factory, Action<T> cleanup)
        {
            ThrowIfDisposed();
            var resource = factory();
            _resources.Push(new ActionDisposable(() => cleanup(resource)));
            return resource;
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                // Dispose resources in reverse order (LIFO)
                while (_resources.Count > 0)
                {
                    try
                    {
                        _resources.Pop().Dispose();
                    }
                    catch (Exception ex)
                    {
                        // Log but don't stop cleanup
                        Console.WriteLine($"Error disposing resource: {ex.Message}");
                    }
                }

                _disposed = true;
            }
        }

        private void ThrowIfDisposed()
        {
            if (_disposed)
                throw new ObjectDisposedException(nameof(ScopedResourceManager));
        }
    }

    // Usage example
    public void UseScopedResources()
    {
        using var scope = new ScopedResourceManager();

        var fileStream = scope.AddResource(new FileStream("test.txt", FileMode.Create));
        var writer = scope.AddResource(new StreamWriter(fileStream));
        var timer = scope.AddResource(
            () => new Timer(_ => { }, null, TimeSpan.Zero, TimeSpan.FromSeconds(1)),
            t => t.Dispose()
        );

        // Use resources...
        writer.WriteLine("Hello World");

        // All resources automatically disposed in correct order
    }
}
```

## Performance Profiling and Diagnostics

### Memory Profiling Tools and Techniques

```csharp
public class MemoryProfiling
{
    // Built-in memory diagnostics
    public static class MemoryDiagnostics
    {
        public static void PrintMemoryStats()
        {
            Console.WriteLine("=== Memory Statistics ===");
            Console.WriteLine($"Total Memory: {GC.GetTotalMemory(false):N0} bytes");
            Console.WriteLine($"Total Memory (after GC): {GC.GetTotalMemory(true):N0} bytes");
            Console.WriteLine($"Gen 0 Collections: {GC.CollectionCount(0)}");
            Console.WriteLine($"Gen 1 Collections: {GC.CollectionCount(1)}");
            Console.WriteLine($"Gen 2 Collections: {GC.CollectionCount(2)}");
            Console.WriteLine($"Server GC: {GCSettings.IsServerGC}");
            Console.WriteLine($"Latency Mode: {GCSettings.LatencyMode}");

            // Working set information
            var process = Process.GetCurrentProcess();
            Console.WriteLine($"Working Set: {process.WorkingSet64:N0} bytes");
            Console.WriteLine($"Private Memory: {process.PrivateMemorySize64:N0} bytes");
            Console.WriteLine($"Virtual Memory: {process.VirtualMemorySize64:N0} bytes");
        }

        public static void AnalyzeObjectAllocations<T>(Func<T> factory, int iterations = 10000)
        {
            // Warm up
            for (int i = 0; i < 100; i++)
            {
                factory();
            }

            // Force GC and measure
            GC.Collect();
            GC.WaitForPendingFinalizers();
            GC.Collect();

            long memoryBefore = GC.GetTotalMemory(false);
            int gen0Before = GC.CollectionCount(0);
            int gen1Before = GC.CollectionCount(1);
            int gen2Before = GC.CollectionCount(2);

            var objects = new T[iterations];
            for (int i = 0; i < iterations; i++)
            {
                objects[i] = factory();
            }

            long memoryAfter = GC.GetTotalMemory(false);
            int gen0After = GC.CollectionCount(0);
            int gen1After = GC.CollectionCount(1);
            int gen2After = GC.CollectionCount(2);

            Console.WriteLine($"=== Allocation Analysis for {typeof(T).Name} ===");
            Console.WriteLine($"Objects created: {iterations:N0}");
            Console.WriteLine($"Memory allocated: {memoryAfter - memoryBefore:N0} bytes");
            Console.WriteLine($"Average per object: {(double)(memoryAfter - memoryBefore) / iterations:F2} bytes");
            Console.WriteLine($"Gen 0 collections: {gen0After - gen0Before}");
            Console.WriteLine($"Gen 1 collections: {gen1After - gen1Before}");
            Console.WriteLine($"Gen 2 collections: {gen2After - gen2Before}");

            // Keep reference to prevent early collection
            GC.KeepAlive(objects);
        }
    }

    // Custom memory profiler
    public class SimpleMemoryProfiler : IDisposable
    {
        private readonly Dictionary<string, ProfileEntry> _profiles = new Dictionary<string, ProfileEntry>();
        private readonly Timer _reportTimer;

        private class ProfileEntry
        {
            public long InitialMemory { get; set; }
            public long PeakMemory { get; set; }
            public int Gen0Collections { get; set; }
            public int Gen1Collections { get; set; }
            public int Gen2Collections { get; set; }
            public DateTime StartTime { get; set; }
        }

        public SimpleMemoryProfiler(TimeSpan reportInterval)
        {
            _reportTimer = new Timer(GenerateReport, null, reportInterval, reportInterval);
        }

        public IDisposable StartProfile(string name)
        {
            var entry = new ProfileEntry
            {
                InitialMemory = GC.GetTotalMemory(false),
                PeakMemory = 0,
                Gen0Collections = GC.CollectionCount(0),
                Gen1Collections = GC.CollectionCount(1),
                Gen2Collections = GC.CollectionCount(2),
                StartTime = DateTime.UtcNow
            };

            _profiles[name] = entry;

            return new ActionDisposable(() => EndProfile(name));
        }

        private void EndProfile(string name)
        {
            if (_profiles.TryGetValue(name, out var entry))
            {
                long endMemory = GC.GetTotalMemory(false);
                entry.PeakMemory = Math.Max(entry.PeakMemory, endMemory);

                Console.WriteLine($"Profile '{name}' completed:");
                Console.WriteLine($"  Duration: {DateTime.UtcNow - entry.StartTime:c}");
                Console.WriteLine($"  Initial memory: {entry.InitialMemory:N0} bytes");
                Console.WriteLine($"  Final memory: {endMemory:N0} bytes");
                Console.WriteLine($"  Peak memory: {entry.PeakMemory:N0} bytes");
                Console.WriteLine($"  Memory delta: {endMemory - entry.InitialMemory:N0} bytes");
                Console.WriteLine($"  Gen 0 collections: {GC.CollectionCount(0) - entry.Gen0Collections}");
                Console.WriteLine($"  Gen 1 collections: {GC.CollectionCount(1) - entry.Gen1Collections}");
                Console.WriteLine($"  Gen 2 collections: {GC.CollectionCount(2) - entry.Gen2Collections}");

                _profiles.Remove(name);
            }
        }

        public void UpdatePeakMemory()
        {
            long currentMemory = GC.GetTotalMemory(false);
            foreach (var entry in _profiles.Values)
            {
                entry.PeakMemory = Math.Max(entry.PeakMemory, currentMemory);
            }
        }

        private void GenerateReport(object state)
        {
            UpdatePeakMemory();

            if (_profiles.Count > 0)
            {
                Console.WriteLine($"=== Active Profiles ({_profiles.Count}) ===");
                foreach (var kvp in _profiles)
                {
                    var entry = kvp.Value;
                    Console.WriteLine($"{kvp.Key}: {DateTime.UtcNow - entry.StartTime:c} elapsed, " +
                                    $"{entry.PeakMemory - entry.InitialMemory:N0} bytes delta");
                }
            }
        }

        public void Dispose()
        {
            _reportTimer?.Dispose();
        }
    }

    // Benchmark helper
    public static class MemoryBenchmark
    {
        public static void Compare(string name1, Action action1, string name2, Action action2, int iterations = 1000)
        {
            Console.WriteLine($"=== Comparing {name1} vs {name2} ===");

            // Warm up
            action1();
            action2();

            // Benchmark first approach
            var result1 = BenchmarkSingle(name1, action1, iterations);

            // Force GC between tests
            GC.Collect();
            GC.WaitForPendingFinalizers();
            GC.Collect();
            Thread.Sleep(100);

            // Benchmark second approach
            var result2 = BenchmarkSingle(name2, action2, iterations);

            // Compare results
            Console.WriteLine("\n=== Comparison Results ===");
            Console.WriteLine($"{name1}: {result1.ElapsedTime.TotalMilliseconds:F2}ms, {result1.MemoryAllocated:N0} bytes");
            Console.WriteLine($"{name2}: {result2.ElapsedTime.TotalMilliseconds:F2}ms, {result2.MemoryAllocated:N0} bytes");

            double timeRatio = result2.ElapsedTime.TotalMilliseconds / result1.ElapsedTime.TotalMilliseconds;
            double memoryRatio = (double)result2.MemoryAllocated / result1.MemoryAllocated;

            Console.WriteLine($"Time ratio: {timeRatio:F2}x");
            Console.WriteLine($"Memory ratio: {memoryRatio:F2}x");
        }

        private static (TimeSpan ElapsedTime, long MemoryAllocated) BenchmarkSingle(string name, Action action, int iterations)
        {
            GC.Collect();
            GC.WaitForPendingFinalizers();
            GC.Collect();

            long memoryBefore = GC.GetTotalMemory(false);
            var sw = Stopwatch.StartNew();

            for (int i = 0; i < iterations; i++)
            {
                action();
            }

            sw.Stop();
            long memoryAfter = GC.GetTotalMemory(false);

            return (sw.Elapsed, memoryAfter - memoryBefore);
        }
    }
}
```

### Event Tracing for Windows (ETW) Integration

```csharp
// ETW event source for memory tracking
[EventSource(Name = "MyApp-Memory")]
public sealed class MemoryEventSource : EventSource
{
    public static readonly MemoryEventSource Log = new MemoryEventSource();

    private MemoryEventSource() { }

    [Event(1, Level = EventLevel.Informational)]
    public void AllocationEvent(string objectType, int size, string location)
    {
        WriteEvent(1, objectType, size, location);
    }

    [Event(2, Level = EventLevel.Warning)]
    public void LargeObjectAllocation(string objectType, int size, string location)
    {
        WriteEvent(2, objectType, size, location);
    }

    [Event(3, Level = EventLevel.Informational)]
    public void GCEvent(int generation, long memoryBefore, long memoryAfter)
    {
        WriteEvent(3, generation, memoryBefore, memoryAfter);
    }

    [Event(4, Level = EventLevel.Error)]
    public void MemoryPressureHigh(long totalMemory, double pressureRatio)
    {
        WriteEvent(4, totalMemory, pressureRatio);
    }
}

public class ETWMemoryTracker
{
    private readonly Timer _gcMonitor;
    private long _lastTotalMemory;

    public ETWMemoryTracker()
    {
        _gcMonitor = new Timer(MonitorGC, null, TimeSpan.Zero, TimeSpan.FromSeconds(5));
        _lastTotalMemory = GC.GetTotalMemory(false);
    }

    private void MonitorGC(object state)
    {
        long currentMemory = GC.GetTotalMemory(false);

        if (Math.Abs(currentMemory - _lastTotalMemory) > 1024 * 1024) // 1MB change
        {
            MemoryEventSource.Log.GCEvent(-1, _lastTotalMemory, currentMemory);
        }

        // Check memory pressure
        var process = Process.GetCurrentProcess();
        double pressureRatio = (double)process.WorkingSet64 / (1024 * 1024 * 1024); // GB

        if (pressureRatio > 0.8) // 80% of 1GB
        {
            MemoryEventSource.Log.MemoryPressureHigh(currentMemory, pressureRatio);
        }

        _lastTotalMemory = currentMemory;
    }

    public static void TrackAllocation<T>(Func<T> factory, [CallerMemberName] string memberName = "")
    {
        long sizeBefore = GC.GetTotalMemory(false);
        T result = factory();
        long sizeAfter = GC.GetTotalMemory(false);

        int approximateSize = (int)(sizeAfter - sizeBefore);

        MemoryEventSource.Log.AllocationEvent(typeof(T).Name, approximateSize, memberName);

        if (approximateSize > 85000) // LOH threshold
        {
            MemoryEventSource.Log.LargeObjectAllocation(typeof(T).Name, approximateSize, memberName);
        }
    }

    public void Dispose()
    {
        _gcMonitor?.Dispose();
    }
}
```

### Memory Dump Analysis

```csharp
public class MemoryDumpAnalysis
{
    // Simple object tracking for debugging
    public static class ObjectTracker
    {
        private static readonly ConcurrentDictionary<string, long> _typeCounts
            = new ConcurrentDictionary<string, long>();
        private static readonly ConcurrentDictionary<string, long> _typeSizes
            = new ConcurrentDictionary<string, long>();

        public static void TrackObject(object obj)
        {
            if (obj == null) return;

            string typeName = obj.GetType().Name;
            _typeCounts.AddOrUpdate(typeName, 1, (key, count) => count + 1);

            // Approximate size calculation
            long size = EstimateObjectSize(obj);
            _typeSizes.AddOrUpdate(typeName, size, (key, totalSize) => totalSize + size);
        }

        private static long EstimateObjectSize(object obj)
        {
            if (obj == null) return 0;

            Type type = obj.GetType();

            // Handle arrays specially
            if (type.IsArray)
            {
                Array array = (Array)obj;
                Type elementType = type.GetElementType();
                int elementSize = Marshal.SizeOf(elementType.IsValueType ? elementType : typeof(IntPtr));
                return array.Length * elementSize + IntPtr.Size * 3; // Array overhead
            }

            // For value types, use Marshal.SizeOf
            if (type.IsValueType)
            {
                return Marshal.SizeOf(type);
            }

            // For reference types, estimate based on fields
            long size = IntPtr.Size * 2; // Object header overhead

            foreach (FieldInfo field in type.GetFields(BindingFlags.Instance | BindingFlags.NonPublic | BindingFlags.Public))
            {
                if (field.FieldType.IsValueType)
                {
                    size += Marshal.SizeOf(field.FieldType);
                }
                else
                {
                    size += IntPtr.Size; // Reference size
                }
            }

            return size;
        }

        public static void PrintReport()
        {
            Console.WriteLine("=== Object Tracking Report ===");

            var sortedByCount = _typeCounts.OrderByDescending(kvp => kvp.Value);
            Console.WriteLine("\nTop 10 by Count:");
            foreach (var kvp in sortedByCount.Take(10))
            {
                string typeName = kvp.Key;
                long count = kvp.Value;
                long totalSize = _typeSizes.TryGetValue(typeName, out long size) ? size : 0;

                Console.WriteLine($"{typeName}: {count:N0} instances, {totalSize:N0} bytes total");
            }

            var sortedBySize = _typeSizes.OrderByDescending(kvp => kvp.Value);
            Console.WriteLine("\nTop 10 by Size:");
            foreach (var kvp in sortedBySize.Take(10))
            {
                string typeName = kvp.Key;
                long totalSize = kvp.Value;
                long count = _typeCounts.TryGetValue(typeName, out long c) ? c : 0;

                Console.WriteLine($"{typeName}: {totalSize:N0} bytes total, {count:N0} instances");
            }
        }

        public static void Reset()
        {
            _typeCounts.Clear();
            _typeSizes.Clear();
        }
    }

    // Memory snapshot for comparison
    public class MemorySnapshot
    {
        public DateTime Timestamp { get; }
        public long TotalMemory { get; }
        public long WorkingSet { get; }
        public int Gen0Collections { get; }
        public int Gen1Collections { get; }
        public int Gen2Collections { get; }

        public MemorySnapshot()
        {
            Timestamp = DateTime.UtcNow;
            TotalMemory = GC.GetTotalMemory(false);

            var process = Process.GetCurrentProcess();
            WorkingSet = process.WorkingSet64;

            Gen0Collections = GC.CollectionCount(0);
            Gen1Collections = GC.CollectionCount(1);
            Gen2Collections = GC.CollectionCount(2);
        }

        public void CompareTo(MemorySnapshot other)
        {
            if (other == null) throw new ArgumentNullException(nameof(other));

            TimeSpan elapsed = Timestamp - other.Timestamp;
            long memoryDelta = TotalMemory - other.TotalMemory;
            long workingSetDelta = WorkingSet - other.WorkingSet;

            Console.WriteLine($"=== Memory Comparison ({elapsed:c} elapsed) ===");
            Console.WriteLine($"Total Memory: {TotalMemory:N0} bytes ({memoryDelta:+#,0;-#,0} delta)");
            Console.WriteLine($"Working Set: {WorkingSet:N0} bytes ({workingSetDelta:+#,0;-#,0} delta)");
            Console.WriteLine($"Gen 0 Collections: {Gen0Collections - other.Gen0Collections}");
            Console.WriteLine($"Gen 1 Collections: {Gen1Collections - other.Gen1Collections}");
            Console.WriteLine($"Gen 2 Collections: {Gen2Collections - other.Gen2Collections}");
        }
    }

    // Usage example
    public void PerformMemoryAnalysis()
    {
        var initialSnapshot = new MemorySnapshot();

        // Perform memory-intensive operations
        var largeList = new List<byte[]>();
        for (int i = 0; i < 1000; i++)
        {
            byte[] array = new byte[1024 * 100]; // 100KB each
            ObjectTracker.TrackObject(array);
            largeList.Add(array);
        }

        var afterAllocationSnapshot = new MemorySnapshot();
        afterAllocationSnapshot.CompareTo(initialSnapshot);

        // Clear references and force GC
        largeList.Clear();
        largeList = null;

        GC.Collect();
        GC.WaitForPendingFinalizers();
        GC.Collect();

        var afterGCSnapshot = new MemorySnapshot();
        afterGCSnapshot.CompareTo(afterAllocationSnapshot);

        ObjectTracker.PrintReport();
    }
}
```

## Advanced Memory Topics (Unsafe Code, Pinning, Interop)

### Unsafe Code and Pointers

```csharp
public unsafe class UnsafeMemoryOperations
{
    // Basic pointer operations
    public void BasicPointerOperations()
    {
        // Stack allocation
        int* stackArray = stackalloc int[100];

        // Initialize using pointer arithmetic
        for (int i = 0; i < 100; i++)
        {
            *(stackArray + i) = i * i;
        }

        // Alternative indexing syntax
        for (int i = 0; i < 100; i++)
        {
            stackArray[i] = i * i;
        }

        // Pointer to pointer
        int** ptrToPtr = &stackArray;
        Console.WriteLine($"Value at index 5: {(*ptrToPtr)[5]}");
    }

    // Working with fixed buffers in structs
    public struct FixedBufferStruct
    {
        public fixed byte Buffer[256]; // Fixed-size buffer
        public int Length;
    }

    public void UseFixedBuffer()
    {
        FixedBufferStruct buffer;
        buffer.Length = 256;

        fixed (byte* ptr = buffer.Buffer)
        {
            // Initialize buffer
            for (int i = 0; i < buffer.Length; i++)
            {
                ptr[i] = (byte)(i % 256);
            }

            // Calculate checksum
            uint checksum = 0;
            for (int i = 0; i < buffer.Length; i++)
            {
                checksum += ptr[i];
            }

            Console.WriteLine($"Checksum: {checksum}");
        }
    }

    // Memory copy operations
    public void FastMemoryCopy(byte[] source, byte[] destination)
    {
        if (source.Length > destination.Length)
            throw new ArgumentException("Destination too small");

        fixed (byte* src = source)
        fixed (byte* dst = destination)
        {
            // Fast memory copy using pointers
            byte* srcPtr = src;
            byte* dstPtr = dst;
            int remaining = source.Length;

            // Copy 8 bytes at a time (on 64-bit systems)
            while (remaining >= sizeof(long))
            {
                *(long*)dstPtr = *(long*)srcPtr;
                srcPtr += sizeof(long);
                dstPtr += sizeof(long);
                remaining -= sizeof(long);
            }

            // Copy remaining bytes
            while (remaining > 0)
            {
                *dstPtr = *srcPtr;
                srcPtr++;
                dstPtr++;
                remaining--;
            }
        }
    }

    // Custom memory allocator
    public class UnsafeAllocator : IDisposable
    {
        private readonly IntPtr _memory;
        private readonly int _size;
        private int _offset;
        private bool _disposed;

        public UnsafeAllocator(int size)
        {
            _size = size;
            _memory = Marshal.AllocHGlobal(size);
            _offset = 0;
        }

        public T* Allocate<T>(int count = 1) where T : unmanaged
        {
            ThrowIfDisposed();

            int sizeNeeded = sizeof(T) * count;
            if (_offset + sizeNeeded > _size)
                throw new OutOfMemoryException("Not enough space in allocator");

            T* ptr = (T*)((byte*)_memory + _offset);
            _offset += sizeNeeded;

            return ptr;
        }

        public Span<T> AllocateSpan<T>(int count) where T : unmanaged
        {
            T* ptr = Allocate<T>(count);
            return new Span<T>(ptr, count);
        }

        public void Reset()
        {
            ThrowIfDisposed();
            _offset = 0;
        }

        public int BytesUsed => _offset;
        public int BytesRemaining => _size - _offset;

        private void ThrowIfDisposed()
        {
            if (_disposed)
                throw new ObjectDisposedException(nameof(UnsafeAllocator));
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                Marshal.FreeHGlobal(_memory);
                _disposed = true;
            }
        }
    }

    // Usage of custom allocator
    public void UseCustomAllocator()
    {
        using var allocator = new UnsafeAllocator(1024 * 1024); // 1MB

        // Allocate various types
        int* integers = allocator.Allocate<int>(100);
        float* floats = allocator.Allocate<float>(50);

        // Initialize
        for (int i = 0; i < 100; i++)
        {
            integers[i] = i;
        }

        for (int i = 0; i < 50; i++)
        {
            floats[i] = i * 3.14f;
        }

        Console.WriteLine($"Allocator used: {allocator.BytesUsed} bytes");
    }
}
```

### Memory Pinning and GCHandle

```csharp
public class MemoryPinningExamples
{
    // Pinning with fixed statement
    public void PinWithFixed()
    {
        byte[] array = new byte[1024];

        fixed (byte* ptr = array)
        {
            // ptr is pinned for duration of fixed block
            // Can safely pass to unmanaged code

            // Simulate unmanaged call
            ProcessBuffer(ptr, array.Length);
        }
        // Array is automatically unpinned here
    }

    private unsafe void ProcessBuffer(byte* buffer, int length)
    {
        for (int i = 0; i < length; i++)
        {
            buffer[i] = (byte)(i % 256);
        }
    }

    // Pinning with GCHandle
    public void PinWithGCHandle()
    {
        byte[] array = new byte[1024];
        GCHandle handle = GCHandle.Alloc(array, GCHandleType.Pinned);

        try
        {
            IntPtr ptr = handle.AddrOfPinnedObject();

            // Can use ptr for extended period
            // Array won't move during GC

            unsafe
            {
                ProcessBuffer((byte*)ptr, array.Length);
            }
        }
        finally
        {
            handle.Free(); // MUST free handle
        }
    }

    // Weak reference with GCHandle
    public class WeakReferenceExample
    {
        private GCHandle _weakHandle;

        public void SetTarget(object target)
        {
            if (_weakHandle.IsAllocated)
                _weakHandle.Free();

            _weakHandle = GCHandle.Alloc(target, GCHandleType.Weak);
        }

        public object GetTarget()
        {
            if (_weakHandle.IsAllocated)
                return _weakHandle.Target;

            return null;
        }

        ~WeakReferenceExample()
        {
            if (_weakHandle.IsAllocated)
                _weakHandle.Free();
        }
    }

    // Memory-mapped file with pinning
    public class PinnedMemoryMappedFile : IDisposable
    {
        private readonly MemoryMappedFile _mmf;
        private readonly MemoryMappedViewAccessor _accessor;
        private readonly GCHandle _handle;
        private readonly SafeBuffer _buffer;
        private bool _disposed;

        public unsafe PinnedMemoryMappedFile(string name, long size)
        {
            _mmf = MemoryMappedFile.CreateNew(name, size);
            _accessor = _mmf.CreateViewAccessor(0, size);
            _buffer = _accessor.SafeMemoryMappedViewHandle;

            // Pin the buffer
            byte* ptr = null;
            _buffer.AcquirePointer(ref ptr);
            _handle = GCHandle.Alloc(ptr, GCHandleType.Pinned);
        }

        public unsafe byte* GetPointer()
        {
            ThrowIfDisposed();
            return (byte*)_handle.AddrOfPinnedObject();
        }

        public long Size => _accessor.Capacity;

        private void ThrowIfDisposed()
        {
            if (_disposed)
                throw new ObjectDisposedException(nameof(PinnedMemoryMappedFile));
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                if (_handle.IsAllocated)
                    _handle.Free();

                _buffer?.ReleasePointer();
                _accessor?.Dispose();
                _mmf?.Dispose();

                _disposed = true;
            }
        }
    }

    // High-performance buffer pool with pinning
    public class PinnedBufferPool : IDisposable
    {
        private readonly ConcurrentQueue<PinnedBuffer> _buffers = new();
        private readonly int _bufferSize;
        private readonly int _maxBuffers;
        private int _currentBuffers;
        private bool _disposed;

        public PinnedBufferPool(int bufferSize, int maxBuffers = 10)
        {
            _bufferSize = bufferSize;
            _maxBuffers = maxBuffers;
        }

        public PinnedBuffer Rent()
        {
            ThrowIfDisposed();

            if (_buffers.TryDequeue(out var buffer))
            {
                return buffer;
            }

            if (_currentBuffers < _maxBuffers)
            {
                Interlocked.Increment(ref _currentBuffers);
                return new PinnedBuffer(_bufferSize);
            }

            throw new InvalidOperationException("Buffer pool exhausted");
        }

        public void Return(PinnedBuffer buffer)
        {
            if (!_disposed && buffer != null)
            {
                buffer.Reset();
                _buffers.Enqueue(buffer);
            }
        }

        private void ThrowIfDisposed()
        {
            if (_disposed)
                throw new ObjectDisposedException(nameof(PinnedBufferPool));
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                while (_buffers.TryDequeue(out var buffer))
                {
                    buffer.Dispose();
                }

                _disposed = true;
            }
        }
    }

    public class PinnedBuffer : IDisposable
    {
        private readonly byte[] _buffer;
        private readonly GCHandle _handle;
        private bool _disposed;

        public unsafe PinnedBuffer(int size)
        {
            _buffer = new byte[size];
            _handle = GCHandle.Alloc(_buffer, GCHandleType.Pinned);
        }

        public unsafe byte* Pointer
        {
            get
            {
                ThrowIfDisposed();
                return (byte*)_handle.AddrOfPinnedObject();
            }
        }

        public int Length => _buffer.Length;

        public Span<byte> AsSpan() => _buffer.AsSpan();

        public void Reset()
        {
            ThrowIfDisposed();
            Array.Clear(_buffer, 0, _buffer.Length);
        }

        private void ThrowIfDisposed()
        {
            if (_disposed)
                throw new ObjectDisposedException(nameof(PinnedBuffer));
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                _handle.Free();
                _disposed = true;
            }
        }
    }
}
```

### Platform Invoke (P/Invoke) and Interop

```csharp
public class PInvokeMemoryOperations
{
    // Windows API memory functions
    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern IntPtr VirtualAlloc(
        IntPtr lpAddress,
        UIntPtr dwSize,
        uint flAllocationType,
        uint flProtect);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool VirtualFree(
        IntPtr lpAddress,
        UIntPtr dwSize,
        uint dwFreeType);

    [DllImport("msvcrt.dll", EntryPoint = "memcpy")]
    private static extern IntPtr memcpy(IntPtr dest, IntPtr src, UIntPtr count);

    [DllImport("msvcrt.dll", EntryPoint = "memset")]
    private static extern IntPtr memset(IntPtr dest, int value, UIntPtr count);

    // Windows memory constants
    private const uint MEM_COMMIT = 0x1000;
    private const uint MEM_RESERVE = 0x2000;
    private const uint PAGE_READWRITE = 0x04;
    private const uint MEM_RELEASE = 0x8000;

    // Custom virtual memory allocator
    public class VirtualMemoryAllocator : IDisposable
    {
        private IntPtr _memory;
        private readonly UIntPtr _size;
        private bool _disposed;

        public VirtualMemoryAllocator(long size)
        {
            _size = new UIntPtr((ulong)size);
            _memory = VirtualAlloc(IntPtr.Zero, _size, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);

            if (_memory == IntPtr.Zero)
            {
                throw new OutOfMemoryException($"Failed to allocate {size} bytes of virtual memory");
            }
        }

        public IntPtr Pointer
        {
            get
            {
                ThrowIfDisposed();
                return _memory;
            }
        }

        public long Size => (long)_size.ToUInt64();

        public unsafe Span<T> AsSpan<T>() where T : unmanaged
        {
            ThrowIfDisposed();
            int count = (int)(Size / sizeof(T));
            return new Span<T>((T*)_memory, count);
        }

        public void Zero()
        {
            ThrowIfDisposed();
            memset(_memory, 0, _size);
        }

        private void ThrowIfDisposed()
        {
            if (_disposed)
                throw new ObjectDisposedException(nameof(VirtualMemoryAllocator));
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                if (_memory != IntPtr.Zero)
                {
                    VirtualFree(_memory, UIntPtr.Zero, MEM_RELEASE);
                    _memory = IntPtr.Zero;
                }
                _disposed = true;
            }
        }
    }

    // High-performance memory operations
    public unsafe class FastMemoryOperations
    {
        public static void FastCopy(void* source, void* destination, int byteCount)
        {
            if (byteCount <= 0) return;

            byte* src = (byte*)source;
            byte* dst = (byte*)destination;

            // Use native memcpy for large copies
            if (byteCount > 1024)
            {
                memcpy(new IntPtr(dst), new IntPtr(src), new UIntPtr((uint)byteCount));
                return;
            }

            // Optimized copy for smaller buffers
            int remaining = byteCount;

            // Copy 8 bytes at a time
            while (remaining >= 8)
            {
                *(long*)dst = *(long*)src;
                src += 8;
                dst += 8;
                remaining -= 8;
            }

            // Copy 4 bytes
            if (remaining >= 4)
            {
                *(int*)dst = *(int*)src;
                src += 4;
                dst += 4;
                remaining -= 4;
            }

            // Copy remaining bytes
            while (remaining > 0)
            {
                *dst = *src;
                src++;
                dst++;
                remaining--;
            }
        }

        public static void FastZero(void* memory, int byteCount)
        {
            if (byteCount <= 0) return;

            if (byteCount > 1024)
            {
                memset(new IntPtr(memory), 0, new UIntPtr((uint)byteCount));
                return;
            }

            byte* ptr = (byte*)memory;
            int remaining = byteCount;

            // Zero 8 bytes at a time
            while (remaining >= 8)
            {
                *(long*)ptr = 0;
                ptr += 8;
                remaining -= 8;
            }

            // Zero remaining bytes
            while (remaining > 0)
            {
                *ptr = 0;
                ptr++;
                remaining--;
            }
        }

        public static bool FastEquals(void* left, void* right, int byteCount)
        {
            if (byteCount <= 0) return true;

            byte* l = (byte*)left;
            byte* r = (byte*)right;
            int remaining = byteCount;

            // Compare 8 bytes at a time
            while (remaining >= 8)
            {
                if (*(long*)l != *(long*)r)
                    return false;

                l += 8;
                r += 8;
                remaining -= 8;
            }

            // Compare remaining bytes
            while (remaining > 0)
            {
                if (*l != *r)
                    return false;

                l++;
                r++;
                remaining--;
            }

            return true;
        }
    }

    // Marshaling complex structures
    [StructLayout(LayoutKind.Sequential)]
    public struct ComplexStruct
    {
        public int IntValue;
        public double DoubleValue;
        [MarshalAs(UnmanagedType.ByValTStr, SizeConst = 256)]
        public string StringValue;
        public IntPtr PointerValue;
    }

    // Example of marshaling to/from unmanaged memory
    public void MarshalExample()
    {
        var managed = new ComplexStruct
        {
            IntValue = 42,
            DoubleValue = 3.14159,
            StringValue = "Hello, World!",
            PointerValue = new IntPtr(0x12345678)
        };

        // Calculate size needed
        int size = Marshal.SizeOf<ComplexStruct>();

        // Allocate unmanaged memory
        IntPtr ptr = Marshal.AllocHGlobal(size);

        try
        {
            // Marshal to unmanaged memory
            Marshal.StructureToPtr(managed, ptr, false);

            // Marshal back from unmanaged memory
            var unmarshaled = Marshal.PtrToStructure<ComplexStruct>(ptr);

            Console.WriteLine($"Original: {managed.IntValue}, {managed.DoubleValue}, {managed.StringValue}");
            Console.WriteLine($"Unmarshaled: {unmarshaled.IntValue}, {unmarshaled.DoubleValue}, {unmarshaled.StringValue}");
        }
        finally
        {
            Marshal.FreeHGlobal(ptr);
        }
    }

    // Custom marshaling for performance-critical scenarios
    public unsafe class CustomMarshaler
    {
        public static void MarshalArray<T>(T[] array, void* destination) where T : unmanaged
        {
            fixed (T* src = array)
            {
                FastMemoryOperations.FastCopy(src, destination, array.Length * sizeof(T));
            }
        }

        public static void UnmarshalArray<T>(void* source, T[] array) where T : unmanaged
        {
            fixed (T* dst = array)
            {
                FastMemoryOperations.FastCopy(source, dst, array.Length * sizeof(T));
            }
        }

        public static T* AllocateAndMarshal<T>(T[] array) where T : unmanaged
        {
            int sizeInBytes = array.Length * sizeof(T);
            T* ptr = (T*)Marshal.AllocHGlobal(sizeInBytes);

            MarshalArray(array, ptr);

            return ptr;
        }

        public static void FreeMarshaled<T>(T* ptr) where T : unmanaged
        {
            Marshal.FreeHGlobal(new IntPtr(ptr));
        }
    }
}
```

### Memory-Mapped Files and Shared Memory

```csharp
public class MemoryMappedFileExamples
{
    // Basic memory-mapped file
    public class BasicMemoryMappedFile : IDisposable
    {
        private readonly MemoryMappedFile _mmf;
        private readonly MemoryMappedViewAccessor _accessor;
        private bool _disposed;

        public BasicMemoryMappedFile(string name, long size)
        {
            _mmf = MemoryMappedFile.CreateNew(name, size);
            _accessor = _mmf.CreateViewAccessor(0, size);
        }

        public void WriteBytes(long offset, byte[] data)
        {
            ThrowIfDisposed();
            _accessor.WriteArray(offset, data, 0, data.Length);
        }

        public byte[] ReadBytes(long offset, int count)
        {
            ThrowIfDisposed();
            byte[] buffer = new byte[count];
            _accessor.ReadArray(offset, buffer, 0, count);
            return buffer;
        }

        public void WriteStruct<T>(long offset, T value) where T : struct
        {
            ThrowIfDisposed();
            _accessor.Write(offset, ref value);
        }

        public T ReadStruct<T>(long offset) where T : struct
        {
            ThrowIfDisposed();
            return _accessor.ReadStruct<T>(offset);
        }

        private void ThrowIfDisposed()
        {
            if (_disposed)
                throw new ObjectDisposedException(nameof(BasicMemoryMappedFile));
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                _accessor?.Dispose();
                _mmf?.Dispose();
                _disposed = true;
            }
        }
    }

    // High-performance memory-mapped buffer
    public unsafe class HighPerformanceMemoryMappedBuffer : IDisposable
    {
        private readonly MemoryMappedFile _mmf;
        private readonly MemoryMappedViewAccessor _accessor;
        private readonly byte* _basePointer;
        private readonly long _size;
        private bool _disposed;

        public HighPerformanceMemoryMappedBuffer(string name, long size)
        {
            _size = size;
            _mmf = MemoryMappedFile.CreateNew(name, size);
            _accessor = _mmf.CreateViewAccessor(0, size);

            byte* ptr = null;
            _accessor.SafeMemoryMappedViewHandle.AcquirePointer(ref ptr);
            _basePointer = ptr;
        }

        public byte* GetPointer(long offset = 0)
        {
            ThrowIfDisposed();
            if (offset < 0 || offset >= _size)
                throw new ArgumentOutOfRangeException(nameof(offset));

            return _basePointer + offset;
        }

        public Span<T> GetSpan<T>(long offset, int count) where T : unmanaged
        {
            ThrowIfDisposed();

            long sizeNeeded = count * sizeof(T);
            if (offset + sizeNeeded > _size)
                throw new ArgumentOutOfRangeException();

            return new Span<T>(GetPointer(offset), count);
        }

        public void FastCopy(long destinationOffset, void* source, int byteCount)
        {
            ThrowIfDisposed();

            if (destinationOffset + byteCount > _size)
                throw new ArgumentOutOfRangeException();

            PInvokeMemoryOperations.FastMemoryOperations.FastCopy(
                source,
                GetPointer(destinationOffset),
                byteCount);
        }

        public void Zero(long offset, int byteCount)
        {
            ThrowIfDisposed();

            if (offset + byteCount > _size)
                throw new ArgumentOutOfRangeException();

            PInvokeMemoryOperations.FastMemoryOperations.FastZero(
                GetPointer(offset),
                byteCount);
        }

        public long Size => _size;

        private void ThrowIfDisposed()
        {
            if (_disposed)
                throw new ObjectDisposedException(nameof(HighPerformanceMemoryMappedBuffer));
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                _accessor?.SafeMemoryMappedViewHandle?.ReleasePointer();
                _accessor?.Dispose();
                _mmf?.Dispose();
                _disposed = true;
            }
        }
    }

    // Inter-process communication using memory-mapped files
    public class SharedMemoryQueue<T> : IDisposable where T : unmanaged
    {
        private readonly MemoryMappedFile _mmf;
        private readonly MemoryMappedViewAccessor _accessor;
        private readonly string _mutexName;
        private readonly Mutex _mutex;
        private bool _disposed;

        private const int HeaderSize = sizeof(int) * 3; // head, tail, capacity
        private readonly int _itemSize = sizeof(T);
        private readonly int _capacity;

        public SharedMemoryQueue(string name, int capacity)
        {
            _capacity = capacity;
            _mutexName = name + "_mutex";

            long totalSize = HeaderSize + (capacity * _itemSize);
            _mmf = MemoryMappedFile.CreateOrOpen(name, totalSize);
            _accessor = _mmf.CreateViewAccessor(0, totalSize);

            bool createdNew;
            _mutex = new Mutex(false, _mutexName, out createdNew);

            if (createdNew)
            {
                // Initialize as empty queue
                _accessor.Write(0, 0); // head
                _accessor.Write(4, 0); // tail
                _accessor.Write(8, capacity); // capacity
            }
        }

        public bool TryEnqueue(T item)
        {
            ThrowIfDisposed();

            _mutex.WaitOne();
            try
            {
                int head = _accessor.ReadInt32(0);
                int tail = _accessor.ReadInt32(4);
                int capacity = _accessor.ReadInt32(8);

                int nextTail = (tail + 1) % capacity;
                if (nextTail == head) // Queue is full
                    return false;

                // Write item
                long itemOffset = HeaderSize + (tail * _itemSize);
                _accessor.Write(itemOffset, ref item);

                // Update tail
                _accessor.Write(4, nextTail);
                return true;
            }
            finally
            {
                _mutex.ReleaseMutex();
            }
        }

        public bool TryDequeue(out T item)
        {
            ThrowIfDisposed();

            item = default;

            _mutex.WaitOne();
            try
            {
                int head = _accessor.ReadInt32(0);
                int tail = _accessor.ReadInt32(4);

                if (head == tail) // Queue is empty
                    return false;

                // Read item
                long itemOffset = HeaderSize + (head * _itemSize);
                item = _accessor.ReadStruct<T>(itemOffset);

                // Update head
                int nextHead = (head + 1) % _capacity;
                _accessor.Write(0, nextHead);

                return true;
            }
            finally
            {
                _mutex.ReleaseMutex();
            }
        }

        public int Count
        {
            get
            {
                ThrowIfDisposed();

                _mutex.WaitOne();
                try
                {
                    int head = _accessor.ReadInt32(0);
                    int tail = _accessor.ReadInt32(4);

                    if (tail >= head)
                        return tail - head;
                    else
                        return (_capacity - head) + tail;
                }
                finally
                {
                    _mutex.ReleaseMutex();
                }
            }
        }

        private void ThrowIfDisposed()
        {
            if (_disposed)
                throw new ObjectDisposedException(nameof(SharedMemoryQueue<T>));
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                _mutex?.Dispose();
                _accessor?.Dispose();
                _mmf?.Dispose();
                _disposed = true;
            }
        }
    }

    // Usage example
    public void DemonstrateSharedMemory()
    {
        // Process 1: Producer
        using var producer = new SharedMemoryQueue<int>("TestQueue", 100);

        for (int i = 0; i < 50; i++)
        {
            if (producer.TryEnqueue(i))
            {
                Console.WriteLine($"Enqueued: {i}");
            }
            else
            {
                Console.WriteLine("Queue is full");
                break;
            }
        }

        // Process 2: Consumer (would be in different process)
        using var consumer = new SharedMemoryQueue<int>("TestQueue", 100);

        while (consumer.TryDequeue(out int value))
        {
            Console.WriteLine($"Dequeued: {value}");
        }
    }
}
```

## Summary and Best Practices

### Memory Management Checklist

- **Always dispose IDisposable objects** - Use `using` statements or explicit disposal
- **Understand value vs reference semantics** - Know when copying occurs
- **Minimize allocations in hot paths** - Use object pooling, `Span<T>`, and `stackalloc`
- **Avoid boxing** - Use generics instead of `object` parameters
- **Be careful with event handlers** - Always unsubscribe to prevent leaks
- **Monitor memory usage** - Use profiling tools to identify issues early
- **Prefer `readonly struct`** - For immutable value types to avoid copying
- **Use `in` parameters** - For large structs to avoid copying
- **Pool large objects** - To reduce GC pressure
- **Pin memory sparingly** - Only when necessary for interop

### Performance Guidelines

1. **Allocation Patterns**
   - Prefer stack allocation for temporary data
   - Use object pools for frequently allocated objects
   - Minimize allocations in loops

2. **Collection Management**
   - Pre-size collections when possible
   - Use appropriate collection types for access patterns
   - Clear collections rather than replacing them

3. **String Handling**
   - Use `StringBuilder` for multiple concatenations
   - Use `string.Join()` for simple joining
   - Leverage `Span<char>` for parsing without allocation

4. **Unsafe Code**
   - Only use when performance is critical
   - Always validate bounds and handle exceptions
   - Prefer safe alternatives when possible

This comprehensive guide covers the essential aspects of C# memory management, from basic concepts to advanced techniques. Understanding these principles will help you write efficient, performant C# applications with proper resource management.