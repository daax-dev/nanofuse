# Rust Local Build Toolchain

## Core Tools Installation

### Rustup - The Rust Toolchain Installer
```bash
# Install rustup (primary method)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Alternative: homebrew (macOS)
brew install rustup-init
rustup-init

# Alternative: package managers (Linux)
# Debian/Ubuntu
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Configure PATH
source ~/.cargo/env
```

### Toolchain Management
```bash
# Show installed toolchains
rustup show

# Install specific toolchain
rustup install stable
rustup install beta
rustup install nightly
rustup install 1.70.0  # Specific version

# Set default toolchain
rustup default stable
rustup default nightly
rustup default 1.70.0

# Update toolchains
rustup update
rustup update stable
rustup update nightly

# Add components to toolchain
rustup component add rustfmt
rustup component add clippy
rustup component add rust-src
rustup component add rust-analyzer

# Add targets for cross-compilation
rustup target add x86_64-pc-windows-gnu
rustup target add aarch64-apple-darwin
rustup target add wasm32-unknown-unknown

# List available targets
rustup target list
```

### Cargo - Build System and Package Manager
```bash
# Check cargo version
cargo --version

# Create new project
cargo new my_project
cargo new my_project --lib  # Library project
cargo new my_project --bin  # Binary project (default)

# Initialize in existing directory
cargo init
cargo init --lib
cargo init --bin

# Build project
cargo build              # Debug build
cargo build --release    # Release build
cargo build --target x86_64-pc-windows-gnu  # Cross-compile

# Run project
cargo run                # Run debug build
cargo run --release      # Run release build
cargo run --bin binary_name  # Run specific binary

# Test project
cargo test               # Run all tests
cargo test test_name     # Run specific test
cargo test --release     # Run tests in release mode
cargo test --lib         # Test only library
cargo test --bin binary_name  # Test specific binary

# Check project (compile without building)
cargo check              # Faster than build for syntax checking
cargo check --all-targets  # Check all targets

# Clean build artifacts
cargo clean

# Generate documentation
cargo doc                # Build documentation
cargo doc --open         # Build and open in browser
cargo doc --no-deps      # Don't include dependencies
```

## Project Structure and Configuration

### Standard Project Layout
```
my_project/
├── Cargo.toml          # Package manifest
├── Cargo.lock          # Dependency lock file (auto-generated)
├── src/
│   ├── main.rs         # Main binary entry point
│   ├── lib.rs          # Library root
│   ├── bin/            # Additional binaries
│   │   ├── binary1.rs
│   │   └── binary2.rs
│   └── modules/        # Module organization
│       ├── mod.rs
│       ├── module1.rs
│       └── module2.rs
├── tests/              # Integration tests
│   ├── integration_test.rs
│   └── common/
│       └── mod.rs
├── examples/           # Example programs
│   └── example1.rs
├── benches/            # Benchmarks
│   └── benchmark1.rs
├── docs/               # Additional documentation
├── assets/             # Static assets
└── README.md
```

