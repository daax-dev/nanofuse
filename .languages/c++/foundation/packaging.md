# C++ Packaging, Distribution, and Deployment

## Overview

Modern C++ packaging involves creating distributable libraries, applications, and containers. This guide covers library creation, package management integration, cross-platform distribution, containerization, and deployment strategies for C++ projects.

## Library Creation and Packaging

### Static Library Creation

```cmake
# CMakeLists.txt for static library
cmake_minimum_required(VERSION 3.20)
project(MyStaticLib VERSION 1.0.0)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Create static library
add_library(MyStaticLib STATIC
    src/core.cpp
    src/utils.cpp
    src/algorithm.cpp
)

# Alias for consistent naming
add_library(MyStaticLib::MyStaticLib ALIAS MyStaticLib)

# Public headers
target_include_directories(MyStaticLib
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src
)

# Dependencies
find_package(fmt REQUIRED)
target_link_libraries(MyStaticLib PUBLIC fmt::fmt)

# Compiler features
target_compile_features(MyStaticLib PUBLIC cxx_std_20)

# Install targets
install(TARGETS MyStaticLib
    EXPORT MyStaticLibTargets
    ARCHIVE DESTINATION lib
    LIBRARY DESTINATION lib
    RUNTIME DESTINATION bin
    INCLUDES DESTINATION include
)

# Install headers
install(DIRECTORY include/
    DESTINATION include
    FILES_MATCHING PATTERN "*.hpp"
)

# Generate and install config files
include(CMakePackageConfigHelpers)

configure_package_config_file(
    "${CMAKE_CURRENT_SOURCE_DIR}/cmake/MyStaticLibConfig.cmake.in"
    "${CMAKE_CURRENT_BINARY_DIR}/MyStaticLibConfig.cmake"
    INSTALL_DESTINATION "lib/cmake/MyStaticLib"
)

write_basic_package_version_file(
    "${CMAKE_CURRENT_BINARY_DIR}/MyStaticLibConfigVersion.cmake"
    VERSION "${PROJECT_VERSION}"
    COMPATIBILITY SameMajorVersion
)

install(FILES
    "${CMAKE_CURRENT_BINARY_DIR}/MyStaticLibConfig.cmake"
    "${CMAKE_CURRENT_BINARY_DIR}/MyStaticLibConfigVersion.cmake"
    DESTINATION "lib/cmake/MyStaticLib"
)

install(EXPORT MyStaticLibTargets
    FILE MyStaticLibTargets.cmake
    NAMESPACE MyStaticLib::
    DESTINATION lib/cmake/MyStaticLib
)
```

### Shared Library Creation

```cmake
# CMakeLists.txt for shared library
cmake_minimum_required(VERSION 3.20)
project(MySharedLib VERSION 1.0.0)

# Create shared library
add_library(MySharedLib SHARED
    src/core.cpp
    src/utils.cpp
    src/algorithm.cpp
)

# Export symbols for Windows
include(GenerateExportHeader)
generate_export_header(MySharedLib
    BASE_NAME MYSHAREDLIB
    EXPORT_MACRO_NAME MYSHAREDLIB_EXPORT
    EXPORT_FILE_NAME ${CMAKE_CURRENT_BINARY_DIR}/mysharedlib_export.h
)

# Configure library properties
set_target_properties(MySharedLib PROPERTIES
    VERSION ${PROJECT_VERSION}
    SOVERSION ${PROJECT_VERSION_MAJOR}
    PUBLIC_HEADER "include/mysharedlib.hpp"
)

# Include directories
target_include_directories(MySharedLib
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<BUILD_INTERFACE:${CMAKE_CURRENT_BINARY_DIR}>
        $<INSTALL_INTERFACE:include>
    PRIVATE
        src
)

# Platform-specific settings
if(WIN32)
    # Windows-specific settings
    target_compile_definitions(MySharedLib PRIVATE
        MYSHAREDLIB_EXPORTS
        WIN32_LEAN_AND_MEAN
        NOMINMAX
    )
endif()

# Install the library
install(TARGETS MySharedLib
    EXPORT MySharedLibTargets
    ARCHIVE DESTINATION lib
    LIBRARY DESTINATION lib
    RUNTIME DESTINATION bin
    PUBLIC_HEADER DESTINATION include
)

# Install export header
install(FILES ${CMAKE_CURRENT_BINARY_DIR}/mysharedlib_export.h
    DESTINATION include
)
```

