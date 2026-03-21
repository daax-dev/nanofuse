# Dart Memory Management

> Comprehensive guide to Dart's memory management, garbage collection, and performance optimization for mobile applications.

## Memory Management Overview

Dart uses **automatic memory management** with a **generational garbage collector**. The runtime handles memory allocation and deallocation automatically, but understanding how it works is crucial for building performant mobile applications.

### Key Concepts
- **Automatic Memory Management**: No manual memory allocation/deallocation
- **Generational Garbage Collection**: Optimized for different object lifespans
- **Reference Counting**: Objects are tracked by references
- **Weak References**: References that don't prevent garbage collection
- **Memory Pools**: Efficient allocation for small objects

---

## Dart Memory Model

### Object Allocation and Lifecycle

```dart
// Object creation and lifecycle
void memoryLifecycleExample() {
  // 1. Object allocation (heap memory)
  var user = User(name: 'John', age: 30);  // Allocated in young generation

  // 2. Object usage
  print(user.displayName);  // Object is alive and referenced

  // 3. Reference goes out of scope
  user = null;  // No longer referenced

  // 4. Eligible for garbage collection
  // GC will reclaim memory when needed
}

// Long-lived objects move to old generation
class ApplicationState {
  static final ApplicationState _instance = ApplicationState._internal();
  static ApplicationState get instance => _instance;

  ApplicationState._internal(); // Singleton - lives in old generation

  final Map<String, dynamic> _cache = {}; // Long-lived cache
}
```

### Memory Layout

```dart
// Stack vs Heap allocation
void memoryLayoutExample() {
  // Stack allocated (method parameters, local variables)
  int localVariable = 42;        // Stack
  String localString = 'hello';  // Reference on stack, object on heap

  // Heap allocated (objects, collections)
  List<String> items = ['a', 'b', 'c'];  // List and strings on heap
  User user = User(name: 'Alice');        // User object on heap

  // Value types vs Reference types
  int a = 5;
  int b = a;    // Copy of value
  b = 10;       // a is still 5

  List<int> list1 = [1, 2, 3];
  List<int> list2 = list1;  // Reference copy
  list2.add(4);             // Both list1 and list2 see the change
}
```

---

## Garbage Collection

### Generational Garbage Collection

```dart
// Understanding object generations
class MemoryGenerationExample {
  // Young generation objects (short-lived)
  void processTemporaryData() {
    for (int i = 0; i < 1000; i++) {
      var tempData = ProcessingData(i);  // Young generation
      tempData.process();                // Used briefly
      // tempData becomes eligible for GC at end of iteration
    }
  }

  // Old generation objects (long-lived)
  static final Map<String, String> _cache = {};  // Moves to old generation
  static final Logger _logger = Logger('App');   // Long-lived singleton

  void cacheData(String key, String value) {
    _cache[key] = value;  // Long-lived data
  }
}

// Triggering garbage collection manually (rare usage)
import 'dart:io' as io;

void forceGarbageCollection() {
  // Force GC - only for debugging/testing
  io.ProcessResult.run('', [], runInShell: true).then((_) {
    print('GC triggered');
  });
}
```

### GC Performance Characteristics

```dart
class GCPerformanceExample {
  // Short allocation bursts (GC friendly)
  List<ProcessedItem> processItems(List<RawItem> items) {
    final results = <ProcessedItem>[];

    for (final item in items) {
      // Short-lived temporary objects
      final validator = ItemValidator();
      final processor = ItemProcessor();

      if (validator.isValid(item)) {
        final processed = processor.process(item);
        results.add(processed);
      }
      // validator and processor become eligible for GC
    }

    return results;
  }

  // Avoiding memory pressure during processing
  Future<void> processLargeDataset(List<Data> dataset) async {
    const batchSize = 100;

    for (int i = 0; i < dataset.length; i += batchSize) {
      final batch = dataset.skip(i).take(batchSize);
      await processBatch(batch);

      // Allow GC between batches
      await Future.delayed(Duration.zero);
    }
  }
}
```

---

## Memory Leaks and Prevention

### Common Memory Leak Patterns

```dart
// ❌ Memory leak: Unclosed StreamSubscription
class BadEventHandler {
  StreamSubscription? _subscription;

  void startListening() {
    _subscription = eventStream.listen((event) {
      handleEvent(event);
    });
    // Missing: _subscription.cancel() somewhere
  }
}

// ✅ Proper resource cleanup
class GoodEventHandler {
  StreamSubscription? _subscription;

  void startListening() {
    _subscription = eventStream.listen((event) {
      handleEvent(event);
    });
  }

  void dispose() {
    _subscription?.cancel();
    _subscription = null;
  }
}

// ❌ Memory leak: Timer not canceled
class BadPeriodicTask {
  Timer? _timer;

  void startPeriodicTask() {
    _timer = Timer.periodic(Duration(seconds: 1), (timer) {
      performTask();
    });
    // Missing: timer.cancel()
  }
}

// ✅ Proper timer management
class GoodPeriodicTask {
  Timer? _timer;

  void startPeriodicTask() {
    _timer = Timer.periodic(Duration(seconds: 1), (timer) {
      performTask();
    });
  }

  void dispose() {
    _timer?.cancel();
    _timer = null;
  }
}
```

