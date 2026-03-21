# C++ Keywords and Syntax Reference

Complete reference of all C++ keywords, operators, and syntax constructs across all standards from C++98 to C++23.

## Keywords by Category

### Access Control
```cpp
private     // Private access specifier
protected   // Protected access specifier
public      // Public access specifier
friend      // Friend function/class declaration
```

### Storage Classes and Linkage
```cpp
static      // Static storage duration or internal linkage
extern      // External linkage declaration
thread_local // Thread-local storage (C++11)
mutable     // Mutable class member
register    // Register storage hint (deprecated in C++17)
```

### Type Modifiers
```cpp
const       // Constant qualifier
volatile    // Volatile qualifier
constexpr   // Compile-time constant expression (C++11)
consteval   // Immediate function (C++20)
constinit   // Constant initialization (C++20)
```

### Fundamental Types
```cpp
// Integral types
bool        // Boolean type
char        // Character type
char8_t     // UTF-8 character type (C++20)
char16_t    // UTF-16 character type (C++11)
char32_t    // UTF-32 character type (C++11)
wchar_t     // Wide character type
short       // Short integer
int         // Integer
long        // Long integer
signed      // Signed integer modifier
unsigned    // Unsigned integer modifier

// Floating-point types
float       // Single precision floating-point
double      // Double precision floating-point

// Void type
void        // No type/return value
```

### Object-Oriented Programming
```cpp
class       // Class declaration
struct      // Structure declaration (public by default)
union       // Union declaration
enum        // Enumeration declaration
namespace   // Namespace declaration
using       // Using declaration/directive (C++11 enhanced)
typedef     // Type alias declaration
typename    // Template parameter/dependent name disambiguation
template    // Template declaration
explicit    // Explicit constructor/conversion
virtual     // Virtual function/inheritance
override    // Virtual function override (C++11)
final       // Final virtual function/class (C++11)
```

### Control Flow
```cpp
if          // Conditional statement
else        // Alternative conditional branch
switch      // Multi-way branch statement
case        // Switch case label
default     // Default case/constructor
for         // For loop
while       // While loop
do          // Do-while loop
break       // Break from loop/switch
continue    // Continue to next iteration
goto        // Jump to label
return      // Return from function
```

### Exception Handling
```cpp
try         // Exception handling block
catch       // Exception handler
throw       // Throw exception
noexcept    // No-exception specification (C++11)
```

### Memory Management
```cpp
new         // Dynamic memory allocation
delete      // Dynamic memory deallocation
```

### Operators
```cpp
operator    // Operator overloading
sizeof      // Size-of operator
alignof     // Alignment-of operator (C++11)
alignas     // Alignment specifier (C++11)
decltype    // Decltype specifier (C++11)
typeid      // Type identification
```

### C++11 and Later Keywords
```cpp
// C++11
auto        // Automatic type deduction
nullptr     // Null pointer constant
static_assert // Compile-time assertion

// C++20
concept     // Concept declaration
requires    // Requires clause
co_await    // Coroutine await
co_return   // Coroutine return
co_yield    // Coroutine yield

// C++23
(No new keywords in C++23, focus on library features)
```

## Operator Precedence and Associativity

### Precedence Table (Highest to Lowest)

| Precedence | Operator | Description | Associativity |
|------------|----------|-------------|---------------|
| 1 | `::` | Scope resolution | Left-to-right |
| 2 | `()` `[]` `->` `.` | Function call, subscript, member access | Left-to-right |
| 3 | `++` `--` | Postfix increment/decrement | Left-to-right |
| 4 | `++` `--` `+` `-` `!` `~` `*` `&` `sizeof` `new` `delete` | Prefix, unary, cast | Right-to-left |
| 5 | `.*` `->*` | Pointer-to-member | Left-to-right |
| 6 | `*` `/` `%` | Multiplicative | Left-to-right |
| 7 | `+` `-` | Additive | Left-to-right |
| 8 | `<<` `>>` | Bitwise shift | Left-to-right |
| 9 | `<=>` | Three-way comparison (C++20) | Left-to-right |
| 10 | `<` `<=` `>` `>=` | Relational | Left-to-right |
| 11 | `==` `!=` | Equality | Left-to-right |
| 12 | `&` | Bitwise AND | Left-to-right |
| 13 | `^` | Bitwise XOR | Left-to-right |
| 14 | `|` | Bitwise OR | Left-to-right |
| 15 | `&&` | Logical AND | Left-to-right |
| 16 | `||` | Logical OR | Left-to-right |
| 17 | `?:` | Ternary conditional | Right-to-left |
| 18 | `=` `+=` `-=` `*=` `/=` `%=` etc. | Assignment | Right-to-left |
| 19 | `throw` | Throw expression | Right-to-left |
| 20 | `,` | Comma | Left-to-right |

