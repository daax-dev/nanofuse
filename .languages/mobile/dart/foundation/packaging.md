# Dart Package Management and pub.dev

> Comprehensive guide to Dart package management, pub.dev ecosystem, and dependency management for mobile application development.

## Package Management Overview

Dart uses **pub** as its package manager, with **pub.dev** serving as the central package repository. The package system enables code reuse, dependency management, and ecosystem collaboration.

### Key Concepts
- **Package**: A directory containing code, resources, and metadata
- **pub.dev**: Official Dart package repository
- **pubspec.yaml**: Package configuration and dependency declaration
- **pub get**: Downloads and installs dependencies
- **pub upgrade**: Updates dependencies to latest compatible versions

---

## Package Structure

### Standard Package Layout

```
my_flutter_app/
├── lib/                    # Dart code (public API)
│   ├── main.dart          # App entry point
│   ├── src/               # Private implementation
│   │   ├── models/
│   │   ├── services/
│   │   └── widgets/
│   └── my_flutter_app.dart # Public library export
├── test/                   # Unit and widget tests
│   ├── unit/
│   ├── widget/
│   └── integration/
├── example/                # Example usage (for packages)
├── assets/                 # Images, fonts, data files
│   ├── images/
│   ├── fonts/
│   └── data/
├── android/                # Android-specific files
├── ios/                    # iOS-specific files
├── web/                    # Web-specific files (if applicable)
├── pubspec.yaml           # Package configuration
├── pubspec.lock           # Locked dependency versions
├── README.md              # Package documentation
├── CHANGELOG.md           # Version history
└── LICENSE                # License information
```

### pubspec.yaml Configuration

```yaml
name: my_flutter_app
description: A comprehensive Flutter mobile application
version: 1.2.3+4  # version+buildNumber
publish_to: none  # Prevent accidental publishing

environment:
  sdk: ">=3.0.0 <4.0.0"
  flutter: ">=3.0.0"

dependencies:
  flutter:
    sdk: flutter

  # State Management
  riverpod: ^2.4.9
  flutter_riverpod: ^2.4.9

  # Network & HTTP
  http: ^1.1.2
  dio: ^5.4.0

  # JSON & Serialization
  json_annotation: ^4.8.1
  freezed_annotation: ^2.4.1

  # Storage & Database
  shared_preferences: ^2.2.2
  sqflite: ^2.3.0
  hive: ^2.2.3
  hive_flutter: ^1.1.0

  # UI & Navigation
  go_router: ^13.0.0
  flutter_svg: ^2.0.9
  cached_network_image: ^3.3.0

  # Platform Integration
  package_info_plus: ^5.0.1
  device_info_plus: ^10.0.1
  permission_handler: ^11.3.0

  # Firebase (optional)
  firebase_core: ^2.24.2
  firebase_auth: ^4.16.0
  firebase_firestore: ^4.13.6

  # Development utilities
  logger: ^2.0.2+1
  uuid: ^4.3.3

dev_dependencies:
  flutter_test:
    sdk: flutter

  # Code Generation
  build_runner: ^2.4.7
  json_serializable: ^6.7.1
  freezed: ^2.4.7

  # Linting & Analysis
  flutter_lints: ^3.0.1
  very_good_analysis: ^5.1.0

  # Testing
  mocktail: ^1.0.2
  bloc_test: ^9.1.5
  integration_test:
    sdk: flutter

# Asset configuration
flutter:
  uses-material-design: true

  assets:
    - assets/images/
    - assets/icons/
    - assets/data/config.json

  fonts:
    - family: CustomFont
      fonts:
        - asset: assets/fonts/CustomFont-Regular.ttf
        - asset: assets/fonts/CustomFont-Bold.ttf
          weight: 700

# Platform-specific configuration
flutter:
  assets:
    - assets/images/

  # Generate platform-specific icons
  flutter_icons:
    android: true
    ios: true
    image_path: "assets/icon/app_icon.png"
    min_sdk_android: 21
```

---

## Dependency Management

### Dependency Types

```yaml
dependencies:
  # Regular dependencies (included in final app)
  http: ^1.1.2              # Network requests
  provider: ^6.1.1          # State management

  # SDK dependencies
  flutter:
    sdk: flutter

dev_dependencies:
  # Development-only dependencies
  flutter_test:
    sdk: flutter
  build_runner: ^2.4.7      # Code generation
  flutter_lints: ^3.0.1     # Linting rules

dependency_overrides:
  # Force specific versions (use sparingly)
  meta: 1.9.1

# Git dependencies
dependencies:
  my_custom_package:
    git:
      url: https://github.com/myuser/my_custom_package.git
      ref: main  # or specific commit/tag

  # Git with path
  another_package:
    git:
      url: https://github.com/myuser/monorepo.git
      path: packages/another_package
      ref: v2.0.0

# Local path dependencies (for development)
dependencies:
  local_package:
    path: ../local_package/

# Hosted dependencies (non-pub.dev)
dependencies:
  private_package:
    hosted:
      name: private_package
      url: https://my-private-pub-server.com
    version: ^1.0.0
```

