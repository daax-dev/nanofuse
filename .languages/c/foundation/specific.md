# C Language Specific Features

## The C Preprocessor System

### Macro System
```c
#include <stdio.h>

// Object-like macros
#define PI 3.14159265359
#define MAX_BUFFER_SIZE 1024
#define DEBUG_MODE 1

// Function-like macros
#define SQUARE(x) ((x) * (x))
#define MAX(a, b) ((a) > (b) ? (a) : (b))
#define MIN(a, b) ((a) < (b) ? (a) : (b))

// Multi-statement macros
#define SWAP(a, b, type) do { \
    type temp = (a); \
    (a) = (b); \
    (b) = temp; \
} while(0)

// Stringification
#define STR(x) #x
#define XSTR(x) STR(x)

// Token pasting
#define CONCAT(a, b) a##b

void demonstrate_macros(void) {
    int x = 5, y = 10;
    printf("SQUARE(%d) = %d\n", x, SQUARE(x));
    printf("MAX(%d, %d) = %d\n", x, y, MAX(x, y));

    SWAP(x, y, int);
    printf("After swap: x=%d, y=%d\n", x, y);

    printf("PI as string: %s\n", STR(PI));
    printf("Line number: %s\n", XSTR(__LINE__));

    int CONCAT(var, 123) = 42;
    printf("var123 = %d\n", var123);
}
```

### Conditional Compilation
```c
// Platform-specific compilation
#ifdef _WIN32
    #include <windows.h>
    #define PATH_SEPARATOR '\\'
#elif defined(__linux__)
    #include <unistd.h>
    #define PATH_SEPARATOR '/'
#elif defined(__APPLE__)
    #include <unistd.h>
    #define PATH_SEPARATOR '/'
#else
    #error "Unsupported platform"
#endif

// Feature detection
#if defined(__STDC_VERSION__) && __STDC_VERSION__ >= 199901L
    // C99 features available
    #define HAVE_C99 1
    #include <stdint.h>
    #include <stdbool.h>
#else
    #define HAVE_C99 0
    typedef enum { false = 0, true = 1 } bool;
#endif

// Compiler-specific features
#ifdef __GNUC__
    #define PACKED __attribute__((packed))
    #define UNUSED __attribute__((unused))
    #define NORETURN __attribute__((noreturn))
#elif defined(_MSC_VER)
    #define PACKED
    #define UNUSED
    #define NORETURN __declspec(noreturn)
#else
    #define PACKED
    #define UNUSED
    #define NORETURN
#endif

void demonstrate_conditional_compilation(void) {
    printf("Path separator: %c\n", PATH_SEPARATOR);

    #if HAVE_C99
        bool c99_available = true;
        printf("C99 features: %s\n", c99_available ? "available" : "not available");
    #endif

    #ifdef DEBUG_MODE
        printf("Debug mode is enabled\n");
    #endif
}
```

### Variadic Macros (C99)
```c
#include <stdarg.h>

// C99 variadic macros
#define DEBUG_PRINT(fmt, ...) \
    do { \
        if (DEBUG_MODE) { \
            printf("[DEBUG] " fmt "\n", __VA_ARGS__); \
        } \
    } while(0)

// C23 __VA_OPT__ for optional comma
#if __STDC_VERSION__ >= 202311L
#define LOG(level, fmt, ...) \
    printf("[%s] " fmt "\n", level __VA_OPT__(,) __VA_ARGS__)
#endif

// GNU extension: named variadic arguments
#ifdef __GNUC__
#define DEBUG_PRINTF(format, args...) \
    printf("DEBUG: " format, ##args)
#endif

void demonstrate_variadic_macros(void) {
    DEBUG_PRINT("Value is %d", 42);
    DEBUG_PRINT("Two values: %d and %s", 42, "hello");

    #ifdef __GNUC__
        DEBUG_PRINTF("GNU style: %d\n", 123);
        DEBUG_PRINTF("GNU with no args\n");
    #endif
}
```

## Function Pointers and Callbacks

