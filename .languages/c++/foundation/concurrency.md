# C++ Concurrency and Parallelism

## Overview

Modern C++ provides comprehensive concurrency support through multiple layers: low-level primitives, high-level abstractions, and parallel algorithms. This guide covers threading, synchronization, atomic operations, lock-free programming, and coroutines.

## Threading Fundamentals

### Basic Thread Management

```cpp
#include <thread>
#include <iostream>
#include <vector>
#include <functional>

// Basic thread creation
void worker_function(int id) {
    std::cout << "Worker " << id << " running on thread "
              << std::this_thread::get_id() << std::endl;
}

int main() {
    // Creating threads
    std::thread t1(worker_function, 1);
    std::thread t2([](int id) {
        std::cout << "Lambda worker " << id << std::endl;
    }, 2);

    // Wait for completion
    t1.join();
    t2.join();

    // Detached thread (fire and forget)
    std::thread t3(worker_function, 3);
    t3.detach();

    return 0;
}
```

### Thread Pool Implementation

```cpp
#include <thread>
#include <queue>
#include <mutex>
#include <condition_variable>
#include <functional>
#include <future>

class ThreadPool {
private:
    std::vector<std::thread> workers;
    std::queue<std::function<void()>> tasks;
    std::mutex queue_mutex;
    std::condition_variable condition;
    bool stop = false;

public:
    ThreadPool(size_t threads) {
        for (size_t i = 0; i < threads; ++i) {
            workers.emplace_back([this] {
                while (true) {
                    std::function<void()> task;

                    {
                        std::unique_lock<std::mutex> lock(queue_mutex);
                        condition.wait(lock, [this] {
                            return stop || !tasks.empty();
                        });

                        if (stop && tasks.empty()) return;

                        task = std::move(tasks.front());
                        tasks.pop();
                    }

                    task();
                }
            });
        }
    }

    template<class F, class... Args>
    auto enqueue(F&& f, Args&&... args)
        -> std::future<typename std::result_of<F(Args...)>::type> {

        using return_type = typename std::result_of<F(Args...)>::type;

        auto task = std::make_shared<std::packaged_task<return_type()>>(
            std::bind(std::forward<F>(f), std::forward<Args>(args)...)
        );

        std::future<return_type> result = task->get_future();

        {
            std::unique_lock<std::mutex> lock(queue_mutex);
            if (stop) {
                throw std::runtime_error("enqueue on stopped ThreadPool");
            }

            tasks.emplace([task](){ (*task)(); });
        }

        condition.notify_one();
        return result;
    }

    ~ThreadPool() {
        {
            std::unique_lock<std::mutex> lock(queue_mutex);
            stop = true;
        }

        condition.notify_all();

        for (std::thread &worker : workers) {
            worker.join();
        }
    }
};

// Usage example
void use_thread_pool() {
    ThreadPool pool(4);
    std::vector<std::future<int>> results;

    for (int i = 0; i < 8; ++i) {
        results.emplace_back(
            pool.enqueue([i] {
                std::this_thread::sleep_for(std::chrono::seconds(1));
                return i * i;
            })
        );
    }

    for (auto&& result : results) {
        std::cout << result.get() << ' ';
    }
}
```

## Synchronization Primitives

### Mutexes and Locks

```cpp
#include <mutex>
#include <shared_mutex>
#include <lock_guard>
#include <unique_lock>

class ThreadSafeCounter {
private:
    mutable std::mutex mtx;
    int count = 0;

public:
    void increment() {
        std::lock_guard<std::mutex> lock(mtx);
        ++count;
    }

    void decrement() {
        std::lock_guard<std::mutex> lock(mtx);
        --count;
    }

    int get() const {
        std::lock_guard<std::mutex> lock(mtx);
        return count;
    }
};

// Reader-Writer Lock
class ReadWriteData {
private:
    mutable std::shared_mutex rw_mutex;
    std::string data;

public:
    // Multiple readers can access simultaneously
    std::string read() const {
        std::shared_lock<std::shared_mutex> lock(rw_mutex);
        return data;
    }

    // Exclusive writer access
    void write(const std::string& new_data) {
        std::unique_lock<std::shared_mutex> lock(rw_mutex);
        data = new_data;
    }
};

// Scoped Locking with RAII
class BankAccount {
private:
    mutable std::mutex mtx;
    double balance;

public:
    BankAccount(double initial_balance) : balance(initial_balance) {}

    void transfer(BankAccount& to, double amount) {
        // Avoid deadlock with ordered locking
        std::lock(mtx, to.mtx);
        std::lock_guard<std::mutex> lock1(mtx, std::adopt_lock);
        std::lock_guard<std::mutex> lock2(to.mtx, std::adopt_lock);

        if (balance >= amount) {
            balance -= amount;
            to.balance += amount;
        }
    }

    double get_balance() const {
        std::lock_guard<std::mutex> lock(mtx);
        return balance;
    }
};
```

