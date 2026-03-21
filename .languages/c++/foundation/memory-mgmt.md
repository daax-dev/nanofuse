# C++ Memory Management Comprehensive Guide

Complete guide to memory management in C++, covering stack/heap allocation, RAII, smart pointers, custom allocators, and modern memory safety patterns.

## Memory Layout and Allocation

### Stack vs Heap Allocation

```cpp
#include <memory>
#include <vector>

void demonstrate_stack_heap() {
    // Stack allocation - automatic storage
    int stack_var = 42;                    // Destroyed when leaving scope
    int stack_array[100];                  // Fixed size, fast allocation
    std::string stack_string("hello");    // Constructor called automatically

    // Heap allocation - dynamic storage
    int* heap_var = new int(42);          // Manual memory management
    int* heap_array = new int[100];       // Variable size, slower allocation
    std::string* heap_string = new std::string("hello");

    // Modern heap allocation with smart pointers
    auto smart_var = std::make_unique<int>(42);
    auto smart_array = std::make_unique<int[]>(100);
    auto smart_string = std::make_unique<std::string>("hello");

    // Cleanup (required for raw pointers)
    delete heap_var;
    delete[] heap_array;
    delete heap_string;
    // Smart pointers clean up automatically
}
```

### RAII (Resource Acquisition Is Initialization)

```cpp
class ResourceManager {
private:
    void* resource_;
    size_t size_;

public:
    // Constructor acquires resource
    explicit ResourceManager(size_t size)
        : size_(size), resource_(malloc(size)) {
        if (!resource_) {
            throw std::bad_alloc();
        }
        std::cout << "Acquired resource of size " << size << std::endl;
    }

    // Destructor releases resource
    ~ResourceManager() {
        if (resource_) {
            free(resource_);
            std::cout << "Released resource of size " << size_ << std::endl;
        }
    }

    // Non-copyable (prevent double-free)
    ResourceManager(const ResourceManager&) = delete;
    ResourceManager& operator=(const ResourceManager&) = delete;

    // Movable (transfer ownership)
    ResourceManager(ResourceManager&& other) noexcept
        : resource_(other.resource_), size_(other.size_) {
        other.resource_ = nullptr;
        other.size_ = 0;
    }

    ResourceManager& operator=(ResourceManager&& other) noexcept {
        if (this != &other) {
            if (resource_) free(resource_);
            resource_ = other.resource_;
            size_ = other.size_;
            other.resource_ = nullptr;
            other.size_ = 0;
        }
        return *this;
    }

    void* get() const noexcept { return resource_; }
    size_t size() const noexcept { return size_; }
};
```

## Smart Pointers

### std::unique_ptr - Exclusive Ownership

```cpp
#include <memory>

class Widget {
public:
    Widget(int id) : id_(id) {
        std::cout << "Widget " << id_ << " created\n";
    }
    ~Widget() {
        std::cout << "Widget " << id_ << " destroyed\n";
    }
    void doSomething() const {
        std::cout << "Widget " << id_ << " doing something\n";
    }
private:
    int id_;
};

void demonstrate_unique_ptr() {
    // Creation
    auto widget1 = std::make_unique<Widget>(1);
    std::unique_ptr<Widget> widget2(new Widget(2)); // Less preferred

    // Usage
    widget1->doSomething();
    (*widget1).doSomething();

    // Transfer ownership
    auto widget3 = std::move(widget1); // widget1 is now nullptr

    // Custom deleter
    auto file_deleter = [](FILE* f) {
        if (f) {
            fclose(f);
            std::cout << "File closed\n";
        }
    };
    std::unique_ptr<FILE, decltype(file_deleter)> file(
        fopen("test.txt", "w"), file_deleter);

    // Array version
    auto array = std::make_unique<int[]>(10);
    array[0] = 42;
}

// Factory function returning unique_ptr
std::unique_ptr<Widget> createWidget(int id) {
    return std::make_unique<Widget>(id);
}
```

