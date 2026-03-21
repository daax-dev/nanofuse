# C Packaging and Distribution Guide

## Static vs Dynamic Libraries

### Creating Static Libraries
```bash
# Compile source files to object files
gcc -c -Wall -Wextra -O2 src/math_utils.c -o math_utils.o
gcc -c -Wall -Wextra -O2 src/string_utils.c -o string_utils.o
gcc -c -Wall -Wextra -O2 src/file_utils.c -o file_utils.o

# Create static library (.a file)
ar rcs libmylib.a math_utils.o string_utils.o file_utils.o

# Verify library contents
ar -t libmylib.a

# Link against static library
gcc main.c -L. -lmylib -o program

# Alternative: direct linking
gcc main.c libmylib.a -o program
```

### Creating Shared Libraries
```bash
# Compile with position-independent code (PIC)
gcc -c -fPIC -Wall -Wextra -O2 src/math_utils.c -o math_utils.o
gcc -c -fPIC -Wall -Wextra -O2 src/string_utils.c -o string_utils.o
gcc -c -fPIC -Wall -Wextra -O2 src/file_utils.c -o file_utils.o

# Create shared library (.so file on Linux, .dylib on macOS, .dll on Windows)
gcc -shared -Wl,-soname,libmylib.so.1 -o libmylib.so.1.0.0 \
    math_utils.o string_utils.o file_utils.o

# Create symlinks for versioning
ln -sf libmylib.so.1.0.0 libmylib.so.1
ln -sf libmylib.so.1 libmylib.so

# Link against shared library
gcc main.c -L. -lmylib -o program

# Run with library path
LD_LIBRARY_PATH=. ./program

# Install to system directories
sudo cp libmylib.so.1.0.0 /usr/local/lib/
sudo ldconfig  # Update library cache
```

### Library Makefile Example
```makefile
# Library Makefile
CC = gcc
CFLAGS = -Wall -Wextra -O2 -fPIC
LDFLAGS = -shared

SRCDIR = src
INCDIR = include
OBJDIR = obj
LIBDIR = lib

SOURCES = $(wildcard $(SRCDIR)/*.c)
OBJECTS = $(SOURCES:$(SRCDIR)/%.c=$(OBJDIR)/%.o)
HEADERS = $(wildcard $(INCDIR)/*.h)

LIBRARY_NAME = mylib
STATIC_LIB = $(LIBDIR)/lib$(LIBRARY_NAME).a
SHARED_LIB = $(LIBDIR)/lib$(LIBRARY_NAME).so
VERSION = 1.0.0
SHARED_LIB_VERSIONED = $(SHARED_LIB).$(VERSION)

.PHONY: all static shared clean install

all: static shared

static: $(STATIC_LIB)

shared: $(SHARED_LIB_VERSIONED)

$(STATIC_LIB): $(OBJECTS) | $(LIBDIR)
	ar rcs $@ $^

$(SHARED_LIB_VERSIONED): $(OBJECTS) | $(LIBDIR)
	$(CC) $(LDFLAGS) -Wl,-soname,lib$(LIBRARY_NAME).so.1 -o $@ $^
	cd $(LIBDIR) && ln -sf lib$(LIBRARY_NAME).so.$(VERSION) lib$(LIBRARY_NAME).so.1
	cd $(LIBDIR) && ln -sf lib$(LIBRARY_NAME).so.1 lib$(LIBRARY_NAME).so

$(OBJDIR)/%.o: $(SRCDIR)/%.c | $(OBJDIR)
	$(CC) $(CFLAGS) -I$(INCDIR) -c $< -o $@

$(OBJDIR):
	mkdir -p $(OBJDIR)

$(LIBDIR):
	mkdir -p $(LIBDIR)

clean:
	rm -rf $(OBJDIR) $(LIBDIR)

install: all
	install -d $(DESTDIR)/usr/local/lib
	install -d $(DESTDIR)/usr/local/include/$(LIBRARY_NAME)
	install -m 755 $(SHARED_LIB_VERSIONED) $(DESTDIR)/usr/local/lib/
	install -m 644 $(STATIC_LIB) $(DESTDIR)/usr/local/lib/
	install -m 644 $(HEADERS) $(DESTDIR)/usr/local/include/$(LIBRARY_NAME)/
	ldconfig
```