### Basic Function Pointers
```c
// Function pointer declarations
int add(int a, int b) { return a + b; }
int subtract(int a, int b) { return a - b; }
int multiply(int a, int b) { return a * b; }
int divide_safe(int a, int b) { return b != 0 ? a / b : 0; }

// Function pointer typedef
typedef int (*binary_op_t)(int, int);

void demonstrate_function_pointers(void) {
    // Direct function pointer usage
    int (*operation)(int, int) = add;
    printf("5 + 3 = %d\n", operation(5, 3));

    // Array of function pointers
    binary_op_t operations[] = {add, subtract, multiply, divide_safe};
    const char* op_names[] = {"+", "-", "*", "/"};

    int a = 20, b = 5;
    for (int i = 0; i < 4; i++) {
        printf("%d %s %d = %d\n", a, op_names[i], b, operations[i](a, b));
    }

    // Function pointer as parameter
    apply_operation(10, 3, multiply);
}

void apply_operation(int a, int b, binary_op_t op) {
    printf("Operation result: %d\n", op(a, b));
}
```

### Callback Systems
```c
#include <stdlib.h>

// Callback function type for array processing
typedef void (*array_callback_t)(int value, int index, void* user_data);

void process_array(int* array, size_t length, array_callback_t callback, void* user_data) {
    for (size_t i = 0; i < length; i++) {
        callback(array[i], i, user_data);
    }
}

// Example callbacks
void print_value(int value, int index, void* user_data) {
    char* prefix = (char*)user_data;
    printf("%s[%d] = %d\n", prefix ? prefix : "", index, value);
}

void sum_values(int value, int index, void* user_data) {
    int* total = (int*)user_data;
    *total += value;
}

void demonstrate_callbacks(void) {
    int numbers[] = {10, 20, 30, 40, 50};
    size_t length = sizeof(numbers) / sizeof(numbers[0]);

    // Print with callback
    printf("Array contents:\n");
    process_array(numbers, length, print_value, "nums");

    // Sum with callback
    int total = 0;
    process_array(numbers, length, sum_values, &total);
    printf("Total sum: %d\n", total);
}
```

### State Machines with Function Pointers
```c
typedef enum {
    STATE_IDLE,
    STATE_RUNNING,
    STATE_PAUSED,
    STATE_STOPPED,
    STATE_COUNT
} state_t;

typedef enum {
    EVENT_START,
    EVENT_PAUSE,
    EVENT_RESUME,
    EVENT_STOP,
    EVENT_COUNT
} event_t;

// State transition function type
typedef state_t (*state_handler_t)(event_t event);

// State handler functions
state_t handle_idle(event_t event) {
    switch (event) {
        case EVENT_START: return STATE_RUNNING;
        default: return STATE_IDLE;
    }
}

state_t handle_running(event_t event) {
    switch (event) {
        case EVENT_PAUSE: return STATE_PAUSED;
        case EVENT_STOP: return STATE_STOPPED;
        default: return STATE_RUNNING;
    }
}

state_t handle_paused(event_t event) {
    switch (event) {
        case EVENT_RESUME: return STATE_RUNNING;
        case EVENT_STOP: return STATE_STOPPED;
        default: return STATE_PAUSED;
    }
}

state_t handle_stopped(event_t event) {
    switch (event) {
        case EVENT_START: return STATE_RUNNING;
        default: return STATE_STOPPED;
    }
}

// State machine
typedef struct {
    state_t current_state;
    state_handler_t handlers[STATE_COUNT];
} state_machine_t;

void state_machine_init(state_machine_t* sm) {
    sm->current_state = STATE_IDLE;
    sm->handlers[STATE_IDLE] = handle_idle;
    sm->handlers[STATE_RUNNING] = handle_running;
    sm->handlers[STATE_PAUSED] = handle_paused;
    sm->handlers[STATE_STOPPED] = handle_stopped;
}

void state_machine_handle_event(state_machine_t* sm, event_t event) {
    state_t old_state = sm->current_state;
    sm->current_state = sm->handlers[sm->current_state](event);

    if (old_state != sm->current_state) {
        printf("State transition: %d -> %d\n", old_state, sm->current_state);
    }
}
```

## Variadic Functions

