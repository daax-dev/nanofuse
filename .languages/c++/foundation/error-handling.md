# C++ Error Handling Comprehensive Guide

Complete guide to error handling in C++, covering traditional approaches, exceptions, modern error handling patterns, and best practices.

## Error Handling Philosophies

### 1. Traditional C-Style Error Codes
**Approach**: Functions return error codes, actual results passed via parameters or global state.

```cpp
enum class ErrorCode {
    Success = 0,
    InvalidInput,
    OutOfMemory,
    FileNotFound,
    PermissionDenied
};

// Traditional approach
ErrorCode divide(double a, double b, double& result) {
    if (b == 0.0) {
        return ErrorCode::InvalidInput;
    }
    result = a / b;
    return ErrorCode::Success;
}

// Usage
double result;
ErrorCode error = divide(10.0, 2.0, result);
if (error != ErrorCode::Success) {
    // Handle error
    return error;
}
// Use result
```

**Advantages**:
- Zero overhead when no error occurs
- Explicit error handling at each call site
- Compatible with C code
- Deterministic performance

**Disadvantages**:
- Verbose error checking code
- Easy to ignore errors accidentally
- Difficult to propagate through call stack
- Mixed return value semantics

### 2. Exception-Based Error Handling
**Approach**: Use C++ exceptions for error conditions that should not be ignored.

```cpp
#include <stdexcept>
#include <string>

class DivisionByZeroError : public std::runtime_error {
public:
    explicit DivisionByZeroError(const std::string& message)
        : std::runtime_error("Division by zero: " + message) {}
};

double divide(double a, double b) {
    if (b == 0.0) {
        throw DivisionByZeroError("Cannot divide " + std::to_string(a) + " by zero");
    }
    return a / b;
}

// Usage
try {
    double result = divide(10.0, 0.0);
    std::cout << "Result: " << result << std::endl;
} catch (const DivisionByZeroError& e) {
    std::cerr << "Error: " << e.what() << std::endl;
} catch (const std::exception& e) {
    std::cerr << "Unexpected error: " << e.what() << std::endl;
}
```

**Advantages**:
- Cannot be ignored (program terminates if unhandled)
- Clean separation of normal and error code paths
- Automatic stack unwinding and cleanup
- Rich error information

**Disadvantages**:
- Runtime overhead (usually small)
- More complex control flow
- Not suitable for all environments (embedded, real-time)
- Can be disabled in some contexts

## Exception Safety Guarantees

### Exception Safety Levels

1. **No-Throw Guarantee**: Operations guaranteed not to throw
2. **Strong Guarantee**: Operation succeeds completely or has no effect
3. **Basic Guarantee**: No resources leaked, objects remain in valid state
4. **No Guarantee**: May leak resources or leave objects invalid

```cpp
class SafeContainer {
private:
    std::vector<int> data_;
    size_t size_;

public:
    // No-throw guarantee
    size_t size() const noexcept {
        return size_;
    }

    // Strong guarantee
    void push_back(int value) {
        data_.push_back(value);  // May throw, but if it does, nothing changed
        ++size_;                 // Only increment if push_back succeeded
    }

    // Strong guarantee with copy-and-swap
    SafeContainer& operator=(const SafeContainer& other) {
        SafeContainer temp(other);  // May throw, but doesn't affect *this
        swap(temp);                 // No-throw swap
        return *this;
    }

    void swap(SafeContainer& other) noexcept {
        data_.swap(other.data_);
        std::swap(size_, other.size_);
    }

    // Basic guarantee
    void process_all() {
        for (auto& item : data_) {
            try {
                risky_operation(item);  // May throw
            } catch (...) {
                // Log error but continue processing
                // Object remains in valid state
            }
        }
    }
};
```

### RAII and Exception Safety

