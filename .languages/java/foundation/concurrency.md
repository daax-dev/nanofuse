# Java Concurrency and Threading - Complete Guide

## Thread Fundamentals

### Creating and Managing Threads

#### Thread Class and Runnable Interface
```java
// Method 1: Extending Thread class
public class CustomThread extends Thread {
    private String threadName;

    public CustomThread(String name) {
        this.threadName = name;
    }

    @Override
    public void run() {
        for (int i = 0; i < 5; i++) {
            System.out.println(threadName + " - Count: " + i);
            try {
                Thread.sleep(1000);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }
    }
}

// Method 2: Implementing Runnable (preferred)
public class CustomRunnable implements Runnable {
    private String taskName;

    public CustomRunnable(String name) {
        this.taskName = name;
    }

    @Override
    public void run() {
        System.out.println("Task " + taskName + " running on " + Thread.currentThread().getName());
        // Task implementation
    }
}

// Usage examples
public class ThreadCreationExample {
    public void createThreads() {
        // Using Thread class
        CustomThread thread1 = new CustomThread("Thread-1");
        thread1.start();

        // Using Runnable interface
        Thread thread2 = new Thread(new CustomRunnable("Task-1"));
        thread2.start();

        // Using lambda expression (Java 8+)
        Thread thread3 = new Thread(() -> {
            System.out.println("Lambda thread running");
        });
        thread3.start();
    }
}
```

### Thread Lifecycle and States
```java
public class ThreadLifecycleExample {
    public void demonstrateThreadStates() {
        Thread worker = new Thread(() -> {
            try {
                System.out.println("Worker thread running");
                Thread.sleep(2000); // TIMED_WAITING

                synchronized(this) {
                    wait(); // WAITING
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        });

        System.out.println("Before start: " + worker.getState()); // NEW

        worker.start();
        System.out.println("After start: " + worker.getState()); // RUNNABLE

        try {
            Thread.sleep(100);
            System.out.println("While sleeping: " + worker.getState()); // TIMED_WAITING

            worker.join(); // Wait for completion
            System.out.println("After join: " + worker.getState()); // TERMINATED
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }
}
```

## Synchronization Mechanisms

### Synchronized Blocks and Methods
```java
public class SynchronizationExample {
    private int counter = 0;
    private final Object lock = new Object();

    // Synchronized method
    public synchronized void incrementCounter() {
        counter++; // Thread-safe increment
    }

    // Synchronized block
    public void incrementWithBlock() {
        synchronized(this) {
            counter++; // Thread-safe increment
        }
    }

    // Using custom lock object
    public void incrementWithCustomLock() {
        synchronized(lock) {
            counter++; // Thread-safe increment
        }
    }

    public synchronized int getCounter() {
        return counter;
    }

    // Static synchronization - class-level lock
    public static synchronized void staticSynchronizedMethod() {
        // Only one thread can execute this across all instances
    }
}
```

### Wait, Notify, and NotifyAll
```java
public class ProducerConsumerExample {
    private Queue<Integer> queue = new LinkedList<>();
    private final int CAPACITY = 10;
    private final Object lock = new Object();

    public void produce() throws InterruptedException {
        synchronized(lock) {
            while (queue.size() == CAPACITY) {
                System.out.println("Queue is full, producer waiting");
                lock.wait(); // Release lock and wait
            }

            int item = (int) (Math.random() * 100);
            queue.offer(item);
            System.out.println("Produced: " + item);

            lock.notifyAll(); // Wake up all waiting threads
        }
    }

    public void consume() throws InterruptedException {
        synchronized(lock) {
            while (queue.isEmpty()) {
                System.out.println("Queue is empty, consumer waiting");
                lock.wait(); // Release lock and wait
            }

            int item = queue.poll();
            System.out.println("Consumed: " + item);

            lock.notifyAll(); // Wake up all waiting threads
        }
    }
}
```

## java.util.concurrent Package

