# C Error Handling Comprehensive Guide

## C's Error Handling Philosophy

### Core Principles
1. **No Built-in Exceptions**: C has no try/catch mechanism
2. **Explicit Error Checking**: Errors must be explicitly checked by the programmer
3. **Return Code Convention**: Functions typically return status codes
4. **errno Global Variable**: Standard library sets errno for system call errors
5. **Fail-Fast or Graceful Degradation**: Choose error handling strategy based on context

### Error Handling Strategies
- **Return Status Codes**: Most common pattern
- **Output Parameters**: Return data through pointers, status through return value
- **Global Error State**: errno mechanism
- **Assertion-based**: Debug-time error detection
- **Setjmp/Longjmp**: Exception-like mechanism (use carefully)

## Return Code Error Handling

### Basic Return Code Patterns

#### Success/Failure Boolean
```c
#include <stdbool.h>

// Simple boolean success/failure
bool initialize_system(void) {
    if (/* initialization fails */) {
        return false;
    }
    return true;
}

// Usage
if (!initialize_system()) {
    fprintf(stderr, "System initialization failed\n");
    exit(EXIT_FAILURE);
}
```

#### Integer Status Codes
```c
// Standard status code pattern
#define SUCCESS 0
#define ERROR_INVALID_INPUT -1
#define ERROR_OUT_OF_MEMORY -2
#define ERROR_FILE_NOT_FOUND -3

int process_file(const char* filename) {
    if (!filename) {
        return ERROR_INVALID_INPUT;
    }

    FILE* file = fopen(filename, "r");
    if (!file) {
        return ERROR_FILE_NOT_FOUND;
    }

    // Process file...

    fclose(file);
    return SUCCESS;
}

// Usage with detailed error handling
int result = process_file("data.txt");
switch (result) {
    case SUCCESS:
        printf("File processed successfully\n");
        break;
    case ERROR_INVALID_INPUT:
        fprintf(stderr, "Invalid input parameter\n");
        break;
    case ERROR_FILE_NOT_FOUND:
        fprintf(stderr, "File not found\n");
        break;
    default:
        fprintf(stderr, "Unknown error: %d\n", result);
        break;
}
```

### Output Parameters Pattern
```c
// Return status, output through parameters
int divide_safely(double a, double b, double* result) {
    if (!result) {
        return -1;  // Invalid output parameter
    }

    if (b == 0.0) {
        return -2;  // Division by zero
    }

    *result = a / b;
    return 0;  // Success
}

// Usage
double result;
int status = divide_safely(10.0, 3.0, &result);
if (status == 0) {
    printf("Result: %f\n", result);
} else {
    fprintf(stderr, "Division failed with code %d\n", status);
}
```

### Multiple Return Values Using Structures
```c
// Result structure pattern
typedef struct {
    int status;
    char* data;
    size_t length;
} parse_result_t;

parse_result_t parse_input(const char* input) {
    parse_result_t result = {0};

    if (!input) {
        result.status = -1;
        return result;
    }

    // Parsing logic...
    result.data = malloc(strlen(input) + 1);
    if (!result.data) {
        result.status = -2;  // Out of memory
        return result;
    }

    strcpy(result.data, input);
    result.length = strlen(input);
    result.status = 0;  // Success

    return result;
}

// Usage
parse_result_t result = parse_input("Hello, World!");
if (result.status == 0) {
    printf("Parsed: %s (length: %zu)\n", result.data, result.length);
    free(result.data);
} else {
    fprintf(stderr, "Parse error: %d\n", result.status);
}
```

## errno-based Error Handling

### Understanding errno
```c
#include <errno.h>
#include <string.h>

// errno is a global variable set by system calls and library functions
extern int errno;

void demonstrate_errno(void) {
    FILE* file;

    // Clear errno before operation
    errno = 0;

    file = fopen("nonexistent_file.txt", "r");
    if (file == NULL) {
        // Check specific error codes
        if (errno == ENOENT) {
            printf("File does not exist\n");
        } else if (errno == EACCES) {
            printf("Permission denied\n");
        } else {
            // Generic error message
            printf("Error opening file: %s\n", strerror(errno));
        }
        return;
    }

    fclose(file);
}
```

