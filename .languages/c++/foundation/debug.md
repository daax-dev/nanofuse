# C++ Debugging, Profiling, and Analysis Tools

## Overview

Effective C++ development requires robust debugging, profiling, and static analysis tools. This guide covers GDB, LLDB, Valgrind, AddressSanitizer, static analyzers, profiling tools, and debugging techniques for production C++ applications.

## Debugging with GDB

### Basic GDB Usage

```cpp
// debug_example.cpp
#include <iostream>
#include <vector>
#include <memory>

class Calculator {
private:
    std::vector<int> history;

public:
    int add(int a, int b) {
        int result = a + b;
        history.push_back(result);
        return result;
    }

    void print_history() const {
        for (size_t i = 0; i < history.size(); ++i) {
            std::cout << "Result " << i << ": " << history[i] << std::endl;
        }
    }

    int get_result(size_t index) const {
        return history[index];  // Potential out-of-bounds access
    }
};

int main() {
    Calculator calc;

    int result1 = calc.add(5, 3);
    int result2 = calc.add(10, 7);

    calc.print_history();

    // This might cause an issue
    int bad_result = calc.get_result(10);

    return 0;
}
```

#### Compilation for Debugging

```bash
# Compile with debug information
g++ -g -O0 -Wall -Wextra debug_example.cpp -o debug_example

# Or with Clang
clang++ -g -O0 -Wall -Wextra debug_example.cpp -o debug_example
```

#### GDB Commands and Workflow

```bash
# Start GDB
gdb ./debug_example

# GDB command reference
(gdb) help                          # Show help
(gdb) help breakpoints              # Help for specific topic

# Setting breakpoints
(gdb) break main                    # Break at main function
(gdb) break Calculator::add         # Break at member function
(gdb) break debug_example.cpp:25    # Break at line 25
(gdb) break *0x400567              # Break at address

# Conditional breakpoints
(gdb) break Calculator::get_result if index > 5

# Running the program
(gdb) run                          # Start execution
(gdb) run arg1 arg2                # Run with arguments

# Stepping through code
(gdb) next                         # Step over (n)
(gdb) step                         # Step into (s)
(gdb) continue                     # Continue execution (c)
(gdb) finish                       # Run until function returns

# Examining variables
(gdb) print variable_name          # Print variable (p)
(gdb) print/x variable_name        # Print in hex
(gdb) print calc.history           # Print member variable
(gdb) print *this                  # Print object

# Examining memory
(gdb) x/10i $pc                    # Examine 10 instructions at PC
(gdb) x/10x address                # Examine 10 hex values
(gdb) x/s string_ptr               # Examine string

# Call stack
(gdb) backtrace                    # Show call stack (bt)
(gdb) frame 2                      # Switch to frame 2
(gdb) up                           # Move up stack frame
(gdb) down                         # Move down stack frame

# Watchpoints
(gdb) watch variable_name          # Watch for changes
(gdb) rwatch variable_name         # Watch for reads
(gdb) awatch variable_name         # Watch for access

# Thread debugging
(gdb) info threads                 # List all threads
(gdb) thread 2                     # Switch to thread 2
(gdb) thread apply all bt          # Backtrace all threads
```

#### Advanced GDB Techniques

```bash
# Custom commands and scripting
(gdb) define print_vector
> set $i = 0
> while $i < $arg0.size()
>   print $arg0[$i++]
> end
> end

# Use custom command
(gdb) print_vector calc.history

# Python scripting in GDB
(gdb) python
>import gdb
>class VectorPrinter:
>    def __init__(self, val):
>        self.val = val
>    def to_string(self):
>        size = int(self.val['_M_impl']['_M_finish'] - self.val['_M_impl']['_M_start'])
>        return f"std::vector of size {size}"
>end

# Load GDB scripts
(gdb) source my_gdb_script.py

# Core dump analysis
gdb ./program core_dump_file

# Remote debugging
(gdb) target remote hostname:port
```

#### GDB Configuration