```cpp
class FileHandler {
private:
    FILE* file_;

public:
    explicit FileHandler(const char* filename, const char* mode)
        : file_(fopen(filename, mode)) {
        if (!file_) {
            throw std::runtime_error("Cannot open file: " + std::string(filename));
        }
    }

    ~FileHandler() noexcept {
        if (file_) {
            fclose(file_);
        }
    }

    // Non-copyable but movable
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

    FILE* get() const noexcept { return file_; }
};

// Usage with automatic cleanup
void process_file(const char* filename) {
    try {
        FileHandler file(filename, "r");
        // File automatically closed when leaving scope
        // Even if an exception is thrown
        risky_file_operation(file.get());
    } catch (const std::exception& e) {
        std::cerr << "File processing failed: " << e.what() << std::endl;
        throw; // Re-throw if needed
    }
}
```

## Modern Error Handling Patterns

### 1. std::optional for Optional Values (C++17)

```cpp
#include <optional>

std::optional<int> safe_divide(int a, int b) {
    if (b == 0) {
        return std::nullopt;  // No value
    }
    return a / b;
}

// Usage
auto result = safe_divide(10, 2);
if (result.has_value()) {
    std::cout << "Result: " << result.value() << std::endl;
} else {
    std::cout << "Division failed" << std::endl;
}

// Alternative usage
if (auto result = safe_divide(10, 2)) {
    std::cout << "Result: " << *result << std::endl;
} else {
    std::cout << "Division failed" << std::endl;
}

// With value_or for default values
int result = safe_divide(10, 0).value_or(-1);
```

### 2. std::expected (C++23)

```cpp
#include <expected>

enum class MathError {
    DivisionByZero,
    Overflow,
    InvalidInput
};

std::expected<double, MathError> safe_divide(double a, double b) {
    if (b == 0.0) {
        return std::unexpected(MathError::DivisionByZero);
    }
    if (std::abs(a) > std::numeric_limits<double>::max() / std::abs(b)) {
        return std::unexpected(MathError::Overflow);
    }
    return a / b;
}

// Usage
auto result = safe_divide(10.0, 2.0);
if (result.has_value()) {
    std::cout << "Result: " << result.value() << std::endl;
} else {
    switch (result.error()) {
        case MathError::DivisionByZero:
            std::cout << "Error: Division by zero" << std::endl;
            break;
        case MathError::Overflow:
            std::cout << "Error: Overflow" << std::endl;
            break;
        case MathError::InvalidInput:
            std::cout << "Error: Invalid input" << std::endl;
            break;
    }
}

// Monadic operations
auto chain_operations = [](double x, double y, double z) {
    return safe_divide(x, y)
        .and_then([z](double intermediate) { return safe_divide(intermediate, z); })
        .transform([](double final_result) { return final_result * 2; });
};
```

### 3. Result Type Pattern (Custom Implementation)

```cpp
template<typename T, typename E>
class Result {
private:
    std::variant<T, E> data_;

public:
    Result(T value) : data_(std::move(value)) {}
    Result(E error) : data_(std::move(error)) {}

    bool is_ok() const {
        return std::holds_alternative<T>(data_);
    }

    bool is_error() const {
        return std::holds_alternative<E>(data_);
    }

    T& value() & {
        if (is_error()) {
            throw std::runtime_error("Attempted to get value from error result");
        }
        return std::get<T>(data_);
    }

    const T& value() const& {
        if (is_error()) {
            throw std::runtime_error("Attempted to get value from error result");
        }
        return std::get<T>(data_);
    }

    T&& value() && {
        if (is_error()) {
            throw std::runtime_error("Attempted to get value from error result");
        }
        return std::get<T>(std::move(data_));
    }

    E& error() & {
        if (is_ok()) {
            throw std::runtime_error("Attempted to get error from ok result");
        }
        return std::get<E>(data_);
    }

    const E& error() const& {
        if (is_ok()) {
            throw std::runtime_error("Attempted to get error from ok result");
        }
        return std::get<E>(data_);
    }

    template<typename F>
    auto map(F&& func) -> Result<decltype(func(std::declval<T>())), E> {
        if (is_ok()) {
            return func(value());
        } else {
            return error();
        }
    }

    template<typename F>
    auto and_then(F&& func) {
        if (is_ok()) {
            return func(value());
        } else {
            return Result<typename decltype(func(std::declval<T>()))::value_type, E>(error());
        }
    }
};

// Helper functions for construction
template<typename T, typename E>
Result<T, E> Ok(T value) {
    return Result<T, E>(std::move(value));
}

template<typename T, typename E>
Result<T, E> Err(E error) {
    return Result<T, E>(std::move(error));
}

// Usage
Result<int, std::string> divide(int a, int b) {
    if (b == 0) {
        return Err<int, std::string>("Division by zero");
    }
    return Ok<int, std::string>(a / b);
}
```