### Cargo.toml Configuration
```toml
[package]
name = \"my_project\"
version = \"0.1.0\"
edition = \"2021\"
authors = [\"Your Name <email@example.com>\"]
description = \"A sample Rust project\"
homepage = \"https://example.com\"
repository = \"https://github.com/user/repo\"
documentation = \"https://docs.rs/my_project\"
readme = \"README.md\"
keywords = [\"cli\", \"tool\", \"utility\"]
categories = [\"command-line-utilities\"]
license = \"MIT OR Apache-2.0\"
license-file = \"LICENSE\"
include = [\"src/**/*\", \"Cargo.toml\", \"README.md\", \"LICENSE\"]
exclude = [\"tests/**/*\", \".github/**/*\"]
publish = true

[dependencies]
serde = { version = \"1.0\", features = [\"derive\"] }
tokio = { version = \"1.0\", features = [\"full\"] }
clap = { version = \"4.0\", features = [\"derive\"] }
anyhow = \"1.0\"
thiserror = \"1.0\"

# Optional dependencies
tracing = { version = \"0.1\", optional = true }
reqwest = { version = \"0.11\", optional = true, features = [\"json\"] }

# Development dependencies (only for testing/benching)
[dev-dependencies]
criterion = \"0.4\"
proptest = \"1.0\"
tempfile = \"3.0\"
tokio-test = \"0.4\"

# Build dependencies (for build scripts)
[build-dependencies]
cc = \"1.0\"
pkg-config = \"0.3\"

# Features
[features]
default = [\"logging\"]
logging = [\"dep:tracing\"]
networking = [\"dep:reqwest\"]
full = [\"logging\", \"networking\"]

# Multiple binaries
[[bin]]
name = \"main_tool\"
path = \"src/main.rs\"

[[bin]]
name = \"helper_tool\"
path = \"src/bin/helper.rs\"

# Examples
[[example]]
name = \"basic_usage\"
path = \"examples/basic.rs\"

# Benchmarks
[[bench]]
name = \"performance\"
harness = false

# Library configuration
[lib]
name = \"my_project\"
path = \"src/lib.rs\"
crate-type = [\"cdylib\", \"rlib\"]  # For FFI and regular use

# Profiles for different build configurations
[profile.dev]
opt-level = 0
debug = true
split-debuginfo = 'unpacked'
debug-assertions = true
overflow-checks = true
lto = false
panic = 'unwind'
incremental = true
codegen-units = 256
rpath = false

[profile.release]
opt-level = 3
debug = false
debug-assertions = false
overflow-checks = false
lto = true
panic = 'unwind'
incremental = false
codegen-units = 1
rpath = false

[profile.test]
opt-level = 0
debug = 2
debug-assertions = true
overflow-checks = true
lto = false
panic = 'unwind'
incremental = true
codegen-units = 256

[profile.bench]
opt-level = 3
debug = false
debug-assertions = false
overflow-checks = false
lto = true
panic = 'unwind'
incremental = false
codegen-units = 1

# Custom profiles (Rust 1.57+)
[profile.dist]
inherits = \"release\"
lto = true
codegen-units = 1

# Workspace configuration
[workspace]
members = [\"crate_a\", \"crate_b\", \"tools/*\"]
exclude = [\"third_party/*\"]

[workspace.dependencies]
serde = \"1.0\"
tokio = \"1.0\"

# Metadata for external tools
[package.metadata.docs.rs]
all-features = true
rustdoc-args = [\"--cfg\", \"docsrs\"]

[package.metadata.release]
sign-commit = true
pre-release-replacements = [
  {file=\"README.md\", search=\"Current version: [a-z0-9\\.-]+\", replace=\"Current version: {{version}}\"},
]
```

## Development Tools

### Code Formatting with rustfmt
```bash
# Install rustfmt
rustup component add rustfmt

# Format code
cargo fmt                # Format entire project
rustfmt src/main.rs     # Format specific file
cargo fmt -- --check    # Check formatting without changing files
cargo fmt -- --emit=files  # Show files that would be changed

# Configuration in rustfmt.toml or .rustfmt.toml
```

```toml
# rustfmt.toml
max_width = 100
hard_tabs = false
tab_spaces = 4
newline_style = \"Unix\"
use_small_heuristics = \"Default\"
indent_style = \"Block\"
wrap_comments = true
format_code_in_doc_comments = true
comment_width = 80
normalize_comments = true
normalize_doc_attributes = true
format_strings = true
format_macro_matchers = true
format_macro_bodies = true
empty_item_single_line = true
struct_lit_single_line = true
fn_single_line = false
where_single_line = false
imports_indent = \"Block\"
imports_layout = \"Mixed\"
group_imports = \"StdExternalCrate\"
```

### Linting with Clippy
```bash
# Install clippy
rustup component add clippy

# Run clippy
cargo clippy             # Basic linting
cargo clippy --all-targets  # Lint all targets
cargo clippy --all-features  # Lint with all features enabled
cargo clippy -- -D warnings  # Treat warnings as errors
cargo clippy -- -W clippy::pedantic  # Enable pedantic lints

# Fix suggestions automatically
cargo clippy --fix
cargo clippy --fix -- --allow-dirty --allow-staged

# Configuration in clippy.toml or .clippy.toml
```

```toml
# clippy.toml
avoid-breaking-exported-api = false
msrv = \"1.65.0\"  # Minimum supported Rust version

# Allowed lints
allow = [
    \"clippy::too_many_arguments\",
    \"clippy::type_complexity\",
]

# Denied lints
deny = [
    \"clippy::unwrap_used\",
    \"clippy::expect_used\",
    \"clippy::panic\",
]
```

