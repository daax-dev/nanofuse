# Rust Packaging and Distribution

## Cargo Package Management System

### Package vs Crate Terminology
- **Package**: A bundle of one or more crates with a `Cargo.toml` file
- **Crate**: A compilation unit (library or binary)
- **Module**: Organization unit within a crate
- **Workspace**: Collection of related packages

### Package Types
```rust
// Library crate (src/lib.rs)
// Can be used by other crates
pub fn public_function() {
    println!("This is a library function");
}

// Binary crate (src/main.rs)
// Produces an executable
fn main() {
    println!("This is a binary executable");
}

// Mixed package (both lib.rs and main.rs)
// Has both library and binary components
```

### Cargo.toml Package Configuration
```toml
[package]
name = "my_awesome_crate"
version = "0.1.0"
edition = "2021"
rust-version = "1.65"
authors = ["Your Name <your.email@example.com>"]
license = "MIT OR Apache-2.0"
description = "A comprehensive Rust package for awesome functionality"
documentation = "https://docs.rs/my_awesome_crate"
homepage = "https://github.com/user/my_awesome_crate"
repository = "https://github.com/user/my_awesome_crate"
readme = "README.md"
keywords = ["cli", "tool", "awesome", "rust"]
categories = ["command-line-utilities", "development-tools"]
exclude = [
    "tests/*",
    "benches/*",
    "examples/large_example/*",
    ".github/*"
]
include = [
    "src/**/*",
    "Cargo.toml",
    "Cargo.lock",
    "README.md",
    "LICENSE*"
]

# Minimum Rust version required
rust-version = "1.65.0"

# Control publishing
publish = true  # or false to prevent publishing
# publish = ["my-registry"]  # Publish only to specific registries

[lib]
name = "my_awesome_crate"
path = "src/lib.rs"
# Crate types: rlib (default), cdylib, staticlib, dylib, proc-macro
crate-type = ["rlib", "cdylib"]
# Make library available to other languages via C ABI
# crate-type = ["cdylib"]

# Multiple binary targets
[[bin]]
name = "main_tool"
path = "src/main.rs"

[[bin]]
name = "helper_tool"
path = "src/bin/helper.rs"

# Example targets
[[example]]
name = "basic_usage"
path = "examples/basic.rs"
required-features = ["cli"]

# Benchmark targets
[[bench]]
name = "performance"
harness = false  # Use custom benchmarking framework
required-features = ["benchmarks"]

# Test targets
[[test]]
name = "integration"
path = "tests/integration_test.rs"
```

## Semantic Versioning and Releases

### SemVer in Rust
```toml
# Version format: MAJOR.MINOR.PATCH
version = "1.2.3"

# Pre-release versions
version = "1.0.0-alpha.1"
version = "1.0.0-beta.2"
version = "1.0.0-rc.1"

# Build metadata
version = "1.0.0+build.123"
version = "1.0.0-beta.1+exp.sha.5114f85"
```

### Version Compatibility Rules
```toml
[dependencies]
# Exact version
exact_version = "=1.2.3"

# Compatible versions (default caret requirement)
compatible = "1.2"      # >=1.2.0, <2.0.0
compatible = "1.2.3"    # >=1.2.3, <2.0.0

# Tilde requirement (patch level changes)
patch_updates = "~1.2.3"  # >=1.2.3, <1.3.0
patch_updates = "~1.2"    # >=1.2.0, <1.3.0

# Wildcard requirement
wildcard = "1.*"        # >=1.0.0, <2.0.0
wildcard = "1.2.*"      # >=1.2.0, <1.3.0

# Inequality requirements
range = ">=1.2.0, <1.5.0"
range = ">1.2.0, <=1.4.0"

# Pre-release handling
prerelease = "1.0.0-alpha.1"  # Must specify exactly for pre-releases
```

