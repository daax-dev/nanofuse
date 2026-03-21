# C Debugging Comprehensive Guide

## GDB (GNU Debugger)

### Basic GDB Usage
```bash
# Compile with debug information
gcc -g -O0 program.c -o program

# Start GDB
gdb ./program

# GDB commands
(gdb) run                    # Start program
(gdb) run arg1 arg2         # Start with arguments
(gdb) run < input.txt       # Start with input redirection

# Breakpoints
(gdb) break main            # Break at function
(gdb) break program.c:42    # Break at line
(gdb) break *0x4004f4       # Break at address
(gdb) info breakpoints      # List breakpoints
(gdb) delete 1              # Delete breakpoint 1
(gdb) disable 1             # Disable breakpoint 1

# Execution control
(gdb) continue              # Continue execution
(gdb) next                  # Next line (step over)
(gdb) step                  # Step into function
(gdb) finish                # Finish current function
(gdb) until                 # Run until current line

# Examining variables
(gdb) print variable_name   # Print variable
(gdb) print *pointer        # Print pointed value
(gdb) print array[5]        # Print array element
(gdb) display variable      # Auto-display on each stop
(gdb) undisplay 1           # Remove display 1

# Stack inspection
(gdb) backtrace             # Show call stack
(gdb) frame 2               # Switch to frame 2
(gdb) info locals           # Show local variables
(gdb) info args             # Show function arguments
```

### Advanced GDB Features
```bash
# Watchpoints
(gdb) watch variable        # Break when variable changes
(gdb) rwatch variable       # Break when variable is read
(gdb) awatch variable       # Break on read/write

# Conditional breakpoints
(gdb) break main if argc > 1
(gdb) condition 1 i == 42   # Add condition to breakpoint 1

# Memory examination
(gdb) x/10i main            # Examine 10 instructions at main
(gdb) x/4x 0x4004f4         # Examine 4 hex words at address
(gdb) x/s string_ptr        # Examine as string
(gdb) info registers        # Show CPU registers

# Core dump analysis
gdb program core.dump

# Remote debugging
(gdb) target remote localhost:1234

# GDB scripting
(gdb) source debug_script.gdb
```

### GDB Scripting Example
```gdb
# debug_script.gdb
define print_array
    set $i = 0
    while $i < $arg1
        print array[$i]
        set $i = $i + 1
    end
end

define hook-stop
    info locals
    backtrace 3
end

# Set initial breakpoints
break main
break error_handler

# Custom commands
define parray
    print_array $arg0
end

run
```

### GDB with Core Dumps
```bash
# Enable core dumps
ulimit -c unlimited

# Generate core dump on crash
./program
# Program crashes, generates core file

# Debug with core dump
gdb ./program core

# Analyze crash
(gdb) backtrace
(gdb) frame 0
(gdb) info locals
(gdb) list
```

## LLDB (LLVM Debugger)

### Basic LLDB Usage
```bash
# Start LLDB
lldb ./program

# LLDB commands
(lldb) run                     # Start program
(lldb) run arg1 arg2          # Start with arguments

# Breakpoints
(lldb) breakpoint set -n main          # Break at function
(lldb) breakpoint set -f program.c -l 42  # Break at line
(lldb) breakpoint list                 # List breakpoints
(lldb) breakpoint delete 1             # Delete breakpoint

# Execution control
(lldb) continue               # Continue execution
(lldb) next                   # Next line
(lldb) step                   # Step into
(lldb) finish                 # Finish function

# Variable examination
(lldb) frame variable variable_name    # Print variable
(lldb) expression variable_name       # Evaluate expression
(lldb) memory read 0x4004f4           # Read memory

# Stack inspection
(lldb) thread backtrace       # Show call stack
(lldb) frame select 2         # Select frame 2
(lldb) frame variable         # Show local variables
```

### LLDB vs GDB Command Comparison
| Operation | GDB | LLDB |
|-----------|-----|------|
| Set breakpoint at function | `break main` | `breakpoint set -n main` |
| Set breakpoint at line | `break file.c:42` | `breakpoint set -f file.c -l 42` |
| Print variable | `print var` | `frame variable var` |
| Show backtrace | `backtrace` | `thread backtrace` |
| Step over | `next` | `next` |
| Step into | `step` | `step` |

## Visual Studio Debugger

### Visual Studio Debugging
```cpp
// Debug build configuration
// Project Properties -> C/C++ -> Optimization -> Disabled (/Od)
// Project Properties -> C/C++ -> Code Generation -> Runtime Library -> Multi-threaded Debug (/MTd)

// Breakpoints
// F9 - Toggle breakpoint
// Ctrl+Shift+F9 - Remove all breakpoints

// Execution
// F5 - Start debugging
// Ctrl+F5 - Start without debugging
// F10 - Step over
// F11 - Step into
// Shift+F11 - Step out

// Windows
// Debug -> Windows -> Locals
// Debug -> Windows -> Watch
// Debug -> Windows -> Call Stack
// Debug -> Windows -> Memory
// Debug -> Windows -> Disassembly
```

