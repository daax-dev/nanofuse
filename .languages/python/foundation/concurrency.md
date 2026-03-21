# Python Concurrency, Threading, and Parallelism Deep Dive

## Understanding Python Concurrency

Python offers multiple approaches to concurrency and parallelism, each with distinct use cases, advantages, and limitations. Understanding when and how to use each approach is crucial for building efficient applications.

### Concurrency vs Parallelism

```
Concurrency: Multiple tasks making progress (not necessarily simultaneously)
┌─────────────────────────────────────────┐
│ Task A: ████░░░░████░░░░████░░░░         │
│ Task B: ░░░░████░░░░████░░░░████         │
│ Task C: ░░░░░░░░░░░░████░░░░████████     │
└─────────────────────────────────────────┘
           Single Core with Context Switching

Parallelism: Multiple tasks running simultaneously
┌─────────────────────────────────────────┐
│ Core 1: ████████████████████████████     │
│ Core 2: ████████████████████████████     │
│ Core 3: ████████████████████████████     │
└─────────────────────────────────────────┘
           Multi-Core True Parallelism
```

## The Global Interpreter Lock (GIL)

### Understanding the GIL

```python
import threading
import time

# The GIL prevents true parallelism in CPU-bound tasks
def cpu_bound_task(n):
    """CPU-intensive task that will be limited by GIL"""
    total = 0
    for i in range(n):
        total += i ** 2
    return total

def demonstrate_gil_limitation():
    start_time = time.time()

    # Single-threaded execution
    result1 = cpu_bound_task(1000000)
    result2 = cpu_bound_task(1000000)
    single_thread_time = time.time() - start_time

    # Multi-threaded execution (won't be faster due to GIL)
    start_time = time.time()

    threads = []
    results = []

    def worker(n, results, index):
        results[index] = cpu_bound_task(n)

    results = [None] * 2
    for i in range(2):
        thread = threading.Thread(target=worker, args=(1000000, results, i))
        threads.append(thread)
        thread.start()

    for thread in threads:
        thread.join()

    multi_thread_time = time.time() - start_time

    print(f"Single-threaded: {single_thread_time:.2f} seconds")
    print(f"Multi-threaded: {multi_thread_time:.2f} seconds")
    print(f"Speedup: {single_thread_time / multi_thread_time:.2f}x")

demonstrate_gil_limitation()
```

### When Threading Works Despite GIL

```python
import threading
import time
import requests

# I/O-bound tasks benefit from threading despite GIL
def io_bound_task(url):
    """I/O-intensive task that releases GIL during I/O operations"""
    response = requests.get(url)
    return len(response.content)

def demonstrate_threading_benefit():
    urls = [
        'https://httpbin.org/delay/1',
        'https://httpbin.org/delay/1',
        'https://httpbin.org/delay/1',
        'https://httpbin.org/delay/1'
    ]

    # Sequential execution
    start_time = time.time()
    sequential_results = [io_bound_task(url) for url in urls]
    sequential_time = time.time() - start_time

    # Threaded execution
    start_time = time.time()

    threads = []
    results = [None] * len(urls)

    def worker(url, results, index):
        results[index] = io_bound_task(url)

    for i, url in enumerate(urls):
        thread = threading.Thread(target=worker, args=(url, results, i))
        threads.append(thread)
        thread.start()

    for thread in threads:
        thread.join()

    threaded_time = time.time() - start_time

    print(f"Sequential: {sequential_time:.2f} seconds")
    print(f"Threaded: {threaded_time:.2f} seconds")
    print(f"Speedup: {sequential_time / threaded_time:.2f}x")
```

## Threading

### Basic Threading

```python
import threading
import time
import queue

# Basic thread creation and management
def worker_function(name, duration):
    """Basic worker function"""
    print(f"Worker {name} started")
    time.sleep(duration)
    print(f"Worker {name} finished")

# Create and start threads
threads = []
for i in range(3):
    thread = threading.Thread(
        target=worker_function,
        args=(f"Thread-{i}", 2),
        name=f"Worker-{i}"
    )
    threads.append(thread)
    thread.start()

# Wait for all threads to complete
for thread in threads:
    thread.join()

print("All threads completed")
```