### Breaking Change Guidelines
```rust
// MAJOR version changes (breaking changes):
// - Remove public API
// - Change function signatures
// - Change behavior significantly

// Before (v1.0.0)
pub fn process_data(data: &str) -> String {
    data.to_uppercase()
}

// After (v2.0.0) - BREAKING: Changed return type
pub fn process_data(data: &str) -> Result<String, ProcessError> {
    if data.is_empty() {
        Err(ProcessError::EmptyInput)
    } else {
        Ok(data.to_uppercase())
    }
}

// MINOR version changes (non-breaking additions):
// - Add new public API
// - Add optional parameters
// - Deprecate (but don't remove) old API

// v1.1.0 - Added new function (non-breaking)
pub fn process_data_advanced(data: &str, options: &ProcessOptions) -> String {
    // New functionality
}

// PATCH version changes (bug fixes):
// - Fix bugs without changing API
// - Internal improvements
// - Documentation updates
```

## Dependency Management

### Dependency Types and Configuration
```toml
[dependencies]
# Production dependencies
serde = { version = "1.0", features = ["derive"] }
tokio = { version = "1.0", features = ["full"] }
clap = { version = "4.0", default-features = false, features = ["derive", "std"] }

# Optional dependencies (for feature gates)
tracing = { version = "0.1", optional = true }
reqwest = { version = "0.11", optional = true }

# Platform-specific dependencies
[target.'cfg(windows)'.dependencies]
winapi = "0.3"

[target.'cfg(unix)'.dependencies]
nix = "0.26"

# Development dependencies (testing, benchmarking)
[dev-dependencies]
criterion = "0.4"
proptest = "1.0"
tempfile = "3.0"
tokio-test = "0.4"

# Build dependencies (build.rs scripts)
[build-dependencies]
cc = "1.0"
bindgen = "0.65"
pkg-config = "0.3"

# Target-specific configurations
[target.x86_64-pc-windows-gnu.dependencies]
winapi = { version = "0.3", features = ["winuser"] }

[target.'cfg(target_os = "macos")'.dependencies]
core-foundation = "0.9"
```

### Feature Management
```toml
[features]
# Default features enabled by default
default = ["std", "serde"]

# Feature definitions
std = []  # Enable standard library
serde = ["dep:serde", "serde/derive"]  # Enable serde with derive
async = ["dep:tokio", "tokio/full"]    # Enable async functionality
cli = ["dep:clap", "clap/derive"]      # Enable CLI functionality

# Mutually exclusive features
tls-native = ["reqwest/native-tls"]
tls-rustls = ["reqwest/rustls-tls"]

# Internal features (not user-facing)
_internal_unstable = []

# Feature combinations
full = ["std", "serde", "async", "cli"]
minimal = []  # No default features
```

```rust
// Conditional compilation based on features
#[cfg(feature = "serde")]
use serde::{Deserialize, Serialize};

#[cfg(feature = "serde")]
#[derive(Serialize, Deserialize)]
pub struct Data {
    value: i32,
}

#[cfg(not(feature = "std"))]
use alloc::vec::Vec;

#[cfg(feature = "std")]
use std::vec::Vec;

// Feature-dependent API
#[cfg(feature = "async")]
pub async fn async_function() -> Result<String, Error> {
    // Async implementation
    Ok("async result".to_string())
}

#[cfg(not(feature = "async"))]
pub fn sync_function() -> Result<String, Error> {
    // Sync implementation
    Ok("sync result".to_string())
}
```

## Publishing to Crates.io

### Pre-Publication Checklist
```bash
# 1. Ensure package builds cleanly
cargo build --release

# 2. Run all tests
cargo test --all-features
cargo test --no-default-features

# 3. Check documentation builds
cargo doc --no-deps --all-features

# 4. Run clippy for linting
cargo clippy --all-targets --all-features -- -D warnings

# 5. Check formatting
cargo fmt --all -- --check

# 6. Dry run publish
cargo publish --dry-run

# 7. Check package contents
cargo package --list
```

### Package Validation
```toml
# Cargo.toml validation checklist
[package]
# Required fields
name = "unique-crate-name"      # Must be unique on crates.io
version = "0.1.0"               # Valid semver
edition = "2021"                # Rust edition

# Highly recommended
description = "Clear, concise description under 300 characters"
license = "MIT OR Apache-2.0"   # Standard Rust licenses
repository = "https://github.com/user/repo"
documentation = "https://docs.rs/crate-name"

# Important metadata
keywords = ["web", "http", "client"]  # Max 5, lowercase, ASCII
categories = ["web-programming::http-client"]  # From crates.io categories
readme = "README.md"
authors = ["Name <email@domain.com>"]

# Control what gets packaged
exclude = [
    "tests/fixtures/*",
    "benches/large_datasets/*",
    ".github/*",
    "docs/internal/*"
]
```