### std::shared_ptr - Shared Ownership

```cpp
#include <memory>
#include <vector>

void demonstrate_shared_ptr() {
    // Creation
    auto shared1 = std::make_shared<Widget>(1);
    std::shared_ptr<Widget> shared2(new Widget(2)); // Less efficient

    std::cout << "shared1 ref count: " << shared1.use_count() << std::endl;

    // Sharing ownership
    auto shared3 = shared1; // Copy increases ref count
    std::cout << "shared1 ref count: " << shared1.use_count() << std::endl;

    // Container of shared pointers
    std::vector<std::shared_ptr<Widget>> widgets;
    widgets.push_back(shared1);
    widgets.push_back(shared2);
    widgets.push_back(std::make_shared<Widget>(3));

    // Circular reference problem and solution
    class Parent; class Child;

    class Parent {
    public:
        std::vector<std::shared_ptr<Child>> children;
        void addChild(std::shared_ptr<Child> child) {
            children.push_back(child);
        }
    };

    class Child {
    public:
        std::weak_ptr<Parent> parent; // Use weak_ptr to break cycle
    };

    auto parent = std::make_shared<Parent>();
    auto child = std::make_shared<Child>();

    parent->addChild(child);
    child->parent = parent; // No circular reference due to weak_ptr
}
```

### std::weak_ptr - Non-owning Observer

```cpp
void demonstrate_weak_ptr() {
    std::shared_ptr<Widget> shared = std::make_shared<Widget>(1);
    std::weak_ptr<Widget> weak = shared;

    std::cout << "shared ref count: " << shared.use_count() << std::endl;
    std::cout << "weak expired: " << weak.expired() << std::endl;

    // Safe access through lock()
    if (auto locked = weak.lock()) {
        locked->doSomething(); // Safe to use
        std::cout << "ref count during lock: " << locked.use_count() << std::endl;
    }

    // Reset shared pointer
    shared.reset();
    std::cout << "weak expired after reset: " << weak.expired() << std::endl;

    // Attempting to lock expired weak_ptr returns nullptr
    if (auto locked = weak.lock()) {
        locked->doSomething(); // This won't execute
    } else {
        std::cout << "Cannot lock expired weak_ptr\n";
    }
}

// Cache implementation using weak_ptr
template<typename Key, typename Value>
class WeakCache {
private:
    std::unordered_map<Key, std::weak_ptr<Value>> cache_;

public:
    std::shared_ptr<Value> get(const Key& key) {
        auto it = cache_.find(key);
        if (it != cache_.end()) {
            if (auto locked = it->second.lock()) {
                return locked; // Return existing value
            } else {
                cache_.erase(it); // Clean up expired entry
            }
        }
        return nullptr; // Not found or expired
    }

    void put(const Key& key, std::shared_ptr<Value> value) {
        cache_[key] = value;
    }
};
```

## Custom Memory Allocators

### Pool Allocator

```cpp
#include <memory>
#include <vector>

template<typename T, size_t PoolSize = 1024>
class PoolAllocator {
private:
    alignas(T) char pool_[PoolSize * sizeof(T)];
    std::vector<bool> used_;
    size_t next_free_;

public:
    using value_type = T;

    PoolAllocator() : used_(PoolSize, false), next_free_(0) {}

    template<typename U>
    PoolAllocator(const PoolAllocator<U, PoolSize>&) noexcept {}

    T* allocate(size_t n) {
        if (n != 1) {
            throw std::bad_alloc(); // This simple allocator only handles single objects
        }

        // Find free slot
        for (size_t i = next_free_; i < PoolSize; ++i) {
            if (!used_[i]) {
                used_[i] = true;
                next_free_ = i + 1;
                return reinterpret_cast<T*>(&pool_[i * sizeof(T)]);
            }
        }

        // Wrap around
        for (size_t i = 0; i < next_free_; ++i) {
            if (!used_[i]) {
                used_[i] = true;
                next_free_ = i + 1;
                return reinterpret_cast<T*>(&pool_[i * sizeof(T)]);
            }
        }

        throw std::bad_alloc(); // Pool exhausted
    }

    void deallocate(T* p, size_t n) noexcept {
        if (n != 1 || !p) return;

        size_t index = (reinterpret_cast<char*>(p) - pool_) / sizeof(T);
        if (index < PoolSize) {
            used_[index] = false;
            next_free_ = std::min(next_free_, index);
        }
    }

    template<typename U>
    bool operator==(const PoolAllocator<U, PoolSize>&) const noexcept {
        return true; // All instances are equivalent
    }
};

// Usage with standard containers
using PoolVector = std::vector<int, PoolAllocator<int>>;
PoolVector vec; // Uses pool allocator
```

