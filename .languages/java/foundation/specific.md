# Java-Specific Features and Unique Capabilities

## Project Loom - Virtual Threads (Java 19+, Standard in 21)

### Virtual Thread Architecture
```java
public class VirtualThreadFeatures {

    // Creating virtual threads
    public void basicVirtualThreads() throws InterruptedException {
        // Method 1: Thread.ofVirtual()
        Thread vThread1 = Thread.ofVirtual()
                               .name("virtual-worker-1")
                               .start(() -> {
                                   System.out.println("Running on: " + Thread.currentThread());
                               });

        // Method 2: Using builder pattern
        Thread.Builder.OfVirtual builder = Thread.ofVirtual().name("worker-", 0);

        for (int i = 0; i < 10; i++) {
            builder.start(() -> {
                try {
                    Thread.sleep(Duration.ofMillis(1000));
                    System.out.println("Virtual thread completed on: " + Thread.currentThread());
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                }
            });
        }

        vThread1.join();
    }

    // Executor service with virtual threads
    public void virtualThreadExecutor() throws InterruptedException {
        try (ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor()) {

            // Can handle millions of virtual threads efficiently
            List<Future<String>> futures = new ArrayList<>();

            for (int i = 0; i < 10000; i++) {
                int taskId = i;
                Future<String> future = executor.submit(() -> {
                    Thread.sleep(Duration.ofSeconds(1));
                    return "Task " + taskId + " completed";
                });
                futures.add(future);
            }

            // Collect results
            for (Future<String> future : futures) {
                System.out.println(future.get());
            }
        }
    }
}
```

### Structured Concurrency (Preview Feature)
```java
public class StructuredConcurrencyExample {

    public String fetchUserData(String userId) throws InterruptedException, ExecutionException {
        try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {

            // Start multiple related tasks
            Future<String> userProfile = scope.fork(() -> fetchUserProfile(userId));
            Future<String> userPreferences = scope.fork(() -> fetchUserPreferences(userId));
            Future<String> userActivity = scope.fork(() -> fetchUserActivity(userId));

            // Wait for all to complete or any to fail
            scope.join();           // Wait for all tasks
            scope.throwIfFailed();  // Propagate any failures

            // All tasks succeeded - combine results
            return combineUserData(
                userProfile.resultNow(),
                userPreferences.resultNow(),
                userActivity.resultNow()
            );
        }
    }

    private String fetchUserProfile(String userId) {
        // Simulate API call
        try { Thread.sleep(100); } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        return "Profile for " + userId;
    }

    private String fetchUserPreferences(String userId) {
        // Simulate API call
        try { Thread.sleep(200); } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        return "Preferences for " + userId;
    }

    private String fetchUserActivity(String userId) {
        // Simulate API call
        try { Thread.sleep(150); } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        return "Activity for " + userId;
    }

    private String combineUserData(String profile, String preferences, String activity) {
        return String.format("User Data: [%s, %s, %s]", profile, preferences, activity);
    }
}
```

## Record Classes (Java 14 Preview, Standard in 16)

### Basic Records
```java
// Immutable data carrier
public record Person(String name, int age, String email) {

    // Compact constructor with validation
    public Person {
        if (name == null || name.isBlank()) {
            throw new IllegalArgumentException("Name cannot be null or blank");
        }
        if (age < 0) {
            throw new IllegalArgumentException("Age cannot be negative");
        }
        if (email == null || !email.contains("@")) {
            throw new IllegalArgumentException("Invalid email");
        }
    }

    // Additional methods
    public boolean isAdult() {
        return age >= 18;
    }

    public String getDisplayName() {
        return name + " (" + age + ")";
    }
}

// Usage
public class RecordExample {
    public void useRecords() {
        Person person = new Person("Alice", 30, "alice@example.com");

        // Automatic getters
        String name = person.name();
        int age = person.age();
        String email = person.email();

        // Automatic equals, hashCode, toString
        System.out.println(person); // Person[name=Alice, age=30, email=alice@example.com]

        // Records work well with pattern matching
        if (person instanceof Person(var n, var a, var e)) {
            System.out.println("Extracted: " + n + ", " + a + ", " + e);
        }
    }
}
```

