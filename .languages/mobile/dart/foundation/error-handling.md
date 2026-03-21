# Dart Error Handling

> Comprehensive guide to error handling patterns and exception management in Dart and Flutter mobile applications.

## Error Handling Philosophy

Dart uses **exceptions** for error handling rather than error codes. The language encourages explicit error handling through try-catch blocks and provides rich exception types for different error scenarios.

### Key Principles
1. **Fail Fast**: Detect errors as early as possible
2. **Explicit Handling**: Use try-catch for expected exceptions
3. **Meaningful Messages**: Provide clear, actionable error information
4. **Recovery Strategies**: Implement appropriate recovery mechanisms
5. **Resource Cleanup**: Always clean up resources in finally blocks

---

## Exception Hierarchy

### Base Exception Types

```dart
// All exceptions implement Exception or Error
abstract class Exception {}
abstract class Error {}

// Common exception types
FormatException      // String parsing errors
ArgumentError        // Invalid arguments
StateError          // Invalid state operations
UnsupportedError    // Unsupported operations
UnimplementedError  // Not yet implemented
TypeError           // Type-related errors
NoSuchMethodError   // Method doesn't exist
RangeError          // Index out of bounds
```

### Custom Exception Creation

```dart
// Simple custom exception
class ValidationException implements Exception {
  final String message;
  const ValidationException(this.message);

  @override
  String toString() => 'ValidationException: $message';
}

// Rich exception with additional data
class NetworkException implements Exception {
  final String message;
  final int? statusCode;
  final String? endpoint;
  final DateTime timestamp;

  NetworkException(
    this.message, {
    this.statusCode,
    this.endpoint,
  }) : timestamp = DateTime.now();

  @override
  String toString() {
    return 'NetworkException: $message'
           '${statusCode != null ? ' (Status: $statusCode)' : ''}'
           '${endpoint != null ? ' at $endpoint' : ''}';
  }
}

// Exception with nested causes
class ServiceException implements Exception {
  final String message;
  final Object? cause;
  final StackTrace? stackTrace;

  ServiceException(this.message, [this.cause, this.stackTrace]);

  @override
  String toString() {
    var buffer = StringBuffer('ServiceException: $message');
    if (cause != null) {
      buffer.write('\nCaused by: $cause');
    }
    return buffer.toString();
  }
}
```

---

## Basic Exception Handling

### Try-Catch Fundamentals

```dart
// Basic try-catch
void basicErrorHandling() {
  try {
    riskyOperation();
  } catch (e) {
    print('Error occurred: $e');
  }
}

// Catch with stack trace
void withStackTrace() {
  try {
    riskyOperation();
  } catch (e, stackTrace) {
    print('Error: $e');
    print('Stack trace: $stackTrace');
    // Log to crash reporting service
    crashReporter.recordError(e, stackTrace);
  }
}

// Multiple catch blocks
void multipleCatch() {
  try {
    riskyOperation();
  } on FormatException catch (e) {
    print('Format error: ${e.message}');
  } on ArgumentError catch (e) {
    print('Argument error: ${e.message}');
  } catch (e) {
    print('Unknown error: $e');
  }
}

// Finally block for cleanup
void withCleanup() {
  Resource? resource;
  try {
    resource = acquireResource();
    useResource(resource);
  } catch (e) {
    handleError(e);
  } finally {
    resource?.dispose(); // Always cleanup
  }
}
```

### Exception Filtering and Conditions

```dart
// Catch specific exceptions with conditions
void conditionalCatch() {
  try {
    processUserInput(input);
  } on ValidationException catch (e) when (e.message.contains('email')) {
    showEmailValidationError(e.message);
  } on ValidationException catch (e) {
    showGenericValidationError(e.message);
  } on NetworkException catch (e) when (e.statusCode == 401) {
    redirectToLogin();
  } on NetworkException catch (e) when (e.statusCode! >= 500) {
    showServerErrorMessage();
  }
}

// Re-throwing exceptions
void processWithRethrow() {
  try {
    criticalOperation();
  } catch (e) {
    logError(e);
    // Re-throw to let caller handle
    rethrow;
  }
}
```

---

## Async Error Handling

### Future Error Handling