## Essential Syntax Constructs

### Class Declaration
```cpp
class MyClass {
private:
    int private_member;

protected:
    int protected_member;

public:
    // Constructors
    MyClass();                              // Default constructor
    MyClass(int value);                     // Parameterized constructor
    MyClass(const MyClass& other);          // Copy constructor
    MyClass(MyClass&& other) noexcept;      // Move constructor (C++11)

    // Destructor
    virtual ~MyClass();

    // Assignment operators
    MyClass& operator=(const MyClass& other);          // Copy assignment
    MyClass& operator=(MyClass&& other) noexcept;      // Move assignment (C++11)

    // Member functions
    void memberFunction() const;
    virtual void virtualFunction();
    virtual void pureVirtual() = 0;         // Pure virtual
    void finalFunction() final;             // Final function (C++11)

    // Static members
    static int staticMember;
    static void staticFunction();

    // Operator overloading
    MyClass operator+(const MyClass& other) const;
    bool operator==(const MyClass& other) const;
    friend std::ostream& operator<<(std::ostream& os, const MyClass& obj);
};
```

### Template Declaration
```cpp
// Function template
template<typename T>
T max(T a, T b) {
    return (a > b) ? a : b;
}

// Class template
template<typename T, int N>
class Array {
private:
    T data[N];

public:
    T& operator[](size_t index) { return data[index]; }
    constexpr size_t size() const { return N; }
};

// Variadic template (C++11)
template<typename... Args>
void print(Args... args) {
    ((std::cout << args << " "), ...); // Fold expression (C++17)
}

// Template specialization
template<>
class Array<bool, 8> {
    // Specialized implementation for bool
    uint8_t bits;
public:
    // Specialized interface
};

// Template alias (C++11)
template<typename T>
using Vec = std::vector<T>;
```

### Lambda Expressions (C++11)
```cpp
// Basic lambda
auto lambda = [](int x, int y) { return x + y; };

// Lambda with capture
int multiplier = 3;
auto multiply = [multiplier](int x) { return x * multiplier; };

// Capture by reference
int counter = 0;
auto increment = [&counter]() { ++counter; };

// Generic lambda (C++14)
auto generic = [](auto x, auto y) { return x + y; };

// Lambda with explicit return type
auto divide = [](double x, double y) -> double {
    return x / y;
};

// Immediately invoked lambda
int result = [](int x) { return x * x; }(5);
```

### Range-Based For Loop (C++11)
```cpp
std::vector<int> numbers = {1, 2, 3, 4, 5};

// Basic range-based for
for (int n : numbers) {
    std::cout << n << " ";
}

// With auto
for (auto n : numbers) {
    std::cout << n << " ";
}

// With reference (for modification)
for (auto& n : numbers) {
    n *= 2;
}

// With const reference (read-only, efficient)
for (const auto& item : container) {
    process(item);
}

// Structured bindings (C++17)
std::map<std::string, int> map;
for (const auto& [key, value] : map) {
    std::cout << key << ": " << value << "\n";
}
```

### Smart Pointers (C++11)
```cpp
#include <memory>

// Unique pointer
std::unique_ptr<int> ptr = std::make_unique<int>(42);
std::unique_ptr<int> moved_ptr = std::move(ptr); // Transfer ownership

// Shared pointer
std::shared_ptr<int> shared = std::make_shared<int>(42);
std::shared_ptr<int> copy = shared; // Shared ownership

// Weak pointer
std::weak_ptr<int> weak = shared; // Non-owning observer

// Custom deleter
auto custom_deleter = [](FILE* f) { if (f) fclose(f); };
std::unique_ptr<FILE, decltype(custom_deleter)> file(fopen("file.txt", "r"), custom_deleter);
```

### Initialization Forms
```cpp
// Direct initialization
int x(42);
std::string s("hello");

// Copy initialization
int y = 42;
std::string t = "hello";

// List initialization (C++11)
int z{42};
std::string u{"hello"};
std::vector<int> vec{1, 2, 3, 4, 5};

// Value initialization
int w{};        // Zero-initialized
int* ptr{};     // Null pointer

// Aggregate initialization
struct Point { int x, y; };
Point p{10, 20};

// Designated initializers (C++20)
Point q{.x = 10, .y = 20};
```

