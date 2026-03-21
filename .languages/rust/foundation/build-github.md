# GitHub Actions Build Pipeline for Rust

## Basic CI/CD Workflow

### Comprehensive Rust CI Pipeline
```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  CARGO_TERM_COLOR: always
  RUST_BACKTRACE: 1

jobs:
  test:
    name: Test Suite
    runs-on: ubuntu-latest
    strategy:
      matrix:
        rust: [stable, beta, nightly]
        include:
          - rust: stable
            can_fail: false
          - rust: beta
            can_fail: false
          - rust: nightly
            can_fail: true
    continue-on-error: ${{ matrix.can_fail }}

    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

    - name: Install Rust toolchain
      uses: dtolnay/rust-toolchain@master
      with:
        toolchain: ${{ matrix.rust }}
        components: rustfmt, clippy

    - name: Configure Rust caching
      uses: Swatinem/rust-cache@v2
      with:
        cache-on-failure: true

    - name: Check formatting
      run: cargo fmt --all -- --check

    - name: Run Clippy
      run: cargo clippy --all-targets --all-features -- -D warnings

    - name: Run tests
      run: cargo test --all-features --verbose

    - name: Run doc tests
      run: cargo test --doc

    - name: Check docs can be built
      run: cargo doc --no-deps --all-features

  coverage:
    name: Code Coverage
    runs-on: ubuntu-latest
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

    - name: Install Rust toolchain
      uses: dtolnay/rust-toolchain@stable

    - name: Install cargo-llvm-cov
      uses: taiki-e/install-action@cargo-llvm-cov

    - name: Configure Rust caching
      uses: Swatinem/rust-cache@v2

    - name: Generate code coverage
      run: cargo llvm-cov --all-features --workspace --lcov --output-path lcov.info

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        files: lcov.info
        fail_ci_if_error: true

  security:
    name: Security Audit
    runs-on: ubuntu-latest
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

    - name: Install Rust toolchain
      uses: dtolnay/rust-toolchain@stable

    - name: Install cargo-audit
      run: cargo install cargo-audit

    - name: Run security audit
      run: cargo audit

    - name: Run cargo-deny
      uses: EmbarkStudios/cargo-deny-action@v1

  cross-platform:
    name: Cross Platform Build
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        rust: [stable]

    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

    - name: Install Rust toolchain
      uses: dtolnay/rust-toolchain@master
      with:
        toolchain: ${{ matrix.rust }}

    - name: Configure Rust caching
      uses: Swatinem/rust-cache@v2

    - name: Build
      run: cargo build --release

    - name: Run tests
      run: cargo test --release
```

### Release Workflow with Artifacts
```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

env:
  CARGO_TERM_COLOR: always

jobs:
  create-release:
    name: Create Release
    runs-on: ubuntu-latest
    outputs:
      upload_url: ${{ steps.release.outputs.upload_url }}
      release_id: ${{ steps.release.outputs.id }}
    steps:
      - name: Create Release
        id: release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: ${{ github.ref }}
          release_name: Release ${{ github.ref }}
          draft: false
          prerelease: false

  build-release:
    name: Build Release
    needs: create-release
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            target: x86_64-unknown-linux-gnu
            asset_name: myapp-linux-x86_64
          - os: ubuntu-latest
            target: x86_64-unknown-linux-musl
            asset_name: myapp-linux-x86_64-musl
          - os: windows-latest
            target: x86_64-pc-windows-msvc
            asset_name: myapp-windows-x86_64.exe
          - os: macos-latest
            target: x86_64-apple-darwin
            asset_name: myapp-macos-x86_64
          - os: macos-latest
            target: aarch64-apple-darwin
            asset_name: myapp-macos-aarch64

    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

    - name: Install Rust toolchain
      uses: dtolnay/rust-toolchain@stable
      with:
        targets: ${{ matrix.target }}

    - name: Configure Rust caching
      uses: Swatinem/rust-cache@v2
      with:
        key: ${{ matrix.target }}

    - name: Build release binary
      run: cargo build --release --target ${{ matrix.target }}

    - name: Strip binary (Linux/macOS)
      if: matrix.os != 'windows-latest'
      run: strip target/${{ matrix.target }}/release/myapp

    - name: Upload Release Asset
      uses: actions/upload-release-asset@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        upload_url: ${{ needs.create-release.outputs.upload_url }}
        asset_path: target/${{ matrix.target }}/release/myapp${{ matrix.os == 'windows-latest' && '.exe' || '' }}
        asset_name: ${{ matrix.asset_name }}
        asset_content_type: application/octet-stream

  publish-crate:
    name: Publish to crates.io
    runs-on: ubuntu-latest
    needs: create-release
    if: startsWith(github.ref, 'refs/tags/')
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

    - name: Install Rust toolchain
      uses: dtolnay/rust-toolchain@stable

    - name: Configure Rust caching
      uses: Swatinem/rust-cache@v2

    - name: Publish to crates.io
      run: cargo publish --token ${{ secrets.CRATES_IO_TOKEN }}
```