## Standard Exception Hierarchy

### Standard Exception Classes

```cpp
#include <stdexcept>
#include <system_error>

// Base exception class
std::exception
├── std::logic_error           // Programming logic errors
│   ├── std::invalid_argument  // Invalid argument passed
│   ├── std::domain_error      // Domain error in math function
│   ├── std::length_error      // Length error (e.g., string too long)
│   ├── std::out_of_range      // Index out of range
│   └── std::future_error      // Future/promise error
├── std::runtime_error         // Runtime errors
│   ├── std::range_error       // Range error in math function
│   ├── std::overflow_error    // Arithmetic overflow
│   ├── std::underflow_error   // Arithmetic underflow
│   └── std::system_error      // System error with error code
├── std::bad_alloc             // Memory allocation failure
├── std::bad_cast              // Bad dynamic_cast
├── std::bad_typeid            // Bad typeid usage
├── std::bad_exception         // Exception specification violation
└── std::bad_function_call     // Bad std::function call

// Usage examples
void demonstrate_standard_exceptions() {
    try {
        // Various operations that can throw
        std::vector<int> vec(5);
        int value = vec.at(10);  // Throws std::out_of_range
    } catch (const std::out_of_range& e) {
        std::cerr << "Index out of range: " << e.what() << std::endl;
    }

    try {
        std::string str("hello");
        str.resize(str.max_size() + 1);  // Throws std::length_error
    } catch (const std::length_error& e) {
        std::cerr << "Length error: " << e.what() << std::endl;
    }
}
```

### Custom Exception Classes

```cpp
// Domain-specific exception hierarchy
class FileSystemError : public std::runtime_error {
public:
    explicit FileSystemError(const std::string& message)
        : std::runtime_error("File system error: " + message) {}
};

class FileNotFoundError : public FileSystemError {
    std::string filename_;
public:
    explicit FileNotFoundError(const std::string& filename)
        : FileSystemError("File not found: " + filename)
        , filename_(filename) {}

    const std::string& filename() const noexcept {
        return filename_;
    }
};

class PermissionDeniedError : public FileSystemError {
    std::string operation_;
    std::string path_;
public:
    PermissionDeniedError(const std::string& operation, const std::string& path)
        : FileSystemError("Permission denied for " + operation + " on " + path)
        , operation_(operation), path_(path) {}

    const std::string& operation() const noexcept { return operation_; }
    const std::string& path() const noexcept { return path_; }
};

// Network-related exceptions
class NetworkError : public std::runtime_error {
protected:
    int error_code_;
public:
    explicit NetworkError(const std::string& message, int code = 0)
        : std::runtime_error("Network error: " + message), error_code_(code) {}

    int error_code() const noexcept { return error_code_; }
};

class ConnectionTimeoutError : public NetworkError {
    std::chrono::milliseconds timeout_;
public:
    explicit ConnectionTimeoutError(std::chrono::milliseconds timeout)
        : NetworkError("Connection timeout after " + std::to_string(timeout.count()) + "ms")
        , timeout_(timeout) {}

    std::chrono::milliseconds timeout() const noexcept { return timeout_; }
};
```

## Error Handling Best Practices

### 1. Exception-Safe Resource Management