### Header-Only Library

```cmake
# CMakeLists.txt for header-only library
cmake_minimum_required(VERSION 3.20)
project(MyHeaderOnlyLib VERSION 1.0.0)

# Create interface library
add_library(MyHeaderOnlyLib INTERFACE)
add_library(MyHeaderOnlyLib::MyHeaderOnlyLib ALIAS MyHeaderOnlyLib)

# Interface properties
target_include_directories(MyHeaderOnlyLib
    INTERFACE
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
)

target_compile_features(MyHeaderOnlyLib INTERFACE cxx_std_20)

# Dependencies
find_package(fmt REQUIRED)
target_link_libraries(MyHeaderOnlyLib INTERFACE fmt::fmt)

# Install
install(TARGETS MyHeaderOnlyLib
    EXPORT MyHeaderOnlyLibTargets
    INCLUDES DESTINATION include
)

install(DIRECTORY include/
    DESTINATION include
    FILES_MATCHING PATTERN "*.hpp"
)

# Export configuration
install(EXPORT MyHeaderOnlyLibTargets
    FILE MyHeaderOnlyLibTargets.cmake
    NAMESPACE MyHeaderOnlyLib::
    DESTINATION lib/cmake/MyHeaderOnlyLib
)
```

### Library API Design

```cpp
// include/mylib/api.hpp
#ifndef MYLIB_API_HPP
#define MYLIB_API_HPP

#include "mylib_export.h"
#include <string>
#include <memory>
#include <vector>

namespace mylib {

// Forward declarations
class Implementation;

// Main API class
class MYLIB_EXPORT Calculator {
public:
    Calculator();
    ~Calculator();

    // Copy operations
    Calculator(const Calculator& other);
    Calculator& operator=(const Calculator& other);

    // Move operations
    Calculator(Calculator&& other) noexcept;
    Calculator& operator=(Calculator&& other) noexcept;

    // Public interface
    double add(double a, double b);
    double multiply(double a, double b);

    std::vector<double> get_history() const;
    void clear_history();

    // Configuration
    void set_precision(int precision);
    int get_precision() const;

private:
    std::unique_ptr<Implementation> pimpl;
};

// Free functions
MYLIB_EXPORT std::string get_version();
MYLIB_EXPORT void initialize_library();
MYLIB_EXPORT void cleanup_library();

// Error handling
class MYLIB_EXPORT CalculatorError : public std::exception {
public:
    explicit CalculatorError(const std::string& message);
    const char* what() const noexcept override;

private:
    std::string message_;
};

} // namespace mylib

#endif // MYLIB_API_HPP
```