```bash
# ~/.gdbinit
set print pretty on
set print array on
set print array-indexes on
set print elements 200
set print object on
set print static-members on
set print vtbl on
set print demangle on
set demangle-style gnu-v3

# Enable history
set history save on
set history size 10000
set history filename ~/.gdb_history

# Auto-load Python scripts
add-auto-load-safe-path /usr/share/gdb/auto-load
add-auto-load-safe-path /usr/local/share/gdb/auto-load

# Custom aliases
alias -a bt = backtrace
alias -a p = print
alias -a ni = nexti
alias -a si = stepi
```

## Debugging with LLDB

### Basic LLDB Usage

```bash
# Start LLDB
lldb ./debug_example

# LLDB command reference
(lldb) help                        # Show help
(lldb) help breakpoint             # Help for breakpoints

# Setting breakpoints
(lldb) breakpoint set --name main  # b main
(lldb) breakpoint set --file debug_example.cpp --line 25
(lldb) breakpoint set --method Calculator::add
(lldb) breakpoint set --address 0x400567

# Conditional breakpoints
(lldb) breakpoint set --name Calculator::get_result --condition 'index > 5'

# Running and stepping
(lldb) run                         # r
(lldb) next                        # n
(lldb) step                        # s
(lldb) continue                    # c
(lldb) finish                      # finish current function

# Examining variables
(lldb) frame variable              # Show all variables
(lldb) frame variable calc         # Show specific variable
(lldb) expression calc.history     # expr or p
(lldb) expression/x variable_name  # Print in hex

# Memory examination
(lldb) memory read --size 4 --count 10 address
(lldb) disassemble --pc            # dis -p

# Call stack
(lldb) thread backtrace            # bt
(lldb) frame select 2              # f 2

# Watchpoints
(lldb) watchpoint set variable variable_name
(lldb) watchpoint set expression -- address

# Thread debugging
(lldb) thread list                 # Show all threads
(lldb) thread select 2             # Switch to thread 2
```

#### LLDB Configuration

```bash
# ~/.lldbinit
settings set target.x86-disassembly-flavor intel
settings set interpreter.prompt-on-quit false
settings set target.process.stop-on-sharedlibrary-events false

# Custom aliases
command alias bfl breakpoint set -f %1 -l %2
command alias bd breakpoint delete
command alias bl breakpoint list

# Python integration
command script import lldb_helpers
```

## Memory Debugging Tools

### AddressSanitizer (ASan)

```cpp
// memory_bugs.cpp - Examples of memory issues
#include <iostream>
#include <memory>

void heap_buffer_overflow() {
    int* arr = new int[10];
    arr[10] = 42;  // Heap buffer overflow
    delete[] arr;
}

void use_after_free() {
    int* ptr = new int(42);
    delete ptr;
    *ptr = 10;  // Use after free
}

void memory_leak() {
    int* ptr = new int(42);
    // Missing delete - memory leak
}

void stack_buffer_overflow() {
    int arr[10];
    arr[10] = 42;  // Stack buffer overflow
}

int main() {
    heap_buffer_overflow();
    use_after_free();
    memory_leak();
    stack_buffer_overflow();
    return 0;
}
```

#### Compiling with AddressSanitizer

```bash
# GCC with AddressSanitizer
g++ -g -O0 -fsanitize=address -fno-omit-frame-pointer memory_bugs.cpp -o memory_bugs

# Clang with AddressSanitizer
clang++ -g -O0 -fsanitize=address -fno-omit-frame-pointer memory_bugs.cpp -o memory_bugs

# Run with AddressSanitizer
./memory_bugs

# Configure AddressSanitizer options
export ASAN_OPTIONS="verbosity=3:halt_on_error=1:check_initialization_order=1"
./memory_bugs
```

#### AddressSanitizer Configuration

```bash
# Common ASAN_OPTIONS
export ASAN_OPTIONS="
halt_on_error=1
log_path=/tmp/asan_log
symbolize=1
print_stacktrace=1
check_initialization_order=1
strict_init_order=1
detect_stack_use_after_return=true
detect_leaks=1
abort_on_error=1
"
```

