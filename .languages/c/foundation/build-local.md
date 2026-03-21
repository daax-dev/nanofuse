# C Local Build Systems and Toolchains

## GCC Toolchain

### Basic GCC Usage
```bash
# Basic compilation
gcc hello.c -o hello

# Compile with debugging information
gcc -g hello.c -o hello_debug

# Compile with optimization
gcc -O2 hello.c -o hello_optimized

# Compile with all warnings
gcc -Wall -Wextra hello.c -o hello

# Compile multiple source files
gcc main.c utils.c math_ops.c -o program

# Generate assembly output
gcc -S hello.c  # Creates hello.s

# Generate preprocessor output
gcc -E hello.c  # Outputs to stdout

# Compile to object files without linking
gcc -c main.c utils.c  # Creates main.o and utils.o

# Link object files
gcc main.o utils.o -o program
```

### Advanced GCC Options
```bash
# Static linking
gcc -static main.c -o program_static

# Shared library creation
gcc -shared -fPIC library.c -o liblibrary.so

# Link against shared library
gcc main.c -L./lib -llibrary -o program

# Cross-compilation
arm-linux-gnueabihf-gcc -o program_arm main.c

# Profile-guided optimization
gcc -fprofile-generate main.c -o program_profile
./program_profile < test_input
gcc -fprofile-use main.c -o program_optimized

# Link-time optimization
gcc -flto main.c utils.c -o program_lto

# Position-independent executable
gcc -fPIE -pie main.c -o program_pie

# Stack protection
gcc -fstack-protector-strong main.c -o program_secure

# Address sanitizer
gcc -fsanitize=address -g main.c -o program_asan

# Undefined behavior sanitizer
gcc -fsanitize=undefined -g main.c -o program_ubsan
```

### GCC Warning Flags
```bash
# Comprehensive warning set
gcc -Wall -Wextra -Wpedantic -Werror main.c -o program

# Specific useful warnings
gcc -Wformat=2 -Wconversion -Wsign-conversion -Wcast-qual \
    -Wcast-align -Wwrite-strings -Wfloat-equal \
    -Wpointer-arith -Wstrict-prototypes \
    main.c -o program

# Shadow variable warnings
gcc -Wshadow main.c -o program

# Unused variable/function warnings
gcc -Wunused main.c -o program

# Security-related warnings
gcc -Wformat-security -Wstack-protector main.c -o program
```

## Clang/LLVM Toolchain

### Basic Clang Usage
```bash
# Basic compilation (similar to GCC)
clang hello.c -o hello

# Generate LLVM IR
clang -emit-llvm -S hello.c  # Creates hello.ll

# Generate LLVM bitcode
clang -emit-llvm -c hello.c  # Creates hello.bc

# Static analysis
clang --analyze main.c

# Cross-compilation with explicit target
clang --target=arm-linux-gnueabihf main.c -o program_arm

# Use different C standards
clang -std=c11 main.c -o program
clang -std=c2x main.c -o program  # C23/C2x features
```

### Clang-Specific Features
```bash
# Clang static analyzer
clang --analyze -Xanalyzer -analyzer-output=html \
      -Xanalyzer -analyzer-output-dir=analysis \
      main.c

# Clang-tidy (static analysis tool)
clang-tidy main.c -- -I./include

# Clang-format (code formatting)
clang-format -i main.c  # Format in-place
clang-format -style=Google main.c  # Use Google style

# AddressSanitizer with better diagnostics
clang -fsanitize=address -fno-omit-frame-pointer -g main.c -o program

# MemorySanitizer
clang -fsanitize=memory -fno-omit-frame-pointer -g main.c -o program

# ThreadSanitizer
clang -fsanitize=thread -g main.c -pthread -o program

# UndefinedBehaviorSanitizer with specific checks
clang -fsanitize=undefined,nullability -g main.c -o program
```

## Microsoft Visual C++ (MSVC)