### Thread-Safe errno Handling
```c
#include <errno.h>
#include <pthread.h>

// In multi-threaded programs, errno is thread-local
void* thread_function(void* arg) {
    FILE* file;
    int local_errno;

    file = fopen("test.txt", "r");
    if (file == NULL) {
        local_errno = errno;  // Save errno immediately

        // Handle error using saved errno
        fprintf(stderr, "Thread error: %s\n", strerror(local_errno));
        return NULL;
    }

    // Process file...
    fclose(file);
    return NULL;
}
```

### Common errno Values
```c
#include <errno.h>

void print_common_errors(void) {
    printf("ENOENT (No such file): %d\n", ENOENT);
    printf("EACCES (Permission denied): %d\n", EACCES);
    printf("ENOMEM (Out of memory): %d\n", ENOMEM);
    printf("EINVAL (Invalid argument): %d\n", EINVAL);
    printf("EIO (I/O error): %d\n", EIO);
    printf("EAGAIN (Resource temporarily unavailable): %d\n", EAGAIN);
    printf("EBUSY (Device or resource busy): %d\n", EBUSY);
    printf("EEXIST (File exists): %d\n", EEXIST);
    printf("ENOTDIR (Not a directory): %d\n", ENOTDIR);
    printf("EISDIR (Is a directory): %d\n", EISDIR);
}
```

## Memory Management Error Handling

### Safe Memory Allocation
```c
#include <stdlib.h>
#include <stdio.h>

void* safe_malloc(size_t size) {
    void* ptr = malloc(size);
    if (!ptr) {
        fprintf(stderr, "Memory allocation failed for size %zu\n", size);
        exit(EXIT_FAILURE);
    }
    return ptr;
}

void* safe_calloc(size_t count, size_t size) {
    void* ptr = calloc(count, size);
    if (!ptr) {
        fprintf(stderr, "Memory allocation failed for %zu items of size %zu\n",
                count, size);
        exit(EXIT_FAILURE);
    }
    return ptr;
}

void* safe_realloc(void* ptr, size_t size) {
    void* new_ptr = realloc(ptr, size);
    if (!new_ptr && size > 0) {
        fprintf(stderr, "Memory reallocation failed for size %zu\n", size);
        // Don't free original pointer here - caller should handle
        exit(EXIT_FAILURE);
    }
    return new_ptr;
}

// Usage with error checking
int* create_array(size_t count) {
    int* array = malloc(count * sizeof(int));
    if (!array) {
        return NULL;  // Let caller handle the error
    }

    // Initialize array...
    for (size_t i = 0; i < count; i++) {
        array[i] = 0;
    }

    return array;
}

// Caller handles allocation failure
int* array = create_array(1000);
if (!array) {
    fprintf(stderr, "Failed to create array\n");
    return EXIT_FAILURE;
}
```

### Resource Management with Cleanup
```c
#include <stdio.h>
#include <stdlib.h>

typedef struct {
    FILE* file;
    char* buffer;
    int fd;
} resource_t;

// Resource management with proper cleanup
int process_file_safe(const char* filename) {
    resource_t res = {0};  // Initialize all fields to NULL/0
    int result = -1;

    // Allocate resources
    res.file = fopen(filename, "r");
    if (!res.file) {
        goto cleanup;  // Jump to cleanup section
    }

    res.buffer = malloc(1024);
    if (!res.buffer) {
        goto cleanup;
    }

    // Process file...
    if (/* processing succeeds */) {
        result = 0;  // Success
    }

cleanup:
    // Clean up resources in reverse order
    if (res.buffer) {
        free(res.buffer);
    }
    if (res.file) {
        fclose(res.file);
    }

    return result;
}
```

## Assertion-Based Error Handling