### Docker Build and Publish
```yaml
# .github/workflows/docker.yml
name: Docker

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

    - name: Log in to Container Registry
      if: github.event_name != 'pull_request'
      uses: docker/login-action@v2
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v4
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}

    - name: Build and push Docker image
      uses: docker/build-push-action@v4
      with:
        context: .
        push: ${{ github.event_name != 'pull_request' }}
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
```

### Dockerfile for Rust Applications
```dockerfile
# Multi-stage Dockerfile for Rust
FROM rust:1.75 as builder

# Create a new empty shell project
RUN USER=root cargo new --bin myapp
WORKDIR /myapp

# Copy our manifests
COPY ./Cargo.lock ./Cargo.lock
COPY ./Cargo.toml ./Cargo.toml

# Build only the dependencies to cache them
RUN cargo build --release
RUN rm src/*.rs

# Copy the source code
COPY ./src ./src

# Build for release
RUN rm ./target/release/deps/myapp*
RUN cargo build --release

# Final stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \\
    ca-certificates \\
    && rm -rf /var/lib/apt/lists/*

# Copy the binary from builder stage
COPY --from=builder /myapp/target/release/myapp /usr/local/bin/myapp

# Create non-root user
RUN useradd -r -s /bin/false myapp
USER myapp

EXPOSE 8080
CMD [\"myapp\"]
```

### Optimized Dockerfile (Alpine-based)
```dockerfile
# Alpine-based multi-stage build
FROM rust:1.75-alpine as builder

RUN apk add --no-cache musl-dev

WORKDIR /app
COPY . .
RUN cargo build --release --target x86_64-unknown-linux-musl

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/target/x86_64-unknown-linux-musl/release/myapp .
CMD [\"./myapp\"]
```

## Advanced CI/CD Patterns

### Matrix Testing with Features
```yaml
# .github/workflows/feature-matrix.yml
name: Feature Matrix Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        rust: [stable, nightly]
        features:
          - default
          - --no-default-features
          - --no-default-features --features minimal
          - --all-features

    steps:
    - uses: actions/checkout@v4
    - uses: dtolnay/rust-toolchain@master
      with:
        toolchain: ${{ matrix.rust }}

    - name: Test with features
      run: cargo test ${{ matrix.features }}
```

### Conditional Deployments
```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]

jobs:
  deploy-staging:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    environment: staging
    steps:
    - name: Deploy to staging
      run: |
        echo \"Deploying to staging environment\"
        # Deployment commands here

  deploy-production:
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')
    environment: production
    steps:
    - name: Deploy to production
      run: |
        echo \"Deploying to production environment\"
        # Production deployment commands here
```

### Dependency Update Automation
```yaml
# .github/workflows/dependency-update.yml
name: Dependency Update

on:
  schedule:
    - cron: '0 0 * * 0' # Weekly on Sunday
  workflow_dispatch: # Allow manual trigger

jobs:
  update-dependencies:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4
      with:
        token: ${{ secrets.GITHUB_TOKEN }}

    - name: Install Rust toolchain
      uses: dtolnay/rust-toolchain@stable

    - name: Update dependencies
      run: cargo update

    - name: Run tests
      run: cargo test

    - name: Create Pull Request
      uses: peter-evans/create-pull-request@v5
      with:
        token: ${{ secrets.GITHUB_TOKEN }}
        commit-message: Update dependencies
        title: 'chore: Update dependencies'
        body: |
          This PR updates dependencies to their latest versions.

          Changes:
          - Updated Cargo.lock with latest compatible versions

          Please review the changes and ensure all tests pass.
        branch: update-dependencies
        delete-branch: true
```

## Security and Quality Gates

