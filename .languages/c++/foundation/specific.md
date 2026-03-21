# C++-Specific Features and Advanced Language Constructs

## Overview

C++ provides unique features that distinguish it from other languages: RAII, templates, metaprogramming, concepts, modules, and other advanced language constructs. This guide covers these C++-specific features with practical examples.

## RAII (Resource Acquisition Is Initialization)

### Core RAII Principles

```cpp
#include <memory>
#include <fstream>
#include <mutex>

// Basic RAII class for file handling
class FileWrapper {
private:
    std::FILE* file;

public:
    explicit FileWrapper(const char* filename, const char* mode)
        : file(std::fopen(filename, mode)) {
        if (!file) {
            throw std::runtime_error("Failed to open file");
        }
    }

    // Disable copy construction and assignment
    FileWrapper(const FileWrapper&) = delete;
    FileWrapper& operator=(const FileWrapper&) = delete;

    // Enable move construction and assignment
    FileWrapper(FileWrapper&& other) noexcept : file(other.file) {
        other.file = nullptr;
    }

    FileWrapper& operator=(FileWrapper&& other) noexcept {
        if (this != &other) {
            if (file) std::fclose(file);
            file = other.file;
            other.file = nullptr;
        }
        return *this;
    }

    ~FileWrapper() {
        if (file) {
            std::fclose(file);
        }
    }

    std::FILE* get() const { return file; }

    bool write(const char* data) {
        return file && std::fputs(data, file) != EOF;
    }

    bool read(char* buffer, size_t size) {
        return file && std::fgets(buffer, size, file) != nullptr;
    }
};

// RAII for custom resources
template<typename Resource, typename Deleter>
class RAIIWrapper {
private:
    Resource resource;
    Deleter deleter;

public:
    template<typename... Args>
    RAIIWrapper(Deleter d, Args&&... args)
        : resource(std::forward<Args>(args)...), deleter(std::move(d)) {}

    ~RAIIWrapper() {
        deleter(resource);
    }

    RAIIWrapper(const RAIIWrapper&) = delete;
    RAIIWrapper& operator=(const RAIIWrapper&) = delete;

    RAIIWrapper(RAIIWrapper&& other) noexcept
        : resource(std::move(other.resource)), deleter(std::move(other.deleter)) {}

    RAIIWrapper& operator=(RAIIWrapper&& other) noexcept {
        if (this != &other) {
            deleter(resource);
            resource = std::move(other.resource);
            deleter = std::move(other.deleter);
        }
        return *this;
    }

    Resource& get() { return resource; }
    const Resource& get() const { return resource; }
};

// Usage examples
void raii_examples() {
    // Automatic file handling
    {
        FileWrapper file("test.txt", "w");
        file.write("Hello, RAII!");
    } // File automatically closed here

    // Custom resource management
    auto malloc_deleter = [](void* ptr) { std::free(ptr); };
    RAIIWrapper<void*, decltype(malloc_deleter)> memory(
        malloc_deleter, std::malloc(1024));

    // Smart pointer examples
    {
        auto ptr = std::make_unique<int>(42);
        auto shared = std::make_shared<std::string>("Hello");
        std::weak_ptr<std::string> weak = shared;
    } // Automatic cleanup

    // Lock guard example
    std::mutex mtx;
    {
        std::lock_guard<std::mutex> lock(mtx);
        // Critical section
    } // Mutex automatically unlocked
}
```

### Custom RAII Patterns