```dart
// Basic async error handling
Future<String> fetchUserData(String userId) async {
  try {
    final response = await httpClient.get('/users/$userId');
    if (response.statusCode == 200) {
      return response.body;
    }
    throw NetworkException(
      'Failed to fetch user data',
      statusCode: response.statusCode,
      endpoint: '/users/$userId',
    );
  } catch (e) {
    if (e is NetworkException) rethrow;
    throw ServiceException('User data fetch failed', e);
  }
}

// Future error handling with catchError
Future<User?> fetchUserSafely(String userId) {
  return fetchUserData(userId)
      .then((data) => User.fromJson(data))
      .catchError((error) {
        if (error is NetworkException && error.statusCode == 404) {
          return null; // User not found
        }
        throw error; // Re-throw other errors
      });
}

// Multiple async operations with error handling
Future<UserProfile> buildUserProfile(String userId) async {
  try {
    final results = await Future.wait([
      fetchUserData(userId),
      fetchUserPreferences(userId),
      fetchUserActivity(userId),
    ]);

    return UserProfile(
      user: User.fromJson(results[0]),
      preferences: Preferences.fromJson(results[1]),
      activity: Activity.fromJson(results[2]),
    );
  } catch (e) {
    throw ServiceException('Failed to build user profile', e);
  }
}
```

### Stream Error Handling

```dart
// Stream error handling
Stream<String> messageStream() async* {
  try {
    await for (final message in websocketStream) {
      // Process and validate message
      if (isValidMessage(message)) {
        yield processMessage(message);
      }
    }
  } catch (e) {
    // Log error and yield error state
    logError('Stream error', e);
    yield 'ERROR: Connection lost';
  }
}

// Stream with error recovery
Stream<List<Item>> itemStreamWithRecovery() async* {
  int retryCount = 0;
  const maxRetries = 3;

  while (retryCount < maxRetries) {
    try {
      await for (final items in dataService.getItemStream()) {
        retryCount = 0; // Reset on success
        yield items;
      }
    } catch (e) {
      retryCount++;
      if (retryCount >= maxRetries) {
        yield []; // Emit empty list as fallback
        throw ServiceException('Max retries exceeded', e);
      }

      // Wait before retry
      await Future.delayed(Duration(seconds: retryCount * 2));
    }
  }
}

// Stream subscription error handling
void subscribeToMessages() {
  messageStream().listen(
    (message) => handleMessage(message),
    onError: (error, stackTrace) {
      logError('Message stream error', error, stackTrace);
      // Implement reconnection logic
      scheduleReconnection();
    },
    onDone: () => print('Stream completed'),
  );
}
```

---

## Mobile-Specific Error Handling

### Flutter Widget Error Handling

```dart
// Error handling in StatefulWidget
class UserProfileWidget extends StatefulWidget {
  final String userId;

  const UserProfileWidget({super.key, required this.userId});

  @override
  State<UserProfileWidget> createState() => _UserProfileWidgetState();
}

class _UserProfileWidgetState extends State<UserProfileWidget> {
  User? user;
  String? errorMessage;
  bool isLoading = true;

  @override
  void initState() {
    super.initState();
    loadUser();
  }

  Future<void> loadUser() async {
    try {
      setState(() {
        isLoading = true;
        errorMessage = null;
      });

      final userData = await userService.fetchUser(widget.userId);

      if (mounted) {
        setState(() {
          user = userData;
          isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          errorMessage = _getErrorMessage(e);
          isLoading = false;
        });
      }
    }
  }

  String _getErrorMessage(Object error) {
    if (error is NetworkException) {
      switch (error.statusCode) {
        case 404:
          return 'User not found';
        case 401:
          return 'Authentication required';
        case 403:
          return 'Access denied';
        case >= 500:
          return 'Server error. Please try again later.';
        default:
          return 'Network error occurred';
      }
    }
    return 'An unexpected error occurred';
  }

  @override
  Widget build(BuildContext context) {
    if (isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (errorMessage != null) {
      return ErrorWidget(
        message: errorMessage!,
        onRetry: loadUser,
      );
    }

    return UserProfileView(user: user!);
  }
}

// Reusable error widget
class ErrorWidget extends StatelessWidget {
  final String message;
  final VoidCallback? onRetry;

  const ErrorWidget({
    super.key,
    required this.message,
    this.onRetry,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.error_outline,
            size: 64,
            color: Theme.of(context).colorScheme.error,
          ),
          const SizedBox(height: 16),
          Text(
            message,
            style: Theme.of(context).textTheme.titleMedium,
            textAlign: TextAlign.center,
          ),
          if (onRetry != null) ...[
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: onRetry,
              child: const Text('Retry'),
            ),
          ],
        ],
      ),
    );
  }
}
```

### Global Error Handling