### Basic Variadic Functions
```c
#include <stdarg.h>

// Simple variadic function
int sum_integers(int count, ...) {
    va_list args;
    va_start(args, count);

    int total = 0;
    for (int i = 0; i < count; i++) {
        total += va_arg(args, int);
    }

    va_end(args);
    return total;
}

// Printf-style variadic function
void debug_printf(const char* format, ...) {
    va_list args;
    va_start(args, format);

    printf("[DEBUG] ");
    vprintf(format, args);
    printf("\n");

    va_end(args);
}

// Type-safe variadic function using sentinel
void print_strings(const char* first, ...) {
    if (!first) return;

    va_list args;
    va_start(args, first);

    printf("Strings: %s", first);

    const char* str;
    while ((str = va_arg(args, const char*)) != NULL) {
        printf(", %s", str);
    }
    printf("\n");

    va_end(args);
}

void demonstrate_variadic_functions(void) {
    int result = sum_integers(5, 10, 20, 30, 40, 50);
    printf("Sum: %d\n", result);

    debug_printf("Value is %d, name is %s", 42, "test");

    print_strings("first", "second", "third", NULL);  // NULL sentinel
}
```

## Inline Assembly

### GCC Inline Assembly
```c
#ifdef __GNUC__

// Basic inline assembly
int add_asm(int a, int b) {
    int result;
    __asm__ ("addl %2, %1; movl %1, %0"
             : "=r" (result)        // output operand
             : "r" (a), "r" (b)     // input operands
             : );                   // clobbered registers
    return result;
}

// Assembly with memory constraints
void memory_barrier(void) {
    __asm__ __volatile__ ("" ::: "memory");
}

// CPU-specific instructions
uint64_t read_tsc(void) {
    uint32_t low, high;
    __asm__ __volatile__ ("rdtsc" : "=a" (low), "=d" (high));
    return ((uint64_t)high << 32) | low;
}

// Atomic operations using assembly
int atomic_exchange(volatile int* ptr, int new_value) {
    int old_value;
    __asm__ __volatile__ (
        "xchgl %0, %1"
        : "=r" (old_value), "+m" (*ptr)
        : "0" (new_value)
        : "memory"
    );
    return old_value;
}

void demonstrate_inline_assembly(void) {
    printf("Assembly add: %d\n", add_asm(15, 25));

    uint64_t tsc1 = read_tsc();
    // Some work...
    uint64_t tsc2 = read_tsc();
    printf("TSC cycles: %llu\n", tsc2 - tsc1);

    volatile int test_var = 100;
    int old_val = atomic_exchange(&test_var, 200);
    printf("Atomic exchange: old=%d, new=%d\n", old_val, test_var);
}

#endif // __GNUC__
```

## Flexible Array Members (C99)

### Flexible Array Members
```c
#include <stdlib.h>
#include <string.h>

// Structure with flexible array member
typedef struct {
    size_t length;
    int data[];  // Flexible array member (C99)
} dynamic_array_t;

// Alternative pre-C99 idiom
typedef struct {
    size_t length;
    int data[1];  // Fake flexible array (allocate more than 1 element)
} old_style_array_t;

dynamic_array_t* create_dynamic_array(size_t size) {
    // Allocate struct plus space for array elements
    dynamic_array_t* arr = malloc(sizeof(dynamic_array_t) + size * sizeof(int));
    if (!arr) return NULL;

    arr->length = size;
    return arr;
}

// String buffer with flexible array member
typedef struct {
    size_t capacity;
    size_t length;
    char buffer[];  // Flexible array for string data
} string_buffer_t;

string_buffer_t* string_buffer_create(size_t capacity) {
    string_buffer_t* sb = malloc(sizeof(string_buffer_t) + capacity);
    if (!sb) return NULL;

    sb->capacity = capacity;
    sb->length = 0;
    sb->buffer[0] = '\0';
    return sb;
}

int string_buffer_append(string_buffer_t* sb, const char* str) {
    if (!sb || !str) return -1;

    size_t str_len = strlen(str);
    if (sb->length + str_len >= sb->capacity) {
        return -1;  // Not enough space
    }

    strcpy(sb->buffer + sb->length, str);
    sb->length += str_len;
    return 0;
}

void demonstrate_flexible_arrays(void) {
    // Create dynamic array
    dynamic_array_t* arr = create_dynamic_array(10);
    if (arr) {
        for (size_t i = 0; i < arr->length; i++) {
            arr->data[i] = i * i;
        }

        printf("Dynamic array contents:\n");
        for (size_t i = 0; i < arr->length; i++) {
            printf("%d ", arr->data[i]);
        }
        printf("\n");

        free(arr);
    }

    // Create string buffer
    string_buffer_t* sb = string_buffer_create(100);
    if (sb) {
        string_buffer_append(sb, "Hello");
        string_buffer_append(sb, ", ");
        string_buffer_append(sb, "World!");

        printf("String buffer: %s (length: %zu)\n", sb->buffer, sb->length);
        free(sb);
    }
}
```