### Visual Studio Code Debugging
```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug C Program",
            "type": "cppdbg",
            "request": "launch",
            "program": "${workspaceFolder}/program",
            "args": ["arg1", "arg2"],
            "stopAtEntry": false,
            "cwd": "${workspaceFolder}",
            "environment": [],
            "externalConsole": false,
            "MIMode": "gdb",
            "setupCommands": [
                {
                    "description": "Enable pretty-printing for gdb",
                    "text": "-enable-pretty-printing",
                    "ignoreFailures": true
                }
            ],
            "preLaunchTask": "build"
        }
    ]
}
```

```json
// .vscode/tasks.json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "build",
            "type": "shell",
            "command": "gcc",
            "args": ["-g", "-O0", "-o", "program", "main.c"],
            "group": {
                "kind": "build",
                "isDefault": true
            }
        }
    ]
}
```

## Static Analysis Tools

### Clang Static Analyzer
```bash
# Run static analyzer
clang --analyze src/*.c

# Generate HTML report
clang --analyze -Xanalyzer -analyzer-output=html \
      -Xanalyzer -analyzer-output-dir=analysis \
      src/*.c

# Specific checkers
clang --analyze -Xanalyzer -analyzer-checker=alpha.security.taint \
      src/*.c

# List available checkers
clang -cc1 -analyzer-checker-help
```

### Clang-Tidy
```bash
# Basic usage
clang-tidy src/*.c -- -I./include

# With configuration file
echo "Checks: '*,-readability-*'" > .clang-tidy
clang-tidy src/*.c

# Fix issues automatically
clang-tidy -fix src/*.c -- -I./include

# Generate compilation database
bear -- make  # or
cmake -DCMAKE_EXPORT_COMPILE_COMMANDS=ON .

# Use with compile_commands.json
clang-tidy src/*.c
```

### Cppcheck
```bash
# Basic scan
cppcheck src/

# Enable all checks
cppcheck --enable=all src/

# Specific platforms
cppcheck --platform=unix64 src/

# XML output
cppcheck --xml --xml-version=2 src/ 2> report.xml

# Exclude false positives
cppcheck --suppress=unusedFunction src/
```

### PVS-Studio (Commercial)
```bash
# Generate compilation database
pvs-studio-analyzer trace -- make

# Run analysis
pvs-studio-analyzer analyze -f compile_commands.json -j8

# Convert report
plog-converter -a GA:1,2 -t tasklist PVS-Studio.log
```

## Dynamic Analysis Tools

### Valgrind Memory Debugging
```bash
# Memory error detection (Memcheck)
valgrind --tool=memcheck ./program

# Detailed memory error reporting
valgrind --tool=memcheck --leak-check=full \
         --show-leak-kinds=all --track-origins=yes \
         --verbose ./program

# Heap profiling (Massif)
valgrind --tool=massif ./program
ms_print massif.out.12345

# Cache profiling (Cachegrind)
valgrind --tool=cachegrind ./program
cg_annotate cachegrind.out.12345

# Thread error detection (Helgrind)
valgrind --tool=helgrind ./program

# Data race detection (DRD)
valgrind --tool=drd ./program
```

### AddressSanitizer (ASan)
```bash
# Compile with AddressSanitizer
gcc -fsanitize=address -g -O1 program.c -o program

# Run with ASan
export ASAN_OPTIONS=detect_leaks=1:abort_on_error=1:print_stats=1
./program

# Symbolize addresses in crash reports
export ASAN_SYMBOLIZER_PATH=/usr/bin/llvm-symbolizer
./program

# ASan options
export ASAN_OPTIONS=verbosity=3:halt_on_error=1:detect_stack_use_after_return=1
```

### UndefinedBehaviorSanitizer (UBSan)
```bash
# Compile with UBSan
gcc -fsanitize=undefined -g program.c -o program

# Run with UBSan
export UBSAN_OPTIONS=print_stacktrace=1:abort_on_error=1
./program

# Specific UB checks
gcc -fsanitize=signed-integer-overflow,null,return,bounds \
    -g program.c -o program
```

### ThreadSanitizer (TSan)
```bash
# Compile with TSan
gcc -fsanitize=thread -g program.c -pthread -o program

# Run with TSan
export TSAN_OPTIONS=abort_on_error=1:halt_on_error=1
./program
```