```cpp
// Scope guard implementation
template<typename Func>
class ScopeGuard {
private:
    Func func;
    bool active;

public:
    explicit ScopeGuard(Func f) : func(std::move(f)), active(true) {}

    ~ScopeGuard() {
        if (active) func();
    }

    void dismiss() { active = false; }

    ScopeGuard(const ScopeGuard&) = delete;
    ScopeGuard& operator=(const ScopeGuard&) = delete;

    ScopeGuard(ScopeGuard&& other) noexcept
        : func(std::move(other.func)), active(other.active) {
        other.active = false;
    }
};

template<typename Func>
auto make_scope_guard(Func&& func) {
    return ScopeGuard<std::decay_t<Func>>(std::forward<Func>(func));
}

#define SCOPE_EXIT(code) \
    auto UNIQUE_NAME(scope_guard) = make_scope_guard([&]() { code; })

void scope_guard_example() {
    std::FILE* file = std::fopen("test.txt", "w");
    SCOPE_EXIT(if (file) std::fclose(file));

    if (!file) return;

    // Use file...
    std::fputs("Hello", file);

    // File automatically closed on scope exit
}

// Resource pool with RAII
template<typename T>
class ResourcePool {
private:
    std::vector<std::unique_ptr<T>> pool;
    std::queue<T*> available;
    std::mutex mtx;

public:
    template<typename... Args>
    void add_resource(Args&&... args) {
        std::lock_guard<std::mutex> lock(mtx);
        auto resource = std::make_unique<T>(std::forward<Args>(args)...);
        available.push(resource.get());
        pool.push_back(std::move(resource));
    }

    class ResourceLease {
    private:
        T* resource;
        ResourcePool* pool;

    public:
        ResourceLease(T* res, ResourcePool* p) : resource(res), pool(p) {}

        ~ResourceLease() {
            if (resource && pool) {
                pool->return_resource(resource);
            }
        }

        ResourceLease(const ResourceLease&) = delete;
        ResourceLease& operator=(const ResourceLease&) = delete;

        ResourceLease(ResourceLease&& other) noexcept
            : resource(other.resource), pool(other.pool) {
            other.resource = nullptr;
            other.pool = nullptr;
        }

        T* get() const { return resource; }
        T& operator*() const { return *resource; }
        T* operator->() const { return resource; }
    };

    ResourceLease acquire() {
        std::lock_guard<std::mutex> lock(mtx);
        if (available.empty()) {
            throw std::runtime_error("No resources available");
        }

        T* resource = available.front();
        available.pop();
        return ResourceLease(resource, this);
    }

private:
    void return_resource(T* resource) {
        std::lock_guard<std::mutex> lock(mtx);
        available.push(resource);
    }

    friend class ResourceLease;
};
```

## Templates

### Function Templates

```cpp
#include <type_traits>
#include <concepts>

// Basic function template
template<typename T>
T max_value(const T& a, const T& b) {
    return (a > b) ? a : b;
}

// Template with multiple type parameters
template<typename T, typename U>
auto add(const T& a, const U& b) -> decltype(a + b) {
    return a + b;
}

// Template with non-type parameters
template<typename T, size_t N>
T sum_array(const T (&arr)[N]) {
    T result = T{};
    for (size_t i = 0; i < N; ++i) {
        result += arr[i];
    }
    return result;
}

// Variadic function template
template<typename T>
T sum(T&& t) {
    return std::forward<T>(t);
}

template<typename T, typename... Args>
T sum(T&& t, Args&&... args) {
    return std::forward<T>(t) + sum(std::forward<Args>(args)...);
}

// SFINAE (Substitution Failure Is Not An Error)
template<typename T>
typename std::enable_if_t<std::is_arithmetic_v<T>, T>
safe_divide(T a, T b) {
    return (b != T{}) ? a / b : T{};
}

// Tag dispatch
struct integral_tag {};
struct floating_point_tag {};

template<typename T>
constexpr auto get_category() {
    if constexpr (std::is_integral_v<T>) {
        return integral_tag{};
    } else {
        return floating_point_tag{};
    }
}

template<typename T>
T process_impl(T value, integral_tag) {
    return value * 2;  // Integer processing
}

template<typename T>
T process_impl(T value, floating_point_tag) {
    return value * 1.5;  // Floating-point processing
}

template<typename T>
T process(T value) {
    return process_impl(value, get_category<T>());
}

void function_template_examples() {
    // Basic usage
    int max_int = max_value(10, 20);
    double max_double = max_value(3.14, 2.71);

    // Auto deduction
    auto result = add(42, 3.14);  // int + double = double

    // Array template
    int arr[] = {1, 2, 3, 4, 5};
    int total = sum_array(arr);

    // Variadic template
    int variadic_sum = sum(1, 2, 3, 4, 5);

    // SFINAE
    double division = safe_divide(10.0, 3.0);

    // Tag dispatch
    int processed_int = process(42);      // Uses integral_tag
    double processed_double = process(3.14); // Uses floating_point_tag
}
```

### Class Templates