```cpp
// src/api.cpp
#include "mylib/api.hpp"
#include <iostream>
#include <stdexcept>
#include <iomanip>

namespace mylib {

// PIMPL implementation
class Implementation {
public:
    std::vector<double> history;
    int precision = 6;

    void add_to_history(double value) {
        history.push_back(value);
    }
};

Calculator::Calculator() : pimpl(std::make_unique<Implementation>()) {}

Calculator::~Calculator() = default;

Calculator::Calculator(const Calculator& other)
    : pimpl(std::make_unique<Implementation>(*other.pimpl)) {}

Calculator& Calculator::operator=(const Calculator& other) {
    if (this != &other) {
        pimpl = std::make_unique<Implementation>(*other.pimpl);
    }
    return *this;
}

Calculator::Calculator(Calculator&& other) noexcept = default;
Calculator& Calculator::operator=(Calculator&& other) noexcept = default;

double Calculator::add(double a, double b) {
    double result = a + b;
    pimpl->add_to_history(result);
    return result;
}

double Calculator::multiply(double a, double b) {
    double result = a * b;
    pimpl->add_to_history(result);
    return result;
}

std::vector<double> Calculator::get_history() const {
    return pimpl->history;
}

void Calculator::clear_history() {
    pimpl->history.clear();
}

void Calculator::set_precision(int precision) {
    if (precision < 0 || precision > 15) {
        throw CalculatorError("Precision must be between 0 and 15");
    }
    pimpl->precision = precision;
}

int Calculator::get_precision() const {
    return pimpl->precision;
}

std::string get_version() {
    return "1.0.0";
}

void initialize_library() {
    std::cout << "MyLib " << get_version() << " initialized" << std::endl;
}

void cleanup_library() {
    std::cout << "MyLib cleanup complete" << std::endl;
}

CalculatorError::CalculatorError(const std::string& message) : message_(message) {}

const char* CalculatorError::what() const noexcept {
    return message_.c_str();
}

} // namespace mylib
```

## Package Manager Integration

### Conan Package Recipe

```python
# conanfile.py
from conan import ConanFile
from conan.tools.cmake import CMakeToolchain, CMakeDeps, CMake, cmake_layout
from conan.tools.files import copy, save, load
import os

class MyLibConan(ConanFile):
    name = "mylib"
    version = "1.0.0"

    # Metadata
    license = "MIT"
    description = "A high-performance C++ calculator library"
    url = "https://github.com/user/mylib"
    homepage = "https://mylib.example.com"
    topics = ("calculator", "math", "cpp")

    # Configuration
    settings = "os", "compiler", "build_type", "arch"
    options = {
        "shared": [True, False],
        "fPIC": [True, False],
        "with_examples": [True, False],
    }
    default_options = {
        "shared": False,
        "fPIC": True,
        "with_examples": False,
    }

    # Sources
    exports_sources = "CMakeLists.txt", "src/*", "include/*", "cmake/*"

    def config_options(self):
        if self.settings.os == "Windows":
            self.options.rm_safe("fPIC")

    def configure(self):
        if self.options.shared:
            self.options.rm_safe("fPIC")

    def layout(self):
        cmake_layout(self)

    def requirements(self):
        self.requires("fmt/9.1.0")
        if self.options.with_examples:
            self.requires("spdlog/1.11.0")

    def build_requirements(self):
        self.test_requires("catch2/3.2.1")

    def generate(self):
        tc = CMakeToolchain(self)
        tc.variables["BUILD_EXAMPLES"] = self.options.with_examples
        tc.generate()

        deps = CMakeDeps(self)
        deps.generate()

    def build(self):
        cmake = CMake(self)
        cmake.configure()
        cmake.build()

    def package(self):
        cmake = CMake(self)
        cmake.install()

        # Copy license
        copy(self, "LICENSE", src=self.source_folder, dst=os.path.join(self.package_folder, "licenses"))

    def package_info(self):
        self.cpp_info.libs = ["mylib"]
        self.cpp_info.includedirs = ["include"]

        if self.settings.os == "Linux":
            self.cpp_info.system_libs = ["m", "pthread"]
        elif self.settings.os == "Windows":
            self.cpp_info.system_libs = ["ws2_32"]

        # CMake integration
        self.cpp_info.names["cmake_find_package"] = "MyLib"
        self.cpp_info.names["cmake_find_package_multi"] = "MyLib"

        # pkg-config integration
        self.cpp_info.names["pkg_config"] = "mylib"

    def package_id(self):
        if self.info.options.with_examples:
            self.info.options.with_examples = "Any"
```

### vcpkg Port Configuration