### Condition Variables

```cpp
#include <condition_variable>
#include <queue>

template<typename T>
class ThreadSafeQueue {
private:
    mutable std::mutex mtx;
    std::queue<T> data_queue;
    std::condition_variable condition;

public:
    void push(T item) {
        std::lock_guard<std::mutex> lock(mtx);
        data_queue.push(std::move(item));
        condition.notify_one();
    }

    void wait_and_pop(T& item) {
        std::unique_lock<std::mutex> lock(mtx);
        condition.wait(lock, [this] { return !data_queue.empty(); });
        item = std::move(data_queue.front());
        data_queue.pop();
    }

    std::shared_ptr<T> wait_and_pop() {
        std::unique_lock<std::mutex> lock(mtx);
        condition.wait(lock, [this] { return !data_queue.empty(); });
        auto result = std::make_shared<T>(std::move(data_queue.front()));
        data_queue.pop();
        return result;
    }

    bool try_pop(T& item) {
        std::lock_guard<std::mutex> lock(mtx);
        if (data_queue.empty()) return false;
        item = std::move(data_queue.front());
        data_queue.pop();
        return true;
    }

    bool empty() const {
        std::lock_guard<std::mutex> lock(mtx);
        return data_queue.empty();
    }

    size_t size() const {
        std::lock_guard<std::mutex> lock(mtx);
        return data_queue.size();
    }
};

// Producer-Consumer Example
void producer_consumer_example() {
    ThreadSafeQueue<int> queue;

    // Producer thread
    std::thread producer([&queue] {
        for (int i = 0; i < 100; ++i) {
            queue.push(i);
            std::this_thread::sleep_for(std::chrono::milliseconds(10));
        }
    });

    // Consumer thread
    std::thread consumer([&queue] {
        int item;
        for (int i = 0; i < 100; ++i) {
            queue.wait_and_pop(item);
            std::cout << "Consumed: " << item << std::endl;
        }
    });

    producer.join();
    consumer.join();
}
```

## Atomic Operations

### Basic Atomics

```cpp
#include <atomic>
#include <memory>

class AtomicCounter {
private:
    std::atomic<int> count{0};

public:
    void increment() {
        count.fetch_add(1, std::memory_order_relaxed);
    }

    void decrement() {
        count.fetch_sub(1, std::memory_order_relaxed);
    }

    int get() const {
        return count.load(std::memory_order_relaxed);
    }

    // Compare and swap
    bool compare_exchange_weak(int& expected, int desired) {
        return count.compare_exchange_weak(expected, desired,
                                         std::memory_order_release,
                                         std::memory_order_relaxed);
    }
};

// Atomic operations with different memory orderings
class AtomicFlag {
private:
    std::atomic<bool> flag{false};
    std::atomic<int> data{0};

public:
    void set_data(int value) {
        data.store(value, std::memory_order_relaxed);
        flag.store(true, std::memory_order_release);  // Synchronizes-with acquire
    }

    int get_data() {
        while (!flag.load(std::memory_order_acquire)) {  // Synchronized-with release
            std::this_thread::yield();
        }
        return data.load(std::memory_order_relaxed);
    }
};

// Atomic smart pointers
template<typename T>
class AtomicSharedPtr {
private:
    std::atomic<std::shared_ptr<T>> ptr;

public:
    AtomicSharedPtr() = default;
    AtomicSharedPtr(std::shared_ptr<T> p) : ptr(std::move(p)) {}

    std::shared_ptr<T> load() const {
        return std::atomic_load(&ptr);
    }

    void store(std::shared_ptr<T> p) {
        std::atomic_store(&ptr, std::move(p));
    }

    std::shared_ptr<T> exchange(std::shared_ptr<T> p) {
        return std::atomic_exchange(&ptr, std::move(p));
    }

    bool compare_exchange_weak(std::shared_ptr<T>& expected,
                              std::shared_ptr<T> desired) {
        return std::atomic_compare_exchange_weak(&ptr, &expected, std::move(desired));
    }
};
```