```cpp
// Basic class template
template<typename T>
class Vector3D {
private:
    T x, y, z;

public:
    Vector3D() : x(T{}), y(T{}), z(T{}) {}
    Vector3D(T x, T y, T z) : x(x), y(y), z(z) {}

    T dot(const Vector3D& other) const {
        return x * other.x + y * other.y + z * other.z;
    }

    Vector3D operator+(const Vector3D& other) const {
        return Vector3D(x + other.x, y + other.y, z + other.z);
    }

    T length() const {
        return std::sqrt(x*x + y*y + z*z);
    }

    // Template member function
    template<typename U>
    Vector3D<U> cast() const {
        return Vector3D<U>(static_cast<U>(x),
                          static_cast<U>(y),
                          static_cast<U>(z));
    }
};

// Template specialization
template<>
class Vector3D<bool> {
private:
    bool x, y, z;

public:
    Vector3D(bool x = false, bool y = false, bool z = false) : x(x), y(y), z(z) {}

    bool any() const { return x || y || z; }
    bool all() const { return x && y && z; }
};

// Partial template specialization
template<typename T>
class Matrix<T*> {
    // Specialized for pointer types
};

// Template template parameters
template<template<typename> class Container, typename T>
class Adapter {
private:
    Container<T> container;

public:
    void add(const T& item) {
        container.push_back(item);
    }

    size_t size() const {
        return container.size();
    }
};

// Advanced class template with policies
template<typename T, typename ComparePolicy = std::less<T>>
class SortedVector {
private:
    std::vector<T> data;
    ComparePolicy compare;

public:
    void insert(const T& value) {
        auto pos = std::lower_bound(data.begin(), data.end(), value, compare);
        data.insert(pos, value);
    }

    bool contains(const T& value) const {
        auto it = std::lower_bound(data.begin(), data.end(), value, compare);
        return it != data.end() && !compare(value, *it) && !compare(*it, value);
    }

    const std::vector<T>& get_data() const { return data; }
};

void class_template_examples() {
    // Basic usage
    Vector3D<float> vec1(1.0f, 2.0f, 3.0f);
    Vector3D<int> vec2(1, 2, 3);

    auto dot_product = vec1.dot(Vector3D<float>(2.0f, 3.0f, 4.0f));
    auto int_vec = vec1.template cast<int>();

    // Specialized template
    Vector3D<bool> bool_vec(true, false, true);
    bool has_any = bool_vec.any();

    // Template template parameter
    Adapter<std::vector, int> adapter;
    adapter.add(42);

    // Policy-based design
    SortedVector<int> sorted_ints;
    SortedVector<int, std::greater<int>> reverse_sorted;

    sorted_ints.insert(5);
    sorted_ints.insert(2);
    sorted_ints.insert(8);
}
```

### Template Metaprogramming

```cpp
#include <type_traits>

// Compile-time factorial
template<int N>
struct Factorial {
    static constexpr int value = N * Factorial<N-1>::value;
};

template<>
struct Factorial<0> {
    static constexpr int value = 1;
};

// Type list
template<typename...>
struct TypeList {};

// Type list operations
template<typename List>
struct Length;

template<typename... Types>
struct Length<TypeList<Types...>> {
    static constexpr size_t value = sizeof...(Types);
};

template<typename List, typename T>
struct Append;

template<typename... Types, typename T>
struct Append<TypeList<Types...>, T> {
    using type = TypeList<Types..., T>;
};

// Template metafunction for checking if type is in list
template<typename T, typename List>
struct Contains;

template<typename T>
struct Contains<T, TypeList<>> : std::false_type {};

template<typename T, typename Head, typename... Tail>
struct Contains<T, TypeList<Head, Tail...>>
    : std::conditional_t<std::is_same_v<T, Head>,
                        std::true_type,
                        Contains<T, TypeList<Tail...>>> {};

// CRTP (Curiously Recurring Template Pattern)
template<typename Derived>
class Printable {
public:
    void print() const {
        static_cast<const Derived*>(this)->print_impl();
    }
};

class Document : public Printable<Document> {
private:
    std::string content;

public:
    Document(std::string content) : content(std::move(content)) {}

    void print_impl() const {
        std::cout << "Document: " << content << std::endl;
    }
};

// Expression templates
template<typename E>
struct VecExpression {
    const E& self() const { return static_cast<const E&>(*this); }
    E& self() { return static_cast<E&>(*this); }

    operator E&() { return self(); }
    operator const E&() const { return self(); }
};

template<typename T>
struct Vec : VecExpression<Vec<T>> {
    std::vector<T> data;

    Vec(std::initializer_list<T> init) : data(init) {}
    Vec(size_t size) : data(size) {}

    T& operator[](size_t i) { return data[i]; }
    const T& operator[](size_t i) const { return data[i]; }

    size_t size() const { return data.size(); }

    template<typename E>
    Vec& operator=(const VecExpression<E>& expr) {
        const E& e = expr.self();
        for (size_t i = 0; i < size(); ++i) {
            data[i] = e[i];
        }
        return *this;
    }
};

template<typename LHS, typename RHS>
struct VecAdd : VecExpression<VecAdd<LHS, RHS>> {
    const LHS& lhs;
    const RHS& rhs;

    VecAdd(const LHS& l, const RHS& r) : lhs(l), rhs(r) {}

    auto operator[](size_t i) const { return lhs[i] + rhs[i]; }
    size_t size() const { return lhs.size(); }
};

template<typename LHS, typename RHS>
auto operator+(const VecExpression<LHS>& lhs, const VecExpression<RHS>& rhs) {
    return VecAdd<LHS, RHS>(lhs.self(), rhs.self());
}

void metaprogramming_examples() {
    // Compile-time computation
    constexpr int fact5 = Factorial<5>::value;  // 120

    // Type list operations
    using MyTypes = TypeList<int, double, std::string>;
    constexpr size_t length = Length<MyTypes>::value;  // 3

    using ExtendedTypes = Append<MyTypes, bool>::type;
    constexpr bool has_int = Contains<int, MyTypes>::value;  // true

    // CRTP
    Document doc("Hello, World!");
    doc.print();  // Calls Document::print_impl()

    // Expression templates (lazy evaluation)
    Vec<double> a{1.0, 2.0, 3.0};
    Vec<double> b{4.0, 5.0, 6.0};
    Vec<double> c{7.0, 8.0, 9.0};

    Vec<double> result(3);
    result = a + b + c;  // No temporaries created!
}
```

