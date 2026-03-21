# Dart Types and Collections System

> Comprehensive guide to Dart's type system, built-in types, and collections for mobile application development.

## Type System Overview

Dart features a **sound type system** with **static typing** and **null safety**. The type system ensures type safety at compile time while providing flexibility through type inference and dynamic types.

### Core Type Principles
- **Sound Null Safety**: Prevents null pointer exceptions at compile time
- **Type Inference**: Automatically determines types when not explicitly declared
- **Strong Typing**: Once a type is determined, it cannot change
- **Optional Typing**: Use `dynamic` when type flexibility is needed

---

## Built-in Types

### Numeric Types

```dart
// Integer types
int wholeNumber = 42;
int hexValue = 0xFF;
int binaryValue = 0b1010;

// Double precision floating point
double pi = 3.14159;
double scientificNotation = 1.42e5;

// Num (supertype of int and double)
num someNumber = 42;        // Can be int or double
someNumber = 3.14;          // Now it's a double

// Type checking and conversion
if (someNumber is int) {
  print('It\'s an integer');
}

int intValue = someNumber.toInt();
double doubleValue = someNumber.toDouble();
String stringValue = someNumber.toString();

// Parsing from strings
int parsed = int.parse('42');
double parsedDouble = double.parse('3.14');

// Safe parsing with tryParse
int? maybeParsed = int.tryParse('not a number'); // Returns null
```

### Boolean Type

```dart
bool isActive = true;
bool isComplete = false;

// Boolean operations
bool result = isActive && !isComplete;
bool result2 = isActive || isComplete;

// Null-safe boolean checking
bool? maybeTrue;
bool definitelyFalse = maybeTrue ?? false;

// Common boolean patterns in mobile development
bool get hasNetworkConnection => networkStatus == NetworkStatus.connected;
bool get isDataLoaded => data != null && !isLoading;
```

### String Type

```dart
// String literals
String singleQuotes = 'Hello, World!';
String doubleQuotes = "Hello, World!";
String multiline = '''
  This is a
  multi-line string
  with preserved formatting.
''';

// String interpolation
String name = 'Alice';
int age = 30;
String greeting = 'Hello, $name! You are $age years old.';
String calculation = 'Next year you will be ${age + 1}.';

// Raw strings (no escape sequences processed)
String rawString = r'This is a raw string with \n literal backslashes.';

// String operations
String text = 'Flutter Development';
print(text.length);                    // 18
print(text.toUpperCase());             // FLUTTER DEVELOPMENT
print(text.toLowerCase());             // flutter development
print(text.substring(0, 7));          // Flutter
print(text.split(' '));               // [Flutter, Development]
print(text.contains('Flutter'));       // true
print(text.startsWith('Flutter'));     // true
print(text.endsWith('Development'));   // true
print(text.replaceAll('Flutter', 'Dart')); // Dart Development

// String formatting for mobile development
String formatFileSize(int bytes) {
  if (bytes < 1024) return '${bytes}B';
  if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)}KB';
  return '${(bytes / (1024 * 1024)).toStringAsFixed(1)}MB';
}

// Regular expressions
final emailRegex = RegExp(r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$');
bool isValidEmail(String email) => emailRegex.hasMatch(email);
```

---

## Null Safety

### Nullable and Non-nullable Types

```dart
// Non-nullable types (default)
String name = 'John';           // Cannot be null
int age = 30;                   // Cannot be null

// Nullable types (with ? suffix)
String? nickname;               // Can be null
int? score;                     // Can be null

// Working with nullable types
void processUser(String? name) {
  if (name != null) {
    // name is promoted to non-nullable String here
    print(name.length);         // Safe to use
  }

  // Null-aware operators
  int length = name?.length ?? 0;        // Safe navigation + null coalescing
  name?.toUpperCase();                   // Only calls if not null

  // Assertion operator (use carefully!)
  String definitelyNotNull = name!;      // Throws if null
}

// Late variables
late String expensiveValue;

void initializeExpensiveValue() {
  expensiveValue = performExpensiveComputation();
}

// Late final for one-time initialization
late final String appVersion = getAppVersion();
```