### Standard assert Macro
```c
#include <assert.h>
#include <stdio.h>

void process_array(int* array, size_t size) {
    // Precondition checks
    assert(array != NULL);
    assert(size > 0);

    for (size_t i = 0; i < size; i++) {
        assert(i < size);  // Bounds check
        array[i] *= 2;
    }
}

// Custom assertion macro with better messages
#define ASSERT_MSG(condition, message) \
    do { \
        if (!(condition)) { \
            fprintf(stderr, "Assertion failed: %s at %s:%d in %s()\n", \
                    message, __FILE__, __LINE__, __func__); \
            abort(); \
        } \
    } while(0)

void safe_function(void* ptr) {
    ASSERT_MSG(ptr != NULL, "Pointer cannot be NULL");
    // Function implementation...
}
```

### Custom Debug Assertions
```c
#include <stdio.h>
#include <stdlib.h>

// Debug-only assertions
#ifdef DEBUG
    #define DEBUG_ASSERT(condition) \
        do { \
            if (!(condition)) { \
                fprintf(stderr, "DEBUG: Assertion failed: %s at %s:%d\n", \
                        #condition, __FILE__, __LINE__); \
                abort(); \
            } \
        } while(0)
#else
    #define DEBUG_ASSERT(condition) ((void)0)
#endif

void example_function(int value) {
    DEBUG_ASSERT(value >= 0);  // Only checked in debug builds

    // Function implementation...
}
```

## setjmp/longjmp Exception-like Handling

### Basic setjmp/longjmp Usage
```c
#include <setjmp.h>
#include <stdio.h>

jmp_buf jump_buffer;

void function_that_might_fail(void) {
    // Simulate an error condition
    if (/* error condition */) {
        longjmp(jump_buffer, 1);  // Jump back with error code 1
    }

    printf("Function executed successfully\n");
}

int main(void) {
    int result = setjmp(jump_buffer);

    if (result == 0) {
        // First time through - normal execution
        printf("Calling function...\n");
        function_that_might_fail();
        printf("Function completed normally\n");
    } else {
        // Returned from longjmp - error handling
        printf("Error occurred with code: %d\n", result);
        return 1;
    }

    return 0;
}
```

### Exception-like Error Handling Framework
```c
#include <setjmp.h>
#include <stdio.h>
#include <stdbool.h>

#define MAX_EXCEPTION_STACK 10

typedef struct {
    jmp_buf buffer;
    int exception_code;
    char* message;
} exception_frame_t;

static exception_frame_t exception_stack[MAX_EXCEPTION_STACK];
static int exception_stack_top = -1;

// Exception handling macros
#define TRY \
    do { \
        exception_stack_top++; \
        if (setjmp(exception_stack[exception_stack_top].buffer) == 0) {

#define CATCH(code) \
        } else if (exception_stack[exception_stack_top].exception_code == (code)) {

#define CATCH_ANY \
        } else {

#define FINALLY \
        } \
        exception_stack_top--; \
    } while(0)

#define THROW(code, msg) \
    do { \
        if (exception_stack_top >= 0) { \
            exception_stack[exception_stack_top].exception_code = (code); \
            exception_stack[exception_stack_top].message = (msg); \
            longjmp(exception_stack[exception_stack_top].buffer, 1); \
        } else { \
            fprintf(stderr, "Uncaught exception: %d - %s\n", (code), (msg)); \
            exit(EXIT_FAILURE); \
        } \
    } while(0)

// Usage example
#define ERROR_INVALID_INPUT 1
#define ERROR_OUT_OF_MEMORY 2

void risky_function(int value) {
    if (value < 0) {
        THROW(ERROR_INVALID_INPUT, "Value cannot be negative");
    }

    char* buffer = malloc(1000);
    if (!buffer) {
        THROW(ERROR_OUT_OF_MEMORY, "Failed to allocate memory");
    }

    // Normal processing...
    free(buffer);
}

int main(void) {
    TRY {
        risky_function(-5);
        printf("Function completed successfully\n");
    }
    CATCH(ERROR_INVALID_INPUT) {
        printf("Input error: %s\n", exception_stack[exception_stack_top].message);
    }
    CATCH(ERROR_OUT_OF_MEMORY) {
        printf("Memory error: %s\n", exception_stack[exception_stack_top].message);
    }
    CATCH_ANY {
        printf("Unknown error: %d - %s\n",
               exception_stack[exception_stack_top].exception_code,
               exception_stack[exception_stack_top].message);
    }
    FINALLY;

    return 0;
}
```

