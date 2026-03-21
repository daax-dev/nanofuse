# Dart Concurrency and Asynchronous Programming

> Comprehensive guide to Dart's concurrency model, isolates, async/await, and threading for mobile application development.

## Dart Concurrency Model

Dart uses a **single-threaded** execution model with an **event loop**, similar to JavaScript. All Dart code runs in **isolates** - independent workers with their own memory and event loop.

### Key Concepts
- **Single-threaded**: No shared memory between concurrent operations
- **Event Loop**: Processes events and microtasks sequentially
- **Isolates**: Independent execution contexts
- **Futures**: Represent eventual completion of asynchronous operations
- **Streams**: Sequences of asynchronous events

---

## Event Loop and Execution Model

### Event Loop Structure

```dart
// The Dart event loop processes:
// 1. Microtask queue (higher priority)
// 2. Event queue (lower priority)

void demonstrateEventLoop() {
  print('1. Synchronous start');

  // Event queue
  Future(() => print('4. Future event'));

  // Microtask queue (higher priority)
  scheduleMicrotask(() => print('3. Microtask'));

  // Synchronous operation
  print('2. Synchronous end');

  // Output order:
  // 1. Synchronous start
  // 2. Synchronous end
  // 3. Microtask
  // 4. Future event
}
```

### Microtasks vs Events

```dart
void understandingQueues() {
  print('Start');

  // Events (Timer, HTTP, File I/O)
  Timer(Duration.zero, () => print('Timer event'));
  Future(() => print('Future event'));

  // Microtasks (higher priority)
  scheduleMicrotask(() => print('Microtask 1'));
  Future.microtask(() => print('Microtask 2'));

  // More microtasks can be scheduled by existing microtasks
  scheduleMicrotask(() {
    print('Microtask 3');
    scheduleMicrotask(() => print('Nested microtask'));
  });

  print('End');
}
```

---

## Futures and Async/Await

### Future Fundamentals

```dart
// Creating Futures
Future<String> fetchUserName() async {
  await Future.delayed(Duration(seconds: 1));
  return 'John Doe';
}

// Using Futures with then/catchError
void futureThenPattern() {
  fetchUserName()
      .then((name) => print('User: $name'))
      .catchError((error) => print('Error: $error'));
}

// Using async/await (preferred)
Future<void> asyncAwaitPattern() async {
  try {
    final name = await fetchUserName();
    print('User: $name');
  } catch (error) {
    print('Error: $error');
  }
}
```

### Future Combinators

```dart
// Multiple concurrent operations
Future<UserProfile> fetchUserProfile(String userId) async {
  try {
    // Run operations concurrently
    final results = await Future.wait([
      fetchUserData(userId),
      fetchUserPreferences(userId),
      fetchUserStats(userId),
    ]);

    return UserProfile(
      user: User.fromJson(results[0]),
      preferences: Preferences.fromJson(results[1]),
      stats: Stats.fromJson(results[2]),
    );
  } catch (e) {
    throw Exception('Failed to fetch user profile: $e');
  }
}

// Future.any - first to complete
Future<String> fetchFromMultipleServers() async {
  return await Future.any([
    fetchFromServer('https://api1.example.com'),
    fetchFromServer('https://api2.example.com'),
    fetchFromServer('https://api3.example.com'),
  ]);
}

// Timeout handling
Future<String> fetchWithTimeout() async {
  try {
    return await fetchUserData('123').timeout(Duration(seconds: 5));
  } on TimeoutException {
    throw Exception('Request timed out');
  }
}

// Sequential processing with error handling
Future<List<ProcessedItem>> processItemsSequentially(List<Item> items) async {
  final results = <ProcessedItem>[];

  for (final item in items) {
    try {
      final processed = await processItem(item);
      results.add(processed);
    } catch (e) {
      // Log error but continue processing
      print('Failed to process item ${item.id}: $e');
    }
  }

  return results;
}
```

---

## Streams and Stream Processing

### Stream Fundamentals

