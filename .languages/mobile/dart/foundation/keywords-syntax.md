# Dart Keywords and Syntax Reference

> Complete reference guide to all Dart language keywords, operators, and syntax patterns for mobile development.

## Reserved Keywords

Dart has **64 reserved keywords** (as of Dart 3.9). These cannot be used as identifiers.

### Control Flow Keywords

```dart
// Conditional statements
if (condition) { } else { }
if (user.isActive) print('Active') else print('Inactive');

// Switch statements (enhanced in Dart 3.0+)
switch (value) {
  case 1: return 'one';
  case 2: return 'two';
  default: return 'other';
}

// Pattern matching switch (Dart 3.0+)
String describe(Object obj) => switch (obj) {
  1 => 'one',
  String s when s.length > 5 => 'long string',
  String s => 'short string: $s',
  _ => 'unknown'
};

// Loops
for (int i = 0; i < items.length; i++) { }
for (final item in items) { }
while (condition) { }
do { } while (condition);

// Loop control
break;      // Exit loop
continue;   // Skip to next iteration
```

### Exception Handling Keywords

```dart
try {
  riskyOperation();
} catch (e) {
  handleError(e);
} on SpecificException catch (e, stackTrace) {
  handleSpecificError(e, stackTrace);
} finally {
  cleanup();
}

// Throwing exceptions
throw Exception('Something went wrong');
throw 'Custom error message';
rethrow; // Re-throw caught exception
```

### Class and Object Keywords

```dart
// Class definition
class User extends Person implements Serializable {
  // Class body
}

abstract class Shape {
  void draw(); // Abstract method
}

// Interface (implicit)
class Drawable {
  void draw() { } // All classes are implicit interfaces
}

// Mixins
mixin Flyable {
  void fly() { }
}

class Bird with Flyable { }

// Inheritance and interfaces
extends    // Class inheritance
implements // Interface implementation
with       // Mixin application

// Object creation
new User(); // Optional 'new' keyword
User();     // Preferred syntax

// Type checking and casting
is         // Type check
as         // Type cast
```

### Variable and Function Keywords

```dart
// Variable declarations
var name = 'John';           // Type inferred
String name = 'John';        // Explicit type
final name = 'John';         // Immutable (runtime constant)
const name = 'John';         // Compile-time constant
late String name;            // Late initialization
static int count = 0;        // Class-level variable

// Function keywords
void function() { }          // No return value
int calculate() { }          // Returns int
dynamic getValue() { }       // Dynamic return type
```

### Access Modifiers and Visibility

```dart
// No explicit access modifiers in Dart
// Underscore prefix indicates private to library
String _privateField;
void _privateMethod() { }

// External (for native/JS interop)
external void nativeFunction();
```

### Async Programming Keywords

```dart
// Asynchronous programming
async    // Marks function as asynchronous
await    // Waits for Future completion
yield    // Produces value in generator
sync*    // Synchronous generator
async*   // Asynchronous generator

// Examples
Future<String> fetchData() async {
  final response = await httpClient.get(url);
  return response.body;
}

Stream<int> countStream() async* {
  for (int i = 0; i < 10; i++) {
    await Future.delayed(Duration(seconds: 1));
    yield i;
  }
}

Iterable<int> generateNumbers() sync* {
  for (int i = 0; i < 10; i++) {
    yield i;
  }
}
```

## Complete Keyword List

### Core Language Keywords (33)
```dart
abstract    as          assert      async       await
break       case        catch       class       const
continue    covariant   default     deferred    do
else        enum        export      extends     extension
external    factory     false       final       finally
for         Function    get         hide        if
implements  import      in          interface   is
late        library     mixin       new         null
on          operator    part        required    rethrow
return      sealed      set         show        static
super       switch      sync        this        throw
true        try         typedef     var         void
when        while       with        yield
```

### Contextual Keywords (16)
```dart
// These have special meaning in specific contexts
async*      augment     base        covariant   deferred
export      extension   hide        implements  interface
library     mixin       on          part        sealed
show        sync*       when
```

### Built-in Identifiers (15)
```dart
// These have special meaning but can be used as identifiers
abstract    as          augment     base        covariant
deferred    dynamic     export      extension   external
factory     Function    get         implements  import
interface   late        library     mixin       operator
part        required    sealed      set         static
typedef
```

---