### Complex Records
```java
public record Address(String street, String city, String zipCode) {}

public record Employee(String name, int id, Address address, List<String> skills) {

    // Alternative constructor
    public Employee(String name, int id, Address address) {
        this(name, id, address, new ArrayList<>());
    }

    // Defensive copying for mutable fields
    public Employee(String name, int id, Address address, List<String> skills) {
        this.name = name;
        this.id = id;
        this.address = address;
        this.skills = List.copyOf(skills); // Immutable copy
    }

    // Custom accessor with logic
    public List<String> skills() {
        return List.copyOf(skills); // Return immutable copy
    }

    public Employee withSkill(String skill) {
        List<String> newSkills = new ArrayList<>(this.skills);
        newSkills.add(skill);
        return new Employee(name, id, address, newSkills);
    }
}
```

## Pattern Matching Enhancements

### Pattern Matching with instanceof (Java 16+)
```java
public class PatternMatchingExample {

    public String processObject(Object obj) {
        if (obj instanceof String s) {
            return "String: " + s.toUpperCase();
        } else if (obj instanceof Integer i) {
            return "Integer: " + (i * 2);
        } else if (obj instanceof List<?> list && !list.isEmpty()) {
            return "Non-empty list with " + list.size() + " elements";
        } else if (obj instanceof Person(var name, var age, var email)) {
            return "Person: " + name + " is " + age + " years old";
        }
        return "Unknown type: " + obj.getClass().getSimpleName();
    }

    // Pattern matching in switch (Preview in Java 17+)
    public String switchPatterns(Object value) {
        return switch (value) {
            case null -> "null value";
            case String s -> "String: " + s;
            case Integer i -> "Integer: " + i;
            case Double d -> String.format("Double: %.2f", d);
            case List<?> list -> "List with " + list.size() + " elements";
            case Person(var name, var age, var email) ->
                "Person: " + name + " (" + age + ")";
            default -> "Unknown: " + value.toString();
        };
    }
}
```

## Sealed Classes (Java 17+)

### Controlling Inheritance
```java
// Sealed class - restricts which classes can extend it
public abstract sealed class Shape permits Circle, Rectangle, Triangle {
    public abstract double area();
}

// Permitted subclasses
public final class Circle extends Shape {
    private final double radius;

    public Circle(double radius) {
        this.radius = radius;
    }

    @Override
    public double area() {
        return Math.PI * radius * radius;
    }
}

public final class Rectangle extends Shape {
    private final double width, height;

    public Rectangle(double width, double height) {
        this.width = width;
        this.height = height;
    }

    @Override
    public double area() {
        return width * height;
    }
}

public non-sealed class Triangle extends Shape {
    private final double base, height;

    public Triangle(double base, double height) {
        this.base = base;
        this.height = height;
    }

    @Override
    public double area() {
        return 0.5 * base * height;
    }
}

// Usage with exhaustive pattern matching
public class SealedClassExample {
    public String describeShape(Shape shape) {
        return switch (shape) {
            case Circle(var radius) ->
                "Circle with radius " + radius;
            case Rectangle(var width, var height) ->
                "Rectangle " + width + "x" + height;
            case Triangle(var base, var height) ->
                "Triangle with base " + base + " and height " + height;
            // No default needed - compiler knows all possibilities
        };
    }
}
```

## Text Blocks (Java 15+)

### Multi-line String Literals
```java
public class TextBlockExample {

    public void demonstrateTextBlocks() {
        // Traditional string concatenation
        String oldHtml = "<html>\n" +
                        "    <head>\n" +
                        "        <title>My Page</title>\n" +
                        "    </head>\n" +
                        "    <body>\n" +
                        "        <h1>Hello World</h1>\n" +
                        "    </body>\n" +
                        "</html>";

        // Text block (much cleaner)
        String newHtml = """
            <html>
                <head>
                    <title>My Page</title>
                </head>
                <body>
                    <h1>Hello World</h1>
                </body>
            </html>
            """;

        // SQL query example
        String sql = """
            SELECT u.name, u.email, p.title
            FROM users u
            INNER JOIN posts p ON u.id = p.user_id
            WHERE u.status = 'active'
            AND p.created_date >= ?
            ORDER BY p.created_date DESC
            """;

        // JSON example with formatting
        String json = """
            {
                "name": "John Doe",
                "age": 30,
                "address": {
                    "street": "123 Main St",
                    "city": "Anytown",
                    "zipCode": "12345"
                },
                "hobbies": ["reading", "swimming", "coding"]
            }
            """;

        // Text block with escape sequences
        String formatted = """
            Line 1
            Line 2 with "quotes"
            Line 3 with \\ backslash
            Line 4 with \t tab
            """;
    }
}
```