### Comprehensive Security Checks
```yaml
# .github/workflows/security.yml
name: Security

on: [push, pull_request]

jobs:
  security-audit:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: dtolnay/rust-toolchain@stable

    - name: Security Audit
      uses: actions-rs/audit-check@v1
      with:
        token: ${{ secrets.GITHUB_TOKEN }}

  cargo-deny:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: EmbarkStudios/cargo-deny-action@v1

  supply-chain:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: dtolnay/rust-toolchain@stable

    - name: Install cargo-supply-chain
      run: cargo install cargo-supply-chain

    - name: Run supply chain analysis
      run: cargo supply-chain

  semver:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: dtolnay/rust-toolchain@stable

    - name: Check SemVer
      uses: obi1kenobi/cargo-semver-checks-action@v2
```

### Code Quality Gates
```yaml
# .github/workflows/quality.yml
name: Quality

on: [push, pull_request]

jobs:
  formatting:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: dtolnay/rust-toolchain@stable
      with:
        components: rustfmt
    - run: cargo fmt --all -- --check

  linting:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: dtolnay/rust-toolchain@stable
      with:
        components: clippy
    - run: cargo clippy --all-targets --all-features -- -D warnings

  documentation:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: dtolnay/rust-toolchain@stable
    - run: cargo doc --no-deps --all-features
    - run: cargo test --doc

  unused-dependencies:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: dtolnay/rust-toolchain@nightly
    - run: cargo install cargo-udeps
    - run: cargo +nightly udeps

  typos:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: crate-ci/typos@master
```

## Performance and Benchmarking

### Benchmark Regression Testing
```yaml
# .github/workflows/benchmark.yml
name: Benchmark

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: dtolnay/rust-toolchain@stable

    - name: Run benchmark
      run: cargo bench -- --output-format json > benchmark_results.json

    - name: Store benchmark result
      uses: benchmark-action/github-action-benchmark@v1
      with:
        tool: 'cargo'
        output-file-path: benchmark_results.json
        github-token: ${{ secrets.GITHUB_TOKEN }}
        auto-push: true
        comment-on-alert: true
        alert-threshold: '200%'
        fail-on-alert: true
```

### Binary Size Tracking
```yaml
# .github/workflows/size-tracking.yml
name: Binary Size

on: [push, pull_request]

jobs:
  size-tracking:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: dtolnay/rust-toolchain@stable

    - name: Build release
      run: cargo build --release

    - name: Install cargo-bloat
      run: cargo install cargo-bloat

    - name: Analyze binary size
      run: |
        cargo bloat --release --crates > bloat_report.txt
        ls -la target/release/ > binary_sizes.txt

    - name: Upload size reports
      uses: actions/upload-artifact@v3
      with:
        name: size-reports
        path: |
          bloat_report.txt
          binary_sizes.txt
```

## Secrets and Environment Management

### Environment-Specific Deployments
```yaml
# .github/workflows/environments.yml
name: Deploy to Environments

on:
  push:
    branches: [ main, develop ]

jobs:
  deploy-dev:
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    environment: development
    env:
      DATABASE_URL: ${{ secrets.DEV_DATABASE_URL }}
      API_KEY: ${{ secrets.DEV_API_KEY }}
    steps:
    - uses: actions/checkout@v4
    - name: Deploy to dev
      run: echo \"Deploying to development\"

  deploy-prod:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production
    env:
      DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}
      API_KEY: ${{ secrets.PROD_API_KEY }}
    steps:
    - uses: actions/checkout@v4
    - name: Deploy to production
      run: echo \"Deploying to production\"
```

### Cargo Configuration for CI
```toml
# .cargo/config.toml (for CI optimizations)
[build]
rustflags = [\"-D\", \"warnings\"]

[target.x86_64-unknown-linux-gnu]
rustflags = [\"-C\", \"link-arg=-fuse-ld=lld\"]

[env]
RUST_BACKTRACE = \"1\"
```

### Deny Configuration
```toml
# deny.toml
[graph]
targets = [
    { triple = \"x86_64-unknown-linux-gnu\" },
    { triple = \"x86_64-pc-windows-msvc\" },
    { triple = \"x86_64-apple-darwin\" },
]

[advisories]
version = 2
ignore = []

[licenses]
version = 2
allow = [
    \"MIT\",
    \"Apache-2.0\",
    \"Apache-2.0 WITH LLVM-exception\",
    \"BSD-2-Clause\",
    \"BSD-3-Clause\",
    \"ISC\",
    \"Unicode-DFS-2016\",
]

[bans]
multiple-versions = \"deny\"
deny = [
    { name = \"openssl\", version = \"*\" },
]

[sources]
unknown-registry = \"deny\"
unknown-git = \"deny\"
```

This comprehensive GitHub Actions setup provides robust CI/CD pipelines for Rust projects with security, quality, and performance gates ensuring reliable software delivery.