### MSVC Command Line
```cmd
REM Basic compilation
cl hello.c

REM Compile with debugging information
cl /Zi hello.c

REM Compile with optimization
cl /O2 hello.c

REM Compile with all warnings
cl /Wall hello.c

REM Compile multiple files
cl main.c utils.c /Fe:program.exe

REM Generate assembly listing
cl /Fa hello.c

REM Generate preprocessor output
cl /EP hello.c

REM Compile to object file
cl /c main.c

REM Link object files
link main.obj utils.obj /OUT:program.exe

REM Static linking to CRT
cl /MT main.c

REM Dynamic linking to CRT
cl /MD main.c

REM Address sanitizer (VS 2019+)
cl /fsanitize=address main.c
```

### MSVC Project Structure
```xml
<!-- project.vcxproj example -->
<?xml version="1.0" encoding="utf-8"?>
<Project DefaultTargets="Build" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
  <ItemGroup Label="ProjectConfigurations">
    <ProjectConfiguration Include="Debug|x64">
      <Configuration>Debug</Configuration>
      <Platform>x64</Platform>
    </ProjectConfiguration>
    <ProjectConfiguration Include="Release|x64">
      <Configuration>Release</Configuration>
      <Platform>x64</Platform>
    </ProjectConfiguration>
  </ItemGroup>

  <PropertyGroup Label="Globals">
    <ProjectGuid>{YOUR-GUID-HERE}</ProjectGuid>
    <RootNamespace>YourProject</RootNamespace>
  </PropertyGroup>

  <PropertyGroup Condition="'$(Configuration)|$(Platform)'=='Debug|x64'">
    <LinkIncremental>true</LinkIncremental>
  </PropertyGroup>

  <ItemDefinitionGroup Condition="'$(Configuration)|$(Platform)'=='Debug|x64'">
    <ClCompile>
      <WarningLevel>Level3</WarningLevel>
      <Optimization>Disabled</Optimization>
      <PreprocessorDefinitions>_DEBUG;%(PreprocessorDefinitions)</PreprocessorDefinitions>
    </ClCompile>
  </ItemDefinitionGroup>

  <ItemGroup>
    <ClCompile Include="main.c" />
    <ClCompile Include="utils.c" />
  </ItemGroup>

  <ItemGroup>
    <ClInclude Include="utils.h" />
  </ItemGroup>
</Project>
```

## Make Build System

### Basic Makefile
```makefile
# Variables
CC = gcc
CFLAGS = -Wall -Wextra -std=c11 -O2
LDFLAGS = -lm
TARGET = program
SRCDIR = src
OBJDIR = obj
SOURCES = $(wildcard $(SRCDIR)/*.c)
OBJECTS = $(SOURCES:$(SRCDIR)/%.c=$(OBJDIR)/%.o)

# Default target
all: $(TARGET)

# Link object files to create executable
$(TARGET): $(OBJECTS)
	$(CC) $(OBJECTS) -o $@ $(LDFLAGS)

# Compile source files to object files
$(OBJDIR)/%.o: $(SRCDIR)/%.c | $(OBJDIR)
	$(CC) $(CFLAGS) -c $< -o $@

# Create object directory
$(OBJDIR):
	mkdir -p $(OBJDIR)

# Clean build artifacts
clean:
	rm -rf $(OBJDIR) $(TARGET)

# Install target
install: $(TARGET)
	cp $(TARGET) /usr/local/bin/

# Debug build
debug: CFLAGS += -g -DDEBUG
debug: clean $(TARGET)

# Release build
release: CFLAGS += -DNDEBUG
release: clean $(TARGET)

# Static analysis
analyze:
	clang --analyze $(SOURCES)

# Format code
format:
	clang-format -i $(SOURCES) include/*.h

# Generate documentation
docs:
	doxygen Doxyfile

.PHONY: all clean install debug release analyze format docs
```

