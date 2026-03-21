# Debugging Rust Applications

## Debugging Fundamentals

### Debug vs Release Builds
```bash
# Debug build (default) - includes debug symbols
cargo build
cargo run

# Release build - optimized, minimal debug info
cargo build --release
cargo run --release

# Release with debug info
cargo build --release --debug
```

### Debug Configuration in Cargo.toml
```toml
[profile.dev]
debug = true           # Include debug symbols
debug-assertions = true
overflow-checks = true
opt-level = 0         # No optimization for better debugging
panic = 'unwind'      # Enable stack unwinding
incremental = true
split-debuginfo = 'unpacked'  # macOS: separate debug files

[profile.release]
debug = true          # Include symbols even in release
debug-assertions = false
overflow-checks = false
opt-level = 3
panic = 'unwind'
incremental = false
lto = false           # Disable LTO for better debug symbols

# Custom debug profile
[profile.debug-opt]
inherits = \"dev\"
opt-level = 1         # Light optimization for better performance
```

## Built-in Debugging Tools

### Panic and Backtrace Information
```rust
use std::env;

fn main() {
    // Enable backtraces
    env::set_var(\"RUST_BACKTRACE\", \"1\");
    // For full backtraces with more detail
    env::set_var(\"RUST_BACKTRACE\", \"full\");

    // This will panic and show backtrace
    let v = vec![1, 2, 3];
    println!(\"{}\", v[99]); // Index out of bounds
}
```

```bash
# Set backtrace environment variable
export RUST_BACKTRACE=1
export RUST_BACKTRACE=full

# Run with backtrace
RUST_BACKTRACE=1 cargo run
```

### Debug Printing
```rust
// Debug trait for custom types
#[derive(Debug)]
struct Person {
    name: String,
    age: u32,
}

fn main() {
    let person = Person {
        name: \"Alice\".to_string(),
        age: 30,
    };

    // Debug printing
    println!(\"{:?}\", person);          // Compact debug
    println!(\"{:#?}\", person);         // Pretty debug

    // Debug macro for quick debugging
    dbg!(&person);                      // Prints file, line, and value
    dbg!(person.age);                   // Shows expression and value

    // Conditional debug printing
    #[cfg(debug_assertions)]
    println!(\"Debug mode: {:?}\", person);
}
```

### Assert Macros for Debugging
```rust
fn main() {
    let x = 5;
    let y = 10;

    // Basic assertion
    assert!(x < y);
    assert!(x < y, \"x should be less than y\");

    // Equality assertions
    assert_eq!(x + y, 15);
    assert_eq!(x + y, 15, \"Addition failed: {} + {} != 15\", x, y);

    // Inequality assertions
    assert_ne!(x, y);
    assert_ne!(x, y, \"x and y should not be equal\");

    // Debug-only assertions (removed in release builds)
    debug_assert!(x < y);
    debug_assert_eq!(x + 5, y);
}
```

## Command-Line Debuggers

### GDB (GNU Debugger) Setup
```bash
# Install GDB with Rust support
# On Ubuntu/Debian
sudo apt-get install gdb

# On macOS (use lldb instead, but GDB available via homebrew)
brew install gdb

# Build with debug symbols
cargo build

# Start debugging with rust-gdb wrapper
rust-gdb target/debug/my_program

# Or run directly
gdb target/debug/my_program
```

### GDB Commands for Rust
```gdb
# Basic GDB commands for Rust
(gdb) break main                    # Set breakpoint at main
(gdb) break src/main.rs:15         # Set breakpoint at line 15
(gdb) break my_function            # Set breakpoint at function
(gdb) run                          # Start program
(gdb) run arg1 arg2                # Start with arguments

# Navigation
(gdb) step                         # Step into functions
(gdb) next                         # Step over functions
(gdb) continue                     # Continue execution
(gdb) finish                       # Finish current function

# Inspection
(gdb) print variable_name          # Print variable value
(gdb) print *pointer               # Dereference pointer
(gdb) print array[0]               # Print array element
(gdb) print struct_var.field       # Print struct field

# Rust-specific printing
(gdb) print vec                    # Print Vec contents
(gdb) print string                 # Print String contents
(gdb) print option                 # Print Option enum

# Stack and threads
(gdb) backtrace                    # Show call stack
(gdb) frame 2                      # Switch to frame 2
(gdb) info locals                  # Show local variables
(gdb) info args                    # Show function arguments

# Watchpoints
(gdb) watch variable_name          # Break when variable changes
(gdb) rwatch variable_name         # Break when variable is read
```