### Executor Framework
```java
public class ExecutorExample {
    public void demonstrateExecutors() {
        // Single thread executor
        ExecutorService singleExecutor = Executors.newSingleThreadExecutor();
        singleExecutor.submit(() -> System.out.println("Single thread task"));

        // Fixed thread pool
        ExecutorService fixedPool = Executors.newFixedThreadPool(4);
        for (int i = 0; i < 10; i++) {
            int taskId = i;
            fixedPool.submit(() -> {
                System.out.println("Task " + taskId + " on " + Thread.currentThread().getName());
            });
        }

        // Cached thread pool
        ExecutorService cachedPool = Executors.newCachedThreadPool();
        cachedPool.submit(() -> System.out.println("Cached pool task"));

        // Scheduled executor
        ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(2);

        // Schedule one-time task
        scheduler.schedule(() -> System.out.println("Delayed task"), 2, TimeUnit.SECONDS);

        // Schedule recurring task
        scheduler.scheduleAtFixedRate(() -> System.out.println("Recurring task"),
                                     1, 3, TimeUnit.SECONDS);

        // Proper shutdown
        shutdownExecutor(singleExecutor);
        shutdownExecutor(fixedPool);
        shutdownExecutor(cachedPool);
        shutdownExecutor(scheduler);
    }

    private void shutdownExecutor(ExecutorService executor) {
        executor.shutdown();
        try {
            if (!executor.awaitTermination(60, TimeUnit.SECONDS)) {
                executor.shutdownNow();
            }
        } catch (InterruptedException e) {
            executor.shutdownNow();
            Thread.currentThread().interrupt();
        }
    }
}
```

### CompletableFuture for Asynchronous Programming
```java
public class CompletableFutureExample {

    public void basicCompletableFuture() {
        // Simple async operation
        CompletableFuture<String> future = CompletableFuture.supplyAsync(() -> {
            try {
                Thread.sleep(1000);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
            return "Hello, CompletableFuture!";
        });

        // Non-blocking processing of result
        future.thenAccept(result -> System.out.println("Result: " + result));

        // Wait for completion
        future.join();
    }

    public void chainedOperations() {
        CompletableFuture<String> future = CompletableFuture
            .supplyAsync(() -> "Hello")
            .thenApply(s -> s + " World")
            .thenApply(String::toUpperCase)
            .thenCompose(s -> CompletableFuture.supplyAsync(() -> s + "!"));

        System.out.println("Final result: " + future.join());
    }

    public void combiningFutures() {
        CompletableFuture<String> future1 = CompletableFuture
            .supplyAsync(() -> "Hello");

        CompletableFuture<String> future2 = CompletableFuture
            .supplyAsync(() -> "World");

        // Combine results
        CompletableFuture<String> combinedFuture = future1
            .thenCombine(future2, (s1, s2) -> s1 + " " + s2);

        System.out.println("Combined: " + combinedFuture.join());

        // Wait for any of multiple futures
        CompletableFuture<Object> anyOf = CompletableFuture
            .anyOf(future1, future2);
        System.out.println("First completed: " + anyOf.join());
    }

    public void errorHandling() {
        CompletableFuture<String> future = CompletableFuture
            .supplyAsync(() -> {
                if (Math.random() > 0.5) {
                    throw new RuntimeException("Random failure");
                }
                return "Success";
            })
            .handle((result, exception) -> {
                if (exception != null) {
                    return "Error: " + exception.getMessage();
                }
                return result;
            });

        System.out.println("Result with error handling: " + future.join());
    }
}
```