### Thread Communication and Synchronization

```python
import threading
import time
import queue
from concurrent.collections import defaultdict

# Shared data and locks
class ThreadSafeCounter:
    def __init__(self):
        self._value = 0
        self._lock = threading.Lock()

    def increment(self):
        with self._lock:
            self._value += 1

    def value(self):
        with self._lock:
            return self._value

# Producer-Consumer pattern with Queue
def producer_consumer_example():
    # Thread-safe queue
    work_queue = queue.Queue(maxsize=10)
    results_queue = queue.Queue()

    def producer(queue, num_items):
        """Produces work items"""
        for i in range(num_items):
            item = f"item-{i}"
            queue.put(item)
            print(f"Produced: {item}")
            time.sleep(0.1)

        # Signal end of production
        queue.put(None)

    def consumer(work_queue, results_queue, consumer_id):
        """Consumes work items"""
        while True:
            item = work_queue.get()
            if item is None:
                # End signal received
                work_queue.put(None)  # Pass signal to other consumers
                break

            # Process item
            result = f"processed-{item}-by-{consumer_id}"
            results_queue.put(result)
            print(f"Consumer {consumer_id} processed: {item}")
            time.sleep(0.2)

            work_queue.task_done()

    # Start producer
    producer_thread = threading.Thread(
        target=producer,
        args=(work_queue, 20)
    )
    producer_thread.start()

    # Start consumers
    consumer_threads = []
    for i in range(3):
        thread = threading.Thread(
            target=consumer,
            args=(work_queue, results_queue, i)
        )
        consumer_threads.append(thread)
        thread.start()

    # Wait for completion
    producer_thread.join()
    for thread in consumer_threads:
        thread.join()

    # Collect results
    results = []
    while not results_queue.empty():
        results.append(results_queue.get())

    print(f"Processing completed. Results: {len(results)}")

producer_consumer_example()
```

### Advanced Threading Patterns

```python
import threading
import time
import weakref
from contextlib import contextmanager

# Thread-local storage
thread_local_data = threading.local()

def set_thread_data(name, value):
    """Set thread-local data"""
    thread_local_data.name = name
    thread_local_data.value = value

def get_thread_data():
    """Get thread-local data"""
    return getattr(thread_local_data, 'name', None), getattr(thread_local_data, 'value', None)

def thread_local_example():
    def worker(thread_id):
        set_thread_data(f"Thread-{thread_id}", thread_id * 10)
        time.sleep(1)  # Simulate work
        name, value = get_thread_data()
        print(f"Thread {thread_id}: name={name}, value={value}")

    threads = []
    for i in range(3):
        thread = threading.Thread(target=worker, args=(i,))
        threads.append(thread)
        thread.start()

    for thread in threads:
        thread.join()

# Condition variables for complex synchronization
class BoundedBuffer:
    def __init__(self, capacity):
        self.capacity = capacity
        self.buffer = []
        self.condition = threading.Condition()

    def put(self, item):
        with self.condition:
            while len(self.buffer) >= self.capacity:
                print(f"Buffer full, waiting... (size: {len(self.buffer)})")
                self.condition.wait()

            self.buffer.append(item)
            print(f"Put {item}, buffer size: {len(self.buffer)}")
            self.condition.notify_all()

    def get(self):
        with self.condition:
            while len(self.buffer) == 0:
                print("Buffer empty, waiting...")
                self.condition.wait()

            item = self.buffer.pop(0)
            print(f"Got {item}, buffer size: {len(self.buffer)}")
            self.condition.notify_all()
            return item

# Thread pool pattern
class ThreadPool:
    def __init__(self, num_threads):
        self.num_threads = num_threads
        self.queue = queue.Queue()
        self.threads = []
        self.shutdown = False

        for _ in range(num_threads):
            thread = threading.Thread(target=self._worker)
            thread.daemon = True
            thread.start()
            self.threads.append(thread)

    def _worker(self):
        while not self.shutdown:
            try:
                func, args, kwargs = self.queue.get(timeout=1)
                func(*args, **kwargs)
                self.queue.task_done()
            except queue.Empty:
                continue
            except Exception as e:
                print(f"Worker error: {e}")

    def submit(self, func, *args, **kwargs):
        if not self.shutdown:
            self.queue.put((func, args, kwargs))

    def wait_completion(self):
        self.queue.join()

    def shutdown_pool(self):
        self.shutdown = True
        for thread in self.threads:
            thread.join()

thread_local_example()
```