```dart
// Global error handler for Flutter apps
class GlobalErrorHandler {
  static void initialize() {
    // Catch Flutter framework errors
    FlutterError.onError = (FlutterErrorDetails details) {
      FlutterError.presentError(details);
      _reportError(details.exception, details.stack);
    };

    // Catch errors outside of Flutter framework
    PlatformDispatcher.instance.onError = (error, stack) {
      _reportError(error, stack);
      return true;
    };
  }

  static void _reportError(Object error, StackTrace? stack) {
    // Log locally
    debugPrint('Global error: $error');
    if (stack != null) {
      debugPrint('Stack trace: $stack');
    }

    // Report to crash analytics (Firebase Crashlytics, etc.)
    crashAnalytics.recordError(error, stack);

    // Show user-friendly message in production
    if (kReleaseMode) {
      _showErrorSnackBar('An error occurred. Please try again.');
    }
  }

  static void _showErrorSnackBar(String message) {
    final context = navigatorKey.currentContext;
    if (context != null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(message)),
      );
    }
  }
}
```

### Platform-Specific Error Handling

```dart
// Platform channel error handling
class PlatformService {
  static const platform = MethodChannel('com.example.app/platform');

  static Future<String> getNativeData() async {
    try {
      final result = await platform.invokeMethod('getNativeData');
      return result as String;
    } on PlatformException catch (e) {
      throw _handlePlatformException(e);
    } catch (e) {
      throw ServiceException('Platform communication failed', e);
    }
  }

  static Exception _handlePlatformException(PlatformException e) {
    switch (e.code) {
      case 'PERMISSION_DENIED':
        return PermissionException('Required permission not granted');
      case 'NOT_AVAILABLE':
        return ServiceException('Feature not available on this device');
      case 'NETWORK_ERROR':
        return NetworkException('Network connection required');
      default:
        return ServiceException('Platform error: ${e.message}', e);
    }
  }
}

// Permission handling
class PermissionHandler {
  static Future<bool> requestCameraPermission() async {
    try {
      final status = await Permission.camera.request();
      return status == PermissionStatus.granted;
    } catch (e) {
      throw PermissionException('Failed to request camera permission', e);
    }
  }

  static Future<void> handlePermissionDenied(BuildContext context) async {
    await showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Permission Required'),
        content: const Text(
          'Camera permission is required for this feature. '
          'Please grant permission in app settings.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              openAppSettings();
            },
            child: const Text('Settings'),
          ),
        ],
      ),
    );
  }
}
```

---

## Error Handling Patterns

### Result Pattern

```dart
// Result type for explicit error handling
abstract class Result<T, E> {
  const Result();

  bool get isSuccess => this is Success<T, E>;
  bool get isError => this is Error<T, E>;

  T get value => (this as Success<T, E>).value;
  E get error => (this as Error<T, E>).error;
}

class Success<T, E> extends Result<T, E> {
  final T value;
  const Success(this.value);
}

class Error<T, E> extends Result<T, E> {
  final E error;
  const Error(this.error);
}

// Usage in services
class UserService {
  Future<Result<User, String>> fetchUser(String id) async {
    try {
      final user = await apiClient.getUser(id);
      return Success(user);
    } catch (e) {
      return Error('Failed to fetch user: $e');
    }
  }
}

// Usage in UI
class UserController {
  Future<void> loadUser(String id) async {
    final result = await userService.fetchUser(id);

    switch (result) {
      case Success(value: final user):
        updateUI(user);
      case Error(error: final errorMessage):
        showError(errorMessage);
    }
  }
}
```

### Either Pattern

```dart
// Either type for two possible outcomes
abstract class Either<L, R> {
  const Either();

  bool get isLeft => this is Left<L, R>;
  bool get isRight => this is Right<L, R>;
}

class Left<L, R> extends Either<L, R> {
  final L value;
  const Left(this.value);
}

class Right<L, R> extends Either<L, R> {
  final R value;
  const Right(this.value);
}

// Extension methods for convenience
extension EitherExtensions<L, R> on Either<L, R> {
  T fold<T>(T Function(L) onLeft, T Function(R) onRight) {
    return switch (this) {
      Left(value: final l) => onLeft(l),
      Right(value: final r) => onRight(r),
    };
  }

  Either<L, T> map<T>(T Function(R) mapper) {
    return switch (this) {
      Left(value: final l) => Left(l),
      Right(value: final r) => Right(mapper(r)),
    };
  }
}

// Usage example
Future<Either<String, List<Post>>> fetchPosts() async {
  try {
    final posts = await apiClient.getPosts();
    return Right(posts);
  } catch (e) {
    return Left('Failed to fetch posts: $e');
  }
}
```

### Circuit Breaker Pattern