## Concepts (C++20)

### Basic Concepts

```cpp
#include <concepts>

// Built-in concepts
template<std::integral T>
void process_integer(T value) {
    std::cout << "Processing integer: " << value << std::endl;
}

template<std::floating_point T>
void process_float(T value) {
    std::cout << "Processing float: " << value << std::endl;
}

// Custom concepts
template<typename T>
concept Printable = requires(const T& t) {
    std::cout << t;
};

template<typename T>
concept Hashable = requires(const T& t) {
    { std::hash<T>{}(t) } -> std::convertible_to<size_t>;
};

template<typename T>
concept Container = requires(T t) {
    t.begin();
    t.end();
    t.size();
    typename T::value_type;
};

// Compound concepts
template<typename T>
concept PrintableHashable = Printable<T> && Hashable<T>;

// Function with concept constraints
template<Container T>
void print_container(const T& container) {
    for (const auto& item : container) {
        std::cout << item << " ";
    }
    std::cout << std::endl;
}

// Concept with multiple parameters
template<typename T, typename U>
concept Addable = requires(T a, U b) {
    { a + b } -> std::common_with<T>;
};

template<Addable<int> T>
T add_to_int(T value, int increment) {
    return value + increment;
}

// Advanced concepts
template<typename Iter>
concept RandomAccessIterator = std::random_access_iterator<Iter>;

template<typename T>
concept Arithmetic = std::integral<T> || std::floating_point<T>;

template<typename T>
concept SignedArithmetic = Arithmetic<T> && std::signed_integral<T>;

// Concept specialization
template<typename T>
void algorithm_impl(T data) requires std::forward_iterator<T> {
    std::cout << "Forward iterator algorithm" << std::endl;
}

template<typename T>
void algorithm_impl(T data) requires RandomAccessIterator<T> {
    std::cout << "Random access iterator algorithm" << std::endl;
}

void concepts_examples() {
    // Built-in concepts
    process_integer(42);      // Works with integral types
    process_float(3.14);      // Works with floating-point types

    // Custom concepts
    std::vector<int> vec{1, 2, 3, 4, 5};
    print_container(vec);     // Works with containers

    // Concept constraints
    auto result = add_to_int(5.5, 10);  // Works with types addable to int

    // Algorithm selection based on iterator category
    std::vector<int> v{1, 2, 3};
    algorithm_impl(v.begin());  // Selects random access version

    std::list<int> l{1, 2, 3};
    algorithm_impl(l.begin());  // Selects forward iterator version
}
```

### Advanced Concept Patterns