### Advanced Makefile Features
```makefile
# Automatic dependency generation
DEPDIR = deps
DEPFLAGS = -MT $@ -MMD -MP -MF $(DEPDIR)/$*.d

# Include dependency files
DEPFILES := $(SOURCES:$(SRCDIR)/%.c=$(DEPDIR)/%.d)
$(DEPFILES):
include $(wildcard $(DEPFILES))

# Compile with dependency generation
$(OBJDIR)/%.o: $(SRCDIR)/%.c $(DEPDIR)/%.d | $(OBJDIR) $(DEPDIR)
	$(CC) $(DEPFLAGS) $(CFLAGS) -c $< -o $@

$(DEPDIR):
	mkdir -p $(DEPDIR)

# Configuration detection
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    LDFLAGS += -pthread
    PLATFORM = linux
endif
ifeq ($(UNAME_S),Darwin)
    LDFLAGS += -framework CoreFoundation
    PLATFORM = macos
endif

# Compiler detection
ifneq (,$(findstring clang,$(CC)))
    CFLAGS += -Weverything -Wno-padded
else ifneq (,$(findstring gcc,$(CC)))
    CFLAGS += -Wcast-align -Wwrite-strings
endif

# Build variants
VARIANTS = debug release profile
debug: CFLAGS += -g -O0 -DDEBUG
release: CFLAGS += -O3 -DNDEBUG -flto
profile: CFLAGS += -g -O2 -pg

$(VARIANTS): $(TARGET)

# Parallel builds
MAKEFLAGS += -j$(shell nproc)

# Test target
test: $(TARGET)
	./tests/run_tests.sh $(TARGET)

# Coverage target
coverage: CFLAGS += --coverage
coverage: LDFLAGS += --coverage
coverage: clean $(TARGET)
	./$(TARGET)
	gcov $(SOURCES)
	lcov -c -d . -o coverage.info
	genhtml coverage.info -o coverage_html
```

## CMake Build System

### Basic CMakeLists.txt
```cmake
cmake_minimum_required(VERSION 3.10)
project(MyProject C)

# Set C standard
set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

# Add executable
add_executable(program
    src/main.c
    src/utils.c
    src/math_ops.c
)

# Add include directories
target_include_directories(program PRIVATE include)

# Link libraries
target_link_libraries(program m)  # Math library

# Compiler-specific options
if(CMAKE_C_COMPILER_ID STREQUAL "GNU")
    target_compile_options(program PRIVATE -Wall -Wextra)
elseif(CMAKE_C_COMPILER_ID STREQUAL "Clang")
    target_compile_options(program PRIVATE -Wall -Wextra -Weverything)
elseif(CMAKE_C_COMPILER_ID STREQUAL "MSVC")
    target_compile_options(program PRIVATE /W4)
endif()

# Build type specific settings
set(CMAKE_C_FLAGS_DEBUG "-g -O0 -DDEBUG")
set(CMAKE_C_FLAGS_RELEASE "-O3 -DNDEBUG")

# Install rules
install(TARGETS program DESTINATION bin)
install(FILES include/utils.h DESTINATION include)
```