### LLDB (LLVM Debugger) Setup
```bash
# LLDB comes with Xcode on macOS
# On Linux, install via package manager
sudo apt-get install lldb

# Build with debug symbols
cargo build

# Start debugging with rust-lldb wrapper
rust-lldb target/debug/my_program

# Or run directly
lldb target/debug/my_program
```

### LLDB Commands for Rust
```lldb
# Basic LLDB commands
(lldb) breakpoint set --name main           # Set breakpoint at main
(lldb) breakpoint set --file main.rs --line 15  # Set breakpoint at line
(lldb) breakpoint set --name my_function    # Set breakpoint at function
(lldb) run                                  # Start program
(lldb) run arg1 arg2                        # Start with arguments

# Navigation
(lldb) step                                 # Step into
(lldb) next                                 # Step over
(lldb) continue                             # Continue execution
(lldb) finish                               # Finish function

# Inspection
(lldb) print variable_name                  # Print variable
(lldb) print *pointer                       # Dereference pointer
(lldb) frame variable                       # Show all local variables
(lldb) frame variable variable_name         # Show specific variable

# Stack navigation
(lldb) bt                                   # Show backtrace
(lldb) frame select 2                       # Select frame 2
(lldb) up                                   # Move up stack
(lldb) down                                 # Move down stack

# Watchpoints
(lldb) watchpoint set variable variable_name
(lldb) watchpoint set expression --size 4 variable_name
```

## IDE Integration

### VS Code Debugging Setup
```json
// .vscode/launch.json
{
    \"version\": \"0.2.0\",
    \"configurations\": [
        {
            \"type\": \"lldb\",
            \"request\": \"launch\",
            \"name\": \"Debug executable 'my_program'\",
            \"cargo\": {
                \"args\": [
                    \"build\",
                    \"--bin=my_program\",
                    \"--package=my_package\"
                ],
                \"filter\": {
                    \"name\": \"my_program\",
                    \"kind\": \"bin\"
                }
            },
            \"args\": [],
            \"cwd\": \"${workspaceFolder}\",
            \"console\": \"integratedTerminal\",
            \"sourceLanguages\": [\"rust\"]
        },
        {
            \"type\": \"lldb\",
            \"request\": \"launch\",
            \"name\": \"Debug unit tests\",
            \"cargo\": {
                \"args\": [
                    \"test\",
                    \"--no-run\",
                    \"--bin=my_program\",
                    \"--package=my_package\"
                ],
                \"filter\": {
                    \"name\": \"my_program\",
                    \"kind\": \"bin\"
                }
            },
            \"args\": [],
            \"cwd\": \"${workspaceFolder}\",
            \"console\": \"integratedTerminal\",
            \"sourceLanguages\": [\"rust\"]
        },
        {
            \"type\": \"lldb\",
            \"request\": \"launch\",
            \"name\": \"Debug with arguments\",
            \"cargo\": {
                \"args\": [
                    \"build\",
                    \"--bin=my_program\",
                    \"--package=my_package\"
                ],
                \"filter\": {
                    \"name\": \"my_program\",
                    \"kind\": \"bin\"
                }
            },
            \"args\": [\"--verbose\", \"input.txt\"],
            \"cwd\": \"${workspaceFolder}\",
            \"console\": \"integratedTerminal\",
            \"sourceLanguages\": [\"rust\"]
        }
    ]
}
```

### VS Code Settings for Rust Debugging
```json
// .vscode/settings.json
{
    \"rust-analyzer.checkOnSave.command\": \"check\",
    \"rust-analyzer.checkOnSave.allTargets\": false,
    \"rust-analyzer.debug.engine\": \"vadimcn.vscode-lldb\",
    \"rust-analyzer.debug.sourceFileMap\": {
        \"/rustc/*\": \"${env:USERPROFILE}/.rustup/toolchains/stable-x86_64-pc-windows-msvc/lib/rustlib/src/rust/*\"
    }
}
```