## Multiprocessing

### Basic Multiprocessing

```python
import multiprocessing
import time
import os

def cpu_intensive_task(n, process_id):
    """CPU-intensive task that benefits from multiprocessing"""
    print(f"Process {process_id} (PID: {os.getpid()}) started")

    total = 0
    for i in range(n):
        total += i ** 2

    print(f"Process {process_id} finished. Result: {total}")
    return total

def multiprocessing_example():
    start_time = time.time()

    # Create processes
    processes = []
    num_processes = multiprocessing.cpu_count()

    for i in range(num_processes):
        process = multiprocessing.Process(
            target=cpu_intensive_task,
            args=(1000000, i)
        )
        processes.append(process)
        process.start()

    # Wait for all processes to complete
    for process in processes:
        process.join()

    end_time = time.time()
    print(f"Multiprocessing completed in {end_time - start_time:.2f} seconds")

if __name__ == '__main__':
    multiprocessing_example()
```

### Process Communication

```python
import multiprocessing
import time
import queue

# Shared memory and communication
def shared_memory_example():
    # Shared Value
    shared_counter = multiprocessing.Value('i', 0)  # 'i' for integer

    # Shared Array
    shared_array = multiprocessing.Array('d', range(10))  # 'd' for double

    def worker_with_shared_memory(counter, array, process_id):
        # Access shared counter
        with counter.get_lock():
            counter.value += 1
            print(f"Process {process_id}: Counter = {counter.value}")

        # Modify shared array
        with array.get_lock():
            for i in range(len(array)):
                array[i] *= 2
            print(f"Process {process_id}: Array modified")

    processes = []
    for i in range(4):
        process = multiprocessing.Process(
            target=worker_with_shared_memory,
            args=(shared_counter, shared_array, i)
        )
        processes.append(process)
        process.start()

    for process in processes:
        process.join()

    print(f"Final counter value: {shared_counter.value}")
    print(f"Final array: {list(shared_array[:])}")

# Queue-based communication
def queue_communication_example():
    # Create queues
    task_queue = multiprocessing.Queue()
    result_queue = multiprocessing.Queue()

    def producer(queue, num_tasks):
        """Producer process"""
        for i in range(num_tasks):
            task = f"task-{i}"
            queue.put(task)
            print(f"Produced: {task}")

    def consumer(task_queue, result_queue, consumer_id):
        \"\"\"Consumer process\"\"\"
        while True:
            try:
                task = task_queue.get(timeout=2)
                if task is None:
                    break

                # Process task
                result = f"processed-{task}-by-{consumer_id}"
                result_queue.put(result)
                print(f"Consumer {consumer_id} processed: {task}")
                time.sleep(0.1)  # Simulate processing time

            except queue.Empty:
                break

    # Start producer
    producer_process = multiprocessing.Process(
        target=producer,
        args=(task_queue, 10)
    )
    producer_process.start()

    # Start consumers
    consumer_processes = []
    for i in range(3):
        process = multiprocessing.Process(
            target=consumer,
            args=(task_queue, result_queue, i)
        )
        consumer_processes.append(process)
        process.start()

    # Wait for producer to finish
    producer_process.join()

    # Signal consumers to stop
    for _ in consumer_processes:
        task_queue.put(None)

    # Wait for consumers to finish
    for process in consumer_processes:
        process.join()

    # Collect results
    results = []
    while not result_queue.empty():
        results.append(result_queue.get())

    print(f"Collected {len(results)} results")

if __name__ == '__main__':
    shared_memory_example()
    queue_communication_example()
```