```dart
// Creating streams
Stream<int> countStream(int max) async* {
  for (int i = 0; i <= max; i++) {
    await Future.delayed(Duration(milliseconds: 100));
    yield i;
  }
}

// Stream from Future
Stream<String> singleValueStream() async* {
  final value = await fetchData();
  yield value;
}

// Infinite stream
Stream<DateTime> timestampStream() async* {
  while (true) {
    yield DateTime.now();
    await Future.delayed(Duration(seconds: 1));
  }
}
```

### Stream Operations

```dart
// Stream transformations
Future<void> streamOperations() async {
  final numbers = Stream.periodic(
    Duration(milliseconds: 100),
    (count) => count,
  ).take(10);

  // Map transformation
  final doubled = numbers.map((n) => n * 2);

  // Filtering
  final evenNumbers = numbers.where((n) => n.isEven);

  // Reduce operations
  final sum = await numbers.fold(0, (sum, n) => sum + n);

  // Listen to stream
  await for (final number in doubled) {
    print('Doubled: $number');
  }
}

// Complex stream processing
class DataProcessor {
  Stream<ProcessedData> processDataStream(Stream<RawData> input) async* {
    await for (final rawData in input) {
      try {
        // Validate data
        if (!isValidData(rawData)) {
          continue;
        }

        // Transform data
        final processed = await transformData(rawData);

        // Additional processing
        if (processed.isSignificant) {
          yield processed;
        }
      } catch (e) {
        // Log error and continue
        print('Processing error: $e');
      }
    }
  }

  // Stream with backpressure handling
  Stream<T> throttleStream<T>(Stream<T> source, Duration interval) async* {
    T? lastValue;
    Timer? timer;

    await for (final value in source) {
      lastValue = value;

      timer?.cancel();
      timer = Timer(interval, () {
        if (lastValue != null) {
          // In real implementation, we'd need a different approach
          // This is conceptual demonstration
        }
      });
    }
  }
}
```

### Stream Controllers

```dart
// StreamController for custom streams
class EventBus {
  final StreamController<Event> _controller = StreamController<Event>.broadcast();

  Stream<Event> get events => _controller.stream;

  void publishEvent(Event event) {
    if (!_controller.isClosed) {
      _controller.add(event);
    }
  }

  void publishError(Object error) {
    if (!_controller.isClosed) {
      _controller.addError(error);
    }
  }

  void close() {
    _controller.close();
  }
}

// Usage
class NotificationService {
  final EventBus _eventBus = EventBus();

  Stream<NotificationEvent> get notifications =>
      _eventBus.events.where((event) => event is NotificationEvent).cast();

  void sendNotification(String message) {
    _eventBus.publishEvent(NotificationEvent(message));
  }
}
```

---

## Isolates

### Understanding Isolates

```dart
import 'dart:isolate';

// Isolates provide true parallelism in Dart
// Each isolate has its own memory space
// Communication happens through message passing

// Compute function (recommended for CPU-intensive tasks)
Future<int> calculateFibonacci(int n) async {
  return await compute(_fibonacci, n);
}

int _fibonacci(int n) {
  if (n <= 1) return n;
  return _fibonacci(n - 1) + _fibonacci(n - 2);
}

// Custom isolate creation
Future<String> processDataInIsolate(List<dynamic> data) async {
  final receivePort = ReceivePort();

  await Isolate.spawn(_isolateEntry, receivePort.sendPort);

  final sendPort = await receivePort.first as SendPort;

  final responsePort = ReceivePort();
  sendPort.send([data, responsePort.sendPort]);

  final result = await responsePort.first as String;

  return result;
}

void _isolateEntry(SendPort sendPort) async {
  final receivePort = ReceivePort();
  sendPort.send(receivePort.sendPort);

  await for (final message in receivePort) {
    final data = message[0] as List<dynamic>;
    final responsePort = message[1] as SendPort;

    // Perform CPU-intensive work
    final result = processLargeDataSet(data);

    responsePort.send(result);
    break;
  }
}
```

### Mobile-Optimized Isolate Usage