### Flutter-Specific Memory Management

```dart
// Flutter StatefulWidget proper cleanup
class UserProfileWidget extends StatefulWidget {
  @override
  _UserProfileWidgetState createState() => _UserProfileWidgetState();
}

class _UserProfileWidgetState extends State<UserProfileWidget> {
  StreamSubscription<User>? _userSubscription;
  Timer? _refreshTimer;
  AnimationController? _animationController;

  @override
  void initState() {
    super.initState();

    // Initialize resources
    _animationController = AnimationController(
      vsync: this,
      duration: Duration(milliseconds: 300),
    );

    _userSubscription = userService.userStream.listen((user) {
      if (mounted) setState(() => this.user = user);
    });

    _refreshTimer = Timer.periodic(Duration(minutes: 5), (_) {
      refreshData();
    });
  }

  @override
  void dispose() {
    // Critical: Clean up all resources
    _userSubscription?.cancel();
    _refreshTimer?.cancel();
    _animationController?.dispose();

    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(/* widget content */);
  }
}

// Global state management cleanup
class AppStateManager {
  static final AppStateManager _instance = AppStateManager._internal();
  static AppStateManager get instance => _instance;

  AppStateManager._internal();

  final List<StreamSubscription> _subscriptions = [];
  final List<Timer> _timers = [];

  void addSubscription(StreamSubscription subscription) {
    _subscriptions.add(subscription);
  }

  void addTimer(Timer timer) {
    _timers.add(timer);
  }

  void dispose() {
    // Clean up all managed resources
    for (final subscription in _subscriptions) {
      subscription.cancel();
    }
    _subscriptions.clear();

    for (final timer in _timers) {
      timer.cancel();
    }
    _timers.clear();
  }
}
```

---

## Weak References

### Using Weak References to Prevent Leaks

```dart
import 'dart:core' as core;

// WeakReference prevents strong references
class ObserverPattern {
  final Set<WeakReference<Observer>> _observers = {};

  void addObserver(Observer observer) {
    _observers.add(WeakReference(observer));
  }

  void removeObserver(Observer observer) {
    _observers.removeWhere((ref) => ref.target == observer);
  }

  void notifyObservers(Event event) {
    // Clean up dead references
    _observers.removeWhere((ref) => ref.target == null);

    // Notify active observers
    for (final ref in _observers) {
      ref.target?.onEvent(event);
    }
  }
}

// Cache with weak references
class WeakCache<K, V> {
  final Map<K, WeakReference<V>> _cache = {};

  V? get(K key) {
    final ref = _cache[key];
    final value = ref?.target;

    if (value == null) {
      _cache.remove(key);  // Clean up dead reference
    }

    return value;
  }

  void put(K key, V value) {
    _cache[key] = WeakReference(value);
  }

  void cleanUp() {
    _cache.removeWhere((key, ref) => ref.target == null);
  }
}
```

---

## Memory Optimization Techniques

### Object Pooling

```dart
// Object pool for frequently created/destroyed objects
class ObjectPool<T> {
  final List<T> _available = [];
  final T Function() _factory;
  final void Function(T)? _reset;
  final int _maxSize;

  ObjectPool({
    required T Function() factory,
    void Function(T)? reset,
    int maxSize = 50,
  }) : _factory = factory,
       _reset = reset,
       _maxSize = maxSize;

  T acquire() {
    if (_available.isEmpty) {
      return _factory();
    }
    return _available.removeLast();
  }

  void release(T object) {
    if (_available.length < _maxSize) {
      _reset?.call(object);
      _available.add(object);
    }
  }
}

// Usage example
class NetworkRequestPool {
  static final ObjectPool<HttpRequest> _pool = ObjectPool<HttpRequest>(
    factory: () => HttpRequest(),
    reset: (request) => request.reset(),
    maxSize: 10,
  );

  static Future<String> makeRequest(String url) async {
    final request = _pool.acquire();
    try {
      final response = await request.get(url);
      return response.body;
    } finally {
      _pool.release(request);
    }
  }
}
```

### Memory-Efficient Collections