### Version Constraints

```yaml
dependencies:
  # Exact version
  package_name: 1.2.3

  # Caret syntax (compatible versions)
  http: ^1.1.2              # >=1.1.2 <2.0.0

  # Range constraints
  json_annotation: ">=4.8.0 <5.0.0"

  # Any version (not recommended)
  some_package: any

  # Pre-release versions
  experimental_package: 1.0.0-beta.1
```

### Managing Dependencies

```bash
# Install dependencies
flutter pub get

# Upgrade to latest compatible versions
flutter pub upgrade

# Upgrade to specific version
flutter pub upgrade http

# Add new dependency
flutter pub add http
flutter pub add --dev flutter_lints

# Remove dependency
flutter pub remove unused_package

# Show dependency tree
flutter pub deps

# Analyze dependencies for issues
flutter pub analyze

# Check for outdated packages
flutter pub outdated

# Clean package cache
flutter pub cache clean
```

---

## Popular Mobile Development Packages

### State Management

```yaml
# BLoC Pattern
dependencies:
  bloc: ^8.1.2
  flutter_bloc: ^8.1.3
  bloc_test: ^9.1.5  # dev_dependency

# Riverpod (Recommended)
dependencies:
  riverpod: ^2.4.9
  flutter_riverpod: ^2.4.9
  riverpod_annotation: ^2.3.3

# Provider
dependencies:
  provider: ^6.1.1

# GetX
dependencies:
  get: ^4.6.6

# MobX
dependencies:
  mobx: ^2.2.0
  flutter_mobx: ^2.1.1
```

### Network & HTTP

```yaml
dependencies:
  # Basic HTTP client
  http: ^1.1.2

  # Advanced HTTP client
  dio: ^5.4.0
  pretty_dio_logger: ^1.3.1  # Logging interceptor

  # GraphQL
  graphql_flutter: ^5.1.2

  # WebSockets
  web_socket_channel: ^2.4.0
```

### Data Persistence

```yaml
dependencies:
  # Key-value storage
  shared_preferences: ^2.2.2
  flutter_secure_storage: ^9.0.0

  # NoSQL database
  hive: ^2.2.3
  hive_flutter: ^1.1.0

  # SQL database
  sqflite: ^2.3.0
  sqlite3_flutter_libs: ^0.5.0

  # Object-document mapping
  objectbox: ^2.3.1
  objectbox_flutter_libs: ^2.3.1
```

### UI Components

```yaml
dependencies:
  # Images
  cached_network_image: ^3.3.0
  flutter_svg: ^2.0.9
  image: ^4.1.3

  # Icons
  cupertino_icons: ^1.0.6
  font_awesome_flutter: ^10.6.0

  # Animations
  lottie: ^2.7.0
  rive: ^0.12.4

  # Charts
  fl_chart: ^0.66.0
  charts_flutter: ^0.12.0
```

### Navigation & Routing

```yaml
dependencies:
  # Modern routing (recommended)
  go_router: ^13.0.0

  # Auto route generation
  auto_route: ^7.8.4

dev_dependencies:
  auto_route_generator: ^7.3.2
```

### Platform Integration

```yaml
dependencies:
  # Device information
  device_info_plus: ^10.0.1
  package_info_plus: ^5.0.1

  # Permissions
  permission_handler: ^11.3.0

  # Platform channels
  flutter/services  # Built-in

  # File system
  path_provider: ^2.1.2
  file_picker: ^6.1.1

  # Camera & Media
  camera: ^0.10.5+5
  image_picker: ^1.0.5
  video_player: ^2.8.1
```

### Firebase Integration

```yaml
dependencies:
  # Core Firebase
  firebase_core: ^2.24.2

  # Authentication
  firebase_auth: ^4.16.0

  # Database
  cloud_firestore: ^4.13.6
  firebase_database: ^10.4.0

  # Storage
  firebase_storage: ^11.6.0

  # Cloud Functions
  cloud_functions: ^4.6.0

  # Messaging
  firebase_messaging: ^14.7.10

  # Analytics
  firebase_analytics: ^10.8.0

  # Crashlytics
  firebase_crashlytics: ^3.4.9
```

---

## Package Development

### Creating a Package

```bash
# Create Dart package
dart create -t package my_dart_package

# Create Flutter package
flutter create --template=package my_flutter_package

# Create Flutter plugin (with platform code)
flutter create --template=plugin my_flutter_plugin

# Create federated plugin
flutter create --template=plugin --platforms=android,ios my_federated_plugin
```

