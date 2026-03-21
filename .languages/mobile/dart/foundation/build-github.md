# GitHub Actions CI/CD for Flutter/Dart

> Comprehensive guide to setting up continuous integration and continuous deployment pipelines for Flutter mobile applications using GitHub Actions.

## CI/CD Overview

GitHub Actions provides powerful automation for Flutter/Dart projects, enabling automatic testing, building, and deployment of mobile applications. This guide covers complete pipeline setup from basic CI to production deployment.

### Pipeline Benefits
- **Automated Testing**: Run tests on every commit and pull request
- **Multi-Platform Builds**: Build for Android, iOS, and web simultaneously
- **Quality Assurance**: Code analysis, formatting checks, and security scanning
- **Automated Deployment**: Deploy to app stores and distribution platforms
- **Performance Monitoring**: Track build times and app performance

---

## Basic CI Pipeline

### Starter Workflow

```yaml
# .github/workflows/ci.yml
name: CI Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    name: Test and Analyze
    runs-on: ubuntu-latest

    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: '3.24.0'
        channel: 'stable'
        cache: true

    - name: Get dependencies
      run: flutter pub get

    - name: Verify formatting
      run: dart format --set-exit-if-changed .

    - name: Analyze project source
      run: flutter analyze --fatal-infos --fatal-warnings

    - name: Run tests
      run: flutter test --coverage --test-randomize-ordering-seed random

    - name: Upload coverage reports
      uses: codecov/codecov-action@v3
      with:
        file: coverage/lcov.info
        name: codecov-umbrella
        fail_ci_if_error: true
```

---

## Advanced CI Pipeline

### Comprehensive Quality Checks

```yaml
# .github/workflows/quality.yml
name: Code Quality and Security

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  FLUTTER_VERSION: '3.24.0'

jobs:
  analyze:
    name: Static Analysis
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: ${{ env.FLUTTER_VERSION }}
        cache: true

    - name: Cache pub dependencies
      uses: actions/cache@v3
      with:
        path: ${{ env.PUB_CACHE }}
        key: ${{ runner.os }}-pub-${{ hashFiles('**/pubspec.lock') }}
        restore-keys: ${{ runner.os }}-pub-

    - name: Install dependencies
      run: flutter pub get

    - name: Generate code (if needed)
      run: |
        if grep -q build_runner pubspec.yaml; then
          flutter packages pub run build_runner build --delete-conflicting-outputs
        fi

    - name: Check formatting
      run: dart format --set-exit-if-changed .

    - name: Analyze code
      run: flutter analyze --fatal-infos --fatal-warnings

    - name: Check pub publish
      run: dart pub publish --dry-run

    - name: Run custom lints
      run: |
        flutter pub run custom_lint

  security:
    name: Security Scan
    runs-on: ubuntu-latest

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
      with:
        sarif_file: 'trivy-results.sarif'

  test:
    name: Unit and Widget Tests
    runs-on: ubuntu-latest

    strategy:
      matrix:
        flutter-version: ['3.22.0', '3.24.0']

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: ${{ matrix.flutter-version }}
        cache: true

    - name: Install dependencies
      run: flutter pub get

    - name: Run unit tests
      run: |
        flutter test \
          --coverage \
          --test-randomize-ordering-seed random \
          --reporter expanded

    - name: Generate coverage report
      run: |
        sudo apt-get update
        sudo apt-get install -y lcov
        genhtml coverage/lcov.info -o coverage/html

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: coverage/lcov.info
        fail_ci_if_error: true

    - name: Archive coverage results
      uses: actions/upload-artifact@v3
      with:
        name: coverage-report-${{ matrix.flutter-version }}
        path: coverage/html/
```

---

## Multi-Platform Build Pipeline

### Android and iOS Builds