### CountDownLatch, CyclicBarrier, and Semaphore
```java
public class SynchronizationUtilitiesExample {

    // CountDownLatch - wait for multiple threads to complete
    public void countDownLatchExample() throws InterruptedException {
        int numThreads = 3;
        CountDownLatch latch = new CountDownLatch(numThreads);

        for (int i = 0; i < numThreads; i++) {
            int workerId = i;
            new Thread(() -> {
                try {
                    System.out.println("Worker " + workerId + " starting");
                    Thread.sleep((long)(Math.random() * 2000));
                    System.out.println("Worker " + workerId + " completed");
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                } finally {
                    latch.countDown(); // Signal completion
                }
            }).start();
        }

        latch.await(); // Wait for all workers to complete
        System.out.println("All workers completed");
    }

    // CyclicBarrier - synchronize threads at a common point
    public void cyclicBarrierExample() {
        int numThreads = 3;
        CyclicBarrier barrier = new CyclicBarrier(numThreads, () -> {
            System.out.println("All threads reached barrier");
        });

        for (int i = 0; i < numThreads; i++) {
            int workerId = i;
            new Thread(() -> {
                try {
                    System.out.println("Worker " + workerId + " working");
                    Thread.sleep((long)(Math.random() * 2000));
                    System.out.println("Worker " + workerId + " waiting at barrier");
                    barrier.await(); // Wait for all threads
                    System.out.println("Worker " + workerId + " continuing");
                } catch (InterruptedException | BrokenBarrierException e) {
                    Thread.currentThread().interrupt();
                }
            }).start();
        }
    }

    // Semaphore - control access to limited resources
    public void semaphoreExample() {
        Semaphore semaphore = new Semaphore(2); // Allow 2 concurrent accesses

        for (int i = 0; i < 5; i++) {
            int workerId = i;
            new Thread(() -> {
                try {
                    semaphore.acquire(); // Acquire permit
                    System.out.println("Worker " + workerId + " acquired resource");
                    Thread.sleep(2000); // Use resource
                    System.out.println("Worker " + workerId + " releasing resource");
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                } finally {
                    semaphore.release(); // Release permit
                }
            }).start();
        }
    }
}
```

## Thread-Safe Collections

### Concurrent Collections
```java
public class ConcurrentCollectionsExample {

    public void demonstrateConcurrentCollections() {
        // ConcurrentHashMap - thread-safe map
        ConcurrentHashMap<String, Integer> concurrentMap = new ConcurrentHashMap<>();

        // Thread-safe operations
        concurrentMap.put("key1", 1);
        concurrentMap.putIfAbsent("key2", 2);
        concurrentMap.compute("key3", (k, v) -> v == null ? 1 : v + 1);

        // CopyOnWriteArrayList - thread-safe for read-heavy scenarios
        CopyOnWriteArrayList<String> copyOnWriteList = new CopyOnWriteArrayList<>();
        copyOnWriteList.add("item1");
        copyOnWriteList.add("item2");

        // BlockingQueue - producer-consumer scenarios
        BlockingQueue<String> queue = new ArrayBlockingQueue<>(10);

        // Producer
        new Thread(() -> {
            try {
                for (int i = 0; i < 5; i++) {
                    queue.put("Item " + i); // Blocks if queue is full
                    System.out.println("Produced: Item " + i);
                    Thread.sleep(1000);
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }).start();

        // Consumer
        new Thread(() -> {
            try {
                for (int i = 0; i < 5; i++) {
                    String item = queue.take(); // Blocks if queue is empty
                    System.out.println("Consumed: " + item);
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }).start();
    }
}
```

## Project Loom - Virtual Threads (Java 21+)