## pkg-config Integration

### Creating .pc Files
```ini
# mylib.pc.in template
prefix=@PREFIX@
exec_prefix=${prefix}
libdir=${exec_prefix}/lib
includedir=${prefix}/include

Name: MyLib
Description: A useful C library for various operations
Version: @VERSION@
Requires:
Conflicts:
Libs: -L${libdir} -lmylib
Libs.private: -lm
Cflags: -I${includedir}
```

### CMake pkg-config Generation
```cmake
# CMakeLists.txt pkg-config support
configure_file(${CMAKE_CURRENT_SOURCE_DIR}/mylib.pc.in
               ${CMAKE_CURRENT_BINARY_DIR}/mylib.pc
               @ONLY)

install(FILES ${CMAKE_CURRENT_BINARY_DIR}/mylib.pc
        DESTINATION ${CMAKE_INSTALL_LIBDIR}/pkgconfig)
```

### Using pkg-config
```bash
# Check if library is installed
pkg-config --exists mylib

# Get compilation flags
pkg-config --cflags mylib

# Get linking flags
pkg-config --libs mylib

# Get both
pkg-config --cflags --libs mylib

# Use in Makefile
CFLAGS += $(shell pkg-config --cflags mylib)
LDFLAGS += $(shell pkg-config --libs mylib)

# Use in build
gcc $(shell pkg-config --cflags mylib) main.c \
    $(shell pkg-config --libs mylib) -o program
```

## Package Managers

### vcpkg (Microsoft C/C++ Package Manager)
```json
# vcpkg.json
{
    "name": "myproject",
    "version": "1.0.0",
    "dependencies": [
        "curl",
        "json-c",
        "sqlite3",
        {
            "name": "openssl",
            "features": ["tools"]
        }
    ]
}
```

```cmake
# CMakeLists.txt with vcpkg
find_package(PkgConfig REQUIRED)
pkg_check_modules(CURL REQUIRED libcurl)
pkg_check_modules(JSON_C REQUIRED json-c)

find_package(SQLite3 REQUIRED)
find_package(OpenSSL REQUIRED)

target_link_libraries(myproject
    ${CURL_LIBRARIES}
    ${JSON_C_LIBRARIES}
    SQLite3::SQLite3
    OpenSSL::SSL
    OpenSSL::Crypto
)
```

### Conan Package Manager
```ini
# conanfile.txt
[requires]
zlib/1.2.13
openssl/1.1.1s
sqlite3/3.40.0

[generators]
cmake_find_package

[options]
openssl:shared=False
```

```python
# conanfile.py
from conan import ConanFile
from conan.tools.cmake import cmake_deps, cmake_layout, CMakeToolchain

class MyProjectConan(ConanFile):
    settings = "os", "compiler", "build_type", "arch"
    requires = "zlib/1.2.13", "openssl/1.1.1s"
    generators = "CMakeDeps", "CMakeToolchain"

    def layout(self):
        cmake_layout(self)

    def build_requirements(self):
        self.tool_requires("cmake/3.25.0")
```

### Linux Distribution Packaging

#### Debian/Ubuntu (.deb)
```bash
# debian/control
Source: mylib
Section: libs
Priority: optional
Maintainer: Your Name <your.email@example.com>
Build-Depends: debhelper (>= 9), cmake, gcc
Standards-Version: 3.9.8

Package: libmylib1
Architecture: any
Depends: ${shlibs:Depends}, ${misc:Depends}
Description: MyLib - A useful C library
 This package contains the shared library for MyLib.

Package: libmylib-dev
Architecture: any
Section: libdevel
Depends: libmylib1 (= ${binary:Version}), ${misc:Depends}
Description: Development files for MyLib
 This package contains the header files and static library
 needed to compile applications that use MyLib.
```