### Memory Ordering

```cpp
#include <atomic>

class MemoryOrderingExample {
private:
    std::atomic<int> x{0};
    std::atomic<int> y{0};

public:
    // Sequential consistency (default)
    void sequential_consistency() {
        x.store(1);  // std::memory_order_seq_cst
        int r1 = y.load();  // std::memory_order_seq_cst
    }

    // Relaxed ordering
    void relaxed_ordering() {
        x.store(1, std::memory_order_relaxed);
        int r1 = y.load(std::memory_order_relaxed);
    }

    // Release-Acquire ordering
    void release_acquire() {
        x.store(1, std::memory_order_release);
        int r1 = y.load(std::memory_order_acquire);
    }

    // Consume ordering (deprecated in C++17)
    void consume_ordering() {
        x.store(1, std::memory_order_release);
        int r1 = y.load(std::memory_order_consume);
    }

    // Fence operations
    void fence_example() {
        x.store(1, std::memory_order_relaxed);
        std::atomic_thread_fence(std::memory_order_release);
        int r1 = y.load(std::memory_order_relaxed);
        std::atomic_thread_fence(std::memory_order_acquire);
    }
};
```

## Lock-Free Programming

### Lock-Free Stack

```cpp
#include <atomic>
#include <memory>

template<typename T>
class LockFreeStack {
private:
    struct Node {
        T data;
        std::atomic<Node*> next;

        Node(T const& data_) : data(data_), next(nullptr) {}
    };

    std::atomic<Node*> head{nullptr};
    std::atomic<size_t> threads_in_pop{0};
    std::atomic<Node*> to_be_deleted{nullptr};

    static void delete_nodes(Node* nodes) {
        while (nodes) {
            Node* next = nodes->next;
            delete nodes;
            nodes = next;
        }
    }

    void try_reclaim(Node* old_head) {
        if (threads_in_pop == 1) {
            Node* nodes_to_delete = to_be_deleted.exchange(nullptr);
            if (!--threads_in_pop) {
                delete_nodes(nodes_to_delete);
            } else if (nodes_to_delete) {
                chain_pending_nodes(nodes_to_delete);
            }
            delete old_head;
        } else {
            chain_pending_node(old_head);
            --threads_in_pop;
        }
    }

    void chain_pending_nodes(Node* nodes) {
        Node* last = nodes;
        while (Node* const next = last->next) {
            last = next;
        }
        chain_pending_nodes(nodes, last);
    }

    void chain_pending_nodes(Node* first, Node* last) {
        last->next = to_be_deleted;
        while (!to_be_deleted.compare_exchange_weak(last->next, first));
    }

    void chain_pending_node(Node* n) {
        chain_pending_nodes(n, n);
    }

public:
    void push(T const& data) {
        Node* const new_node = new Node(data);
        new_node->next = head.load();
        while (!head.compare_exchange_weak(new_node->next, new_node));
    }

    std::shared_ptr<T> pop() {
        ++threads_in_pop;
        Node* old_head = head.load();

        while (old_head &&
               !head.compare_exchange_weak(old_head, old_head->next));

        std::shared_ptr<T> res;
        if (old_head) {
            res = std::make_shared<T>(old_head->data);
        }

        try_reclaim(old_head);
        return res;
    }

    ~LockFreeStack() {
        while (pop());
        delete_nodes(to_be_deleted);
    }
};
```

### Lock-Free Queue

