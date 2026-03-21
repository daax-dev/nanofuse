# C++ GitHub Actions CI/CD Pipeline

## Overview

Modern C++ projects require robust continuous integration and deployment pipelines. This guide covers GitHub Actions workflows for building, testing, static analysis, packaging, and deployment of C++ projects across multiple platforms.

## Basic CI/CD Workflow

### Multi-Platform Build Matrix

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop, 'feature/*' ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main, develop ]
  schedule:
    - cron: '0 2 * * 0'  # Weekly build on Sunday 2 AM

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

env:
  CMAKE_VERSION: 3.28.1
  CONAN_VERSION: 2.0.17

jobs:
  build-and-test:
    name: ${{ matrix.config.name }}
    runs-on: ${{ matrix.config.os }}

    strategy:
      fail-fast: false
      matrix:
        config:
          - {
              name: "Ubuntu 22.04 GCC 13",
              os: ubuntu-22.04,
              compiler: gcc,
              version: "13",
              build_type: Release,
              cc: "gcc-13",
              cxx: "g++-13",
              packages: "gcc-13 g++-13"
            }
          - {
              name: "Ubuntu 22.04 Clang 17",
              os: ubuntu-22.04,
              compiler: clang,
              version: "17",
              build_type: Release,
              cc: "clang-17",
              cxx: "clang++-17",
              packages: "clang-17 libc++-17-dev libc++abi-17-dev"
            }
          - {
              name: "Windows 2022 MSVC 19.3",
              os: windows-2022,
              compiler: msvc,
              version: "19.3",
              build_type: Release
            }
          - {
              name: "macOS 13 Apple Clang 14",
              os: macos-13,
              compiler: clang,
              version: "14",
              build_type: Release,
              cc: "clang",
              cxx: "clang++"
            }
          - {
              name: "Ubuntu 22.04 GCC 13 Debug",
              os: ubuntu-22.04,
              compiler: gcc,
              version: "13",
              build_type: Debug,
              cc: "gcc-13",
              cxx: "g++-13",
              packages: "gcc-13 g++-13",
              coverage: true
            }

    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        submodules: recursive
        fetch-depth: 0

    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: |
          ~/.conan2
          ~/.cache/vcpkg
          ${{ github.workspace }}/build/_deps
        key: ${{ matrix.config.os }}-${{ matrix.config.compiler }}-${{ matrix.config.version }}-${{ hashFiles('**/conanfile.py', '**/vcpkg.json') }}
        restore-keys: |
          ${{ matrix.config.os }}-${{ matrix.config.compiler }}-${{ matrix.config.version }}-

    - name: Setup Linux environment
      if: runner.os == 'Linux'
      run: |
        sudo apt-get update
        sudo apt-get install -y ninja-build ${{ matrix.config.packages }}

        # Install additional tools
        sudo apt-get install -y \
          valgrind \
          cppcheck \
          clang-tidy-17 \
          lcov \
          gcovr

        # Set environment variables
        echo "CC=${{ matrix.config.cc }}" >> $GITHUB_ENV
        echo "CXX=${{ matrix.config.cxx }}" >> $GITHUB_ENV

    - name: Setup Windows environment
      if: runner.os == 'Windows'
      run: |
        choco install ninja
        # Ensure MSVC is in PATH
        echo "C:\Program Files\Microsoft Visual Studio\2022\Enterprise\VC\Tools\MSVC\14.37.32822\bin\Hostx64\x64" >> $GITHUB_PATH

    - name: Setup macOS environment
      if: runner.os == 'macOS'
      run: |
        brew install ninja
        echo "CC=${{ matrix.config.cc }}" >> $GITHUB_ENV
        echo "CXX=${{ matrix.config.cxx }}" >> $GITHUB_ENV

    - name: Setup Python and Conan
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'

    - name: Install Conan
      run: |
        pip install conan==${{ env.CONAN_VERSION }}
        conan profile detect --force

    - name: Setup CMake
      uses: jwlawson/actions-setup-cmake@v1.14
      with:
        cmake-version: ${{ env.CMAKE_VERSION }}

    - name: Configure project
      run: |
        # Create Conan profile for this build
        conan profile show default

        # Install dependencies
        conan install . \
          --output-folder=build \
          --build=missing \
          --settings=build_type=${{ matrix.config.build_type }} \
          ${{ matrix.config.compiler == 'clang' && '--settings=compiler.libcxx=libc++' || '' }}

        # Configure CMake
        cmake -B build -S . -G Ninja \
          -DCMAKE_TOOLCHAIN_FILE=build/conan_toolchain.cmake \
          -DCMAKE_BUILD_TYPE=${{ matrix.config.build_type }} \
          -DBUILD_TESTS=ON \
          -DBUILD_BENCHMARKS=ON \
          -DENABLE_COVERAGE=${{ matrix.config.coverage == true && 'ON' || 'OFF' }} \
          -DCMAKE_EXPORT_COMPILE_COMMANDS=ON

    - name: Build project
      run: |
        cmake --build build --parallel $(nproc 2>/dev/null || sysctl -n hw.ncpu || echo 4)

    - name: Run tests
      run: |
        cd build
        ctest --output-on-failure --parallel $(nproc 2>/dev/null || sysctl -n hw.ncpu || echo 4)

    - name: Generate coverage report
      if: matrix.config.coverage == true
      run: |
        cd build
        gcovr --root .. --xml --xml-pretty --exclude-unreachable-branches --print-summary -o coverage.xml .
        lcov --capture --directory . --output-file coverage.info
        lcov --remove coverage.info '/usr/*' --output-file coverage.info
        lcov --remove coverage.info '*/test/*' --output-file coverage.info

    - name: Upload coverage to Codecov
      if: matrix.config.coverage == true
      uses: codecov/codecov-action@v3
      with:
        files: build/coverage.xml
        flags: unittests
        name: codecov-umbrella

    - name: Upload build artifacts
      if: matrix.config.build_type == 'Release'
      uses: actions/upload-artifact@v3
      with:
        name: ${{ matrix.config.name }}-artifacts
        path: |
          build/bin/
          build/lib/
        retention-days: 7