### Null Safety Patterns

```dart
// Factory constructor with nullable return
class User {
  final String name;
  final String email;

  User._(this.name, this.email);

  // Factory that can return null
  static User? fromJson(Map<String, dynamic>? json) {
    if (json == null || json['name'] == null || json['email'] == null) {
      return null;
    }
    return User._(json['name'], json['email']);
  }

  // Safe getter with null check
  static String getDisplayName(User? user) {
    return user?.name ?? 'Anonymous';
  }
}

// Null-aware cascade operator
User? user;
user?..name = 'John'
    ..email = 'john@example.com';

// Extension methods for null safety
extension NullableStringExtension on String? {
  bool get isNullOrEmpty => this == null || this!.isEmpty;
  String get orEmpty => this ?? '';
}

// Usage
String? userInput;
if (userInput.isNullOrEmpty) {
  print('No input provided');
}
```

---

## Collections

### Lists

```dart
// List creation
List<String> fruits = ['apple', 'banana', 'orange'];
var numbers = <int>[1, 2, 3, 4, 5];
List<dynamic> mixed = ['text', 42, true];

// List operations
fruits.add('grape');                    // Add single item
fruits.addAll(['kiwi', 'mango']);      // Add multiple items
fruits.insert(1, 'cherry');           // Insert at index
fruits.remove('banana');              // Remove by value
fruits.removeAt(0);                   // Remove by index
fruits.clear();                       // Remove all items

// List properties
print(fruits.length);                 // Number of items
print(fruits.isEmpty);               // Check if empty
print(fruits.isNotEmpty);            // Check if not empty
print(fruits.first);                 // First item (throws if empty)
print(fruits.last);                  // Last item (throws if empty)

// Safe access
print(fruits.firstOrNull);           // First item or null
print(fruits.lastOrNull);            // Last item or null

// List iteration
for (String fruit in fruits) {
  print(fruit);
}

fruits.forEach((fruit) => print(fruit));

for (int i = 0; i < fruits.length; i++) {
  print('${i}: ${fruits[i]}');
}

// List transformations
List<String> upperCaseFruits = fruits.map((fruit) => fruit.toUpperCase()).toList();
List<String> longNames = fruits.where((fruit) => fruit.length > 5).toList();
String joined = fruits.join(', ');

// List sorting
fruits.sort();                        // Sort in place
List<String> sorted = [...fruits]..sort(); // Create sorted copy

// Custom sorting
List<User> users = [/* ... */];
users.sort((a, b) => a.name.compareTo(b.name));

// List generation
List<int> range = List.generate(10, (index) => index * 2);
List<String> filled = List.filled(5, 'default');
```

### Sets

```dart
// Set creation
Set<String> uniqueFruits = {'apple', 'banana', 'apple'}; // {'apple', 'banana'}
var numbers = <int>{1, 2, 3, 4, 5};
Set<String> emptySet = <String>{};

// Set operations
uniqueFruits.add('orange');
uniqueFruits.addAll(['kiwi', 'mango']);
uniqueFruits.remove('banana');

// Set properties
print(uniqueFruits.length);
print(uniqueFruits.contains('apple'));

// Set operations (mathematical)
Set<int> set1 = {1, 2, 3, 4};
Set<int> set2 = {3, 4, 5, 6};

Set<int> union = set1.union(set2);           // {1, 2, 3, 4, 5, 6}
Set<int> intersection = set1.intersection(set2); // {3, 4}
Set<int> difference = set1.difference(set2);     // {1, 2}

// Converting between collections
List<String> fruitList = uniqueFruits.toList();
Set<String> fruitSet = fruitList.toSet();
```

### Maps

