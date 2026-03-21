# C++ Local Build Systems and Development Setup

## Overview

Modern C++ development relies on robust build systems, package managers, and toolchains. This guide covers CMake, build configuration, package management with Conan and vcpkg, and local development environment setup.

## CMake Build System

### Basic CMake Structure

```cmake
# CMakeLists.txt - Project root
cmake_minimum_required(VERSION 3.20)

project(MyProject
    VERSION 1.0.0
    DESCRIPTION "A modern C++ project"
    LANGUAGES CXX
)

# Set C++ standard
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Compiler-specific options
if(MSVC)
    add_compile_options(/W4)
else()
    add_compile_options(-Wall -Wextra -Wpedantic)
endif()

# Build types
set(CMAKE_CONFIGURATION_TYPES "Debug;Release;RelWithDebInfo;MinSizeRel")

if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE "Release" CACHE STRING "Build type" FORCE)
endif()

# Output directories
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)

# Find packages
find_package(Threads REQUIRED)
find_package(fmt REQUIRED)
find_package(spdlog REQUIRED)

# Add subdirectories
add_subdirectory(src)
add_subdirectory(tests)
add_subdirectory(examples)

# Optional documentation
option(BUILD_DOCS "Build documentation" OFF)
if(BUILD_DOCS)
    add_subdirectory(docs)
endif()

# Optional benchmarks
option(BUILD_BENCHMARKS "Build benchmarks" OFF)
if(BUILD_BENCHMARKS)
    add_subdirectory(benchmarks)
endif()
```

### Library Target Configuration

```cmake
# src/CMakeLists.txt
# Create library target
add_library(mylib
    mylib.cpp
    mylib.hpp
    utils.cpp
    utils.hpp
    internal/private.cpp
    internal/private.hpp
)

# Alias for consistent naming
add_library(MyProject::mylib ALIAS mylib)

# Target properties
target_compile_features(mylib PUBLIC cxx_std_20)

# Include directories
target_include_directories(mylib
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/internal
)

# Link libraries
target_link_libraries(mylib
    PUBLIC
        fmt::fmt
        Threads::Threads
    PRIVATE
        spdlog::spdlog
)

# Compiler definitions
target_compile_definitions(mylib
    PUBLIC
        MYLIB_VERSION_MAJOR=${PROJECT_VERSION_MAJOR}
        MYLIB_VERSION_MINOR=${PROJECT_VERSION_MINOR}
    PRIVATE
        $<$<CONFIG:Debug>:MYLIB_DEBUG>
        $<$<CONFIG:Release>:MYLIB_RELEASE>
)

# Platform-specific settings
if(WIN32)
    target_compile_definitions(mylib PRIVATE MYLIB_WINDOWS)
    target_link_libraries(mylib PRIVATE ws2_32)
elseif(UNIX)
    target_compile_definitions(mylib PRIVATE MYLIB_UNIX)
    target_link_libraries(mylib PRIVATE ${CMAKE_DL_LIBS})
endif()

# Export targets for find_package
install(TARGETS mylib
    EXPORT MyProjectTargets
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
    INCLUDES DESTINATION include
)

install(FILES
    mylib.hpp
    utils.hpp
    DESTINATION include
)
```

### Executable Target Configuration

```cmake
# Create executable
add_executable(myapp
    main.cpp
    app.cpp
    app.hpp
)

# Link to library
target_link_libraries(myapp
    PRIVATE
        MyProject::mylib
        fmt::fmt
)

# Install executable
install(TARGETS myapp
    DESTINATION bin
)

# Resource files
configure_file(
    ${CMAKE_CURRENT_SOURCE_DIR}/config.json.in
    ${CMAKE_CURRENT_BINARY_DIR}/config.json
    @ONLY
)

install(FILES
    ${CMAKE_CURRENT_BINARY_DIR}/config.json
    DESTINATION share/myapp
)
```

### Testing Configuration