```bash
# debian/rules
#!/usr/bin/make -f
%:
	dh $@ --buildsystem=cmake

override_dh_auto_configure:
	dh_auto_configure -- -DCMAKE_BUILD_TYPE=Release
```

```bash
# Build package
dpkg-buildpackage -us -uc
```

#### RPM (Red Hat/CentOS/SUSE)
```spec
# mylib.spec
Name:           mylib
Version:        1.0.0
Release:        1%{?dist}
Summary:        A useful C library

License:        MIT
URL:            https://github.com/user/mylib
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  gcc cmake make
Requires:       glibc-devel

%description
MyLib is a useful C library that provides various utility functions.

%package devel
Summary:        Development files for %{name}
Requires:       %{name}%{?_isa} = %{version}-%{release}

%description devel
The %{name}-devel package contains libraries and header files for
developing applications that use %{name}.

%prep
%setup -q

%build
mkdir build
cd build
%cmake ..
%make_build

%install
cd build
%make_install

%files
%license LICENSE
%doc README.md
%{_libdir}/libmylib.so.*

%files devel
%{_includedir}/mylib/
%{_libdir}/libmylib.so
%{_libdir}/pkgconfig/mylib.pc

%changelog
* Tue Jan 01 2024 Your Name <your.email@example.com> - 1.0.0-1
- Initial package
```

### Homebrew Formula (macOS)
```ruby
# Formula/mylib.rb
class Mylib < Formula
  desc "A useful C library for various operations"
  homepage "https://github.com/user/mylib"
  url "https://github.com/user/mylib/archive/v1.0.0.tar.gz"
  sha256 "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
  license "MIT"

  depends_on "cmake" => :build

  def install
    system "cmake", ".", *std_cmake_args
    system "make", "install"
  end

  test do
    system "#{bin}/mylib-test"
  end
end
```

## Distribution Formats

### Source Distribution (Tarball)
```bash
# Create source distribution
make dist

# Or manually
PROJECT_NAME="mylib"
VERSION="1.0.0"

# Create archive
git archive --format=tar.gz --prefix=${PROJECT_NAME}-${VERSION}/ \
    HEAD > ${PROJECT_NAME}-${VERSION}.tar.gz

# Verify archive
tar -tzf ${PROJECT_NAME}-${VERSION}.tar.gz | head -20
```

### Binary Distribution
```bash
# Create binary package structure
mkdir -p mylib-1.0.0-linux-x64/{bin,lib,include,share/doc}

# Copy files
cp program mylib-1.0.0-linux-x64/bin/
cp libmylib.so* mylib-1.0.0-linux-x64/lib/
cp include/*.h mylib-1.0.0-linux-x64/include/
cp README.md LICENSE mylib-1.0.0-linux-x64/share/doc/

# Create archive
tar -czf mylib-1.0.0-linux-x64.tar.gz mylib-1.0.0-linux-x64/
```

### Windows Distribution
```batch
REM Create installer with NSIS
; installer.nsi
!define PRODUCT_NAME "MyLib"
!define PRODUCT_VERSION "1.0.0"
!define PRODUCT_PUBLISHER "Your Company"

Name "${PRODUCT_NAME} ${PRODUCT_VERSION}"
OutFile "mylib-${PRODUCT_VERSION}-setup.exe"
InstallDir "$PROGRAMFILES\${PRODUCT_NAME}"

Section "Main"
  SetOutPath "$INSTDIR\bin"
  File "bin\mylib.dll"
  File "bin\mylib-test.exe"

  SetOutPath "$INSTDIR\include"
  File "include\*.h"

  SetOutPath "$INSTDIR\lib"
  File "lib\mylib.lib"

  WriteRegStr HKLM "Software\${PRODUCT_NAME}" "" $INSTDIR
  WriteUninstaller "$INSTDIR\uninstall.exe"
SectionEnd
```