### Publishing Process
```bash
# 1. Login to crates.io (one-time setup)
cargo login <your-api-token>

# 2. Package the crate
cargo package

# 3. Verify package contents
tar -tzf target/package/my-crate-0.1.0.crate

# 4. Dry run publish
cargo publish --dry-run

# 5. Publish to crates.io
cargo publish

# 6. Verify publication
cargo search my-crate
```

### Post-Publication Management
```bash
# Yank a version (doesn't delete, prevents new dependents)
cargo yank --vers 0.1.1

# Un-yank a version
cargo yank --vers 0.1.1 --undo

# Add owners
cargo owner --add github-username crate-name
cargo owner --add user@email.com crate-name

# List owners
cargo owner --list crate-name

# Remove owners
cargo owner --remove github-username crate-name
```

## Workspaces for Multi-Crate Projects

### Workspace Configuration
```toml
# Root workspace Cargo.toml
[workspace]
members = [
    "core",
    "cli",
    "web-server",
    "utils/*",         # Glob pattern
    "plugins/plugin-*" # Multiple plugins
]
exclude = [
    "third-party/*",
    "experimental/*"
]

# Shared dependency versions across workspace
[workspace.dependencies]
serde = "1.0"
tokio = "1.0"
anyhow = "1.0"
thiserror = "1.0"

# Shared metadata
[workspace.package]
version = "0.1.0"
edition = "2021"
license = "MIT OR Apache-2.0"
repository = "https://github.com/user/mono-repo"
authors = ["Team Name <team@company.com>"]

# Workspace-level profiles
[profile.release]
lto = true
codegen-units = 1
```

### Individual Crate in Workspace
```toml
# core/Cargo.toml
[package]
name = "my-project-core"
version.workspace = true
edition.workspace = true
license.workspace = true
repository.workspace = true

[dependencies]
# Use workspace dependencies
serde.workspace = true
tokio = { workspace = true, features = ["full"] }

# Local workspace dependencies
my-project-utils = { path = "../utils" }

# External dependencies not in workspace
uuid = "1.0"
```

### Workspace Commands
```bash
# Build entire workspace
cargo build

# Build specific package
cargo build -p my-project-core
cargo build -p my-project-cli

# Test entire workspace
cargo test

# Test specific package
cargo test -p my-project-core

# Check all packages
cargo check --workspace

# Update workspace dependencies
cargo update

# Publish all packages (in dependency order)
cargo ws publish

# Generate unified documentation
cargo doc --workspace --no-deps
```

## Alternative Registries and Private Packages

### Private Registry Setup
```toml
# .cargo/config.toml
[registries]
my-company = { index = "https://my-company.com/git/crate-index" }

[source.my-company]
registry = "https://my-company.com/git/crate-index"

# Authentication
[registries.my-company]
index = "https://my-company.com/git/crate-index"
token = "your-private-registry-token"
```

### Using Private Registry
```toml
# Cargo.toml
[dependencies]
# From default registry (crates.io)
serde = "1.0"

# From private registry
internal-lib = { version = "0.1", registry = "my-company" }

# Git dependencies
git-dep = { git = "https://github.com/company/repo", tag = "v0.1.0" }
git-dep-branch = { git = "https://github.com/company/repo", branch = "main" }
git-dep-rev = { git = "https://github.com/company/repo", rev = "abc123" }

# Path dependencies (local development)
local-lib = { path = "../local-lib" }
```

### Publishing to Private Registry
```bash
# Configure registry
cargo login --registry=my-company <token>

# Publish to private registry
cargo publish --registry=my-company

# List from private registry
cargo search --registry=my-company query
```

## Binary Distribution

### Cross-Platform Binary Builds
```toml
# Cargo.toml for binary distribution
[package]
name = "my-cli-tool"
version = "1.0.0"
edition = "2021"

[[bin]]
name = "my-tool"
path = "src/main.rs"

[profile.release]
lto = true
codegen-units = 1
panic = "abort"
strip = true  # Remove debug symbols
opt-level = "z"  # Optimize for size
```