```cmake
# tests/CMakeLists.txt
enable_testing()

# Find testing framework
find_package(Catch2 3 REQUIRED)
# or
# find_package(GTest REQUIRED)

# Test executable
add_executable(test_mylib
    test_main.cpp
    test_mylib.cpp
    test_utils.cpp
)

target_link_libraries(test_mylib
    PRIVATE
        MyProject::mylib
        Catch2::Catch2WithMain
        # or GTest::gtest_main
)

# Discover tests automatically
include(Catch)
catch_discover_tests(test_mylib)

# or for GoogleTest:
# include(GoogleTest)
# gtest_discover_tests(test_mylib)

# Custom test command
add_test(NAME unit_tests COMMAND test_mylib)

# Coverage target (optional)
option(ENABLE_COVERAGE "Enable coverage reporting" OFF)
if(ENABLE_COVERAGE AND CMAKE_COMPILER_IS_GNUCXX)
    target_compile_options(test_mylib PRIVATE --coverage)
    target_link_libraries(test_mylib PRIVATE --coverage)
endif()
```

### Advanced CMake Features

```cmake
# Feature detection
include(CheckCXXSourceCompiles)
include(CheckIncludeFileCXX)

check_cxx_source_compiles("
    #include <filesystem>
    int main() { std::filesystem::path p; return 0; }
" HAS_FILESYSTEM)

if(HAS_FILESYSTEM)
    target_compile_definitions(mylib PRIVATE HAS_FILESYSTEM)
endif()

# Custom functions
function(add_myproject_executable target)
    set(options INSTALL)
    set(oneValueArgs DIRECTORY)
    set(multiValueArgs SOURCES LIBRARIES)

    cmake_parse_arguments(ARG "${options}" "${oneValueArgs}" "${multiValueArgs}" ${ARGN})

    add_executable(${target} ${ARG_SOURCES})

    target_link_libraries(${target} PRIVATE ${ARG_LIBRARIES})

    if(ARG_INSTALL)
        install(TARGETS ${target} DESTINATION ${ARG_DIRECTORY})
    endif()
endfunction()

# Usage
add_myproject_executable(myapp
    SOURCES main.cpp app.cpp
    LIBRARIES MyProject::mylib
    DIRECTORY bin
    INSTALL
)

# Custom targets
add_custom_target(format
    COMMAND clang-format -i ${CMAKE_SOURCE_DIR}/src/*.cpp ${CMAKE_SOURCE_DIR}/src/*.hpp
    COMMENT "Formatting source code"
)

add_custom_target(lint
    COMMAND clang-tidy ${CMAKE_SOURCE_DIR}/src/*.cpp -- -I${CMAKE_SOURCE_DIR}/include
    COMMENT "Running static analysis"
)

# Generate compile commands for IDEs
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
```

### Package Configuration

```cmake
# cmake/MyProjectConfig.cmake.in
@PACKAGE_INIT@

include(CMakeFindDependencyMacro)

find_dependency(fmt)
find_dependency(Threads)

if(@BUILD_SHARED_LIBS@)
    include("${CMAKE_CURRENT_LIST_DIR}/MyProjectTargets.cmake")
else()
    include("${CMAKE_CURRENT_LIST_DIR}/MyProjectTargets-static.cmake")
endif()

check_required_components(MyProject)
```

```cmake
# Root CMakeLists.txt - Package installation
include(CMakePackageConfigHelpers)

# Generate config file
configure_package_config_file(
    cmake/MyProjectConfig.cmake.in
    ${CMAKE_CURRENT_BINARY_DIR}/MyProjectConfig.cmake
    INSTALL_DESTINATION lib/cmake/MyProject
)

# Generate version file
write_basic_package_version_file(
    ${CMAKE_CURRENT_BINARY_DIR}/MyProjectConfigVersion.cmake
    VERSION ${PROJECT_VERSION}
    COMPATIBILITY SameMajorVersion
)

# Install config files
install(FILES
    ${CMAKE_CURRENT_BINARY_DIR}/MyProjectConfig.cmake
    ${CMAKE_CURRENT_BINARY_DIR}/MyProjectConfigVersion.cmake
    DESTINATION lib/cmake/MyProject
)

# Export targets
install(EXPORT MyProjectTargets
    FILE MyProjectTargets.cmake
    NAMESPACE MyProject::
    DESTINATION lib/cmake/MyProject
)
```

## Package Managers

### Conan Package Manager

#### Basic Conanfile