### Virtual Threads Basics
```java
public class VirtualThreadsExample {

    // Creating virtual threads
    public void basicVirtualThreads() throws InterruptedException {
        // Create and start virtual thread
        Thread virtualThread = Thread.ofVirtual().name("virtual-1").start(() -> {
            System.out.println("Virtual thread: " + Thread.currentThread());
        });

        virtualThread.join();

        // Using thread builder
        Thread.Builder builder = Thread.ofVirtual().name("virtual-", 0);

        for (int i = 0; i < 10; i++) {
            builder.start(() -> {
                System.out.println("Running on: " + Thread.currentThread());
                try {
                    Thread.sleep(1000); // Non-blocking for virtual threads
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                }
            });
        }
    }

    // Executor with virtual threads
    public void virtualThreadExecutor() throws InterruptedException {
        try (ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor()) {

            // Submit many tasks - virtual threads scale better
            for (int i = 0; i < 10000; i++) {
                int taskId = i;
                executor.submit(() -> {
                    try {
                        Thread.sleep(Duration.ofSeconds(1));
                        System.out.println("Task " + taskId + " completed");
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                    }
                });
            }

            // Executor automatically closes with try-with-resources
        }
    }

    // Structured Concurrency (Preview)
    public void structuredConcurrency() {
        try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {

            // Start multiple concurrent tasks
            Future<String> task1 = scope.fork(() -> {
                Thread.sleep(Duration.ofSeconds(1));
                return "Result 1";
            });

            Future<String> task2 = scope.fork(() -> {
                Thread.sleep(Duration.ofSeconds(2));
                return "Result 2";
            });

            Future<String> task3 = scope.fork(() -> {
                Thread.sleep(Duration.ofSeconds(1));
                return "Result 3";
            });

            // Wait for all tasks to complete or any to fail
            scope.join();
            scope.throwIfFailed();

            // All tasks succeeded - get results
            System.out.println("Results: " + task1.resultNow() +
                             ", " + task2.resultNow() +
                             ", " + task3.resultNow());

        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        } catch (ExecutionException e) {
            System.err.println("Task failed: " + e.getCause());
        }
    }
}
```

## Atomic Operations and Lock-Free Programming

### Atomic Classes
```java
public class AtomicExample {
    private AtomicInteger counter = new AtomicInteger(0);
    private AtomicReference<String> reference = new AtomicReference<>("initial");
    private AtomicLong timestamp = new AtomicLong(System.currentTimeMillis());

    public void atomicOperations() {
        // Atomic increment
        int newValue = counter.incrementAndGet();
        System.out.println("New counter value: " + newValue);

        // Compare and set
        boolean updated = counter.compareAndSet(1, 2);
        System.out.println("Counter updated: " + updated);

        // Atomic reference operations
        reference.set("updated");
        String previous = reference.getAndSet("final");
        System.out.println("Previous reference: " + previous);

        // Update with function
        counter.updateAndGet(current -> current * 2);
        reference.updateAndGet(String::toUpperCase);
    }

    // Lock-free data structure example
    public static class LockFreeStack<T> {
        private AtomicReference<Node<T>> top = new AtomicReference<>();

        private static class Node<T> {
            final T data;
            final Node<T> next;

            Node(T data, Node<T> next) {
                this.data = data;
                this.next = next;
            }
        }

        public void push(T item) {
            Node<T> newNode;
            Node<T> currentTop;

            do {
                currentTop = top.get();
                newNode = new Node<>(item, currentTop);
            } while (!top.compareAndSet(currentTop, newNode));
        }

        public T pop() {
            Node<T> currentTop;
            Node<T> newTop;

            do {
                currentTop = top.get();
                if (currentTop == null) {
                    return null;
                }
                newTop = currentTop.next;
            } while (!top.compareAndSet(currentTop, newTop));

            return currentTop.data;
        }
    }
}
```

## Advanced Synchronization

### ReadWriteLock
```java
public class ReadWriteLockExample {
    private final ReadWriteLock lock = new ReentrantReadWriteLock();
    private final Lock readLock = lock.readLock();
    private final Lock writeLock = lock.writeLock();
    private Map<String, String> cache = new HashMap<>();

    public String read(String key) {
        readLock.lock();
        try {
            return cache.get(key);
        } finally {
            readLock.unlock();
        }
    }

    public void write(String key, String value) {
        writeLock.lock();
        try {
            cache.put(key, value);
        } finally {
            writeLock.unlock();
        }
    }

    public void readMultiple(List<String> keys) {
        readLock.lock();
        try {
            // Multiple readers can execute concurrently
            for (String key : keys) {
                System.out.println(key + ": " + cache.get(key));
            }
        } finally {
            readLock.unlock();
        }
    }
}
```