```bash
# Install cross-compilation targets
rustup target add x86_64-pc-windows-gnu
rustup target add x86_64-unknown-linux-musl
rustup target add aarch64-apple-darwin

# Cross-compile binaries
cargo build --release --target x86_64-pc-windows-gnu
cargo build --release --target x86_64-unknown-linux-musl
cargo build --release --target aarch64-apple-darwin

# Using cross for complex cross-compilation
cargo install cross
cross build --release --target x86_64-pc-windows-gnu
```

### Binary Installation Methods
```bash
# Install from crates.io
cargo install my-cli-tool

# Install specific version
cargo install my-cli-tool --version 1.2.3

# Install from git
cargo install --git https://github.com/user/repo

# Install from local path
cargo install --path .

# Install with specific features
cargo install my-cli-tool --features "feature1,feature2"

# Uninstall
cargo uninstall my-cli-tool

# List installed binaries
cargo install --list
```

### Distribution via Package Managers

#### Homebrew Formula
```ruby
# my-tool.rb
class MyTool < Formula
  desc "Description of my Rust CLI tool"
  homepage "https://github.com/user/my-tool"
  url "https://github.com/user/my-tool/archive/v1.0.0.tar.gz"
  sha256 "sha256-hash-here"
  license "MIT"

  depends_on "rust" => :build

  def install
    system "cargo", "install", *std_cargo_args
  end

  test do
    assert_match "my-tool", shell_output("#{bin}/my-tool --version")
  end
end
```

#### Arch Linux PKGBUILD
```bash
# PKGBUILD
pkgname=my-tool
pkgver=1.0.0
pkgrel=1
pkgdesc="Description of my Rust CLI tool"
arch=('x86_64')
url="https://github.com/user/my-tool"
license=('MIT')
depends=()
makedepends=('rust')
source=("$pkgname-$pkgver.tar.gz::https://github.com/user/my-tool/archive/v$pkgver.tar.gz")
sha256sums=('SKIP')

build() {
    cd "$pkgname-$pkgver"
    cargo build --release --locked
}

package() {
    cd "$pkgname-$pkgver"
    install -Dm755 "target/release/$pkgname" "$pkgdir/usr/bin/$pkgname"
}
```

#### Debian Package
```bash
# Create debian package structure
mkdir -p my-tool-1.0.0/DEBIAN
mkdir -p my-tool-1.0.0/usr/bin

# Create control file
cat > my-tool-1.0.0/DEBIAN/control << EOF
Package: my-tool
Version: 1.0.0
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Your Name <email@example.com>
Description: Description of my Rust CLI tool
EOF

# Copy binary and build package
cp target/release/my-tool my-tool-1.0.0/usr/bin/
dpkg-deb --build my-tool-1.0.0
```

## Documentation and README

### README.md Template
```markdown
# My Awesome Rust Crate

[![Crates.io](https://img.shields.io/crates/v/my-crate.svg)](https://crates.io/crates/my-crate)
[![Documentation](https://docs.rs/my-crate/badge.svg)](https://docs.rs/my-crate)
[![Build Status](https://github.com/user/my-crate/workflows/CI/badge.svg)](https://github.com/user/my-crate/actions)
[![License](https://img.shields.io/crates/l/my-crate.svg)](LICENSE)

Brief description of what your crate does.

## Features

- Feature 1
- Feature 2
- Feature 3

## Installation

### As a library
```toml
[dependencies]
my-crate = "1.0"
```

### As a binary
```bash
cargo install my-crate
```

## Usage

### Basic Example
```rust
use my_crate::MyStruct;

let instance = MyStruct::new();
let result = instance.do_something();
println!("Result: {}", result);
```

### Advanced Example
```rust
use my_crate::{MyStruct, Config};

let config = Config::builder()
    .option1(true)
    .option2("value")
    .build();

let instance = MyStruct::with_config(config);
let result = instance.advanced_operation()?;
```

## Feature Flags

- `default`: Standard functionality
- `serde`: Serialization support
- `async`: Async/await support
- `cli`: Command-line interface

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

Licensed under either of

- Apache License, Version 2.0 ([LICENSE-APACHE](LICENSE-APACHE))
- MIT license ([LICENSE-MIT](LICENSE-MIT))

at your option.
```