```yaml
# .github/workflows/build.yml
name: Build Applications

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

env:
  FLUTTER_VERSION: '3.24.0'
  JAVA_VERSION: '17'

jobs:
  build-android:
    name: Build Android
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Java
      uses: actions/setup-java@v3
      with:
        distribution: 'adopt'
        java-version: ${{ env.JAVA_VERSION }}

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: ${{ env.FLUTTER_VERSION }}
        cache: true

    - name: Setup Android SDK
      uses: android-actions/setup-android@v3
      with:
        api-level: 34
        build-tools: 34.0.0

    - name: Cache Gradle dependencies
      uses: actions/cache@v3
      with:
        path: |
          ~/.gradle/caches
          ~/.gradle/wrapper
        key: gradle-${{ runner.os }}-${{ hashFiles('**/*.gradle*', '**/gradle-wrapper.properties') }}

    - name: Install dependencies
      run: flutter pub get

    - name: Create keystore
      run: |
        echo "${{ secrets.KEYSTORE_BASE64 }}" | base64 --decode > android/app/keystore.jks
        echo "storePassword=${{ secrets.KEYSTORE_PASSWORD }}" >> android/key.properties
        echo "keyPassword=${{ secrets.KEY_PASSWORD }}" >> android/key.properties
        echo "keyAlias=${{ secrets.KEY_ALIAS }}" >> android/key.properties
        echo "storeFile=keystore.jks" >> android/key.properties

    - name: Build APK
      run: flutter build apk --release --split-per-abi

    - name: Build App Bundle
      run: flutter build appbundle --release

    - name: Upload APK artifacts
      uses: actions/upload-artifact@v3
      with:
        name: android-apk
        path: build/app/outputs/flutter-apk/*.apk

    - name: Upload AAB artifact
      uses: actions/upload-artifact@v3
      with:
        name: android-aab
        path: build/app/outputs/bundle/release/*.aab

  build-ios:
    name: Build iOS
    runs-on: macos-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: ${{ env.FLUTTER_VERSION }}
        cache: true

    - name: Install dependencies
      run: flutter pub get

    - name: Setup iOS signing
      env:
        PROVISIONING_PROFILE_BASE64: ${{ secrets.PROVISIONING_PROFILE_BASE64 }}
        P12_BASE64: ${{ secrets.P12_BASE64 }}
        P12_PASSWORD: ${{ secrets.P12_PASSWORD }}
      run: |
        # Create temporary keychain
        security create-keychain -p "" build.keychain
        security default-keychain -s build.keychain
        security unlock-keychain -p "" build.keychain

        # Import certificate
        echo $P12_BASE64 | base64 --decode > certificate.p12
        security import certificate.p12 -k build.keychain -P $P12_PASSWORD -T /usr/bin/codesign
        security set-key-partition-list -S apple-tool:,apple: -s -k "" build.keychain

        # Install provisioning profile
        mkdir -p ~/Library/MobileDevice/Provisioning\ Profiles
        echo $PROVISIONING_PROFILE_BASE64 | base64 --decode > ~/Library/MobileDevice/Provisioning\ Profiles/profile.mobileprovision

    - name: Install CocoaPods
      run: |
        cd ios
        pod install

    - name: Build iOS
      run: |
        flutter build ios --release --no-codesign

    - name: Build IPA
      run: |
        cd ios
        xcodebuild -workspace Runner.xcworkspace \
                   -scheme Runner \
                   -configuration Release \
                   -destination generic/platform=iOS \
                   -archivePath Runner.xcarchive \
                   archive

        xcodebuild -exportArchive \
                   -archivePath Runner.xcarchive \
                   -exportPath . \
                   -exportOptionsPlist ExportOptions.plist

    - name: Upload IPA artifact
      uses: actions/upload-artifact@v3
      with:
        name: ios-ipa
        path: ios/*.ipa

  build-web:
    name: Build Web
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: ${{ env.FLUTTER_VERSION }}
        cache: true

    - name: Enable Flutter web
      run: flutter config --enable-web

    - name: Install dependencies
      run: flutter pub get

    - name: Build web
      run: flutter build web --release --web-renderer canvaskit

    - name: Upload web artifacts
      uses: actions/upload-artifact@v3
      with:
        name: web-build
        path: build/web/
```

---

## Integration Testing Pipeline

### E2E Testing with Real Devices