### Documentation Generation
```bash
# Generate documentation
cargo doc                # Build docs
cargo doc --open         # Build and open in browser
cargo doc --no-deps      # Don't include dependencies
cargo doc --document-private-items  # Include private items

# Custom documentation settings
cargo doc --features \"feature1,feature2\"
cargo doc --all-features
cargo doc --no-default-features
```

### Testing Infrastructure
```bash
# Run tests
cargo test               # All tests
cargo test test_name     # Specific test
cargo test --lib         # Unit tests only
cargo test --test integration_test  # Specific integration test
cargo test --release     # Test optimized builds
cargo test -- --nocapture  # Show println! output
cargo test -- --test-threads=1  # Single threaded testing

# Test with features
cargo test --features \"feature1,feature2\"
cargo test --all-features
cargo test --no-default-features

# Documentation tests
cargo test --doc
```

### Benchmarking
```bash
# Install criterion (add to dev-dependencies)
cargo bench             # Run benchmarks
cargo bench bench_name  # Specific benchmark

# Using built-in unstable benchmarking (nightly only)
cargo +nightly bench
```

## Build Optimization and Profiles

### Custom Build Scripts (build.rs)
```rust
// build.rs
use std::env;
use std::path::PathBuf;

fn main() {
    // Tell cargo to rerun this build script if these files change
    println!(\"cargo:rerun-if-changed=src/native.c\");
    println!(\"cargo:rerun-if-changed=build.rs\");

    // Compile C code
    cc::Build::new()
        .file(\"src/native.c\")
        .compile(\"native\");

    // Link against system libraries
    println!(\"cargo:rustc-link-lib=ssl\");
    println!(\"cargo:rustc-link-lib=crypto\");

    // Set environment variables
    println!(\"cargo:rustc-env=VERSION={}\", env!(\"CARGO_PKG_VERSION\"));

    // Generate bindings
    let bindings = bindgen::Builder::default()
        .header(\"src/wrapper.h\")
        .generate()
        .expect(\"Unable to generate bindings\");

    let out_path = PathBuf::from(env::var(\"OUT_DIR\").unwrap());
    bindings
        .write_to_file(out_path.join(\"bindings.rs\"))
        .expect(\"Couldn't write bindings!\");
}
```

### Conditional Compilation
```rust
// Conditional features
#[cfg(feature = \"serde\")]
use serde::{Deserialize, Serialize};

#[cfg(feature = \"serde\")]
#[derive(Serialize, Deserialize)]
struct Data {
    value: i32,
}

// Target-specific code
#[cfg(target_os = \"windows\")]
fn platform_specific() {
    println!(\"Running on Windows\");
}

#[cfg(target_os = \"unix\")]
fn platform_specific() {
    println!(\"Running on Unix-like system\");
}

// Debug vs Release
#[cfg(debug_assertions)]
fn debug_only() {
    println!(\"This only runs in debug builds\");
}
```

### Cross-Compilation Setup
```bash
# Add cross-compilation targets
rustup target add x86_64-pc-windows-gnu
rustup target add aarch64-apple-darwin
rustup target add x86_64-unknown-linux-musl
rustup target add wasm32-unknown-unknown

# Cross-compile
cargo build --target x86_64-pc-windows-gnu
cargo build --target aarch64-apple-darwin

# For complex cross-compilation, use cross
cargo install cross
cross build --target x86_64-pc-windows-gnu
```

## Development Environment Setup

### VS Code Configuration
```json
// .vscode/settings.json
{
    \"rust-analyzer.cargo.features\": \"all\",
    \"rust-analyzer.checkOnSave.command\": \"clippy\",
    \"rust-analyzer.imports.granularity.group\": \"module\",
    \"rust-analyzer.imports.prefix\": \"crate\",
    \"editor.formatOnSave\": true,
    \"[rust]\": {
        \"editor.defaultFormatter\": \"rust-lang.rust-analyzer\",
        \"editor.formatOnSave\": true
    }
}

// .vscode/launch.json for debugging
{
    \"version\": \"0.2.0\",
    \"configurations\": [
        {
            \"type\": \"lldb\",
            \"request\": \"launch\",
            \"name\": \"Debug executable 'my_project'\",
            \"cargo\": {
                \"args\": [
                    \"build\",
                    \"--bin=my_project\",
                    \"--package=my_project\"
                ],
                \"filter\": {
                    \"name\": \"my_project\",
                    \"kind\": \"bin\"
                }
            },
            \"args\": [],
            \"cwd\": \"${workspaceFolder}\"
        }
    ]
}
```