```cmake
# ports/mylib/portfile.cmake
vcpkg_from_github(
    OUT_SOURCE_PATH SOURCE_PATH
    REPO user/mylib
    REF v1.0.0
    SHA512 <sha512-hash>
    HEAD_REF main
)

vcpkg_cmake_configure(
    SOURCE_PATH ${SOURCE_PATH}
    OPTIONS
        -DBUILD_EXAMPLES=OFF
        -DBUILD_TESTS=OFF
)

vcpkg_cmake_build()
vcpkg_cmake_install()

vcpkg_cmake_config_fixup(CONFIG_PATH lib/cmake/MyLib)

file(REMOVE_RECURSE "${CURRENT_PACKAGES_DIR}/debug/include")
file(REMOVE_RECURSE "${CURRENT_PACKAGES_DIR}/debug/share")

file(INSTALL "${SOURCE_PATH}/LICENSE" DESTINATION "${CURRENT_PACKAGES_DIR}/share/${PORT}" RENAME copyright)
```

```json
# ports/mylib/vcpkg.json
{
  "name": "mylib",
  "version": "1.0.0",
  "description": "A high-performance C++ calculator library",
  "homepage": "https://github.com/user/mylib",
  "license": "MIT",
  "dependencies": [
    {
      "name": "vcpkg-cmake",
      "host": true
    },
    {
      "name": "vcpkg-cmake-config",
      "host": true
    },
    "fmt"
  ],
  "features": {
    "examples": {
      "description": "Build examples",
      "dependencies": [
        "spdlog"
      ]
    }
  }
}
```

## Cross-Platform Distribution

### Universal Build Script