### ThreadSanitizer (TSan)

```cpp
// race_condition.cpp
#include <thread>
#include <iostream>

int global_counter = 0;

void increment() {
    for (int i = 0; i < 1000000; ++i) {
        global_counter++;  // Race condition
    }
}

int main() {
    std::thread t1(increment);
    std::thread t2(increment);

    t1.join();
    t2.join();

    std::cout << "Counter: " << global_counter << std::endl;
    return 0;
}
```

```bash
# Compile with ThreadSanitizer
g++ -g -O1 -fsanitize=thread race_condition.cpp -o race_condition

# Run with ThreadSanitizer
./race_condition

# Configure ThreadSanitizer
export TSAN_OPTIONS="halt_on_error=1:log_path=/tmp/tsan_log"
```

### MemorySanitizer (MSan)

```bash
# Compile with MemorySanitizer (Clang only)
clang++ -g -O1 -fsanitize=memory -fsanitize-memory-track-origins memory_bugs.cpp -o memory_bugs_msan

# Configure MemorySanitizer
export MSAN_OPTIONS="halt_on_error=1:print_stats=1"
./memory_bugs_msan
```

### UndefinedBehaviorSanitizer (UBSan)

```cpp
// undefined_behavior.cpp
#include <iostream>

int main() {
    int x = 2147483647;  // INT_MAX
    int y = x + 1;       // Integer overflow

    int* p = nullptr;
    int z = *p;          // Null pointer dereference

    int arr[5];
    int w = arr[10];     // Array bounds violation

    return 0;
}
```

```bash
# Compile with UBSan
g++ -g -O1 -fsanitize=undefined undefined_behavior.cpp -o undefined_behavior

# Configure UBSan
export UBSAN_OPTIONS="halt_on_error=1:print_stacktrace=1"
./undefined_behavior
```

### Valgrind

```bash
# Memory checking with Memcheck
valgrind --tool=memcheck \
         --leak-check=full \
         --show-leak-kinds=all \
         --track-origins=yes \
         --verbose \
         --log-file=valgrind.log \
         ./memory_bugs

# Helgrind for race condition detection
valgrind --tool=helgrind ./race_condition

# Cachegrind for cache profiling
valgrind --tool=cachegrind ./program

# Massif for heap profiling
valgrind --tool=massif ./program
```

#### Valgrind Suppressions

```bash
# valgrind.supp
{
   ignore_std_string_leaks
   Memcheck:Leak
   match-leak-kinds: reachable
   ...
   fun:*std::string*
}
```

```bash
# Use suppressions
valgrind --suppressions=valgrind.supp --tool=memcheck ./program
```

## Static Analysis Tools

### Clang Static Analyzer

```bash
# Run scan-build
scan-build make

# Or with specific checkers
scan-build --use-cc=clang --use-c++=clang++ \
           -enable-checker alpha.core.CastSize \
           -enable-checker alpha.core.CastToStruct \
           -enable-checker alpha.security.taint.TaintPropagation \
           make

# Generate HTML report
scan-build -o /tmp/scan-build-results make
```

### Clang-Tidy

```bash
# Basic usage
clang-tidy src/*.cpp -- -I./include -std=c++20

# With compile commands database
clang-tidy -p build src/*.cpp

# Specific checks
clang-tidy -checks='bugprone-*,modernize-*' src/*.cpp

# Fix issues automatically
clang-tidy -fix src/*.cpp --
```

#### Clang-Tidy Configuration

```yaml
# .clang-tidy
Checks: >
  -*,
  bugprone-*,
  clang-analyzer-*,
  cppcoreguidelines-*,
  modernize-*,
  performance-*,
  readability-*

CheckOptions:
  - key: readability-identifier-naming.ClassCase
    value: CamelCase
  - key: readability-identifier-naming.FunctionCase
    value: lower_case
```

### PVS-Studio