```dart
// Image processing isolate for Flutter apps
class ImageProcessor {
  static Future<Uint8List> resizeImage({
    required Uint8List imageData,
    required int targetWidth,
    required int targetHeight,
  }) async {
    final parameters = {
      'imageData': imageData,
      'targetWidth': targetWidth,
      'targetHeight': targetHeight,
    };

    return await compute(_resizeImageIsolate, parameters);
  }

  static Uint8List _resizeImageIsolate(Map<String, dynamic> params) {
    final imageData = params['imageData'] as Uint8List;
    final targetWidth = params['targetWidth'] as int;
    final targetHeight = params['targetHeight'] as int;

    // Image processing logic (using image package)
    final image = img.decodeImage(imageData);
    if (image == null) throw Exception('Invalid image data');

    final resized = img.copyResize(
      image,
      width: targetWidth,
      height: targetHeight,
    );

    return Uint8List.fromList(img.encodePng(resized));
  }
}

// Background data sync isolate
class BackgroundSyncManager {
  static Isolate? _syncIsolate;
  static SendPort? _syncPort;

  static Future<void> startBackgroundSync() async {
    if (_syncIsolate != null) return;

    final receivePort = ReceivePort();
    _syncIsolate = await Isolate.spawn(_syncIsolateEntry, receivePort.sendPort);
    _syncPort = await receivePort.first as SendPort;
  }

  static Future<void> syncData(Map<String, dynamic> data) async {
    _syncPort?.send({'action': 'sync', 'data': data});
  }

  static void stopBackgroundSync() {
    _syncPort?.send({'action': 'stop'});
    _syncIsolate?.kill();
    _syncIsolate = null;
    _syncPort = null;
  }

  static void _syncIsolateEntry(SendPort mainPort) async {
    final receivePort = ReceivePort();
    mainPort.send(receivePort.sendPort);

    await for (final message in receivePort) {
      final action = message['action'] as String;

      switch (action) {
        case 'sync':
          final data = message['data'] as Map<String, dynamic>;
          await _performSync(data);
          break;
        case 'stop':
          return;
      }
    }
  }

  static Future<void> _performSync(Map<String, dynamic> data) async {
    // Background sync logic
    try {
      await uploadDataToServer(data);
      await updateLocalDatabase(data);
    } catch (e) {
      // Handle sync errors
      print('Sync error: $e');
    }
  }
}
```

---

## Mobile Development Patterns

### Flutter-Specific Concurrency

```dart
// StatefulWidget with async operations
class UserListWidget extends StatefulWidget {
  @override
  _UserListWidgetState createState() => _UserListWidgetState();
}

class _UserListWidgetState extends State<UserListWidget> {
  List<User> users = [];
  bool isLoading = false;
  String? errorMessage;

  @override
  void initState() {
    super.initState();
    loadUsers();
  }

  Future<void> loadUsers() async {
    if (!mounted) return;

    setState(() {
      isLoading = true;
      errorMessage = null;
    });

    try {
      final fetchedUsers = await userService.fetchUsers();

      if (mounted) {
        setState(() {
          users = fetchedUsers;
          isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          errorMessage = e.toString();
          isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    if (isLoading) {
      return Center(child: CircularProgressIndicator());
    }

    if (errorMessage != null) {
      return Center(
        child: Column(
          children: [
            Text('Error: $errorMessage'),
            ElevatedButton(
              onPressed: loadUsers,
              child: Text('Retry'),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      itemCount: users.length,
      itemBuilder: (context, index) => UserTile(user: users[index]),
    );
  }
}

// StreamBuilder for reactive UI
class ChatWidget extends StatelessWidget {
  final Stream<List<Message>> messageStream;

  const ChatWidget({super.key, required this.messageStream});

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<List<Message>>(
      stream: messageStream,
      builder: (context, snapshot) {
        if (snapshot.hasError) {
          return ErrorWidget('Chat error: ${snapshot.error}');
        }

        if (!snapshot.hasData) {
          return Center(child: CircularProgressIndicator());
        }

        final messages = snapshot.data!;
        return ListView.builder(
          itemCount: messages.length,
          itemBuilder: (context, index) => MessageTile(message: messages[index]),
        );
      },
    );
  }
}
```