### CLion Debugging
```rust
// For CLion, create a run configuration:
// 1. Go to Run -> Edit Configurations
// 2. Add Cargo Command
// 3. Set Command: run
// 4. Set Arguments if needed
// 5. Set Environment variables
// 6. Enable \"Emulate terminal in output console\"
```

## Advanced Debugging Techniques

### Custom Debug Implementations
```rust
use std::fmt;

struct Person {
    name: String,
    age: u32,
    secrets: Vec<String>,
}

impl fmt::Debug for Person {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct(\"Person\")
            .field(\"name\", &self.name)
            .field(\"age\", &self.age)
            .field(\"secrets\", &\"<redacted>\")  // Hide sensitive data
            .finish()
    }
}

// Conditional debug formatting
impl fmt::Display for Person {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, \"{} (age {})\", self.name, self.age)
    }
}
```

### Logging for Debugging
```rust
use log::{debug, error, info, trace, warn};

fn main() {
    env_logger::init();

    info!(\"Application starting\");
    debug!(\"Debug information\");
    warn!(\"Warning message\");
    error!(\"Error occurred\");
    trace!(\"Detailed trace information\");
}

// Using tracing for structured logging
use tracing::{debug, error, info, warn, instrument, span, Level};

#[instrument]
fn complex_function(input: i32) -> i32 {
    let span = span!(Level::DEBUG, \"calculation\", input = %input);
    let _enter = span.enter();

    debug!(\"Processing input: {}\", input);
    let result = input * 2;
    debug!(\"Result calculated: {}\", result);
    result
}
```

### Memory Debugging with Valgrind
```bash
# Install cargo-valgrind
cargo install cargo-valgrind

# Run with valgrind (Linux only)
cargo valgrind run
cargo valgrind test

# Check for memory leaks
valgrind --leak-check=full --show-leak-kinds=all target/debug/my_program

# Check for threading issues
valgrind --tool=helgrind target/debug/my_program
valgrind --tool=drd target/debug/my_program
```

### AddressSanitizer (ASan)
```bash
# Enable AddressSanitizer (nightly required)
export RUSTFLAGS=\"-Z sanitizer=address\"
export ASAN_OPTIONS=\"detect_odr_violation=0\"
cargo +nightly run --target x86_64-unknown-linux-gnu

# Or add to Cargo.toml
[profile.dev]
opt-level = 1

[target.x86_64-unknown-linux-gnu]
rustflags = [\"-Z\", \"sanitizer=address\"]
```

### Memory Sanitizer (MSan)
```bash
# Enable MemorySanitizer
export RUSTFLAGS=\"-Z sanitizer=memory -Z sanitizer-memory-track-origins\"
cargo +nightly run --target x86_64-unknown-linux-gnu
```

### Thread Sanitizer (TSan)
```bash
# Enable ThreadSanitizer
export RUSTFLAGS=\"-Z sanitizer=thread\"
cargo +nightly run --target x86_64-unknown-linux-gnu
```

## Performance Debugging

### CPU Profiling with perf
```bash
# Record performance data
cargo build --release
perf record --call-graph=dwarf target/release/my_program

# View performance report
perf report

# Install cargo-flamegraph for flame graphs
cargo install flamegraph
cargo flamegraph

# Profile specific functions
perf record -g --call-graph=dwarf -F 99 target/release/my_program
```

### Memory Profiling
```bash
# Install cargo-instruments (macOS)
cargo install cargo-instruments

# Profile allocations (macOS)
cargo instruments -t \"Allocations\" --bin my_program

# Use heaptrack (Linux)
heaptrack target/release/my_program
heaptrack_gui heaptrack.my_program.*.gz
```