### MemorySanitizer (MSan)
```bash
# Compile with MSan (Clang only)
clang -fsanitize=memory -g program.c -o program

# Run with MSan
export MSAN_OPTIONS=abort_on_error=1:print_stats=1
./program
```

## Debugging Techniques

### Printf Debugging
```c
#include <stdio.h>
#include <stdarg.h>

// Debug macro
#ifdef DEBUG
    #define DBG_PRINT(fmt, ...) \
        fprintf(stderr, "[DEBUG] %s:%d %s(): " fmt "\n", \
                __FILE__, __LINE__, __func__, ##__VA_ARGS__)
#else
    #define DBG_PRINT(fmt, ...) ((void)0)
#endif

// Trace macro
#define TRACE_ENTER() DBG_PRINT("Entering")
#define TRACE_EXIT() DBG_PRINT("Exiting")

// Variable dumping
#define DUMP_INT(var) DBG_PRINT(#var " = %d", var)
#define DUMP_STR(var) DBG_PRINT(#var " = %s", var ? var : "(null)")
#define DUMP_PTR(var) DBG_PRINT(#var " = %p", (void*)var)

void example_function(int x, const char* str) {
    TRACE_ENTER();
    DUMP_INT(x);
    DUMP_STR(str);

    // Function logic...

    TRACE_EXIT();
}
```

### Assertion-Based Debugging
```c
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>

// Custom assert with message
#define ASSERT_MSG(condition, message) \
    do { \
        if (!(condition)) { \
            fprintf(stderr, "Assertion failed: %s at %s:%d in %s()\n", \
                    message, __FILE__, __LINE__, __func__); \
            abort(); \
        } \
    } while(0)

// Precondition and postcondition checks
#define REQUIRE(condition) assert(condition)
#define ENSURE(condition) assert(condition)

void safe_array_access(int* array, size_t size, size_t index) {
    REQUIRE(array != NULL);
    REQUIRE(index < size);

    int value = array[index];

    ENSURE(value == array[index]);  // Postcondition
}
```

### Memory Debugging Helpers
```c
#include <stdlib.h>
#include <string.h>

#ifdef DEBUG_MEMORY
static size_t allocated_bytes = 0;
static size_t allocation_count = 0;

void* debug_malloc(size_t size, const char* file, int line) {
    void* ptr = malloc(size);
    if (ptr) {
        allocated_bytes += size;
        allocation_count++;
        printf("MALLOC: %zu bytes at %p (%s:%d) - Total: %zu bytes, Count: %zu\n",
               size, ptr, file, line, allocated_bytes, allocation_count);
        memset(ptr, 0xAA, size);  // Fill with pattern
    }
    return ptr;
}

void debug_free(void* ptr, const char* file, int line) {
    if (ptr) {
        allocation_count--;
        printf("FREE: %p (%s:%d) - Count: %zu\n", ptr, file, line, allocation_count);
        free(ptr);
    }
}

#define malloc(size) debug_malloc(size, __FILE__, __LINE__)
#define free(ptr) debug_free(ptr, __FILE__, __LINE__)

void print_memory_stats(void) {
    printf("Memory stats - Allocated: %zu bytes, Outstanding: %zu allocations\n",
           allocated_bytes, allocation_count);
}
#endif
```

### Stack Trace Generation
```c
#include <execinfo.h>
#include <stdio.h>
#include <stdlib.h>
#include <signal.h>

void print_stack_trace(void) {
    void* array[20];
    size_t size;
    char** strings;

    size = backtrace(array, 20);
    strings = backtrace_symbols(array, size);

    printf("Stack trace (%zu frames):\n", size);
    for (size_t i = 0; i < size; i++) {
        printf("  %s\n", strings[i]);
    }

    free(strings);
}

void signal_handler(int sig) {
    printf("Signal %d caught:\n", sig);
    print_stack_trace();
    exit(1);
}

void setup_signal_handlers(void) {
    signal(SIGSEGV, signal_handler);
    signal(SIGABRT, signal_handler);
    signal(SIGFPE, signal_handler);
}
```

## Embedded Systems Debugging

### JTAG/SWD Debugging
```bash
# OpenOCD configuration
# openocd.cfg
source [find interface/stlink-v2.cfg]
source [find target/stm32f4x.cfg]

# Start OpenOCD
openocd -f openocd.cfg

# Connect GDB to OpenOCD
arm-none-eabi-gdb program.elf
(gdb) target remote localhost:3333
(gdb) monitor reset halt
(gdb) load
(gdb) continue
```