```cpp
// Concept composition
template<typename T>
concept Serializable = requires(const T& obj, std::ostream& os, std::istream& is) {
    { obj.serialize(os) } -> std::same_as<void>;
    { T::deserialize(is) } -> std::same_as<T>;
};

template<typename T>
concept Comparable = requires(const T& a, const T& b) {
    { a < b } -> std::convertible_to<bool>;
    { a > b } -> std::convertible_to<bool>;
    { a <= b } -> std::convertible_to<bool>;
    { a >= b } -> std::convertible_to<bool>;
    { a == b } -> std::convertible_to<bool>;
    { a != b } -> std::convertible_to<bool>;
};

// SFINAE replacement with concepts
template<typename T>
concept Incrementable = requires(T& t) {
    ++t;
    t++;
    { ++t } -> std::same_as<T&>;
};

template<Incrementable T>
T increment_copy(T value) {
    return ++value;
}

// Concepts with type constraints
template<typename T>
concept Numeric = requires {
    std::numeric_limits<T>::is_specialized;
} && (std::integral<T> || std::floating_point<T>);

template<Numeric T>
class Calculator {
public:
    static T add(T a, T b) { return a + b; }
    static T multiply(T a, T b) { return a * b; }
    static T max_value() { return std::numeric_limits<T>::max(); }
};

// Concept for checking member functions
template<typename T>
concept HasToString = requires(const T& t) {
    { t.to_string() } -> std::convertible_to<std::string>;
};

template<typename T>
std::string stringify(const T& obj) {
    if constexpr (HasToString<T>) {
        return obj.to_string();
    } else if constexpr (Printable<T>) {
        std::ostringstream oss;
        oss << obj;
        return oss.str();
    } else {
        return "Unstringifiable object";
    }
}
```

## Modules (C++20)

### Module Basics

```cpp
// math_utils.ixx (module interface)
export module math_utils;

import <iostream>;
import <cmath>;

// Exported declarations
export namespace math {
    double square(double x);
    double cube(double x);

    template<typename T>
    T power(T base, int exponent) {
        T result = T{1};
        for (int i = 0; i < exponent; ++i) {
            result *= base;
        }
        return result;
    }

    class Calculator {
    public:
        double add(double a, double b);
        double subtract(double a, double b);
    private:
        void log_operation(const std::string& op);
    };
}

// Internal (non-exported) helper
namespace {
    bool is_positive(double x) {
        return x > 0.0;
    }
}

// Implementation in same file or separate implementation unit
double math::square(double x) {
    return x * x;
}

double math::cube(double x) {
    return x * x * x;
}

double math::Calculator::add(double a, double b) {
    log_operation("addition");
    return a + b;
}

double math::Calculator::subtract(double a, double b) {
    log_operation("subtraction");
    return a - b;
}

void math::Calculator::log_operation(const std::string& op) {
    std::cout << "Performed " << op << std::endl;
}
```

### Module Partitions

```cpp
// geometry.ixx (primary module interface)
export module geometry;

export import :shapes;    // Import and re-export partition
export import :algorithms;

export namespace geometry {
    void print_info();
}

// geometry-shapes.ixx (interface partition)
export module geometry:shapes;

import <vector>;

export namespace geometry {
    class Point {
    public:
        Point(double x, double y) : x_(x), y_(y) {}
        double x() const { return x_; }
        double y() const { return y_; }
    private:
        double x_, y_;
    };

    class Circle {
    public:
        Circle(Point center, double radius) : center_(center), radius_(radius) {}
        Point center() const { return center_; }
        double radius() const { return radius_; }
        double area() const;
    private:
        Point center_;
        double radius_;
    };
}

// geometry-algorithms.ixx (interface partition)
export module geometry:algorithms;
import :shapes;

export namespace geometry {
    double distance(const Point& a, const Point& b);
    bool intersects(const Circle& a, const Circle& b);
}

// geometry-impl.cpp (implementation partition)
module geometry:shapes;
import <cmath>;

double geometry::Circle::area() const {
    return M_PI * radius_ * radius_;
}

module geometry:algorithms;

double geometry::distance(const Point& a, const Point& b) {
    double dx = a.x() - b.x();
    double dy = a.y() - b.y();
    return std::sqrt(dx*dx + dy*dy);
}

bool geometry::intersects(const Circle& a, const Circle& b) {
    double d = distance(a.center(), b.center());
    return d < (a.radius() + b.radius());
}
```

### Using Modules

```cpp
// main.cpp
import math_utils;
import geometry;
import <iostream>;

int main() {
    // Use exported functions from math_utils
    double sq = math::square(5.0);
    double cb = math::cube(3.0);

    // Use exported template
    int pow_result = math::power(2, 8);

    // Use exported class
    math::Calculator calc;
    double sum = calc.add(10.5, 20.3);

    // Use geometry module
    geometry::Point p1{0.0, 0.0};
    geometry::Point p2{3.0, 4.0};
    geometry::Circle circle{p1, 2.5};

    double dist = geometry::distance(p1, p2);
    double area = circle.area();

    std::cout << "Square: " << sq << std::endl;
    std::cout << "Cube: " << cb << std::endl;
    std::cout << "Power: " << pow_result << std::endl;
    std::cout << "Sum: " << sum << std::endl;
    std::cout << "Distance: " << dist << std::endl;
    std::cout << "Circle area: " << area << std::endl;

    return 0;
}
```