### Background Processing

```dart
// Background task management
class BackgroundTaskManager {
  static final Map<String, Completer> _tasks = {};
  static final Map<String, Timer> _timers = {};

  // Schedule periodic background task
  static void schedulePeriodicTask(
    String taskId,
    Duration interval,
    Future<void> Function() task,
  ) {
    cancelTask(taskId);

    _timers[taskId] = Timer.periodic(interval, (_) async {
      if (_tasks.containsKey(taskId)) return; // Already running

      final completer = Completer<void>();
      _tasks[taskId] = completer;

      try {
        await task();
        completer.complete();
      } catch (e) {
        completer.completeError(e);
      } finally {
        _tasks.remove(taskId);
      }
    });
  }

  // One-time background task
  static Future<void> runBackgroundTask(
    String taskId,
    Future<void> Function() task,
  ) async {
    if (_tasks.containsKey(taskId)) {
      return _tasks[taskId]!.future;
    }

    final completer = Completer<void>();
    _tasks[taskId] = completer;

    try {
      await task();
      completer.complete();
    } catch (e) {
      completer.completeError(e);
    } finally {
      _tasks.remove(taskId);
    }

    return completer.future;
  }

  static void cancelTask(String taskId) {
    _timers[taskId]?.cancel();
    _timers.remove(taskId);
    _tasks.remove(taskId);
  }

  static void cancelAllTasks() {
    _timers.values.forEach((timer) => timer.cancel());
    _timers.clear();
    _tasks.clear();
  }
}

// Usage in mobile app
class DataSyncService {
  static void startAutoSync() {
    BackgroundTaskManager.schedulePeriodicTask(
      'data-sync',
      Duration(minutes: 15),
      _performDataSync,
    );
  }

  static Future<void> _performDataSync() async {
    try {
      // Check network connectivity
      if (!await NetworkService.isConnected()) return;

      // Perform sync operations
      await _syncUserData();
      await _syncAppSettings();
      await _uploadPendingActions();

      print('Background sync completed');
    } catch (e) {
      print('Background sync failed: $e');
    }
  }
}
```

---

## Performance Optimization

### Avoiding Common Pitfalls

```dart
// ❌ Blocking the UI thread
Future<void> badExample() async {
  // This blocks the UI thread
  for (int i = 0; i < 1000000; i++) {
    expensiveCalculation(i);
  }
}

// ✅ Using compute for CPU-intensive work
Future<void> goodExample() async {
  final result = await compute(expensiveWork, largeDataSet);
  updateUI(result);
}

List<int> expensiveWork(List<int> data) {
  // CPU-intensive work in isolate
  return data.map((n) => n * n * n).toList();
}

// ❌ Creating unnecessary Futures
Stream<String> badStreamExample() async* {
  for (int i = 0; i < 100; i++) {
    yield await Future.value('Item $i'); // Unnecessary Future
  }
}

// ✅ Direct yield for synchronous values
Stream<String> goodStreamExample() async* {
  for (int i = 0; i < 100; i++) {
    yield 'Item $i';
  }
}
```

### Memory Management

```dart
// Proper resource cleanup
class ResourceManager {
  StreamSubscription? _subscription;
  Timer? _timer;
  final StreamController _controller = StreamController();

  void startListening() {
    _subscription = dataStream.listen((data) {
      _controller.add(data);
    });

    _timer = Timer.periodic(Duration(seconds: 1), (_) {
      checkStatus();
    });
  }

  void dispose() {
    // Always cleanup resources
    _subscription?.cancel();
    _subscription = null;

    _timer?.cancel();
    _timer = null;

    _controller.close();
  }
}

// Weak references for avoiding memory leaks
class WeakNotificationManager {
  final Set<WeakReference<NotificationListener>> _listeners = {};

  void addListener(NotificationListener listener) {
    _listeners.add(WeakReference(listener));
  }

  void notifyListeners(String message) {
    // Clean up garbage collected listeners
    _listeners.removeWhere((ref) => ref.target == null);

    // Notify active listeners
    for (final ref in _listeners) {
      ref.target?.onNotification(message);
    }
  }
}
```