```dart
// Map creation
Map<String, int> scores = {'Alice': 95, 'Bob': 87, 'Charlie': 92};
var ages = <String, int>{'John': 30, 'Jane': 25};
Map<String, dynamic> userdata = {
  'name': 'John Doe',
  'age': 30,
  'isActive': true,
};

// Map operations
scores['David'] = 88;                 // Add/update entry
scores.putIfAbsent('Eve', () => 90); // Add only if key doesn't exist
scores.remove('Bob');                 // Remove entry
scores.clear();                       // Remove all entries

// Map properties
print(scores.length);
print(scores.isEmpty);
print(scores.isNotEmpty);
print(scores.keys);                   // Iterable of keys
print(scores.values);                 // Iterable of values

// Map access
int? aliceScore = scores['Alice'];    // Nullable return
int davidScore = scores['David'] ?? 0; // With default value

// Map iteration
scores.forEach((name, score) {
  print('$name: $score');
});

for (MapEntry entry in scores.entries) {
  print('${entry.key}: ${entry.value}');
}

for (String name in scores.keys) {
  print('$name: ${scores[name]}');
}

// Map transformations
Map<String, String> grades = scores.map(
  (name, score) => MapEntry(name, _scoreToGrade(score)),
);

// Filtering maps
Map<String, int> highScores = Map.fromEntries(
  scores.entries.where((entry) => entry.value >= 90),
);
```

---

## Advanced Collection Operations

### Collection Processing Patterns

```dart
// Common mobile development patterns
class DataProcessor {
  // Filter and transform data
  List<UserViewModel> processUsers(List<User> users) {
    return users
        .where((user) => user.isActive)
        .map((user) => UserViewModel.fromUser(user))
        .toList();
  }

  // Group data
  Map<String, List<User>> groupUsersByDepartment(List<User> users) {
    final grouped = <String, List<User>>{};
    for (final user in users) {
      grouped.putIfAbsent(user.department, () => []).add(user);
    }
    return grouped;
  }

  // Aggregate data
  Statistics calculateStatistics(List<int> scores) {
    if (scores.isEmpty) return Statistics.empty();

    final total = scores.fold<int>(0, (sum, score) => sum + score);
    final average = total / scores.length;
    final max = scores.reduce((a, b) => a > b ? a : b);
    final min = scores.reduce((a, b) => a < b ? a : b);

    return Statistics(total, average, max, min);
  }

  // Find elements
  User? findUserById(List<User> users, String id) {
    try {
      return users.firstWhere((user) => user.id == id);
    } on StateError {
      return null;
    }
  }

  // Safe first/last access
  T? firstOrNull<T>(List<T> list) => list.isEmpty ? null : list.first;
  T? lastOrNull<T>(List<T> list) => list.isEmpty ? null : list.last;
}

// Pagination helper
class PaginatedList<T> {
  final List<T> items;
  final int totalCount;
  final int page;
  final int pageSize;

  PaginatedList(this.items, this.totalCount, this.page, this.pageSize);

  bool get hasNextPage => (page * pageSize) < totalCount;
  bool get hasPreviousPage => page > 1;
  int get totalPages => (totalCount / pageSize).ceil();

  List<T> getPage(int pageNumber) {
    final startIndex = (pageNumber - 1) * pageSize;
    final endIndex = (startIndex + pageSize).clamp(0, items.length);
    return items.sublist(startIndex, endIndex);
  }
}
```

### Immutable Collections

```dart
// Creating immutable collections
class ImmutableUserList {
  final List<User> _users;

  const ImmutableUserList(this._users);

  // Return new instance with added user
  ImmutableUserList add(User user) {
    return ImmutableUserList([..._users, user]);
  }

  // Return new instance with user removed
  ImmutableUserList remove(String userId) {
    final filtered = _users.where((user) => user.id != userId).toList();
    return ImmutableUserList(filtered);
  }

  // Return new instance with user updated
  ImmutableUserList update(String userId, User updatedUser) {
    final updated = _users.map((user) {
      return user.id == userId ? updatedUser : user;
    }).toList();
    return ImmutableUserList(updated);
  }

  // Access methods
  List<User> get users => List.unmodifiable(_users);
  int get length => _users.length;
  bool get isEmpty => _users.isEmpty;

  User? findById(String id) {
    try {
      return _users.firstWhere((user) => user.id == id);
    } on StateError {
      return null;
    }
  }
}
```

---

## Custom Types and Classes

### Class Definition