```python
# conanfile.py
from conan import ConanFile
from conan.tools.cmake import CMakeToolchain, CMakeDeps, cmake_layout
from conan.tools.files import copy

class MyProjectConan(ConanFile):
    name = "myproject"
    version = "1.0.0"

    # Package metadata
    license = "MIT"
    description = "A modern C++ library"
    url = "https://github.com/user/myproject"

    # Settings and options
    settings = "os", "compiler", "build_type", "arch"
    options = {
        "shared": [True, False],
        "with_tests": [True, False],
        "fPIC": [True, False]
    }
    default_options = {
        "shared": False,
        "with_tests": False,
        "fPIC": True
    }

    # Requirements
    def requirements(self):
        self.requires("fmt/9.1.0")
        self.requires("spdlog/1.11.0")

        if self.options.with_tests:
            self.test_requires("catch2/3.2.1")

    def configure(self):
        if self.settings.os == "Windows":
            self.options.rm_safe("fPIC")

    def layout(self):
        cmake_layout(self)

    def generate(self):
        tc = CMakeToolchain(self)
        tc.variables["BUILD_TESTS"] = self.options.with_tests
        tc.generate()

        deps = CMakeDeps(self)
        deps.generate()

    def build(self):
        cmake = CMake(self)
        cmake.configure()
        cmake.build()

        if self.options.with_tests:
            self.run("ctest --output-on-failure")

    def package(self):
        copy(self, "*.hpp", src=self.source_folder, dst=f"{self.package_folder}/include")
        copy(self, "*.a", src=self.build_folder, dst=f"{self.package_folder}/lib", keep_path=False)
        copy(self, "*.so", src=self.build_folder, dst=f"{self.package_folder}/lib", keep_path=False)
        copy(self, "*.dll", src=self.build_folder, dst=f"{self.package_folder}/bin", keep_path=False)

    def package_info(self):
        self.cpp_info.libs = ["myproject"]
        self.cpp_info.includedirs = ["include"]

        if self.settings.os in ["Linux", "FreeBSD"]:
            self.cpp_info.system_libs = ["pthread", "m"]
        elif self.settings.os == "Windows":
            self.cpp_info.system_libs = ["ws2_32"]
```

#### Consumer Conanfile

```python
# Consumer conanfile.txt
[requires]
fmt/9.1.0
spdlog/1.11.0
boost/1.81.0

[options]
boost:header_only=True
fmt:shared=False

[generators]
CMakeDeps
CMakeToolchain

[layout]
cmake_layout
```

#### Conan Profile Configuration

```ini
# profiles/gcc-11
[settings]
os=Linux
arch=x86_64
compiler=gcc
compiler.version=11
compiler.libcxx=libstdc++11
build_type=Release

[options]
*:shared=False

[env]
CC=gcc-11
CXX=g++-11
```

#### Conan Build Commands

```bash
# Create and build with Conan
conan create . --profile gcc-11

# Install dependencies
conan install . --build missing --profile gcc-11

# Build with dependencies
mkdir build && cd build
conan install .. --build missing
cmake .. -DCMAKE_TOOLCHAIN_FILE=conan_toolchain.cmake
cmake --build .

# Upload to repository
conan upload myproject/1.0.0 -r artifactory --all
```

### vcpkg Package Manager

#### vcpkg.json Manifest

```json
{
  "name": "myproject",
  "version-string": "1.0.0",
  "description": "A modern C++ project",
  "homepage": "https://github.com/user/myproject",
  "license": "MIT",
  "supports": "!(uwp | arm)",
  "dependencies": [
    "fmt",
    "spdlog",
    {
      "name": "boost",
      "features": ["system", "filesystem", "thread"]
    }
  ],
  "features": {
    "tests": {
      "description": "Build tests",
      "dependencies": [
        "catch2"
      ]
    },
    "benchmarks": {
      "description": "Build benchmarks",
      "dependencies": [
        "benchmark"
      ]
    }
  }
}
```

#### CMake Integration with vcpkg