```dart
class MemoryEfficientCollections {
  // Use typed collections for better memory usage
  void efficientCollections() {
    // ✅ Typed lists are more memory efficient
    final List<int> numbers = <int>[1, 2, 3, 4, 5];
    final Set<String> uniqueStrings = <String>{'a', 'b', 'c'};

    // ❌ Avoid dynamic collections when type is known
    final List<dynamic> mixedList = <dynamic>[1, 'a', true];

    // ✅ Use fixed-length lists when size is known
    final List<int> fixedNumbers = List<int>.filled(100, 0);

    // ✅ Use growable: false for better memory usage
    final List<String> names = <String>[]..length = 10;
  }

  // Efficient string building
  String buildLargeString(List<String> parts) {
    // ✅ Use StringBuffer for multiple concatenations
    final buffer = StringBuffer();
    for (final part in parts) {
      buffer.write(part);
    }
    return buffer.toString();

    // ❌ Avoid repeated string concatenation
    // String result = '';
    // for (final part in parts) {
    //   result += part;  // Creates new string each time
    // }
    // return result;
  }

  // Memory-conscious data processing
  Stream<ProcessedData> processLargeDataset(Stream<RawData> input) async* {
    await for (final chunk in input) {
      // Process in chunks to avoid loading entire dataset
      final processed = await processChunk(chunk);
      yield processed;

      // Allow garbage collection between chunks
      await Future.delayed(Duration.zero);
    }
  }
}
```

---

## Memory Profiling and Monitoring

### Memory Usage Monitoring

```dart
import 'dart:io';
import 'dart:developer' as developer;

class MemoryMonitor {
  static void logMemoryUsage() {
    final info = ProcessInfo.currentRss;
    print('Current RSS: ${info ~/ 1024 ~/ 1024} MB');

    developer.log('Memory usage: ${info ~/ 1024} KB', name: 'Memory');
  }

  static Future<void> trackMemoryDuringOperation(
    String operationName,
    Future<void> Function() operation,
  ) async {
    final startRss = ProcessInfo.currentRss;
    print('[$operationName] Start RSS: ${startRss ~/ 1024 ~/ 1024} MB');

    await operation();

    // Force GC to get accurate measurement
    await Future.delayed(Duration(milliseconds: 100));

    final endRss = ProcessInfo.currentRss;
    final delta = endRss - startRss;
    print('[$operationName] End RSS: ${endRss ~/ 1024 ~/ 1024} MB');
    print('[$operationName] Delta: ${delta ~/ 1024 ~/ 1024} MB');
  }
}

// Usage in tests or debugging
void main() async {
  await MemoryMonitor.trackMemoryDuringOperation(
    'Large Data Processing',
    () => processLargeDataset(),
  );
}
```

### Memory Leak Detection

```dart
class MemoryLeakDetector {
  static final Map<Type, int> _objectCounts = {};
  static Timer? _monitoringTimer;

  static void registerObject(Object obj) {
    final type = obj.runtimeType;
    _objectCounts[type] = (_objectCounts[type] ?? 0) + 1;
  }

  static void unregisterObject(Object obj) {
    final type = obj.runtimeType;
    final count = _objectCounts[type] ?? 0;
    if (count > 0) {
      _objectCounts[type] = count - 1;
    }
  }

  static void startMonitoring() {
    _monitoringTimer = Timer.periodic(Duration(seconds: 30), (_) {
      _logObjectCounts();
    });
  }

  static void stopMonitoring() {
    _monitoringTimer?.cancel();
    _monitoringTimer = null;
  }

  static void _logObjectCounts() {
    print('=== Object Count Report ===');
    _objectCounts.forEach((type, count) {
      if (count > 0) {
        print('$type: $count instances');
      }
    });
    print('==========================');
  }
}

// Usage in classes to track instance counts
class TrackedUser {
  TrackedUser() {
    MemoryLeakDetector.registerObject(this);
  }

  void dispose() {
    MemoryLeakDetector.unregisterObject(this);
  }
}
```

---

## Mobile-Specific Memory Considerations

### Flutter Memory Management

```dart
// Image memory management in Flutter
class ImageMemoryManager {
  static const int _maxCacheSize = 50 * 1024 * 1024; // 50 MB

  static Widget efficientNetworkImage(String url) {
    return CachedNetworkImage(
      imageUrl: url,
      memCacheWidth: 800,  // Limit memory usage
      memCacheHeight: 600,
      placeholder: (context, url) => CircularProgressIndicator(),
      errorWidget: (context, url, error) => Icon(Icons.error),

      // Configure caching
      cacheManager: CacheManager(
        Config(
          'customCacheKey',
          stalePeriod: Duration(days: 7),
          maxNrOfCacheObjects: 100,
        ),
      ),
    );
  }

  static void clearImageCache() {
    // Clear Flutter's image cache
    imageCache.clear();
    imageCache.clearLiveImages();
  }

  static void configureImageCache() {
    // Configure image cache size
    imageCache.maximumSize = 200;  // Number of images
    imageCache.maximumSizeBytes = _maxCacheSize;
  }
}

// List view memory optimization
class EfficientListView extends StatelessWidget {
  final List<Item> items;

  const EfficientListView({super.key, required this.items});

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      itemCount: items.length,
      // Only build visible items
      itemBuilder: (context, index) {
        if (index >= items.length) return null;
        return ItemWidget(item: items[index]);
      },

      // Cache extent for smooth scrolling
      cacheExtent: 1000.0,

      // Optimize for large lists
      addRepaintBoundaries: true,
      addSemanticIndexes: true,
    );
  }
}
```