```yaml
# .github/workflows/integration-tests.yml
name: Integration Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  integration-test-android:
    name: Android Integration Tests
    runs-on: macos-latest

    strategy:
      matrix:
        api-level: [28, 30, 34]

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: '3.24.0'
        cache: true

    - name: Install dependencies
      run: flutter pub get

    - name: Setup Android emulator
      uses: reactivecircus/android-emulator-runner@v2
      with:
        api-level: ${{ matrix.api-level }}
        arch: x86_64
        profile: Nexus 6
        script: |
          flutter drive \
            --driver=test_driver/integration_test.dart \
            --target=integration_test/app_test.dart \
            --device-id=emulator-5554

  integration-test-ios:
    name: iOS Integration Tests
    runs-on: macos-latest

    strategy:
      matrix:
        device: ['iPhone 14', 'iPhone 15 Pro']

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: '3.24.0'
        cache: true

    - name: Install dependencies
      run: flutter pub get

    - name: Start iOS Simulator
      run: |
        xcrun simctl create test-simulator com.apple.CoreSimulator.SimDeviceType.iPhone-15-Pro com.apple.CoreSimulator.SimRuntime.iOS-17-0
        xcrun simctl boot test-simulator

    - name: Run integration tests
      run: |
        flutter drive \
          --driver=test_driver/integration_test.dart \
          --target=integration_test/app_test.dart \
          --device-id=test-simulator

    - name: Upload test results
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: integration-test-results-${{ matrix.device }}
        path: test_driver/screenshots/
```

---

## Deployment Pipeline

### Automated App Store Deployment

```yaml
# .github/workflows/deploy.yml
name: Deploy to App Stores

on:
  push:
    tags: [ 'v*' ]

env:
  FLUTTER_VERSION: '3.24.0'

jobs:
  deploy-android:
    name: Deploy to Google Play
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: ${{ env.FLUTTER_VERSION }}

    - name: Install dependencies
      run: flutter pub get

    - name: Setup signing
      run: |
        echo "${{ secrets.KEYSTORE_BASE64 }}" | base64 --decode > android/app/keystore.jks
        echo "storePassword=${{ secrets.KEYSTORE_PASSWORD }}" >> android/key.properties
        echo "keyPassword=${{ secrets.KEY_PASSWORD }}" >> android/key.properties
        echo "keyAlias=${{ secrets.KEY_ALIAS }}" >> android/key.properties
        echo "storeFile=keystore.jks" >> android/key.properties

    - name: Build App Bundle
      run: flutter build appbundle --release

    - name: Upload to Google Play
      uses: r0adkll/upload-google-play@v1
      with:
        serviceAccountJsonPlainText: ${{ secrets.GOOGLE_PLAY_SERVICE_ACCOUNT_JSON }}
        packageName: com.example.myapp
        releaseFiles: build/app/outputs/bundle/release/*.aab
        track: production
        status: completed
        whatsNewDirectory: android/release-notes/

  deploy-ios:
    name: Deploy to App Store
    runs-on: macos-latest
    if: startsWith(github.ref, 'refs/tags/v')

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: ${{ env.FLUTTER_VERSION }}

    - name: Install dependencies
      run: flutter pub get

    - name: Setup iOS signing
      env:
        PROVISIONING_PROFILE_BASE64: ${{ secrets.PROVISIONING_PROFILE_BASE64 }}
        P12_BASE64: ${{ secrets.P12_BASE64 }}
        P12_PASSWORD: ${{ secrets.P12_PASSWORD }}
      run: |
        # Setup signing as in build pipeline
        security create-keychain -p "" build.keychain
        security default-keychain -s build.keychain
        security unlock-keychain -p "" build.keychain

        echo $P12_BASE64 | base64 --decode > certificate.p12
        security import certificate.p12 -k build.keychain -P $P12_PASSWORD -T /usr/bin/codesign
        security set-key-partition-list -S apple-tool:,apple: -s -k "" build.keychain

        mkdir -p ~/Library/MobileDevice/Provisioning\ Profiles
        echo $PROVISIONING_PROFILE_BASE64 | base64 --decode > ~/Library/MobileDevice/Provisioning\ Profiles/profile.mobileprovision

    - name: Build and archive
      run: |
        cd ios
        pod install
        flutter build ios --release --no-codesign

        xcodebuild -workspace Runner.xcworkspace \
                   -scheme Runner \
                   -configuration Release \
                   -destination generic/platform=iOS \
                   -archivePath Runner.xcarchive \
                   archive

    - name: Upload to App Store
      env:
        APP_STORE_CONNECT_API_KEY_ID: ${{ secrets.APP_STORE_CONNECT_API_KEY_ID }}
        APP_STORE_CONNECT_API_ISSUER_ID: ${{ secrets.APP_STORE_CONNECT_API_ISSUER_ID }}
        APP_STORE_CONNECT_API_KEY: ${{ secrets.APP_STORE_CONNECT_API_KEY }}
      run: |
        cd ios
        xcodebuild -exportArchive \
                   -archivePath Runner.xcarchive \
                   -exportPath . \
                   -exportOptionsPlist ExportOptions.plist

        xcrun altool --upload-app \
                     --type ios \
                     --file *.ipa \
                     --apiKey $APP_STORE_CONNECT_API_KEY_ID \
                     --apiIssuer $APP_STORE_CONNECT_API_ISSUER_ID

  deploy-web:
    name: Deploy to GitHub Pages
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: ${{ env.FLUTTER_VERSION }}

    - name: Enable Flutter web
      run: flutter config --enable-web

    - name: Install dependencies
      run: flutter pub get

    - name: Build web
      run: flutter build web --release --base-href "/my-flutter-app/"

    - name: Deploy to GitHub Pages
      uses: peaceiris/actions-gh-pages@v3
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
        publish_dir: build/web
```