```cpp
template<typename T>
class LockFreeQueue {
private:
    struct Node {
        std::atomic<T*> data;
        std::atomic<Node*> next;

        Node() : data(nullptr), next(nullptr) {}
    };

    std::atomic<Node*> head;
    std::atomic<Node*> tail;

public:
    LockFreeQueue() : head(new Node), tail(head.load()) {}

    ~LockFreeQueue() {
        while (Node* const old_head = head.load()) {
            head = old_head->next;
            delete old_head;
        }
    }

    void enqueue(T item) {
        Node* const new_node = new Node;
        T* const data = new T(std::move(item));
        new_node->data.store(data);

        Node* prev_tail = tail.load();
        while (true) {
            Node* const tail_next = prev_tail->next.load();
            if (prev_tail == tail.load()) {
                if (tail_next == nullptr) {
                    if (prev_tail->next.compare_exchange_weak(tail_next, new_node)) {
                        break;
                    }
                } else {
                    tail.compare_exchange_weak(prev_tail, tail_next);
                }
            }
        }
        tail.compare_exchange_weak(prev_tail, new_node);
    }

    std::unique_ptr<T> dequeue() {
        Node* old_head = head.load();
        while (true) {
            Node* const tail_node = tail.load();
            Node* const head_next = old_head->next.load();

            if (old_head == head.load()) {
                if (old_head == tail_node) {
                    if (head_next == nullptr) {
                        return std::unique_ptr<T>();
                    }
                    tail.compare_exchange_weak(tail_node, head_next);
                } else {
                    if (head_next == nullptr) {
                        continue;
                    }
                    T* const data = head_next->data.load();
                    if (data == nullptr) {
                        continue;
                    }
                    if (head.compare_exchange_weak(old_head, head_next)) {
                        std::unique_ptr<T> result(data);
                        delete old_head;
                        return result;
                    }
                }
            }
        }
    }
};
```

## Parallel Algorithms

### STL Parallel Algorithms (C++17)

```cpp
#include <algorithm>
#include <execution>
#include <vector>
#include <numeric>

void parallel_algorithms_example() {
    std::vector<int> data(1000000);
    std::iota(data.begin(), data.end(), 0);

    // Parallel sort
    std::sort(std::execution::par_unseq, data.begin(), data.end(),
              std::greater<int>());

    // Parallel for_each
    std::for_each(std::execution::par, data.begin(), data.end(),
                  [](int& x) { x = x * x; });

    // Parallel transform
    std::vector<int> result(data.size());
    std::transform(std::execution::par_unseq,
                   data.begin(), data.end(), result.begin(),
                   [](int x) { return x * 2; });

    // Parallel reduce
    int sum = std::reduce(std::execution::par,
                         data.begin(), data.end(), 0);

    // Parallel find
    auto it = std::find(std::execution::par,
                       data.begin(), data.end(), 42);
}

// Custom parallel algorithm
template<typename Iterator, typename Func>
void parallel_for_each(Iterator first, Iterator last, Func f) {
    const auto length = std::distance(first, last);
    if (length == 0) return;

    const auto min_per_thread = 25;
    const auto max_threads = (length + min_per_thread - 1) / min_per_thread;
    const auto hardware_threads = std::thread::hardware_concurrency();
    const auto num_threads = std::min(hardware_threads != 0 ? hardware_threads : 2,
                                     static_cast<unsigned long>(max_threads));

    const auto block_size = length / num_threads;

    std::vector<std::thread> threads(num_threads - 1);

    Iterator block_start = first;
    for (unsigned long i = 0; i < (num_threads - 1); ++i) {
        Iterator block_end = block_start;
        std::advance(block_end, block_size);

        threads[i] = std::thread([block_start, block_end, f] {
            std::for_each(block_start, block_end, f);
        });

        block_start = block_end;
    }

    std::for_each(block_start, last, f);

    for (auto& t : threads) {
        t.join();
    }
}
```

## Coroutines (C++20)

### Basic Coroutine Implementation