### Package Structure for Publishing

```yaml
# pubspec.yaml for a publishable package
name: my_awesome_package
description: A comprehensive utility package for Flutter development
version: 1.0.0

# Homepage and repository
homepage: https://github.com/myusername/my_awesome_package
repository: https://github.com/myusername/my_awesome_package
issue_tracker: https://github.com/myusername/my_awesome_package/issues
documentation: https://my_awesome_package.dev

# Publishing configuration
publish_to: https://pub.dev  # or 'none' for private packages

# Supported platforms
platforms:
  android:
  ios:
  linux:
  macos:
  web:
  windows:

# Topics for discoverability
topics:
  - flutter
  - mobile
  - utility
  - state-management

environment:
  sdk: ">=3.0.0 <4.0.0"
  flutter: ">=3.0.0"

dependencies:
  flutter:
    sdk: flutter
  meta: ^1.10.0

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^3.0.1
```

### Package Library Structure

```dart
// lib/my_awesome_package.dart (main library file)
library my_awesome_package;

export 'src/core/awesome_manager.dart';
export 'src/widgets/awesome_widget.dart';
export 'src/models/awesome_model.dart';
export 'src/utils/awesome_utils.dart';

// Do not export src/ files directly
// export 'src/internal/private_implementation.dart'; // ❌

// lib/src/core/awesome_manager.dart
class AwesomeManager {
  static const String version = '1.0.0';

  /// Initialize the awesome manager with configuration
  static void initialize(AwesomeConfig config) {
    _config = config;
    _initialized = true;
  }

  // Implementation details...
}

// lib/src/widgets/awesome_widget.dart
import 'package:flutter/material.dart';
import '../core/awesome_manager.dart';

class AwesomeWidget extends StatelessWidget {
  const AwesomeWidget({
    super.key,
    required this.title,
    this.onTap,
  });

  final String title;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Text(title),
        ),
      ),
    );
  }
}
```

### Package Documentation

```dart
// Well-documented public API
/// A comprehensive utility package for Flutter development.
///
/// This package provides utilities for common Flutter development tasks
/// including state management, network operations, and UI components.
///
/// ## Getting Started
///
/// Initialize the package in your app's main function:
///
/// ```dart
/// void main() {
///   AwesomeManager.initialize(AwesomeConfig(
///     apiKey: 'your-api-key',
///     debugMode: kDebugMode,
///   ));
///   runApp(MyApp());
/// }
/// ```
library my_awesome_package;

/// Configuration class for AwesomeManager
///
/// Example:
/// ```dart
/// final config = AwesomeConfig(
///   apiKey: 'abc123',
///   debugMode: true,
/// );
/// ```
class AwesomeConfig {
  /// Creates an AwesomeConfig with the given parameters.
  const AwesomeConfig({
    required this.apiKey,
    this.debugMode = false,
    this.timeout = const Duration(seconds: 30),
  });

  /// The API key for authentication
  final String apiKey;

  /// Whether debug mode is enabled
  final bool debugMode;

  /// Network request timeout duration
  final Duration timeout;
}
```

---

## Publishing Packages

### Pre-publish Checklist

```bash
# 1. Analyze package for issues
dart analyze

# 2. Run tests
flutter test

# 3. Check package structure and metadata
dart pub publish --dry-run

# 4. Format code consistently
dart format .

# 5. Generate documentation
dart doc

# 6. Verify example works
cd example && flutter run

# 7. Check pub.dev scoring criteria
# - Follows Dart conventions
# - Has documentation
# - Supports multiple platforms
# - Has example
# - Passes static analysis
```

### Publishing Process

```bash
# Publish to pub.dev (requires verification)
dart pub publish

# Publish with confirmation
dart pub publish --force

# Publish specific version
dart pub publish --tag=v1.0.0
```

### Package Maintenance

```yaml
# Semantic versioning
version: 1.2.3
# MAJOR.MINOR.PATCH
# MAJOR: Breaking changes
# MINOR: New features (backward compatible)
# PATCH: Bug fixes (backward compatible)

# Pre-release versions
version: 1.2.3-alpha.1
version: 1.2.3-beta.2
version: 1.2.3-rc.1
```

---

## Private Package Management

### Creating Private Repositories

```yaml
# Using Git dependencies for private packages
dependencies:
  private_package:
    git:
      url: https://github.com/company/private_package.git
      ref: main

# Using SSH for private repositories
dependencies:
  ssh_private_package:
    git:
      url: git@github.com:company/ssh_private_package.git
      ref: v1.0.0
```

### Local Development Workflow

```yaml
# pubspec.yaml during development
dependencies:
  my_package:
    path: ../my_package/  # Local development

# Switch to published version for release
dependencies:
  my_package: ^1.0.0