```cmake
# CMakeLists.txt
cmake_minimum_required(VERSION 3.20)

# vcpkg toolchain
if(DEFINED ENV{VCPKG_ROOT} AND NOT DEFINED CMAKE_TOOLCHAIN_FILE)
    set(CMAKE_TOOLCHAIN_FILE "$ENV{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake"
        CACHE STRING "")
endif()

project(MyProject)

# Find packages installed by vcpkg
find_package(fmt CONFIG REQUIRED)
find_package(spdlog CONFIG REQUIRED)

# Optional features
find_package(Boost COMPONENTS system filesystem thread)
if(Boost_FOUND)
    target_compile_definitions(mylib PRIVATE HAS_BOOST)
    target_link_libraries(mylib PRIVATE Boost::system Boost::filesystem Boost::thread)
endif()

# Test dependencies
if(BUILD_TESTS)
    find_package(Catch2 3 CONFIG REQUIRED)
    target_link_libraries(test_mylib PRIVATE Catch2::Catch2WithMain)
endif()
```

#### vcpkg Build Commands

```bash
# Install vcpkg
git clone https://github.com/Microsoft/vcpkg.git
cd vcpkg
./bootstrap-vcpkg.sh

# Install packages
./vcpkg install fmt spdlog catch2

# With manifest mode
./vcpkg install --triplet x64-linux

# Custom triplets
./vcpkg install --triplet x64-linux-static

# Integration
./vcpkg integrate install

# Build project
mkdir build && cd build
cmake .. -DCMAKE_TOOLCHAIN_FILE=../vcpkg/scripts/buildsystems/vcpkg.cmake
cmake --build .
```

### Custom Triplet Configuration

```cmake
# triplets/x64-linux-static.cmake
set(VCPKG_TARGET_ARCHITECTURE x64)
set(VCPKG_CRT_LINKAGE static)
set(VCPKG_LIBRARY_LINKAGE static)
set(VCPKG_CMAKE_SYSTEM_NAME Linux)

# Custom compiler
set(VCPKG_C_FLAGS "-march=native")
set(VCPKG_CXX_FLAGS "-march=native")
```

## Toolchain Configuration

### Cross-Compilation Toolchain

```cmake
# toolchains/arm-linux-gnueabihf.cmake
set(CMAKE_SYSTEM_NAME Linux)
set(CMAKE_SYSTEM_PROCESSOR arm)

# Compiler paths
set(CMAKE_C_COMPILER arm-linux-gnueabihf-gcc)
set(CMAKE_CXX_COMPILER arm-linux-gnueabihf-g++)
set(CMAKE_ASM_COMPILER arm-linux-gnueabihf-gcc)

# Root path
set(CMAKE_FIND_ROOT_PATH /usr/arm-linux-gnueabihf)

# Search settings
set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)
set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE ONLY)

# Compiler flags
set(CMAKE_C_FLAGS "-march=armv7-a -mfpu=neon")
set(CMAKE_CXX_FLAGS "-march=armv7-a -mfpu=neon")
```

### MSVC Toolchain Configuration

```cmake
# toolchains/msvc-x64.cmake
set(CMAKE_SYSTEM_NAME Windows)
set(CMAKE_SYSTEM_PROCESSOR x64)

# Use specific Visual Studio version
set(CMAKE_GENERATOR_PLATFORM x64)
set(CMAKE_GENERATOR_TOOLSET v143)

# Compiler flags
set(CMAKE_CXX_FLAGS "/W4 /permissive- /Zc:__cplusplus")
set(CMAKE_CXX_FLAGS_DEBUG "/MDd /Zi /Od /RTC1")
set(CMAKE_CXX_FLAGS_RELEASE "/MD /O2 /Ob2 /DNDEBUG")

# Linker flags
set(CMAKE_EXE_LINKER_FLAGS "/SUBSYSTEM:CONSOLE")
set(CMAKE_SHARED_LINKER_FLAGS "/SUBSYSTEM:WINDOWS")
```

### Clang Toolchain Configuration