### StampedLock (Java 8+)
```java
public class StampedLockExample {
    private final StampedLock lock = new StampedLock();
    private double x, y;

    // Optimistic read
    public double distanceFromOrigin() {
        long stamp = lock.tryOptimisticRead();
        double currentX = x, currentY = y; // Read variables

        if (!lock.validate(stamp)) {
            // Fallback to read lock if optimistic read failed
            stamp = lock.readLock();
            try {
                currentX = x;
                currentY = y;
            } finally {
                lock.unlockRead(stamp);
            }
        }

        return Math.sqrt(currentX * currentX + currentY * currentY);
    }

    // Write operation
    public void moveToOrigin() {
        long stamp = lock.writeLock();
        try {
            x = 0.0;
            y = 0.0;
        } finally {
            lock.unlockWrite(stamp);
        }
    }

    // Convert read lock to write lock
    public void moveIfAtOrigin(double newX, double newY) {
        long stamp = lock.readLock();
        try {
            while (x == 0.0 && y == 0.0) {
                long writeStamp = lock.tryConvertToWriteLock(stamp);
                if (writeStamp != 0L) {
                    stamp = writeStamp;
                    x = newX;
                    y = newY;
                    break;
                } else {
                    lock.unlockRead(stamp);
                    stamp = lock.writeLock();
                }
            }
        } finally {
            lock.unlock(stamp);
        }
    }
}
```

## Performance Considerations and Best Practices

### Thread Pool Sizing
```java
public class ThreadPoolSizing {

    public ExecutorService createOptimalThreadPool() {
        int cpuCores = Runtime.getRuntime().availableProcessors();

        // CPU-intensive tasks: cores + 1
        int cpuIntensiveThreads = cpuCores + 1;

        // I/O-intensive tasks: cores * (1 + wait_time/service_time)
        // Typical ratio for I/O bound: 2 * cores
        int ioIntensiveThreads = cpuCores * 2;

        // Create thread pool based on workload type
        return new ThreadPoolExecutor(
            cpuCores,                           // Core pool size
            ioIntensiveThreads,                 // Maximum pool size
            60L,                                // Keep alive time
            TimeUnit.SECONDS,
            new LinkedBlockingQueue<>(1000),    // Work queue
            new ThreadFactoryBuilder()
                .setNameFormat("worker-%d")
                .setDaemon(true)
                .build(),
            new ThreadPoolExecutor.CallerRunsPolicy() // Rejection policy
        );
    }
}
```

### Avoiding Common Concurrency Pitfalls
```java
public class ConcurrencyBestPractices {

    // Avoid: Synchronizing on mutable objects
    private List<String> list = new ArrayList<>();

    // Wrong - synchronizing on mutable object
    public void wrongSynchronization() {
        synchronized(list) { // Bad: list reference might change
            list.add("item");
        }
    }

    // Correct - using immutable lock object
    private final Object lock = new Object();

    public void correctSynchronization() {
        synchronized(lock) { // Good: lock object never changes
            list.add("item");
        }
    }

    // Double-checked locking pattern (correct implementation)
    private volatile Singleton instance;

    public Singleton getSingleton() {
        if (instance == null) {
            synchronized(this) {
                if (instance == null) {
                    instance = new Singleton();
                }
            }
        }
        return instance;
    }

    // Thread-safe initialization using enum
    public enum ThreadSafeSingleton {
        INSTANCE;

        public void doSomething() {
            // Thread-safe by JVM guarantee
        }
    }
}
```

This comprehensive guide covers Java concurrency from basic threading concepts to advanced features like virtual threads, providing practical examples for building robust concurrent applications.