```dart
// Circuit breaker for handling repeated failures
class CircuitBreaker {
  final int failureThreshold;
  final Duration timeout;
  final Duration resetTimeout;

  int _failureCount = 0;
  DateTime? _lastFailureTime;
  CircuitBreakerState _state = CircuitBreakerState.closed;

  CircuitBreaker({
    this.failureThreshold = 5,
    this.timeout = const Duration(seconds: 30),
    this.resetTimeout = const Duration(minutes: 1),
  });

  Future<T> execute<T>(Future<T> Function() operation) async {
    if (_state == CircuitBreakerState.open) {
      if (_shouldAttemptReset()) {
        _state = CircuitBreakerState.halfOpen;
      } else {
        throw CircuitBreakerException('Circuit breaker is open');
      }
    }

    try {
      final result = await operation();
      _onSuccess();
      return result;
    } catch (e) {
      _onFailure();
      rethrow;
    }
  }

  void _onSuccess() {
    _failureCount = 0;
    _state = CircuitBreakerState.closed;
  }

  void _onFailure() {
    _failureCount++;
    _lastFailureTime = DateTime.now();

    if (_failureCount >= failureThreshold) {
      _state = CircuitBreakerState.open;
    }
  }

  bool _shouldAttemptReset() {
    return _lastFailureTime != null &&
           DateTime.now().difference(_lastFailureTime!) > resetTimeout;
  }
}

enum CircuitBreakerState { closed, open, halfOpen }

class CircuitBreakerException implements Exception {
  final String message;
  CircuitBreakerException(this.message);
}
```

---

## Testing Error Handling

### Unit Testing Exceptions

```dart
import 'package:test/test.dart';

void main() {
  group('UserService Error Handling', () {
    late UserService userService;
    late MockApiClient mockApiClient;

    setUp(() {
      mockApiClient = MockApiClient();
      userService = UserService(mockApiClient);
    });

    test('should throw NetworkException on network error', () async {
      // Arrange
      when(mockApiClient.getUser(any))
          .thenThrow(HttpException('Network error'));

      // Act & Assert
      await expectLater(
        () => userService.fetchUser('123'),
        throwsA(isA<NetworkException>()),
      );
    });

    test('should handle 404 error correctly', () async {
      // Arrange
      when(mockApiClient.getUser(any))
          .thenThrow(HttpException('Not found', uri: Uri.parse('/users/123')));

      // Act & Assert
      final exception = await expectLater(
        () => userService.fetchUser('123'),
        throwsA(isA<NetworkException>()
            .having((e) => e.statusCode, 'statusCode', 404)),
      );
    });

    test('should retry on transient failures', () async {
      // Arrange
      when(mockApiClient.getUser(any))
          .thenThrow(HttpException('Temporary failure'))
          .thenReturn(User(id: '123', name: 'John'));

      // Act
      final user = await userService.fetchUserWithRetry('123');

      // Assert
      expect(user.name, 'John');
      verify(mockApiClient.getUser('123')).called(2);
    });
  });

  group('Error Handling Utilities', () {
    test('should format error messages correctly', () {
      final exception = NetworkException(
        'Connection failed',
        statusCode: 500,
        endpoint: '/api/users',
      );

      expect(
        exception.toString(),
        contains('Connection failed'),
      );
      expect(
        exception.toString(),
        contains('Status: 500'),
      );
    });
  });
}
```

### Widget Error Testing

```dart
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('UserProfileWidget Error Handling', () {
    testWidgets('should show error message when loading fails', (tester) async {
      // Arrange
      final mockUserService = MockUserService();
      when(mockUserService.fetchUser(any))
          .thenThrow(NetworkException('Network error'));

      // Act
      await tester.pumpWidget(
        MaterialApp(
          home: UserProfileWidget(
            userId: '123',
            userService: mockUserService,
          ),
        ),
      );

      await tester.pumpAndSettle();

      // Assert
      expect(find.text('Network error occurred'), findsOneWidget);
      expect(find.byType(ElevatedButton), findsOneWidget); // Retry button
    });

    testWidgets('should retry loading when retry button is pressed', (tester) async {
      // Arrange
      final mockUserService = MockUserService();
      when(mockUserService.fetchUser(any))
          .thenThrow(NetworkException('Network error'))
          .thenReturn(User(id: '123', name: 'John'));

      await tester.pumpWidget(
        MaterialApp(
          home: UserProfileWidget(
            userId: '123',
            userService: mockUserService,
          ),
        ),
      );
      await tester.pumpAndSettle();

      // Act
      await tester.tap(find.text('Retry'));
      await tester.pumpAndSettle();

      // Assert
      expect(find.text('John'), findsOneWidget);
      verify(mockUserService.fetchUser('123')).called(2);
    });
  });
}
```