## Operators

### Arithmetic Operators

```dart
// Basic arithmetic
int a = 5, b = 3;
print(a + b);  // Addition: 8
print(a - b);  // Subtraction: 2
print(a * b);  // Multiplication: 15
print(a / b);  // Division: 1.6666...
print(a ~/ b); // Integer division: 1
print(a % b);  // Modulo: 2

// Unary operators
print(-a);     // Negation: -5
print(++a);    // Pre-increment: 6
print(a++);    // Post-increment: 6 (then becomes 7)
print(--a);    // Pre-decrement: 6
print(a--);    // Post-decrement: 6 (then becomes 5)
```

### Equality and Relational Operators

```dart
// Equality
print(a == b);  // Equal to: false
print(a != b);  // Not equal to: true

// Relational
print(a > b);   // Greater than: true
print(a >= b);  // Greater than or equal: true
print(a < b);   // Less than: false
print(a <= b);  // Less than or equal: false
```

### Type Test Operators

```dart
Object obj = 'Hello';

// Type checking
if (obj is String) {
  print('obj is a String');
}

if (obj is! int) {
  print('obj is not an int');
}

// Type casting
String str = obj as String;
```

### Assignment Operators

```dart
int a = 5;

// Compound assignment
a += 3;  // a = a + 3
a -= 2;  // a = a - 2
a *= 4;  // a = a * 4
a ~/= 2; // a = a ~/ 2 (integer division)
a %= 3;  // a = a % 3

// Null-aware assignment
String? name;
name ??= 'Default'; // Assign only if name is null
```

### Logical Operators

```dart
bool a = true, b = false;

print(!a);        // NOT: false
print(a && b);    // AND: false
print(a || b);    // OR: true
```

### Bitwise and Shift Operators

```dart
int a = 5;  // 101 in binary
int b = 3;  // 011 in binary

print(a & b);   // AND: 1 (001)
print(a | b);   // OR: 7 (111)
print(a ^ b);   // XOR: 6 (110)
print(~a);      // NOT: -6 (inverts all bits)
print(a << 1);  // Left shift: 10 (1010)
print(a >> 1);  // Right shift: 2 (010)
```

### Null-Aware Operators

```dart
String? name;
int? length = name?.length;  // Safe navigation
String display = name ?? 'Unknown';  // Null coalescing

// Null-aware method invocation
name?.toUpperCase();

// Null-aware cascade
User? user;
user?..name = 'John'
    ..email = 'john@example.com';
```

### Cascade Operator

```dart
// Traditional approach
Paint paint = Paint();
paint.color = Colors.red;
paint.strokeWidth = 5.0;
paint.style = PaintingStyle.stroke;

// Using cascade operator
Paint paint = Paint()
  ..color = Colors.red
  ..strokeWidth = 5.0
  ..style = PaintingStyle.stroke;

// Null-aware cascade
Paint? paint;
paint?..color = Colors.red
      ..strokeWidth = 5.0;
```

### Conditional (Ternary) Operator

```dart
String status = user.isActive ? 'Online' : 'Offline';

// Nested ternary (use sparingly)
String priority = score > 90 ? 'High' :
                 score > 60 ? 'Medium' : 'Low';
```

---

## Syntax Patterns

### Variable Declarations

```dart
// Type inference
var name = 'John';           // String inferred
var age = 25;               // int inferred
var isActive = true;        // bool inferred

// Explicit types
String name = 'John';
int age = 25;
bool isActive = true;

// Final (runtime constant)
final String name = 'John';
final timestamp = DateTime.now(); // Set once at runtime

// Const (compile-time constant)
const String name = 'John';
const pi = 3.14159;

// Late initialization
late String name;
late final String computedValue;

void initializeValues() {
  name = 'John';
  computedValue = expensiveComputation();
}

// Nullable types
String? name;           // Can be null
int? age;              // Can be null
late String? value;    // Late and nullable
```

### Function Definitions

```dart
// Basic function
void greet(String name) {
  print('Hello, $name!');
}

// Function with return type
int add(int a, int b) {
  return a + b;
}

// Arrow function (expression body)
int multiply(int a, int b) => a * b;

// Optional parameters
void greet(String name, [String greeting = 'Hello']) {
  print('$greeting, $name!');
}

// Named parameters
void createUser({
  required String name,
  int age = 0,
  String? email,
}) {
  // Function body
}

// Function as first-class objects
Function operation = add;
int result = operation(5, 3);

// Higher-order functions
void executeOperation(int a, int b, int Function(int, int) op) {
  print(op(a, b));
}

executeOperation(5, 3, add);
executeOperation(5, 3, (a, b) => a * b); // Anonymous function
```