### Process Pools

```python
import multiprocessing
import time
import math

def compute_factorial(n):
    \"\"\"CPU-intensive computation\"\"\"
    return math.factorial(n)

def pool_example():
    numbers = [5000, 6000, 7000, 8000] * 5

    # Sequential execution
    start_time = time.time()
    sequential_results = [compute_factorial(n) for n in numbers]
    sequential_time = time.time() - start_time

    # Pool execution
    start_time = time.time()
    with multiprocessing.Pool() as pool:
        pool_results = pool.map(compute_factorial, numbers)
    pool_time = time.time() - start_time

    print(f"Sequential: {sequential_time:.2f} seconds")
    print(f"Pool: {pool_time:.2f} seconds")
    print(f"Speedup: {sequential_time / pool_time:.2f}x")

# Advanced pool usage
def advanced_pool_example():
    def slow_square(x):
        time.sleep(0.1)  # Simulate slow operation
        return x * x

    numbers = list(range(50))

    with multiprocessing.Pool(processes=4) as pool:
        # Asynchronous execution
        async_results = []
        for num in numbers:
            result = pool.apply_async(slow_square, (num,))
            async_results.append(result)

        # Get results as they complete
        for i, async_result in enumerate(async_results):
            result = async_result.get(timeout=10)
            print(f"Result {i}: {result}")

    # Using map with multiple arguments
    def multiply(args):
        x, y = args
        return x * y

    pairs = [(i, i+1) for i in range(10)]

    with multiprocessing.Pool() as pool:
        products = pool.map(multiply, pairs)
        print(f"Products: {products}")

if __name__ == '__main__':
    pool_example()
    advanced_pool_example()
```

## Asyncio and Async/Await

### Basic Asyncio

```python
import asyncio
import aiohttp
import time

# Basic async function
async def simple_async_function():
    print("Starting async function")
    await asyncio.sleep(1)  # Non-blocking sleep
    print("Async function completed")
    return "Done"

# Running async code
async def main():
    result = await simple_async_function()
    print(f"Result: {result}")

# asyncio.run(main())  # Uncomment to run

# Concurrent execution with asyncio
async def async_task(name, duration):
    print(f"Task {name} started")
    await asyncio.sleep(duration)
    print(f"Task {name} completed")
    return f"Result from {name}"

async def concurrent_tasks():
    # Create multiple tasks
    tasks = [
        asyncio.create_task(async_task("A", 2)),
        asyncio.create_task(async_task("B", 1)),
        asyncio.create_task(async_task("C", 3))
    ]

    # Wait for all tasks to complete
    results = await asyncio.gather(*tasks)
    print(f"All results: {results}")
```

### Advanced Asyncio Patterns