## Switch Expressions (Java 14+)

### Enhanced Switch Syntax
```java
public class SwitchExpressionExample {

    // Traditional switch statement
    public String oldStyleSwitch(int day) {
        String result;
        switch (day) {
            case 1:
            case 2:
            case 3:
            case 4:
            case 5:
                result = "Weekday";
                break;
            case 6:
            case 7:
                result = "Weekend";
                break;
            default:
                result = "Invalid day";
                break;
        }
        return result;
    }

    // Modern switch expression
    public String newStyleSwitch(int day) {
        return switch (day) {
            case 1, 2, 3, 4, 5 -> "Weekday";
            case 6, 7 -> "Weekend";
            default -> "Invalid day";
        };
    }

    // Switch expression with yield
    public String switchWithYield(int value) {
        return switch (value) {
            case 1 -> "one";
            case 2 -> "two";
            case 3 -> {
                System.out.println("Processing three");
                yield "three"; // yield keyword for block
            }
            default -> {
                System.out.println("Processing other: " + value);
                yield "other";
            }
        };
    }

    // Switch with pattern matching (Preview)
    public String patternSwitch(Object obj) {
        return switch (obj) {
            case String s when s.length() > 5 -> "Long string: " + s;
            case String s -> "Short string: " + s;
            case Integer i when i > 0 -> "Positive integer: " + i;
            case Integer i -> "Non-positive integer: " + i;
            case null -> "null value";
            default -> "Other type";
        };
    }
}
```

## Local Variable Type Inference (var) - Java 10+

### Effective Use of var
```java
public class VarExample {

    public void demonstrateVar() {
        // Good uses of var
        var list = new ArrayList<String>();           // Clear from right side
        var map = new HashMap<String, Integer>();     // Avoids repetition
        var stream = list.stream()                    // Complex return types
                        .filter(s -> s.startsWith("A"))
                        .map(String::toUpperCase);

        // var with generic methods
        var optional = Optional.of("Hello");
        var future = CompletableFuture.supplyAsync(() -> "World");

        // var in try-with-resources
        try (var reader = Files.newBufferedReader(Paths.get("file.txt"))) {
            // Use reader
        } catch (IOException e) {
            e.printStackTrace();
        }

        // var in loops
        var numbers = List.of(1, 2, 3, 4, 5);
        for (var number : numbers) {
            System.out.println(number);
        }

        // Anonymous classes with var
        var comparator = new Comparator<String>() {
            @Override
            public int compare(String s1, String s2) {
                return s1.length() - s2.length();
            }
        };
    }

    public void varGuidelines() {
        // Good - type is obvious
        var fileName = "document.pdf";
        var users = getUserList();
        var count = list.size();

        // Questionable - type not obvious
        // var data = getData();        // What type is data?
        // var result = process();      // What does process() return?

        // Better - use explicit types when unclear
        String data = getData();
        ProcessResult result = process();
    }

    private List<String> getUserList() { return new ArrayList<>(); }
    private String getData() { return "data"; }
    private ProcessResult process() { return new ProcessResult(); }

    private static class ProcessResult {}
}
```

## Module System (Java 9+)

### Module Declaration
```java
// module-info.java
module com.example.myapp {
    // Dependencies on other modules
    requires java.base;          // Implicit - always present
    requires java.logging;       // Standard library module
    requires transitive java.sql; // Transitive - consumers get this too

    // Optional dependency
    requires static java.compiler; // Compile-time only

    // External library
    requires com.fasterxml.jackson.core;

    // Exports - what other modules can access
    exports com.example.myapp.api;           // Public API
    exports com.example.myapp.internal to   // Qualified export
        com.example.myapp.test,
        com.example.myapp.integration;

    // Services
    uses com.example.myapp.spi.PluginProvider;    // Service consumer
    provides com.example.myapp.spi.PluginProvider  // Service provider
        with com.example.myapp.impl.DefaultPluginProvider;

    // Reflection access (if needed)
    opens com.example.myapp.model;           // For frameworks
    opens com.example.myapp.config to       // Qualified opens
        com.fasterxml.jackson.databind;
}
```