---

## Error Reporting and Logging

### Structured Logging

```dart
// Custom logger with structured error logging
class Logger {
  static const String _tag = 'MyApp';

  static void info(String message, [Map<String, dynamic>? context]) {
    _log('INFO', message, context);
  }

  static void warning(String message, [Map<String, dynamic>? context]) {
    _log('WARNING', message, context);
  }

  static void error(String message, Object? error, StackTrace? stackTrace) {
    final context = <String, dynamic>{
      if (error != null) 'error': error.toString(),
      if (stackTrace != null) 'stackTrace': stackTrace.toString(),
    };
    _log('ERROR', message, context);
  }

  static void _log(String level, String message, Map<String, dynamic>? context) {
    final timestamp = DateTime.now().toIso8601String();
    final logEntry = {
      'timestamp': timestamp,
      'level': level,
      'tag': _tag,
      'message': message,
      if (context != null) ...context,
    };

    // Print to console in debug mode
    if (kDebugMode) {
      debugPrint('[$level] $_tag: $message');
      if (context != null) {
        debugPrint('Context: $context');
      }
    }

    // Send to logging service in production
    if (kReleaseMode) {
      _sendToLoggingService(logEntry);
    }
  }

  static void _sendToLoggingService(Map<String, dynamic> logEntry) {
    // Implementation for sending logs to remote service
  }
}
```

### Crash Reporting Integration

```dart
// Firebase Crashlytics integration
class CrashReporter {
  static Future<void> initialize() async {
    await FirebaseCrashlytics.instance.setCrashlyticsCollectionEnabled(true);

    // Set custom keys for better crash analysis
    await FirebaseCrashlytics.instance.setCustomKey('flutter_version', '3.24.0');
    await FirebaseCrashlytics.instance.setCustomKey('dart_version', '3.9.2');
  }

  static Future<void> recordError(
    Object error,
    StackTrace? stackTrace, {
    Map<String, String>? context,
    bool fatal = false,
  }) async {
    // Add context information
    if (context != null) {
      for (final entry in context.entries) {
        await FirebaseCrashlytics.instance.setCustomKey(entry.key, entry.value);
      }
    }

    // Record the error
    await FirebaseCrashlytics.instance.recordError(
      error,
      stackTrace,
      fatal: fatal,
      information: context?.entries.map((e) => '${e.key}: ${e.value}').toList() ?? [],
    );
  }

  static Future<void> log(String message) async {
    await FirebaseCrashlytics.instance.log(message);
  }

  static Future<void> setUserId(String userId) async {
    await FirebaseCrashlytics.instance.setUserIdentifier(userId);
  }
}
```

---

## Error Handling Best Practices

### Mobile Development Guidelines

1. **User Experience First**
   - Provide meaningful error messages
   - Offer recovery actions when possible
   - Use loading states during error recovery

2. **Network Error Handling**
   - Distinguish between different network states
   - Implement retry mechanisms with backoff
   - Cache data for offline scenarios

3. **Resource Management**
   - Always cleanup resources in finally blocks
   - Handle permission errors gracefully
   - Manage memory during error scenarios

4. **Testing Strategy**
   - Test error scenarios explicitly
   - Mock external dependencies for error testing
   - Verify error messages and recovery mechanisms

5. **Production Monitoring**
   - Implement comprehensive crash reporting
   - Log errors with sufficient context
   - Monitor error rates and patterns

### Code Quality Checklist

```dart
// ✅ Good error handling example
class UserRepository {
  Future<User> getUser(String id) async {
    try {
      validateUserId(id);
      final userData = await _fetchUserData(id);
      return User.fromJson(userData);
    } on ValidationException catch (e) {
      throw UserValidationException('Invalid user ID: ${e.message}');
    } on NetworkException catch (e) {
      throw UserFetchException('Failed to fetch user data', e);
    } catch (e, stackTrace) {
      Logger.error('Unexpected error in getUser', e, stackTrace);
      throw UserRepositoryException('Unexpected error occurred', e);
    }
  }
}

// ❌ Poor error handling example
class BadUserRepository {
  Future<User> getUser(String id) async {
    final userData = await _fetchUserData(id); // No error handling
    return User.fromJson(userData); // Could throw parsing error
  }
}
```

---

**Last Updated**: September 13, 2025
**Dart Version**: 3.9.2+
**Flutter Compatibility**: All current Flutter versions
**Mobile Focus**: Android and iOS error handling patterns included