```python
import asyncio
import aiohttp
import aiofiles
from asyncio import Queue

# Async context managers
class AsyncDatabaseConnection:
    async def __aenter__(self):
        print("Opening database connection")
        await asyncio.sleep(0.1)  # Simulate connection setup
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        print("Closing database connection")
        await asyncio.sleep(0.1)  # Simulate connection cleanup

async def async_context_example():
    async with AsyncDatabaseConnection() as conn:
        print("Using database connection")
        await asyncio.sleep(0.5)

# Async generators
async def async_number_generator(limit):
    for i in range(limit):
        await asyncio.sleep(0.1)  # Simulate async operation
        yield i

async def async_generator_example():
    async for number in async_number_generator(5):
        print(f"Generated: {number}")

# Producer-consumer with asyncio
async def async_producer_consumer():
    queue = Queue(maxsize=10)

    async def producer(queue, num_items):
        for i in range(num_items):
            item = f"item-{i}"
            await queue.put(item)
            print(f"Produced: {item}")
            await asyncio.sleep(0.1)

        # Signal completion
        await queue.put(None)

    async def consumer(queue, consumer_id):
        while True:
            item = await queue.get()
            if item is None:
                # Put None back for other consumers
                await queue.put(None)
                break

            print(f"Consumer {consumer_id} processing: {item}")
            await asyncio.sleep(0.2)  # Simulate processing
            queue.task_done()

    # Start producer and consumers
    producer_task = asyncio.create_task(producer(queue, 10))
    consumer_tasks = [
        asyncio.create_task(consumer(queue, i)) for i in range(3)
    ]

    # Wait for producer to finish
    await producer_task

    # Wait for all items to be processed
    await queue.join()

    # Cancel consumer tasks
    for task in consumer_tasks:
        task.cancel()

    await asyncio.gather(*consumer_tasks, return_exceptions=True)
    print("Producer-consumer completed")

# HTTP requests with asyncio
async def fetch_url(session, url):
    async with session.get(url) as response:
        content = await response.text()
        return len(content)

async def fetch_multiple_urls():
    urls = [
        'https://httpbin.org/delay/1',
        'https://httpbin.org/delay/1',
        'https://httpbin.org/delay/1',
        'https://httpbin.org/delay/1'
    ]

    async with aiohttp.ClientSession() as session:
        start_time = time.time()

        # Concurrent requests
        tasks = [fetch_url(session, url) for url in urls]
        results = await asyncio.gather(*tasks)

        end_time = time.time()

        print(f"Fetched {len(results)} URLs in {end_time - start_time:.2f} seconds")
        print(f"Content lengths: {results}")

# Running all examples
async def run_asyncio_examples():
    print("=== Basic Async ===")
    await main()

    print("\\n=== Concurrent Tasks ===")
    await concurrent_tasks()

    print("\\n=== Async Context Manager ===")
    await async_context_example()

    print("\\n=== Async Generator ===")
    await async_generator_example()

    print("\\n=== Producer-Consumer ===")
    await async_producer_consumer()

    print("\\n=== HTTP Requests ===")
    await fetch_multiple_urls()

# asyncio.run(run_asyncio_examples())  # Uncomment to run
```

### Asyncio Synchronization

```python
import asyncio
import random

# Async locks and synchronization primitives
async def shared_resource_example():
    # Async lock
    lock = asyncio.Lock()
    shared_resource = {"value": 0}

    async def worker(worker_id, lock, resource):
        for _ in range(5):
            async with lock:
                # Critical section
                current_value = resource["value"]
                await asyncio.sleep(0.1)  # Simulate work
                resource["value"] = current_value + 1
                print(f"Worker {worker_id}: value = {resource['value']}")

    # Start multiple workers
    workers = [worker(i, lock, shared_resource) for i in range(3)]
    await asyncio.gather(*workers)

    print(f"Final value: {shared_resource['value']}")

# Semaphores for limiting concurrency
async def semaphore_example():
    # Limit to 2 concurrent operations
    semaphore = asyncio.Semaphore(2)

    async def limited_operation(op_id):
        async with semaphore:
            print(f"Operation {op_id} started (semaphore acquired)")
            await asyncio.sleep(random.uniform(1, 3))
            print(f"Operation {op_id} completed (semaphore released)")

    # Start many operations
    operations = [limited_operation(i) for i in range(10)]
    await asyncio.gather(*operations)

# Events for coordination
async def event_example():
    event = asyncio.Event()

    async def waiter(name):
        print(f"{name} waiting for event")
        await event.wait()
        print(f"{name} received event!")

    async def setter():
        await asyncio.sleep(2)
        print("Setting event")
        event.set()

    # Start waiters and setter
    waiters = [waiter(f"Waiter-{i}") for i in range(3)]
    await asyncio.gather(*waiters, setter())

# asyncio.run(shared_resource_example())
# asyncio.run(semaphore_example())
# asyncio.run(event_example())
```