### Stack Allocator

```cpp
template<size_t StackSize>
class StackAllocator {
private:
    alignas(std::max_align_t) char stack_[StackSize];
    size_t offset_;

public:
    StackAllocator() : offset_(0) {}

    template<typename T>
    T* allocate(size_t n) {
        size_t required = n * sizeof(T);
        size_t alignment = alignof(T);

        // Align offset
        size_t aligned_offset = (offset_ + alignment - 1) & ~(alignment - 1);

        if (aligned_offset + required > StackSize) {
            throw std::bad_alloc();
        }

        offset_ = aligned_offset + required;
        return reinterpret_cast<T*>(&stack_[aligned_offset]);
    }

    template<typename T>
    void deallocate(T* p, size_t n) noexcept {
        // Stack allocator typically deallocates in reverse order
        // This is a simplified version
        size_t ptr_offset = reinterpret_cast<char*>(p) - stack_;
        if (ptr_offset + n * sizeof(T) == offset_) {
            offset_ = ptr_offset; // Only deallocate if it's the last allocation
        }
    }

    void reset() noexcept {
        offset_ = 0; // Reset entire stack
    }

    size_t bytes_used() const noexcept { return offset_; }
    size_t bytes_remaining() const noexcept { return StackSize - offset_; }
};
```

## Memory Debugging and Profiling

### Debug Allocator

```cpp
#include <unordered_map>
#include <mutex>

class DebugAllocator {
private:
    struct AllocInfo {
        size_t size;
        const char* file;
        int line;
        std::chrono::time_point<std::chrono::steady_clock> timestamp;
    };

    static std::unordered_map<void*, AllocInfo> allocations_;
    static std::mutex mutex_;
    static size_t total_allocated_;
    static size_t peak_allocated_;
    static size_t current_allocated_;

public:
    static void* allocate(size_t size, const char* file = __FILE__, int line = __LINE__) {
        void* ptr = std::malloc(size);
        if (!ptr) return nullptr;

        std::lock_guard<std::mutex> lock(mutex_);
        allocations_[ptr] = {size, file, line, std::chrono::steady_clock::now()};
        total_allocated_ += size;
        current_allocated_ += size;
        peak_allocated_ = std::max(peak_allocated_, current_allocated_);

        return ptr;
    }

    static void deallocate(void* ptr) noexcept {
        if (!ptr) return;

        std::lock_guard<std::mutex> lock(mutex_);
        auto it = allocations_.find(ptr);
        if (it != allocations_.end()) {
            current_allocated_ -= it->second.size;
            allocations_.erase(it);
        }
        std::free(ptr);
    }

    static void report_leaks() {
        std::lock_guard<std::mutex> lock(mutex_);
        if (!allocations_.empty()) {
            std::cout << "Memory leaks detected:\n";
            for (const auto& [ptr, info] : allocations_) {
                auto duration = std::chrono::steady_clock::now() - info.timestamp;
                auto seconds = std::chrono::duration_cast<std::chrono::seconds>(duration).count();

                std::cout << "Leak: " << info.size << " bytes at " << ptr
                         << " allocated in " << info.file << ":" << info.line
                         << " (alive for " << seconds << " seconds)\n";
            }
        }

        std::cout << "Memory Statistics:\n"
                 << "  Total allocated: " << total_allocated_ << " bytes\n"
                 << "  Peak usage: " << peak_allocated_ << " bytes\n"
                 << "  Current usage: " << current_allocated_ << " bytes\n";
    }
};

// Static member definitions
std::unordered_map<void*, DebugAllocator::AllocInfo> DebugAllocator::allocations_;
std::mutex DebugAllocator::mutex_;
size_t DebugAllocator::total_allocated_ = 0;
size_t DebugAllocator::peak_allocated_ = 0;
size_t DebugAllocator::current_allocated_ = 0;

#ifdef DEBUG_MEMORY
#define DEBUG_NEW(size) DebugAllocator::allocate(size, __FILE__, __LINE__)
#define DEBUG_DELETE(ptr) DebugAllocator::deallocate(ptr)
#else
#define DEBUG_NEW(size) std::malloc(size)
#define DEBUG_DELETE(ptr) std::free(ptr)
#endif
```