```bash
#!/bin/bash
# build-all-platforms.sh

set -e

PROJECT_NAME="mylib"
VERSION="1.0.0"
BUILD_DIR="build"
INSTALL_DIR="install"
PACKAGE_DIR="packages"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Platform detection
detect_platform() {
    case "$(uname -s)" in
        Linux*)     PLATFORM="linux";;
        Darwin*)    PLATFORM="macos";;
        CYGWIN*|MINGW32*|MSYS*|MINGW*) PLATFORM="windows";;
        *)          PLATFORM="unknown";;
    esac

    case "$(uname -m)" in
        x86_64)     ARCH="x64";;
        arm64|aarch64) ARCH="arm64";;
        armv7l)     ARCH="arm";;
        *)          ARCH="unknown";;
    esac

    echo "${PLATFORM}-${ARCH}"
}

# Build function
build_platform() {
    local platform=$1
    local build_type=${2:-Release}
    local target_triple=""

    log_info "Building for platform: $platform (Build type: $build_type)"

    # Platform-specific configuration
    case $platform in
        "linux-x64")
            target_triple="x86_64-linux-gnu"
            CMAKE_ARGS="-DCMAKE_C_COMPILER=gcc -DCMAKE_CXX_COMPILER=g++"
            ;;
        "linux-arm64")
            target_triple="aarch64-linux-gnu"
            CMAKE_ARGS="-DCMAKE_C_COMPILER=aarch64-linux-gnu-gcc -DCMAKE_CXX_COMPILER=aarch64-linux-gnu-g++"
            ;;
        "windows-x64")
            target_triple="x86_64-w64-mingw32"
            CMAKE_ARGS="-DCMAKE_TOOLCHAIN_FILE=cmake/toolchains/mingw-w64.cmake"
            ;;
        "macos-x64")
            target_triple="x86_64-apple-darwin"
            CMAKE_ARGS="-DCMAKE_OSX_ARCHITECTURES=x86_64"
            ;;
        "macos-arm64")
            target_triple="arm64-apple-darwin"
            CMAKE_ARGS="-DCMAKE_OSX_ARCHITECTURES=arm64"
            ;;
        *)
            log_error "Unsupported platform: $platform"
            return 1
            ;;
    esac

    # Create build directory
    local build_dir="${BUILD_DIR}-${platform}-${build_type,,}"
    local install_dir="${INSTALL_DIR}-${platform}-${build_type,,}"

    rm -rf "$build_dir" "$install_dir"
    mkdir -p "$build_dir" "$install_dir"

    # Install dependencies with Conan
    conan install . \
        --output-folder="$build_dir" \
        --build=missing \
        --settings=build_type="$build_type" \
        --settings=arch="$(echo $platform | cut -d'-' -f2)"

    # Configure
    cmake -B "$build_dir" -S . \
        -DCMAKE_BUILD_TYPE="$build_type" \
        -DCMAKE_INSTALL_PREFIX="$install_dir" \
        -DCMAKE_TOOLCHAIN_FILE="$build_dir/conan_toolchain.cmake" \
        -DBUILD_SHARED_LIBS=OFF \
        -DBUILD_TESTS=OFF \
        -DBUILD_EXAMPLES=OFF \
        $CMAKE_ARGS

    # Build
    cmake --build "$build_dir" --parallel 4

    # Install
    cmake --install "$build_dir"

    # Package
    package_build "$platform" "$build_type" "$install_dir"
}

# Package function
package_build() {
    local platform=$1
    local build_type=$2
    local install_dir=$3

    log_info "Packaging $platform build"

    mkdir -p "$PACKAGE_DIR"

    case $platform in
        windows-*)
            # Create zip package for Windows
            (cd "$install_dir" && zip -r "../$PACKAGE_DIR/${PROJECT_NAME}-${VERSION}-${platform}-${build_type,,}.zip" .)
            ;;
        *)
            # Create tar.gz package for Unix-like systems
            tar -czf "$PACKAGE_DIR/${PROJECT_NAME}-${VERSION}-${platform}-${build_type,,}.tar.gz" -C "$install_dir" .
            ;;
    esac
}

# Main execution
main() {
    local target_platform=${1:-$(detect_platform)}
    local build_type=${2:-Release}

    log_info "Starting build for $target_platform"

    # Clean previous packages
    rm -rf "$PACKAGE_DIR"
    mkdir -p "$PACKAGE_DIR"

    if [ "$target_platform" == "all" ]; then
        # Build for all supported platforms
        for platform in "linux-x64" "windows-x64" "macos-x64" "macos-arm64"; do
            build_platform "$platform" "$build_type"
        done
    else
        build_platform "$target_platform" "$build_type"
    fi

    log_info "Build completed. Packages available in $PACKAGE_DIR/"
    ls -la "$PACKAGE_DIR/"
}

# Usage information
usage() {
    echo "Usage: $0 [platform] [build_type]"
    echo "Platforms: linux-x64, linux-arm64, windows-x64, macos-x64, macos-arm64, all"
    echo "Build types: Debug, Release, RelWithDebInfo, MinSizeRel"
    echo "Default: $(detect_platform) Release"
}

# Command line parsing
if [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    usage
    exit 0
fi

main "$@"
```

### CMake Toolchain Files

```cmake
# cmake/toolchains/mingw-w64.cmake
set(CMAKE_SYSTEM_NAME Windows)
set(CMAKE_SYSTEM_PROCESSOR x86_64)

# Compiler configuration
set(CMAKE_C_COMPILER x86_64-w64-mingw32-gcc)
set(CMAKE_CXX_COMPILER x86_64-w64-mingw32-g++)
set(CMAKE_RC_COMPILER x86_64-w64-mingw32-windres)

# Root path
set(CMAKE_FIND_ROOT_PATH /usr/x86_64-w64-mingw32)

# Search settings
set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)
set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)

# Windows-specific flags
set(CMAKE_CXX_FLAGS "-static-libgcc -static-libstdc++")
set(CMAKE_EXE_LINKER_FLAGS "-static")
```

