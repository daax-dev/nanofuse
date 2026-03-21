# Awesome Kotlin for Mobile Development

## Overview
This curated collection of Kotlin resources focuses specifically on mobile development, enhancing and expanding upon the excellent [awesome-kotlin](https://github.com/KotlinBy/awesome-kotlin) repository with mobile-first insights, production-ready patterns, and real-world implementation guidance.

## Table of Contents
- [Core Libraries and Frameworks](#core-libraries-and-frameworks)
- [Architecture and Design Patterns](#architecture-and-design-patterns)
- [UI Development](#ui-development)
- [Networking and API Integration](#networking-and-api-integration)
- [Data Persistence](#data-persistence)
- [Reactive Programming](#reactive-programming)
- [Testing](#testing)
- [Performance and Optimization](#performance-and-optimization)
- [Kotlin Multiplatform Mobile](#kotlin-multiplatform-mobile)
- [Developer Tools and IDE](#developer-tools-and-ide)
- [Learning Resources](#learning-resources)
- [Production Examples](#production-examples)

## Core Libraries and Frameworks

### Essential Android Kotlin Libraries
- **[Retrofit](https://square.github.io/retrofit/)** - Type-safe HTTP client for Android and Kotlin
  - *Production Usage*: Used by 90%+ of Android apps for network requests
  - *Mobile Benefits*: Automatic JSON parsing, coroutines support, built-in error handling
  - *Best Practices*: Use with OkHttp for caching and interceptors
  ```kotlin
  interface GitHubService {
      @GET("users/{user}/repos")
      suspend fun listRepos(@Path("user") user: String): List<Repository>
  }
  ```

- **[OkHttp](https://square.github.io/okhttp/)** - Efficient HTTP client with connection pooling
  - *Production Usage*: Foundation for most Android networking libraries
  - *Mobile Benefits*: Connection reuse, automatic GZIP, HTTP/2 support
  - *Performance*: Reduces battery usage through connection pooling

- **[Hilt](https://dagger.dev/hilt/)** - Dependency injection for Android
  - *Production Usage*: Google's recommended DI solution for Android
  - *Mobile Benefits*: Lifecycle-aware injection, simplified setup
  - *Architecture*: Enables clean, testable code architecture

### Jetpack Compose Libraries
- **[Accompanist](https://github.com/google/accompanist)** - Jetpack Compose utilities
  - *Components*: Permissions, Pager, System UI Controller, Navigation Material
  - *Mobile Focus*: Solves common mobile UI challenges in Compose
  - *Stability*: Battle-tested in production applications

- **[Compose Destinations](https://github.com/raamcosta/compose-destinations)** - Type-safe navigation for Compose
  - *Code Generation*: Compile-time safe navigation
  - *Developer Experience*: Reduces navigation boilerplate by 90%
  - *Integration*: Works seamlessly with ViewModels and Hilt

## Architecture and Design Patterns

### Architecture Frameworks
- **[Circuit](https://github.com/slackhq/circuit)** - Compose-driven architecture framework
  - *Production*: Powers Slack's mobile applications
  - *Benefits*: Unidirectional data flow, type-safe navigation, testable UI
  - *Pattern*: Presenter + UI separation with event-driven communication

- **[MVI-Kotlin](https://github.com/arkivanov/MVIKotlin)** - Model-View-Intent framework
  - *Mobile Focus*: Predictable state management for complex mobile UIs
  - *Benefits*: Time-travel debugging, reactive state management
  - *Multiplatform*: Works across Android, iOS, and desktop

### State Management
- **[Molecule](https://github.com/cashapp/molecule)** - Build reactive state with Compose
  - *Innovation*: Use Compose compiler for state management
  - *Performance*: Leverages Compose's diffing for efficient updates
  - *Pattern*: Functional reactive programming with familiar Compose syntax

### Architecture Components Extensions
- **[Store](https://github.com/MobileNativeFoundation/Store)** - Async data loading and caching
  - *Mobile Optimized*: Handles network state, offline scenarios
  - *Performance*: Intelligent caching with memory/disk strategies
  - *Integration*: Works with Room, Retrofit, and any data source

## UI Development

### Jetpack Compose Ecosystem
- **[Compose Material3](https://developer.android.com/jetpack/compose/designsystems/material3)** - Material Design 3 components
  - *Modern Design*: Latest Material Design system implementation
  - *Theming*: Dynamic color, custom themes, dark mode support
  - *Components*: 40+ production-ready UI components

- **[Lottie Compose](https://github.com/airbnb/lottie-android)** - After Effects animations
  - *Rich Animations*: Complex animations from After Effects
  - *Performance*: GPU-accelerated rendering, optimized for mobile
  - *Designer Collaboration*: Direct designer-to-developer workflow

### Custom Component Libraries
- **[Compose Charts](https://github.com/patrykandpatryk/vico)** - Chart library for Compose
  - *Charts*: Line, bar, pie, candlestick charts with animations
  - *Customization*: Highly customizable appearance and behavior
  - *Performance*: Optimized for smooth animations and large datasets

- **[Compose Calendar](https://github.com/kizitonwose/CalendarView)** - Calendar components
  - *Flexibility*: Monthly, weekly, daily views with custom layouts
  - *Interactions*: Touch handling, selection states, custom decorations
  - *Localization*: Full internationalization support

## Networking and API Integration

### Advanced HTTP Clients
- **[Ktor Client](https://ktor.io/docs/client.html)** - Multiplatform HTTP client
  - *Multiplatform*: Shared networking code across Android, iOS
  - *Coroutines*: Built from ground up with coroutines support
  - *Features*: JSON serialization, auth, logging, MockEngine for testing

- **[Apollo GraphQL](https://www.apollographql.com/docs/android/)** - Type-safe GraphQL client
  - *Code Generation*: Compile-time safe GraphQL queries
  - *Caching*: Intelligent normalized caching
  - *Real-time*: WebSocket subscriptions for live data

### JSON Processing
- **[Kotlinx Serialization](https://github.com/Kotlin/kotlinx.serialization)** - Multiplatform serialization
  - *Performance*: Faster than reflection-based solutions
  - *Multiplatform*: Works across all Kotlin platforms
  - *Type Safety*: Compile-time verification of serialization

- **[Moshi](https://github.com/square/moshi)** with **[MoshiX](https://github.com/ZacSweers/MoshiX)**
  - *Performance*: Fast JSON parsing optimized for mobile
  - *Extensions*: Advanced adapters and sealed class support
  - *Integration*: Seamless Retrofit integration

## Data Persistence

### Database Solutions
- **[Room](https://developer.android.com/training/data-storage/room)** - SQLite abstraction with compile-time verification
  - *Type Safety*: SQL queries verified at compile time
  - *Coroutines*: Built-in suspend function support
  - *Migration*: Automated database migrations

- **[SQLDelight](https://github.com/cashapp/sqldelight)** - Generates type-safe Kotlin APIs from SQL
  - *Multiplatform*: Shared database logic across platforms
  - *SQL First*: Write actual SQL, get type-safe Kotlin
  - *Performance*: No reflection, minimal runtime overhead

### Data Storage
- **[DataStore](https://developer.android.com/topic/libraries/architecture/datastore)** - Modern data storage solution
  - *Type Safety*: Protocol Buffers or Preferences with type safety
  - *Async*: Built for coroutines and Flow
  - *Migration*: Easy migration from SharedPreferences

- **[Encrypted SharedPreferences](https://developer.android.com/reference/androidx/security/crypto/EncryptedSharedPreferences)** - Secure local storage
  - *Security*: AES-256 encryption for sensitive data
  - *Compatibility*: Works with existing SharedPreferences API
  - *Performance*: Minimal overhead for encryption/decryption

## Reactive Programming

### Coroutines and Flow
- **[Kotlinx Coroutines](https://github.com/Kotlin/kotlinx.coroutines)** - Asynchronous programming
  - *Mobile Optimized*: Main-safe operations, lifecycle awareness
  - *Structured Concurrency*: Automatic cancellation and error handling
  - *Integration*: Built-in support in Android Jetpack libraries

- **[Flow Extensions](https://github.com/hoc081098/FlowExt)** - Advanced Flow operators
  - *Utilities*: flatMapFirst, retryWithExponentialBackoff, throttleLatest
  - *Mobile Patterns*: Operators designed for mobile app scenarios
  - *Performance*: Optimized implementations for common use cases

### Reactive UI
- **[Compose State](https://developer.android.com/jetpack/compose/state)** - Reactive UI state management
  - *Declarative*: UI automatically updates with state changes
  - *Performance*: Efficient recomposition and state hoisting
  - *Integration*: Works seamlessly with ViewModels and repositories

## Testing

### Testing Frameworks
- **[MockK](https://mockk.io/)** - Mocking library designed for Kotlin
  - *Kotlin First*: Leverages Kotlin language features
  - *Coroutines*: Built-in support for testing suspend functions
  - *Annotations*: Easy setup with annotations

- **[Turbine](https://github.com/cashapp/turbine)** - Testing library for Flow
  - *Flow Testing*: Simplified testing of kotlinx.coroutines Flow
  - *Assertions*: Fluent API for Flow testing scenarios
  - *Coroutines*: Works seamlessly with coroutine test frameworks

### UI Testing
- **[Compose Testing](https://developer.android.com/jetpack/compose/testing)** - UI testing for Compose
  - *Semantics*: Test UI based on semantic properties
  - *Performance*: Faster than View-based UI tests
  - *Debugging*: Rich debugging tools and test reports

- **[Robot Pattern](https://medium.com/android-testing/robot-pattern-testing-6c55b7bb5549)** - Maintainable UI tests
  - *Readability*: Domain-specific language for test scenarios
  - *Maintenance*: Reduces test maintenance overhead
  - *Reusability*: Shared test components across test suites

### Test Utilities
- **[Kotest](https://kotest.io/)** - Kotlin test framework
  - *Styles*: Multiple testing styles (FunSpec, StringSpec, etc.)
  - *Matchers*: Rich assertion library
  - *Property Testing*: Property-based testing support

## Performance and Optimization

### Memory Management
- **[LeakCanary](https://square.github.io/leakcanary/)** - Memory leak detection
  - *Automatic*: Detects Activity and Fragment leaks automatically
  - *Analysis*: Detailed leak analysis with fix suggestions
  - *Integration*: Zero-config setup for most applications

- **[Memory Profiler Tools](https://developer.android.com/studio/profile/memory-profiler)** - Android Studio integration
  - *Real-time*: Live memory usage monitoring
  - *Heap Dumps*: Detailed memory analysis
  - *Allocation Tracking*: Track object allocations

### Performance Monitoring
- **[Firebase Performance](https://firebase.google.com/products/performance)** - App performance monitoring
  - *Automatic*: Automatic tracing of app starts, screen rendering
  - *Custom*: Custom performance traces for business logic
  - *Analytics*: Performance analytics and alerting

- **[Flipper](https://fbflipper.com/)** - Desktop debugging platform
  - *Network*: Inspect network requests and responses
  - *Database*: Browse and edit local databases
  - *Layout*: Inspect view hierarchies and properties

### Build Optimization
- **[Gradle Build Scans](https://scans.gradle.com/)** - Build performance analysis
  - *Insights*: Detailed build performance analytics
  - *Optimization*: Identify build bottlenecks
  - *Collaboration*: Share build scans with team members

## Kotlin Multiplatform Mobile

### Core KMP Libraries
- **[Kotlin Multiplatform Mobile](https://kotlinlang.org/docs/multiplatform-mobile-getting-started.html)** - Share code between iOS and Android
  - *Business Logic*: Share repositories, use cases, models
  - *Network*: Ktor client for shared networking
  - *Storage*: SQLDelight for shared database logic

### KMP UI Frameworks
- **[Compose Multiplatform](https://www.jetbrains.com/lp/compose-multiplatform/)** - Shared UI across platforms
  - *UI Sharing*: Share UI code between Android, iOS, desktop
  - *Native Performance*: Compiles to native code on each platform
  - *Ecosystem*: Growing ecosystem of multiplatform libraries

### KMP Utilities
- **[KMM-ViewModel](https://github.com/rickclephas/KMM-ViewModel)** - Shared ViewModels for KMP
  - *Architecture*: Share presentation logic across platforms
  - *Lifecycle*: Platform-specific lifecycle management
  - *State*: Shared state management patterns

- **[Koin](https://insert-koin.io/)** - Multiplatform dependency injection
  - *Lightweight*: No code generation, reflection-based
  - *DSL*: Kotlin DSL for dependency declaration
  - *Testing*: Built-in testing utilities

## Developer Tools and IDE

### Android Studio Extensions
- **[Kotlin Multiplatform Mobile Plugin](https://plugins.jetbrains.com/plugin/14936-kotlin-multiplatform-mobile)** - KMP support
  - *Project Templates*: Ready-to-use KMP project templates
  - *Code Sharing*: Tools for managing shared code
  - *Build*: Integrated build and deployment tools

- **[Rainbow Brackets](https://plugins.jetbrains.com/plugin/10080-rainbow-brackets)** - Code readability
  - *Visual Aid*: Color-coded bracket matching
  - *Kotlin*: Especially helpful with Kotlin's concise syntax
  - *Productivity*: Reduces cognitive load when reading code

### Code Quality Tools
- **[Detekt](https://detekt.github.io/detekt/)** - Static code analysis for Kotlin
  - *Code Quality*: 140+ rules for Kotlin code quality
  - *Customization*: Configurable rules and custom rule sets
  - *Integration*: Gradle plugin and IDE integration

- **[ktlint](https://ktlint.github.io/)** - Kotlin linter and formatter
  - *Consistency*: Enforces consistent Kotlin code style
  - *Integration*: Git hooks, Gradle plugin, IDE integration
  - *Standards*: Based on official Kotlin coding conventions

### Documentation
- **[Dokka](https://github.com/Kotlin/dokka)** - Documentation engine for Kotlin
  - *KDoc*: Generate documentation from KDoc comments
  - *Multiformat*: HTML, Markdown, Javadoc formats
  - *Multiplatform*: Supports Kotlin Multiplatform projects

## Learning Resources

### Official Documentation
- **[Kotlin Language Guide](https://kotlinlang.org/docs/home.html)** - Comprehensive language documentation
- **[Android Kotlin Guide](https://developer.android.com/kotlin)** - Android-specific Kotlin guidance
- **[Jetpack Compose Guide](https://developer.android.com/jetpack/compose)** - Modern UI toolkit documentation
- **[Kotlin Multiplatform Guide](https://kotlinlang.org/docs/multiplatform.html)** - Cross-platform development

### Books and Courses
- **\"Effective Kotlin\" by Marcin Moskala** - Best practices for Kotlin development
- **\"Kotlin Coroutines\" by Marcin Moskala** - Deep dive into asynchronous programming
- **\"Android Development with Kotlin\" by Marcin Moskala** - Comprehensive Android guide
- **Google Codelabs** - Hands-on tutorials for Android and Kotlin

### Video Learning
- **[Philipp Lackner YouTube](https://youtube.com/@PhilippLackner)** - 150k+ subscribers, modern Android development
- **[Android Developers YouTube](https://youtube.com/@AndroidDevelopers)** - Official Google content
- **[KotlinConf Talks](https://youtube.com/kotlinconf)** - Conference presentations from Kotlin experts
- **[Talking Kotlin Podcast](https://talkingkotlin.com/)** - Official JetBrains podcast

### Community Resources
- **[r/Kotlin](https://reddit.com/r/Kotlin)** - Active Reddit community
- **[Kotlin Slack](https://kotlinlang.slack.com/)** - Official Slack workspace
- **[Android Weekly](https://androidweekly.net/)** - Newsletter with Kotlin content
- **[Kotlin Weekly](http://kotlinweekly.net/)** - Dedicated Kotlin newsletter

## Production Examples

### Open Source Mobile Apps
- **[Tivi](https://github.com/chrisbanes/tivi)** - TV show tracking app showcasing modern Android development
  - *Architecture*: MVVM with Jetpack Compose
  - *Libraries*: Room, Retrofit, Hilt, Coroutines
  - *Patterns*: Production-ready architecture patterns

- **[Now in Android](https://github.com/android/nowinandroid)** - Google's reference Android app
  - *Modern Stack*: Latest Android development practices
  - *Architecture*: Multi-module, clean architecture
  - *Testing*: Comprehensive testing strategy

### Company Open Source Projects
- **[Slack Circuit](https://github.com/slackhq/circuit)** - Production architecture framework from Slack
- **[Square Workflow](https://github.com/square/workflow-kotlin)** - State machine framework from Square
- **[Netflix DGS Framework](https://github.com/Netflix/dgs-framework)** - GraphQL framework (JVM, not mobile-specific but Kotlin-friendly)

### Architecture Examples
- **[Android Architecture Samples](https://github.com/android/architecture-samples)** - Google's official architecture examples
- **[Kotlin Multiplatform Samples](https://github.com/Kotlin/kmm-production-sample)** - Production-ready KMP examples
- **[Compose Samples](https://github.com/android/compose-samples)** - Jetpack Compose examples

## Development Environment Setup

### Recommended IDE Setup
```kotlin
// Gradle configuration for modern Kotlin Android project
android {
    compileSdk 34

    compileOptions {
        sourceCompatibility JavaVersion.VERSION_17
        targetCompatibility JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = \"17\"
        freeCompilerArgs += listOf(
            \"-opt-in=kotlin.RequiresOptIn\",
            \"-opt-in=kotlinx.coroutines.ExperimentalCoroutinesApi\",
            \"-opt-in=androidx.compose.material3.ExperimentalMaterial3Api\"
        )
    }

    buildFeatures {
        compose = true
    }

    composeOptions {
        kotlinCompilerExtensionVersion = \"1.5.8\"
    }
}

dependencies {
    // Core Android + Kotlin
    implementation(\"androidx.core:core-ktx:1.12.0\")
    implementation(\"androidx.lifecycle:lifecycle-runtime-ktx:2.7.0\")
    implementation(\"androidx.activity:activity-compose:1.8.2\")

    // Jetpack Compose
    implementation(platform(\"androidx.compose:compose-bom:2024.02.00\"))
    implementation(\"androidx.compose.ui:ui\")
    implementation(\"androidx.compose.material3:material3\")
    implementation(\"androidx.compose.ui:ui-tooling-preview\")

    // Architecture Components
    implementation(\"androidx.lifecycle:lifecycle-viewmodel-compose:2.7.0\")
    implementation(\"androidx.navigation:navigation-compose:2.7.6\")

    // Dependency Injection
    implementation(\"com.google.dagger:hilt-android:2.48.1\")
    kapt(\"com.google.dagger:hilt-compiler:2.48.1\")
    implementation(\"androidx.hilt:hilt-navigation-compose:1.1.0\")

    // Networking
    implementation(\"com.squareup.retrofit2:retrofit:2.9.0\")
    implementation(\"com.squareup.retrofit2:converter-moshi:2.9.0\")
    implementation(\"com.squareup.okhttp3:logging-interceptor:4.12.0\")

    // Async Programming
    implementation(\"org.jetbrains.kotlinx:kotlinx-coroutines-android:1.7.3\")

    // Testing
    testImplementation(\"junit:junit:4.13.2\")
    testImplementation(\"io.mockk:mockk:1.13.8\")
    testImplementation(\"org.jetbrains.kotlinx:kotlinx-coroutines-test:1.7.3\")
    testImplementation(\"app.cash.turbine:turbine:1.0.0\")

    androidTestImplementation(\"androidx.test.ext:junit:1.1.5\")
    androidTestImplementation(\"androidx.compose.ui:ui-test-junit4\")

    debugImplementation(\"androidx.compose.ui:ui-tooling\")
    debugImplementation(\"androidx.compose.ui:ui-test-manifest\")
}
```

## Contributing to the Ecosystem

### How to Contribute
1. **Identify Gaps**: Find missing tools or libraries in the mobile development workflow
2. **Start Small**: Contribute to existing projects before creating new ones
3. **Focus on Mobile**: Consider mobile-specific constraints (battery, memory, network)
4. **Document Well**: Provide clear documentation and examples
5. **Test Thoroughly**: Ensure production-ready quality

### Mobile-Specific Considerations
- **Battery Life**: Optimize for minimal battery usage
- **Memory Constraints**: Consider memory usage in library design
- **Network Efficiency**: Handle poor network conditions gracefully
- **User Experience**: Prioritize smooth, responsive user interfaces
- **Platform Integration**: Work well with Android/iOS platform features

## Future of Kotlin Mobile Development

### Emerging Trends (2024-2025)
- **Compose Multiplatform**: Shared UI development across platforms
- **Kotlin Multiplatform Mobile**: Increased adoption for business logic sharing
- **AI Integration**: ML Kit and TensorFlow Lite integration patterns
- **Wearable Development**: Wear OS development with Kotlin and Compose
- **Foldable Devices**: Adaptive UI patterns for foldable and large screen devices

### Performance Evolution
- **Compilation**: Kotlin/Native improvements for better mobile performance
- **Runtime**: Reduced memory footprint and faster startup times
- **Tooling**: Better profiling and debugging tools for Kotlin mobile apps
- **Cross-Platform**: Improved performance parity between native and shared code

---

## Conclusion

This awesome list represents the current state of Kotlin mobile development, focusing on production-ready tools and patterns. The ecosystem continues to evolve rapidly, with new libraries and frameworks emerging regularly.

For the most up-to-date information:
- Follow [Kotlin Blog](https://blog.jetbrains.com/kotlin/) for official updates
- Monitor [Android Developers Blog](https://android-developers.googleblog.com/) for Android-specific developments
- Join [Kotlin Slack](https://kotlinlang.slack.com/) for community discussions
- Watch [KotlinConf](https://kotlinconf.com/) for annual ecosystem updates

**Key Takeaway**: The Kotlin mobile ecosystem prioritizes developer experience, type safety, and performance. Choose libraries that align with these principles and have active maintenance and community support.

---

*This document is a living resource. Contributions and updates are welcome to keep it current with the rapidly evolving Kotlin mobile ecosystem.*