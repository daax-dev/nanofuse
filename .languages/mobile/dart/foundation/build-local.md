# Local Flutter/Dart Development Setup

> Comprehensive guide to setting up a complete local development environment for Dart programming and Flutter mobile development.

## Development Environment Overview

A proper Flutter/Dart development setup requires multiple tools and SDKs working together. This guide covers installation, configuration, and optimization for mobile development productivity.

### Core Requirements
- **Flutter SDK**: Framework and tooling
- **Dart SDK**: Included with Flutter
- **Platform SDKs**: Android Studio/Xcode for mobile targets
- **IDE/Editor**: VS Code, Android Studio, or IntelliJ IDEA
- **Version Control**: Git integration

---

## System Requirements

### Minimum Hardware Requirements

```yaml
# Minimum specifications for Flutter development
CPU: 64-bit processor (x64 or ARM64)
RAM: 8 GB (16 GB recommended)
Storage: 15 GB free space (SSD recommended)
Display: 1280x800 minimum resolution

# Platform-specific requirements
Windows: Windows 10 or later (64-bit)
macOS: macOS 10.14 or later
Linux: Ubuntu 18.04+ or equivalent
```

### Recommended Hardware Configuration

```yaml
# Optimal development setup
CPU: Multi-core processor (Intel i7/AMD Ryzen 7 or Apple M1/M2)
RAM: 16-32 GB
Storage: 500+ GB SSD
Display: 2K+ resolution, dual monitor setup
Network: Stable broadband connection
```

---

## Flutter SDK Installation

### Windows Installation

```powershell
# Method 1: Direct download
# Download Flutter SDK from https://docs.flutter.dev/get-started/install/windows

# Method 2: Using Git
git clone https://github.com/flutter/flutter.git -b stable
cd flutter

# Add to PATH (PowerShell)
$env:PATH += ";C:\flutter\bin"

# Permanent PATH addition (add to System Environment Variables)
# C:\flutter\bin

# Verify installation
flutter doctor
```

### macOS Installation

```bash
# Method 1: Using Homebrew (recommended)
brew install --cask flutter

# Method 2: Direct download and extract
curl -o flutter_macos.zip https://storage.googleapis.com/flutter_infra_release/releases/stable/macos/flutter_macos_3.24.0-stable.zip
unzip flutter_macos.zip
sudo mv flutter /usr/local/

# Add to PATH
echo 'export PATH="$PATH:/usr/local/flutter/bin"' >> ~/.zshrc
source ~/.zshrc

# Verify installation
flutter doctor
```

### Linux Installation

```bash
# Download and extract Flutter
wget https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_3.24.0-stable.tar.xz
tar xf flutter_linux_3.24.0-stable.tar.xz
sudo mv flutter /opt/

# Add to PATH
echo 'export PATH="$PATH:/opt/flutter/bin"' >> ~/.bashrc
source ~/.bashrc

# Install dependencies
sudo apt-get update
sudo apt-get install curl git unzip xz-utils zip libglu1-mesa

# Verify installation
flutter doctor
```

---

## IDE and Editor Setup

### Visual Studio Code Setup

```bash
# Install VS Code extensions
code --install-extension Dart-Code.dart-code
code --install-extension Dart-Code.flutter
code --install-extension ms-vscode.vscode-json
code --install-extension bradlc.vscode-tailwindcss
code --install-extension PKief.material-icon-theme
code --install-extension formulahendry.auto-rename-tag
```

#### VS Code Configuration

```json
// .vscode/settings.json
{
  "dart.flutterSdkPath": "/path/to/flutter",
  "dart.previewFlutterUiGuides": true,
  "dart.previewFlutterUiGuidesCustomTracking": true,
  "dart.debugExtensionBackendProtocol": "sse",
  "dart.debugExternalPackageLibraries": false,
  "dart.debugSdkLibraries": false,
  "dart.analysisExcludedFolders": [
    ".dart_tool",
    "build"
  ],
  "dart.lineLength": 100,
  "editor.formatOnSave": true,
  "editor.formatOnType": true,
  "editor.rulers": [80, 100],
  "files.associations": {
    "*.dart": "dart"
  },
  "emmet.includeLanguages": {
    "dart": "html"
  }
}
```

#### VS Code Launch Configuration

```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Flutter (Debug)",
      "type": "dart",
      "request": "launch",
      "program": "lib/main.dart",
      "flutterMode": "debug"
    },
    {
      "name": "Flutter (Profile)",
      "type": "dart",
      "request": "launch",
      "program": "lib/main.dart",
      "flutterMode": "profile"
    },
    {
      "name": "Flutter (Release)",
      "type": "dart",
      "request": "launch",
      "program": "lib/main.dart",
      "flutterMode": "release"
    },
    {
      "name": "Flutter Tests",
      "type": "dart",
      "request": "launch",
      "program": "test/"
    }
  ]
}
```