---

## Testing Concurrent Code

### Testing Async Operations

```dart
import 'package:test/test.dart';

void main() {
  group('Async Tests', () {
    test('should handle async operation completion', () async {
      final result = await fetchUserData('123');
      expect(result.id, '123');
    });

    test('should handle async operation timeout', () async {
      await expectLater(
        fetchUserData('timeout').timeout(Duration(milliseconds: 100)),
        throwsA(isA<TimeoutException>()),
      );
    });

    test('should handle stream events', () async {
      final events = <String>[];
      final stream = countStream(3);

      await for (final event in stream) {
        events.add('Event: $event');
      }

      expect(events, ['Event: 0', 'Event: 1', 'Event: 2', 'Event: 3']);
    });

    test('should handle concurrent operations', () async {
      final stopwatch = Stopwatch()..start();

      final results = await Future.wait([
        delayedOperation(Duration(milliseconds: 100), 'A'),
        delayedOperation(Duration(milliseconds: 100), 'B'),
        delayedOperation(Duration(milliseconds: 100), 'C'),
      ]);

      stopwatch.stop();

      expect(results, ['A', 'B', 'C']);
      expect(stopwatch.elapsedMilliseconds, lessThan(200)); // Concurrent execution
    });
  });

  group('Isolate Tests', () {
    test('should process data in isolate', () async {
      final largeList = List.generate(10000, (i) => i);
      final result = await compute(sumList, largeList);
      expect(result, 49995000); // Sum of 0..9999
    });
  });
}

// Helper functions for tests
Future<String> delayedOperation(Duration delay, String value) async {
  await Future.delayed(delay);
  return value;
}

int sumList(List<int> numbers) {
  return numbers.fold(0, (sum, n) => sum + n);
}
```

### Testing Flutter Async Widgets

```dart
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('Async Widget Tests', () {
    testWidgets('should show loading indicator initially', (tester) async {
      await tester.pumpWidget(MaterialApp(
        home: UserProfileWidget(userId: '123'),
      ));

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('should show user data after loading', (tester) async {
      await tester.pumpWidget(MaterialApp(
        home: UserProfileWidget(userId: '123'),
      ));

      // Wait for async operation to complete
      await tester.pumpAndSettle();

      expect(find.text('John Doe'), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsNothing);
    });

    testWidgets('should handle stream updates', (tester) async {
      final controller = StreamController<String>();

      await tester.pumpWidget(MaterialApp(
        home: StreamBuilder<String>(
          stream: controller.stream,
          builder: (context, snapshot) {
            if (snapshot.hasData) {
              return Text(snapshot.data!);
            }
            return Text('No data');
          },
        ),
      ));

      expect(find.text('No data'), findsOneWidget);

      controller.add('Hello');
      await tester.pump();

      expect(find.text('Hello'), findsOneWidget);

      controller.close();
    });
  });
}
```

---

## Best Practices Summary

### Mobile Development Guidelines

1. **UI Responsiveness**
   - Use `compute()` for CPU-intensive work
   - Keep UI thread free for rendering
   - Show loading indicators for long operations

2. **Memory Management**
   - Always cancel subscriptions and timers
   - Use weak references to avoid memory leaks
   - Clean up resources in dispose methods

3. **Error Handling**
   - Handle errors in async operations
   - Provide user-friendly error messages
   - Implement retry mechanisms for network operations

4. **Performance**
   - Use `await` only when necessary
   - Prefer `Stream.listen()` over `await for` for long-lived streams
   - Batch operations when possible

5. **Testing**
   - Test async operations thoroughly
   - Use `pumpAndSettle()` for Flutter widget tests
   - Mock external dependencies

---

**Last Updated**: September 13, 2025
**Dart Version**: 3.9.2+
**Flutter Compatibility**: All current Flutter versions
**Mobile Focus**: Optimized for Android and iOS development patterns