```

### Static Analysis Workflow

```yaml
# .github/workflows/static-analysis.yml
name: Static Analysis

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  clang-tidy:
    name: Clang-Tidy Analysis
    runs-on: ubuntu-22.04

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup environment
      run: |
        sudo apt-get update
        sudo apt-get install -y \
          clang-17 \
          clang-tidy-17 \
          libc++-17-dev \
          libc++abi-17-dev \
          ninja-build

    - name: Setup Python and Conan
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'

    - name: Install Conan
      run: |
        pip install conan
        conan profile detect --force

    - name: Install dependencies
      run: |
        conan install . --output-folder=build --build=missing

    - name: Configure project
      run: |
        cmake -B build -S . -G Ninja \
          -DCMAKE_TOOLCHAIN_FILE=build/conan_toolchain.cmake \
          -DCMAKE_BUILD_TYPE=Debug \
          -DCMAKE_CXX_CLANG_TIDY="clang-tidy-17;--config-file=.clang-tidy" \
          -DCMAKE_EXPORT_COMPILE_COMMANDS=ON

    - name: Run Clang-Tidy
      run: |
        cmake --build build --target all -- -k 0

    - name: Generate Clang-Tidy report
      run: |
        run-clang-tidy-17 \
          -p build \
          -header-filter='src/.*\.hpp$' \
          src/ > clang-tidy-report.txt || true

    - name: Upload Clang-Tidy report
      uses: actions/upload-artifact@v3
      with:
        name: clang-tidy-report
        path: clang-tidy-report.txt

  cppcheck:
    name: Cppcheck Analysis
    runs-on: ubuntu-22.04

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Cppcheck
      run: |
        sudo apt-get update
        sudo apt-get install -y cppcheck

    - name: Run Cppcheck
      run: |
        cppcheck \
          --enable=all \
          --std=c++20 \
          --platform=unix64 \
          --suppress=missingIncludeSystem \
          --suppress=unusedFunction \
          --inline-suppr \
          --xml \
          --xml-version=2 \
          --output-file=cppcheck-report.xml \
          src/ || true

    - name: Upload Cppcheck report
      uses: actions/upload-artifact@v3
      with:
        name: cppcheck-report
        path: cppcheck-report.xml

  sonarcloud:
    name: SonarCloud Analysis
    runs-on: ubuntu-22.04

    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Setup environment
      run: |
        sudo apt-get update
        sudo apt-get install -y \
          gcc-13 \
          g++-13 \
          ninja-build \
          gcovr \
          lcov

    - name: Setup Python and Conan
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'

    - name: Install Conan
      run: |
        pip install conan
        conan profile detect --force

    - name: Configure and build with coverage
      run: |
        conan install . --output-folder=build --build=missing

        cmake -B build -S . -G Ninja \
          -DCMAKE_TOOLCHAIN_FILE=build/conan_toolchain.cmake \
          -DCMAKE_BUILD_TYPE=Debug \
          -DENABLE_COVERAGE=ON \
          -DBUILD_TESTS=ON

        cmake --build build --parallel 4

    - name: Run tests and collect coverage
      run: |
        cd build
        ctest --output-on-failure
        gcovr --root .. --sonarqube coverage.xml .

    - name: SonarCloud Scan
      uses: SonarSource/sonarcloud-github-action@master
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
```

### Benchmarking Workflow

```yaml
# .github/workflows/benchmark.yml
name: Performance Benchmarks

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 3 * * *'  # Daily at 3 AM