## Concurrent.futures

### ThreadPoolExecutor and ProcessPoolExecutor

```python
from concurrent.futures import ThreadPoolExecutor, ProcessPoolExecutor, as_completed
import time
import requests

def io_bound_task(url):
    \"\"\"I/O-bound task for ThreadPoolExecutor\"\"\"
    response = requests.get(url)
    return len(response.content)

def cpu_bound_task(n):
    \"\"\"CPU-bound task for ProcessPoolExecutor\"\"\"
    total = 0
    for i in range(n):
        total += i ** 2
    return total

def thread_pool_example():
    urls = [
        'https://httpbin.org/delay/1',
        'https://httpbin.org/delay/1',
        'https://httpbin.org/delay/1',
        'https://httpbin.org/delay/1'
    ]

    # Using ThreadPoolExecutor
    with ThreadPoolExecutor(max_workers=4) as executor:
        # Submit tasks
        future_to_url = {
            executor.submit(io_bound_task, url): url
            for url in urls
        }

        # Process completed futures
        for future in as_completed(future_to_url):
            url = future_to_url[future]
            try:
                result = future.result()
                print(f"URL {url}: {result} bytes")
            except Exception as exc:
                print(f"URL {url} generated exception: {exc}")

def process_pool_example():
    numbers = [1000000, 2000000, 3000000, 4000000]

    # Using ProcessPoolExecutor
    with ProcessPoolExecutor(max_workers=4) as executor:
        # Submit tasks
        future_to_number = {
            executor.submit(cpu_bound_task, num): num
            for num in numbers
        }

        # Process completed futures
        for future in as_completed(future_to_number):
            number = future_to_number[future]
            try:
                result = future.result()
                print(f"Number {number}: result = {result}")
            except Exception as exc:
                print(f"Number {number} generated exception: {exc}")

# Advanced usage with map
def advanced_futures_example():
    def process_with_timeout(n):
        time.sleep(1)
        return n * n

    numbers = list(range(10))

    # Using map with ThreadPoolExecutor
    with ThreadPoolExecutor(max_workers=3) as executor:
        start_time = time.time()
        results = list(executor.map(process_with_timeout, numbers))
        end_time = time.time()

        print(f"Processed {len(numbers)} items in {end_time - start_time:.2f} seconds")
        print(f"Results: {results}")

    # Handling timeouts
    with ThreadPoolExecutor(max_workers=2) as executor:
        future = executor.submit(process_with_timeout, 5)
        try:
            result = future.result(timeout=0.5)  # 0.5 second timeout
            print(f"Result: {result}")
        except TimeoutError:
            print("Task timed out")
            future.cancel()  # Try to cancel (may not work if already running)

thread_pool_example()
process_pool_example()
advanced_futures_example()
```

## Performance Comparison and Best Practices

### Choosing the Right Concurrency Model