### Android Studio Setup

```bash
# Install Android Studio from https://developer.android.com/studio

# Required plugins (install via Preferences > Plugins):
# - Flutter
# - Dart
# - Git Integration
# - Material Theme UI (optional)
```

#### Android Studio Configuration

```properties
# studio.vmoptions (increase memory for better performance)
-Xms2g
-Xmx8g
-XX:ReservedCodeCacheSize=1g
-XX:+UseConcMarkSweepGC
-XX:SoftRefLRUPolicyMSPerMB=50
-Dsun.io.useCanonPrefixCache=false
-Djdk.http.auth.tunneling.disabledSchemes=""
-Djsse.enableSNIExtension=false
-XX:+UseCompressedOops
-Dfile.encoding=UTF-8
-XX:+UseConcMarkSweepGC
-XX:+CMSClassUnloadingEnabled
-XX:+CMSPermGenSweepingEnabled
```

### IntelliJ IDEA Setup

```bash
# Install IntelliJ IDEA Community or Ultimate
# Install plugins:
# - Flutter
# - Dart
# - Git Integration
```

---

## Platform SDK Setup

### Android Development Setup

```bash
# Install Android Studio (includes Android SDK)
# Or install Android SDK separately

# Required SDK components:
# - Android SDK Platform-Tools
# - Android SDK Build-Tools
# - Android SDK Platform (API 34+)
# - Android Emulator
# - Intel x86 Emulator Accelerator (HAXM) - Intel only
# - Google USB Driver (Windows only)
```

#### Android SDK Configuration

```bash
# Set Android SDK environment variables
export ANDROID_HOME=/path/to/Android/Sdk
export PATH=$PATH:$ANDROID_HOME/platform-tools
export PATH=$PATH:$ANDROID_HOME/tools
export PATH=$PATH:$ANDROID_HOME/tools/bin
export PATH=$PATH:$ANDROID_HOME/emulator

# Windows (PowerShell)
$env:ANDROID_HOME = "C:\Users\YourName\AppData\Local\Android\Sdk"
$env:PATH += ";$env:ANDROID_HOME\platform-tools"
$env:PATH += ";$env:ANDROID_HOME\tools"
$env:PATH += ";$env:ANDROID_HOME\emulator"

# Accept Android licenses
flutter doctor --android-licenses
```

#### Android Emulator Setup

```bash
# Create Android emulator via Android Studio:
# Tools > AVD Manager > Create Virtual Device

# Command line emulator creation
avdmanager create avd -n "Pixel_7_API_34" -k "system-images;android-34;google_apis;x86_64"

# List available emulators
flutter emulators

# Launch emulator
flutter emulators --launch Pixel_7_API_34

# Or use emulator command
emulator -avd Pixel_7_API_34
```

### iOS Development Setup (macOS only)

```bash
# Install Xcode from App Store
# Install Xcode Command Line Tools
xcode-select --install

# Accept Xcode license
sudo xcodebuild -license accept

# Install CocoaPods for iOS dependencies
sudo gem install cocoapods

# Or using Homebrew
brew install cocoapods

# Verify Xcode setup
flutter doctor
```

#### iOS Simulator Setup

```bash
# Open iOS Simulator
open -a Simulator

# List available iOS simulators
xcrun simctl list devices

# Create new iOS simulator
xcrun simctl create "iPhone 15 Pro" "iPhone 15 Pro" "17.0"

# Launch specific simulator
xcrun simctl boot "iPhone 15 Pro"
```

---

## Development Tools Setup

### Git Configuration

```bash
# Configure Git (required for Flutter)
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Set up SSH key for GitHub/GitLab
ssh-keygen -t rsa -b 4096 -C "your.email@example.com"

# Add SSH key to ssh-agent
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_rsa

# Copy public key to clipboard (macOS)
pbcopy < ~/.ssh/id_rsa.pub

# Copy public key to clipboard (Linux)
xclip -selection clipboard < ~/.ssh/id_rsa.pub

# Copy public key to clipboard (Windows)
clip < ~/.ssh/id_rsa.pub
```

### Chrome/Edge for Flutter Web

```bash
# Chrome is required for Flutter web development
# Install Google Chrome or Microsoft Edge

# Set Chrome as default for Flutter web debugging
flutter config --enable-web
```

---

## Project Setup and Structure

### Creating a New Flutter Project