```cmake
# cmake/toolchains/arm64-linux.cmake
set(CMAKE_SYSTEM_NAME Linux)
set(CMAKE_SYSTEM_PROCESSOR aarch64)

set(CMAKE_C_COMPILER aarch64-linux-gnu-gcc)
set(CMAKE_CXX_COMPILER aarch64-linux-gnu-g++)

set(CMAKE_FIND_ROOT_PATH /usr/aarch64-linux-gnu)

set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)
set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)

set(CMAKE_C_FLAGS "-march=armv8-a")
set(CMAKE_CXX_FLAGS "-march=armv8-a")
```

## Container-Based Distribution

### Docker Multi-Stage Build

```dockerfile
# Dockerfile
# Multi-stage build for C++ application
ARG UBUNTU_VERSION=22.04
FROM ubuntu:${UBUNTU_VERSION} AS base

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    ninja-build \
    git \
    pkg-config \
    curl \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Install Conan
RUN pip3 install conan

# Development stage
FROM base AS development
WORKDIR /workspace

# Copy source code
COPY . .

# Configure Conan
RUN conan profile detect --force

# Install dependencies
RUN conan install . --output-folder=build --build=missing

# Configure and build
RUN cmake -B build -S . -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_TOOLCHAIN_FILE=build/conan_toolchain.cmake \
    -DBUILD_TESTS=ON \
    -DBUILD_EXAMPLES=ON

RUN cmake --build build --parallel $(nproc)

# Run tests
RUN cd build && ctest --output-on-failure

# Production build stage
FROM base AS builder
WORKDIR /src

COPY . .

RUN conan profile detect --force
RUN conan install . --output-folder=build --build=missing

RUN cmake -B build -S . -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_TOOLCHAIN_FILE=build/conan_toolchain.cmake \
    -DBUILD_TESTS=OFF \
    -DBUILD_EXAMPLES=OFF \
    -DCMAKE_INSTALL_PREFIX=/usr/local

RUN cmake --build build --parallel $(nproc)
RUN cmake --install build

# Runtime stage
FROM ubuntu:${UBUNTU_VERSION} AS runtime

# Install runtime dependencies only
RUN apt-get update && apt-get install -y \
    libstdc++6 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r myapp && useradd --no-log-init -r -g myapp myapp

# Copy application from builder stage
COPY --from=builder /usr/local/ /usr/local/
COPY --chown=myapp:myapp --from=builder /usr/local/bin/myapp /usr/local/bin/

USER myapp

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/myapp"]
CMD ["--config", "/etc/myapp/config.json"]
```

### Docker Compose for Development

```yaml
# docker-compose.yml
version: '3.8'

services:
  myapp:
    build:
      context: .
      dockerfile: Dockerfile
      target: development
    volumes:
      - .:/workspace
      - build-cache:/workspace/build
    environment:
      - CMAKE_BUILD_TYPE=Debug
    ports:
      - "8080:8080"
    command: ["tail", "-f", "/dev/null"]  # Keep container running

  myapp-prod:
    build:
      context: .
      dockerfile: Dockerfile
      target: runtime
    ports:
      - "8081:8080"
    environment:
      - LOG_LEVEL=info

volumes:
  build-cache:
```

### Kubernetes Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-deployment
  labels:
    app: myapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: myapp:1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: LOG_LEVEL
          value: "info"
        - name: CONFIG_PATH
          value: "/etc/myapp/config.json"
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: config-volume
          mountPath: /etc/myapp
      volumes:
      - name: config-volume
        configMap:
          name: myapp-config

---
apiVersion: v1
kind: Service
metadata:
  name: myapp-service
spec:
  selector:
    app: myapp
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: LoadBalancer

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
data:
  config.json: |
    {
      "server": {
        "port": 8080,
        "threads": 4
      },
      "logging": {
        "level": "info",
        "format": "json"
      }
    }
```

## Application Installers

### Windows Installer (NSIS)

```nsis
# installer.nsi
!include "MUI2.nsh"