```bash
# Generate compile commands
pvs-studio-analyzer trace -- make

# Run analysis
pvs-studio-analyzer analyze -o project.log

# Convert to readable format
plog-converter -t tasklist project.log
```

### Cppcheck

```bash
# Basic usage
cppcheck --enable=all --std=c++20 src/

# With specific checks
cppcheck --enable=warning,style,performance,portability \
         --inconclusive \
         --inline-suppr \
         --suppress=missingIncludeSystem \
         src/

# Generate XML report
cppcheck --xml --xml-version=2 --output-file=cppcheck.xml src/
```

## Profiling Tools

### GNU gprof

```cpp
// performance_test.cpp
#include <iostream>
#include <vector>
#include <chrono>

void expensive_function() {
    std::vector<int> vec(1000000);
    for (int i = 0; i < 1000000; ++i) {
        vec[i] = i * i;
    }
}

void another_function() {
    for (int i = 0; i < 1000; ++i) {
        expensive_function();
    }
}

int main() {
    another_function();
    return 0;
}
```

```bash
# Compile with profiling
g++ -pg -O2 performance_test.cpp -o performance_test

# Run program
./performance_test

# Generate profile report
gprof ./performance_test gmon.out > profile_report.txt
```

### perf (Linux Performance Tools)

```bash
# Record performance data
perf record -g ./performance_test

# View report
perf report

# Live monitoring
perf top

# Specific events
perf stat -e cycles,instructions,cache-misses ./performance_test

# Memory access patterns
perf mem record ./performance_test
perf mem report

# Call graph profiling
perf record --call-graph dwarf ./performance_test
```

### Intel VTune Profiler

```bash
# Basic profiling
vtune -collect hotspots ./performance_test

# Memory access analysis
vtune -collect memory-access ./performance_test

# Threading analysis
vtune -collect threading ./performance_test

# Microarchitecture analysis
vtune -collect uarch-exploration ./performance_test

# View results
vtune-gui
```

### Google Benchmark Integration

```cpp
// benchmark_example.cpp
#include <benchmark/benchmark.h>
#include <vector>
#include <algorithm>

static void BM_VectorSort(benchmark::State& state) {
    std::vector<int> vec(state.range(0));

    for (auto _ : state) {
        // Setup
        std::iota(vec.begin(), vec.end(), 0);
        std::reverse(vec.begin(), vec.end());

        // Benchmark
        benchmark::DoNotOptimize(vec.data());
        std::sort(vec.begin(), vec.end());
        benchmark::ClobberMemory();
    }

    state.SetComplexityN(state.range(0));
}

BENCHMARK(BM_VectorSort)
    ->Range(1<<10, 1<<20)
    ->Complexity(benchmark::oNLogN);

BENCHMARK_MAIN();
```

```bash
# Compile with benchmark
g++ -O2 -lbenchmark -lpthread benchmark_example.cpp -o benchmark_example

# Run benchmarks
./benchmark_example --benchmark_format=json --benchmark_out=results.json
```

## Advanced Debugging Techniques

### Core Dump Analysis

```bash
# Enable core dumps
ulimit -c unlimited

# Generate core dump on crash
./program_that_crashes

# Analyze core dump
gdb ./program core

# Or with LLDB
lldb ./program -c core
```

### Remote Debugging

```bash
# Start gdbserver on remote machine
gdbserver :2345 ./program

# Connect from local machine
gdb ./program
(gdb) target remote hostname:2345
```

### Debugging Optimized Code

```bash
# Compile with debug info and optimization
g++ -g -O2 -DNDEBUG program.cpp -o program

# Use specific GDB settings for optimized code
(gdb) set print frame-arguments all
(gdb) set print entry-values both
(gdb) set scheduler-locking on
```

### Debugging Templates and Metaprogramming

```cpp
// template_debug.cpp
template<int N>
struct Factorial {
    static constexpr int value = N * Factorial<N-1>::value;
};

template<>
struct Factorial<0> {
    static constexpr int value = 1;
};

int main() {
    constexpr int result = Factorial<5>::value;
    return result;
}
```