### Module Usage Example
```java
package com.example.myapp.api;

// Public API class - exported by module
public class Calculator {
    public int add(int a, int b) {
        return a + b;
    }

    public int subtract(int a, int b) {
        return a - b;
    }
}

package com.example.myapp.internal;

// Internal implementation - not exported
class CalculatorImpl {
    public double complexCalculation(double x, double y) {
        // Internal implementation details
        return Math.sqrt(x * x + y * y);
    }
}
```

## Enhanced NullPointerException Messages (Java 14+)

### Helpful NPE Messages
```java
public class NPEExample {
    static class Person {
        String name;
        Address address;

        public String getName() { return name; }
        public Address getAddress() { return address; }
    }

    static class Address {
        String street;
        public String getStreet() { return street; }
    }

    public void demonstrateEnhancedNPE() {
        Person person = new Person();
        // person.address is null

        try {
            // This will throw NPE with helpful message:
            // "Cannot invoke 'Address.getStreet()' because the return value
            //  of 'Person.getAddress()' is null"
            String street = person.getAddress().getStreet();
        } catch (NullPointerException e) {
            System.out.println("Enhanced NPE message: " + e.getMessage());
        }

        String[] array = null;
        try {
            // "Cannot read the array length because '<local1>' is null"
            int length = array.length;
        } catch (NullPointerException e) {
            System.out.println("Array NPE: " + e.getMessage());
        }
    }
}
```

## Foreign Function & Memory API (Project Panama - Preview)

### Accessing Native Libraries
```java
import java.lang.foreign.*;
import java.lang.invoke.MethodHandle;

public class ForeignFunctionExample {

    // Define C function signature: int strlen(const char* str)
    private static final FunctionDescriptor STRLEN_DESC =
        FunctionDescriptor.of(ValueLayout.JAVA_INT, ValueLayout.ADDRESS);

    public void callNativeFunction() {
        try (Arena arena = Arena.ofConfined()) {
            // Get system linker
            Linker linker = Linker.nativeLinker();

            // Look up native function
            SymbolLookup lookup = SymbolLookup.loaderLookup();
            MemorySegment strlenAddr = lookup.find("strlen").orElseThrow();

            // Create method handle
            MethodHandle strlen = linker.downcallHandle(strlenAddr, STRLEN_DESC);

            // Allocate native memory for string
            MemorySegment cString = arena.allocateUtf8String("Hello, native world!");

            // Call native function
            int length = (int) strlen.invoke(cString);
            System.out.println("String length from C strlen: " + length);

        } catch (Throwable e) {
            e.printStackTrace();
        }
    }

    // Memory segment operations
    public void memorySegmentExample() {
        try (Arena arena = Arena.ofConfined()) {
            // Allocate native memory
            MemorySegment segment = arena.allocate(1000);

            // Write data
            segment.setAtIndex(ValueLayout.JAVA_INT, 0, 42);
            segment.setAtIndex(ValueLayout.JAVA_DOUBLE, 1, 3.14159);

            // Read data back
            int intValue = segment.getAtIndex(ValueLayout.JAVA_INT, 0);
            double doubleValue = segment.getAtIndex(ValueLayout.JAVA_DOUBLE, 1);

            System.out.println("Int: " + intValue + ", Double: " + doubleValue);
        }
    }
}
```

## JDK Flight Recorder (JFR) Integration

### Custom JFR Events
```java
import jdk.jfr.*;

@Name("com.example.UserAction")
@Label("User Action")
@Description("Tracks user actions in the application")
@Category("Application")
public class UserActionEvent extends Event {

    @Label("User ID")
    String userId;

    @Label("Action Type")
    String actionType;

    @Label("Duration")
    @Timespan(Timespan.MILLISECONDS)
    long duration;

    @Label("Success")
    boolean success;
}

// Usage
public class JFRExample {

    public void trackUserAction(String userId, String actionType) {
        UserActionEvent event = new UserActionEvent();
        event.userId = userId;
        event.actionType = actionType;
        event.begin(); // Start timing

        try {
            // Perform user action
            performAction(actionType);
            event.success = true;
        } catch (Exception e) {
            event.success = false;
            throw e;
        } finally {
            event.end();   // End timing and record event
            event.commit(); // Send to JFR
        }
    }

    private void performAction(String actionType) {
        // Simulate action
        try {
            Thread.sleep((long) (Math.random() * 100));
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }
}
```

These Java-specific features demonstrate the language's continued evolution, focusing on developer productivity, performance, and maintainability while maintaining backward compatibility and enterprise reliability.