## CMake Packaging Support

### CPack Configuration
```cmake
# CMakeLists.txt packaging
include(CPack)

set(CPACK_PACKAGE_NAME "mylib")
set(CPACK_PACKAGE_VENDOR "Your Company")
set(CPACK_PACKAGE_VERSION_MAJOR ${PROJECT_VERSION_MAJOR})
set(CPACK_PACKAGE_VERSION_MINOR ${PROJECT_VERSION_MINOR})
set(CPACK_PACKAGE_VERSION_PATCH ${PROJECT_VERSION_PATCH})
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "A useful C library")
set(CPACK_PACKAGE_DESCRIPTION_FILE "${CMAKE_CURRENT_SOURCE_DIR}/README.md")
set(CPACK_RESOURCE_FILE_LICENSE "${CMAKE_CURRENT_SOURCE_DIR}/LICENSE")

# Source package
set(CPACK_SOURCE_GENERATOR "TGZ;ZIP")
set(CPACK_SOURCE_IGNORE_FILES "/build/;/.git/;/.gitignore")

# Binary packages
if(WIN32)
    set(CPACK_GENERATOR "NSIS;ZIP")
elseif(APPLE)
    set(CPACK_GENERATOR "DragNDrop;TGZ")
else()
    set(CPACK_GENERATOR "DEB;RPM;TGZ")
endif()

# Debian-specific
set(CPACK_DEBIAN_PACKAGE_MAINTAINER "your.email@example.com")
set(CPACK_DEBIAN_PACKAGE_DEPENDS "libc6")

# RPM-specific
set(CPACK_RPM_PACKAGE_LICENSE "MIT")
set(CPACK_RPM_PACKAGE_GROUP "Development/Libraries")
```

### CMake Package Configuration
```cmake
# CMakeListst.txt - Package config files
include(CMakePackageConfigHelpers)

write_basic_package_version_file(
    "${CMAKE_CURRENT_BINARY_DIR}/mylibConfigVersion.cmake"
    VERSION ${PROJECT_VERSION}
    COMPATIBILITY SameMajorVersion
)

configure_package_config_file(
    "${CMAKE_CURRENT_SOURCE_DIR}/mylibConfig.cmake.in"
    "${CMAKE_CURRENT_BINARY_DIR}/mylibConfig.cmake"
    INSTALL_DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/mylib
)

install(FILES
    "${CMAKE_CURRENT_BINARY_DIR}/mylibConfig.cmake"
    "${CMAKE_CURRENT_BINARY_DIR}/mylibConfigVersion.cmake"
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/mylib
)

install(EXPORT mylibTargets
    FILE mylibTargets.cmake
    NAMESPACE mylib::
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/mylib
)
```

```cmake
# mylibConfig.cmake.in
@PACKAGE_INIT@

include(CMakeFindDependencyMacro)
find_dependency(Threads)

include("${CMAKE_CURRENT_LIST_DIR}/mylibTargets.cmake")

check_required_components(mylib)
```

## Container Packaging

### Docker Multi-stage Build
```dockerfile
# Dockerfile
FROM gcc:11 as builder

WORKDIR /app
COPY . .

RUN make clean && make CFLAGS="-O3 -static"

# Runtime image
FROM scratch
COPY --from=builder /app/program /program
ENTRYPOINT ["/program"]
```

### Alpine-based Minimal Image
```dockerfile
FROM alpine:latest as builder

RUN apk add --no-cache gcc musl-dev make cmake

WORKDIR /app
COPY . .
RUN cmake . && make

FROM alpine:latest
RUN apk add --no-cache libc6-compat
COPY --from=builder /app/program /usr/local/bin/
CMD ["program"]
```