```python
import time
import asyncio
import threading
import multiprocessing
from concurrent.futures import ThreadPoolExecutor, ProcessPoolExecutor

# Test functions
def cpu_task(n):
    total = 0
    for i in range(n):
        total += i ** 2
    return total

def io_task():
    time.sleep(0.1)  # Simulate I/O
    return "completed"

async def async_io_task():
    await asyncio.sleep(0.1)
    return "completed"

# Performance comparison
def compare_cpu_bound():
    \"\"\"Compare approaches for CPU-bound tasks\"\"\"
    n = 500000
    num_tasks = 4

    print("=== CPU-Bound Task Comparison ===")

    # Sequential
    start = time.time()
    results = [cpu_task(n) for _ in range(num_tasks)]
    sequential_time = time.time() - start
    print(f"Sequential: {sequential_time:.2f}s")

    # Threading (will be slower due to GIL)
    start = time.time()
    with ThreadPoolExecutor(max_workers=4) as executor:
        results = list(executor.map(cpu_task, [n] * num_tasks))
    threading_time = time.time() - start
    print(f"Threading: {threading_time:.2f}s")

    # Multiprocessing (should be faster)
    start = time.time()
    with ProcessPoolExecutor(max_workers=4) as executor:
        results = list(executor.map(cpu_task, [n] * num_tasks))
    multiprocessing_time = time.time() - start
    print(f"Multiprocessing: {multiprocessing_time:.2f}s")

    print(f"Multiprocessing speedup: {sequential_time / multiprocessing_time:.2f}x")

def compare_io_bound():
    \"\"\"Compare approaches for I/O-bound tasks\"\"\"
    num_tasks = 20

    print("\\n=== I/O-Bound Task Comparison ===")

    # Sequential
    start = time.time()
    results = [io_task() for _ in range(num_tasks)]
    sequential_time = time.time() - start
    print(f"Sequential: {sequential_time:.2f}s")

    # Threading (should be much faster)
    start = time.time()
    with ThreadPoolExecutor(max_workers=10) as executor:
        results = list(executor.map(lambda x: io_task(), range(num_tasks)))
    threading_time = time.time() - start
    print(f"Threading: {threading_time:.2f}s")

    # Asyncio (should be fastest)
    async def async_comparison():
        start = time.time()
        results = await asyncio.gather(*[async_io_task() for _ in range(num_tasks)])
        return time.time() - start

    asyncio_time = asyncio.run(async_comparison())
    print(f"Asyncio: {asyncio_time:.2f}s")

    print(f"Threading speedup: {sequential_time / threading_time:.2f}x")
    print(f"Asyncio speedup: {sequential_time / asyncio_time:.2f}x")

compare_cpu_bound()
compare_io_bound()
```

### Best Practices and Common Patterns

```python
import asyncio
import threading
import multiprocessing
import signal
import sys
from contextlib import contextmanager

# Graceful shutdown pattern
class GracefulKiller:
    \"\"\"Handle graceful shutdown signals\"\"\"
    def __init__(self):
        self.kill_now = threading.Event()
        signal.signal(signal.SIGINT, self._handle_signal)
        signal.signal(signal.SIGTERM, self._handle_signal)

    def _handle_signal(self, signum, frame):
        print(f"Received signal {signum}, shutting down gracefully...")
        self.kill_now.set()

# Thread-safe singleton pattern
class ThreadSafeSingleton:
    _instances = {}
    _lock = threading.Lock()

    def __new__(cls, *args, **kwargs):
        if cls not in cls._instances:
            with cls._lock:
                if cls not in cls._instances:
                    cls._instances[cls] = super().__new__(cls)
        return cls._instances[cls]

# Connection pool pattern
class ConnectionPool:
    def __init__(self, max_connections=10):
        self.max_connections = max_connections
        self.pool = queue.Queue(maxsize=max_connections)
        self.all_connections = set()
        self.lock = threading.Lock()

        # Initialize pool
        for _ in range(max_connections):
            conn = self._create_connection()
            self.pool.put(conn)
            self.all_connections.add(conn)

    def _create_connection(self):
        # Simulate connection creation
        return f"connection-{id(threading.current_thread())}"

    @contextmanager
    def get_connection(self, timeout=10):
        conn = None
        try:
            conn = self.pool.get(timeout=timeout)
            yield conn
        finally:
            if conn:
                self.pool.put(conn)

    def close_all(self):
        with self.lock:
            # Close all connections
            for conn in self.all_connections:
                # conn.close()  # Would close actual connections
                pass

# Error handling in concurrent code
def robust_worker(task_queue, result_queue):
    \"\"\"Worker with proper error handling\"\"\"
    while True:
        try:
            task = task_queue.get(timeout=1)
            if task is None:
                break

            # Process task
            result = process_task(task)
            result_queue.put(('success', result))

        except queue.Empty:
            continue
        except Exception as e:
            result_queue.put(('error', str(e)))
        finally:
            task_queue.task_done()

# Monitoring and metrics
class ConcurrencyMetrics:
    def __init__(self):
        self.lock = threading.Lock()
        self.active_threads = 0
        self.completed_tasks = 0
        self.failed_tasks = 0

    def thread_started(self):
        with self.lock:
            self.active_threads += 1

    def thread_finished(self):
        with self.lock:
            self.active_threads -= 1

    def task_completed(self):
        with self.lock:
            self.completed_tasks += 1

    def task_failed(self):
        with self.lock:
            self.failed_tasks += 1

    def get_stats(self):
        with self.lock:
            return {
                'active_threads': self.active_threads,
                'completed_tasks': self.completed_tasks,
                'failed_tasks': self.failed_tasks
            }

# Async context manager for resource management
class AsyncResourceManager:
    def __init__(self, resource_limit=10):
        self.semaphore = asyncio.Semaphore(resource_limit)
        self.active_resources = 0

    async def __aenter__(self):
        await self.semaphore.acquire()
        self.active_resources += 1
        print(f"Acquired resource (active: {self.active_resources})")
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        self.active_resources -= 1
        self.semaphore.release()
        print(f"Released resource (active: {self.active_resources})")

print("Best practices examples loaded. Use with appropriate runners.")
```