```bash
# Create new Flutter app
flutter create my_awesome_app
cd my_awesome_app

# Create project with specific organization
flutter create --org com.example my_awesome_app

# Create project with platform support
flutter create --platforms android,ios,web my_awesome_app

# Create package
flutter create --template package my_package

# Create plugin
flutter create --template plugin my_plugin
```

### Project Structure Best Practices

```
my_flutter_app/
├── android/                 # Android-specific files
├── ios/                     # iOS-specific files
├── web/                     # Web-specific files (if enabled)
├── lib/                     # Dart source code
│   ├── main.dart           # App entry point
│   ├── app/                # App-level configuration
│   ├── core/               # Core functionality
│   │   ├── constants/
│   │   ├── errors/
│   │   ├── network/
│   │   └── utils/
│   ├── features/           # Feature-based modules
│   │   └── auth/
│   │       ├── data/       # Data layer
│   │       ├── domain/     # Business logic
│   │       └── presentation/ # UI layer
│   ├── shared/             # Shared components
│   │   ├── widgets/
│   │   ├── models/
│   │   └── services/
│   └── gen/                # Generated code
├── test/                   # Unit tests
├── integration_test/       # Integration tests
├── assets/                 # Images, fonts, data files
│   ├── images/
│   ├── icons/
│   └── fonts/
├── pubspec.yaml           # Dependencies and configuration
├── pubspec.lock           # Locked dependency versions
├── analysis_options.yaml  # Code analysis rules
├── README.md              # Project documentation
└── CHANGELOG.md           # Version history
```

### Essential Project Configuration

```yaml
# pubspec.yaml
name: my_awesome_app
description: A comprehensive Flutter mobile application
version: 1.0.0+1
publish_to: 'none'

environment:
  sdk: ">=3.0.0 <4.0.0"
  flutter: ">=3.0.0"

dependencies:
  flutter:
    sdk: flutter

  # Essential dependencies
  cupertino_icons: ^1.0.6

  # State management (choose one)
  flutter_riverpod: ^2.4.9
  # flutter_bloc: ^8.1.3
  # provider: ^6.1.1

  # Navigation
  go_router: ^13.0.0

  # Network
  http: ^1.1.2
  dio: ^5.4.0

  # Local storage
  shared_preferences: ^2.2.2
  hive_flutter: ^1.1.0

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^3.0.1
  build_runner: ^2.4.7
  json_serializable: ^6.7.1

flutter:
  uses-material-design: true
  assets:
    - assets/images/
    - assets/icons/
  fonts:
    - family: CustomFont
      fonts:
        - asset: assets/fonts/CustomFont-Regular.ttf
        - asset: assets/fonts/CustomFont-Bold.ttf
          weight: 700
```

---

## Development Workflow Setup

### Hot Reload Configuration

```bash
# Enable hot reload (default)
flutter run

# Run with hot reload on file changes
flutter run --hot

# Run in debug mode with verbose logging
flutter run --debug --verbose

# Run on specific device
flutter run -d "iPhone 15 Pro"
```

### Build Configuration

```yaml
# analysis_options.yaml
include: package:flutter_lints/flutter.yaml

analyzer:
  exclude:
    - "**/*.g.dart"
    - "**/*.freezed.dart"
    - build/**

  language:
    strict-casts: true
    strict-inference: true
    strict-raw-types: true

linter:
  rules:
    - avoid_print
    - prefer_single_quotes
    - sort_constructors_first
    - sort_unnamed_constructors_first
    - always_use_package_imports
    - avoid_relative_lib_imports
    - prefer_relative_imports
```

### Scripts and Automation

```json
// package.json (for project scripts)
{
  "scripts": {
    "clean": "flutter clean && flutter pub get",
    "build:android": "flutter build apk --release",
    "build:ios": "flutter build ios --release",
    "build:web": "flutter build web --release",
    "test": "flutter test",
    "test:coverage": "flutter test --coverage",
    "analyze": "flutter analyze",
    "format": "dart format .",
    "gen": "flutter packages pub run build_runner build --delete-conflicting-outputs"
  }
}
```

### Makefile for Common Tasks

```makefile
# Makefile
.PHONY: clean build test analyze format

# Clean and get dependencies
clean:
	flutter clean
	flutter pub get

# Build for all platforms
build:
	flutter build apk --release
	flutter build ios --release
	flutter build web --release

# Run tests with coverage
test:
	flutter test --coverage
	lcov --remove coverage/lcov.info 'lib/*/*.g.dart' -o coverage/lcov.info

# Analyze code
analyze:
	flutter analyze

# Format code
format:
	dart format .

# Generate code
gen:
	flutter packages pub run build_runner build --delete-conflicting-outputs

# Run on Android emulator
android:
	flutter run -d android

# Run on iOS simulator
ios:
	flutter run -d ios

# Setup project
setup:
	flutter pub get
	cd ios && pod install
```