## Continuous Integration Packaging

### GitHub Actions Release
```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  create-release:
    runs-on: ubuntu-latest
    outputs:
      upload_url: ${{ steps.create_release.outputs.upload_url }}

    steps:
    - uses: actions/checkout@v4

    - name: Create Release
      id: create_release
      uses: actions/create-release@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        tag_name: ${{ github.ref }}
        release_name: Release ${{ github.ref }}
        draft: false
        prerelease: false

  build-and-upload:
    needs: create-release
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]

    runs-on: ${{ matrix.os }}

    steps:
    - uses: actions/checkout@v4

    - name: Build
      run: |
        mkdir build && cd build
        cmake .. -DCMAKE_BUILD_TYPE=Release
        cmake --build . --config Release
        cpack

    - name: Upload Release Asset
      uses: actions/upload-release-asset@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        upload_url: ${{ needs.create-release.outputs.upload_url }}
        asset_path: build/*.tar.gz
        asset_name: mylib-${{ matrix.os }}.tar.gz
        asset_content_type: application/gzip
```

## Best Practices

### Library Design Principles
1. **ABI Stability**: Maintain backward compatibility
2. **Versioning**: Use semantic versioning (MAJOR.MINOR.PATCH)
3. **Header-only vs Compiled**: Consider distribution complexity
4. **Dependencies**: Minimize external dependencies
5. **Platform Support**: Test on target platforms

### Package Structure
```
mylib/
├── CMakeLists.txt
├── LICENSE
├── README.md
├── CHANGELOG.md
├── include/
│   └── mylib/
│       ├── mylib.h
│       └── internal.h
├── src/
│   ├── mylib.c
│   └── internal.c
├── tests/
│   ├── test_main.c
│   └── CMakeLists.txt
├── examples/
│   ├── basic_usage.c
│   └── advanced_usage.c
├── docs/
│   ├── API.md
│   └── BUILDING.md
├── packaging/
│   ├── debian/
│   ├── rpm/
│   └── homebrew/
└── scripts/
    ├── build.sh
    └── install.sh
```

### Installation Layout
```
/usr/local/
├── bin/                    # Executables
├── lib/                    # Shared libraries
│   ├── libmylib.so -> libmylib.so.1
│   ├── libmylib.so.1 -> libmylib.so.1.0.0
│   ├── libmylib.so.1.0.0
│   ├── libmylib.a         # Static library
│   └── pkgconfig/
│       └── mylib.pc
├── include/                # Headers
│   └── mylib/
│       ├── mylib.h
│       └── types.h
└── share/
    ├── doc/mylib/         # Documentation
    └── man/man3/          # Man pages
```

### Testing Package Installation
```bash
#!/bin/bash
# test_package.sh

set -e

PACKAGE_NAME="mylib"
TEST_DIR=$(mktemp -d)

echo "Testing package installation..."

# Test pkg-config
if pkg-config --exists ${PACKAGE_NAME}; then
    echo "✓ pkg-config file found"
else
    echo "✗ pkg-config file not found"
    exit 1
fi

# Test compilation
cd ${TEST_DIR}
cat > test.c << EOF
#include <mylib/mylib.h>
int main() { return mylib_version(); }
EOF

if gcc $(pkg-config --cflags ${PACKAGE_NAME}) test.c \
       $(pkg-config --libs ${PACKAGE_NAME}) -o test; then
    echo "✓ Compilation successful"
else
    echo "✗ Compilation failed"
    exit 1
fi

# Test execution
if ./test; then
    echo "✓ Execution successful"
else
    echo "✗ Execution failed"
    exit 1
fi

rm -rf ${TEST_DIR}
echo "Package installation test passed!"
```

Effective C packaging requires careful consideration of distribution methods, target platforms, and user needs. The key is to provide multiple distribution formats while maintaining consistency and ease of installation.