```cmake
# toolchains/clang-17.cmake
set(CMAKE_C_COMPILER clang-17)
set(CMAKE_CXX_COMPILER clang++-17)

# Compiler flags
set(CMAKE_C_FLAGS "-Wall -Wextra -Wpedantic")
set(CMAKE_CXX_FLAGS "-Wall -Wextra -Wpedantic -stdlib=libc++")

# Linker
set(CMAKE_EXE_LINKER_FLAGS "-fuse-ld=lld")
set(CMAKE_SHARED_LINKER_FLAGS "-fuse-ld=lld")

# Static analysis
option(ENABLE_CLANG_TIDY "Enable clang-tidy" OFF)
if(ENABLE_CLANG_TIDY)
    set(CMAKE_CXX_CLANG_TIDY clang-tidy-17)
endif()

# Sanitizers
option(ENABLE_ASAN "Enable AddressSanitizer" OFF)
option(ENABLE_TSAN "Enable ThreadSanitizer" OFF)
option(ENABLE_MSAN "Enable MemorySanitizer" OFF)

if(ENABLE_ASAN)
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fsanitize=address -fno-omit-frame-pointer")
    set(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} -fsanitize=address")
endif()

if(ENABLE_TSAN)
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fsanitize=thread")
    set(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} -fsanitize=thread")
endif()

if(ENABLE_MSAN)
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fsanitize=memory -fsanitize-memory-track-origins")
    set(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} -fsanitize=memory")
endif()
```

## Development Environment Setup

### VSCode Configuration

```json
// .vscode/settings.json
{
    "cmake.configureSettings": {
        "CMAKE_EXPORT_COMPILE_COMMANDS": "ON",
        "CMAKE_BUILD_TYPE": "Debug"
    },
    "cmake.buildDirectory": "${workspaceFolder}/build",
    "cmake.generator": "Ninja",
    "C_Cpp.default.configurationProvider": "ms-vscode.cmake-tools",
    "C_Cpp.default.cppStandard": "c++20",
    "C_Cpp.default.intelliSenseMode": "gcc-x64",
    "C_Cpp.clang_format_style": "file",
    "files.associations": {
        "*.hpp": "cpp",
        "CMakeLists.txt": "cmake"
    },
    "cmake.debugConfig": {
        "args": ["--config", "Debug"],
        "cwd": "${workspaceFolder}/build"
    }
}
```

```json
// .vscode/tasks.json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Configure CMake",
            "type": "shell",
            "command": "cmake",
            "args": ["-B", "build", "-S", ".", "-G", "Ninja"],
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        },
        {
            "label": "Build Project",
            "type": "shell",
            "command": "cmake",
            "args": ["--build", "build", "--parallel"],
            "group": {
                "kind": "build",
                "isDefault": true
            },
            "dependsOn": "Configure CMake"
        },
        {
            "label": "Run Tests",
            "type": "shell",
            "command": "ctest",
            "args": ["--test-dir", "build", "--output-on-failure"],
            "group": "test",
            "dependsOn": "Build Project"
        },
        {
            "label": "Clean Build",
            "type": "shell",
            "command": "cmake",
            "args": ["--build", "build", "--target", "clean"]
        },
        {
            "label": "Format Code",
            "type": "shell",
            "command": "clang-format",
            "args": ["-i", "src/*.cpp", "src/*.hpp", "tests/*.cpp"]
        }
    ]
}
```

```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Application",
            "type": "cppdbg",
            "request": "launch",
            "program": "${workspaceFolder}/build/bin/myapp",
            "args": [],
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
            "preLaunchTask": "Build Project"
        },
        {
            "name": "Debug Tests",
            "type": "cppdbg",
            "request": "launch",
            "program": "${workspaceFolder}/build/tests/test_mylib",
            "args": [],
            "stopAtEntry": false,
            "cwd": "${workspaceFolder}",
            "environment": [],
            "externalConsole": false,
            "MIMode": "gdb",
            "preLaunchTask": "Build Project"
        }
    ]
}
```

### CLion Configuration

```yaml
# .idea/cmake.xml equivalent settings
cmake:
  profiles:
    - name: Debug
      build_type: Debug
      cmake_options: -DBUILD_TESTS=ON -DENABLE_COVERAGE=ON
      build_options: --parallel 4
      toolchain: Default
    - name: Release
      build_type: Release
      cmake_options: -DBUILD_TESTS=OFF -DCMAKE_BUILD_TYPE=Release
      build_options: --parallel 4
      toolchain: Default
    - name: Sanitized
      build_type: Debug
      cmake_options: -DENABLE_ASAN=ON -DENABLE_UBSAN=ON
      build_options: --parallel 4
      toolchain: Clang
```