jobs:
  benchmark:
    name: Performance Benchmarks
    runs-on: ubuntu-22.04

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup environment
      run: |
        sudo apt-get update
        sudo apt-get install -y \
          gcc-13 \
          g++-13 \
          ninja-build \
          linux-tools-generic \
          linux-tools-$(uname -r)

        echo "CC=gcc-13" >> $GITHUB_ENV
        echo "CXX=g++-13" >> $GITHUB_ENV

    - name: Setup Python and Conan
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'

    - name: Install Conan
      run: |
        pip install conan
        conan profile detect --force

    - name: Build benchmarks
      run: |
        conan install . --output-folder=build --build=missing

        cmake -B build -S . -G Ninja \
          -DCMAKE_TOOLCHAIN_FILE=build/conan_toolchain.cmake \
          -DCMAKE_BUILD_TYPE=Release \
          -DBUILD_BENCHMARKS=ON \
          -DCMAKE_CXX_FLAGS="-march=native -mtune=native"

        cmake --build build --parallel 4

    - name: Run benchmarks
      run: |
        cd build
        ./benchmarks/bench_mylib --benchmark_format=json --benchmark_out=benchmark_results.json

    - name: Store benchmark results
      uses: benchmark-action/github-action-benchmark@v1
      with:
        tool: 'googlecpp'
        output-file-path: build/benchmark_results.json
        github-token: ${{ secrets.GITHUB_TOKEN }}
        auto-push: true
        comment-on-alert: true
        alert-threshold: '200%'
        fail-on-alert: true

    - name: Upload benchmark artifacts
      uses: actions/upload-artifact@v3
      with:
        name: benchmark-results
        path: build/benchmark_results.json
```

### Documentation Workflow

```yaml
# .github/workflows/documentation.yml
name: Documentation

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]

jobs:
  build-docs:
    name: Build Documentation
    runs-on: ubuntu-22.04

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Doxygen
      run: |
        sudo apt-get update
        sudo apt-get install -y \
          doxygen \
          graphviz \
          plantuml

    - name: Setup Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'

    - name: Install Python dependencies
      run: |
        pip install \
          sphinx \
          sphinx-rtd-theme \
          breathe \
          exhale \
          myst-parser

    - name: Generate Doxygen documentation
      run: |
        doxygen docs/Doxyfile

    - name: Build Sphinx documentation
      run: |
        cd docs
        sphinx-build -b html . _build/html

    - name: Deploy to GitHub Pages
      if: github.ref == 'refs/heads/main'
      uses: peaceiris/actions-gh-pages@v3
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
        publish_dir: docs/_build/html
        cname: myproject.example.com

    - name: Upload documentation artifacts
      uses: actions/upload-artifact@v3
      with:
        name: documentation
        path: docs/_build/html
```

### Security Scanning Workflow

```yaml
# .github/workflows/security.yml
name: Security Scan

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]
  schedule:
    - cron: '0 4 * * 1'  # Weekly on Monday at 4 AM

jobs:
  codeql-analysis:
    name: CodeQL Analysis
    runs-on: ubuntu-22.04

    strategy:
      fail-fast: false
      matrix:
        language: [ 'cpp' ]

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Initialize CodeQL
      uses: github/codeql-action/init@v2
      with:
        languages: ${{ matrix.language }}
        config-file: ./.github/codeql/codeql-config.yml

    - name: Setup environment
      run: |
        sudo apt-get update
        sudo apt-get install -y gcc-13 g++-13 ninja-build

    - name: Setup Python and Conan
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'

    - name: Install Conan
      run: |
        pip install conan
        conan profile detect --force

    - name: Build project
      run: |
        conan install . --output-folder=build --build=missing

        cmake -B build -S . -G Ninja \
          -DCMAKE_TOOLCHAIN_FILE=build/conan_toolchain.cmake \
          -DCMAKE_BUILD_TYPE=Debug \
          -DBUILD_TESTS=ON

        cmake --build build --parallel 4

    - name: Perform CodeQL Analysis
      uses: github/codeql-action/analyze@v2

  dependency-scan:
    name: Dependency Vulnerability Scan
    runs-on: ubuntu-22.04

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: Upload Trivy scan results
      uses: github/codeql-action/upload-sarif@v2
      if: always()
      with:
        sarif_file: 'trivy-results.sarif'
