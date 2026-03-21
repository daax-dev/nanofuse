# C GitHub Actions CI/CD Build Pipeline

## Basic GitHub Actions Workflow

### Simple C Build Workflow
```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Install dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y build-essential

    - name: Build
      run: |
        gcc -Wall -Wextra -O2 src/*.c -o program

    - name: Test
      run: |
        ./program --test

    - name: Upload artifacts
      uses: actions/upload-artifact@v4
      with:
        name: program-linux
        path: program
```

### Multi-Platform Build Matrix
```yaml
name: Cross Platform Build

on: [push, pull_request]

jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        compiler: [gcc, clang]
        exclude:
          - os: windows-latest
            compiler: clang
        include:
          - os: windows-latest
            compiler: msvc

    runs-on: ${{ matrix.os }}

    steps:
    - uses: actions/checkout@v4

    - name: Setup MSVC (Windows)
      if: matrix.os == 'windows-latest' && matrix.compiler == 'msvc'
      uses: microsoft/setup-msbuild@v2

    - name: Install dependencies (Linux)
      if: matrix.os == 'ubuntu-latest'
      run: |
        sudo apt-get update
        sudo apt-get install -y build-essential
        if [ "${{ matrix.compiler }}" = "clang" ]; then
          sudo apt-get install -y clang
        fi

    - name: Install dependencies (macOS)
      if: matrix.os == 'macos-latest'
      run: |
        if [ "${{ matrix.compiler }}" = "clang" ]; then
          brew install llvm
        fi

    - name: Build (Unix)
      if: matrix.os != 'windows-latest'
      run: |
        export CC=${{ matrix.compiler }}
        make

    - name: Build (Windows MSVC)
      if: matrix.os == 'windows-latest' && matrix.compiler == 'msvc'
      run: |
        cl /Fe:program.exe src/*.c

    - name: Test
      run: |
        if [ "${{ matrix.os }}" = "windows-latest" ]; then
          ./program.exe --test
        else
          ./program --test
        fi
      shell: bash

    - name: Upload artifacts
      uses: actions/upload-artifact@v4
      with:
        name: program-${{ matrix.os }}-${{ matrix.compiler }}
        path: program*
```

## Advanced CMake Integration

### CMake-Based Workflow
```yaml
name: CMake Build

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

env:
  BUILD_TYPE: Release

jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu-20.04, ubuntu-22.04, windows-latest, macos-latest]
        build_type: [Debug, Release]
        c_compiler: [gcc, clang]
        include:
          - os: windows-latest
            c_compiler: cl
            cpp_compiler: cl
          - os: ubuntu-20.04
            c_compiler: gcc
            cpp_compiler: g++
          - os: ubuntu-22.04
            c_compiler: clang
            cpp_compiler: clang++
          - os: macos-latest
            c_compiler: clang
            cpp_compiler: clang++
        exclude:
          - os: windows-latest
            c_compiler: gcc
          - os: windows-latest
            c_compiler: clang
          - os: macos-latest
            c_compiler: gcc

    runs-on: ${{ matrix.os }}

    steps:
    - uses: actions/checkout@v4

    - name: Set reusable strings
      id: strings
      shell: bash
      run: |
        echo "build-output-dir=${{ github.workspace }}/build" >> "$GITHUB_OUTPUT"

    - name: Install dependencies (Linux)
      if: runner.os == 'Linux'
      run: |
        sudo apt-get update
        sudo apt-get install -y cmake ninja-build
        if [ "${{ matrix.c_compiler }}" = "clang" ]; then
          sudo apt-get install -y clang
        fi

    - name: Install dependencies (macOS)
      if: runner.os == 'macOS'
      run: |
        brew install cmake ninja

    - name: Configure CMake
      run: >
        cmake -B ${{ steps.strings.outputs.build-output-dir }}
        -DCMAKE_CXX_COMPILER=${{ matrix.cpp_compiler }}
        -DCMAKE_C_COMPILER=${{ matrix.c_compiler }}
        -DCMAKE_BUILD_TYPE=${{ matrix.build_type }}
        -G Ninja
        -S ${{ github.workspace }}

    - name: Build
      run: cmake --build ${{ steps.strings.outputs.build-output-dir }} --config ${{ matrix.build_type }}

    - name: Test
      working-directory: ${{ steps.strings.outputs.build-output-dir }}
      run: ctest --build-config ${{ matrix.build_type }} --verbose

    - name: Install
      run: cmake --install ${{ steps.strings.outputs.build-output-dir }} --prefix ${{ github.workspace }}/install

    - name: Upload build artifacts
      uses: actions/upload-artifact@v4
      with:
        name: build-${{ matrix.os }}-${{ matrix.c_compiler }}-${{ matrix.build_type }}
        path: ${{ github.workspace }}/install
```