## Defensive Programming Techniques

### Input Validation
```c
#include <stdio.h>
#include <stdbool.h>
#include <limits.h>
#include <ctype.h>

// Validate string input
bool is_valid_string(const char* str, size_t max_length) {
    if (!str) return false;

    size_t length = strlen(str);
    if (length == 0 || length > max_length) return false;

    // Check for valid characters
    for (size_t i = 0; i < length; i++) {
        if (!isprint(str[i]) && !isspace(str[i])) {
            return false;
        }
    }

    return true;
}

// Safe integer parsing
int parse_integer(const char* str, int* result) {
    if (!str || !result) {
        return -1;  // Invalid parameters
    }

    char* endptr;
    errno = 0;

    long value = strtol(str, &endptr, 10);

    // Check for conversion errors
    if (errno == ERANGE) {
        return -2;  // Overflow/underflow
    }

    if (endptr == str || *endptr != '\0') {
        return -3;  // Invalid format
    }

    // Check if value fits in int
    if (value < INT_MIN || value > INT_MAX) {
        return -4;  // Value out of range for int
    }

    *result = (int)value;
    return 0;  // Success
}

// Safe array access
bool safe_array_get(const int* array, size_t array_size, size_t index, int* value) {
    if (!array || !value) {
        return false;
    }

    if (index >= array_size) {
        return false;  // Index out of bounds
    }

    *value = array[index];
    return true;
}
```

### Buffer Overflow Prevention
```c
#include <string.h>
#include <stdio.h>

// Safe string copying
int safe_strcpy(char* dest, size_t dest_size, const char* src) {
    if (!dest || !src || dest_size == 0) {
        return -1;
    }

    size_t src_len = strlen(src);
    if (src_len >= dest_size) {
        return -2;  // Source too large
    }

    strcpy(dest, src);  // Safe because we checked the size
    return 0;
}

// Safe string concatenation
int safe_strcat(char* dest, size_t dest_size, const char* src) {
    if (!dest || !src || dest_size == 0) {
        return -1;
    }

    size_t dest_len = strlen(dest);
    size_t src_len = strlen(src);

    if (dest_len + src_len >= dest_size) {
        return -2;  // Result would be too large
    }

    strcat(dest, src);
    return 0;
}

// Safe formatted string creation
int safe_sprintf(char* buffer, size_t buffer_size, const char* format, ...) {
    if (!buffer || !format || buffer_size == 0) {
        return -1;
    }

    va_list args;
    va_start(args, format);

    int result = vsnprintf(buffer, buffer_size, format, args);

    va_end(args);

    if (result < 0 || (size_t)result >= buffer_size) {
        return -2;  // Formatting error or truncation
    }

    return 0;  // Success
}
```

## Error Logging and Reporting

