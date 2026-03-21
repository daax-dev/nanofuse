# Java Keywords and Syntax - Complete Reference

## Java Reserved Keywords (50 Total)

### Access Modifiers
```java
public    // Accessible from any other class
private   // Accessible only within the same class
protected // Accessible within the same package and subclasses
```

### Class, Method, and Variable Modifiers
```java
abstract  // Cannot be instantiated (classes) or must be overridden (methods)
static    // Belongs to the class rather than any instance
final     // Cannot be changed (variables), overridden (methods), or extended (classes)
native    // Method implemented in another language (typically C/C++)
synchronized // Thread-safe method or block access
transient // Field should not be serialized
volatile  // Field may be modified by multiple threads
strictfp  // Use strict floating-point calculations
```

### Class Definition Keywords
```java
class     // Define a class
interface // Define an interface
enum      // Define an enumeration (Java 5+)
extends   // Inherit from a superclass
implements // Implement an interface
```

### Method Definition and Return
```java
void      // Method returns no value
return    // Return a value from a method or exit a method
```

### Object Creation and Type Testing
```java
new       // Create a new object instance
instanceof // Test if an object is an instance of a specific type
super     // Reference to parent class
this      // Reference to current object
```

### Control Flow - Conditional Statements
```java
if        // Conditional statement
else      // Alternative branch for if statement
switch    // Multi-way branch statement
case      // Label in switch statement
default   // Default case in switch statement
```

### Control Flow - Loops
```java
for       // Loop with initialization, condition, and update
while     // Loop while condition is true
do        // Do-while loop (executes at least once)
break     // Exit from loop or switch
continue  // Skip to next iteration of loop
```

### Exception Handling
```java
try       // Begin exception handling block
catch     // Handle specific exception types
finally   // Execute code regardless of exceptions
throw     // Throw an exception
throws    // Declare exceptions that method may throw
```

### Package and Import
```java
package   // Declare package membership
import    // Import classes or packages
```

### Primitive Data Types
```java
boolean   // True or false value
byte      // 8-bit signed integer (-128 to 127)
short     // 16-bit signed integer (-32,768 to 32,767)
int       // 32-bit signed integer (-2^31 to 2^31-1)
long      // 64-bit signed integer (-2^63 to 2^63-1)
float     // 32-bit floating-point number
double    // 64-bit floating-point number
char      // 16-bit Unicode character
```

### Literal Values
```java
null      // Null reference (not pointing to any object)
true      // Boolean true literal
false     // Boolean false literal
```

### Reserved but Not Used
```java
const     // Reserved but not used (use 'final' instead)
goto      // Reserved but not used (considered harmful)
```

### Additional Reserved Keywords (Context-Sensitive)
```java
var       // Local variable type inference (Java 10+)
yield     // Used in switch expressions (Java 14+)
record    // Define record classes (Java 14+, standardized in Java 16)
sealed    // Restrict which classes can extend/implement (Java 17+)
permits   // Used with sealed classes to specify allowed subclasses
```

## Java Syntax Fundamentals

### Basic Program Structure
```java
// Package declaration (optional, must be first non-comment line)
package com.example.myapp;

// Import statements (optional, after package declaration)
import java.util.List;
import java.util.ArrayList;
import static java.lang.Math.PI; // Static import

// Class declaration
public class MyClass {
    // Class-level variables (fields)
    private int instanceVariable;
    private static final String CLASS_CONSTANT = "CONSTANT";

    // Constructor
    public MyClass() {
        this.instanceVariable = 0;
    }

    // Methods
    public void instanceMethod() {
        // Method body
    }

    public static void main(String[] args) {
        // Main method - program entry point
    }
}
```

### Variable Declaration and Initialization
```java
// Primitive type declarations
int number = 42;
double price = 99.99;
boolean isActive = true;
char initial = 'A';

// Object type declarations
String name = "John Doe";
List<String> names = new ArrayList<>();
Object obj = null;

// Array declarations
int[] numbers = new int[10];
String[] words = {"hello", "world"};
int[][] matrix = new int[3][3];

// Variable declaration with var (Java 10+)
var message = "Hello, World!"; // Inferred as String
var list = new ArrayList<String>(); // Inferred as ArrayList<String>

// Final variables (constants)
final int MAX_SIZE = 100;
final List<String> IMMUTABLE_LIST = List.of("a", "b", "c");
```

