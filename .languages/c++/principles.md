# C++ Principles, Idioms, and Philosophy

Core principles, design patterns, and philosophical approaches that define effective C++ programming and guide best practices.

## Core C++ Philosophy

### 1. Zero-Cost Abstraction
**Principle**: You shouldn't pay for what you don't use, and what you do use should be as efficient as hand-written code.

```cpp
// High-level abstraction with zero runtime cost
std::vector<int> numbers = {1, 2, 3, 4, 5};
auto sum = std::accumulate(numbers.begin(), numbers.end(), 0);
// Compiles to optimal assembly equivalent to manual loop
```

**Implications**:
- Templates enable compile-time optimization
- RAII provides safety without runtime overhead
- Inline functions eliminate call overhead
- Compiler optimizations are crucial

### 2. Direct Resource Management
**Principle**: Provide direct access to hardware and system resources when needed.

```cpp
// Direct memory management when required
void* raw_memory = std::aligned_alloc(64, size);
// Memory-mapped file access
void* mapped = mmap(nullptr, file_size, PROT_READ, MAP_PRIVATE, fd, 0);
```

**Implications**:
- Programmers can control resource usage precisely
- No hidden allocations or garbage collection overhead
- Manual memory management when automatic isn't sufficient

### 3. Type Safety
**Principle**: Catch errors at compile time rather than runtime when possible.

```cpp
// Strong typing prevents errors
enum class Color { Red, Green, Blue };
enum class Size { Small, Medium, Large };

// Color color = Size::Small;  // Compilation error
Color color = Color::Red;      // Correct
```

**Implications**:
- Strong type system prevents many bugs
- Templates enable type-safe generic programming
- Concepts (C++20) improve template error messages

### 4. Performance by Default
**Principle**: The language should enable writing efficient code by default.

```cpp
// Move semantics avoid unnecessary copies
std::vector<std::string> create_strings() {
    std::vector<std::string> result;
    result.emplace_back("Hello");  // Constructed in-place
    return result;  // Move, not copy
}
```

## Fundamental Design Principles

### RAII (Resource Acquisition Is Initialization)
**Core Idiom**: Tie resource lifetime to object lifetime.

```cpp
class FileHandler {
private:
    FILE* file_;
public:
    explicit FileHandler(const char* filename)
        : file_(fopen(filename, "r")) {
        if (!file_) throw std::runtime_error("Cannot open file");
    }

    ~FileHandler() {
        if (file_) fclose(file_);
    }

    // Non-copyable, movable
    FileHandler(const FileHandler&) = delete;
    FileHandler& operator=(const FileHandler&) = delete;

    FileHandler(FileHandler&& other) noexcept : file_(other.file_) {
        other.file_ = nullptr;
    }

    FileHandler& operator=(FileHandler&& other) noexcept {
        if (this != &other) {
            if (file_) fclose(file_);
            file_ = other.file_;
            other.file_ = nullptr;
        }
        return *this;
    }
};
```

### Rule of Zero/Three/Five
**Principle**: Manage special member functions consistently.

```cpp
// Rule of Zero: No manual resource management
class Simple {
    std::string name_;
    std::vector<int> data_;
    // Compiler-generated destructor, copy/move constructors/operators work correctly
};

// Rule of Five: Manual resource management
class ResourceManager {
    Resource* resource_;
public:
    ~ResourceManager();                                    // Destructor
    ResourceManager(const ResourceManager& other);        // Copy constructor
    ResourceManager& operator=(const ResourceManager&);   // Copy assignment
    ResourceManager(ResourceManager&& other) noexcept;    // Move constructor
    ResourceManager& operator=(ResourceManager&&) noexcept; // Move assignment
};
```

### Value Semantics
**Principle**: Objects should behave like values (copyable, comparable, assignable).

```cpp
class Point {
    double x_, y_;
public:
    Point(double x, double y) : x_(x), y_(y) {}

    // Value-like operations
    Point operator+(const Point& other) const {
        return Point(x_ + other.x_, y_ + other.y_);
    }

    bool operator==(const Point& other) const {
        return x_ == other.x_ && y_ == other.y_;
    }

    bool operator<(const Point& other) const {
        return std::tie(x_, y_) < std::tie(other.x_, other.y_);
    }
};
```

## Essential C++ Idioms

### 1. Pimpl (Pointer to Implementation)
**Purpose**: Hide implementation details, reduce compilation dependencies.

```cpp
// Widget.h
class Widget {
public:
    Widget();
    ~Widget();
    Widget(const Widget& other);
    Widget& operator=(const Widget& other);
    Widget(Widget&& other) noexcept;
    Widget& operator=(Widget&& other) noexcept;

    void do_something();

private:
    class Impl;
    std::unique_ptr<Impl> pimpl_;
};

// Widget.cpp
class Widget::Impl {
public:
    void do_something() { /* implementation */ }
    // Private data members
    ComplexType complex_data_;
    AnotherComplexType more_data_;
};

Widget::Widget() : pimpl_(std::make_unique<Impl>()) {}
Widget::~Widget() = default;  // Important: defined where Impl is complete
```