```

### Release and Packaging Workflow

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags: [ 'v*' ]

permissions:
  contents: write
  packages: write

jobs:
  create-release:
    name: Create Release
    runs-on: ubuntu-22.04
    outputs:
      upload_url: ${{ steps.create_release.outputs.upload_url }}
      version: ${{ steps.get_version.outputs.version }}

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Get version from tag
      id: get_version
      run: |
        echo "version=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT

    - name: Generate changelog
      id: changelog
      run: |
        # Generate changelog from commits since last tag
        git log $(git describe --tags --abbrev=0 HEAD^)..HEAD --pretty=format:"* %s" > changelog.txt
        echo "changelog<<EOF" >> $GITHUB_OUTPUT
        cat changelog.txt >> $GITHUB_OUTPUT
        echo "EOF" >> $GITHUB_OUTPUT

    - name: Create Release
      id: create_release
      uses: actions/create-release@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        tag_name: ${{ github.ref }}
        release_name: Release ${{ steps.get_version.outputs.version }}
        body: |
          ## Changes in ${{ steps.get_version.outputs.version }}

          ${{ steps.changelog.outputs.changelog }}
        draft: false
        prerelease: ${{ contains(github.ref, 'alpha') || contains(github.ref, 'beta') || contains(github.ref, 'rc') }}

  build-packages:
    name: Build Packages (${{ matrix.config.name }})
    runs-on: ${{ matrix.config.os }}
    needs: create-release

    strategy:
      fail-fast: false
      matrix:
        config:
          - {
              name: "Linux x64",
              os: ubuntu-22.04,
              cc: "gcc-13",
              cxx: "g++-13",
              packages: "gcc-13 g++-13",
              archive_name: "linux-x64"
            }
          - {
              name: "Windows x64",
              os: windows-2022,
              archive_name: "windows-x64"
            }
          - {
              name: "macOS x64",
              os: macos-13,
              cc: "clang",
              cxx: "clang++",
              archive_name: "macos-x64"
            }

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup environment
      shell: bash
      run: |
        if [ "${{ matrix.config.os }}" = "ubuntu-22.04" ]; then
          sudo apt-get update
          sudo apt-get install -y ninja-build ${{ matrix.config.packages }}
          echo "CC=${{ matrix.config.cc }}" >> $GITHUB_ENV
          echo "CXX=${{ matrix.config.cxx }}" >> $GITHUB_ENV
        elif [ "${{ matrix.config.os }}" = "windows-2022" ]; then
          choco install ninja
        elif [ "${{ matrix.config.os }}" = "macos-13" ]; then
          brew install ninja
          echo "CC=${{ matrix.config.cc }}" >> $GITHUB_ENV
          echo "CXX=${{ matrix.config.cxx }}" >> $GITHUB_ENV
        fi

    - name: Setup Python and Conan
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'

    - name: Install Conan
      run: |
        pip install conan
        conan profile detect --force

    - name: Build and package
      run: |
        # Install dependencies
        conan install . --output-folder=build --build=missing

        # Configure
        cmake -B build -S . -G Ninja \
          -DCMAKE_TOOLCHAIN_FILE=build/conan_toolchain.cmake \
          -DCMAKE_BUILD_TYPE=Release \
          -DCMAKE_INSTALL_PREFIX=install \
          -DBUILD_TESTS=OFF \
          -DBUILD_BENCHMARKS=OFF

        # Build
        cmake --build build --parallel 4

        # Install
        cmake --install build

        # Package
        cd install
        if [ "$RUNNER_OS" = "Windows" ]; then
          7z a ../myproject-${{ needs.create-release.outputs.version }}-${{ matrix.config.archive_name }}.zip *
        else
          tar -czf ../myproject-${{ needs.create-release.outputs.version }}-${{ matrix.config.archive_name }}.tar.gz *
        fi

    - name: Upload package to release
      uses: actions/upload-release-asset@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        upload_url: ${{ needs.create-release.outputs.upload_url }}
        asset_path: ${{ runner.os == 'Windows' && format('./myproject-{0}-{1}.zip', needs.create-release.outputs.version, matrix.config.archive_name) || format('./myproject-{0}-{1}.tar.gz', needs.create-release.outputs.version, matrix.config.archive_name) }}
        asset_name: ${{ runner.os == 'Windows' && format('myproject-{0}-{1}.zip', needs.create-release.outputs.version, matrix.config.archive_name) || format('myproject-{0}-{1}.tar.gz', needs.create-release.outputs.version, matrix.config.archive_name) }}
        asset_content_type: ${{ runner.os == 'Windows' && 'application/zip' || 'application/gzip' }}

  publish-packages:
    name: Publish Packages
    runs-on: ubuntu-22.04
    needs: [create-release, build-packages]

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Python and Conan
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'

    - name: Install Conan
      run: |
        pip install conan
        conan profile detect --force

    - name: Create and upload Conan package
      run: |
        # Create package
        conan create . --build=missing

        # Upload to Conan repository
        conan upload myproject/${{ needs.create-release.outputs.version }} \
          --remote=conancenter \
          --all \
          --confirm
      env:
        CONAN_PASSWORD: ${{ secrets.CONAN_PASSWORD }}
        CONAN_LOGIN_USERNAME: ${{ secrets.CONAN_USERNAME }}

    - name: Build and push Docker image
      if: github.event_name == 'push' && startsWith(github.ref, 'refs/tags/v')
      run: |
        docker build -t myproject:${{ needs.create-release.outputs.version }} .
        docker tag myproject:${{ needs.create-release.outputs.version }} myproject:latest

        echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin

        docker push myproject:${{ needs.create-release.outputs.version }}
        docker push myproject:latest
```