### Operators and Expressions

#### Arithmetic Operators
```java
int a = 10, b = 3;

int sum = a + b;        // Addition: 13
int difference = a - b; // Subtraction: 7
int product = a * b;    // Multiplication: 30
int quotient = a / b;   // Division: 3
int remainder = a % b;  // Modulo: 1

// Compound assignment operators
a += 5;  // a = a + 5
a -= 2;  // a = a - 2
a *= 2;  // a = a * 2
a /= 3;  // a = a / 3
a %= 4;  // a = a % 4

// Increment and decrement
int x = 5;
int y = ++x; // Pre-increment: x becomes 6, y is 6
int z = x++; // Post-increment: z is 6, x becomes 7
```

#### Comparison Operators
```java
int a = 10, b = 20;

boolean equal = a == b;        // false
boolean notEqual = a != b;     // true
boolean less = a < b;          // true
boolean lessEqual = a <= b;    // true
boolean greater = a > b;       // false
boolean greaterEqual = a >= b; // false

// Object comparison
String str1 = "hello";
String str2 = new String("hello");
boolean sameReference = str1 == str2;    // false
boolean sameContent = str1.equals(str2); // true
```

#### Logical Operators
```java
boolean a = true, b = false;

boolean and = a && b;    // Logical AND: false
boolean or = a || b;     // Logical OR: true
boolean not = !a;        // Logical NOT: false

// Short-circuit evaluation
boolean result = (a != null) && a.equals("test"); // Safe null check
```

#### Bitwise Operators
```java
int a = 5;  // Binary: 101
int b = 3;  // Binary: 011

int bitwiseAnd = a & b;    // 001 = 1
int bitwiseOr = a | b;     // 111 = 7
int bitwiseXor = a ^ b;    // 110 = 6
int bitwiseNot = ~a;       // ...11111010 (two's complement)
int leftShift = a << 2;    // 10100 = 20
int rightShift = a >> 1;   // 10 = 2
int unsignedRightShift = a >>> 1; // 10 = 2
```

#### Ternary Operator
```java
// Conditional operator: condition ? valueIfTrue : valueIfFalse
int age = 18;
String status = age >= 18 ? "adult" : "minor";

// Nested ternary (use sparingly)
String grade = score >= 90 ? "A" : score >= 80 ? "B" : score >= 70 ? "C" : "F";
```

### Control Flow Structures

#### If-Else Statements
```java
// Simple if statement
if (condition) {
    // Execute if condition is true
}

// If-else statement
if (condition) {
    // Execute if condition is true
} else {
    // Execute if condition is false
}

// If-else if chain
if (condition1) {
    // Execute if condition1 is true
} else if (condition2) {
    // Execute if condition2 is true
} else {
    // Execute if all conditions are false
}

// Single statement (braces optional but recommended)
if (x > 0)
    System.out.println("Positive");
```

#### Switch Statements
```java
// Traditional switch statement
int day = 3;
switch (day) {
    case 1:
        System.out.println("Monday");
        break;
    case 2:
        System.out.println("Tuesday");
        break;
    case 3:
        System.out.println("Wednesday");
        break;
    default:
        System.out.println("Other day");
        break;
}

// Switch with multiple cases
switch (month) {
    case 12: case 1: case 2:
        System.out.println("Winter");
        break;
    case 3: case 4: case 5:
        System.out.println("Spring");
        break;
    case 6: case 7: case 8:
        System.out.println("Summer");
        break;
    case 9: case 10: case 11:
        System.out.println("Fall");
        break;
}

// Switch expressions (Java 14+)
String season = switch (month) {
    case 12, 1, 2 -> "Winter";
    case 3, 4, 5 -> "Spring";
    case 6, 7, 8 -> "Summer";
    case 9, 10, 11 -> "Fall";
    default -> "Unknown";
};

// Switch with yield (Java 14+)
String result = switch (value) {
    case 1 -> "one";
    case 2 -> "two";
    case 3 -> {
        System.out.println("Processing three");
        yield "three";
    }
    default -> "other";
};
```