## Static Analysis Integration

### Comprehensive Analysis Workflow
```yaml
name: Static Analysis

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  static-analysis:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Install analysis tools
      run: |
        sudo apt-get update
        sudo apt-get install -y clang clang-tools cppcheck

    - name: Clang Static Analyzer
      run: |
        mkdir -p analysis-results/clang-analyzer
        clang --analyze -Xanalyzer -analyzer-output=html \
              -Xanalyzer -analyzer-output-dir=analysis-results/clang-analyzer \
              src/*.c

    - name: Clang-Tidy
      run: |
        clang-tidy src/*.c -- -Iinclude > clang-tidy-results.txt

    - name: Cppcheck
      run: |
        cppcheck --enable=all --xml --xml-version=2 src/ 2> cppcheck-results.xml

    - name: Upload analysis results
      uses: actions/upload-artifact@v4
      with:
        name: static-analysis-results
        path: |
          analysis-results/
          clang-tidy-results.txt
          cppcheck-results.xml

  security-scan:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Run CodeQL Analysis
      uses: github/codeql-action/init@v3
      with:
        languages: c

    - name: Build for CodeQL
      run: |
        gcc -c src/*.c

    - name: Perform CodeQL Analysis
      uses: github/codeql-action/analyze@v3
```

## Sanitizer Testing

### Sanitizer Matrix
```yaml
name: Sanitizer Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  sanitizer-tests:
    strategy:
      matrix:
        sanitizer: [address, undefined, memory, thread]
        compiler: [gcc, clang]
        exclude:
          - sanitizer: memory
            compiler: gcc

    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Install dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y build-essential
        if [ "${{ matrix.compiler }}" = "clang" ]; then
          sudo apt-get install -y clang
        fi

    - name: Build with sanitizer
      run: |
        export CC=${{ matrix.compiler }}
        export CFLAGS="-fsanitize=${{ matrix.sanitizer }} -g -fno-omit-frame-pointer"
        if [ "${{ matrix.sanitizer }}" = "thread" ]; then
          export CFLAGS="$CFLAGS -pthread"
        fi
        make clean && make

    - name: Run tests
      run: |
        export ASAN_OPTIONS=detect_leaks=1:abort_on_error=1
        export UBSAN_OPTIONS=print_stacktrace=1:abort_on_error=1
        export MSAN_OPTIONS=abort_on_error=1
        export TSAN_OPTIONS=abort_on_error=1
        ./run_tests.sh

    - name: Upload sanitizer reports
      if: failure()
      uses: actions/upload-artifact@v4
      with:
        name: sanitizer-${{ matrix.sanitizer }}-${{ matrix.compiler }}-report
        path: |
          *.log
          core*
```

## Code Coverage

### Coverage Workflow
```yaml
name: Code Coverage

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  coverage:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Install dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y gcc gcov lcov

    - name: Build with coverage
      run: |
        export CFLAGS="--coverage -g -O0"
        export LDFLAGS="--coverage"
        make clean && make

    - name: Run tests
      run: |
        ./run_tests.sh

    - name: Generate coverage report
      run: |
        gcov src/*.c
        lcov --capture --directory . --output-file coverage.info
        lcov --remove coverage.info '/usr/*' --output-file coverage.info
        lcov --list coverage.info

    - name: Upload to Codecov
      uses: codecov/codecov-action@v4
      with:
        files: ./coverage.info
        flags: unittests
        name: codecov-umbrella

    - name: Generate HTML report
      run: |
        genhtml coverage.info --output-directory coverage_html

    - name: Upload coverage HTML
      uses: actions/upload-artifact@v4
      with:
        name: coverage-report
        path: coverage_html/
```

## Release and Deployment