```cpp
#include <coroutine>
#include <exception>

// Simple generator coroutine
template<typename T>
class Generator {
public:
    struct promise_type {
        T current_value;

        Generator get_return_object() {
            return Generator{std::coroutine_handle<promise_type>::from_promise(*this)};
        }

        std::suspend_always initial_suspend() { return {}; }
        std::suspend_always final_suspend() noexcept { return {}; }

        std::suspend_always yield_value(T value) {
            current_value = value;
            return {};
        }

        void return_void() {}
        void unhandled_exception() { std::terminate(); }
    };

private:
    std::coroutine_handle<promise_type> h;

public:
    explicit Generator(std::coroutine_handle<promise_type> handle) : h(handle) {}

    ~Generator() {
        if (h) h.destroy();
    }

    Generator(const Generator&) = delete;
    Generator& operator=(const Generator&) = delete;

    Generator(Generator&& other) noexcept : h(std::exchange(other.h, {})) {}
    Generator& operator=(Generator&& other) noexcept {
        if (this != &other) {
            if (h) h.destroy();
            h = std::exchange(other.h, {});
        }
        return *this;
    }

    T next() {
        h.resume();
        return h.promise().current_value;
    }

    bool done() const {
        return h.done();
    }
};

// Usage example
Generator<int> fibonacci() {
    int a = 0, b = 1;
    while (true) {
        co_yield a;
        auto temp = a;
        a = b;
        b += temp;
    }
}

// Async coroutine
#include <future>

template<typename T>
class Task {
public:
    struct promise_type {
        T result;
        std::exception_ptr exception;

        Task get_return_object() {
            return Task{std::coroutine_handle<promise_type>::from_promise(*this)};
        }

        std::suspend_never initial_suspend() { return {}; }
        std::suspend_never final_suspend() noexcept { return {}; }

        void return_value(T value) {
            result = value;
        }

        void unhandled_exception() {
            exception = std::current_exception();
        }
    };

private:
    std::coroutine_handle<promise_type> h;

public:
    explicit Task(std::coroutine_handle<promise_type> handle) : h(handle) {}

    ~Task() {
        if (h) h.destroy();
    }

    T get() {
        if (!h.done()) {
            h.resume();
        }

        if (h.promise().exception) {
            std::rethrow_exception(h.promise().exception);
        }

        return h.promise().result;
    }

    bool is_ready() const {
        return h.done();
    }
};

// Async function example
Task<int> async_computation(int x) {
    // Simulate async work
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    co_return x * x;
}
```

### Awaitable Types

```cpp
#include <chrono>

// Custom awaitable
class TimerAwaitable {
private:
    std::chrono::milliseconds duration;

public:
    explicit TimerAwaitable(std::chrono::milliseconds d) : duration(d) {}

    bool await_ready() const { return false; }

    void await_suspend(std::coroutine_handle<> h) {
        std::thread([h, d = duration] {
            std::this_thread::sleep_for(d);
            h.resume();
        }).detach();
    }

    void await_resume() {}
};

TimerAwaitable sleep_for(std::chrono::milliseconds duration) {
    return TimerAwaitable{duration};
}

// Coroutine using custom awaitable
Task<void> example_coroutine() {
    std::cout << "Starting work..." << std::endl;
    co_await sleep_for(std::chrono::milliseconds(1000));
    std::cout << "Work completed!" << std::endl;
}
```

## Synchronization Patterns

### Barrier Synchronization

```cpp
#include <barrier>

void barrier_example() {
    const int num_threads = 4;
    std::barrier sync_point(num_threads);

    auto worker = [&sync_point](int id) {
        // Phase 1
        std::cout << "Thread " << id << " phase 1\n";
        sync_point.arrive_and_wait();

        // Phase 2
        std::cout << "Thread " << id << " phase 2\n";
        sync_point.arrive_and_wait();

        // Phase 3
        std::cout << "Thread " << id << " phase 3\n";
    };

    std::vector<std::thread> threads;
    for (int i = 0; i < num_threads; ++i) {
        threads.emplace_back(worker, i);
    }

    for (auto& t : threads) {
        t.join();
    }
}
```

### Semaphore (C++20)

```cpp
#include <semaphore>

class ResourcePool {
private:
    std::counting_semaphore<10> semaphore{3}; // 3 resources available
    std::vector<Resource> resources;
    std::mutex resource_mutex;

public:
    Resource* acquire() {
        semaphore.acquire(); // Wait for available resource

        std::lock_guard<std::mutex> lock(resource_mutex);
        // Find and return available resource
        for (auto& resource : resources) {
            if (!resource.in_use) {
                resource.in_use = true;
                return &resource;
            }
        }
        return nullptr;
    }

    void release(Resource* resource) {
        {
            std::lock_guard<std::mutex> lock(resource_mutex);
            resource->in_use = false;
        }
        semaphore.release(); // Signal that resource is available
    }
};
```