### Cargo Documentation
```rust
//! # My Crate
//!
//! This crate provides functionality for...
//!
//! ## Examples
//!
//! ```
//! use my_crate::MyStruct;
//!
//! let instance = MyStruct::new();
//! assert_eq!(instance.value(), 42);
//! ```

/// A struct that does something useful
///
/// # Examples
///
/// ```
/// use my_crate::MyStruct;
///
/// let instance = MyStruct::new();
/// println!("Value: {}", instance.value());
/// ```
pub struct MyStruct {
    value: i32,
}

impl MyStruct {
    /// Creates a new instance
    ///
    /// # Examples
    ///
    /// ```
    /// use my_crate::MyStruct;
    ///
    /// let instance = MyStruct::new();
    /// assert_eq!(instance.value(), 42);
    /// ```
    pub fn new() -> Self {
        Self { value: 42 }
    }

    /// Returns the internal value
    ///
    /// # Examples
    ///
    /// ```
    /// use my_crate::MyStruct;
    ///
    /// let instance = MyStruct::new();
    /// assert_eq!(instance.value(), 42);
    /// ```
    pub fn value(&self) -> i32 {
        self.value
    }
}
```

## Security and Supply Chain

### Cargo Audit and Security
```bash
# Install cargo-audit
cargo install cargo-audit

# Check for known vulnerabilities
cargo audit

# Check for yanked crates
cargo audit --stale

# Generate security report
cargo audit --json > security-report.json
```

### Supply Chain Verification
```bash
# Install cargo-supply-chain
cargo install cargo-supply-chain

# Analyze dependency authors and publishers
cargo supply-chain

# Check for suspicious patterns
cargo supply-chain --json | jq '.crates[] | select(.publishers | length == 1)'
```

### Dependency Policies with cargo-deny
```toml
# deny.toml
[graph]
targets = [
    { triple = "x86_64-unknown-linux-gnu" },
    { triple = "x86_64-pc-windows-msvc" },
    { triple = "x86_64-apple-darwin" },
]

[advisories]
version = 2
ignore = []
informational = "warn"
severity-threshold = "low"

[licenses]
version = 2
allow = [
    "MIT",
    "Apache-2.0",
    "Apache-2.0 WITH LLVM-exception",
    "BSD-2-Clause",
    "BSD-3-Clause",
    "ISC",
    "Unicode-DFS-2016",
]
deny = [
    "GPL-2.0",
    "GPL-3.0",
    "AGPL-1.0",
    "AGPL-3.0",
]

[bans]
multiple-versions = "warn"
wildcards = "allow"
highlight = "all"
workspace-default-features = "allow"
external-default-features = "allow"
skip = []
skip-tree = []

deny = [
    { name = "openssl", version = "*", reason = "Use rustls instead" },
]

[sources]
unknown-registry = "warn"
unknown-git = "warn"
allow-registry = ["https://github.com/rust-lang/crates.io-index"]
allow-git = []
```

## Best Practices Summary

### Package Design
1. **Single Responsibility**: Each crate should have a clear, focused purpose
2. **Semantic Versioning**: Follow SemVer strictly for public APIs
3. **Feature Flags**: Use features to make crates modular and lightweight
4. **Documentation**: Comprehensive docs with examples for all public APIs
5. **Testing**: Unit tests, integration tests, and doc tests

### Publication Strategy
1. **Quality Gates**: Automated testing, linting, and formatting checks
2. **Release Automation**: Use tools like `cargo-release` for consistent releases
3. **Changelog**: Maintain detailed changelogs following Keep a Changelog
4. **Deprecation**: Proper deprecation warnings before removing APIs
5. **Security**: Regular audits and prompt response to vulnerabilities

### Ecosystem Participation
1. **Standard Licenses**: Use MIT OR Apache-2.0 for broad compatibility
2. **Naming Conventions**: Follow Rust naming conventions and crates.io guidelines
3. **Categories and Keywords**: Proper categorization for discoverability
4. **Community**: Respond to issues, accept contributions, maintain actively
5. **Integration**: Work well with common Rust ecosystem tools and patterns

This comprehensive packaging guide ensures Rust crates are properly structured, documented, secured, and distributed following ecosystem best practices.