---

## Performance Monitoring

### Build Performance and Size Analysis

```yaml
# .github/workflows/performance.yml
name: Performance Analysis

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  analyze-size:
    name: App Size Analysis
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: '3.24.0'

    - name: Install dependencies
      run: flutter pub get

    - name: Build and analyze APK size
      run: |
        flutter build apk --release --analyze-size --tree-shake-icons

    - name: Extract size information
      run: |
        flutter build apk --analyze-size > size-analysis.txt
        cat size-analysis.txt

    - name: Upload size analysis
      uses: actions/upload-artifact@v3
      with:
        name: size-analysis
        path: size-analysis.txt

    - name: Comment PR with size analysis
      if: github.event_name == 'pull_request'
      uses: actions/github-script@v6
      with:
        script: |
          const fs = require('fs');
          const analysis = fs.readFileSync('size-analysis.txt', 'utf8');
          github.rest.issues.createComment({
            issue_number: context.issue.number,
            owner: context.repo.owner,
            repo: context.repo.repo,
            body: `## 📊 App Size Analysis\n\n\`\`\`\n${analysis}\n\`\`\``
          });

  benchmark:
    name: Performance Benchmarks
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: '3.24.0'

    - name: Install dependencies
      run: flutter pub get

    - name: Run benchmarks
      run: |
        cd test/benchmarks
        flutter test --reporter json > benchmark-results.json

    - name: Process benchmark results
      run: |
        python scripts/process_benchmarks.py benchmark-results.json

    - name: Upload benchmark results
      uses: actions/upload-artifact@v3
      with:
        name: benchmark-results
        path: |
          benchmark-results.json
          benchmark-summary.html
```

---

## Workflow Optimization

### Caching and Performance

```yaml
# .github/workflows/optimized-ci.yml
name: Optimized CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  FLUTTER_VERSION: '3.24.0'