```cpp
class Database {
private:
    void* connection_;

public:
    explicit Database(const std::string& connection_string) {
        connection_ = connect(connection_string.c_str());
        if (!connection_) {
            throw std::runtime_error("Failed to connect to database");
        }
    }

    ~Database() noexcept {
        if (connection_) {
            disconnect(connection_);
        }
    }

    // Non-copyable but movable
    Database(const Database&) = delete;
    Database& operator=(const Database&) = delete;

    Database(Database&& other) noexcept : connection_(other.connection_) {
        other.connection_ = nullptr;
    }

    Database& operator=(Database&& other) noexcept {
        if (this != &other) {
            if (connection_) disconnect(connection_);
            connection_ = other.connection_;
            other.connection_ = nullptr;
        }
        return *this;
    }

    // Strong exception safety
    std::vector<Record> query(const std::string& sql) {
        std::vector<Record> results;

        auto stmt = prepare_statement(connection_, sql.c_str());
        if (!stmt) {
            throw std::runtime_error("Failed to prepare statement");
        }

        // RAII guard for statement
        auto stmt_guard = make_scope_guard([stmt] { free_statement(stmt); });

        while (has_more_results(stmt)) {
            Record record = fetch_record(stmt);  // May throw
            results.push_back(std::move(record)); // Strong guarantee
        }

        return results;  // Move semantics
    }
};
```

### 2. Error Propagation Strategies

```cpp
// Strategy 1: Exception propagation
class ServiceLayer {
public:
    void process_request(const Request& request) {
        try {
            validate_request(request);  // May throw validation_error
            auto data = fetch_data(request.id());  // May throw database_error
            auto result = transform_data(data);     // May throw processing_error
            store_result(result);                   // May throw storage_error
        } catch (const validation_error& e) {
            // Log and re-throw with context
            log_error("Validation failed for request " + request.id(), e);
            throw processing_error("Request validation failed: " + std::string(e.what()));
        } catch (const database_error& e) {
            // Handle database errors specifically
            log_error("Database operation failed", e);
            if (e.is_transient()) {
                schedule_retry(request);
                throw transient_error("Database temporarily unavailable");
            } else {
                throw processing_error("Database operation failed: " + std::string(e.what()));
            }
        }
        // Other exceptions propagate unchanged
    }
};

// Strategy 2: Error code propagation with Result type
class ServiceLayerNoExcept {
public:
    Result<ProcessingResult, ProcessingError> process_request(const Request& request) noexcept {
        auto validation_result = validate_request(request);
        if (validation_result.is_error()) {
            return Err<ProcessingResult, ProcessingError>(
                ProcessingError::ValidationFailed(validation_result.error())
            );
        }

        auto fetch_result = fetch_data(request.id());
        if (fetch_result.is_error()) {
            if (fetch_result.error().is_transient()) {
                schedule_retry(request);
                return Err<ProcessingResult, ProcessingError>(
                    ProcessingError::TransientError("Database temporarily unavailable")
                );
            } else {
                return Err<ProcessingResult, ProcessingError>(
                    ProcessingError::DatabaseError(fetch_result.error())
                );
            }
        }

        return transform_data(fetch_result.value())
            .and_then([this](const auto& transformed) {
                return store_result(transformed);
            })
            .map_error([](const auto& error) {
                return ProcessingError::StorageError(error);
            });
    }
};
```

### 3. Logging and Diagnostics