```

### Private Pub Server

```yaml
# Using custom pub server
dependencies:
  private_package:
    hosted:
      name: private_package
      url: https://pub.company.com
    version: ^1.0.0

# Environment configuration
dependency_overrides:
  private_package:
    hosted:
      name: private_package
      url: $PUB_SERVER_URL
    version: ^1.0.0
```

---

## Package Security

### Security Best Practices

```yaml
# 1. Audit dependencies regularly
dev_dependencies:
  dependency_validator: ^3.2.2

# 2. Pin critical dependencies
dependencies:
  security_critical_package: 1.2.3  # Exact version

# 3. Use dependency overrides carefully
dependency_overrides:
  vulnerable_package: 1.2.4  # Only when necessary

# 4. Avoid packages with security issues
# Check pub.dev security advisories
```

### Vulnerability Scanning

```bash
# Check for known vulnerabilities
dart pub audit

# Update vulnerable packages
dart pub upgrade

# Get security reports
dart pub report
```

---

## Build Configuration

### Asset Management

```yaml
flutter:
  assets:
    # Include all files in directory
    - assets/images/

    # Include specific files
    - assets/data/config.json
    - assets/data/sample.xml

    # Include files from packages
    - packages/my_package/assets/icons/

  fonts:
    - family: Roboto
      fonts:
        - asset: assets/fonts/Roboto-Regular.ttf
        - asset: assets/fonts/Roboto-Bold.ttf
          weight: 700
        - asset: assets/fonts/Roboto-Italic.ttf
          style: italic
```

### Build Flavors

```yaml
# Android build.gradle configuration
android {
    flavorDimensions "default"
    productFlavors {
        dev {
            dimension "default"
            applicationIdSuffix ".dev"
            versionNameSuffix "-dev"
        }
        staging {
            dimension "default"
            applicationIdSuffix ".staging"
            versionNameSuffix "-staging"
        }
        production {
            dimension "default"
        }
    }
}
```

---

## Package Testing

### Testing Package Dependencies

```dart
// test/package_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:my_awesome_package/my_awesome_package.dart';

void main() {
  group('Package Tests', () {
    test('should initialize correctly', () {
      const config = AwesomeConfig(apiKey: 'test-key');
      AwesomeManager.initialize(config);

      expect(AwesomeManager.isInitialized, isTrue);
    });

    test('should handle invalid configuration', () {
      expect(
        () => AwesomeConfig(apiKey: ''),
        throwsA(isA<ArgumentError>()),
      );
    });
  });
}

// Integration testing with dependencies
import 'package:integration_test/integration_test.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('package integration test', (tester) async {
    // Test package functionality in real app context
    await tester.pumpWidget(MyApp());
    await tester.pumpAndSettle();

    // Verify package widgets render correctly
    expect(find.byType(AwesomeWidget), findsOneWidget);
  });
}
```

### Mock Dependencies

```dart
// Using mocktail for testing
import 'package:mocktail/mocktail.dart';

class MockHttpClient extends Mock implements http.Client {}

void main() {
  late MockHttpClient mockClient;

  setUp(() {
    mockClient = MockHttpClient();
  });

  test('should handle network request', () async {
    // Arrange
    when(() => mockClient.get(any())).thenAnswer(
      (_) async => http.Response('{"success": true}', 200),
    );

    // Act
    final service = ApiService(client: mockClient);
    final result = await service.fetchData();

    // Assert
    expect(result.success, isTrue);
    verify(() => mockClient.get(any())).called(1);
  });
}
```

---

## Best Practices

### Dependency Management Guidelines

1. **Version Constraints**
```yaml
# ✅ Good: Use caret constraints
dependencies:
  http: ^1.1.2

# ❌ Avoid: Overly restrictive
dependencies:
  http: 1.1.2

# ❌ Avoid: Too permissive
dependencies:
  http: any
```

2. **Regular Updates**
```bash
# Update packages monthly
flutter pub upgrade

# Check for major version updates
flutter pub outdated

# Review changelogs before updating
```

3. **Package Selection Criteria**
   - **Popularity**: High download count and likes
   - **Maintenance**: Recent updates and active issues
   - **Quality**: Good documentation and examples
   - **Compatibility**: Supports target platforms
   - **License**: Compatible with your project

4. **Minimize Dependencies**
```yaml
# ✅ Good: Essential dependencies only
dependencies:
  flutter:
    sdk: flutter
  http: ^1.1.2
  provider: ^6.1.1

# ❌ Avoid: Unnecessary dependencies
dependencies:
  unused_package: ^1.0.0
  duplicate_functionality: ^2.0.0
```

---

**Last Updated**: September 13, 2025
**Pub Version**: Latest pub client
**pub.dev**: Current ecosystem and guidelines
**Mobile Focus**: Optimized for Flutter mobile development workflows