```dart
class User {
  // Instance variables
  final String id;
  final String name;
  final String email;
  final DateTime createdAt;
  final bool isActive;

  // Constructor
  User({
    required this.id,
    required this.name,
    required this.email,
    DateTime? createdAt,
    this.isActive = true,
  }) : createdAt = createdAt ?? DateTime.now();

  // Named constructors
  User.guest()
      : id = 'guest',
        name = 'Guest User',
        email = 'guest@example.com',
        createdAt = DateTime.now(),
        isActive = false;

  User.fromJson(Map<String, dynamic> json)
      : id = json['id'],
        name = json['name'],
        email = json['email'],
        createdAt = DateTime.parse(json['created_at']),
        isActive = json['is_active'] ?? true;

  // Factory constructor
  factory User.createWithValidation({
    required String id,
    required String name,
    required String email,
  }) {
    if (!_isValidEmail(email)) {
      throw ArgumentError('Invalid email format');
    }
    return User(id: id, name: name, email: email);
  }

  // Methods
  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'email': email,
      'created_at': createdAt.toIso8601String(),
      'is_active': isActive,
    };
  }

  User copyWith({
    String? name,
    String? email,
    bool? isActive,
  }) {
    return User(
      id: id, // id cannot be changed
      name: name ?? this.name,
      email: email ?? this.email,
      createdAt: createdAt, // createdAt cannot be changed
      isActive: isActive ?? this.isActive,
    );
  }

  // Getters
  String get displayName => name.isNotEmpty ? name : email.split('@').first;
  bool get isGuest => id == 'guest';

  // Static methods
  static bool _isValidEmail(String email) {
    return RegExp(r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$').hasMatch(email);
  }

  // Equality and hashing
  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
           other is User &&
           other.id == id &&
           other.name == name &&
           other.email == email;
  }

  @override
  int get hashCode => Object.hash(id, name, email);

  @override
  String toString() => 'User(id: $id, name: $name, email: $email)';
}
```

### Enums

```dart
// Basic enum
enum UserRole { admin, moderator, user, guest }

// Enhanced enum (Dart 2.17+)
enum NetworkStatus {
  disconnected('No Connection', 0),
  connecting('Connecting...', 1),
  connected('Connected', 2),
  reconnecting('Reconnecting...', 3);

  const NetworkStatus(this.displayName, this.priority);

  final String displayName;
  final int priority;

  bool get isConnected => this == NetworkStatus.connected;
  bool get canMakeRequests => this == NetworkStatus.connected;

  Color get statusColor {
    switch (this) {
      case NetworkStatus.connected:
        return Colors.green;
      case NetworkStatus.connecting:
      case NetworkStatus.reconnecting:
        return Colors.orange;
      case NetworkStatus.disconnected:
        return Colors.red;
    }
  }
}

// Usage
void displayNetworkStatus(NetworkStatus status) {
  print('Status: ${status.displayName}');
  print('Priority: ${status.priority}');
  print('Can make requests: ${status.canMakeRequests}');
}
```

### Records (Dart 3.0+)

```dart
// Record types for returning multiple values
(String, int) getUserInfo() {
  return ('John Doe', 30);
}

// Named records
(String name, int age, bool isActive) getUserDetails() {
  return (name: 'John Doe', age: 30, isActive: true);
}

// Using records
void processUserInfo() {
  final (name, age) = getUserInfo();
  print('User: $name, Age: $age');

  final userDetails = getUserDetails();
  print('Name: ${userDetails.name}');
  print('Age: ${userDetails.age}');
  print('Active: ${userDetails.isActive}');
}

// Records in collections
List<(String name, double score)> getHighScores() {
  return [
    ('Alice', 95.5),
    ('Bob', 87.2),
    ('Charlie', 92.8),
  ];
}

// Record destructuring in loops
void displayScores() {
  for (final (name, score) in getHighScores()) {
    print('$name: ${score.toStringAsFixed(1)}');
  }
}
```

---

## Type Checking and Casting

### Runtime Type Information