## Designated Initializers (C99)

### Designated Initializers
```c
#include <stdio.h>

// Structure for demonstration
typedef struct {
    int id;
    char name[20];
    float salary;
    int department;
} employee_t;

// Enum for array indexing
enum color_index {
    RED_INDEX = 0,
    GREEN_INDEX = 1,
    BLUE_INDEX = 2
};

void demonstrate_designated_initializers(void) {
    // Designated initializers for structures
    employee_t emp1 = {
        .id = 12345,
        .name = "John Doe",
        .salary = 50000.0f,
        .department = 10
    };

    employee_t emp2 = {
        .salary = 60000.0f,  // Can be in any order
        .id = 67890,
        .name = "Jane Smith"
        // department will be initialized to 0
    };

    // Designated initializers for arrays
    int days_in_month[12] = {
        [0] = 31,   // January
        [1] = 28,   // February
        [2] = 31,   // March
        [3] = 30,   // April
        [11] = 31   // December
        // Others initialized to 0
    };

    // Mixed regular and designated initializers
    int mixed_array[10] = {
        1, 2, 3,        // First three elements
        [7] = 8,        // 8th element
        [9] = 10        // 10th element
    };

    // Color array using enum indices
    const char* colors[3] = {
        [RED_INDEX] = "Red",
        [GREEN_INDEX] = "Green",
        [BLUE_INDEX] = "Blue"
    };

    printf("Employee 1: ID=%d, Name=%s, Salary=%.2f\n",
           emp1.id, emp1.name, emp1.salary);
    printf("Employee 2: ID=%d, Name=%s, Salary=%.2f, Dept=%d\n",
           emp2.id, emp2.name, emp2.salary, emp2.department);

    printf("Days in months: Jan=%d, Feb=%d, Dec=%d\n",
           days_in_month[0], days_in_month[1], days_in_month[11]);

    for (int i = 0; i < 3; i++) {
        printf("Color %d: %s\n", i, colors[i]);
    }
}
```

## Compound Literals (C99)

### Compound Literals
```c
// Function that takes a point structure
typedef struct { int x, y; } point_t;

void print_point(point_t p) {
    printf("Point: (%d, %d)\n", p.x, p.y);
}

int* get_array_pointer(void) {
    // Return pointer to compound literal array
    // Note: This creates the array with static storage duration
    return (int[]){10, 20, 30, 40, 50};
}

void demonstrate_compound_literals(void) {
    // Compound literal for structure
    print_point((point_t){.x = 100, .y = 200});

    // Compound literal for array
    int* numbers = (int[]){1, 2, 3, 4, 5};
    for (int i = 0; i < 5; i++) {
        printf("%d ", numbers[i]);
    }
    printf("\n");

    // Compound literal with function call
    int* static_array = get_array_pointer();
    printf("Static array: ");
    for (int i = 0; i < 5; i++) {
        printf("%d ", static_array[i]);
    }
    printf("\n");

    // Modifying compound literal
    (int[]){1, 2, 3}[1] = 99;  // Modifies the compound literal

    // String compound literal
    char* greeting = (char[]){"Hello, World!"};
    printf("Greeting: %s\n", greeting);
}
```

## Restrict Qualifier and Aliasing