### Background Memory Management

```dart
class BackgroundMemoryManager {
  static void optimizeForBackground() {
    // Clear non-essential caches when app goes to background
    imageCache.clearLiveImages();

    // Cancel unnecessary timers
    _cancelNonCriticalTimers();

    // Reduce memory footprint
    _clearTemporaryCaches();
  }

  static void restoreFromBackground() {
    // Restore essential functionality when app returns to foreground
    _restoreEssentialServices();
  }

  static void _cancelNonCriticalTimers() {
    // Cancel UI update timers, animation controllers, etc.
  }

  static void _clearTemporaryCaches() {
    // Clear UI-related caches that can be rebuilt
  }

  static void _restoreEssentialServices() {
    // Restart critical services
  }
}

// App lifecycle integration
class AppLifecycleManager extends WidgetsBindingObserver {
  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    switch (state) {
      case AppLifecycleState.paused:
        BackgroundMemoryManager.optimizeForBackground();
        break;
      case AppLifecycleState.resumed:
        BackgroundMemoryManager.restoreFromBackground();
        break;
      case AppLifecycleState.inactive:
      case AppLifecycleState.detached:
        // Handle other states
        break;
      case AppLifecycleState.hidden:
        // Handle hidden state
        break;
    }
  }
}
```

---

## Memory Testing

### Unit Testing Memory Management

```dart
import 'package:test/test.dart';

void main() {
  group('Memory Management Tests', () {
    test('should dispose resources properly', () {
      final manager = ResourceManager();
      manager.initialize();

      expect(manager.isInitialized, isTrue);

      manager.dispose();

      expect(manager.isDisposed, isTrue);
      expect(manager.hasActiveResources, isFalse);
    });

    test('should not leak memory in loop processing', () async {
      final initialRss = ProcessInfo.currentRss;

      // Process data in loop
      for (int i = 0; i < 1000; i++) {
        final processor = DataProcessor();
        await processor.process(generateTestData());
        processor.dispose();
      }

      // Allow GC to run
      await Future.delayed(Duration(milliseconds: 100));

      final finalRss = ProcessInfo.currentRss;
      final memoryIncrease = finalRss - initialRss;

      // Memory increase should be reasonable (< 10MB for this test)
      expect(memoryIncrease, lessThan(10 * 1024 * 1024));
    });
  });
}
```

### Integration Testing with Memory Constraints

```dart
import 'package:integration_test/integration_test.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('Memory Integration Tests', () {
    testWidgets('app should handle low memory conditions', (tester) async {
      await tester.pumpWidget(MyApp());

      // Simulate low memory condition
      await tester.binding.defaultBinaryMessenger.handlePlatformMessage(
        'flutter/system',
        StandardMethodCodec().encodeMethodCall(
          MethodCall('System.requestAppExitAndRestartIfNeeded'),
        ),
        (data) {},
      );

      await tester.pumpAndSettle();

      // App should still be functional
      expect(find.byType(MyApp), findsOneWidget);
    });
  });
}
```

---

## Best Practices Summary

### Memory Management Guidelines

1. **Resource Cleanup**
   - Always cancel StreamSubscriptions
   - Dispose AnimationControllers
   - Cancel Timers and periodic operations
   - Close database connections

2. **Avoid Memory Leaks**
   - Use WeakReferences for observer patterns
   - Implement dispose methods properly
   - Clean up global state
   - Monitor long-lived objects

3. **Optimize Collections**
   - Use typed collections over dynamic
   - Consider fixed-length lists when appropriate
   - Use object pools for frequently created objects
   - Process large datasets in chunks

4. **Flutter-Specific**
   - Configure image cache appropriately
   - Use ListView.builder for large lists
   - Clear caches during background transitions
   - Monitor widget tree complexity

5. **Testing and Monitoring**
   - Test memory usage in critical paths
   - Monitor memory in production
   - Profile memory during development
   - Use debugging tools for leak detection

---

**Last Updated**: September 13, 2025
**Dart Version**: 3.9.2+
**Flutter Compatibility**: All current Flutter versions
**Mobile Focus**: Optimized for Android and iOS memory constraints