```cpp
#include <source_location>

class Logger {
public:
    template<typename... Args>
    static void error(const std::string& message, Args&&... args,
                     const std::source_location& loc = std::source_location::current()) {
        log_with_location(LogLevel::Error, message, loc, std::forward<Args>(args)...);
    }

    template<typename... Args>
    static void warning(const std::string& message, Args&&... args,
                       const std::source_location& loc = std::source_location::current()) {
        log_with_location(LogLevel::Warning, message, loc, std::forward<Args>(args)...);
    }

private:
    template<typename... Args>
    static void log_with_location(LogLevel level, const std::string& message,
                                 const std::source_location& loc, Args&&... args) {
        std::ostringstream oss;
        oss << "[" << level << "] " << loc.file_name() << ":" << loc.line()
            << " in " << loc.function_name() << " - " << message;

        // Format additional arguments if provided
        if constexpr (sizeof...(args) > 0) {
            oss << " [";
            ((oss << args << " "), ...);
            oss << "]";
        }

        write_log(oss.str());
    }
};

// Exception with rich diagnostic information
class DiagnosticException : public std::runtime_error {
private:
    std::source_location location_;
    std::map<std::string, std::string> context_;

public:
    template<typename... Args>
    DiagnosticException(const std::string& message, Args&&... context_pairs,
                       const std::source_location& loc = std::source_location::current())
        : std::runtime_error(message), location_(loc) {
        add_context(std::forward<Args>(context_pairs)...);
    }

    const std::source_location& location() const noexcept {
        return location_;
    }

    const std::map<std::string, std::string>& context() const noexcept {
        return context_;
    }

    void add_context(const std::string& key, const std::string& value) {
        context_[key] = value;
    }

private:
    template<typename... Args>
    void add_context(const std::string& key, const std::string& value, Args&&... rest) {
        context_[key] = value;
        if constexpr (sizeof...(rest) > 0) {
            add_context(std::forward<Args>(rest)...);
        }
    }
};

// Usage
void risky_operation(const std::string& user_id, int retry_count) {
    try {
        perform_operation();
    } catch (const std::exception& e) {
        Logger::error("Operation failed", "user_id", user_id, "retry_count", retry_count);
        throw DiagnosticException("Operation failed for user",
                                 "user_id", user_id,
                                 "retry_count", std::to_string(retry_count),
                                 "original_error", e.what());
    }
}
```

### 4. Testing Error Conditions

```cpp
#include <gtest/gtest.h>

class ErrorHandlingTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Setup test fixtures
    }

    void TearDown() override {
        // Cleanup
    }
};

TEST_F(ErrorHandlingTest, DivisionByZeroThrowsException) {
    EXPECT_THROW(divide(10, 0), DivisionByZeroError);
}

TEST_F(ErrorHandlingTest, ExceptionMessageIsCorrect) {
    try {
        divide(10, 0);
        FAIL() << "Expected DivisionByZeroError to be thrown";
    } catch (const DivisionByZeroError& e) {
        EXPECT_THAT(e.what(), ::testing::HasSubstr("Division by zero"));
        EXPECT_THAT(e.what(), ::testing::HasSubstr("10"));
    }
}

TEST_F(ErrorHandlingTest, NoExceptOperationsAreNoThrow) {
    EXPECT_NO_THROW({
        SafeContainer container;
        auto size = container.size();  // Should be noexcept
        static_cast<void>(size);
    });
}

TEST_F(ErrorHandlingTest, ExceptionSafetyGuarantees) {
    SafeContainer container;
    container.push_back(1);
    container.push_back(2);

    auto original_size = container.size();

    // Simulate exception during operation
    EXPECT_THROW(container.risky_operation(), std::runtime_error);

    // Container should still be in valid state
    EXPECT_EQ(container.size(), original_size);
    EXPECT_NO_THROW(container.push_back(3));
}

// Property-based testing for error conditions
TEST_F(ErrorHandlingTest, SafeDividePropertiesHold) {
    for (int i = -100; i <= 100; ++i) {
        for (int j = -100; j <= 100; ++j) {
            auto result = safe_divide(i, j);

            if (j == 0) {
                EXPECT_FALSE(result.has_value()) << "Division by zero should return nullopt";
            } else {
                EXPECT_TRUE(result.has_value()) << "Valid division should return value";
                EXPECT_DOUBLE_EQ(*result, static_cast<double>(i) / j);
            }
        }
    }
}
```

This comprehensive error handling guide covers all major approaches to managing errors in C++, from traditional error codes to modern patterns like `std::expected`, with practical examples and best practices for building robust applications.