## Advanced Language Features

### Perfect Forwarding and Universal References

```cpp
#include <utility>
#include <type_traits>

// Perfect forwarding function
template<typename Func, typename... Args>
decltype(auto) invoke_function(Func&& func, Args&&... args) {
    return std::forward<Func>(func)(std::forward<Args>(args)...);
}

// Factory function with perfect forwarding
template<typename T, typename... Args>
std::unique_ptr<T> make_unique_forwarding(Args&&... args) {
    return std::unique_ptr<T>(new T(std::forward<Args>(args)...));
}

// Universal reference vs rvalue reference
template<typename T>
void process(T&& param) {  // Universal reference (forwarding reference)
    if constexpr (std::is_lvalue_reference_v<T>) {
        std::cout << "Lvalue reference" << std::endl;
    } else {
        std::cout << "Rvalue reference" << std::endl;
    }
}

void process(std::string&& param) {  // Rvalue reference
    std::cout << "Specific rvalue reference: " << param << std::endl;
}

// Perfect forwarding wrapper
template<typename Callable>
class FunctionWrapper {
private:
    Callable func;

public:
    template<typename F>
    FunctionWrapper(F&& f) : func(std::forward<F>(f)) {}

    template<typename... Args>
    decltype(auto) operator()(Args&&... args) {
        return func(std::forward<Args>(args)...);
    }
};

void perfect_forwarding_examples() {
    // Perfect forwarding in action
    auto lambda = [](const std::string& s, int x) {
        return s + std::to_string(x);
    };

    std::string str = "Value: ";
    auto result1 = invoke_function(lambda, str, 42);        // Lvalue
    auto result2 = invoke_function(lambda, std::string("Count: "), 10); // Rvalue

    // Factory with perfect forwarding
    auto ptr = make_unique_forwarding<std::string>(10, 'A');  // String of 10 'A's

    // Universal reference deduction
    int x = 42;
    process(x);           // Lvalue - T deduced as int&
    process(100);         // Rvalue - T deduced as int
    process(std::string("hello")); // Specific overload chosen

    // Wrapper usage
    FunctionWrapper wrapper(lambda);
    auto result3 = wrapper(str, 123);
}
```

### CRTP and Mixin Patterns

```cpp
// Static polymorphism with CRTP
template<typename Derived>
class Shape {
public:
    double area() const {
        return static_cast<const Derived*>(this)->area_impl();
    }

    void draw() const {
        static_cast<const Derived*>(this)->draw_impl();
    }

    // Common functionality
    void describe() const {
        std::cout << "Shape with area: " << area() << std::endl;
    }
};

class Rectangle : public Shape<Rectangle> {
private:
    double width, height;

public:
    Rectangle(double w, double h) : width(w), height(h) {}

    double area_impl() const {
        return width * height;
    }

    void draw_impl() const {
        std::cout << "Drawing rectangle " << width << "x" << height << std::endl;
    }
};

class Circle : public Shape<Circle> {
private:
    double radius;

public:
    Circle(double r) : radius(r) {}

    double area_impl() const {
        return M_PI * radius * radius;
    }

    void draw_impl() const {
        std::cout << "Drawing circle with radius " << radius << std::endl;
    }
};

// Mixin pattern
template<typename Base>
class Serializable : public Base {
public:
    template<typename... Args>
    Serializable(Args&&... args) : Base(std::forward<Args>(args)...) {}

    void serialize(std::ostream& os) const {
        os << "Serializing object of type: " << typeid(*this).name() << std::endl;
        // Custom serialization logic
    }

    void deserialize(std::istream& is) {
        // Custom deserialization logic
        std::cout << "Deserializing object" << std::endl;
    }
};

template<typename Base>
class Loggable : public Base {
public:
    template<typename... Args>
    Loggable(Args&&... args) : Base(std::forward<Args>(args)...) {}

    void log(const std::string& message) const {
        std::cout << "[LOG] " << message << std::endl;
    }
};

// Multiple mixins
using SerializableRectangle = Serializable<Rectangle>;
using LoggableSerializableRectangle = Loggable<Serializable<Rectangle>>;

void crtp_mixin_examples() {
    Rectangle rect(5.0, 3.0);
    Circle circle(2.0);

    // CRTP static polymorphism
    rect.describe();   // Calls Rectangle::area_impl()
    circle.describe(); // Calls Circle::area_impl()

    // Mixins
    SerializableRectangle ser_rect(4.0, 2.0);
    ser_rect.describe();
    ser_rect.serialize(std::cout);

    LoggableSerializableRectangle log_ser_rect(6.0, 4.0);
    log_ser_rect.log("Creating rectangle");
    log_ser_rect.describe();
    log_ser_rect.serialize(std::cout);
}
```