## Memory Safety Patterns

### RAII Wrappers for C Resources

```cpp
template<typename T, typename Deleter>
class RAIIWrapper {
private:
    T resource_;
    Deleter deleter_;

public:
    explicit RAIIWrapper(T resource, Deleter deleter)
        : resource_(resource), deleter_(deleter) {}

    ~RAIIWrapper() {
        if (resource_) {
            deleter_(resource_);
        }
    }

    // Non-copyable
    RAIIWrapper(const RAIIWrapper&) = delete;
    RAIIWrapper& operator=(const RAIIWrapper&) = delete;

    // Movable
    RAIIWrapper(RAIIWrapper&& other) noexcept
        : resource_(other.resource_), deleter_(std::move(other.deleter_)) {
        other.resource_ = T{};
    }

    RAIIWrapper& operator=(RAIIWrapper&& other) noexcept {
        if (this != &other) {
            if (resource_) deleter_(resource_);
            resource_ = other.resource_;
            deleter_ = std::move(other.deleter_);
            other.resource_ = T{};
        }
        return *this;
    }

    T get() const noexcept { return resource_; }
    T release() noexcept {
        T temp = resource_;
        resource_ = T{};
        return temp;
    }

    explicit operator bool() const noexcept {
        return resource_ != T{};
    }
};

// Helper function for creating RAII wrappers
template<typename T, typename Deleter>
auto make_raii(T resource, Deleter deleter) {
    return RAIIWrapper<T, Deleter>(resource, deleter);
}

// Usage examples
void demonstrate_raii_wrappers() {
    // File handle
    auto file = make_raii(fopen("test.txt", "r"), [](FILE* f) {
        if (f) fclose(f);
    });

    if (file) {
        // Use file.get() to access FILE*
        fread(buffer, 1, sizeof(buffer), file.get());
    }

    // Socket handle (hypothetical)
    auto socket = make_raii(create_socket(), [](int sock) {
        if (sock != -1) close_socket(sock);
    });
}
```

### Memory Pool Management

```cpp
class MemoryPool {
private:
    struct Block {
        Block* next;
    };

    void* pool_;
    Block* free_list_;
    size_t block_size_;
    size_t pool_size_;
    size_t blocks_allocated_;

public:
    MemoryPool(size_t block_size, size_t num_blocks)
        : block_size_(std::max(block_size, sizeof(Block)))
        , pool_size_(block_size_ * num_blocks)
        , blocks_allocated_(0) {

        pool_ = std::aligned_alloc(alignof(std::max_align_t), pool_size_);
        if (!pool_) {
            throw std::bad_alloc();
        }

        // Initialize free list
        char* current = static_cast<char*>(pool_);
        free_list_ = reinterpret_cast<Block*>(current);

        for (size_t i = 0; i < num_blocks - 1; ++i) {
            Block* block = reinterpret_cast<Block*>(current);
            current += block_size_;
            block->next = reinterpret_cast<Block*>(current);
        }

        // Last block points to nullptr
        reinterpret_cast<Block*>(current)->next = nullptr;
    }

    ~MemoryPool() {
        std::free(pool_);
    }

    void* allocate() {
        if (!free_list_) {
            return nullptr; // Pool exhausted
        }

        void* result = free_list_;
        free_list_ = free_list_->next;
        ++blocks_allocated_;
        return result;
    }

    void deallocate(void* ptr) noexcept {
        if (!ptr) return;

        Block* block = static_cast<Block*>(ptr);
        block->next = free_list_;
        free_list_ = block;
        --blocks_allocated_;
    }

    size_t blocks_allocated() const noexcept { return blocks_allocated_; }
    bool empty() const noexcept { return blocks_allocated_ == 0; }
};
```