### Advanced CMake Configuration
```cmake
cmake_minimum_required(VERSION 3.16)
project(AdvancedProject C)

# Options
option(BUILD_SHARED_LIBS "Build shared libraries" OFF)
option(ENABLE_TESTS "Build tests" ON)
option(ENABLE_SANITIZERS "Enable sanitizers" OFF)
option(ENABLE_STATIC_ANALYSIS "Enable static analysis" OFF)

# Find packages
find_package(PkgConfig REQUIRED)
pkg_check_modules(GLIB REQUIRED glib-2.0)

# Create library
add_library(mylib
    src/library.c
    src/utilities.c
)

target_include_directories(mylib
    PUBLIC include
    PRIVATE src
)

target_link_libraries(mylib ${GLIB_LIBRARIES})
target_include_directories(mylib PRIVATE ${GLIB_INCLUDE_DIRS})
target_compile_options(mylib PRIVATE ${GLIB_CFLAGS_OTHER})

# Create executable
add_executable(program src/main.c)
target_link_libraries(program mylib)

# Sanitizers
if(ENABLE_SANITIZERS)
    if(CMAKE_C_COMPILER_ID STREQUAL "GNU" OR CMAKE_C_COMPILER_ID STREQUAL "Clang")
        target_compile_options(program PRIVATE -fsanitize=address,undefined)
        target_link_options(program PRIVATE -fsanitize=address,undefined)
    endif()
endif()

# Static analysis
if(ENABLE_STATIC_ANALYSIS)
    find_program(CLANG_TIDY clang-tidy)
    if(CLANG_TIDY)
        set_target_properties(mylib PROPERTIES C_CLANG_TIDY "${CLANG_TIDY}")
        set_target_properties(program PROPERTIES C_CLANG_TIDY "${CLANG_TIDY}")
    endif()
endif()

# Testing
if(ENABLE_TESTS)
    enable_testing()
    add_executable(test_program tests/test_main.c)
    target_link_libraries(test_program mylib)
    add_test(NAME UnitTests COMMAND test_program)
endif()

# Custom targets
add_custom_target(format
    COMMAND clang-format -i src/*.c include/*.h
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    COMMENT "Formatting source code"
)

add_custom_target(docs
    COMMAND doxygen Doxyfile
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    COMMENT "Generating documentation"
)

# Installation
include(GNUInstallDirs)
install(TARGETS program mylib
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
    LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
    ARCHIVE DESTINATION ${CMAKE_INSTALL_LIBDIR}
)
install(FILES include/mylib.h DESTINATION ${CMAKE_INSTALL_INCLUDEDIR})

# Package configuration
include(CMakePackageConfigHelpers)
write_basic_package_version_file(
    mylibConfigVersion.cmake
    VERSION 1.0
    COMPATIBILITY SameMajorVersion
)

configure_package_config_file(
    mylibConfig.cmake.in
    mylibConfig.cmake
    INSTALL_DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/mylib
)

install(FILES
    ${CMAKE_CURRENT_BINARY_DIR}/mylibConfig.cmake
    ${CMAKE_CURRENT_BINARY_DIR}/mylibConfigVersion.cmake
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/mylib
)
```

### CMake Usage
```bash
# Basic build
mkdir build && cd build
cmake ..
make

# Debug build
cmake -DCMAKE_BUILD_TYPE=Debug ..
make

# Release build with custom options
cmake -DCMAKE_BUILD_TYPE=Release \
      -DENABLE_TESTS=ON \
      -DENABLE_SANITIZERS=ON \
      ..
make

# Install
make install

# Run tests
make test

# Generate packages
cpack

# Cross-compilation
cmake -DCMAKE_TOOLCHAIN_FILE=../toolchains/arm-linux.cmake ..
```

## Meson Build System

### Basic meson.build
```meson
project('myproject', 'c',
  version : '1.0',
  default_options : ['warning_level=3',
                     'c_std=c11'])

# Dependencies
cc = meson.get_compiler('c')
m_dep = cc.find_library('m', required : false)

# Sources
sources = [
  'src/main.c',
  'src/utils.c',
  'src/math_ops.c'
]

# Include directories
inc = include_directories('include')

# Executable
executable('program',
  sources,
  include_directories : inc,
  dependencies : m_dep,
  install : true)

# Library
mylib = library('mylib',
  ['src/library.c', 'src/utilities.c'],
  include_directories : inc,
  install : true)

# Tests
if get_option('enable_tests')
  test_exe = executable('test_program',
    'tests/test_main.c',
    include_directories : inc,
    link_with : mylib)

  test('basic test', test_exe)
endif

# Custom targets
run_target('format',
  command : ['clang-format', '-i', 'src/*.c', 'include/*.h'])
```

### Advanced Meson Features
```meson
project('advanced_project', 'c',
  version : '2.0',
  license : 'MIT',
  default_options : [
    'warning_level=3',
    'c_std=c11',
    'optimization=2',
    'debug=true'
  ])

# Configuration
conf_data = configuration_data()
conf_data.set('VERSION', '"@0@"'.format(meson.project_version()))
conf_data.set('PREFIX', '"@0@"'.format(get_option('prefix')))

configure_file(input : 'config.h.in',
               output : 'config.h',
               configuration : conf_data)

# Find dependencies
glib_dep = dependency('glib-2.0', version : '>= 2.50')
thread_dep = dependency('threads')

# Compiler checks
if cc.has_function('strlcpy')
  add_project_arguments('-DHAVE_STRLCPY', language : 'c')
endif

# Build options
sanitize = get_option('b_sanitize')
if sanitize != 'none'
  message('Building with sanitizer: ' + sanitize)
endif

# Subdirectories
subdir('src')
subdir('tests')
subdir('docs')

# Package configuration
pkg = import('pkgconfig')
pkg.generate(mylib,
  description : 'A useful C library',
  url : 'https://example.com/mylib')
```