### 2. CRTP (Curiously Recurring Template Pattern)
**Purpose**: Static polymorphism, interface enforcement.

```cpp
template<typename Derived>
class Drawable {
public:
    void draw() const {
        static_cast<const Derived*>(this)->draw_impl();
    }

    void render() const {
        prepare();
        draw();
        cleanup();
    }

private:
    void prepare() const { /* common preparation */ }
    void cleanup() const { /* common cleanup */ }
};

class Circle : public Drawable<Circle> {
public:
    void draw_impl() const {
        // Circle-specific drawing
    }
};

class Rectangle : public Drawable<Rectangle> {
public:
    void draw_impl() const {
        // Rectangle-specific drawing
    }
};
```

### 3. Tag Dispatching
**Purpose**: Select algorithm implementation based on type properties.

```cpp
namespace detail {
    template<typename Iterator>
    void advance_impl(Iterator& it, int n, std::random_access_iterator_tag) {
        it += n;  // O(1) for random access iterators
    }

    template<typename Iterator>
    void advance_impl(Iterator& it, int n, std::input_iterator_tag) {
        while (n--) ++it;  // O(n) for input iterators
    }
}

template<typename Iterator>
void advance(Iterator& it, int n) {
    detail::advance_impl(it, n,
        typename std::iterator_traits<Iterator>::iterator_category{});
}
```

### 4. SFINAE (Substitution Failure Is Not An Error)
**Purpose**: Enable/disable template specializations based on type properties.

```cpp
// C++11/14 style
template<typename T>
typename std::enable_if_t<std::is_integral_v<T>, T>
safe_divide(T a, T b) {
    if (b == 0) throw std::invalid_argument("Division by zero");
    return a / b;
}

template<typename T>
typename std::enable_if_t<std::is_floating_point_v<T>, T>
safe_divide(T a, T b) {
    if (std::abs(b) < std::numeric_limits<T>::epsilon())
        throw std::invalid_argument("Division by near-zero");
    return a / b;
}

// C++20 style with concepts
template<std::integral T>
T safe_divide(T a, T b) {
    if (b == 0) throw std::invalid_argument("Division by zero");
    return a / b;
}

template<std::floating_point T>
T safe_divide(T a, T b) {
    if (std::abs(b) < std::numeric_limits<T>::epsilon())
        throw std::invalid_argument("Division by near-zero");
    return a / b;
}
```

### 5. Type Erasure
**Purpose**: Hide type information while maintaining value semantics.

```cpp
class Drawable {
public:
    template<typename T>
    Drawable(T object) : object_(std::make_unique<Model<T>>(std::move(object))) {}

    void draw() const { object_->draw(); }

    Drawable(const Drawable& other) : object_(other.object_->clone()) {}
    Drawable& operator=(const Drawable& other) {
        if (this != &other) {
            object_ = other.object_->clone();
        }
        return *this;
    }

    Drawable(Drawable&&) = default;
    Drawable& operator=(Drawable&&) = default;

private:
    struct Concept {
        virtual ~Concept() = default;
        virtual void draw() const = 0;
        virtual std::unique_ptr<Concept> clone() const = 0;
    };

    template<typename T>
    struct Model : Concept {
        Model(T object) : object_(std::move(object)) {}
        void draw() const override { object_.draw(); }
        std::unique_ptr<Concept> clone() const override {
            return std::make_unique<Model>(*this);
        }
        T object_;
    };

    std::unique_ptr<Concept> object_;
};
```

## Modern C++ Principles (C++11 and Later)

### 1. Prefer Auto
**Principle**: Let the compiler deduce types when it improves readability or correctness.

```cpp
// Clear improvement
auto it = container.begin();  // Instead of std::vector<T>::iterator it
auto lambda = [](int x) { return x * 2; };

// Avoid hidden conversions
auto value = some_function();  // Gets exact return type
```

### 2. Use Range-Based For Loops
**Principle**: Prefer range-based loops over traditional index-based loops.

```cpp
// Modern approach
for (const auto& element : container) {
    process(element);
}

// When you need index
for (const auto& [index, element] : std::views::enumerate(container)) {
    process(index, element);
}
```

### 3. Prefer Smart Pointers
**Principle**: Use smart pointers for automatic memory management.

```cpp
// Ownership transfer
std::unique_ptr<Resource> create_resource() {
    return std::make_unique<Resource>(params);
}

// Shared ownership
std::shared_ptr<Resource> shared_resource = std::make_shared<Resource>(params);

// Observer pattern
std::weak_ptr<Resource> observer = shared_resource;
```