## Decision Matrix: When to Use Each Approach

```python
# Decision guide for concurrency choices

CONCURRENCY_GUIDE = {
    "CPU-bound tasks": {
        "recommendation": "multiprocessing.ProcessPoolExecutor",
        "reasoning": "Bypasses GIL, utilizes multiple CPU cores",
        "alternatives": ["multiprocessing.Pool", "concurrent.futures.ProcessPoolExecutor"]
    },

    "I/O-bound tasks (simple)": {
        "recommendation": "concurrent.futures.ThreadPoolExecutor",
        "reasoning": "Easy to use, good for blocking I/O operations",
        "alternatives": ["threading.Thread", "asyncio for async I/O"]
    },

    "I/O-bound tasks (many concurrent)": {
        "recommendation": "asyncio",
        "reasoning": "Most efficient for high-concurrency I/O operations",
        "alternatives": ["aiohttp", "aiofiles", "asyncpg"]
    },

    "Mixed workloads": {
        "recommendation": "Hybrid approach",
        "reasoning": "Use asyncio for I/O, ProcessPoolExecutor for CPU tasks",
        "pattern": "asyncio.run_in_executor() with ProcessPoolExecutor"
    },

    "Web scraping": {
        "recommendation": "asyncio + aiohttp",
        "reasoning": "Handle many concurrent requests efficiently",
        "alternatives": ["ThreadPoolExecutor + requests"]
    },

    "Data processing": {
        "recommendation": "multiprocessing.Pool",
        "reasoning": "Parallel processing of data chunks",
        "alternatives": ["concurrent.futures.ProcessPoolExecutor"]
    },

    "Real-time systems": {
        "recommendation": "asyncio",
        "reasoning": "Low latency, event-driven programming",
        "considerations": "Avoid blocking operations"
    }
}

def get_concurrency_recommendation(task_type):
    \"\"\"Get concurrency recommendation for task type\"\"\"
    guide = CONCURRENCY_GUIDE.get(task_type)
    if guide:
        print(f"Task Type: {task_type}")
        print(f"Recommendation: {guide['recommendation']}")
        print(f"Reasoning: {guide['reasoning']}")
        if 'alternatives' in guide:
            print(f"Alternatives: {', '.join(guide['alternatives'])}")
        if 'pattern' in guide:
            print(f"Pattern: {guide['pattern']}")
        if 'considerations' in guide:
            print(f"Considerations: {guide['considerations']}")
    else:
        print(f"No specific guidance for: {task_type}")

# Example usage
for task_type in CONCURRENCY_GUIDE:
    get_concurrency_recommendation(task_type)
    print("-" * 50)
```

This comprehensive guide covers all aspects of Python concurrency, from basic threading to advanced asyncio patterns, providing the foundation needed to choose and implement the right concurrency approach for any application.