### Fold Expressions (C++17)

```cpp
// Unary fold expressions
template<typename... Args>
auto sum(Args... args) {
    return (args + ...);  // Unary right fold
}

template<typename... Args>
auto sum_left(Args... args) {
    return (... + args);  // Unary left fold
}

// Binary fold expressions
template<typename... Args>
void print_with_spaces(Args... args) {
    ((std::cout << args << " "), ...);  // Comma operator with fold
    std::cout << std::endl;
}

template<typename... Args>
bool all_positive(Args... args) {
    return ((args > 0) && ...);  // Logical AND fold
}

template<typename... Args>
bool any_negative(Args... args) {
    return ((args < 0) || ...);  // Logical OR fold
}

// Fold with function calls
template<typename... Funcs>
void call_all(Funcs... funcs) {
    (funcs(), ...);  // Call all functions
}

// Complex fold expression
template<typename Container, typename... Values>
void push_back_all(Container& container, Values... values) {
    (container.push_back(values), ...);
}

template<typename... Types>
size_t total_size() {
    return (sizeof(Types) + ...);
}

void fold_expression_examples() {
    // Basic arithmetic folds
    auto result1 = sum(1, 2, 3, 4, 5);        // 15
    auto result2 = sum_left(1, 2, 3, 4, 5);   // 15

    // Printing with fold
    print_with_spaces("Hello", "world", "from", "fold", "expressions");

    // Logical operations
    bool all_pos = all_positive(1, 2, 3, 4);     // true
    bool any_neg = any_negative(1, -2, 3, 4);    // true

    // Function call fold
    auto f1 = []() { std::cout << "Function 1 "; };
    auto f2 = []() { std::cout << "Function 2 "; };
    auto f3 = []() { std::cout << "Function 3 "; };
    call_all(f1, f2, f3);
    std::cout << std::endl;

    // Container operations
    std::vector<int> vec;
    push_back_all(vec, 1, 2, 3, 4, 5);

    // Type size calculation
    constexpr size_t total = total_size<int, double, char, bool>();
    std::cout << "Total size: " << total << " bytes" << std::endl;
}
```

### Constexpr and Compile-Time Programming

```cpp
#include <array>
#include <string_view>

// Constexpr functions
constexpr int factorial(int n) {
    if (n <= 1) return 1;
    return n * factorial(n - 1);
}

constexpr bool is_prime(int n) {
    if (n < 2) return false;
    for (int i = 2; i * i <= n; ++i) {
        if (n % i == 0) return false;
    }
    return true;
}

// Constexpr class
class constexpr_string {
private:
    const char* data_;
    size_t size_;

public:
    constexpr constexpr_string(const char* str) : data_(str), size_(0) {
        while (str[size_]) ++size_;  // Compute length
    }

    constexpr size_t size() const { return size_; }
    constexpr const char* data() const { return data_; }
    constexpr char operator[](size_t index) const { return data_[index]; }

    constexpr bool starts_with(const constexpr_string& prefix) const {
        if (prefix.size() > size_) return false;
        for (size_t i = 0; i < prefix.size(); ++i) {
            if (data_[i] != prefix.data_[i]) return false;
        }
        return true;
    }
};

// Constexpr algorithms
template<typename T, size_t N>
constexpr std::array<T, N> generate_primes() {
    std::array<T, N> primes{};
    size_t count = 0;
    int candidate = 2;

    while (count < N) {
        if (is_prime(candidate)) {
            primes[count++] = candidate;
        }
        ++candidate;
    }

    return primes;
}

// Consteval functions (C++20) - must be evaluated at compile time
consteval int compile_time_only(int x) {
    return x * x;
}

// constinit (C++20) - initialization must be at compile time
constinit int global_value = compile_time_only(10);

// Template metaprogramming with constexpr
template<int N>
struct FibonacciArray {
    static constexpr std::array<int, N> generate() {
        std::array<int, N> arr{};
        if (N > 0) arr[0] = 0;
        if (N > 1) arr[1] = 1;

        for (int i = 2; i < N; ++i) {
            arr[i] = arr[i-1] + arr[i-2];
        }

        return arr;
    }

    static constexpr auto value = generate();
};

void constexpr_examples() {
    // Compile-time computations
    constexpr int fact5 = factorial(5);           // 120
    constexpr bool prime_check = is_prime(17);    // true

    // Compile-time string operations
    constexpr constexpr_string str("Hello, World!");
    constexpr size_t len = str.size();            // 13
    constexpr bool starts = str.starts_with("Hello"); // true

    // Compile-time array generation
    constexpr auto first_10_primes = generate_primes<int, 10>();

    // Consteval
    constexpr int ct_value = compile_time_only(5); // 25

    // Template metaprogramming
    constexpr auto fib_10 = FibonacciArray<10>::value;

    // Print results
    std::cout << "Factorial of 5: " << fact5 << std::endl;
    std::cout << "Is 17 prime: " << prime_check << std::endl;
    std::cout << "String length: " << len << std::endl;
    std::cout << "First 10 primes: ";
    for (auto p : first_10_primes) {
        std::cout << p << " ";
    }
    std::cout << std::endl;
}
```