### 4. Use Uniform Initialization
**Principle**: Prefer brace initialization for consistency and safety.

```cpp
// Uniform initialization
std::vector<int> numbers{1, 2, 3, 4, 5};
MyClass object{param1, param2};
int value{42};  // Prevents narrowing conversions

// Most vexing parse avoidance
MyClass object{};  // Default construction, not function declaration
```

### 5. Move Semantics
**Principle**: Prefer moving over copying when the source object won't be used.

```cpp
class Resource {
    std::vector<int> data_;
public:
    // Move constructor
    Resource(Resource&& other) noexcept : data_(std::move(other.data_)) {}

    // Move assignment
    Resource& operator=(Resource&& other) noexcept {
        if (this != &other) {
            data_ = std::move(other.data_);
        }
        return *this;
    }
};

// Usage
Resource resource = create_resource();  // Move, not copy
container.push_back(std::move(resource));  // Explicit move
```

## Performance-Oriented Principles

### 1. Minimize Dynamic Allocation
**Principle**: Use stack allocation and avoid `new`/`delete` when possible.

```cpp
// Prefer stack allocation
void process_data() {
    std::array<int, 1000> buffer;  // Stack allocation
    // Use buffer...
}

// Use containers for dynamic size
void process_variable_data(size_t size) {
    std::vector<int> buffer(size);  // Single allocation
    // Use buffer...
}
```

### 2. Prefer Algorithms Over Hand-Written Loops
**Principle**: STL algorithms are optimized and less error-prone.

```cpp
// Instead of manual loop
std::vector<int> numbers = get_numbers();
std::sort(numbers.begin(), numbers.end());

// Use algorithm compositions
auto result = numbers
    | std::views::filter([](int x) { return x > 0; })
    | std::views::transform([](int x) { return x * 2; })
    | std::views::take(10);
```

### 3. Minimize Copying
**Principle**: Use references, move semantics, and in-place construction.

```cpp
// Pass by const reference for large objects
void process(const std::vector<Data>& data);

// Return by value (move optimization)
std::vector<Data> create_data();

// Emplace instead of push
container.emplace_back(constructor_args);  // Construct in-place
```

## Error Handling Philosophy

### 1. Fail Fast
**Principle**: Detect and report errors as early as possible.

```cpp
class BankAccount {
    double balance_;
public:
    void withdraw(double amount) {
        if (amount < 0) {
            throw std::invalid_argument("Negative withdrawal amount");
        }
        if (amount > balance_) {
            throw std::runtime_error("Insufficient funds");
        }
        balance_ -= amount;
    }
};
```

### 2. Exception Safety
**Principle**: Provide appropriate exception safety guarantees.

```cpp
class Vector {
    T* data_;
    size_t size_, capacity_;
public:
    void push_back(const T& value) {
        if (size_ == capacity_) {
            // Strong exception safety: either succeeds completely or fails with no effect
            size_t new_capacity = capacity_ * 2;
            T* new_data = new T[new_capacity];

            try {
                for (size_t i = 0; i < size_; ++i) {
                    new_data[i] = data_[i];  // May throw
                }
                new_data[size_] = value;  // May throw
            } catch (...) {
                delete[] new_data;
                throw;  // Rethrow, original state unchanged
            }

            delete[] data_;
            data_ = new_data;
            capacity_ = new_capacity;
        } else {
            data_[size_] = value;  // May throw, but strong safety maintained
        }
        ++size_;
    }
};
```

## Code Organization Principles

### 1. Single Responsibility Principle
**Principle**: Each class/function should have one reason to change.

```cpp
// Bad: Multiple responsibilities
class FileProcessor {
    void read_file(const std::string& filename);
    void process_data();
    void write_file(const std::string& filename);
    void log_operations();
    void send_email_notification();
};

// Good: Separated responsibilities
class FileReader {
    std::string read_file(const std::string& filename);
};

class DataProcessor {
    ProcessedData process(const std::string& data);
};

class FileWriter {
    void write_file(const std::string& filename, const ProcessedData& data);
};
```

### 2. Dependency Inversion
**Principle**: Depend on abstractions, not concretions.

```cpp
// Abstract interface
class Logger {
public:
    virtual ~Logger() = default;
    virtual void log(const std::string& message) = 0;
};

// Concrete implementations
class FileLogger : public Logger {
public:
    void log(const std::string& message) override;
};

class ConsoleLogger : public Logger {
public:
    void log(const std::string& message) override;
};

// Client depends on abstraction
class Application {
    std::unique_ptr<Logger> logger_;
public:
    explicit Application(std::unique_ptr<Logger> logger)
        : logger_(std::move(logger)) {}

    void run() {
        logger_->log("Application started");
        // ...
    }
};
```

These principles and idioms form the foundation of effective C++ programming, emphasizing performance, safety, and maintainability while leveraging the language's unique strengths.