#### Loop Structures
```java
// For loop
for (int i = 0; i < 10; i++) {
    System.out.println(i);
}

// Enhanced for loop (for-each)
int[] numbers = {1, 2, 3, 4, 5};
for (int number : numbers) {
    System.out.println(number);
}

// While loop
int i = 0;
while (i < 10) {
    System.out.println(i);
    i++;
}

// Do-while loop
int j = 0;
do {
    System.out.println(j);
    j++;
} while (j < 10);

// Nested loops with labeled break
outer: for (int x = 0; x < 3; x++) {
    for (int y = 0; y < 3; y++) {
        if (x == 1 && y == 1) {
            break outer; // Break out of both loops
        }
        System.out.println("(" + x + "," + y + ")");
    }
}
```

### Method Declaration Syntax
```java
// Method signature components
[access_modifier] [static] [final] [abstract] [synchronized]
return_type method_name(parameter_list) [throws exception_list] {
    // Method body
}

// Examples
public static void main(String[] args) { }

private int calculateSum(int a, int b) {
    return a + b;
}

protected abstract void abstractMethod();

public synchronized void threadSafeMethod() {
    // Method body
}

// Varargs method (variable arguments)
public void printNumbers(int... numbers) {
    for (int number : numbers) {
        System.out.println(number);
    }
}

// Generic method
public <T> T genericMethod(T item) {
    return item;
}
```

### Class Declaration Syntax
```java
// Class declaration components
[access_modifier] [abstract|final] class ClassName
    [extends SuperClass] [implements Interface1, Interface2] {

    // Class members: fields, constructors, methods, nested classes
}

// Examples
public class Animal {
    private String name;

    public Animal(String name) {
        this.name = name;
    }
}

public abstract class Shape {
    abstract double getArea();
}

public final class ImmutableClass {
    private final String value;

    public ImmutableClass(String value) {
        this.value = value;
    }
}

public class Dog extends Animal implements Comparable<Dog> {
    public Dog(String name) {
        super(name);
    }

    @Override
    public int compareTo(Dog other) {
        return this.name.compareTo(other.name);
    }
}
```

### Interface Declaration Syntax
```java
// Interface declaration
[access_modifier] interface InterfaceName
    [extends Interface1, Interface2] {

    // Constants (implicitly public, static, final)
    String CONSTANT = "value";

    // Abstract methods (implicitly public, abstract)
    void abstractMethod();

    // Default methods (Java 8+)
    default void defaultMethod() {
        System.out.println("Default implementation");
    }

    // Static methods (Java 8+)
    static void staticMethod() {
        System.out.println("Static method in interface");
    }

    // Private methods (Java 9+)
    private void privateHelperMethod() {
        System.out.println("Private helper method");
    }
}
```

### Exception Handling Syntax
```java
// Try-catch-finally block
try {
    // Code that may throw exceptions
    int result = divide(10, 0);
} catch (ArithmeticException e) {
    // Handle specific exception
    System.err.println("Division by zero: " + e.getMessage());
} catch (Exception e) {
    // Handle general exception
    System.err.println("General error: " + e.getMessage());
} finally {
    // Always executes (optional)
    System.out.println("Cleanup code");
}

// Multi-catch (Java 7+)
try {
    // Code that may throw multiple exception types
} catch (IOException | SQLException e) {
    // Handle multiple exception types
    System.err.println("IO or SQL error: " + e.getMessage());
}

// Try-with-resources (Java 7+)
try (FileInputStream file = new FileInputStream("data.txt");
     BufferedReader reader = new BufferedReader(new InputStreamReader(file))) {

    // Use resources
    String line = reader.readLine();

} catch (IOException e) {
    // Handle exception
    System.err.println("File error: " + e.getMessage());
}
// Resources automatically closed

// Throwing exceptions
public void validateAge(int age) throws IllegalArgumentException {
    if (age < 0) {
        throw new IllegalArgumentException("Age cannot be negative");
    }
}
```

### Annotation Syntax
```java
// Built-in annotations
@Override
public String toString() {
    return "MyClass instance";
}

@Deprecated
public void oldMethod() {
    // Deprecated method
}

@SuppressWarnings("unchecked")
public void methodWithWarnings() {
    List rawList = new ArrayList();
}

// Custom annotation definition
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.METHOD)
public @interface MyAnnotation {
    String value() default "";
    int priority() default 0;
}

// Using custom annotation
@MyAnnotation(value = "important", priority = 1)
public void annotatedMethod() {
    // Method implementation
}
```