### Benchmark Debugging
```rust
use criterion::{criterion_group, criterion_main, Criterion, black_box};

fn fibonacci(n: u64) -> u64 {
    match n {
        0 => 1,
        1 => 1,
        n => fibonacci(n-1) + fibonacci(n-2),
    }
}

fn criterion_benchmark(c: &mut Criterion) {
    c.bench_function(\"fib 20\", |b| b.iter(|| fibonacci(black_box(20))));
}

criterion_group!(benches, criterion_benchmark);
criterion_main!(benches);
```

## Testing and Debugging

### Unit Test Debugging
```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_function() {
        let result = my_function(5);

        // Use dbg! in tests for debugging
        dbg!(&result);

        assert_eq!(result, 10);
    }

    #[test]
    #[should_panic(expected = \"division by zero\")]
    fn test_panic() {
        divide(10, 0);
    }

    #[test]
    fn test_with_debug_output() {
        println!(\"Debug output in test\");
        // Run with: cargo test -- --nocapture
        assert!(true);
    }
}
```

### Integration Test Debugging
```rust
// tests/integration_test.rs
use my_crate::*;

#[test]
fn integration_test() {
    // Set up test environment
    let temp_dir = tempfile::tempdir().unwrap();

    // Debug the test setup
    println!(\"Temp dir: {:?}\", temp_dir.path());

    // Test logic here
    assert!(true);
}
```

### Property-Based Testing with Debugging
```rust
use proptest::prelude::*;

proptest! {
    #[test]
    fn test_reverse_property(s in \".*\") {
        let reversed = s.chars().rev().collect::<String>();
        let double_reversed = reversed.chars().rev().collect::<String>();

        // Debug failing cases
        if s != double_reversed {
            println!(\"Original: {:?}\", s);
            println!(\"Reversed: {:?}\", reversed);
            println!(\"Double reversed: {:?}\", double_reversed);
        }

        prop_assert_eq!(s, double_reversed);
    }
}
```

## Environment-Specific Debugging

### Cross-Platform Debugging
```rust
#[cfg(target_os = \"windows\")]
fn platform_debug() {
    println!(\"Debugging on Windows\");
    // Windows-specific debug code
}

#[cfg(target_os = \"linux\")]
fn platform_debug() {
    println!(\"Debugging on Linux\");
    // Linux-specific debug code
}

#[cfg(target_os = \"macos\")]
fn platform_debug() {
    println!(\"Debugging on macOS\");
    // macOS-specific debug code
}
```

### Remote Debugging
```bash
# For remote debugging, use gdbserver (Linux)
gdbserver :1234 target/debug/my_program

# Connect from local machine
gdb target/debug/my_program
(gdb) target remote remote_host:1234
```

### Docker Container Debugging
```dockerfile
# Dockerfile for debugging
FROM rust:1.75

# Install debugging tools
RUN apt-get update && apt-get install -y \\
    gdb \\
    valgrind \\
    strace

WORKDIR /app
COPY . .

# Build with debug symbols
RUN cargo build

# Keep container running for debugging
CMD [\"sleep\", \"infinity\"]
```

```bash
# Debug in container
docker run -it --security-opt seccomp=unconfined my_rust_app /bin/bash
gdb target/debug/my_program
```

## Best Practices

### Debugging Checklist
1. **Build Configuration**
   - Use debug build for development
   - Include debug symbols in release builds when needed
   - Configure appropriate optimization levels

2. **Logging Strategy**
   - Use structured logging with tracing
   - Set appropriate log levels
   - Don't log sensitive information
   - Use conditional compilation for debug logs

3. **Error Handling**
   - Use Result<T, E> for recoverable errors
   - Provide meaningful error messages
   - Include context in error chains
   - Use panic! only for unrecoverable errors

4. **Testing**
   - Write unit tests for all functions
   - Include edge cases and error conditions
   - Use property-based testing for complex logic
   - Test debug and release builds

5. **Performance**
   - Profile before optimizing
   - Use benchmarks to track performance
   - Monitor memory usage and allocations
   - Test with realistic data sizes

6. **Tool Usage**
   - Learn your debugger thoroughly
   - Use IDE integration when available
   - Leverage Rust-specific tools (clippy, rustfmt)
   - Keep tools updated

This comprehensive debugging guide provides the tools and techniques necessary for effective Rust application debugging across different environments and scenarios.