### Meson Usage
```bash
# Setup build directory
meson setup builddir

# Compile
meson compile -C builddir

# Test
meson test -C builddir

# Install
meson install -C builddir

# Configure options
meson configure builddir -Denable_tests=true -Db_sanitize=address

# Cross-compilation
meson setup --cross-file cross/arm-linux.ini builddir-arm
```

## Bazel Build System

### Basic BUILD File
```python
# BUILD file
cc_binary(
    name = "program",
    srcs = ["src/main.c", "src/utils.c"],
    hdrs = ["include/utils.h"],
    includes = ["include"],
    linkopts = ["-lm"],
)

cc_library(
    name = "mylib",
    srcs = ["src/library.c"],
    hdrs = ["include/library.h"],
    includes = ["include"],
    visibility = ["//visibility:public"],
)

cc_test(
    name = "test_program",
    srcs = ["tests/test_main.c"],
    deps = [":mylib"],
)
```

### WORKSPACE File
```python
workspace(name = "myproject")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

# External dependencies
http_archive(
    name = "gtest",
    urls = ["https://github.com/google/googletest/archive/release-1.11.0.tar.gz"],
    strip_prefix = "googletest-release-1.11.0",
)
```

## Autotools (Autoconf/Automake)

### configure.ac
```bash
AC_INIT([myproject], [1.0], [support@example.com])
AM_INIT_AUTOMAKE([-Wall -Werror foreign])
AC_PROG_CC
AC_CONFIG_HEADERS([config.h])
AC_CONFIG_FILES([
 Makefile
 src/Makefile
])
AC_OUTPUT
```

### Makefile.am
```makefile
SUBDIRS = src

bin_PROGRAMS = program
program_SOURCES = src/main.c src/utils.c
program_CPPFLAGS = -Iinclude
program_LDFLAGS = -lm

# Library
lib_LTLIBRARIES = libmylib.la
libmylib_la_SOURCES = src/library.c

# Headers
include_HEADERS = include/mylib.h

# Tests
TESTS = test_program
check_PROGRAMS = test_program
test_program_SOURCES = tests/test_main.c
test_program_LDADD = libmylib.la
```

### Autotools Usage
```bash
# Generate configure script
autoreconf -fiv

# Configure and build
./configure
make

# Install
make install

# Run tests
make check

# Create distribution
make dist
```

## Build Tool Comparison

| Tool | Learning Curve | Speed | Features | Ecosystem |
|------|----------------|-------|----------|-----------|
| **Make** | Low | Fast | Basic | Universal |
| **CMake** | Medium | Medium | Comprehensive | Large |
| **Meson** | Low | Very Fast | Modern | Growing |
| **Bazel** | High | Very Fast | Advanced | Google ecosystem |
| **Autotools** | High | Slow | Portable | Legacy projects |

## Best Practices

### Project Structure
```
project/
├── src/           # Source files
├── include/       # Public headers
├── tests/         # Test files
├── docs/          # Documentation
├── scripts/       # Build/utility scripts
├── extern/        # External dependencies
├── CMakeLists.txt # Build files
├── Makefile
├── README.md
└── .gitignore
```

### Build Configuration
1. **Separate build directories** for different configurations
2. **Use compiler flags consistently** across all build systems
3. **Enable all relevant warnings** and treat warnings as errors
4. **Include static analysis** in build process
5. **Set up continuous integration** with multiple compilers
6. **Use sanitizers** in debug builds
7. **Generate and install pkg-config files** for libraries

The choice of build system depends on project requirements, team expertise, and integration needs. Make remains the most portable, while CMake provides the best balance of features and ecosystem support for most C projects.