## Best Practices and Patterns

### Modern C++ Idioms

```cpp
// Rule of Five/Zero
class Resource {
private:
    std::unique_ptr<int[]> data;
    size_t size;

public:
    // Constructor
    explicit Resource(size_t s) : data(std::make_unique<int[]>(s)), size(s) {}

    // Rule of Five
    ~Resource() = default;  // Compiler-generated is fine with unique_ptr

    Resource(const Resource& other) : data(std::make_unique<int[]>(other.size)), size(other.size) {
        std::copy(other.data.get(), other.data.get() + size, data.get());
    }

    Resource& operator=(const Resource& other) {
        if (this != &other) {
            data = std::make_unique<int[]>(other.size);
            size = other.size;
            std::copy(other.data.get(), other.data.get() + size, data.get());
        }
        return *this;
    }

    Resource(Resource&&) = default;
    Resource& operator=(Resource&&) = default;
};

// PIMPL idiom
class Widget {
private:
    class Impl;
    std::unique_ptr<Impl> pimpl;

public:
    Widget();
    ~Widget();  // Must be defined in .cpp file

    Widget(const Widget& other);
    Widget& operator=(const Widget& other);
    Widget(Widget&&) noexcept;
    Widget& operator=(Widget&&) noexcept;

    void do_something();
    int get_value() const;
};

// Type erasure
class Drawable {
private:
    struct DrawableConcept {
        virtual ~DrawableConcept() = default;
        virtual void draw() const = 0;
        virtual std::unique_ptr<DrawableConcept> clone() const = 0;
    };

    template<typename T>
    struct DrawableModel : DrawableConcept {
        T object;

        explicit DrawableModel(T obj) : object(std::move(obj)) {}

        void draw() const override {
            object.draw();
        }

        std::unique_ptr<DrawableConcept> clone() const override {
            return std::make_unique<DrawableModel>(*this);
        }
    };

    std::unique_ptr<DrawableConcept> concept_;

public:
    template<typename T>
    Drawable(T obj) : concept_(std::make_unique<DrawableModel<T>>(std::move(obj))) {}

    Drawable(const Drawable& other) : concept_(other.concept_->clone()) {}
    Drawable& operator=(const Drawable& other) {
        if (this != &other) {
            concept_ = other.concept_->clone();
        }
        return *this;
    }

    Drawable(Drawable&&) = default;
    Drawable& operator=(Drawable&&) = default;

    void draw() const {
        concept_->draw();
    }
};

// Policy-based design
template<typename OutputPolicy, typename FilterPolicy>
class Logger : private OutputPolicy, private FilterPolicy {
public:
    template<typename... Args>
    void log(const std::string& level, Args&&... args) {
        if (FilterPolicy::should_log(level)) {
            OutputPolicy::output(level, std::forward<Args>(args)...);
        }
    }
};

struct ConsoleOutput {
    template<typename... Args>
    void output(const std::string& level, Args&&... args) {
        std::cout << "[" << level << "] ";
        ((std::cout << args << " "), ...);
        std::cout << std::endl;
    }
};

struct DebugFilter {
    bool should_log(const std::string& level) {
        return level == "DEBUG" || level == "ERROR";
    }
};

using DebugLogger = Logger<ConsoleOutput, DebugFilter>;
```

This comprehensive guide covers C++-specific features from RAII and templates to modern language constructs like concepts and modules, providing practical examples for production use.