### Lambda Expressions and Functional Interfaces (Java 8+)
```java
// Lambda expression syntax
(parameters) -> { body }

// Examples
Runnable r = () -> System.out.println("Hello, Lambda!");

Function<String, Integer> stringLength = s -> s.length();

BiFunction<Integer, Integer, Integer> add = (a, b) -> a + b;

Predicate<String> isEmpty = String::isEmpty; // Method reference

// Using lambdas with streams
List<String> names = Arrays.asList("Alice", "Bob", "Charlie");
names.stream()
     .filter(name -> name.startsWith("A"))
     .map(String::toUpperCase)
     .forEach(System.out::println);
```

### Generics Syntax
```java
// Generic class
public class Box<T> {
    private T content;

    public void set(T content) {
        this.content = content;
    }

    public T get() {
        return content;
    }
}

// Bounded generics
public class NumberBox<T extends Number> {
    private T number;
}

// Generic methods
public <T> void swap(T[] array, int i, int j) {
    T temp = array[i];
    array[i] = array[j];
    array[j] = temp;
}

// Wildcards
List<? extends Number> numbers = new ArrayList<Integer>();
List<? super Integer> integers = new ArrayList<Number>();
List<?> unknown = new ArrayList<String>();
```

### Record Syntax (Java 14+, Standard in Java 16)
```java
// Record declaration
public record Person(String name, int age) {
    // Compact constructor
    public Person {
        if (age < 0) {
            throw new IllegalArgumentException("Age cannot be negative");
        }
    }

    // Additional methods
    public boolean isAdult() {
        return age >= 18;
    }
}

// Usage
Person person = new Person("Alice", 30);
String name = person.name(); // Accessor method
int age = person.age();      // Accessor method
```

### Pattern Matching (Java 17+)
```java
// Pattern matching with instanceof
if (obj instanceof String s) {
    // s is automatically cast to String
    System.out.println(s.toUpperCase());
}

// Pattern matching in switch (Preview in Java 17+)
public String formatValue(Object value) {
    return switch (value) {
        case null -> "null";
        case Integer i -> String.format("Integer: %d", i);
        case String s -> String.format("String: %s", s);
        case Double d -> String.format("Double: %.2f", d);
        default -> value.toString();
    };
}
```

### Text Blocks (Java 15+)
```java
// Multi-line string literals
String html = """
    <html>
        <head>
            <title>My Page</title>
        </head>
        <body>
            <h1>Hello, World!</h1>
        </body>
    </html>
    """;

String sql = """
    SELECT name, email, phone
    FROM users
    WHERE status = 'active'
    ORDER BY name
    """;
```

## Syntax Rules and Conventions

### Naming Conventions
```java
// Classes: PascalCase
class MyClass { }
class CustomerService { }

// Methods and variables: camelCase
int myVariable;
public void calculateTotal() { }

// Constants: UPPER_SNAKE_CASE
public static final int MAX_SIZE = 100;
private static final String DEFAULT_NAME = "Unknown";

// Packages: lowercase, dot-separated
package com.example.myapp.service;

// Interfaces: Often end with -able or start with I
interface Drawable { }
interface IUserService { } // Less common in Java
```

### Code Formatting and Style
```java
// Braces placement (K&R style - most common in Java)
public class MyClass {
    public void myMethod() {
        if (condition) {
            // Code here
        } else {
            // Code here
        }
    }
}

// Indentation: 4 spaces (most common) or 2 spaces
public void method() {
    if (condition) {
        doSomething();
        doSomethingElse();
    }
}

// Line length: typically 80-120 characters
public void longMethodName(String parameterOne,
                          String parameterTwo,
                          String parameterThree) {
    // Method body
}
```

### Comments
```java
// Single-line comment
int x = 5; // End-of-line comment

/*
 * Multi-line comment
 * Can span multiple lines
 */

/**
 * Javadoc comment for classes, methods, and fields
 * @param name the name parameter
 * @return the calculated result
 * @throws IllegalArgumentException if name is null
 * @since 1.0
 * @author John Doe
 */
public int calculateValue(String name) {
    return name.length();
}
```

This comprehensive reference covers all Java keywords, syntax structures, and conventions, providing a complete foundation for understanding and writing Java code at any level.