### Class Definitions

```dart
// Basic class
class Person {
  String name;
  int age;

  // Constructor
  Person(this.name, this.age);

  // Named constructor
  Person.baby(this.name) : age = 0;

  // Factory constructor
  factory Person.fromJson(Map<String, dynamic> json) {
    return Person(json['name'], json['age']);
  }

  // Method
  void introduce() {
    print('Hi, I\'m $name, $age years old.');
  }

  // Getter
  String get description => '$name ($age)';

  // Setter
  set name(String newName) {
    if (newName.isNotEmpty) {
      _name = newName;
    }
  }
}

// Inheritance
class Student extends Person {
  String school;

  Student(super.name, super.age, this.school);

  @override
  void introduce() {
    super.introduce();
    print('I study at $school.');
  }
}

// Abstract class
abstract class Animal {
  void makeSound(); // Abstract method

  void sleep() {    // Concrete method
    print('Sleeping...');
  }
}

// Implementation
class Dog extends Animal {
  @override
  void makeSound() {
    print('Woof!');
  }
}
```

### Enums

```dart
// Basic enum
enum Color { red, green, blue }

// Enhanced enum (Dart 2.17+)
enum Planet {
  mercury(3.303e+23, 2.4397e6),
  venus(4.869e+24, 6.0518e6),
  earth(5.976e+24, 6.37814e6);

  const Planet(this.mass, this.radius);
  final double mass;       // in kilograms
  final double radius;     // in meters

  double get surfaceGravity => 6.67300E-11 * mass / (radius * radius);
}

// Using enums
Color favoriteColor = Color.blue;
print('Surface gravity of Earth: ${Planet.earth.surfaceGravity}');
```

### Pattern Matching (Dart 3.0+)

```dart
// Switch expressions
String describe(Object obj) => switch (obj) {
  1 => 'one',
  2 => 'two',
  String s when s.length > 5 => 'long string',
  String s => 'string: $s',
  List<int> l when l.isEmpty => 'empty list',
  List<int> l => 'list with ${l.length} items',
  _ => 'something else'
};

// Destructuring
var (x, y) = getCoordinates(); // Returns a record

// Pattern matching in if-case
if (obj case String s when s.length > 10) {
  print('Long string: $s');
}

// Record patterns
var person = (name: 'John', age: 30);
var (name: n, age: a) = person;
```

### Collections

```dart
// Lists
List<int> numbers = [1, 2, 3, 4, 5];
var names = <String>['John', 'Jane', 'Bob'];

// Sets
Set<String> fruits = {'apple', 'banana', 'orange'};
var uniqueNumbers = <int>{1, 2, 3, 2, 1}; // {1, 2, 3}

// Maps
Map<String, int> scores = {'Alice': 95, 'Bob': 87};
var ages = <String, int>{'John': 30, 'Jane': 25};

// Collection operations
var doubled = numbers.map((n) => n * 2).toList();
var evens = numbers.where((n) => n.isEven).toList();
var sum = numbers.fold(0, (sum, n) => sum + n);
```

### String Syntax

```dart
// Basic strings
String single = 'Hello, World!';
String double = "Hello, World!";

// String interpolation
String name = 'John';
int age = 30;
String message = 'Hello, $name! You are $age years old.';
String calculation = 'Next year you will be ${age + 1}.';

// Multi-line strings
String multiline = '''
This is a
multi-line
string.
''';

// Raw strings
String raw = r'This is a raw string with \n escape sequences';

// String concatenation
String full = 'Hello, ' + 'World!';
String interpolated = 'Hello, $name!';
```

### Comments

```dart
// Single-line comment

/*
Multi-line
comment
*/

/// Documentation comment for classes, functions, etc.
/// This will appear in generated documentation.
class User {
  /// The user's name
  String name;

  /// Creates a new user with the given [name]
  User(this.name);
}
```

---

## Special Syntax Features

### Extension Methods