### Automated Release Workflow
```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build-release:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        arch: [x64, arm64]
        exclude:
          - os: windows-latest
            arch: arm64

    runs-on: ${{ matrix.os }}

    steps:
    - uses: actions/checkout@v4

    - name: Setup build environment
      run: |
        if [ "${{ runner.os }}" = "Linux" ]; then
          sudo apt-get update
          sudo apt-get install -y build-essential
          if [ "${{ matrix.arch }}" = "arm64" ]; then
            sudo apt-get install -y gcc-aarch64-linux-gnu
          fi
        elif [ "${{ runner.os }}" = "macOS" ]; then
          if [ "${{ matrix.arch }}" = "arm64" ]; then
            export CFLAGS="-target arm64-apple-macos11"
          fi
        fi
      shell: bash

    - name: Build release
      run: |
        if [ "${{ runner.os }}" = "Linux" ] && [ "${{ matrix.arch }}" = "arm64" ]; then
          export CC=aarch64-linux-gnu-gcc
        fi
        export CFLAGS="-O3 -DNDEBUG"
        make clean && make

        # Create release archive
        if [ "${{ runner.os }}" = "Windows" ]; then
          7z a program-${{ matrix.os }}-${{ matrix.arch }}.zip program.exe
        else
          tar -czf program-${{ matrix.os }}-${{ matrix.arch }}.tar.gz program
        fi
      shell: bash

    - name: Upload release artifacts
      uses: actions/upload-artifact@v4
      with:
        name: release-${{ matrix.os }}-${{ matrix.arch }}
        path: program-${{ matrix.os }}-${{ matrix.arch }}.*

  create-release:
    needs: build-release
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Download all artifacts
      uses: actions/download-artifact@v4
      with:
        path: release-artifacts

    - name: Create Release
      uses: actions/create-release@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        tag_name: ${{ github.ref }}
        release_name: Release ${{ github.ref }}
        body: |
          Changes in this Release
          - Feature 1
          - Feature 2
          - Bug fixes
        draft: false
        prerelease: false

    - name: Upload Release Assets
      uses: actions/upload-release-asset@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        upload_url: ${{ steps.create_release.outputs.upload_url }}
        asset_path: release-artifacts/**/*
        asset_name: ${{ matrix.os }}-${{ matrix.arch }}
        asset_content_type: application/octet-stream
```

## Docker Integration

### Docker Build Workflow
```yaml
name: Docker Build

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
  docker-build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
    - uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Log in to Container Registry
      if: github.event_name != 'pull_request'
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        push: ${{ github.event_name != 'pull_request' }}
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        platforms: linux/amd64,linux/arm64
```

### Multi-stage Dockerfile
```dockerfile
# Dockerfile
FROM gcc:latest as builder

WORKDIR /app
COPY . .

RUN make clean && \
    make CFLAGS="-O3 -static" && \
    strip program

FROM scratch
COPY --from=builder /app/program /program
ENTRYPOINT ["/program"]
```

## Performance Testing

### Benchmark Workflow
```yaml
name: Performance Tests

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

    - name: Install dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y build-essential valgrind

    - name: Build optimized version
      run: |
        export CFLAGS="-O3 -DNDEBUG"
        make clean && make

    - name: Run benchmarks
      run: |
        ./benchmarks/run_benchmarks.sh > benchmark_results.txt

    - name: Memory profiling
      run: |
        valgrind --tool=massif --massif-out-file=massif.out ./program
        ms_print massif.out > memory_profile.txt

    - name: Performance profiling
      run: |
        valgrind --tool=callgrind --callgrind-out-file=callgrind.out ./program
        callgrind_annotate callgrind.out > performance_profile.txt

    - name: Upload benchmark results
      uses: actions/upload-artifact@v4
      with:
        name: performance-results
        path: |
          benchmark_results.txt
          memory_profile.txt
          performance_profile.txt
          massif.out
          callgrind.out
```

## Documentation Generation

### Documentation Workflow
```yaml
name: Documentation

on:
  push:
    branches: [ main ]

jobs:
  docs:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Install Doxygen
      run: |
        sudo apt-get update
        sudo apt-get install -y doxygen graphviz

    - name: Generate documentation
      run: |
        doxygen Doxyfile

    - name: Deploy to GitHub Pages
      uses: peaceiris/actions-gh-pages@v3
      if: github.ref == 'refs/heads/main'
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
        publish_dir: ./docs/html
```

## Best Practices Summary

### Workflow Organization
1. **Separate workflows** for different purposes (CI, release, docs)
2. **Use build matrices** for cross-platform testing
3. **Cache dependencies** to speed up builds
4. **Fail fast** on critical errors
5. **Parallel job execution** where possible

### Security Practices
1. **Use official actions** from trusted sources
2. **Pin action versions** to specific commits
3. **Minimize token permissions**
4. **Scan for vulnerabilities** regularly
5. **Never commit secrets** to repository

### Performance Optimization
1. **Cache build artifacts** between runs
2. **Use appropriate runners** for workload
3. **Optimize Docker layers** for faster builds
4. **Parallel test execution** when possible
5. **Skip unnecessary steps** with conditions

This comprehensive GitHub Actions setup provides robust CI/CD for C projects with cross-platform builds, extensive testing, and automated deployment capabilities.