```dart
void processData(dynamic data) {
  // Type checking with 'is'
  if (data is String) {
    print('String length: ${data.length}');
  } else if (data is int) {
    print('Integer value: $data');
  } else if (data is List) {
    print('List with ${data.length} items');
  } else if (data is Map<String, dynamic>) {
    print('Map with keys: ${data.keys.join(', ')}');
  }

  // Negative type checking
  if (data is! String) {
    print('Data is not a string');
  }

  // Type casting with 'as'
  try {
    String stringData = data as String;
    print('Casted to string: $stringData');
  } on TypeError catch (e) {
    print('Cast failed: $e');
  }

  // Safe casting
  String? maybeString = data is String ? data : null;
  String? anotherWay = data as String?;
}

// Generic type checking
bool isListOfStrings(dynamic value) {
  return value is List<String>;
}

bool isMapWithStringKeys(dynamic value) {
  return value is Map<String, dynamic>;
}
```

---

## Collections for Mobile Development

### Common Mobile Patterns

```dart
// Caching with Map
class DataCache<K, V> {
  final Map<K, V> _cache = {};
  final int maxSize;

  DataCache({this.maxSize = 100});

  V? get(K key) => _cache[key];

  void put(K key, V value) {
    if (_cache.length >= maxSize) {
      // Remove oldest entry (simple LRU)
      final firstKey = _cache.keys.first;
      _cache.remove(firstKey);
    }
    _cache[key] = value;
  }

  void clear() => _cache.clear();
  bool containsKey(K key) => _cache.containsKey(key);
  int get size => _cache.length;
}

// Search functionality
class SearchHelper {
  static List<T> search<T>(
    List<T> items,
    String query,
    String Function(T) getSearchText,
  ) {
    if (query.isEmpty) return items;

    final normalizedQuery = query.toLowerCase();
    return items.where((item) {
      final searchText = getSearchText(item).toLowerCase();
      return searchText.contains(normalizedQuery);
    }).toList();
  }

  static List<User> searchUsers(List<User> users, String query) {
    return search(users, query, (user) => '${user.name} ${user.email}');
  }
}

// Data loading states
enum LoadingState { initial, loading, loaded, error }

class DataState<T> {
  final LoadingState state;
  final T? data;
  final String? error;

  const DataState({
    required this.state,
    this.data,
    this.error,
  });

  const DataState.initial() : this(state: LoadingState.initial);
  const DataState.loading() : this(state: LoadingState.loading);
  const DataState.loaded(T data) : this(state: LoadingState.loaded, data: data);
  const DataState.error(String error) : this(state: LoadingState.error, error: error);

  bool get isInitial => state == LoadingState.initial;
  bool get isLoading => state == LoadingState.loading;
  bool get isLoaded => state == LoadingState.loaded;
  bool get hasError => state == LoadingState.error;

  T get value {
    if (data == null) throw StateError('No data available');
    return data!;
  }
}
```

---

## Best Practices

### Type Safety Guidelines

1. **Prefer explicit types for public APIs**
```dart
// Good: Clear parameter and return types
Future<List<User>> fetchUsers(String departmentId) async { }

// Avoid: Generic dynamic types
Future<dynamic> fetchData(dynamic id) async { }
```

2. **Use null safety effectively**
```dart
// Good: Handle nullability explicitly
String? getUserDisplayName(User? user) {
  return user?.name ?? 'Unknown';
}

// Avoid: Unnecessary null assertions
String getBadDisplayName(User? user) {
  return user!.name; // Dangerous - can throw
}
```

3. **Choose appropriate collection types**
```dart
// Use List for ordered, indexed data
List<String> menuItems = ['Home', 'Profile', 'Settings'];

// Use Set for unique values
Set<String> uniqueTags = {'flutter', 'mobile', 'dart'};

// Use Map for key-value associations
Map<String, String> localizations = {'en': 'Hello', 'es': 'Hola'};
```

4. **Implement proper equality**
```dart
class Item {
  final String id;
  final String name;

  Item(this.id, this.name);

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
           other is Item && other.id == id;
  }

  @override
  int get hashCode => id.hashCode;
}
```

---

**Last Updated**: September 13, 2025
**Dart Version**: 3.9.2+
**Null Safety**: Sound null safety enabled
**Mobile Focus**: Optimized patterns for Android and iOS development