```dart
// Extending existing types
extension StringExtension on String {
  bool get isEmail => contains('@');

  String capitalize() {
    if (isEmpty) return this;
    return '${this[0].toUpperCase()}${substring(1)}';
  }
}

// Usage
String email = 'user@example.com';
print(email.isEmail); // true
print('hello'.capitalize()); // Hello
```

### Generics

```dart
// Generic classes
class Box<T> {
  T value;
  Box(this.value);
}

// Generic functions
T getFirst<T>(List<T> items) {
  return items.first;
}

// Bounded generics
class NumberBox<T extends num> {
  T value;
  NumberBox(this.value);

  T add(T other) => value + other as T;
}

// Usage
var stringBox = Box<String>('Hello');
var intBox = Box<int>(42);
```

### Mixins

```dart
// Mixin definition
mixin Flyable {
  void fly() => print('Flying!');
}

mixin Swimmable {
  void swim() => print('Swimming!');
}

// Using mixins
class Bird with Flyable {
  void chirp() => print('Chirp!');
}

class Duck extends Bird with Swimmable {
  @override
  void fly() {
    print('Duck flying!');
  }
}

// Usage
var duck = Duck();
duck.fly();    // Duck flying!
duck.swim();   // Swimming!
duck.chirp();  // Chirp!
```

---

## Modern Dart Features (3.0+)

### Records

```dart
// Record types
(int, String) userInfo = (25, 'John');
(int age, String name) namedRecord = (age: 25, name: 'John');

// Accessing record fields
print(userInfo.$1); // 25
print(userInfo.$2); // John
print(namedRecord.age);  // 25
print(namedRecord.name); // John

// Functions returning records
(String, int) getUserInfo() {
  return ('John Doe', 30);
}

var (name, age) = getUserInfo(); // Destructuring
```

### Sealed Classes

```dart
// Sealed class (exhaustive subclassing)
sealed class Shape {}

class Circle extends Shape {
  final double radius;
  Circle(this.radius);
}

class Rectangle extends Shape {
  final double width, height;
  Rectangle(this.width, this.height);
}

// Switch must be exhaustive
double calculateArea(Shape shape) {
  return switch (shape) {
    Circle(radius: var r) => 3.14 * r * r,
    Rectangle(width: var w, height: var h) => w * h,
    // No default case needed - compiler knows all cases are covered
  };
}
```

### Class Modifiers

```dart
// Base class (can only be extended, not implemented)
base class Vehicle {
  void move() => print('Moving');
}

class Car extends Vehicle {} // OK
// class MockVehicle implements Vehicle {} // Error

// Interface class (can only be implemented, not extended)
interface class Drawable {
  void draw();
}

class Shape implements Drawable { // OK
  @override
  void draw() {}
}
// class ExtendedDrawable extends Drawable {} // Error

// Final class (cannot be extended or implemented)
final class ApiClient {
  void makeRequest() {}
}
// class MyApiClient extends ApiClient {} // Error
```

---

## Syntax Summary for Mobile Development

### Common Mobile Development Patterns

```dart
// Widget classes (Flutter-specific)
class MyWidget extends StatelessWidget {
  final String title;

  const MyWidget({super.key, required this.title});

  @override
  Widget build(BuildContext context) {
    return Text(title);
  }
}

// Stateful widgets
class CounterWidget extends StatefulWidget {
  @override
  State<CounterWidget> createState() => _CounterWidgetState();
}

class _CounterWidgetState extends State<CounterWidget> {
  int _counter = 0;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text('Counter: $_counter'),
        ElevatedButton(
          onPressed: () => setState(() => _counter++),
          child: Text('Increment'),
        ),
      ],
    );
  }
}

// Async operations for mobile
Future<List<User>> fetchUsers() async {
  try {
    final response = await http.get(Uri.parse('/api/users'));
    if (response.statusCode == 200) {
      final jsonData = json.decode(response.body) as List;
      return jsonData.map((item) => User.fromJson(item)).toList();
    }
    throw HttpException('Failed to load users');
  } catch (e) {
    rethrow;
  }
}

// Stream handling for real-time updates
Stream<ChatMessage> messageStream() async* {
  await for (final data in websocketStream) {
    yield ChatMessage.fromJson(data);
  }
}
```

---

**Last Updated**: September 13, 2025
**Dart Version**: 3.9.2+
**Completeness**: All current Dart keywords and major syntax patterns included