### Development Scripts

```bash
#!/bin/bash
# scripts/build.sh

set -e

# Configuration
BUILD_TYPE=${1:-Release}
BUILD_DIR="build-${BUILD_TYPE,,}"
INSTALL_DIR="install"
PARALLEL_JOBS=$(nproc)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Clean previous build
if [ -d "$BUILD_DIR" ]; then
    echo_info "Cleaning previous build directory"
    rm -rf "$BUILD_DIR"
fi

# Configure
echo_info "Configuring project (Build Type: $BUILD_TYPE)"
cmake -B "$BUILD_DIR" -S . \
    -DCMAKE_BUILD_TYPE="$BUILD_TYPE" \
    -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
    -DBUILD_TESTS=ON \
    -DBUILD_BENCHMARKS=ON \
    -DCMAKE_EXPORT_COMPILE_COMMANDS=ON \
    -G Ninja

# Build
echo_info "Building project with $PARALLEL_JOBS parallel jobs"
cmake --build "$BUILD_DIR" --parallel "$PARALLEL_JOBS"

# Test
echo_info "Running tests"
ctest --test-dir "$BUILD_DIR" --output-on-failure --parallel "$PARALLEL_JOBS"

# Install
echo_info "Installing to $INSTALL_DIR"
cmake --install "$BUILD_DIR"

echo_info "Build completed successfully!"
```

```bash
#!/bin/bash
# scripts/setup-dev.sh

# Install system dependencies (Ubuntu/Debian)
sudo apt-get update
sudo apt-get install -y \
    build-essential \
    cmake \
    ninja-build \
    clang-17 \
    clang-tools-17 \
    lldb-17 \
    libc++-17-dev \
    libc++abi-17-dev \
    git \
    pkg-config \
    curl \
    zip \
    unzip \
    tar

# Install vcpkg
if [ ! -d "vcpkg" ]; then
    git clone https://github.com/Microsoft/vcpkg.git
    cd vcpkg
    ./bootstrap-vcpkg.sh
    ./vcpkg integrate install
    cd ..
fi

# Install Conan
pip3 install --user conan

# Setup Conan profile
conan profile detect --force

# Create development directories
mkdir -p build
mkdir -p install
mkdir -p .vscode

echo "Development environment setup complete!"
echo "Run './scripts/build.sh Debug' to build the project"
```

### Continuous Integration Preparation

```yaml
# .github/workflows/build.yml (preparation for CI)
name: Build and Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    name: ${{ matrix.os }}-${{ matrix.build_type }}-${{ matrix.compiler }}
    runs-on: ${{ matrix.os }}

    strategy:
      matrix:
        os: [ubuntu-22.04, windows-2022, macos-12]
        build_type: [Debug, Release]
        compiler: [gcc, clang, msvc]
        exclude:
          - os: windows-2022
            compiler: gcc
          - os: windows-2022
            compiler: clang
          - os: ubuntu-22.04
            compiler: msvc
          - os: macos-12
            compiler: msvc

    steps:
    - uses: actions/checkout@v3

    - name: Setup environment
      shell: bash
      run: |
        if [ "${{ matrix.os }}" = "ubuntu-22.04" ]; then
          sudo apt-get update
          sudo apt-get install -y ninja-build
          if [ "${{ matrix.compiler }}" = "clang" ]; then
            sudo apt-get install -y clang-17 libc++-17-dev libc++abi-17-dev
          fi
        elif [ "${{ matrix.os }}" = "macos-12" ]; then
          brew install ninja
        fi

    - name: Configure CMake
      run: |
        cmake -B build -S . -G Ninja \
          -DCMAKE_BUILD_TYPE=${{ matrix.build_type }} \
          -DBUILD_TESTS=ON

    - name: Build
      run: cmake --build build --parallel 4

    - name: Test
      run: ctest --test-dir build --output-on-failure
```

This comprehensive guide covers modern C++ build systems, package management, toolchain configuration, and development environment setup for production-ready projects.