### Restrict Qualifier Usage
```c
#include <string.h>

// Function using restrict for optimization hints
void vector_add_restrict(double* restrict result,
                        const double* restrict a,
                        const double* restrict b,
                        size_t n) {
    for (size_t i = 0; i < n; i++) {
        result[i] = a[i] + b[i];  // Compiler can optimize knowing no aliasing
    }
}

// Memory copying with restrict
void* memcpy_restrict(void* restrict dest, const void* restrict src, size_t n) {
    char* d = (char*)dest;
    const char* s = (const char*)src;

    for (size_t i = 0; i < n; i++) {
        d[i] = s[i];  // No overlap assumed due to restrict
    }

    return dest;
}

// Matrix multiplication with restrict
void matrix_multiply(double* restrict c,
                    const double* restrict a,
                    const double* restrict b,
                    int n) {
    for (int i = 0; i < n; i++) {
        for (int j = 0; j < n; j++) {
            double sum = 0.0;
            for (int k = 0; k < n; k++) {
                sum += a[i * n + k] * b[k * n + j];
            }
            c[i * n + j] = sum;
        }
    }
}

void demonstrate_restrict(void) {
    double a[] = {1.0, 2.0, 3.0, 4.0};
    double b[] = {5.0, 6.0, 7.0, 8.0};
    double result[4];

    vector_add_restrict(result, a, b, 4);

    printf("Vector addition result: ");
    for (int i = 0; i < 4; i++) {
        printf("%.1f ", result[i]);
    }
    printf("\n");
}
```

## Zero-Length Arrays and Empty Structures

### GNU C Extensions
```c
#ifdef __GNUC__

// Zero-length array (GNU extension, pre-C99)
struct gnu_flexible {
    int length;
    char data[0];  // Zero-length array
};

// Empty structure (GNU extension)
struct empty {
    // No members
};

void demonstrate_gnu_extensions(void) {
    // Zero-length arrays work similarly to flexible arrays
    struct gnu_flexible* gf = malloc(sizeof(struct gnu_flexible) + 20);
    if (gf) {
        gf->length = 20;
        strcpy(gf->data, "GNU extension");
        printf("GNU flexible: %s\n", gf->data);
        free(gf);
    }

    // Empty structure has size 0 in GCC (1 in C++)
    printf("Empty struct size: %zu\n", sizeof(struct empty));
}

#endif // __GNUC__
```

## Pragma Directives

### Standard and Compiler-Specific Pragmas
```c
// Standard pragma
#pragma STDC FP_CONTRACT ON
#pragma STDC FENV_ACCESS OFF
#pragma STDC CX_LIMITED_RANGE OFF

// GCC-specific pragmas
#ifdef __GNUC__
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wunused-variable"
#pragma GCC optimize("O3")

void gcc_pragma_example(void) {
    int unused_var = 42;  // Warning suppressed by pragma
    printf("GCC pragma example\n");
}

#pragma GCC diagnostic pop
#endif

// Microsoft Visual C++ pragmas
#ifdef _MSC_VER
#pragma warning(push)
#pragma warning(disable: 4996)  // Disable deprecated function warnings

void msvc_pragma_example(void) {
    char buffer[100];
    strcpy(buffer, "MSVC example");  // Would normally warn about strcpy
    printf("%s\n", buffer);
}

#pragma warning(pop)
#endif

// Pack pragma for structure alignment
#pragma pack(push, 1)  // Pack structures to 1-byte alignment

struct packed_struct {
    char c;     // 1 byte
    int i;      // 4 bytes, no padding
    short s;    // 2 bytes
};  // Total: 7 bytes

#pragma pack(pop)  // Restore original packing

void demonstrate_pragmas(void) {
    printf("Packed struct size: %zu bytes\n", sizeof(struct packed_struct));

    #ifdef __GNUC__
        gcc_pragma_example();
    #endif

    #ifdef _MSC_VER
        msvc_pragma_example();
    #endif
}
```

These C-specific features set the language apart from other programming languages, providing low-level control, flexibility, and powerful abstraction mechanisms that have made C the foundation for system programming and embedded development.