```bash
# Compile with template debugging
g++ -g -ftemplate-backtrace-limit=0 template_debug.cpp

# GDB template debugging
(gdb) set print demangle on
(gdb) set print asm-demangle on
(gdb) info types  # Show all types
(gdb) ptype Factorial<5>  # Show template instantiation
```

### Custom Debugging Aids

```cpp
// debug_utils.hpp
#ifndef DEBUG_UTILS_HPP
#define DEBUG_UTILS_HPP

#include <iostream>
#include <string>
#include <chrono>

// Debug macros
#ifdef DEBUG
    #define DBG(x) std::cout << #x << " = " << (x) << std::endl
    #define TRACE() std::cout << "TRACE: " << __FILE__ << ":" << __LINE__ << std::endl
    #define LOG(msg) std::cout << "LOG: " << msg << std::endl
#else
    #define DBG(x)
    #define TRACE()
    #define LOG(msg)
#endif

// RAII timer for performance measurement
class Timer {
private:
    std::chrono::high_resolution_clock::time_point start;
    std::string name;

public:
    Timer(const std::string& name) : name(name) {
        start = std::chrono::high_resolution_clock::now();
    }

    ~Timer() {
        auto end = std::chrono::high_resolution_clock::now();
        auto duration = std::chrono::duration_cast<std::chrono::microseconds>(end - start);
        std::cout << name << " took " << duration.count() << " microseconds" << std::endl;
    }
};

#define TIME_SCOPE(name) Timer timer(name)

// Memory usage tracking
class MemoryTracker {
private:
    static size_t allocated_bytes;

public:
    static void* allocate(size_t size) {
        allocated_bytes += size;
        std::cout << "Allocated " << size << " bytes (total: " << allocated_bytes << ")" << std::endl;
        return std::malloc(size);
    }

    static void deallocate(void* ptr, size_t size) {
        allocated_bytes -= size;
        std::cout << "Deallocated " << size << " bytes (total: " << allocated_bytes << ")" << std::endl;
        std::free(ptr);
    }

    static size_t get_allocated_bytes() { return allocated_bytes; }
};

size_t MemoryTracker::allocated_bytes = 0;

#endif // DEBUG_UTILS_HPP
```

### Debugging Multi-threaded Applications

```cpp
// thread_debug.cpp
#include <thread>
#include <mutex>
#include <condition_variable>
#include <iostream>

std::mutex mtx;
std::condition_variable cv;
bool ready = false;

void thread_function(int id) {
    std::unique_lock<std::mutex> lock(mtx);

    // Wait for signal
    cv.wait(lock, []{ return ready; });

    std::cout << "Thread " << id << " is running" << std::endl;

    // Simulate work
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
}

int main() {
    std::vector<std::thread> threads;

    // Create threads
    for (int i = 0; i < 4; ++i) {
        threads.emplace_back(thread_function, i);
    }

    // Signal all threads
    {
        std::lock_guard<std::mutex> lock(mtx);
        ready = true;
    }
    cv.notify_all();

    // Wait for completion
    for (auto& t : threads) {
        t.join();
    }

    return 0;
}
```

```bash
# Debug multi-threaded applications
gdb ./thread_debug
(gdb) set scheduler-locking on
(gdb) info threads
(gdb) thread apply all bt
```

## Best Practices

### Debugging Workflow

1. **Reproduce the Issue**
   - Create minimal test case
   - Document steps to reproduce
   - Use consistent environment

2. **Use Appropriate Tools**
   - Debug builds for development
   - Sanitizers for memory issues
   - Valgrind for detailed analysis
   - Static analyzers for code quality

3. **Logging and Tracing**
   - Strategic logging placement
   - Different log levels
   - Thread-safe logging
   - Performance impact consideration

4. **Testing Integration**
   - Unit tests for components
   - Integration tests for workflows
   - Stress tests for reliability
   - Continuous testing in CI/CD

This comprehensive guide provides the tools and techniques needed for effective debugging, profiling, and analysis of C++ applications in production environments.