---

## Performance Optimization

### Development Performance

```bash
# Increase Dart VM memory
export DART_VM_OPTIONS="--old-gen-heap-size=4096"

# Enable Dart development compiler optimizations
flutter run --enable-software-rendering
flutter run --enable-impeller  # For new rendering engine
```

### Build Optimization

```yaml
# android/app/build.gradle
android {
    compileSdkVersion 34

    defaultConfig {
        minSdkVersion 21
        targetSdkVersion 34

        // Enable multidex for large apps
        multiDexEnabled true
    }

    buildTypes {
        release {
            // Enable obfuscation
            minifyEnabled true
            useProguard true
            proguardFiles getDefaultProguardFile('proguard-android.txt'), 'proguard-rules.pro'

            // Enable split APKs
            splits {
                abi {
                    enable true
                    reset()
                    include 'x86', 'x86_64', 'arm64-v8a', 'armeabi-v7a'
                    universalApk false
                }
            }
        }
    }
}
```

---

## Device Testing Setup

### Physical Device Configuration

```bash
# Android device setup
# 1. Enable Developer Options
# 2. Enable USB Debugging
# 3. Connect device via USB

# Check connected devices
flutter devices
adb devices

# iOS device setup (macOS only)
# 1. Connect device via USB
# 2. Trust computer
# 3. Device should appear in Xcode

# List connected iOS devices
xcrun xctrace list devices
```

### Network Testing

```bash
# Test over WiFi (Android)
adb tcpip 5555
adb connect <device_ip>:5555

# Run app on WiFi-connected device
flutter run -d <device_id>
```

---

## Troubleshooting Common Issues

### Flutter Doctor Issues

```bash
# Run Flutter doctor to check setup
flutter doctor

# Fix common issues
flutter doctor --android-licenses  # Accept Android licenses
flutter config --android-sdk /path/to/android/sdk
flutter config --android-studio-dir /path/to/android-studio
```

### Dependency Issues

```bash
# Clean and reinstall dependencies
flutter clean
flutter pub get

# Clear pub cache
flutter pub cache clean

# Repair broken packages
flutter pub deps
flutter pub upgrade
```

### Platform-Specific Issues

```bash
# Android build issues
cd android && ./gradlew clean
flutter clean && flutter pub get

# iOS build issues
cd ios
rm -rf Pods Podfile.lock
pod install --repo-update

# Clear derived data (iOS)
rm -rf ~/Library/Developer/Xcode/DerivedData
```

---

## Development Best Practices

### Environment Configuration

```bash
# Use version management
# Flutter Version Management (fvm)
dart pub global activate fvm

# Install specific Flutter version
fvm install 3.24.0
fvm use 3.24.0

# Use different versions per project
fvm use 3.24.0 --force
```

### Code Quality Tools

```yaml
# .github/workflows/quality.yml
name: Code Quality
on: [push, pull_request]

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: subosito/flutter-action@v2
        with:
          flutter-version: '3.24.0'

      - name: Install dependencies
        run: flutter pub get

      - name: Analyze code
        run: flutter analyze

      - name: Check formatting
        run: dart format --set-exit-if-changed .

      - name: Run tests
        run: flutter test --coverage
```

### Security Configuration

```bash
# Secure sensitive files
echo "*.keystore" >> .gitignore
echo "key.properties" >> .gitignore
echo "GoogleService-Info.plist" >> .gitignore
echo "google-services.json" >> .gitignore

# Use environment variables for secrets
export API_KEY="your-secret-key"
export DATABASE_URL="your-database-url"
```

---

## Verification and Testing

### Complete Setup Verification

```bash
# Verify Flutter installation
flutter doctor -v

# Create test project
flutter create test_app
cd test_app

# Run on available devices
flutter run

# Run tests
flutter test

# Build for release
flutter build apk --release
flutter build ios --release
flutter build web --release
```

### Performance Testing

```bash
# Profile app performance
flutter run --profile

# Analyze app size
flutter build apk --analyze-size

# Memory profiling
flutter run --enable-dart-profiling
```

This comprehensive local development setup provides a solid foundation for productive Flutter/Dart mobile development. Regular maintenance and updates of tools and SDKs ensure optimal performance and access to the latest features.

---

**Last Updated**: September 13, 2025
**Flutter Version**: 3.24.0+
**Dart Version**: 3.9.2+
**Platform Support**: Android, iOS, Web, Desktop