jobs:
  test:
    name: Optimized Testing
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Cache Flutter SDK
      id: flutter-cache
      uses: actions/cache@v3
      with:
        path: |
          /opt/hostedtoolcache/flutter
          ~/.pub-cache
        key: flutter-${{ env.FLUTTER_VERSION }}-${{ runner.os }}

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: ${{ env.FLUTTER_VERSION }}
        cache: true

    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: |
          ${{ env.PUB_CACHE }}
          ~/.pub-cache
        key: pub-cache-${{ runner.os }}-${{ hashFiles('**/pubspec.lock') }}
        restore-keys: |
          pub-cache-${{ runner.os }}-

    - name: Install dependencies
      run: flutter pub get

    - name: Cache generated files
      uses: actions/cache@v3
      with:
        path: |
          lib/**/*.g.dart
          lib/**/*.freezed.dart
        key: generated-${{ runner.os }}-${{ hashFiles('lib/**/*.dart', 'pubspec.lock') }}

    - name: Generate code if needed
      run: |
        if [ ! -f "lib/generated_flag" ]; then
          flutter packages pub run build_runner build --delete-conflicting-outputs
          touch lib/generated_flag
        fi

    - name: Run tests in parallel
      run: |
        flutter test --concurrency=$(nproc) --coverage

  build-matrix:
    name: Build Matrix
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
        flutter-version: ['3.22.0', '3.24.0']
      fail-fast: false

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: ${{ matrix.flutter-version }}
        cache: true

    - name: Build for platform
      run: |
        if [ "$RUNNER_OS" == "Linux" ]; then
          flutter build apk --release
          flutter build web --release
        elif [ "$RUNNER_OS" == "macOS" ]; then
          flutter build ios --release --no-codesign
          flutter build macos --release
        fi
```

---

## Secret Management

### Required Secrets Configuration

```yaml
# Repository secrets needed for CI/CD
secrets:
  # Android signing
  KEYSTORE_BASE64: # Base64 encoded keystore file
  KEYSTORE_PASSWORD: # Keystore password
  KEY_ALIAS: # Key alias
  KEY_PASSWORD: # Key password

  # iOS signing
  PROVISIONING_PROFILE_BASE64: # Base64 encoded provisioning profile
  P12_BASE64: # Base64 encoded certificate
  P12_PASSWORD: # Certificate password

  # App Store deployment
  APP_STORE_CONNECT_API_KEY_ID: # App Store Connect API key ID
  APP_STORE_CONNECT_API_ISSUER_ID: # Issuer ID
  APP_STORE_CONNECT_API_KEY: # Private key content

  # Google Play deployment
  GOOGLE_PLAY_SERVICE_ACCOUNT_JSON: # Service account JSON

  # Other services
  CODECOV_TOKEN: # Codecov upload token
  SLACK_WEBHOOK_URL: # Slack notifications
```

### Environment-Specific Configuration

```yaml
# .github/workflows/deploy-staging.yml
name: Deploy Staging

on:
  push:
    branches: [ develop ]

jobs:
  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    environment: staging

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: '3.24.0'

    - name: Configure for staging
      run: |
        echo "API_URL=${{ vars.STAGING_API_URL }}" > .env
        echo "APP_NAME=${{ vars.STAGING_APP_NAME }}" >> .env

    - name: Build and deploy
      run: |
        flutter build apk --release --flavor staging
        # Deploy to staging environment
```

---

## Monitoring and Notifications

### Slack Integration

```yaml
# Add to any workflow for notifications
- name: Notify Slack on failure
  if: failure()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    channel: '#dev-notifications'
    webhook_url: ${{ secrets.SLACK_WEBHOOK_URL }}
    message: |
      Build failed for ${{ github.repository }}
      Branch: ${{ github.ref }}
      Commit: ${{ github.sha }}
```

### Email Notifications

```yaml
- name: Send email notification
  if: failure()
  uses: dawidd6/action-send-mail@v3
  with:
    server_address: smtp.gmail.com
    server_port: 465
    username: ${{ secrets.EMAIL_USERNAME }}
    password: ${{ secrets.EMAIL_PASSWORD }}
    subject: Build Failed - ${{ github.repository }}
    body: |
      Build failed for repository: ${{ github.repository }}
      Branch: ${{ github.ref }}
      Commit: ${{ github.sha }}
    to: dev-team@company.com
    from: ci-cd@company.com
```

This comprehensive GitHub Actions setup provides robust CI/CD capabilities for Flutter mobile development, ensuring code quality, automated testing, and reliable deployment processes.

---

**Last Updated**: September 13, 2025
**GitHub Actions Version**: Latest
**Flutter Version**: 3.24.0+
**Platform Support**: Android, iOS, Web, Desktop