### Latch (C++20)

```cpp
#include <latch>

void latch_example() {
    const int num_workers = 5;
    std::latch work_done(num_workers);
    std::latch start_work(1);

    auto worker = [&](int id) {
        // Wait for signal to start
        start_work.wait();

        // Do work
        std::cout << "Worker " << id << " working...\n";
        std::this_thread::sleep_for(std::chrono::milliseconds(100 * id));

        // Signal completion
        std::cout << "Worker " << id << " done\n";
        work_done.count_down();
    };

    // Start all worker threads
    std::vector<std::thread> workers;
    for (int i = 0; i < num_workers; ++i) {
        workers.emplace_back(worker, i);
    }

    // Signal all workers to start
    std::cout << "Starting all workers...\n";
    start_work.count_down();

    // Wait for all workers to complete
    work_done.wait();
    std::cout << "All workers completed\n";

    for (auto& w : workers) {
        w.join();
    }
}
```

## Performance Considerations

### Cache-Friendly Design

```cpp
// False sharing prevention
struct alignas(64) CacheLinePadded {
    std::atomic<int> counter{0};
    // Padding to prevent false sharing
};

// Lock-free design patterns
class SpinLock {
private:
    std::atomic_flag locked = ATOMIC_FLAG_INIT;

public:
    void lock() {
        while (locked.test_and_set(std::memory_order_acquire)) {
            // Busy wait with pause instruction
            #ifdef __x86_64__
            __builtin_ia32_pause();
            #else
            std::this_thread::yield();
            #endif
        }
    }

    void unlock() {
        locked.clear(std::memory_order_release);
    }
};

// NUMA-aware allocation
template<typename T>
class NumaAllocator {
public:
    using value_type = T;

    T* allocate(size_t n) {
        // NUMA-aware allocation logic
        void* ptr = numa_alloc_local(n * sizeof(T));
        if (!ptr) throw std::bad_alloc();
        return static_cast<T*>(ptr);
    }

    void deallocate(T* p, size_t n) {
        numa_free(p, n * sizeof(T));
    }
};
```

### Thread Affinity and Priority

```cpp
#include <pthread.h>
#include <sched.h>

class ThreadManager {
public:
    static void set_thread_affinity(std::thread& t, int cpu_id) {
        cpu_set_t cpuset;
        CPU_ZERO(&cpuset);
        CPU_SET(cpu_id, &cpuset);

        pthread_t native_handle = t.native_handle();
        pthread_setaffinity_np(native_handle, sizeof(cpu_set_t), &cpuset);
    }

    static void set_thread_priority(std::thread& t, int priority) {
        sched_param sch_params;
        sch_params.sched_priority = priority;

        pthread_t native_handle = t.native_handle();
        pthread_setschedparam(native_handle, SCHED_FIFO, &sch_params);
    }

    static void set_thread_name(std::thread& t, const std::string& name) {
        pthread_t native_handle = t.native_handle();
        pthread_setname_np(native_handle, name.c_str());
    }
};
```

## Best Practices

1. **Use RAII for Resource Management**
   - Always use lock guards instead of manual locking
   - Prefer unique_lock over lock_guard for complex scenarios

2. **Minimize Lock Contention**
   - Use reader-writer locks for read-heavy workloads
   - Consider lock-free data structures for high-contention scenarios

3. **Avoid Deadlocks**
   - Always acquire locks in the same order
   - Use std::lock for multiple mutex acquisition
   - Consider using std::scoped_lock (C++17)

4. **Memory Ordering**
   - Use relaxed ordering for simple counters
   - Use acquire-release for synchronization
   - Default to sequential consistency when in doubt

5. **Thread Pool Design**
   - Size thread pools based on hardware concurrency
   - Implement work stealing for load balancing
   - Use futures for result handling

6. **Atomic Operations**
   - Prefer atomic operations over mutexes for simple operations
   - Use compare-and-swap for lock-free algorithms
   - Be aware of ABA problem in lock-free code

This comprehensive guide covers modern C++ concurrency from basic threading to advanced coroutines and lock-free programming, providing practical examples and performance considerations for production use.