### Simple Logging Framework
```c
#include <stdio.h>
#include <stdarg.h>
#include <time.h>
#include <string.h>

typedef enum {
    LOG_DEBUG,
    LOG_INFO,
    LOG_WARNING,
    LOG_ERROR,
    LOG_FATAL
} log_level_t;

static const char* log_level_strings[] = {
    "DEBUG", "INFO", "WARNING", "ERROR", "FATAL"
};

static log_level_t current_log_level = LOG_INFO;
static FILE* log_file = NULL;

void log_init(const char* filename, log_level_t level) {
    current_log_level = level;

    if (filename) {
        log_file = fopen(filename, "a");
        if (!log_file) {
            fprintf(stderr, "Warning: Could not open log file %s\n", filename);
            log_file = stderr;
        }
    } else {
        log_file = stderr;
    }
}

void log_message(log_level_t level, const char* format, ...) {
    if (level < current_log_level || !log_file) {
        return;
    }

    // Get current time
    time_t now;
    time(&now);
    struct tm* local_time = localtime(&now);

    char time_str[64];
    strftime(time_str, sizeof(time_str), "%Y-%m-%d %H:%M:%S", local_time);

    // Print timestamp and log level
    fprintf(log_file, "[%s] %s: ", time_str, log_level_strings[level]);

    // Print the actual message
    va_list args;
    va_start(args, format);
    vfprintf(log_file, format, args);
    va_end(args);

    fprintf(log_file, "\n");
    fflush(log_file);
}

void log_cleanup(void) {
    if (log_file && log_file != stderr) {
        fclose(log_file);
    }
}

// Convenience macros
#define LOG_DEBUG(...) log_message(LOG_DEBUG, __VA_ARGS__)
#define LOG_INFO(...) log_message(LOG_INFO, __VA_ARGS__)
#define LOG_WARNING(...) log_message(LOG_WARNING, __VA_ARGS__)
#define LOG_ERROR(...) log_message(LOG_ERROR, __VA_ARGS__)
#define LOG_FATAL(...) log_message(LOG_FATAL, __VA_ARGS__)

// Usage example
void example_with_logging(void) {
    log_init("application.log", LOG_DEBUG);

    LOG_INFO("Application started");

    FILE* file = fopen("data.txt", "r");
    if (!file) {
        LOG_ERROR("Failed to open data.txt: %s", strerror(errno));
        return;
    }

    LOG_DEBUG("File opened successfully");

    // Process file...

    fclose(file);
    LOG_INFO("Processing completed");

    log_cleanup();
}
```

## Best Practices for C Error Handling

### 1. Consistent Error Codes
```c
// Define error codes in a header file
#ifndef ERRORS_H
#define ERRORS_H

typedef enum {
    ERR_SUCCESS = 0,
    ERR_INVALID_ARGUMENT = -1,
    ERR_OUT_OF_MEMORY = -2,
    ERR_FILE_NOT_FOUND = -3,
    ERR_PERMISSION_DENIED = -4,
    ERR_TIMEOUT = -5,
    ERR_NETWORK_ERROR = -6,
    ERR_UNKNOWN = -999
} error_code_t;

const char* error_to_string(error_code_t error);

#endif /* ERRORS_H */
```

### 2. Error Context Information
```c
// Error context structure
typedef struct {
    error_code_t code;
    char message[256];
    char file[64];
    int line;
    char function[64];
} error_context_t;

#define SET_ERROR(ctx, err_code, msg) \
    do { \
        (ctx)->code = (err_code); \
        snprintf((ctx)->message, sizeof((ctx)->message), "%s", (msg)); \
        snprintf((ctx)->file, sizeof((ctx)->file), "%s", __FILE__); \
        (ctx)->line = __LINE__; \
        snprintf((ctx)->function, sizeof((ctx)->function), "%s", __func__); \
    } while(0)
```

### 3. Graceful Degradation
```c
// Fallback mechanisms
int load_configuration(const char* filename, config_t* config) {
    FILE* file = fopen(filename, "r");
    if (!file) {
        // Use default configuration
        LOG_WARNING("Could not load config file %s, using defaults", filename);
        load_default_config(config);
        return 1;  // Warning: using defaults
    }

    // Load from file...
    fclose(file);
    return 0;  // Success
}
```

### 4. Error Recovery
```c
// Retry mechanism
int network_operation_with_retry(const char* url, int max_retries) {
    int attempts = 0;
    int result;

    do {
        result = network_operation(url);
        if (result == 0) {
            return 0;  // Success
        }

        attempts++;
        if (attempts < max_retries) {
            LOG_WARNING("Network operation failed, retrying... (%d/%d)",
                       attempts, max_retries);
            sleep(1);  // Wait before retry
        }
    } while (attempts < max_retries);

    LOG_ERROR("Network operation failed after %d attempts", max_retries);
    return result;
}
```

C error handling requires discipline and consistency. The key is choosing the appropriate strategy for each situation and applying it consistently throughout your codebase.