Name "MyApp 1.0.0"
OutFile "MyApp-1.0.0-Setup.exe"
InstallDir "$PROGRAMFILES64\MyApp"

!define MUI_ABORTWARNING

# Installer pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE.txt"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

# Uninstaller pages
!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

!insertmacro MUI_LANGUAGE "English"

Section "Install"
    SetOutPath $INSTDIR

    # Main executable
    File "bin\myapp.exe"
    File "bin\*.dll"

    # Documentation
    File "README.md"
    File "LICENSE.txt"

    # Create shortcuts
    CreateDirectory "$SMPROGRAMS\MyApp"
    CreateShortcut "$SMPROGRAMS\MyApp\MyApp.lnk" "$INSTDIR\myapp.exe"
    CreateShortcut "$DESKTOP\MyApp.lnk" "$INSTDIR\myapp.exe"

    # Registry entries
    WriteRegStr HKLM "Software\MyApp" "InstallDir" $INSTDIR
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MyApp" "DisplayName" "MyApp 1.0.0"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MyApp" "UninstallString" "$INSTDIR\Uninstall.exe"

    # Create uninstaller
    WriteUninstaller "$INSTDIR\Uninstall.exe"
SectionEnd

Section "Uninstall"
    Delete "$INSTDIR\myapp.exe"
    Delete "$INSTDIR\*.dll"
    Delete "$INSTDIR\README.md"
    Delete "$INSTDIR\LICENSE.txt"
    Delete "$INSTDIR\Uninstall.exe"

    RMDir "$INSTDIR"

    Delete "$SMPROGRAMS\MyApp\MyApp.lnk"
    RMDir "$SMPROGRAMS\MyApp"
    Delete "$DESKTOP\MyApp.lnk"

    DeleteRegKey HKLM "Software\MyApp"
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MyApp"
SectionEnd
```

### macOS App Bundle

```bash
#!/bin/bash
# create-macos-bundle.sh

APP_NAME="MyApp"
VERSION="1.0.0"
BUNDLE_ID="com.example.myapp"
EXECUTABLE="myapp"

# Create app bundle structure
BUNDLE_DIR="${APP_NAME}.app"
CONTENTS_DIR="${BUNDLE_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"

mkdir -p "$MACOS_DIR" "$RESOURCES_DIR"

# Copy executable
cp "build/bin/$EXECUTABLE" "$MACOS_DIR/"

# Create Info.plist
cat > "${CONTENTS_DIR}/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>$EXECUTABLE</string>
    <key>CFBundleIdentifier</key>
    <string>$BUNDLE_ID</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundleVersion</key>
    <string>$VERSION</string>
    <key>CFBundleShortVersionString</key>
    <string>$VERSION</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleSignature</key>
    <string>????</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
</dict>
</plist>
EOF

# Copy resources
cp "resources/icon.icns" "$RESOURCES_DIR/" 2>/dev/null || echo "No icon found"

# Code signing (if certificates available)
if [ -n "$CODE_SIGN_IDENTITY" ]; then
    codesign --force --sign "$CODE_SIGN_IDENTITY" "$BUNDLE_DIR"
fi

# Create DMG
if command -v create-dmg &> /dev/null; then
    create-dmg \
        --volname "$APP_NAME Installer" \
        --window-size 500 300 \
        --icon-size 100 \
        --icon "$APP_NAME.app" 125 125 \
        --app-drop-link 375 125 \
        "${APP_NAME}-${VERSION}.dmg" \
        "$BUNDLE_DIR"
fi

echo "macOS app bundle created: $BUNDLE_DIR"
```

### Linux Package Creation

```bash
#!/bin/bash
# create-deb-package.sh

PACKAGE_NAME="myapp"
VERSION="1.0.0"
ARCH="amd64"