### Configuration Files

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
  portability-*,
  readability-*,
  -modernize-use-trailing-return-type,
  -readability-identifier-length,
  -cppcoreguidelines-avoid-magic-numbers,
  -readability-magic-numbers

WarningsAsErrors: '*'

CheckOptions:
  - key: readability-identifier-naming.ClassCase
    value: CamelCase
  - key: readability-identifier-naming.FunctionCase
    value: lower_case
  - key: readability-identifier-naming.VariableCase
    value: lower_case
  - key: readability-identifier-naming.ConstantCase
    value: UPPER_CASE
  - key: cppcoreguidelines-special-member-functions.AllowSoleDefaultDtor
    value: true
  - key: modernize-loop-convert.MaxCopySize
    value: 16
  - key: performance-move-const-arg.CheckTriviallyCopyableMove
    value: false
```

#### SonarCloud Configuration

```properties
# sonar-project.properties
sonar.projectKey=myproject
sonar.organization=myorg
sonar.projectName=MyProject
sonar.projectVersion=1.0.0

sonar.sources=src/
sonar.tests=tests/
sonar.cfamily.build-wrapper-output=build/
sonar.cfamily.cache.enabled=true
sonar.cfamily.cache.path=build/sonar-cache
sonar.cfamily.threads=4
sonar.coverageReportPaths=build/coverage.xml

sonar.exclusions=**/third_party/**,**/external/**
sonar.test.exclusions=tests/**
sonar.coverage.exclusions=tests/**,examples/**,benchmarks/**

sonar.cpd.exclusions=tests/**
sonar.lang.patterns.cpp=**/*.cpp,**/*.hpp
```

#### CodeQL Configuration

```yaml
# .github/codeql/codeql-config.yml
name: "CodeQL Config"

disable-default-queries: false

queries:
  - uses: security-and-quality
  - uses: security-extended

paths-ignore:
  - "**/third_party/**"
  - "**/external/**"
  - "**/build/**"
  - "**/tests/**"

paths:
  - "src/"
  - "include/"
```

#### Dockerfile for Packaging

```dockerfile
# Dockerfile
FROM ubuntu:22.04 AS builder

RUN apt-get update && apt-get install -y \
    gcc-13 \
    g++-13 \
    cmake \
    ninja-build \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

RUN pip3 install conan

WORKDIR /app
COPY . .

RUN conan profile detect --force && \
    conan install . --output-folder=build --build=missing && \
    cmake -B build -S . -G Ninja \
        -DCMAKE_TOOLCHAIN_FILE=build/conan_toolchain.cmake \
        -DCMAKE_BUILD_TYPE=Release \
        -DBUILD_TESTS=OFF && \
    cmake --build build --parallel 4 && \
    cmake --install build --prefix /usr/local

FROM ubuntu:22.04
RUN apt-get update && apt-get install -y \
    libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/ /usr/local/

ENTRYPOINT ["/usr/local/bin/myapp"]
```

This comprehensive GitHub Actions CI/CD pipeline provides automated building, testing, static analysis, benchmarking, documentation generation, security scanning, and package deployment for modern C++ projects.