## Common Memory Pitfalls and Solutions

### Double Delete Prevention

```cpp
template<typename T>
class SafePtr {
private:
    mutable T* ptr_;

public:
    explicit SafePtr(T* ptr = nullptr) : ptr_(ptr) {}

    ~SafePtr() {
        delete ptr_;
        ptr_ = nullptr;
    }

    // Non-copyable (prevents double delete)
    SafePtr(const SafePtr&) = delete;
    SafePtr& operator=(const SafePtr&) = delete;

    // Movable
    SafePtr(SafePtr&& other) noexcept : ptr_(other.ptr_) {
        other.ptr_ = nullptr;
    }

    SafePtr& operator=(SafePtr&& other) noexcept {
        if (this != &other) {
            delete ptr_;
            ptr_ = other.ptr_;
            other.ptr_ = nullptr;
        }
        return *this;
    }

    T* get() const noexcept { return ptr_; }
    T* release() const noexcept {
        T* temp = ptr_;
        ptr_ = nullptr;
        return temp;
    }

    T& operator*() const { return *ptr_; }
    T* operator->() const { return ptr_; }

    explicit operator bool() const noexcept { return ptr_ != nullptr; }
};
```

### Memory Alignment

```cpp
// Aligned allocation utilities
template<typename T>
T* aligned_new(size_t count, size_t alignment = alignof(T)) {
    void* ptr = std::aligned_alloc(alignment, count * sizeof(T));
    if (!ptr) throw std::bad_alloc();
    return static_cast<T*>(ptr);
}

template<typename T>
void aligned_delete(T* ptr) noexcept {
    std::free(ptr);
}

// SIMD-friendly aligned container
template<typename T, size_t Alignment = 32>
class AlignedVector {
private:
    T* data_;
    size_t size_;
    size_t capacity_;

public:
    AlignedVector() : data_(nullptr), size_(0), capacity_(0) {}

    explicit AlignedVector(size_t count)
        : data_(aligned_new<T>(count, Alignment))
        , size_(count), capacity_(count) {
        std::uninitialized_default_construct_n(data_, count);
    }

    ~AlignedVector() {
        if (data_) {
            std::destroy_n(data_, size_);
            aligned_delete(data_);
        }
    }

    void push_back(const T& value) {
        if (size_ >= capacity_) {
            reserve(capacity_ == 0 ? 1 : capacity_ * 2);
        }
        new(&data_[size_]) T(value);
        ++size_;
    }

    void reserve(size_t new_capacity) {
        if (new_capacity > capacity_) {
            T* new_data = aligned_new<T>(new_capacity, Alignment);

            if (data_) {
                std::uninitialized_move_n(data_, size_, new_data);
                std::destroy_n(data_, size_);
                aligned_delete(data_);
            }

            data_ = new_data;
            capacity_ = new_capacity;
        }
    }

    T* data() noexcept { return data_; }
    const T* data() const noexcept { return data_; }
    size_t size() const noexcept { return size_; }
    bool is_aligned() const noexcept {
        return reinterpret_cast<uintptr_t>(data_) % Alignment == 0;
    }
};
```

This memory management guide provides comprehensive coverage of C++ memory handling from basic concepts to advanced patterns, ensuring efficient and safe memory usage in production code.