# C# Concurrency & Asynchronous Programming Guide

## Table of Contents
- [Threading Fundamentals](#threading-fundamentals)
- [Task-based Asynchronous Programming (TAP)](#task-based-asynchronous-programming)
- [Parallel Programming](#parallel-programming)
- [Synchronization Primitives](#synchronization-primitives)
- [Thread-safe Collections](#thread-safe-collections)
- [Cancellation Tokens](#cancellation-tokens)
- [Producer-Consumer Patterns](#producer-consumer-patterns)
- [Actor Model & Concurrent Data Structures](#actor-model-concurrent-data-structures)
- [Performance Considerations](#performance-considerations)
- [Advanced Topics](#advanced-topics)
- [Real-world Patterns](#real-world-patterns)

## Threading Fundamentals

### Thread Class Basics

```csharp
// Basic thread creation and management
public class ThreadingBasics
{
    public static void BasicThreadDemo()
    {
        // Creating a thread
        Thread thread = new Thread(WorkerMethod);
        thread.Name = "WorkerThread";
        thread.IsBackground = true; // Dies when main thread dies
        thread.Start();

        // Wait for thread completion
        thread.Join(5000); // Timeout after 5 seconds

        if (thread.IsAlive)
        {
            thread.Abort(); // Deprecated in .NET Core+
        }
    }

    private static void WorkerMethod()
    {
        Console.WriteLine($"Thread {Thread.CurrentThread.ManagedThreadId} is working");
        Thread.Sleep(2000);
        Console.WriteLine("Work completed");
    }

    // Parameterized thread start
    public static void ParameterizedThreadDemo()
    {
        Thread thread = new Thread(ParameterizedWorker);
        thread.Start("Hello from thread!");
        thread.Join();
    }

    private static void ParameterizedWorker(object data)
    {
        string message = data as string;
        Console.WriteLine($"Thread received: {message}");
    }
}
```

### Thread Lifecycle and States

```csharp
public class ThreadLifecycle
{
    public static void DemonstrateThreadStates()
    {
        Thread thread = new Thread(() =>
        {
            Console.WriteLine("Thread started - Running state");
            Thread.Sleep(1000);
            Console.WriteLine("Thread completing");
        });

        Console.WriteLine($"Before Start: {thread.ThreadState}"); // Unstarted

        thread.Start();
        Console.WriteLine($"After Start: {thread.ThreadState}"); // Running

        Thread.Sleep(500);
        Console.WriteLine($"During execution: {thread.ThreadState}"); // Running or WaitSleepJoin

        thread.Join();
        Console.WriteLine($"After completion: {thread.ThreadState}"); // Stopped
    }

    // Thread-local storage
    private static ThreadLocal<int> threadLocalValue = new ThreadLocal<int>(() => 0);

    public static void ThreadLocalStorageDemo()
    {
        Parallel.For(0, 5, i =>
        {
            threadLocalValue.Value = i * 10;
            Thread.Sleep(100);
            Console.WriteLine($"Thread {Thread.CurrentThread.ManagedThreadId}: {threadLocalValue.Value}");
        });
    }
}
```

## Task-based Asynchronous Programming

### Async/Await Fundamentals

```csharp
public class AsyncAwaitBasics
{
    // Basic async method
    public static async Task<string> FetchDataAsync()
    {
        await Task.Delay(1000); // Simulate async operation
        return "Data fetched successfully";
    }

    // Async method without return value
    public static async Task ProcessDataAsync()
    {
        Console.WriteLine("Starting processing...");
        await Task.Delay(2000);
        Console.WriteLine("Processing completed");
    }

    // Handling multiple async operations
    public static async Task<string[]> FetchMultipleDataAsync()
    {
        // Start all tasks concurrently
        Task<string> task1 = FetchDataFromSource1Async();
        Task<string> task2 = FetchDataFromSource2Async();
        Task<string> task3 = FetchDataFromSource3Async();

        // Wait for all to complete
        string[] results = await Task.WhenAll(task1, task2, task3);
        return results;
    }

    private static async Task<string> FetchDataFromSource1Async()
    {
        await Task.Delay(1000);
        return "Data from source 1";
    }

    private static async Task<string> FetchDataFromSource2Async()
    {
        await Task.Delay(1500);
        return "Data from source 2";
    }

    private static async Task<string> FetchDataFromSource3Async()
    {
        await Task.Delay(800);
        return "Data from source 3";
    }
}
```

### ConfigureAwait and Context Handling

```csharp
public class ContextHandling
{
    // Avoiding deadlocks in UI applications
    public static async Task<string> SafeAsyncMethod()
    {
        // ConfigureAwait(false) prevents deadlocks
        string result = await FetchDataAsync().ConfigureAwait(false);

        // This won't run on the original context
        await ProcessResultAsync(result).ConfigureAwait(false);

        return result;
    }

    // ValueTask for performance-critical scenarios
    public static ValueTask<int> GetCachedValueAsync(int key)
    {
        if (_cache.TryGetValue(key, out int cachedValue))
        {
            // Return synchronously if cached
            return new ValueTask<int>(cachedValue);
        }

        // Return async task if not cached
        return new ValueTask<int>(FetchFromDatabaseAsync(key));
    }

    private static readonly Dictionary<int, int> _cache = new();

    private static async Task<int> FetchFromDatabaseAsync(int key)
    {
        await Task.Delay(100); // Simulate database call
        int value = key * 10;
        _cache[key] = value;
        return value;
    }

    private static async Task ProcessResultAsync(string result)
    {
        await Task.Delay(500);
        Console.WriteLine($"Processed: {result}");
    }

    private static async Task<string> FetchDataAsync()
    {
        await Task.Delay(1000);
        return "Async data";
    }
}
```

### Task Continuation and Chaining

```csharp
public class TaskContinuation
{
    public static async Task DemonstrateContinuations()
    {
        // Using ContinueWith
        Task<string> initialTask = Task.Run(() =>
        {
            Thread.Sleep(1000);
            return "Initial result";
        });

        Task<string> continuationTask = initialTask.ContinueWith(t =>
        {
            if (t.IsFaulted)
            {
                return "Error occurred";
            }
            return $"Processed: {t.Result}";
        }, TaskContinuationOptions.OnlyOnRanToCompletion);

        string result = await continuationTask;
        Console.WriteLine(result);
    }

    // Task chaining with different contexts
    public static async Task TaskChainingDemo()
    {
        await Task.Run(() => Console.WriteLine("Background thread"))
            .ContinueWith(_ => Console.WriteLine("Continuation 1"), TaskScheduler.Default)
            .ContinueWith(_ => Console.WriteLine("Continuation 2"), TaskScheduler.Current);
    }

    // Error handling in continuations
    public static async Task ErrorHandlingInContinuations()
    {
        try
        {
            await Task.Run(() => throw new InvalidOperationException("Simulated error"))
                .ContinueWith(t =>
                {
                    if (t.IsFaulted)
                    {
                        Console.WriteLine($"Error: {t.Exception?.GetBaseException().Message}");
                        return "Error handled";
                    }
                    return t.Result?.ToString();
                }, TaskContinuationOptions.ExecuteSynchronously);
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Outer catch: {ex.Message}");
        }
    }
}
```

## Parallel Programming

### PLINQ (Parallel LINQ)

```csharp
public class PLINQExamples
{
    public static void BasicPLINQOperations()
    {
        var numbers = Enumerable.Range(1, 10_000_000);

        // Sequential vs Parallel comparison
        var sw = Stopwatch.StartNew();
        var sequentialSum = numbers.Where(n => n % 2 == 0).Sum();
        sw.Stop();
        Console.WriteLine($"Sequential: {sw.ElapsedMilliseconds}ms, Result: {sequentialSum}");

        sw.Restart();
        var parallelSum = numbers.AsParallel().Where(n => n % 2 == 0).Sum();
        sw.Stop();
        Console.WriteLine($"Parallel: {sw.ElapsedMilliseconds}ms, Result: {parallelSum}");
    }

    public static void PLINQWithOptions()
    {
        var data = Enumerable.Range(1, 1000).ToList();

        var result = data.AsParallel()
            .WithDegreeOfParallelism(Environment.ProcessorCount)
            .WithExecutionMode(ParallelExecutionMode.ForceParallelism)
            .Where(x => IsExpensiveOperation(x))
            .Select(x => x * x)
            .OrderBy(x => x) // This forces merge
            .ToList();

        Console.WriteLine($"Processed {result.Count} items");
    }

    private static bool IsExpensiveOperation(int value)
    {
        Thread.SpinWait(1000); // Simulate CPU-intensive work
        return value % 7 == 0;
    }

    // Partitioning strategies
    public static void CustomPartitioning()
    {
        var source = Enumerable.Range(0, 1000);

        // Default partitioning
        source.AsParallel().Select(ProcessItem).ToList();

        // Custom partitioner for load balancing
        var partitioner = Partitioner.Create(source, true); // Load balancing
        partitioner.AsParallel().Select(ProcessItem).ToList();

        // Range partitioner for arrays
        var array = source.ToArray();
        var rangePartitioner = Partitioner.Create(array, true);
        rangePartitioner.AsParallel().Select(ProcessItem).ToList();
    }

    private static int ProcessItem(int item)
    {
        Thread.Sleep(10); // Simulate work
        return item * 2;
    }
}
```

### Parallel Class

```csharp
public class ParallelClassExamples
{
    public static void ParallelForDemo()
    {
        // Parallel.For
        Parallel.For(0, 1000, i =>
        {
            ProcessItem(i);
        });

        // Parallel.For with thread-local state
        object lockObject = new object();
        long total = 0;

        Parallel.For(0, 1000, () => 0L, // Thread-local initializer
            (i, loop, localTotal) =>
            {
                localTotal += ProcessItem(i);
                return localTotal;
            },
            localTotal => // Thread-local finalizer
            {
                lock (lockObject)
                {
                    total += localTotal;
                }
            });

        Console.WriteLine($"Total: {total}");
    }

    public static void ParallelForEachDemo()
    {
        var items = Enumerable.Range(1, 1000).ToList();

        // Basic parallel foreach
        Parallel.ForEach(items, item =>
        {
            ProcessItem(item);
        });

        // With partitioner
        var partitioner = Partitioner.Create(items, true);
        Parallel.ForEach(partitioner, item =>
        {
            ProcessItem(item);
        });

        // With parallel options
        var parallelOptions = new ParallelOptions
        {
            MaxDegreeOfParallelism = Environment.ProcessorCount / 2,
            CancellationToken = CancellationToken.None
        };

        Parallel.ForEach(items, parallelOptions, item =>
        {
            ProcessItem(item);
        });
    }

    public static void ParallelInvokeDemo()
    {
        // Execute multiple methods in parallel
        Parallel.Invoke(
            () => DoWork1(),
            () => DoWork2(),
            () => DoWork3()
        );

        // With parallel options
        var parallelOptions = new ParallelOptions
        {
            MaxDegreeOfParallelism = 2
        };

        Parallel.Invoke(parallelOptions,
            () => DoWork1(),
            () => DoWork2(),
            () => DoWork3()
        );
    }

    private static int ProcessItem(int item)
    {
        Thread.Sleep(1);
        return item * 2;
    }

    private static void DoWork1() => Thread.Sleep(1000);
    private static void DoWork2() => Thread.Sleep(1500);
    private static void DoWork3() => Thread.Sleep(800);
}
```

## Synchronization Primitives

### Locks and Mutexes

```csharp
public class SynchronizationPrimitives
{
    private static readonly object _lockObject = new object();
    private static readonly ReaderWriterLockSlim _readerWriterLock = new ReaderWriterLockSlim();
    private static readonly Mutex _mutex = new Mutex(false, "GlobalMutexName");
    private static int _counter = 0;

    public static void LockDemo()
    {
        // Basic lock
        Parallel.For(0, 1000, i =>
        {
            lock (_lockObject)
            {
                _counter++;
            }
        });

        Console.WriteLine($"Counter: {_counter}");
    }

    public static void ReaderWriterLockDemo()
    {
        var data = new Dictionary<string, string>();

        // Multiple readers
        Parallel.For(0, 10, i =>
        {
            _readerWriterLock.EnterReadLock();
            try
            {
                if (data.ContainsKey($"key{i}"))
                {
                    Console.WriteLine($"Read: {data[$"key{i}"]}");
                }
            }
            finally
            {
                _readerWriterLock.ExitReadLock();
            }
        });

        // Single writer
        Task.Run(() =>
        {
            _readerWriterLock.EnterWriteLock();
            try
            {
                data["newKey"] = "newValue";
                Thread.Sleep(100);
            }
            finally
            {
                _readerWriterLock.ExitWriteLock();
            }
        });
    }

    public static void MutexDemo()
    {
        bool acquired = false;
        try
        {
            acquired = _mutex.WaitOne(TimeSpan.FromSeconds(5));
            if (acquired)
            {
                Console.WriteLine("Mutex acquired");
                Thread.Sleep(2000);
            }
            else
            {
                Console.WriteLine("Failed to acquire mutex");
            }
        }
        finally
        {
            if (acquired)
            {
                _mutex.ReleaseMutex();
            }
        }
    }

    // SpinLock for short-lived locks
    private static SpinLock _spinLock = new SpinLock();

    public static void SpinLockDemo()
    {
        Parallel.For(0, 1000, i =>
        {
            bool lockTaken = false;
            try
            {
                _spinLock.Enter(ref lockTaken);
                _counter++;
            }
            finally
            {
                if (lockTaken)
                {
                    _spinLock.Exit();
                }
            }
        });
    }
}
```

### Semaphores and Events

```csharp
public class SemaphoresAndEvents
{
    private static readonly SemaphoreSlim _semaphore = new SemaphoreSlim(3, 3); // Max 3 concurrent
    private static readonly ManualResetEventSlim _manualResetEvent = new ManualResetEventSlim(false);
    private static readonly AutoResetEvent _autoResetEvent = new AutoResetEvent(false);
    private static readonly CountdownEvent _countdownEvent = new CountdownEvent(5);

    public static async Task SemaphoreDemo()
    {
        var tasks = new Task[10];

        for (int i = 0; i < 10; i++)
        {
            int taskId = i;
            tasks[i] = Task.Run(async () =>
            {
                await _semaphore.WaitAsync();
                try
                {
                    Console.WriteLine($"Task {taskId} acquired semaphore");
                    await Task.Delay(2000);
                    Console.WriteLine($"Task {taskId} releasing semaphore");
                }
                finally
                {
                    _semaphore.Release();
                }
            });
        }

        await Task.WhenAll(tasks);
    }

    public static void ManualResetEventDemo()
    {
        // Start multiple waiting tasks
        var waitingTasks = new Task[5];
        for (int i = 0; i < 5; i++)
        {
            int taskId = i;
            waitingTasks[i] = Task.Run(() =>
            {
                Console.WriteLine($"Task {taskId} waiting...");
                _manualResetEvent.Wait();
                Console.WriteLine($"Task {taskId} signaled!");
            });
        }

        Thread.Sleep(2000);
        Console.WriteLine("Setting event - all tasks will proceed");
        _manualResetEvent.Set(); // All waiting tasks proceed

        Task.WaitAll(waitingTasks);
        _manualResetEvent.Reset(); // Reset for next use
    }

    public static void AutoResetEventDemo()
    {
        var tasks = new Task[5];
        for (int i = 0; i < 5; i++)
        {
            int taskId = i;
            tasks[i] = Task.Run(() =>
            {
                Console.WriteLine($"Task {taskId} waiting...");
                _autoResetEvent.WaitOne();
                Console.WriteLine($"Task {taskId} signaled!");
            });
        }

        // Signal one task at a time
        for (int i = 0; i < 5; i++)
        {
            Thread.Sleep(1000);
            Console.WriteLine($"Signaling task {i}");
            _autoResetEvent.Set(); // Only one waiting task proceeds
        }

        Task.WaitAll(tasks);
    }

    public static void CountdownEventDemo()
    {
        // Start worker tasks
        var workers = new Task[5];
        for (int i = 0; i < 5; i++)
        {
            int workerId = i;
            workers[i] = Task.Run(() =>
            {
                Thread.Sleep((workerId + 1) * 1000);
                Console.WriteLine($"Worker {workerId} completed");
                _countdownEvent.Signal(); // Decrement count
            });
        }

        // Wait for all workers to complete
        Console.WriteLine("Waiting for all workers...");
        _countdownEvent.Wait();
        Console.WriteLine("All workers completed!");

        Task.WaitAll(workers);
    }
}
```

### Barriers and Advanced Synchronization

```csharp
public class AdvancedSynchronization
{
    private static readonly Barrier _barrier = new Barrier(3, barrier =>
    {
        Console.WriteLine($"Phase {barrier.CurrentPhaseNumber} completed by all participants");
    });

    public static void BarrierDemo()
    {
        var tasks = new Task[3];

        for (int i = 0; i < 3; i++)
        {
            int taskId = i;
            tasks[i] = Task.Run(() =>
            {
                for (int phase = 0; phase < 3; phase++)
                {
                    // Do work for this phase
                    Thread.Sleep((taskId + 1) * 500);
                    Console.WriteLine($"Task {taskId} completed phase {phase}");

                    // Wait for all tasks to complete this phase
                    _barrier.SignalAndWait();
                }
            });
        }

        Task.WaitAll(tasks);
    }

    // Custom synchronization primitive
    public class AsyncLock
    {
        private readonly SemaphoreSlim _semaphore = new SemaphoreSlim(1, 1);

        public async Task<IDisposable> LockAsync()
        {
            await _semaphore.WaitAsync();
            return new LockReleaser(_semaphore);
        }

        private class LockReleaser : IDisposable
        {
            private readonly SemaphoreSlim _semaphore;

            public LockReleaser(SemaphoreSlim semaphore)
            {
                _semaphore = semaphore;
            }

            public void Dispose()
            {
                _semaphore.Release();
            }
        }
    }

    public static async Task AsyncLockDemo()
    {
        var asyncLock = new AsyncLock();
        var tasks = new Task[5];

        for (int i = 0; i < 5; i++)
        {
            int taskId = i;
            tasks[i] = Task.Run(async () =>
            {
                using (await asyncLock.LockAsync())
                {
                    Console.WriteLine($"Task {taskId} acquired async lock");
                    await Task.Delay(1000);
                    Console.WriteLine($"Task {taskId} releasing async lock");
                }
            });
        }

        await Task.WhenAll(tasks);
    }
}
```

## Thread-safe Collections

### Concurrent Collections

```csharp
public class ConcurrentCollectionsDemo
{
    public static void ConcurrentDictionaryDemo()
    {
        var concurrentDict = new ConcurrentDictionary<string, int>();

        // Thread-safe operations
        Parallel.For(0, 1000, i =>
        {
            string key = $"key{i % 100}";

            // AddOrUpdate - atomic operation
            concurrentDict.AddOrUpdate(key, 1, (k, v) => v + 1);

            // GetOrAdd - atomic operation
            int value = concurrentDict.GetOrAdd($"computed{i}", k => ComputeValue(k));

            // TryUpdate - atomic compare and exchange
            concurrentDict.TryUpdate(key, value * 2, value);
        });

        Console.WriteLine($"Dictionary contains {concurrentDict.Count} items");

        // Safe enumeration
        foreach (var kvp in concurrentDict)
        {
            Console.WriteLine($"{kvp.Key}: {kvp.Value}");
        }
    }

    public static void ConcurrentQueueDemo()
    {
        var concurrentQueue = new ConcurrentQueue<string>();

        // Producer tasks
        var producers = new Task[3];
        for (int i = 0; i < 3; i++)
        {
            int producerId = i;
            producers[i] = Task.Run(() =>
            {
                for (int j = 0; j < 100; j++)
                {
                    concurrentQueue.Enqueue($"Producer{producerId}-Item{j}");
                    Thread.Sleep(10);
                }
            });
        }

        // Consumer tasks
        var consumers = new Task[2];
        for (int i = 0; i < 2; i++)
        {
            int consumerId = i;
            consumers[i] = Task.Run(() =>
            {
                while (!producers.All(t => t.IsCompleted) || !concurrentQueue.IsEmpty)
                {
                    if (concurrentQueue.TryDequeue(out string item))
                    {
                        Console.WriteLine($"Consumer{consumerId}: {item}");
                    }
                    Thread.Sleep(15);
                }
            });
        }

        Task.WaitAll(producers);
        Task.WaitAll(consumers);
    }

    public static void ConcurrentStackDemo()
    {
        var concurrentStack = new ConcurrentStack<int>();

        // Push items from multiple threads
        Parallel.For(0, 1000, i =>
        {
            concurrentStack.Push(i);
        });

        // Pop items safely
        var results = new List<int>();
        while (concurrentStack.TryPop(out int item))
        {
            results.Add(item);
        }

        Console.WriteLine($"Popped {results.Count} items");
    }

    public static void ConcurrentBagDemo()
    {
        var concurrentBag = new ConcurrentBag<int>();

        // Add items from multiple threads
        Parallel.For(0, 1000, i =>
        {
            concurrentBag.Add(i);
        });

        Console.WriteLine($"Bag contains {concurrentBag.Count} items");

        // Thread-safe enumeration
        Parallel.ForEach(concurrentBag, item =>
        {
            ProcessBagItem(item);
        });
    }

    public static void BlockingCollectionDemo()
    {
        using var blockingCollection = new BlockingCollection<string>(100); // Bounded capacity

        // Producer task
        var producer = Task.Run(() =>
        {
            try
            {
                for (int i = 0; i < 50; i++)
                {
                    blockingCollection.Add($"Item {i}");
                    Thread.Sleep(50);
                }
            }
            finally
            {
                blockingCollection.CompleteAdding();
            }
        });

        // Consumer task
        var consumer = Task.Run(() =>
        {
            foreach (string item in blockingCollection.GetConsumingEnumerable())
            {
                Console.WriteLine($"Consumed: {item}");
                Thread.Sleep(100);
            }
        });

        Task.WaitAll(producer, consumer);
    }

    private static int ComputeValue(string key)
    {
        Thread.Sleep(10); // Simulate computation
        return key.GetHashCode() & 0x7FFFFFFF;
    }

    private static void ProcessBagItem(int item)
    {
        Thread.Sleep(1); // Simulate processing
    }
}
```

### Thread-safe Patterns

```csharp
public class ThreadSafePatterns
{
    // Thread-safe lazy initialization
    private static readonly Lazy<ExpensiveResource> _lazyResource =
        new Lazy<ExpensiveResource>(() => new ExpensiveResource());

    public static ExpensiveResource GetResource() => _lazyResource.Value;

    // Thread-safe singleton pattern
    public sealed class ThreadSafeSingleton
    {
        private static readonly Lazy<ThreadSafeSingleton> _instance =
            new Lazy<ThreadSafeSingleton>(() => new ThreadSafeSingleton());

        public static ThreadSafeSingleton Instance => _instance.Value;

        private ThreadSafeSingleton() { }

        public void DoWork() => Console.WriteLine("Singleton working");
    }

    // Interlocked operations for atomic updates
    private static long _counter = 0;

    public static void InterlockedDemo()
    {
        Parallel.For(0, 1000, i =>
        {
            Interlocked.Increment(ref _counter);
            Interlocked.Add(ref _counter, 5);
            Interlocked.Exchange(ref _counter, _counter * 2);
            Interlocked.CompareExchange(ref _counter, 0, 1000000);
        });

        Console.WriteLine($"Final counter value: {Interlocked.Read(ref _counter)}");
    }

    // Volatile fields for memory visibility
    private static volatile bool _shouldStop = false;

    public static void VolatileDemo()
    {
        var worker = Task.Run(() =>
        {
            while (!_shouldStop) // Volatile ensures visibility
            {
                Thread.SpinWait(1000);
            }
            Console.WriteLine("Worker stopped");
        });

        Thread.Sleep(2000);
        _shouldStop = true; // This change will be visible to worker

        worker.Wait();
    }
}

public class ExpensiveResource
{
    public ExpensiveResource()
    {
        Thread.Sleep(1000); // Simulate expensive initialization
        Console.WriteLine("ExpensiveResource initialized");
    }
}
```

## Cancellation Tokens

### Basic Cancellation

```csharp
public class CancellationTokenExamples
{
    public static async Task BasicCancellationDemo()
    {
        using var cts = new CancellationTokenSource();

        // Cancel after 3 seconds
        cts.CancelAfter(TimeSpan.FromSeconds(3));

        try
        {
            await DoLongRunningWorkAsync(cts.Token);
        }
        catch (OperationCanceledException)
        {
            Console.WriteLine("Operation was cancelled");
        }
    }

    public static async Task<string> DoLongRunningWorkAsync(CancellationToken cancellationToken)
    {
        for (int i = 0; i < 10; i++)
        {
            // Check for cancellation
            cancellationToken.ThrowIfCancellationRequested();

            Console.WriteLine($"Working... step {i + 1}");
            await Task.Delay(1000, cancellationToken);
        }

        return "Work completed";
    }

    // Cooperative cancellation in CPU-bound operations
    public static long CalculateSum(int[] numbers, CancellationToken cancellationToken)
    {
        long sum = 0;

        for (int i = 0; i < numbers.Length; i++)
        {
            // Periodic cancellation check
            if (i % 1000 == 0)
            {
                cancellationToken.ThrowIfCancellationRequested();
            }

            sum += numbers[i];
        }

        return sum;
    }

    // Linked cancellation tokens
    public static async Task LinkedCancellationDemo()
    {
        using var globalCts = new CancellationTokenSource();
        using var localCts = new CancellationTokenSource();

        // Create linked token that responds to either cancellation source
        using var linkedCts = CancellationTokenSource.CreateLinkedTokenSource(
            globalCts.Token, localCts.Token);

        // Cancel global after 5 seconds
        globalCts.CancelAfter(TimeSpan.FromSeconds(5));

        try
        {
            await DoWorkWithLinkedTokenAsync(linkedCts.Token);
        }
        catch (OperationCanceledException ex)
        {
            if (globalCts.Token.IsCancellationRequested)
            {
                Console.WriteLine("Global cancellation");
            }
            else if (localCts.Token.IsCancellationRequested)
            {
                Console.WriteLine("Local cancellation");
            }
        }
    }

    private static async Task DoWorkWithLinkedTokenAsync(CancellationToken token)
    {
        await Task.Delay(10000, token); // Will be cancelled after 5 seconds
    }
}
```

### Advanced Cancellation Patterns

```csharp
public class AdvancedCancellationPatterns
{
    // Timeout with cancellation
    public static async Task<T> WithTimeout<T>(Task<T> task, TimeSpan timeout)
    {
        using var cts = new CancellationTokenSource();
        var delayTask = Task.Delay(timeout, cts.Token);

        var completedTask = await Task.WhenAny(task, delayTask);

        if (completedTask == delayTask)
        {
            throw new TimeoutException($"Operation timed out after {timeout}");
        }

        cts.Cancel(); // Cancel the delay task
        return await task;
    }

    // Cancellation with cleanup
    public static async Task DoWorkWithCleanupAsync(CancellationToken cancellationToken)
    {
        var resources = new List<IDisposable>();

        try
        {
            // Register cleanup action
            using var registration = cancellationToken.Register(() =>
            {
                Console.WriteLine("Cancellation requested - starting cleanup");
                foreach (var resource in resources)
                {
                    resource?.Dispose();
                }
            });

            // Simulate work with resource allocation
            for (int i = 0; i < 10; i++)
            {
                cancellationToken.ThrowIfCancellationRequested();

                var resource = new MemoryStream(); // Example resource
                resources.Add(resource);

                await Task.Delay(1000, cancellationToken);
            }
        }
        finally
        {
            // Ensure cleanup even without cancellation
            foreach (var resource in resources)
            {
                resource?.Dispose();
            }
        }
    }

    // Cancellation with progress reporting
    public static async Task DoWorkWithProgressAsync(
        IProgress<int> progress,
        CancellationToken cancellationToken)
    {
        const int totalWork = 100;

        for (int i = 0; i < totalWork; i++)
        {
            cancellationToken.ThrowIfCancellationRequested();

            // Simulate work
            await Task.Delay(50, cancellationToken);

            // Report progress
            progress?.Report((i + 1) * 100 / totalWork);
        }
    }

    // Graceful shutdown pattern
    public class WorkerService
    {
        private readonly CancellationTokenSource _shutdownCts = new();
        private Task _backgroundTask;

        public async Task StartAsync()
        {
            _backgroundTask = RunBackgroundWorkAsync(_shutdownCts.Token);
            await Task.CompletedTask;
        }

        public async Task StopAsync()
        {
            _shutdownCts.Cancel();

            try
            {
                await _backgroundTask;
            }
            catch (OperationCanceledException)
            {
                // Expected when cancelled
            }
        }

        private async Task RunBackgroundWorkAsync(CancellationToken cancellationToken)
        {
            while (!cancellationToken.IsCancellationRequested)
            {
                try
                {
                    await DoWorkAsync(cancellationToken);
                }
                catch (OperationCanceledException)
                {
                    Console.WriteLine("Background work cancelled gracefully");
                    throw;
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"Error in background work: {ex.Message}");
                    // Continue working despite errors
                }
            }
        }

        private async Task DoWorkAsync(CancellationToken cancellationToken)
        {
            Console.WriteLine("Doing background work...");
            await Task.Delay(5000, cancellationToken);
        }
    }
}
```

## Producer-Consumer Patterns

### Classic Producer-Consumer

```csharp
public class ProducerConsumerPatterns
{
    // Using BlockingCollection
    public static async Task BlockingCollectionPatternDemo()
    {
        using var collection = new BlockingCollection<WorkItem>(boundedCapacity: 10);

        // Producer tasks
        var producers = new Task[2];
        for (int i = 0; i < 2; i++)
        {
            int producerId = i;
            producers[i] = Task.Run(() =>
            {
                try
                {
                    for (int j = 0; j < 20; j++)
                    {
                        var workItem = new WorkItem { Id = j, ProducerId = producerId };
                        collection.Add(workItem);
                        Console.WriteLine($"Producer {producerId} added item {j}");
                        Thread.Sleep(100);
                    }
                }
                finally
                {
                    Console.WriteLine($"Producer {producerId} completed");
                }
            });
        }

        // Consumer tasks
        var consumers = new Task[3];
        for (int i = 0; i < 3; i++)
        {
            int consumerId = i;
            consumers[i] = Task.Run(() =>
            {
                foreach (var workItem in collection.GetConsumingEnumerable())
                {
                    ProcessWorkItem(workItem, consumerId);
                }
                Console.WriteLine($"Consumer {consumerId} completed");
            });
        }

        // Wait for producers to complete, then signal completion
        await Task.WhenAll(producers);
        collection.CompleteAdding();

        // Wait for consumers to finish processing
        await Task.WhenAll(consumers);
    }

    // Using Channels (.NET Core 3.0+)
    public static async Task ChannelPatternDemo()
    {
        var channel = Channel.CreateBounded<WorkItem>(new BoundedChannelOptions(10)
        {
            FullMode = BoundedChannelFullMode.Wait,
            SingleReader = false,
            SingleWriter = false
        });

        var writer = channel.Writer;
        var reader = channel.Reader;

        // Producer task
        var producer = Task.Run(async () =>
        {
            try
            {
                for (int i = 0; i < 50; i++)
                {
                    var workItem = new WorkItem { Id = i, ProducerId = 0 };
                    await writer.WriteAsync(workItem);
                    Console.WriteLine($"Produced item {i}");
                }
            }
            finally
            {
                writer.Complete();
            }
        });

        // Consumer tasks
        var consumers = new Task[3];
        for (int i = 0; i < 3; i++)
        {
            int consumerId = i;
            consumers[i] = Task.Run(async () =>
            {
                await foreach (var workItem in reader.ReadAllAsync())
                {
                    ProcessWorkItem(workItem, consumerId);
                }
            });
        }

        await Task.WhenAll(producer);
        await Task.WhenAll(consumers);
    }

    private static void ProcessWorkItem(WorkItem item, int consumerId)
    {
        Console.WriteLine($"Consumer {consumerId} processing item {item.Id} from producer {item.ProducerId}");
        Thread.Sleep(200); // Simulate processing
    }
}

public class WorkItem
{
    public int Id { get; set; }
    public int ProducerId { get; set; }
}
```

### Advanced Channel Patterns

```csharp
public class AdvancedChannelPatterns
{
    // Pipeline pattern with channels
    public static async Task PipelinePatternDemo()
    {
        // Stage 1: Generate numbers
        var stage1Channel = Channel.CreateUnbounded<int>();
        var stage1Task = Task.Run(async () =>
        {
            var writer = stage1Channel.Writer;
            try
            {
                for (int i = 0; i < 100; i++)
                {
                    await writer.WriteAsync(i);
                }
            }
            finally
            {
                writer.Complete();
            }
        });

        // Stage 2: Square numbers
        var stage2Channel = Channel.CreateUnbounded<int>();
        var stage2Task = Task.Run(async () =>
        {
            var reader = stage1Channel.Reader;
            var writer = stage2Channel.Writer;

            try
            {
                await foreach (var number in reader.ReadAllAsync())
                {
                    await writer.WriteAsync(number * number);
                }
            }
            finally
            {
                writer.Complete();
            }
        });

        // Stage 3: Filter even numbers
        var stage3Channel = Channel.CreateUnbounded<int>();
        var stage3Task = Task.Run(async () =>
        {
            var reader = stage2Channel.Reader;
            var writer = stage3Channel.Writer;

            try
            {
                await foreach (var number in reader.ReadAllAsync())
                {
                    if (number % 2 == 0)
                    {
                        await writer.WriteAsync(number);
                    }
                }
            }
            finally
            {
                writer.Complete();
            }
        });

        // Final consumer
        var consumerTask = Task.Run(async () =>
        {
            var reader = stage3Channel.Reader;
            await foreach (var result in reader.ReadAllAsync())
            {
                Console.WriteLine($"Final result: {result}");
            }
        });

        await Task.WhenAll(stage1Task, stage2Task, stage3Task, consumerTask);
    }

    // Fan-out/Fan-in pattern
    public static async Task FanOutFanInDemo()
    {
        // Input channel
        var inputChannel = Channel.CreateBounded<string>(10);

        // Multiple processing channels
        var processingChannels = new Channel<string>[3];
        for (int i = 0; i < processingChannels.Length; i++)
        {
            processingChannels[i] = Channel.CreateUnbounded<string>();
        }

        // Output channel
        var outputChannel = Channel.CreateUnbounded<string>();

        // Producer
        var producer = Task.Run(async () =>
        {
            var writer = inputChannel.Writer;
            try
            {
                for (int i = 0; i < 30; i++)
                {
                    await writer.WriteAsync($"Item {i}");
                }
            }
            finally
            {
                writer.Complete();
            }
        });

        // Fan-out dispatcher
        var dispatcher = Task.Run(async () =>
        {
            var reader = inputChannel.Reader;
            int channelIndex = 0;

            await foreach (var item in reader.ReadAllAsync())
            {
                var targetChannel = processingChannels[channelIndex];
                await targetChannel.Writer.WriteAsync(item);
                channelIndex = (channelIndex + 1) % processingChannels.Length;
            }

            // Complete all processing channels
            foreach (var channel in processingChannels)
            {
                channel.Writer.Complete();
            }
        });

        // Processing workers
        var workers = new Task[processingChannels.Length];
        for (int i = 0; i < workers.Length; i++)
        {
            int workerId = i;
            workers[i] = Task.Run(async () =>
            {
                var reader = processingChannels[workerId].Reader;
                var writer = outputChannel.Writer;

                await foreach (var item in reader.ReadAllAsync())
                {
                    // Simulate processing
                    await Task.Delay(100);
                    var processedItem = $"Processed by worker {workerId}: {item}";
                    await writer.WriteAsync(processedItem);
                }
            });
        }

        // Complete output channel when all workers are done
        var completionTask = Task.Run(async () =>
        {
            await Task.WhenAll(workers);
            outputChannel.Writer.Complete();
        });

        // Consumer
        var consumer = Task.Run(async () =>
        {
            var reader = outputChannel.Reader;
            await foreach (var result in reader.ReadAllAsync())
            {
                Console.WriteLine(result);
            }
        });

        await Task.WhenAll(producer, dispatcher, completionTask, consumer);
    }
}
```

## Actor Model & Concurrent Data Structures

### Actor Model Implementation

```csharp
public abstract class Actor
{
    private readonly Channel<object> _mailbox;
    private readonly Task _processingTask;
    private readonly CancellationTokenSource _cancellationTokenSource;

    protected Actor()
    {
        _mailbox = Channel.CreateUnbounded<object>();
        _cancellationTokenSource = new CancellationTokenSource();
        _processingTask = Task.Run(ProcessMessages);
    }

    public async Task SendAsync(object message)
    {
        await _mailbox.Writer.WriteAsync(message);
    }

    protected abstract Task HandleMessageAsync(object message);

    private async Task ProcessMessages()
    {
        var reader = _mailbox.Reader;

        try
        {
            await foreach (var message in reader.ReadAllAsync(_cancellationTokenSource.Token))
            {
                try
                {
                    await HandleMessageAsync(message);
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"Error processing message: {ex.Message}");
                }
            }
        }
        catch (OperationCanceledException)
        {
            // Expected when stopping
        }
    }

    public async Task StopAsync()
    {
        _mailbox.Writer.Complete();
        _cancellationTokenSource.Cancel();

        try
        {
            await _processingTask;
        }
        catch (OperationCanceledException)
        {
            // Expected
        }
    }
}

// Example actor implementation
public class CounterActor : Actor
{
    private int _count = 0;

    protected override async Task HandleMessageAsync(object message)
    {
        switch (message)
        {
            case IncrementMessage:
                _count++;
                Console.WriteLine($"Count incremented to: {_count}");
                break;

            case DecrementMessage:
                _count--;
                Console.WriteLine($"Count decremented to: {_count}");
                break;

            case GetCountMessage getCountMsg:
                getCountMsg.TaskCompletionSource.SetResult(_count);
                break;

            default:
                Console.WriteLine($"Unknown message type: {message.GetType()}");
                break;
        }

        await Task.CompletedTask;
    }
}

public class IncrementMessage { }
public class DecrementMessage { }

public class GetCountMessage
{
    public TaskCompletionSource<int> TaskCompletionSource { get; set; } = new();
}

// Usage example
public class ActorModelDemo
{
    public static async Task RunActorDemo()
    {
        var counterActor = new CounterActor();

        // Send increment messages from multiple tasks
        var tasks = new Task[10];
        for (int i = 0; i < 10; i++)
        {
            tasks[i] = Task.Run(async () =>
            {
                await counterActor.SendAsync(new IncrementMessage());
                await Task.Delay(100);
                await counterActor.SendAsync(new DecrementMessage());
            });
        }

        await Task.WhenAll(tasks);

        // Get final count
        var getCountMsg = new GetCountMessage();
        await counterActor.SendAsync(getCountMsg);
        int finalCount = await getCountMsg.TaskCompletionSource.Task;

        Console.WriteLine($"Final count: {finalCount}");

        await counterActor.StopAsync();
    }
}
```

### Lock-free Data Structures

```csharp
public class LockFreeDataStructures
{
    // Lock-free stack using CAS (Compare-And-Swap)
    public class LockFreeStack<T>
    {
        private class Node
        {
            public T Value;
            public Node Next;
        }

        private Node _head;

        public void Push(T item)
        {
            var newNode = new Node { Value = item };

            Node currentHead;
            do
            {
                currentHead = _head;
                newNode.Next = currentHead;
            }
            while (Interlocked.CompareExchange(ref _head, newNode, currentHead) != currentHead);
        }

        public bool TryPop(out T result)
        {
            Node currentHead;
            Node newHead;

            do
            {
                currentHead = _head;
                if (currentHead == null)
                {
                    result = default(T);
                    return false;
                }

                newHead = currentHead.Next;
            }
            while (Interlocked.CompareExchange(ref _head, newHead, currentHead) != currentHead);

            result = currentHead.Value;
            return true;
        }
    }

    // Lock-free queue (simplified version)
    public class LockFreeQueue<T>
    {
        private class Node
        {
            public volatile T Value;
            public volatile Node Next;
        }

        private volatile Node _head;
        private volatile Node _tail;

        public LockFreeQueue()
        {
            var dummy = new Node();
            _head = _tail = dummy;
        }

        public void Enqueue(T item)
        {
            var newNode = new Node { Value = item };

            while (true)
            {
                var tail = _tail;
                var next = tail.Next;

                if (tail == _tail) // Ensure tail hasn't changed
                {
                    if (next == null)
                    {
                        // Try to link new node
                        if (Interlocked.CompareExchange(ref tail.Next, newNode, null) == null)
                        {
                            break; // Successfully enqueued
                        }
                    }
                    else
                    {
                        // Try to advance tail
                        Interlocked.CompareExchange(ref _tail, next, tail);
                    }
                }
            }

            // Try to advance tail
            Interlocked.CompareExchange(ref _tail, newNode, _tail);
        }

        public bool TryDequeue(out T result)
        {
            while (true)
            {
                var head = _head;
                var tail = _tail;
                var next = head.Next;

                if (head == _head) // Ensure head hasn't changed
                {
                    if (head == tail)
                    {
                        if (next == null)
                        {
                            result = default(T);
                            return false; // Queue is empty
                        }

                        // Try to advance tail
                        Interlocked.CompareExchange(ref _tail, next, tail);
                    }
                    else
                    {
                        if (next == null)
                        {
                            continue; // Inconsistent state, retry
                        }

                        // Read value before dequeue
                        result = next.Value;

                        // Try to advance head
                        if (Interlocked.CompareExchange(ref _head, next, head) == head)
                        {
                            return true; // Successfully dequeued
                        }
                    }
                }
            }
        }
    }

    public static void LockFreeStackDemo()
    {
        var stack = new LockFreeStack<int>();

        // Push items from multiple threads
        Parallel.For(0, 1000, i =>
        {
            stack.Push(i);
        });

        // Pop items from multiple threads
        var results = new ConcurrentBag<int>();
        Parallel.For(0, 1000, i =>
        {
            if (stack.TryPop(out int value))
            {
                results.Add(value);
            }
        });

        Console.WriteLine($"Popped {results.Count} items from lock-free stack");
    }

    public static void LockFreeQueueDemo()
    {
        var queue = new LockFreeQueue<string>();

        // Enqueue from multiple producers
        var producers = Parallel.For(0, 5, producerId =>
        {
            for (int i = 0; i < 100; i++)
            {
                queue.Enqueue($"Producer{producerId}-Item{i}");
            }
        });

        // Dequeue from multiple consumers
        var results = new ConcurrentBag<string>();
        Parallel.For(0, 3, consumerId =>
        {
            for (int i = 0; i < 100; i++)
            {
                if (queue.TryDequeue(out string item))
                {
                    results.Add(item);
                }
                else
                {
                    Thread.SpinWait(100); // Brief spin before retry
                    i--; // Retry this iteration
                }
            }
        });

        Console.WriteLine($"Dequeued {results.Count} items from lock-free queue");
    }
}
```

## Performance Considerations

### Thread Pool Management

```csharp
public class ThreadPoolManagement
{
    public static void ThreadPoolConfiguration()
    {
        // Get current thread pool settings
        ThreadPool.GetMaxThreads(out int maxWorkerThreads, out int maxCompletionPortThreads);
        ThreadPool.GetMinThreads(out int minWorkerThreads, out int minCompletionPortThreads);

        Console.WriteLine($"Max Worker Threads: {maxWorkerThreads}");
        Console.WriteLine($"Max Completion Port Threads: {maxCompletionPortThreads}");
        Console.WriteLine($"Min Worker Threads: {minWorkerThreads}");
        Console.WriteLine($"Min Completion Port Threads: {minCompletionPortThreads}");

        // Configure thread pool for high-throughput scenarios
        int processorCount = Environment.ProcessorCount;
        ThreadPool.SetMinThreads(processorCount * 2, minCompletionPortThreads);
        ThreadPool.SetMaxThreads(processorCount * 100, maxCompletionPortThreads);

        // Monitor thread pool performance
        MonitorThreadPool();
    }

    private static void MonitorThreadPool()
    {
        Timer timer = new Timer(state =>
        {
            ThreadPool.GetAvailableThreads(out int availableWorkerThreads, out int availableCompletionPortThreads);
            ThreadPool.GetMaxThreads(out int maxWorkerThreads, out int maxCompletionPortThreads);

            int busyWorkerThreads = maxWorkerThreads - availableWorkerThreads;
            int busyCompletionPortThreads = maxCompletionPortThreads - availableCompletionPortThreads;

            Console.WriteLine($"Busy Worker Threads: {busyWorkerThreads}, Busy I/O Threads: {busyCompletionPortThreads}");
        }, null, TimeSpan.Zero, TimeSpan.FromSeconds(5));

        // Keep timer alive
        Console.ReadKey();
        timer.Dispose();
    }

    // Custom thread pool for specific workloads
    public class CustomTaskScheduler : TaskScheduler
    {
        private readonly BlockingCollection<Task> _tasks = new BlockingCollection<Task>();
        private readonly Thread[] _threads;

        public CustomTaskScheduler(int threadCount)
        {
            _threads = new Thread[threadCount];

            for (int i = 0; i < threadCount; i++)
            {
                _threads[i] = new Thread(() =>
                {
                    foreach (var task in _tasks.GetConsumingEnumerable())
                    {
                        TryExecuteTask(task);
                    }
                })
                {
                    IsBackground = true,
                    Name = $"CustomScheduler-{i}"
                };

                _threads[i].Start();
            }
        }

        protected override IEnumerable<Task> GetScheduledTasks() => _tasks;

        protected override void QueueTask(Task task)
        {
            _tasks.Add(task);
        }

        protected override bool TryExecuteTaskInline(Task task, bool taskWasPreviouslyQueued)
        {
            return TryExecuteTask(task);
        }

        public void Shutdown()
        {
            _tasks.CompleteAdding();
            foreach (var thread in _threads)
            {
                thread.Join();
            }
            _tasks.Dispose();
        }
    }
}
```

### Performance Optimization Patterns

```csharp
public class PerformanceOptimization
{
    // Memory-efficient async enumerable
    public static async IAsyncEnumerable<T> ProcessLargeDatasetAsync<T>(
        IEnumerable<T> source,
        Func<T, Task<T>> processor,
        [EnumeratorCancellation] CancellationToken cancellationToken = default)
    {
        var semaphore = new SemaphoreSlim(Environment.ProcessorCount);

        foreach (var item in source)
        {
            cancellationToken.ThrowIfCancellationRequested();

            await semaphore.WaitAsync(cancellationToken);

            try
            {
                var result = await processor(item).ConfigureAwait(false);
                yield return result;
            }
            finally
            {
                semaphore.Release();
            }
        }
    }

    // Batching pattern for high-throughput scenarios
    public static async Task ProcessInBatchesAsync<T>(
        IEnumerable<T> items,
        Func<IEnumerable<T>, Task> batchProcessor,
        int batchSize = 100,
        CancellationToken cancellationToken = default)
    {
        var batch = new List<T>(batchSize);

        foreach (var item in items)
        {
            cancellationToken.ThrowIfCancellationRequested();

            batch.Add(item);

            if (batch.Count >= batchSize)
            {
                await batchProcessor(batch).ConfigureAwait(false);
                batch.Clear();
            }
        }

        // Process remaining items
        if (batch.Count > 0)
        {
            await batchProcessor(batch).ConfigureAwait(false);
        }
    }

    // Object pooling for reducing allocations
    public class ObjectPool<T> where T : class, new()
    {
        private readonly ConcurrentQueue<T> _queue = new ConcurrentQueue<T>();
        private readonly Func<T> _factory;
        private readonly Action<T> _resetAction;

        public ObjectPool(Func<T> factory = null, Action<T> resetAction = null)
        {
            _factory = factory ?? (() => new T());
            _resetAction = resetAction;
        }

        public T Rent()
        {
            if (_queue.TryDequeue(out T item))
            {
                return item;
            }

            return _factory();
        }

        public void Return(T item)
        {
            if (item != null)
            {
                _resetAction?.Invoke(item);
                _queue.Enqueue(item);
            }
        }
    }

    // Memory-mapped files for large data processing
    public static async Task ProcessLargeFileAsync(string filePath)
    {
        using var mmf = MemoryMappedFile.CreateFromFile(filePath, FileMode.Open, "data", 0);
        using var accessor = mmf.CreateViewAccessor(0, 0);

        var fileInfo = new FileInfo(filePath);
        long fileSize = fileInfo.Length;

        const int chunkSize = 1024 * 1024; // 1MB chunks
        var tasks = new List<Task>();

        for (long offset = 0; offset < fileSize; offset += chunkSize)
        {
            long currentOffset = offset;
            long currentSize = Math.Min(chunkSize, fileSize - offset);

            var task = Task.Run(() =>
            {
                ProcessChunk(accessor, currentOffset, currentSize);
            });

            tasks.Add(task);
        }

        await Task.WhenAll(tasks);
    }

    private static void ProcessChunk(MemoryMappedViewAccessor accessor, long offset, long size)
    {
        // Process the chunk of data
        for (long i = offset; i < offset + size; i++)
        {
            byte data = accessor.ReadByte(i);
            // Process byte
        }
    }
}
```

## Advanced Topics

### Custom Schedulers and Contexts

```csharp
public class AdvancedScheduling
{
    // Single-threaded task scheduler
    public class SingleThreadedTaskScheduler : TaskScheduler, IDisposable
    {
        private readonly BlockingCollection<Task> _tasks = new BlockingCollection<Task>();
        private readonly Thread _thread;

        public SingleThreadedTaskScheduler()
        {
            _thread = new Thread(() =>
            {
                foreach (var task in _tasks.GetConsumingEnumerable())
                {
                    TryExecuteTask(task);
                }
            })
            {
                IsBackground = true,
                Name = "SingleThreadedScheduler"
            };

            _thread.Start();
        }

        protected override IEnumerable<Task> GetScheduledTasks() => _tasks;

        protected override void QueueTask(Task task)
        {
            _tasks.Add(task);
        }

        protected override bool TryExecuteTaskInline(Task task, bool taskWasPreviouslyQueued)
        {
            return Thread.CurrentThread == _thread && TryExecuteTask(task);
        }

        public void Dispose()
        {
            _tasks.CompleteAdding();
            _thread.Join();
            _tasks.Dispose();
        }
    }

    // Priority-based task scheduler
    public class PriorityTaskScheduler : TaskScheduler
    {
        private readonly object _lock = new object();
        private readonly List<Task> _tasks = new List<Task>();
        private readonly Dictionary<Task, int> _priorities = new Dictionary<Task, int>();

        public void QueueTask(Task task, int priority)
        {
            _priorities[task] = priority;
            QueueTask(task);
        }

        protected override void QueueTask(Task task)
        {
            lock (_lock)
            {
                _tasks.Add(task);
                _tasks.Sort((t1, t2) =>
                {
                    int p1 = _priorities.GetValueOrDefault(t1, 0);
                    int p2 = _priorities.GetValueOrDefault(t2, 0);
                    return p2.CompareTo(p1); // Higher priority first
                });
            }

            ThreadPool.QueueUserWorkItem(ProcessTasks);
        }

        private void ProcessTasks(object state)
        {
            Task taskToExecute = null;

            lock (_lock)
            {
                if (_tasks.Count > 0)
                {
                    taskToExecute = _tasks[0];
                    _tasks.RemoveAt(0);
                    _priorities.Remove(taskToExecute);
                }
            }

            if (taskToExecute != null)
            {
                TryExecuteTask(taskToExecute);
            }
        }

        protected override bool TryExecuteTaskInline(Task task, bool taskWasPreviouslyQueued)
        {
            return TryExecuteTask(task);
        }

        protected override IEnumerable<Task> GetScheduledTasks()
        {
            lock (_lock)
            {
                return _tasks.ToArray();
            }
        }
    }

    // Custom synchronization context
    public class CustomSynchronizationContext : SynchronizationContext
    {
        private readonly BlockingCollection<(SendOrPostCallback callback, object state)> _callbacks
            = new BlockingCollection<(SendOrPostCallback, object)>();
        private readonly Thread _thread;

        public CustomSynchronizationContext()
        {
            _thread = new Thread(() =>
            {
                SetSynchronizationContext(this);

                foreach (var (callback, state) in _callbacks.GetConsumingEnumerable())
                {
                    try
                    {
                        callback(state);
                    }
                    catch (Exception ex)
                    {
                        Console.WriteLine($"Error in sync context: {ex.Message}");
                    }
                }
            })
            {
                IsBackground = true,
                Name = "CustomSyncContext"
            };

            _thread.Start();
        }

        public override void Post(SendOrPostCallback d, object state)
        {
            _callbacks.Add((d, state));
        }

        public override void Send(SendOrPostCallback d, object state)
        {
            if (Thread.CurrentThread == _thread)
            {
                d(state);
            }
            else
            {
                var mre = new ManualResetEventSlim(false);
                Exception exception = null;

                Post(s =>
                {
                    try
                    {
                        d(state);
                    }
                    catch (Exception ex)
                    {
                        exception = ex;
                    }
                    finally
                    {
                        mre.Set();
                    }
                }, null);

                mre.Wait();

                if (exception != null)
                {
                    throw exception;
                }
            }
        }

        public void Shutdown()
        {
            _callbacks.CompleteAdding();
            _thread.Join();
        }
    }
}
```

## Real-world Patterns

### Web Service Patterns

```csharp
public class WebServicePatterns
{
    // Circuit breaker pattern for fault tolerance
    public class CircuitBreaker
    {
        private readonly int _failureThreshold;
        private readonly TimeSpan _timeout;
        private readonly object _lock = new object();

        private int _failureCount;
        private DateTime _lastFailureTime;
        private CircuitState _state = CircuitState.Closed;

        public CircuitBreaker(int failureThreshold, TimeSpan timeout)
        {
            _failureThreshold = failureThreshold;
            _timeout = timeout;
        }

        public async Task<T> ExecuteAsync<T>(Func<Task<T>> operation)
        {
            if (_state == CircuitState.Open)
            {
                lock (_lock)
                {
                    if (DateTime.UtcNow - _lastFailureTime >= _timeout)
                    {
                        _state = CircuitState.HalfOpen;
                    }
                    else
                    {
                        throw new CircuitBreakerOpenException();
                    }
                }
            }

            try
            {
                var result = await operation();

                if (_state == CircuitState.HalfOpen)
                {
                    Reset();
                }

                return result;
            }
            catch (Exception)
            {
                RecordFailure();
                throw;
            }
        }

        private void RecordFailure()
        {
            lock (_lock)
            {
                _failureCount++;
                _lastFailureTime = DateTime.UtcNow;

                if (_failureCount >= _failureThreshold)
                {
                    _state = CircuitState.Open;
                }
            }
        }

        private void Reset()
        {
            lock (_lock)
            {
                _failureCount = 0;
                _state = CircuitState.Closed;
            }
        }
    }

    public enum CircuitState
    {
        Closed,
        Open,
        HalfOpen
    }

    public class CircuitBreakerOpenException : Exception
    {
        public CircuitBreakerOpenException() : base("Circuit breaker is open") { }
    }

    // Retry pattern with exponential backoff
    public static async Task<T> RetryWithExponentialBackoffAsync<T>(
        Func<Task<T>> operation,
        int maxRetries = 3,
        TimeSpan initialDelay = default,
        CancellationToken cancellationToken = default)
    {
        if (initialDelay == default)
        {
            initialDelay = TimeSpan.FromMilliseconds(100);
        }

        Exception lastException = null;

        for (int attempt = 0; attempt <= maxRetries; attempt++)
        {
            try
            {
                return await operation();
            }
            catch (Exception ex) when (attempt < maxRetries)
            {
                lastException = ex;

                var delay = TimeSpan.FromMilliseconds(
                    initialDelay.TotalMilliseconds * Math.Pow(2, attempt));

                Console.WriteLine($"Attempt {attempt + 1} failed, retrying in {delay.TotalMilliseconds}ms");

                await Task.Delay(delay, cancellationToken);
            }
        }

        throw lastException ?? new InvalidOperationException("Max retries exceeded");
    }

    // Bulk processing with parallelism control
    public static async Task<IEnumerable<TResult>> ProcessInParallelAsync<TSource, TResult>(
        IEnumerable<TSource> source,
        Func<TSource, Task<TResult>> processor,
        int maxConcurrency = 10,
        CancellationToken cancellationToken = default)
    {
        using var semaphore = new SemaphoreSlim(maxConcurrency, maxConcurrency);

        var tasks = source.Select(async item =>
        {
            await semaphore.WaitAsync(cancellationToken);
            try
            {
                return await processor(item);
            }
            finally
            {
                semaphore.Release();
            }
        });

        return await Task.WhenAll(tasks);
    }
}
```

### Background Service Patterns

```csharp
public class BackgroundServicePatterns
{
    // Long-running background service
    public abstract class BackgroundServiceBase : IDisposable
    {
        private readonly CancellationTokenSource _cancellationTokenSource = new();
        private Task _backgroundTask;

        public virtual async Task StartAsync()
        {
            _backgroundTask = ExecuteAsync(_cancellationTokenSource.Token);
            await Task.CompletedTask;
        }

        public virtual async Task StopAsync()
        {
            if (_backgroundTask == null)
                return;

            _cancellationTokenSource.Cancel();

            try
            {
                await _backgroundTask;
            }
            catch (OperationCanceledException)
            {
                // Expected when cancelled
            }
        }

        protected abstract Task ExecuteAsync(CancellationToken cancellationToken);

        public virtual void Dispose()
        {
            StopAsync().GetAwaiter().GetResult();
            _cancellationTokenSource?.Dispose();
        }
    }

    // Scheduled task service
    public class ScheduledTaskService : BackgroundServiceBase
    {
        private readonly Dictionary<string, ScheduledTask> _scheduledTasks = new();
        private readonly object _lock = new object();

        public void ScheduleTask(string taskName, Func<Task> taskFunc, TimeSpan interval)
        {
            lock (_lock)
            {
                _scheduledTasks[taskName] = new ScheduledTask
                {
                    Name = taskName,
                    TaskFunc = taskFunc,
                    Interval = interval,
                    NextRun = DateTime.UtcNow + interval
                };
            }
        }

        public void UnscheduleTask(string taskName)
        {
            lock (_lock)
            {
                _scheduledTasks.Remove(taskName);
            }
        }

        protected override async Task ExecuteAsync(CancellationToken cancellationToken)
        {
            while (!cancellationToken.IsCancellationRequested)
            {
                var tasksToRun = new List<ScheduledTask>();
                var now = DateTime.UtcNow;

                lock (_lock)
                {
                    foreach (var task in _scheduledTasks.Values)
                    {
                        if (now >= task.NextRun)
                        {
                            tasksToRun.Add(task);
                            task.NextRun = now + task.Interval;
                        }
                    }
                }

                // Run scheduled tasks in parallel
                var runningTasks = tasksToRun.Select(async task =>
                {
                    try
                    {
                        await task.TaskFunc();
                        Console.WriteLine($"Scheduled task '{task.Name}' completed");
                    }
                    catch (Exception ex)
                    {
                        Console.WriteLine($"Scheduled task '{task.Name}' failed: {ex.Message}");
                    }
                });

                await Task.WhenAll(runningTasks);

                // Wait before checking again
                await Task.Delay(TimeSpan.FromSeconds(1), cancellationToken);
            }
        }
    }

    private class ScheduledTask
    {
        public string Name { get; set; }
        public Func<Task> TaskFunc { get; set; }
        public TimeSpan Interval { get; set; }
        public DateTime NextRun { get; set; }
    }

    // Message processing service
    public class MessageProcessorService : BackgroundServiceBase
    {
        private readonly Channel<ProcessingMessage> _messageChannel;
        private readonly int _maxConcurrency;

        public MessageProcessorService(int maxConcurrency = 10)
        {
            _maxConcurrency = maxConcurrency;
            _messageChannel = Channel.CreateUnbounded<ProcessingMessage>();
        }

        public async Task EnqueueMessageAsync(string messageId, string content)
        {
            var message = new ProcessingMessage
            {
                Id = messageId,
                Content = content,
                EnqueuedAt = DateTime.UtcNow
            };

            await _messageChannel.Writer.WriteAsync(message);
        }

        protected override async Task ExecuteAsync(CancellationToken cancellationToken)
        {
            using var semaphore = new SemaphoreSlim(_maxConcurrency, _maxConcurrency);
            var reader = _messageChannel.Reader;

            await foreach (var message in reader.ReadAllAsync(cancellationToken))
            {
                await semaphore.WaitAsync(cancellationToken);

                // Process message in background
                _ = Task.Run(async () =>
                {
                    try
                    {
                        await ProcessMessageAsync(message);
                    }
                    finally
                    {
                        semaphore.Release();
                    }
                }, cancellationToken);
            }
        }

        private async Task ProcessMessageAsync(ProcessingMessage message)
        {
            Console.WriteLine($"Processing message {message.Id}");

            // Simulate processing
            await Task.Delay(TimeSpan.FromSeconds(2));

            Console.WriteLine($"Message {message.Id} processed successfully");
        }

        public override async Task StopAsync()
        {
            _messageChannel.Writer.Complete();
            await base.StopAsync();
        }
    }

    private class ProcessingMessage
    {
        public string Id { get; set; }
        public string Content { get; set; }
        public DateTime EnqueuedAt { get; set; }
    }
}
```

## Best Practices and Common Pitfalls

### Best Practices

1. **Prefer async/await over raw Tasks**
   - Use ConfigureAwait(false) in library code
   - Avoid async void except for event handlers
   - Use ValueTask for frequently called methods that may complete synchronously

2. **Thread Safety Guidelines**
   - Prefer immutable data structures
   - Use concurrent collections over locks when possible
   - Minimize shared mutable state

3. **Performance Considerations**
   - Use thread pool for short-lived operations
   - Create dedicated threads for long-running tasks
   - Consider object pooling for high-allocation scenarios

4. **Error Handling**
   - Always handle AggregateException in Task.Wait scenarios
   - Use try-catch around await calls
   - Implement proper cancellation support

### Common Pitfalls to Avoid

```csharp
public class CommonPitfalls
{
    // PITFALL 1: Deadlock with synchronous blocking on async code
    public static void DeadlockExample()
    {
        // DON'T DO THIS - can cause deadlock
        // string result = SomeAsyncMethod().Result;

        // DO THIS INSTEAD
        string result = SomeAsyncMethod().GetAwaiter().GetResult();
        // OR better yet, make the calling method async
    }

    // PITFALL 2: Fire-and-forget without proper error handling
    public static async Task FireAndForgetPitfall()
    {
        // DON'T DO THIS - exceptions will be lost
        // _ = SomeAsyncMethod();

        // DO THIS INSTEAD
        _ = Task.Run(async () =>
        {
            try
            {
                await SomeAsyncMethod();
            }
            catch (Exception ex)
            {
                // Log the exception
                Console.WriteLine($"Fire-and-forget task failed: {ex.Message}");
            }
        });
    }

    // PITFALL 3: Improper exception handling in parallel operations
    public static async Task ProperExceptionHandling()
    {
        var tasks = new[]
        {
            SomeAsyncMethod(),
            SomeAsyncMethod(),
            SomeAsyncMethod()
        };

        try
        {
            await Task.WhenAll(tasks);
        }
        catch (Exception)
        {
            // This only catches the first exception
            // To get all exceptions:
            foreach (var task in tasks)
            {
                if (task.IsFaulted)
                {
                    foreach (var ex in task.Exception.InnerExceptions)
                    {
                        Console.WriteLine($"Task failed: {ex.Message}");
                    }
                }
            }
        }
    }

    private static async Task<string> SomeAsyncMethod()
    {
        await Task.Delay(1000);
        return "Done";
    }
}
```

This comprehensive guide covers the essential aspects of C# concurrency and asynchronous programming, from basic threading concepts to advanced patterns used in enterprise applications. The examples demonstrate practical implementations that can be adapted to real-world scenarios.