# Create package directory structure
PKG_DIR="${PACKAGE_NAME}_${VERSION}_${ARCH}"
mkdir -p "$PKG_DIR/DEBIAN"
mkdir -p "$PKG_DIR/usr/bin"
mkdir -p "$PKG_DIR/usr/share/applications"
mkdir -p "$PKG_DIR/usr/share/doc/$PACKAGE_NAME"

# Control file
cat > "$PKG_DIR/DEBIAN/control" << EOF
Package: $PACKAGE_NAME
Version: $VERSION
Section: utils
Priority: optional
Architecture: $ARCH
Depends: libstdc++6 (>= 10), libc6 (>= 2.31)
Maintainer: Your Name <email@example.com>
Description: MyApp - A sample C++ application
 This is a sample C++ application demonstrating
 packaging and distribution.
EOF

# Copy files
cp "build/bin/myapp" "$PKG_DIR/usr/bin/"
chmod 755 "$PKG_DIR/usr/bin/myapp"

# Desktop entry
cat > "$PKG_DIR/usr/share/applications/myapp.desktop" << EOF
[Desktop Entry]
Name=MyApp
Comment=A sample C++ application
Exec=/usr/bin/myapp
Icon=myapp
Terminal=false
Type=Application
Categories=Utility;
EOF

# Documentation
cp README.md "$PKG_DIR/usr/share/doc/$PACKAGE_NAME/"
cp LICENSE "$PKG_DIR/usr/share/doc/$PACKAGE_NAME/copyright"

# Build package
dpkg-deb --build "$PKG_DIR"

echo "Debian package created: ${PKG_DIR}.deb"

# RPM spec file for Red Hat-based distributions
cat > myapp.spec << EOF
Name: myapp
Version: 1.0.0
Release: 1%{?dist}
Summary: A sample C++ application

License: MIT
URL: https://github.com/user/myapp
Source0: %{name}-%{version}.tar.gz

BuildRequires: gcc-c++, cmake, ninja-build

%description
A sample C++ application demonstrating packaging and distribution.

%prep
%setup -q

%build
cmake -B build -S . -G Ninja -DCMAKE_BUILD_TYPE=Release
cmake --build build --parallel

%install
rm -rf %{buildroot}
cmake --install build --prefix %{buildroot}/usr

%files
%license LICENSE
%doc README.md
/usr/bin/myapp

%changelog
* Thu Jan 01 2024 Your Name <email@example.com> - 1.0.0-1
- Initial package
EOF
```

## Continuous Deployment

### GitHub Actions Release Workflow

```yaml
# .github/workflows/release.yml
name: Release and Deploy

on:
  push:
    tags: ['v*']

jobs:
  build-and-package:
    name: Build and Package (${{ matrix.os }})
    runs-on: ${{ matrix.os }}

    strategy:
      matrix:
        os: [ubuntu-22.04, windows-2022, macos-13]
        include:
          - os: ubuntu-22.04
            platform: linux
            package_cmd: ./scripts/create-deb-package.sh
          - os: windows-2022
            platform: windows
            package_cmd: makensis installer.nsi
          - os: macos-13
            platform: macos
            package_cmd: ./scripts/create-macos-bundle.sh

    steps:
    - uses: actions/checkout@v4

    - name: Build application
      run: |
        # Platform-specific build steps

    - name: Create package
      run: ${{ matrix.package_cmd }}

    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: ${{ matrix.platform }}-package
        path: |
          *.deb
          *.exe
          *.dmg

  deploy:
    needs: build-and-package
    runs-on: ubuntu-22.04

    steps:
    - name: Download artifacts
      uses: actions/download-artifact@v3

    - name: Create GitHub Release
      uses: softprops/action-gh-release@v1
      with:
        files: |
          **/*.deb
          **/*.exe
          **/*.dmg
        token: ${{ secrets.GITHUB_TOKEN }}

    - name: Deploy to package repositories
      run: |
        # Deploy to various package repositories
```

This comprehensive guide covers all aspects of C++ packaging and distribution, from library creation to cross-platform deployment strategies for modern C++ applications.