### Attributes (C++11+)
```cpp
// Standard attributes
[[noreturn]] void terminate();                    // Function never returns
[[deprecated]] void old_function();               // Deprecated function
[[deprecated("use new_function instead")]] void old_func();
[[nodiscard]] int important_function();           // Return value shouldn't be ignored
[[maybe_unused]] int debug_variable = 0;         // Variable might not be used
[[fallthrough]];                                  // Intentional switch fallthrough
[[likely]] if (condition) { }                    // Likely branch (C++20)
[[unlikely]] if (error) { }                      // Unlikely branch (C++20)

// Function attributes
[[nodiscard("check return code")]]
int process_data(const Data& data);

// Class attributes
class [[deprecated("use NewClass instead")]] OldClass {
    [[maybe_unused]] int unused_member;
};
```

### Concepts (C++20)
```cpp
#include <concepts>

// Define a concept
template<typename T>
concept Numeric = std::integral<T> || std::floating_point<T>;

// Use concept in function template
template<Numeric T>
T add(T a, T b) {
    return a + b;
}

// Alternative syntax
template<typename T>
requires Numeric<T>
T multiply(T a, T b) {
    return a * b;
}

// Complex concept
template<typename T>
concept Container = requires(T t) {
    t.begin();
    t.end();
    t.size();
};

// Concept with type requirements
template<typename T>
concept Printable = requires(T t) {
    std::cout << t;
};
```

### Coroutines (C++20)
```cpp
#include <coroutine>

// Simple generator
template<typename T>
struct generator {
    struct promise_type {
        T current_value;

        generator get_return_object() {
            return generator{std::coroutine_handle<promise_type>::from_promise(*this)};
        }

        std::suspend_always initial_suspend() { return {}; }
        std::suspend_always final_suspend() noexcept { return {}; }
        void unhandled_exception() {}

        std::suspend_always yield_value(T value) {
            current_value = value;
            return {};
        }

        void return_void() {}
    };

    std::coroutine_handle<promise_type> h_;

    bool move_next() {
        h_.resume();
        return !h_.done();
    }

    T current_value() {
        return h_.promise().current_value;
    }

    ~generator() { h_.destroy(); }
};

// Usage
generator<int> sequence() {
    for (int i = 0; i < 10; ++i) {
        co_yield i;
    }
}
```

### Modules (C++20)
```cpp
// Module interface unit (math.cppm)
export module math;

export int add(int a, int b) {
    return a + b;
}

export int multiply(int a, int b) {
    return a * b;
}

// Module implementation unit (math.cpp)
module math;
#include <iostream>

// Non-exported function
int internal_helper() {
    return 42;
}

// Using the module (main.cpp)
import math;

int main() {
    int result = add(5, 3);
    return 0;
}
```

### Three-Way Comparison (C++20)
```cpp
#include <compare>

class Point {
    int x, y;

public:
    Point(int x, int y) : x(x), y(y) {}

    // Default three-way comparison
    auto operator<=>(const Point& other) const = default;

    // Custom three-way comparison
    std::strong_ordering operator<=>(const Point& other) const {
        if (auto cmp = x <=> other.x; cmp != 0)
            return cmp;
        return y <=> other.y;
    }
};

// Usage
Point p1{1, 2};
Point p2{3, 4};

if (p1 < p2) { /* ... */ }
if (p1 == p2) { /* ... */ }
if (p1 >= p2) { /* ... */ }
```

## Preprocessor Directives

### Include and Macro Directives
```cpp
#include <header>       // System header
#include "header.h"     // User header
#define MACRO value     // Macro definition
#undef MACRO           // Macro undefinition
#ifdef MACRO           // Conditional compilation
#ifndef MACRO
#if condition
#elif condition
#else
#endif
```

### Pragma Directives
```cpp
#pragma once                    // Include guard
#pragma pack(push, 1)          // Structure packing
#pragma warning(disable: 4996) // Disable warning (MSVC)
#pragma GCC diagnostic ignored "-Wunused-variable" // GCC
```

This comprehensive syntax reference covers all major C++ language constructs from basic keywords to the latest C++20 and C++23 features, providing a complete foundation for C++ programming.