### Serial/UART Debugging
```c
#include <stdio.h>

// Retarget printf to UART (embedded systems)
int _write(int file, char* ptr, int len) {
    for (int i = 0; i < len; i++) {
        uart_send_char(ptr[i]);
    }
    return len;
}

// Debug output macros for embedded
#define DEBUG_UART(fmt, ...) printf("[DEBUG] " fmt "\r\n", ##__VA_ARGS__)
#define ERROR_UART(fmt, ...) printf("[ERROR] " fmt "\r\n", ##__VA_ARGS__)
```

### Logic Analyzer Integration
```c
// GPIO-based debug signals
#define DEBUG_PIN_SET() (GPIOB->BSRR = GPIO_PIN_0)
#define DEBUG_PIN_CLEAR() (GPIOB->BSRR = GPIO_PIN_0 << 16)
#define DEBUG_PIN_TOGGLE() (GPIOB->ODR ^= GPIO_PIN_0)

void debug_function(void) {
    DEBUG_PIN_SET();    // Mark function entry

    // Function logic...

    DEBUG_PIN_CLEAR();  // Mark function exit
}
```

## Performance Profiling

### GNU Profiler (gprof)
```bash
# Compile with profiling
gcc -pg program.c -o program

# Run program to generate profile data
./program

# Analyze profile
gprof program gmon.out > analysis.txt
```

### Perf (Linux)
```bash
# Profile CPU usage
perf record ./program
perf report

# Profile specific events
perf record -e cache-misses ./program
perf record -e branch-misses ./program

# Real-time profiling
perf top

# System-wide profiling
perf record -a sleep 10
```

### Time-based Profiling
```c
#include <time.h>
#include <sys/time.h>

double get_time_seconds(void) {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return tv.tv_sec + tv.tv_usec / 1000000.0;
}

void profile_function(void) {
    double start_time = get_time_seconds();

    // Function to profile
    expensive_operation();

    double end_time = get_time_seconds();
    printf("Function took %.6f seconds\n", end_time - start_time);
}

// Cycle-accurate profiling (x86)
static inline uint64_t rdtsc(void) {
    uint32_t lo, hi;
    __asm__ __volatile__("rdtsc" : "=a" (lo), "=d" (hi));
    return ((uint64_t)hi << 32) | lo;
}
```

## Best Practices

### Debug Build Configuration
```makefile
# Makefile debug configuration
DEBUG ?= 1

ifeq ($(DEBUG), 1)
    CFLAGS += -g -O0 -DDEBUG -Wall -Wextra
    CFLAGS += -fsanitize=address -fsanitize=undefined
    LDFLAGS += -fsanitize=address -fsanitize=undefined
else
    CFLAGS += -O2 -DNDEBUG
endif
```

### Logging Framework
```c
typedef enum {
    LOG_TRACE, LOG_DEBUG, LOG_INFO, LOG_WARN, LOG_ERROR, LOG_FATAL
} log_level_t;

void log_message(log_level_t level, const char* file, int line,
                const char* func, const char* fmt, ...) {
    const char* level_names[] = {"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"};

    va_list args;
    va_start(args, fmt);

    fprintf(stderr, "[%s] %s:%d %s(): ", level_names[level], file, line, func);
    vfprintf(stderr, fmt, args);
    fprintf(stderr, "\n");

    va_end(args);

    if (level == LOG_FATAL) {
        abort();
    }
}

#define LOG_TRACE(...) log_message(LOG_TRACE, __FILE__, __LINE__, __func__, __VA_ARGS__)
#define LOG_DEBUG(...) log_message(LOG_DEBUG, __FILE__, __LINE__, __func__, __VA_ARGS__)
#define LOG_INFO(...) log_message(LOG_INFO, __FILE__, __LINE__, __func__, __VA_ARGS__)
#define LOG_WARN(...) log_message(LOG_WARN, __FILE__, __LINE__, __func__, __VA_ARGS__)
#define LOG_ERROR(...) log_message(LOG_ERROR, __FILE__, __LINE__, __func__, __VA_ARGS__)
#define LOG_FATAL(...) log_message(LOG_FATAL, __FILE__, __LINE__, __func__, __VA_ARGS__)
```

### Debugging Checklist
1. **Compile with debug symbols** (-g flag)
2. **Disable optimization** (-O0) for debugging
3. **Enable all warnings** (-Wall -Wextra)
4. **Use static analysis tools** regularly
5. **Enable runtime sanitizers** in development
6. **Set up proper logging** throughout the application
7. **Use version control** to track working states
8. **Write unit tests** to isolate problems
9. **Document debugging sessions** for future reference
10. **Learn your tools** thoroughly (GDB, Valgrind, etc.)

Effective debugging in C requires a combination of the right tools, techniques, and systematic approaches. The key is to use multiple complementary methods to identify and fix issues efficiently.