### Environment Variables
```bash
# Rust-specific environment variables
export RUSTC_VERSION=$(rustc --version)
export CARGO_INCREMENTAL=1
export RUST_BACKTRACE=1          # Show backtraces on panic
export RUST_BACKTRACE=full       # Show full backtraces
export RUST_LOG=debug            # Logging level

# Build optimization
export CARGO_TARGET_DIR=target   # Build directory
export RUSTFLAGS=\"-C target-cpu=native\"  # Optimize for current CPU

# Cross-compilation helpers
export CC_x86_64_pc_windows_gnu=x86_64-w64-mingw32-gcc
export CARGO_TARGET_X86_64_PC_WINDOWS_GNU_LINKER=x86_64-w64-mingw32-gcc
```

### Cargo Configuration (.cargo/config.toml)
```toml
# .cargo/config.toml
[build]
target-dir = \"target\"
rustflags = [\"-C\", \"target-cpu=native\"]

[target.x86_64-unknown-linux-gnu]
linker = \"clang\"
rustflags = [\"-C\", \"link-arg=-fuse-ld=lld\"]

[target.x86_64-pc-windows-gnu]
linker = \"x86_64-w64-mingw32-gcc\"

[net]
retry = 2
git-fetch-with-cli = true

[registry]
default = \"crates-io\"

[registries.my-registry]
index = \"https://my-intranet:8080/git/index\"

[alias]
b = \"build\"
c = \"check\"
t = \"test\"
r = \"run\"
rr = \"run --release\"
br = \"build --release\"
cr = \"check --release\"

[env]
RUST_LOG = \"info\"
```

## Performance Profiling and Optimization

### Profiling Tools Setup
```bash
# Install profiling tools
cargo install flamegraph
cargo install cargo-criterion
cargo install cargo-benchcmp

# CPU profiling with flamegraph
cargo flamegraph

# Memory profiling with valgrind (Linux)
cargo install cargo-valgrind
cargo valgrind run

# Binary size analysis
cargo install cargo-bloat
cargo bloat --release
```

### Benchmark Configuration
```rust
// benches/my_benchmark.rs
use criterion::{black_box, criterion_group, criterion_main, Criterion};

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

## Debugging Setup

### GDB/LLDB Integration
```bash
# Install debugging symbols
rustup component add rust-src

# Debug with GDB (Linux)
rust-gdb target/debug/my_project

# Debug with LLDB (macOS)
rust-lldb target/debug/my_project

# Debug information in release builds
cargo build --release --debug
```

### Logging and Tracing
```rust
// Using tracing for structured logging
use tracing::{info, warn, error, debug, trace};
use tracing_subscriber;

fn main() {
    tracing_subscriber::fmt::init();

    info!(\"Application starting\");
    debug!(user_id = 42, \"User logged in\");
    warn!(\"This is a warning\");
    error!(\"Something went wrong\");
}
```

## Maintenance and Updates

### Dependency Management
```bash
# Update dependencies
cargo update              # Update to latest compatible versions
cargo update serde        # Update specific dependency
cargo update --precise 1.0.100 serde  # Update to specific version

# Audit dependencies for security issues
cargo install cargo-audit
cargo audit

# Check for outdated dependencies
cargo install cargo-outdated
cargo outdated

# Generate dependency tree
cargo tree
cargo tree --duplicates   # Show duplicate dependencies
cargo tree --format \"{p} {f}\"  # Custom format
```

### Cleaning and Optimization
```bash
# Clean build artifacts
cargo clean

# Remove unused dependencies
cargo install cargo-udeps
cargo +nightly udeps

# Check binary size
cargo install cargo-bloat
cargo bloat --release
cargo bloat --release --crates  # Group by crates

# Strip debug symbols from release builds
strip target/release/my_project
```

This comprehensive local build toolchain setup enables efficient Rust development with proper tooling, optimization, debugging, and maintenance capabilities.