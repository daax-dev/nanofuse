# Java Version History - Complete Feature Evolution (Java 1.5 - Java 23)

## Table of Contents
- [Java 1.5/5.0 (September 2004) - Tiger](#java-15-50-september-2004---tiger)
- [Java 6 (December 2006) - Mustang](#java-6-december-2006---mustang)
- [Java 7 (July 2011) - Dolphin](#java-7-july-2011---dolphin)
- [Java 8 (March 2014) - Spider (LTS)](#java-8-march-2014---spider-lts)
- [Java 9 (September 2017) - Jigsaw](#java-9-september-2017---jigsaw)
- [Java 10 (March 2018)](#java-10-march-2018)
- [Java 11 (September 2018) - LTS](#java-11-september-2018---lts)
- [Java 12 (March 2019)](#java-12-march-2019)
- [Java 13 (September 2019)](#java-13-september-2019)
- [Java 14 (March 2020)](#java-14-march-2020)
- [Java 15 (September 2020)](#java-15-september-2020)
- [Java 16 (March 2021)](#java-16-march-2021)
- [Java 17 (September 2021) - LTS](#java-17-september-2021---lts)
- [Java 18 (March 2022)](#java-18-march-2022)
- [Java 19 (September 2022)](#java-19-september-2022)
- [Java 20 (March 2023)](#java-20-march-2023)
- [Java 21 (September 2023) - LTS](#java-21-september-2023---lts)
- [Java 22 (March 2024)](#java-22-march-2024)
- [Java 23 (September 2024)](#java-23-september-2024)
- [LTS Release Schedule](#lts-release-schedule)
- [Version Adoption Recommendations](#version-adoption-recommendations)

---

## Java 1.5 / 5.0 (September 2004) - Tiger

**Major Theme**: Developer productivity and type safety through language enhancements

### 🔥 Major Features

#### 1. **Generics - Type Safety Revolution**
The most significant language enhancement, providing compile-time type safety for collections.

```java
// Before Java 5 - Raw types, runtime ClassCastException risk
List list = new ArrayList();
list.add("String");
list.add(42);  // No compile-time error
String str = (String) list.get(1);  // Runtime ClassCastException

// Java 5+ - Type-safe collections
List<String> list = new ArrayList<String>();
list.add("String");
// list.add(42);  // Compile-time error!
String str = list.get(0);  // No casting needed
```

**Advanced Generics Features:**
```java
// Bounded type parameters
public class NumberContainer<T extends Number> {
    private T value;
    public T getValue() { return value; }
    public void setValue(T value) { this.value = value; }
}

// Wildcards for flexibility
List<? extends Number> numbers = new ArrayList<Integer>();
List<? super Integer> integers = new ArrayList<Number>();

// Generic methods
public static <T> void swap(T[] array, int i, int j) {
    T temp = array[i];
    array[i] = array[j];
    array[j] = temp;
}
```

#### 2. **Enhanced For-Loop (for-each)**
Simplified iteration over collections and arrays.

```java
// Traditional for loop
List<String> names = Arrays.asList("Alice", "Bob", "Charlie");
for (int i = 0; i < names.size(); i++) {
    System.out.println(names.get(i));
}

// Enhanced for-loop (for-each)
for (String name : names) {
    System.out.println(name);
}

// Works with arrays too
int[] numbers = {1, 2, 3, 4, 5};
for (int number : numbers) {
    System.out.println(number);
}
```

#### 3. **Autoboxing and Unboxing**
Automatic conversion between primitive types and their wrapper classes.

```java
// Before Java 5 - Manual boxing/unboxing
Integer wrapped = new Integer(42);  // Boxing
int primitive = wrapped.intValue(); // Unboxing

// Java 5+ - Automatic boxing/unboxing
Integer wrapped = 42;        // Autoboxing
int primitive = wrapped;     // Auto-unboxing

List<Integer> numbers = new ArrayList<>();
numbers.add(42);            // Autoboxing int to Integer
int first = numbers.get(0); // Auto-unboxing Integer to int
```

#### 4. **Enumerations (Enums)**
Type-safe, powerful alternative to integer constants.

```java
// Basic enum
public enum Day {
    MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY
}

// Advanced enum with methods and fields
public enum Planet {
    MERCURY(3.303e+23, 2.4397e6),
    VENUS(4.869e+24, 6.0518e6),
    EARTH(5.976e+24, 6.37814e6);

    private final double mass;   // in kilograms
    private final double radius; // in meters

    Planet(double mass, double radius) {
        this.mass = mass;
        this.radius = radius;
    }

    public double surfaceGravity() {
        return G * mass / (radius * radius);
    }

    private static final double G = 6.67300E-11;
}

// Usage
Day today = Day.MONDAY;
switch (today) {
    case MONDAY:
        System.out.println("Back to work!");
        break;
    case FRIDAY:
        System.out.println("TGIF!");
        break;
}
```

#### 5. **Variable Arguments (Varargs)**
Methods can accept variable number of arguments.

```java
// Varargs method
public static void printNumbers(int... numbers) {
    System.out.println("Count: " + numbers.length);
    for (int number : numbers) {
        System.out.println(number);
    }
}

// Usage - all valid calls
printNumbers();                    // 0 arguments
printNumbers(1);                   // 1 argument
printNumbers(1, 2, 3, 4, 5);      // 5 arguments
printNumbers(new int[]{1, 2, 3});  // Array argument

// String formatting with varargs
String message = String.format("Hello %s, you have %d messages", "Alice", 5);
```

#### 6. **Annotations**
Metadata that can be processed at compile-time or runtime.

```java
// Built-in annotations
@Override
public String toString() {
    return "MyClass instance";
}

@Deprecated
public void oldMethod() {
    // Legacy method
}

@SuppressWarnings("unchecked")
public void rawTypeMethod() {
    List list = new ArrayList();  // Raw type warning suppressed
}

// Custom annotations
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.METHOD)
public @interface Benchmark {
    String value() default "";
    int iterations() default 1;
}

// Using custom annotation
@Benchmark(value = "performance test", iterations = 1000)
public void performanceMethod() {
    // Method implementation
}
```

#### 7. **Static Imports**
Import static methods and fields directly.

```java
// Without static import
System.out.println(Math.PI);
System.out.println(Math.sqrt(16));

// With static import
import static java.lang.System.out;
import static java.lang.Math.*;

out.println(PI);
out.println(sqrt(16));

// Especially useful with constants and utility methods
import static java.util.Collections.*;
List<String> list = new ArrayList<>();
addAll(list, "a", "b", "c");
sort(list);
```

### 🛠 **Additional Improvements**

#### **StringBuilder Class**
More efficient string concatenation than StringBuffer for single-threaded use.

```java
// StringBuilder - not synchronized, faster
StringBuilder sb = new StringBuilder();
sb.append("Hello").append(" ").append("World");
String result = sb.toString();
```

#### **Scanner Class**
Simplified text parsing and input reading.

```java
Scanner scanner = new Scanner(System.in);
System.out.print("Enter your name: ");
String name = scanner.nextLine();
System.out.print("Enter your age: ");
int age = scanner.nextInt();

// File scanning
Scanner fileScanner = new Scanner(new File("data.txt"));
while (fileScanner.hasNextLine()) {
    System.out.println(fileScanner.nextLine());
}
```

#### **ProcessBuilder**
More flexible process creation than Runtime.exec().

```java
ProcessBuilder pb = new ProcessBuilder("cmd", "/c", "dir");
pb.directory(new File("C:\\"));
pb.redirectErrorStream(true);
Process process = pb.start();
```

### 📈 **Performance Impact**
- **Generics**: No runtime overhead (type erasure)
- **Enhanced for-loop**: Slight performance improvement over traditional loops
- **Autoboxing**: Potential performance cost due to object creation
- **StringBuilder**: 10-50x faster than String concatenation in loops

---

## Java 6 (December 2006) - Mustang

**Major Theme**: Performance improvements, web services, and developer tools

### 🔥 Major Features

#### 1. **Scripting Engine Support (JSR 223)**
Embed and execute scripts in Java applications.

```java
ScriptEngineManager manager = new ScriptEngineManager();
ScriptEngine engine = manager.getEngineByName("JavaScript");

// Execute JavaScript code
engine.eval("print('Hello from JavaScript!')");

// Set variables and call functions
engine.put("x", 10);
engine.put("y", 20);
Object result = engine.eval("x + y");
System.out.println("Result: " + result); // 30

// Load and execute script files
engine.eval(new FileReader("script.js"));
```

#### 2. **Web Services Support**
Built-in JAX-WS for SOAP web services.

```java
// Web service endpoint
@WebService
public class CalculatorService {
    @WebMethod
    public int add(int a, int b) {
        return a + b;
    }

    @WebMethod
    public int multiply(int a, int b) {
        return a * b;
    }
}

// Publishing the service
CalculatorService service = new CalculatorService();
Endpoint endpoint = Endpoint.publish("http://localhost:8080/calculator", service);
```

#### 3. **Compiler API**
Programmatically compile Java source code.

```java
JavaCompiler compiler = ToolProvider.getSystemJavaCompiler();
StandardJavaFileManager fileManager = compiler.getStandardFileManager(null, null, null);

// Compile source files
Iterable<? extends JavaFileObject> compilationUnits =
    fileManager.getJavaFileObjectsFromStrings(Arrays.asList("MyClass.java"));

CompilationTask task = compiler.getTask(null, fileManager, null, null, null, compilationUnits);
boolean success = task.call();
```

#### 4. **Desktop Integration**
Better integration with the native desktop environment.

```java
Desktop desktop = Desktop.getDesktop();

// Open default browser
if (desktop.isSupported(Desktop.Action.BROWSE)) {
    desktop.browse(new URI("https://www.oracle.com"));
}

// Open default email client
if (desktop.isSupported(Desktop.Action.MAIL)) {
    desktop.mail(new URI("mailto:user@example.com?subject=Hello"));
}

// Open files with default applications
if (desktop.isSupported(Desktop.Action.OPEN)) {
    desktop.open(new File("document.pdf"));
}
```

### 🛠 **Performance & JVM Improvements**

#### **Performance Enhancements**
- **Parallel Garbage Collector improvements**: Better throughput for multi-core systems
- **Escape Analysis**: Objects that don't escape method scope can be allocated on stack
- **Lock optimization**: Biased locking and lock elimination
- **HotSpot improvements**: Better compilation and optimization

#### **Memory Management**
```java
// New memory pool management
MemoryMXBean memoryBean = ManagementFactory.getMemoryMXBean();
MemoryUsage heapUsage = memoryBean.getHeapMemoryUsage();

System.out.println("Used: " + heapUsage.getUsed());
System.out.println("Max: " + heapUsage.getMax());
```

### 🔧 **Developer Tools**

#### **Monitoring and Management**
- **JConsole improvements**: Better GUI for monitoring JVM
- **MXBeans**: More comprehensive monitoring capabilities
- **jmap, jstack, jstat**: Command-line diagnostic tools

#### **Annotation Processing**
```java
// Pluggable Annotation Processing API
@SupportedAnnotationTypes("com.example.MyAnnotation")
@SupportedSourceVersion(SourceVersion.RELEASE_6)
public class MyProcessor extends AbstractProcessor {
    @Override
    public boolean process(Set<? extends TypeElement> annotations,
                          RoundEnvironment roundEnv) {
        // Process annotations at compile time
        return true;
    }
}
```

---

## Java 7 (July 2011) - Dolphin

**Major Theme**: Language improvements and "Project Coin" small enhancements

### 🔥 Major Features

#### 1. **Try-with-Resources Statement**
Automatic resource management for objects implementing AutoCloseable.

```java
// Before Java 7 - Manual resource management
FileInputStream fis = null;
try {
    fis = new FileInputStream("file.txt");
    // Use the file
} finally {
    if (fis != null) {
        try {
            fis.close();
        } catch (IOException e) {
            // Handle close exception
        }
    }
}

// Java 7+ - Automatic resource management
try (FileInputStream fis = new FileInputStream("file.txt");
     BufferedReader reader = new BufferedReader(new InputStreamReader(fis))) {

    String line = reader.readLine();
    // Resources automatically closed
} catch (IOException e) {
    // Handle exception
}

// Multiple resources
try (FileInputStream in = new FileInputStream("input.txt");
     FileOutputStream out = new FileOutputStream("output.txt")) {

    // Copy file content
    byte[] buffer = new byte[1024];
    int bytesRead;
    while ((bytesRead = in.read(buffer)) != -1) {
        out.write(buffer, 0, bytesRead);
    }
}
```

#### 2. **Multi-catch Exception Handling**
Handle multiple exception types in a single catch block.

```java
// Before Java 7 - Duplicate code
try {
    // Some operation
} catch (IOException e) {
    logger.log(e);
    throw e;
} catch (SQLException e) {
    logger.log(e);
    throw e;
} catch (ClassNotFoundException e) {
    logger.log(e);
    throw e;
}

// Java 7+ - Multi-catch
try {
    // Some operation
} catch (IOException | SQLException | ClassNotFoundException e) {
    logger.log(e);
    throw e;
}

// Combined with specific handling
try {
    // Database and file operations
} catch (SQLException e) {
    // Handle database errors specifically
    rollbackTransaction();
} catch (IOException | ClassNotFoundException e) {
    // Handle other errors
    cleanupResources();
}
```

#### 3. **String Switch Statements**
Switch statements can now use String literals.

```java
String day = "MONDAY";

// Java 7+ - String switch
switch (day) {
    case "MONDAY":
        System.out.println("Start of work week");
        break;
    case "TUESDAY":
    case "WEDNESDAY":
    case "THURSDAY":
        System.out.println("Midweek");
        break;
    case "FRIDAY":
        System.out.println("TGIF!");
        break;
    case "SATURDAY":
    case "SUNDAY":
        System.out.println("Weekend!");
        break;
    default:
        System.out.println("Invalid day");
}

// Practical example - HTTP status handling
String httpMethod = request.getMethod();
switch (httpMethod) {
    case "GET":
        handleGet(request, response);
        break;
    case "POST":
        handlePost(request, response);
        break;
    case "PUT":
        handlePut(request, response);
        break;
    case "DELETE":
        handleDelete(request, response);
        break;
    default:
        response.sendError(405, "Method Not Allowed");
}
```

#### 4. **Diamond Operator**
Type inference for generic instance creation.

```java
// Before Java 7 - Redundant type information
Map<String, List<Integer>> map = new HashMap<String, List<Integer>>();
List<String> list = new ArrayList<String>();

// Java 7+ - Diamond operator
Map<String, List<Integer>> map = new HashMap<>();
List<String> list = new ArrayList<>();

// Complex nested types
Map<String, Map<Integer, List<String>>> complex = new HashMap<>();
```

#### 5. **Binary Literals and Numeric Literals with Underscores**
Improved readability for numeric literals.

```java
// Binary literals
int binary = 0b1010_1100_0011;  // Binary with underscores
byte flags = 0b0001_0010;

// Underscores in numeric literals for readability
long creditCardNumber = 1234_5678_9012_3456L;
int million = 1_000_000;
double pi = 3.14_15_92_65_35_89_79;

// Hexadecimal
int hex = 0xFF_EC_DE_5E;

// Practical examples
long maxMemory = 8_589_934_592L;  // 8GB in bytes
int bitmask = 0b1111_0000;        // Upper 4 bits set
```

### 🛠 **NIO.2 (New I/O 2)**
Comprehensive file system API improvements.

#### **Path and Files Classes**
```java
// Path operations
Path path = Paths.get("C:", "Users", "John", "Documents", "file.txt");
Path absolutePath = path.toAbsolutePath();
Path parent = path.getParent();
Path filename = path.getFileName();

// File operations
if (Files.exists(path)) {
    long size = Files.size(path);
    System.out.println("File size: " + size + " bytes");
}

// Copy, move, delete operations
Path source = Paths.get("source.txt");
Path target = Paths.get("target.txt");
Files.copy(source, target, StandardCopyOption.REPLACE_EXISTING);
Files.move(source, Paths.get("moved.txt"), StandardCopyOption.ATOMIC_MOVE);
Files.delete(target);

// Directory operations
Path directory = Paths.get("myDirectory");
Files.createDirectories(directory);

// List directory contents
try (DirectoryStream<Path> stream = Files.newDirectoryStream(directory)) {
    for (Path entry : stream) {
        System.out.println(entry.getFileName());
    }
}
```

#### **File Attributes and Permissions**
```java
// File attributes
BasicFileAttributes attrs = Files.readAttributes(path, BasicFileAttributes.class);
System.out.println("Creation time: " + attrs.creationTime());
System.out.println("Last modified: " + attrs.lastModifiedTime());
System.out.println("Size: " + attrs.size());

// POSIX permissions (Unix/Linux)
if (FileSystems.getDefault().supportedFileAttributeViews().contains("posix")) {
    Set<PosixFilePermission> permissions = PosixFilePermissions.fromString("rwxr--r--");
    Files.setPosixFilePermissions(path, permissions);
}
```

#### **Watch Service**
Monitor directory changes in real-time.

```java
WatchService watchService = FileSystems.getDefault().newWatchService();
Path directory = Paths.get("watchedDirectory");
directory.register(watchService,
    StandardWatchEventKinds.ENTRY_CREATE,
    StandardWatchEventKinds.ENTRY_DELETE,
    StandardWatchEventKinds.ENTRY_MODIFY);

while (true) {
    WatchKey key = watchService.take();

    for (WatchEvent<?> event : key.pollEvents()) {
        WatchEvent.Kind<?> kind = event.kind();
        Path fileName = (Path) event.context();

        System.out.println(kind + ": " + fileName);
    }

    if (!key.reset()) {
        break;
    }
}
```

### 🚀 **JVM and Performance Improvements**

#### **G1 Garbage Collector**
Low-latency garbage collector for large heap sizes.

```bash
# Enable G1GC
java -XX:+UseG1GC -XX:MaxGCPauseMillis=200 MyApplication
```

#### **Tiered Compilation**
Combines benefits of client and server compilers.

```bash
# Enable tiered compilation (default in Java 8+)
java -XX:+TieredCompilation MyApplication
```

#### **Compressed OOPs**
Reduced memory footprint for object references on 64-bit platforms.

```bash
# Enable compressed OOPs (default for heaps < 32GB)
java -XX:+UseCompressedOops MyApplication
```

---

## Java 8 (March 2014) - Spider (LTS)

**Major Theme**: Functional programming paradigm with lambdas and streams

### 🔥 Major Features

#### 1. **Lambda Expressions**
The most significant language enhancement since generics.

```java
// Before Java 8 - Anonymous inner classes
List<String> names = Arrays.asList("Alice", "Bob", "Charlie");
Collections.sort(names, new Comparator<String>() {
    @Override
    public int compare(String a, String b) {
        return a.compareTo(b);
    }
});

// Java 8+ - Lambda expressions
Collections.sort(names, (a, b) -> a.compareTo(b));
// Even shorter with method references
Collections.sort(names, String::compareTo);

// Various lambda syntax forms
Runnable r1 = () -> System.out.println("Hello World");
Predicate<String> isEmpty = s -> s.isEmpty();
Function<String, Integer> stringLength = String::length;
BinaryOperator<Integer> sum = (a, b) -> a + b;
Consumer<String> printer = System.out::println;
```

#### **Lambda Expressions in Detail**
```java
// No parameters
Runnable task = () -> System.out.println("Task executed");

// Single parameter (parentheses optional)
Consumer<String> printer = s -> System.out.println(s);
Consumer<String> printer2 = (s) -> System.out.println(s);

// Multiple parameters
BiFunction<Integer, Integer, Integer> add = (x, y) -> x + y;
BiFunction<Integer, Integer, Integer> multiply = (x, y) -> {
    System.out.println("Multiplying " + x + " and " + y);
    return x * y;
};

// Type inference
List<String> list = Arrays.asList("a", "b", "c");
list.forEach(s -> System.out.println(s.toUpperCase()));

// Capturing variables (effectively final)
String prefix = "Item: ";
list.forEach(s -> System.out.println(prefix + s));
```

#### 2. **Stream API**
Functional-style operations on collections.

```java
List<Person> people = Arrays.asList(
    new Person("Alice", 25, "Engineer"),
    new Person("Bob", 30, "Manager"),
    new Person("Charlie", 35, "Engineer"),
    new Person("Diana", 28, "Designer")
);

// Filter, map, and collect
List<String> engineerNames = people.stream()
    .filter(p -> p.getJob().equals("Engineer"))
    .map(Person::getName)
    .collect(Collectors.toList());

// Complex operations
Map<String, List<Person>> peopleByJob = people.stream()
    .collect(Collectors.groupingBy(Person::getJob));

double averageAge = people.stream()
    .mapToInt(Person::getAge)
    .average()
    .orElse(0.0);

// Parallel processing
long engineerCount = people.parallelStream()
    .filter(p -> p.getJob().equals("Engineer"))
    .count();
```

#### **Advanced Stream Operations**
```java
// FlatMap for nested structures
List<List<String>> nestedList = Arrays.asList(
    Arrays.asList("a", "b"),
    Arrays.asList("c", "d", "e"),
    Arrays.asList("f")
);

List<String> flatList = nestedList.stream()
    .flatMap(Collection::stream)
    .collect(Collectors.toList());

// Reduce operations
Optional<Integer> sum = Arrays.asList(1, 2, 3, 4, 5).stream()
    .reduce(Integer::sum);

String concatenated = Arrays.asList("a", "b", "c").stream()
    .reduce("", String::concat);

// Custom collectors
String joinedNames = people.stream()
    .map(Person::getName)
    .collect(Collectors.joining(", ", "[", "]"));

// Statistics
IntSummaryStatistics ageStats = people.stream()
    .mapToInt(Person::getAge)
    .summaryStatistics();
```

#### 3. **Functional Interfaces**
Interfaces with exactly one abstract method, enabling lambda expressions.

```java
// Built-in functional interfaces
@FunctionalInterface
interface MathOperation {
    int operate(int a, int b);
}

// Usage with lambdas
MathOperation addition = (a, b) -> a + b;
MathOperation subtraction = (a, b) -> a - b;
MathOperation multiplication = Integer::sum;

// Common functional interfaces
Predicate<String> isLong = s -> s.length() > 5;
Function<String, Integer> length = String::length;
Consumer<String> print = System.out::println;
Supplier<String> stringSupplier = () -> "Hello";
UnaryOperator<String> upperCase = String::toUpperCase;
BinaryOperator<Integer> max = Integer::max;

// Composing functions
Function<String, String> upperThenReverse = upperCase.andThen(s ->
    new StringBuilder(s).reverse().toString());
```

#### 4. **Method References**
Shorthand notation for lambda expressions calling existing methods.

```java
// Static method references
Function<String, Integer> parser = Integer::parseInt;
BinaryOperator<Integer> max = Integer::max;

// Instance method references
String prefix = "Hello ";
Function<String, String> greeter = prefix::concat;

// Constructor references
Supplier<List<String>> listSupplier = ArrayList::new;
Function<String, Person> personCreator = Person::new;

// Method reference types
List<String> names = Arrays.asList("alice", "bob", "charlie");

// Instance method of arbitrary object
names.stream().map(String::toUpperCase); // s -> s.toUpperCase()

// Static method
names.stream().map(String::valueOf); // s -> String.valueOf(s)

// Constructor
names.stream().map(StringBuilder::new); // s -> new StringBuilder(s)
```

#### 5. **Default Methods in Interfaces**
Interfaces can now have method implementations.

```java
interface Vehicle {
    void start();
    void stop();

    // Default method
    default void honk() {
        System.out.println("Beep beep!");
    }

    // Static method in interface
    static void checkSpeed(int speed) {
        if (speed > 100) {
            System.out.println("Speeding!");
        }
    }
}

class Car implements Vehicle {
    @Override
    public void start() {
        System.out.println("Car started");
    }

    @Override
    public void stop() {
        System.out.println("Car stopped");
    }

    // Can override default method if needed
    @Override
    public void honk() {
        System.out.println("Car horn: HONK!");
    }
}

// Multiple inheritance of behavior
interface Flyable {
    default void fly() {
        System.out.println("Flying...");
    }
}

interface Swimmable {
    default void swim() {
        System.out.println("Swimming...");
    }
}

class Duck implements Flyable, Swimmable {
    // Inherits both fly() and swim() methods
}
```

#### 6. **Optional Class**
Container object to represent potentially null values.

```java
// Creating Optional
Optional<String> optional = Optional.of("Hello");
Optional<String> empty = Optional.empty();
Optional<String> nullable = Optional.ofNullable(getString()); // may be null

// Checking and retrieving values
if (optional.isPresent()) {
    System.out.println(optional.get());
}

// Functional approach
optional.ifPresent(System.out::println);

// Default values
String value = optional.orElse("Default");
String value2 = optional.orElseGet(() -> computeDefault());
String value3 = optional.orElseThrow(() -> new RuntimeException("No value"));

// Transforming Optional values
Optional<Integer> length = optional.map(String::length);
Optional<String> upperCase = optional
    .filter(s -> s.length() > 3)
    .map(String::toUpperCase);

// FlatMap for nested Optionals
Optional<Person> person = findPerson(id);
Optional<String> email = person
    .flatMap(Person::getEmail)
    .filter(e -> e.contains("@"));
```

#### 7. **Date and Time API (JSR 310)**
New comprehensive date and time API replacing java.util.Date.

```java
// Current date and time
LocalDate today = LocalDate.now();
LocalTime now = LocalTime.now();
LocalDateTime dateTime = LocalDateTime.now();
ZonedDateTime zonedDateTime = ZonedDateTime.now();

// Creating specific dates
LocalDate birthday = LocalDate.of(1990, Month.JANUARY, 15);
LocalTime meetingTime = LocalTime.of(14, 30);
LocalDateTime appointment = LocalDateTime.of(2024, 12, 25, 10, 30);

// Parsing from strings
LocalDate parsed = LocalDate.parse("2024-12-25");
LocalTime parsedTime = LocalTime.parse("14:30:00");
LocalDateTime parsedDateTime = LocalDateTime.parse("2024-12-25T14:30:00");

// Date arithmetic
LocalDate nextWeek = today.plusWeeks(1);
LocalDate lastMonth = today.minusMonths(1);
LocalDateTime nextHour = dateTime.plusHours(1);

// Formatting
DateTimeFormatter formatter = DateTimeFormatter.ofPattern("dd/MM/yyyy");
String formatted = today.format(formatter);

// Time zones
ZoneId tokyo = ZoneId.of("Asia/Tokyo");
ZonedDateTime tokyoTime = ZonedDateTime.now(tokyo);
ZonedDateTime utcTime = tokyoTime.withZoneSameInstant(ZoneOffset.UTC);

// Period and Duration
Period period = Period.between(birthday, today);
Duration duration = Duration.between(now.minusHours(2), now);

// Working with different calendar systems
JapaneseDate japaneseDate = JapaneseDate.now();
HijrahDate hijrahDate = HijrahDate.now();
```

### 🛠 **Additional Major Improvements**

#### **Nashorn JavaScript Engine**
Replace Rhino with a more performant JavaScript engine.

```java
ScriptEngineManager manager = new ScriptEngineManager();
ScriptEngine nashorn = manager.getEngineByName("nashorn");

// Execute JavaScript
nashorn.eval("print('Hello from Nashorn')");

// Java-JavaScript interoperability
nashorn.put("javaObject", new MyJavaClass());
nashorn.eval("javaObject.someMethod()");

// Load JavaScript files
nashorn.eval(new FileReader("script.js"));
```

#### **Parallel Array Operations**
Parallel sorting and operations on arrays.

```java
int[] array = {5, 2, 8, 1, 9, 3};

// Parallel sort
Arrays.parallelSort(array);

// Parallel operations
Arrays.parallelSetAll(array, i -> i * 2);
Arrays.parallelPrefix(array, Integer::sum);

// Custom parallel operations
int[] result = Arrays.stream(array)
    .parallel()
    .map(x -> x * x)
    .toArray();
```

#### **CompletableFuture**
Enhanced asynchronous programming support.

```java
// Simple async operation
CompletableFuture<String> future = CompletableFuture.supplyAsync(() -> {
    // Simulate long-running task
    try { Thread.sleep(1000); } catch (InterruptedException e) {}
    return "Hello World";
});

// Chaining operations
CompletableFuture<String> result = future
    .thenApply(String::toUpperCase)
    .thenApply(s -> s + "!")
    .thenCompose(s -> CompletableFuture.supplyAsync(() -> s + " Completed"));

// Combining futures
CompletableFuture<Integer> future1 = CompletableFuture.supplyAsync(() -> 10);
CompletableFuture<Integer> future2 = CompletableFuture.supplyAsync(() -> 20);
CompletableFuture<Integer> combined = future1.thenCombine(future2, Integer::sum);

// Exception handling
CompletableFuture<String> withErrorHandling = CompletableFuture
    .supplyAsync(() -> {
        if (Math.random() > 0.5) throw new RuntimeException("Random error");
        return "Success";
    })
    .exceptionally(ex -> "Error: " + ex.getMessage())
    .thenApply(String::toUpperCase);
```

### 📈 **Performance and JVM Improvements**

#### **Memory Improvements**
- **Metaspace**: Replaced PermGen with native memory Metaspace
- **Compressed OOPs improvements**: Better memory efficiency
- **String deduplication**: Automatic deduplication of duplicate String objects

#### **JVM Enhancements**
```bash
# Remove PermGen, use Metaspace
java -XX:MetaspaceSize=128m -XX:MaxMetaspaceSize=512m MyApp

# Enable string deduplication
java -XX:+UseG1GC -XX:+UseStringDeduplication MyApp
```

---

## Java 9 (September 2017) - Jigsaw

**Major Theme**: Modular system (Project Jigsaw) and platform modularity

### 🔥 Major Features

#### 1. **Java Platform Module System (JPMS)**
The most significant architectural change in Java's history.

```java
// module-info.java
module com.example.myapp {
    requires java.base;        // Implicit, but can be explicit
    requires java.sql;         // Requires SQL module
    requires transitive java.logging;  // Transitive dependency

    exports com.example.myapp.api;     // Export package to other modules
    exports com.example.myapp.internal to com.example.trusted; // Qualified export

    opens com.example.myapp.model;     // Open for reflection (all modules)
    opens com.example.myapp.config to com.fasterxml.jackson.databind; // Qualified opens

    uses com.example.myapp.spi.Service;        // Service consumer
    provides com.example.myapp.spi.Service     // Service provider
        with com.example.myapp.impl.ServiceImpl;
}
```

#### **Module System Benefits**
```java
// Before modules - everything accessible
// com/example/internal/PrivateUtil.java - but still accessible!
public class PrivateUtil {
    public static void sensitiveMethod() {
        // Should be internal only
    }
}

// With modules - strong encapsulation
// Only exported packages are accessible
// Internal packages are truly internal

// Reliable configuration - no more classpath hell
// Explicit dependencies prevent missing JARs at runtime
```

#### **Module Commands**
```bash
# Compile modular application
javac -d out --module-path libs src/module-info.java src/com/example/**/*.java

# Run modular application
java --module-path out --module com.example.myapp/com.example.myapp.Main

# List modules
java --list-modules

# Describe module
java --describe-module java.base

# Generate module graph
jdeps --generate-module-info . myapp.jar
```

#### 2. **JShell - Interactive REPL**
Java finally gets a Read-Eval-Print Loop for experimentation.

```bash
# Start JShell
jshell

# Basic usage
jshell> 2 + 2
$1 ==> 4

jshell> String name = "Java"
name ==> "Java"

jshell> System.out.println("Hello " + name)
Hello Java

# Define methods
jshell> int factorial(int n) {
   ...>     return n <= 1 ? 1 : n * factorial(n-1);
   ...> }
|  created method factorial(int)

jshell> factorial(5)
$4 ==> 120

# Load external JARs
jshell> /env --class-path mylib.jar
jshell> import com.example.MyClass

# Save and load sessions
jshell> /save session.jsh
jshell> /open session.jsh
```

#### 3. **Collection Factory Methods**
Convenient methods to create immutable collections.

```java
// Before Java 9 - verbose immutable collections
List<String> list = Collections.unmodifiableList(
    Arrays.asList("a", "b", "c")
);
Set<String> set = Collections.unmodifiableSet(
    new HashSet<>(Arrays.asList("x", "y", "z"))
);
Map<String, Integer> map = Collections.unmodifiableMap(
    new HashMap<String, Integer>() {{
        put("one", 1);
        put("two", 2);
    }}
);

// Java 9+ - Factory methods
List<String> list = List.of("a", "b", "c");
Set<String> set = Set.of("x", "y", "z");
Map<String, Integer> map = Map.of(
    "one", 1,
    "two", 2,
    "three", 3
);

// Empty collections
List<String> empty = List.of();
Set<Integer> emptySet = Set.of();
Map<String, String> emptyMap = Map.of();

// Important: These collections are immutable!
// list.add("d"); // Throws UnsupportedOperationException
```

#### 4. **Stream API Enhancements**
New methods to make streams more powerful.

```java
// takeWhile - take elements while condition is true
Stream.of(1, 2, 3, 4, 5, 6)
    .takeWhile(x -> x < 4)
    .forEach(System.out::println); // 1, 2, 3

// dropWhile - drop elements while condition is true
Stream.of(1, 2, 3, 4, 5, 6)
    .dropWhile(x -> x < 4)
    .forEach(System.out::println); // 4, 5, 6

// ofNullable - create stream from potentially null value
String value = getValue(); // might return null
Stream<String> stream = Stream.ofNullable(value);

// iterate with predicate
Stream.iterate(1, n -> n < 100, n -> n * 2)
    .forEach(System.out::println); // 1, 2, 4, 8, 16, 32, 64

// Collectors enhancements
Map<Boolean, List<String>> partitioned = names.stream()
    .collect(Collectors.partitioningBy(name -> name.length() > 5));

// Filtering collector
List<String> longNames = names.stream()
    .collect(Collectors.filtering(
        name -> name.length() > 5,
        Collectors.toList()
    ));

// FlatMapping collector
List<Character> chars = words.stream()
    .collect(Collectors.flatMapping(
        word -> word.chars().mapToObj(c -> (char) c),
        Collectors.toList()
    ));
```

#### 5. **Optional Enhancements**
More functional methods for Optional.

```java
Optional<String> optional = Optional.of("Hello");

// ifPresentOrElse - action if present, another if empty
optional.ifPresentOrElse(
    System.out::println,
    () -> System.out.println("Empty")
);

// or - provide alternative Optional if current is empty
Optional<String> result = Optional.empty()
    .or(() -> Optional.of("Default"));

// stream - convert Optional to Stream
Stream<String> stream = optional.stream();

// Chain of Optional operations
String result = findUser(id)
    .flatMap(User::getAddress)
    .flatMap(Address::getCity)
    .orElse("Unknown City");
```

#### 6. **Process API Improvements**
Better process management and information.

```java
// Get current process info
ProcessHandle current = ProcessHandle.current();
System.out.println("PID: " + current.pid());
System.out.println("Info: " + current.info());

// List all processes
ProcessHandle.allProcesses()
    .filter(p -> p.info().command().isPresent())
    .limit(10)
    .forEach(p -> System.out.println(
        p.pid() + " -> " + p.info().command().orElse("Unknown")
    ));

// Start and manage processes
ProcessBuilder pb = new ProcessBuilder("sleep", "10");
Process process = pb.start();
ProcessHandle handle = process.toHandle();

// Monitor process completion
handle.onExit().thenAccept(p -> {
    System.out.println("Process " + p.pid() + " finished");
});

// Destroy process tree
handle.destroyForcibly();
```

### 🛠 **Additional Improvements**

#### **HTTP/2 Client (Incubator)**
New HTTP client with HTTP/2 support.

```java
// Note: This was incubator in Java 9, finalized in Java 11
HttpClient client = HttpClient.newBuilder()
    .version(HttpClient.Version.HTTP_2)
    .build();

HttpRequest request = HttpRequest.newBuilder()
    .uri(URI.create("https://api.example.com/data"))
    .header("Content-Type", "application/json")
    .build();

HttpResponse<String> response = client.send(request,
    HttpResponse.BodyHandlers.ofString());
```

#### **Reactive Streams**
Java 9 includes java.util.concurrent.Flow for reactive programming.

```java
// Publisher-Subscriber pattern
public class SimplePublisher implements Flow.Publisher<Integer> {
    @Override
    public void subscribe(Flow.Subscriber<? super Integer> subscriber) {
        subscriber.onSubscribe(new Flow.Subscription() {
            private int count = 0;

            @Override
            public void request(long n) {
                for (int i = 0; i < n && count < 10; i++) {
                    subscriber.onNext(++count);
                }
                if (count >= 10) {
                    subscriber.onComplete();
                }
            }

            @Override
            public void cancel() {}
        });
    }
}
```

#### **Multi-Resolution Images**
Support for HiDPI displays.

```java
MultiResolutionImage mrImage = Toolkit.getDefaultToolkit()
    .getImage("image.png");

Image variant = mrImage.getResolutionVariant(64, 64);
List<Image> variants = mrImage.getResolutionVariants();
```

### 📦 **JVM and Runtime Improvements**

#### **Compact Strings**
Automatic optimization for Latin-1 strings to use byte arrays instead of char arrays.

```bash
# Enabled by default, but can be controlled
java -XX:-CompactStrings MyApp  # Disable
java -XX:+CompactStrings MyApp  # Enable (default)
```

#### **G1 as Default GC**
G1 Garbage Collector becomes the default for server-class machines.

```bash
# G1 is now default, but can specify others
java -XX:+UseParallelGC MyApp
java -XX:+UseConcMarkSweepGC MyApp
```

#### **Unified Logging**
New logging framework for JVM diagnostics.

```bash
# New logging syntax
java -Xlog:gc:gc.log MyApp
java -Xlog:gc*:gc.log:time,level,tags MyApp
java -Xlog:all:all.log MyApp
```

---

## Java 10 (March 2018)

**Major Theme**: Local variable type inference and runtime improvements

### 🔥 Major Features

#### 1. **Local Variable Type Inference (var)**
Type inference for local variables using the `var` keyword.

```java
// Before Java 10
String message = "Hello World";
List<String> names = new ArrayList<>();
Map<String, Integer> scores = new HashMap<>();

// Java 10+ with var
var message = "Hello World";        // String
var names = new ArrayList<String>(); // ArrayList<String>
var scores = new HashMap<String, Integer>(); // HashMap<String, Integer>

// Complex types made simpler
var result = database.query("SELECT * FROM users")
    .stream()
    .filter(row -> row.getAge() > 18)
    .collect(Collectors.toMap(
        row -> row.getId(),
        row -> row.getName()
    )); // Map<Integer, String>

// With generic methods
var optional = Optional.of("Hello");
var stream = Stream.of(1, 2, 3, 4, 5);
var future = CompletableFuture.supplyAsync(() -> "Async Result");
```

#### **var Guidelines and Limitations**
```java
// Good uses of var - improves readability
var users = getUsersFromDatabase();
var config = ConfigurationLoader.load("app.properties");
var json = objectMapper.readValue(jsonString, UserProfile.class);

// Less ideal uses - reduces readability
var data = getData();  // What type is data?
var x = calculate();   // What does calculate() return?

// var limitations - cannot use in these contexts:
// - Method parameters: void method(var param) {} // ❌
// - Method return type: var getUser() {} // ❌
// - Class fields: private var field; // ❌
// - Array initializers: var array = {1, 2, 3}; // ❌
// - Lambda parameters (Java 10): (var x) -> x + 1 // ❌ (allowed in Java 11)

// Must be initialized
var uninitalized; // ❌ Compilation error
var nullValue = null; // ❌ Cannot infer type from null

// Proper initialization required
var list = new ArrayList<>(); // ❌ Cannot infer generic type
var list = new ArrayList<String>(); // ✅ Correct
```

#### 2. **Unmodifiable Collection Copyof Methods**
Create immutable copies of collections.

```java
List<String> original = new ArrayList<>();
original.add("a");
original.add("b");
original.add("c");

// Create immutable copy
List<String> immutableCopy = List.copyOf(original);
Set<String> immutableSet = Set.copyOf(original);
Map<String, String> immutableMap = Map.copyOf(originalMap);

// Attempting to modify throws exception
// immutableCopy.add("d"); // UnsupportedOperationException

// copyOf vs of() behavior
List<String> list1 = List.of("a", "b", "c");        // Always creates new
List<String> list2 = List.copyOf(list1);            // Returns same instance if already immutable
List<String> list3 = List.copyOf(original);         // Creates new immutable copy

// Useful for defensive copying
public class UserService {
    private final List<String> allowedRoles;

    public UserService(List<String> roles) {
        this.allowedRoles = List.copyOf(roles); // Defensive copy
    }

    public List<String> getAllowedRoles() {
        return allowedRoles; // Safe to return - immutable
    }
}
```

### 🛠 **Performance and JVM Improvements**

#### **Application Class-Data Sharing (AppCDS)**
Improve startup time and memory footprint.

```bash
# Create class list during application run
java -XX:+UseAppCDS -XX:DumpLoadedClassList=classes.lst MyApp

# Create shared archive
java -XX:+UseAppCDS -Xshare:dump -XX:SharedClassListFile=classes.lst \
     -XX:SharedArchiveFile=app.jsa -cp myapp.jar

# Use shared archive
java -XX:+UseAppCDS -Xshare:on -XX:SharedArchiveFile=app.jsa \
     -cp myapp.jar MyApp
```

#### **Parallel Full GC for G1**
G1 Garbage Collector now performs full GC in parallel.

```bash
# G1 parallel full GC is now default
java -XX:+UseG1GC MyApp

# Monitor full GC improvements
java -XX:+UseG1GC -Xlog:gc*:gc.log MyApp
```

#### **Experimental Graal JIT Compiler**
New JIT compiler for improved performance.

```bash
# Enable experimental Graal compiler
java -XX:+UnlockExperimentalVMOptions -XX:+UseJVMCICompiler MyApp

# With specific Graal configurations
java -XX:+UnlockExperimentalVMOptions -XX:+UseJVMCICompiler \
     -XX:+EagerJVMCI MyApp
```

#### **Root Certificates**
Default set of CA certificates included with JDK.

```java
// No longer need to manually configure CA certificates
SSLContext context = SSLContext.getDefault();
HttpsURLConnection connection = (HttpsURLConnection)
    new URL("https://example.com").openConnection();
```

### 🧪 **Additional Improvements**

#### **Thread-Local Handshakes**
Individual thread operations without global VM safepoints.

```bash
# Better performance for operations like:
# - Thread.getAllStackTraces()
# - Thread state changes
# - Some garbage collection operations
```

#### **Heap Allocation on Alternative Memory Devices**
Use non-DRAM memory for heap storage.

```bash
# Allocate heap on non-DRAM memory (like NVDIMMs)
java -XX:AllocateHeapAt=/path/to/nvdimm MyApp
```

#### **Time-Based Release Versioning**
New version numbering scheme: YEAR.MONTH.

```bash
# Java version output changed
$ java -version
java version "10.0.1" 2018-04-17
Java(TM) SE Runtime Environment 18.3 (build 10.0.1+10)
Java HotSpot(TM) 64-Bit Server VM 18.3 (build 10.0.1+10, mixed mode)
```

---

## Java 11 (September 2018) - LTS

**Major Theme**: Long-term support release with HTTP client and string enhancements

### 🔥 Major Features

#### 1. **HTTP Client (Standard)**
The HTTP client introduced as incubator in Java 9 is now standard.

```java
// Basic HTTP GET request
HttpClient client = HttpClient.newHttpClient();
HttpRequest request = HttpRequest.newBuilder()
    .uri(URI.create("https://api.example.com/users"))
    .build();

HttpResponse<String> response = client.send(request,
    HttpResponse.BodyHandlers.ofString());

System.out.println("Status: " + response.statusCode());
System.out.println("Body: " + response.body());

// HTTP POST with JSON
String json = "{\"name\":\"John\",\"age\":30}";
HttpRequest postRequest = HttpRequest.newBuilder()
    .uri(URI.create("https://api.example.com/users"))
    .header("Content-Type", "application/json")
    .POST(HttpRequest.BodyPublishers.ofString(json))
    .build();

HttpResponse<String> postResponse = client.send(postRequest,
    HttpResponse.BodyHandlers.ofString());

// Asynchronous requests
CompletableFuture<HttpResponse<String>> futureResponse = client
    .sendAsync(request, HttpResponse.BodyHandlers.ofString());

futureResponse.thenAccept(resp -> {
    System.out.println("Async response: " + resp.body());
});

// HTTP/2 configuration
HttpClient http2Client = HttpClient.newBuilder()
    .version(HttpClient.Version.HTTP_2)
    .connectTimeout(Duration.ofSeconds(30))
    .build();

// Custom configuration
HttpClient customClient = HttpClient.newBuilder()
    .authenticator(new Authenticator() {
        @Override
        protected PasswordAuthentication getPasswordAuthentication() {
            return new PasswordAuthentication("user", "pass".toCharArray());
        }
    })
    .proxy(ProxySelector.of(new InetSocketAddress("proxy.example.com", 8080)))
    .followRedirects(HttpClient.Redirect.NORMAL)
    .build();
```

#### 2. **String Methods Enhancements**
New utility methods for String manipulation.

```java
// isBlank() - checks if string is empty or contains only whitespace
String empty = "";
String whitespace = "   \t\n  ";
String content = "Hello";

System.out.println(empty.isBlank());      // true
System.out.println(whitespace.isBlank()); // true
System.out.println(content.isBlank());    // false

// lines() - returns stream of lines
String multiline = "Line 1\nLine 2\r\nLine 3\n";
multiline.lines()
    .map(String::toUpperCase)
    .forEach(System.out::println);

// repeat() - repeat string n times
String pattern = "=".repeat(50);
System.out.println(pattern); // ==================================================

String greeting = "Hello! ".repeat(3);
System.out.println(greeting); // Hello! Hello! Hello!

// strip(), stripLeading(), stripTrailing() - Unicode-aware trimming
String unicode = "\u2000\u2001 Hello World \u2000\u2001";
System.out.println(unicode.trim());         // Still has Unicode spaces
System.out.println(unicode.strip());        // Properly removes Unicode spaces
System.out.println(unicode.stripLeading()); // Remove leading spaces only
System.out.println(unicode.stripTrailing()); // Remove trailing spaces only

// Practical examples
public class StringUtils {
    public static void printSection(String title) {
        String divider = "=".repeat(title.length() + 4);
        System.out.println(divider);
        System.out.println("| " + title + " |");
        System.out.println(divider);
    }

    public static List<String> parseLines(String input) {
        return input.lines()
            .map(String::strip)
            .filter(line -> !line.isBlank())
            .collect(Collectors.toList());
    }
}
```

#### 3. **File Methods**
New convenience methods for file operations.

```java
// readString() and writeString()
Path filePath = Paths.get("example.txt");

// Write string to file
String content = "Hello World\nThis is Java 11";
Files.writeString(filePath, content);

// Read string from file
String readContent = Files.readString(filePath);
System.out.println(readContent);

// With specific encoding
Files.writeString(filePath, content, StandardCharsets.UTF_8);
String readContentUtf8 = Files.readString(filePath, StandardCharsets.UTF_8);

// Append to file
Files.writeString(filePath, "\nAppended line", StandardOpenOption.APPEND);

// Advanced file operations
public class FileProcessor {
    public void processConfigFile(Path configPath) throws IOException {
        String config = Files.readString(configPath);

        // Process each line
        String processed = config.lines()
            .map(String::strip)
            .filter(line -> !line.isBlank() && !line.startsWith("#"))
            .map(line -> processConfigLine(line))
            .collect(Collectors.joining("\n"));

        // Write back processed config
        Files.writeString(configPath, processed);
    }

    private String processConfigLine(String line) {
        // Process individual config line
        return line.toLowerCase();
    }
}
```

#### 4. **var in Lambda Parameters**
Use var for lambda parameter types.

```java
// Java 11 allows var in lambda parameters
List<String> names = Arrays.asList("Alice", "Bob", "Charlie");

// Traditional lambda
names.forEach((String name) -> System.out.println(name));

// Java 11 - var in lambda parameters
names.forEach((var name) -> System.out.println(name));

// Useful when you need annotations
names.forEach((@NonNull var name) -> {
    System.out.println(name.toUpperCase());
});

// Multiple parameters
Map<String, Integer> scores = Map.of("Alice", 95, "Bob", 87);
scores.forEach((var name, var score) -> {
    System.out.println(name + ": " + score);
});

// Complex lambda with var
Stream.of("apple", "banana", "cherry")
    .collect(Collectors.toMap(
        (var fruit) -> fruit.charAt(0),
        (var fruit) -> fruit.length(),
        (var existing, var replacement) -> existing + replacement
    ));
```

#### 5. **Nest-Based Access Control**
Better support for nested classes and reflection.

```java
public class OuterClass {
    private String outerField = "outer";

    public class InnerClass {
        private String innerField = "inner";

        public void accessOuter() {
            // Can access private members of outer class
            System.out.println(outerField);
        }
    }

    public void accessInner(InnerClass inner) {
        // Can access private members of inner class
        System.out.println(inner.innerField);
    }
}

// Reflection improvements
Class<?> outerClass = OuterClass.class;
Class<?> innerClass = OuterClass.InnerClass.class;

// Check nest relationships
System.out.println(outerClass.isNestHost());        // true
System.out.println(innerClass.isNestmate(outerClass)); // true
System.out.println(Arrays.toString(outerClass.getNestMembers()));
```

### 🛠 **Runtime and Performance Improvements**

#### **Epsilon Garbage Collector**
No-op garbage collector for performance testing.

```bash
# Use Epsilon GC (handles allocation but no collection)
java -XX:+UnlockExperimentalVMOptions -XX:+UseEpsilonGC MyApp

# Useful for:
# - Short-lived applications
# - Performance testing
# - Applications with known memory bounds
```

#### **Low-Latency Z Garbage Collector (Experimental)**
Experimental low-latency garbage collector.

```bash
# Enable experimental ZGC
java -XX:+UnlockExperimentalVMOptions -XX:+UseZGC MyApp

# ZGC benefits:
# - Sub-millisecond pause times
# - Concurrent collection
# - Scalable (8MB to 16TB heaps)
```

#### **Flight Recorder**
Low-overhead profiling now free and available in OpenJDK.

```bash
# Enable Flight Recorder
java -XX:+FlightRecorder -XX:StartFlightRecording=duration=60s,filename=profile.jfr MyApp

# Analyze with Mission Control or other tools
jfr print --events jdk.CPULoad profile.jfr
```

### 🔧 **Developer Productivity**

#### **Single-File Source-Code Programs**
Run Java programs directly without compilation.

```bash
# hello.java
#!/usr/bin/java --source 11
public class Hello {
    public static void main(String[] args) {
        System.out.println("Hello Java 11!");
    }
}

# Run directly
java hello.java

# Or make executable
chmod +x hello.java
./hello.java
```

#### **Dynamic Class-File Constants**
Support for dynamic constants in bytecode.

```java
// Enables more efficient implementation of:
// - String concatenation
// - Method handles
// - Enum and record generation
```

### 📦 **Removal and Deprecation**

#### **Removed Modules and APIs**
```java
// These were removed in Java 11:
// - Java EE modules (JAX-WS, JAXB, JAX-RS)
// - CORBA modules
// - Nashorn JavaScript engine (deprecated)

// Migration required for:
// javax.xml.bind.* (JAXB)
// javax.xml.ws.* (JAX-WS)
// javax.annotation.* (Common Annotations)
```

#### **Alternative Dependencies**
```xml
<!-- Add as external dependencies if needed -->
<dependency>
    <groupId>javax.xml.bind</groupId>
    <artifactId>jaxb-api</artifactId>
    <version>2.3.1</version>
</dependency>

<dependency>
    <groupId>com.sun.xml.bind</groupId>
    <artifactId>jaxb-runtime</artifactId>
    <version>2.3.1</version>
</dependency>
```

---

## Java 12 (March 2019)

**Major Theme**: Switch expressions preview and microbenchmarks

### 🔥 Major Features

#### 1. **Switch Expressions (Preview)**
Enhanced switch statements that can return values.

```java
// Traditional switch statement
String dayType;
switch (dayOfWeek) {
    case MONDAY:
    case TUESDAY:
    case WEDNESDAY:
    case THURSDAY:
    case FRIDAY:
        dayType = "Weekday";
        break;
    case SATURDAY:
    case SUNDAY:
        dayType = "Weekend";
        break;
    default:
        dayType = "Unknown";
}

// Java 12 switch expression (preview)
String dayType = switch (dayOfWeek) {
    case MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY -> "Weekday";
    case SATURDAY, SUNDAY -> "Weekend";
};

// Complex switch expressions
int numLetters = switch (dayOfWeek) {
    case MONDAY, FRIDAY, SUNDAY -> 6;
    case TUESDAY -> 7;
    case THURSDAY, SATURDAY -> 8;
    case WEDNESDAY -> 9;
};

// Switch with blocks
String description = switch (grade) {
    case A -> "Excellent";
    case B -> "Good";
    case C -> {
        System.out.println("Average grade");
        yield "Average";
    }
    default -> "Below Average";
};

// Enum switch expressions
public enum Status { ACTIVE, INACTIVE, PENDING }

String message = switch (userStatus) {
    case ACTIVE -> "User is currently active";
    case INACTIVE -> "User account is deactivated";
    case PENDING -> "User registration is pending approval";
};
```

### 🛠 **JVM Improvements**

#### **Shenandoah Garbage Collector (Experimental)**
Low-latency garbage collector alternative to ZGC.

```bash
# Enable Shenandoah GC
java -XX:+UnlockExperimentalVMOptions -XX:+UseShenandoahGC MyApp

# Shenandoah features:
# - Low pause times independent of heap size
# - Concurrent evacuation
# - Lower memory overhead than ZGC
```

#### **Microbenchmark Suite**
Built-in microbenchmarking with JMH (Java Microbenchmark Harness).

```java
// Example microbenchmark (requires JMH dependency)
@BenchmarkMode(Mode.AverageTime)
@OutputTimeUnit(TimeUnit.NANOSECONDS)
@State(Scope.Benchmark)
public class StringConcatenationBenchmark {

    @Benchmark
    public String stringConcatenation() {
        return "Hello" + " " + "World";
    }

    @Benchmark
    public String stringBuilder() {
        return new StringBuilder()
            .append("Hello")
            .append(" ")
            .append("World")
            .toString();
    }

    @Benchmark
    public String stringBuffer() {
        return new StringBuffer()
            .append("Hello")
            .append(" ")
            .append("World")
            .toString();
    }
}
```

#### **JVM Constants API**
New API for modeling nominal descriptors of class-file and run-time entities.

```java
// Constant descriptors for reflection
MethodType mt = MethodType.methodType(String.class, int.class);
MethodHandle mh = MethodHandles.lookup()
    .findVirtual(String.class, "substring", mt);

// Dynamic constants support
CallSite callSite = StringConcatFactory.makeConcatWithConstants(
    MethodHandles.lookup(),
    "concat",
    MethodType.methodType(String.class, String.class, int.class),
    "User: \u0001, Age: \u0001",
    new Object[0]
);
```

### 📊 **Collection and String Enhancements**

#### **Compact Number Formatting**
Format numbers in a compact, locale-sensitive manner.

```java
NumberFormat compactFormatter = NumberFormat.getCompactNumberInstance(
    Locale.US, NumberFormat.Style.SHORT);

System.out.println(compactFormatter.format(1000));     // 1K
System.out.println(compactFormatter.format(1500));     // 2K (rounded)
System.out.println(compactFormatter.format(1000000));  // 1M
System.out.println(compactFormatter.format(1234567));  // 1M

// Different locales
NumberFormat germanCompact = NumberFormat.getCompactNumberInstance(
    Locale.GERMANY, NumberFormat.Style.LONG);
System.out.println(germanCompact.format(1500));        // 1 Tausend

// Different styles
NumberFormat longStyle = NumberFormat.getCompactNumberInstance(
    Locale.US, NumberFormat.Style.LONG);
System.out.println(longStyle.format(1500));            // 2 thousand
System.out.println(longStyle.format(1000000));         // 1 million
```

#### **Teeing Collector**
Combine two collectors into one.

```java
List<String> names = List.of("Alice", "Bob", "Charlie", "David", "Eve");

// Collect both average length and concatenated string
record NameStats(double avgLength, String concatenated) {}

NameStats stats = names.stream()
    .collect(Collectors.teeing(
        Collectors.averagingInt(String::length),  // First collector
        Collectors.joining(", "),                 // Second collector
        (avg, concat) -> new NameStats(avg, concat) // Combiner
    ));

System.out.println("Average length: " + stats.avgLength);
System.out.println("All names: " + stats.concatenated);

// Practical example - statistics calculation
record SalesStats(double total, long count) {}

SalesStats salesStats = salesData.stream()
    .collect(Collectors.teeing(
        Collectors.summingDouble(Sale::getAmount),
        Collectors.counting(),
        (sum, count) -> new SalesStats(sum, count)
    ));
```

---

## Java 13 (September 2019)

**Major Theme**: Text blocks preview and switch expression improvements

### 🔥 Major Features

#### 1. **Text Blocks (Preview)**
Multi-line string literals with proper formatting.

```java
// Before text blocks - concatenation nightmare
String json = "{\n" +
              "  \"name\": \"John Doe\",\n" +
              "  \"age\": 30,\n" +
              "  \"address\": {\n" +
              "    \"street\": \"123 Main St\",\n" +
              "    \"city\": \"Anytown\"\n" +
              "  }\n" +
              "}";

// Java 13+ text blocks (preview)
String json = """
    {
      "name": "John Doe",
      "age": 30,
      "address": {
        "street": "123 Main St",
        "city": "Anytown"
      }
    }
    """;

// HTML templates
String html = """
    <!DOCTYPE html>
    <html>
    <head>
        <title>My Page</title>
    </head>
    <body>
        <h1>Welcome to My Site</h1>
        <p>This is a paragraph.</p>
    </body>
    </html>
    """;

// SQL queries
String query = """
    SELECT u.name, u.email, p.title
    FROM users u
    JOIN posts p ON u.id = p.user_id
    WHERE u.created_date > ?
    ORDER BY u.name
    """;

// Indentation handling
String code = """
        public class Example {
            public void method() {
                System.out.println("Hello");
            }
        }
        """; // Common indentation is automatically removed

// Empty lines and formatting
String poem = """
    Roses are red,
    Violets are blue,

    Java 13 is here,
    Text blocks are new!
    """;
```

#### 2. **Switch Expressions (Second Preview)**
Refined switch expressions with yield keyword.

```java
// Switch expressions with yield
String result = switch (grade) {
    case A -> "Excellent";
    case B -> "Good";
    case C -> {
        System.out.println("Could be better");
        yield "Average"; // yield replaces return in switch expressions
    }
    case D, F -> {
        System.out.println("Needs improvement");
        yield "Poor";
    }
};

// Mixed arrow and colon syntax
int daysInMonth = switch (month) {
    case FEBRUARY -> {
        if (isLeapYear) {
            yield 29;
        } else {
            yield 28;
        }
    }
    case APRIL, JUNE, SEPTEMBER, NOVEMBER -> 30;
    default: {
        // Traditional block with yield
        System.out.println("Standard month");
        yield 31;
    }
};

// Complex business logic
Money calculateDiscount(Customer customer, Product product) {
    return switch (customer.getTier()) {
        case GOLD -> {
            double discount = switch (product.getCategory()) {
                case ELECTRONICS -> 0.15;
                case CLOTHING -> 0.20;
                case BOOKS -> 0.10;
                default -> 0.05;
            };
            yield product.getPrice().multiply(discount);
        }
        case SILVER -> product.getPrice().multiply(0.10);
        case BRONZE -> product.getPrice().multiply(0.05);
        case REGULAR -> Money.ZERO;
    };
}
```

### 🛠 **Runtime and Performance**

#### **Dynamic CDS Archives**
Extend Class Data Sharing to application classes.

```bash
# Create CDS archive during application run
java -XX:ArchiveClassesAtExit=app.jsa -cp myapp.jar MyApp

# Use CDS archive
java -XX:SharedArchiveFile=app.jsa -cp myapp.jar MyApp

# Benefits:
# - Faster startup time
# - Reduced memory footprint
# - Better performance for short-lived applications
```

#### **ZGC: Uncommit Unused Memory**
ZGC can now return unused heap memory to the operating system.

```bash
# Enable ZGC with memory uncommit
java -XX:+UnlockExperimentalVMOptions -XX:+UseZGC
     -XX:+UncommitUnusedPages MyApp

# Monitor memory usage
java -XX:+UnlockExperimentalVMOptions -XX:+UseZGC
     -Xlog:gc*:gc.log MyApp
```

### 🔧 **Socket API Reimplementation**

#### **NIO-based Socket Implementation**
New implementation for better maintainability and performance.

```java
// Enhanced socket performance (transparent to applications)
ServerSocket serverSocket = new ServerSocket(8080);
Socket clientSocket = serverSocket.accept();

// Better integration with NIO
SocketChannel socketChannel = SocketChannel.open();
socketChannel.connect(new InetSocketAddress("localhost", 8080));

// Improved debugging and monitoring
// -Djdk.net.usePlainSocketImpl=true to use old implementation if needed
```

---

## Java 14 (March 2020)

**Major Theme**: Records preview, pattern matching, and helpful NullPointerExceptions

### 🔥 Major Features

#### 1. **Records (Preview)**
Concise syntax for immutable data classes.

```java
// Traditional Java class for data
public final class PersonOld {
    private final String name;
    private final int age;
    private final String email;

    public PersonOld(String name, int age, String email) {
        this.name = name;
        this.age = age;
        this.email = email;
    }

    public String getName() { return name; }
    public int getAge() { return age; }
    public String getEmail() { return email; }

    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (obj == null || getClass() != obj.getClass()) return false;
        PersonOld person = (PersonOld) obj;
        return age == person.age &&
               Objects.equals(name, person.name) &&
               Objects.equals(email, person.email);
    }

    @Override
    public int hashCode() {
        return Objects.hash(name, age, email);
    }

    @Override
    public String toString() {
        return "PersonOld{name='" + name + "', age=" + age + ", email='" + email + "'}";
    }
}

// Java 14+ Record (preview)
public record Person(String name, int age, String email) {}

// Usage
Person person = new Person("Alice", 30, "alice@example.com");
System.out.println(person.name());    // Alice
System.out.println(person.age());     // 30
System.out.println(person.email());   // alice@example.com
System.out.println(person.toString()); // Person[name=Alice, age=30, email=alice@example.com]

// Records with validation
public record ValidatedPerson(String name, int age, String email) {
    public ValidatedPerson {
        if (age < 0) {
            throw new IllegalArgumentException("Age cannot be negative");
        }
        if (name == null || name.trim().isEmpty()) {
            throw new IllegalArgumentException("Name cannot be null or empty");
        }
        // Normalize name
        name = name.trim();
    }

    // Additional methods
    public boolean isAdult() {
        return age >= 18;
    }

    public String getDomain() {
        return email.substring(email.indexOf('@') + 1);
    }
}

// Nested records
public record Address(String street, String city, String country) {}
public record PersonWithAddress(String name, int age, Address address) {}

// Records in collections
List<Person> people = List.of(
    new Person("Alice", 30, "alice@example.com"),
    new Person("Bob", 25, "bob@example.com")
);

Map<String, Person> personMap = people.stream()
    .collect(Collectors.toMap(Person::email, Function.identity()));
```

#### 2. **Pattern Matching for instanceof (Preview)**
Eliminate explicit casting after instanceof checks.

```java
// Before Java 14 - traditional instanceof
public String describeAnimal(Object animal) {
    if (animal instanceof Dog) {
        Dog dog = (Dog) animal; // Explicit cast
        return "Dog named " + dog.getName() + " of breed " + dog.getBreed();
    } else if (animal instanceof Cat) {
        Cat cat = (Cat) animal; // Explicit cast
        return "Cat named " + cat.getName() + " with " + cat.getLives() + " lives";
    } else if (animal instanceof Bird) {
        Bird bird = (Bird) animal; // Explicit cast
        return "Bird named " + bird.getName() + " that can " +
               (bird.canFly() ? "fly" : "not fly");
    } else {
        return "Unknown animal";
    }
}

// Java 14+ Pattern matching (preview)
public String describeAnimal(Object animal) {
    if (animal instanceof Dog dog) {
        return "Dog named " + dog.getName() + " of breed " + dog.getBreed();
    } else if (animal instanceof Cat cat) {
        return "Cat named " + cat.getName() + " with " + cat.getLives() + " lives";
    } else if (animal instanceof Bird bird) {
        return "Bird named " + bird.getName() + " that can " +
               (bird.canFly() ? "fly" : "not fly");
    } else {
        return "Unknown animal";
    }
}

// Complex pattern matching
public double calculateArea(Shape shape) {
    if (shape instanceof Rectangle rect && rect.getWidth() > 0 && rect.getHeight() > 0) {
        return rect.getWidth() * rect.getHeight();
    } else if (shape instanceof Circle circle && circle.getRadius() > 0) {
        return Math.PI * circle.getRadius() * circle.getRadius();
    } else if (shape instanceof Triangle triangle) {
        return 0.5 * triangle.getBase() * triangle.getHeight();
    }
    return 0;
}

// Pattern matching with guards
public String processRequest(Object request) {
    if (request instanceof HttpRequest httpReq && httpReq.getMethod().equals("GET")) {
        return "Processing GET request to " + httpReq.getUrl();
    } else if (request instanceof HttpRequest httpReq && httpReq.getMethod().equals("POST")) {
        return "Processing POST request with " + httpReq.getBody().length() + " bytes";
    }
    return "Unknown request type";
}
```

#### 3. **Helpful NullPointerExceptions**
Detailed NPE messages showing which part of an expression was null.

```java
// Before Java 14 - Generic NPE message
person.getAddress().getStreet().toLowerCase();
// Exception: java.lang.NullPointerException

// Java 14+ - Detailed NPE message
// Exception: java.lang.NullPointerException:
//   Cannot invoke "String.toLowerCase()" because the return value of
//   "com.example.Address.getStreet()" is null

// Complex expression with detailed NPE
users.get(userId).getProfile().getPreferences().getTheme().getColor();
// Exception shows exactly which method returned null:
// Cannot invoke "com.example.Theme.getColor()" because the return value of
// "com.example.Preferences.getTheme()" is null

// Enable detailed NPE messages (enabled by default in Java 14+)
// -XX:+ShowCodeDetailsInExceptionMessages

// Example troubleshooting
public void processUserData(Map<String, User> users, String userId) {
    // This will show exactly what's null
    String email = users.get(userId).getContact().getEmail().toLowerCase();

    // Possible detailed error messages:
    // - Cannot invoke "User.getContact()" because the return value of "Map.get(Object)" is null
    // - Cannot invoke "Contact.getEmail()" because the return value of "User.getContact()" is null
    // - Cannot invoke "String.toLowerCase()" because the return value of "Contact.getEmail()" is null
}
```

#### 4. **Text Blocks (Second Preview)**
Enhanced text blocks with new escape sequences.

```java
// New escape sequences in text blocks
String textWithEscapes = """
    Line 1
    Line 2 \
    continues on same line
    Line 3 with explicit newline\n
    Line 4 with space at end \s
    Line 5 normal
    """;

// \<line-terminator> - line continuation
String singleLine = """
    This is a very long line that \
    spans multiple lines in source \
    but appears as single line
    """;

// \s - explicit space (preserves trailing spaces)
String withSpaces = """
    Name:     John Doe    \s
    Age:      30         \s
    Location: New York   \s
    """;

// Practical examples
String sqlQuery = """
    SELECT
        u.id,
        u.name,
        u.email,
        p.title
    FROM users u \
    LEFT JOIN posts p ON u.id = p.user_id \
    WHERE u.active = true \
    ORDER BY u.name
    """;

String jsonTemplate = """
    {
        "user": {
            "name": "%s",
            "age": %d,
            "preferences": {
                "theme": "dark",
                "notifications": true\s
            }
        }
    }
    """.formatted("John Doe", 30);
```

### 🛠 **JVM and Performance**

#### **Foreign Memory Access API (Incubator)**
Safe and efficient access to off-heap memory.

```java
// Access off-heap memory (incubator module)
// Requires --add-modules jdk.incubator.foreign

import jdk.incubator.foreign.*;

// Allocate off-heap memory
try (MemorySegment segment = MemorySegment.allocateNative(100)) {
    MemoryAddress address = segment.baseAddress();

    // Write data
    MemoryAccess.setIntAtOffset(segment, 0, 42);
    MemoryAccess.setIntAtOffset(segment, 4, 84);

    // Read data
    int value1 = MemoryAccess.getIntAtOffset(segment, 0); // 42
    int value2 = MemoryAccess.getIntAtOffset(segment, 4); // 84
}
```

#### **JFR Event Streaming**
Consume JFR events in real-time.

```java
// Stream JFR events in real-time
try (RecordingStream rs = new RecordingStream()) {
    rs.enable("jdk.CPULoad").withPeriod(Duration.ofSeconds(1));
    rs.enable("jdk.JavaMonitorEnter").withThreshold(Duration.ofMillis(10));

    rs.onEvent("jdk.CPULoad", event -> {
        double cpuLoad = event.getDouble("machineTotal");
        System.out.println("CPU Load: " + cpuLoad);
    });

    rs.onEvent("jdk.JavaMonitorEnter", event -> {
        String monitorClass = event.getString("monitorClass");
        System.out.println("Monitor contention on: " + monitorClass);
    });

    rs.start(); // Start consuming events
}
```

#### **Non-Volatile Mapped Byte Buffers**
Support for non-volatile memory (NVM) through MappedByteBuffer.

```java
// Map file to non-volatile memory
try (RandomAccessFile file = new RandomAccessFile("data.nvram", "rw");
     FileChannel channel = file.getChannel()) {

    MappedByteBuffer buffer = channel.map(
        FileChannel.MapMode.READ_WRITE, 0, file.length());

    // Force changes to non-volatile storage
    buffer.force();
}
```

### 🧪 **Packaging and Tools**

#### **Packaging Tool (Incubator)**
Create native installers for Java applications.

```bash
# Create native installer (incubator feature)
jpackage --input target/libs \
         --main-jar myapp.jar \
         --main-class com.example.Main \
         --name MyApplication \
         --app-version 1.0 \
         --type pkg \
         --dest output

# Platform-specific installers
# Windows: exe, msi
# macOS: pkg, dmg
# Linux: deb, rpm

# Advanced packaging
jpackage --input libs \
         --main-jar app.jar \
         --name "My Java App" \
         --app-version "2.1.0" \
         --vendor "My Company" \
         --description "Description of my application" \
         --copyright "Copyright 2020 My Company" \
         --license-file LICENSE.txt \
         --icon app-icon.png \
         --java-options "-Xmx2g -Dfile.encoding=UTF-8"
```

---

## Java 15 (September 2020)

**Major Theme**: Text blocks finalized, sealed classes preview, and hidden classes

### 🔥 Major Features

#### 1. **Text Blocks (Final)**
Text blocks are now a permanent feature with final syntax.

```java
// Text blocks are now standard (no longer preview)
String json = """
    {
        "users": [
            {
                "id": 1,
                "name": "Alice Johnson",
                "email": "alice@example.com",
                "roles": ["admin", "user"]
            },
            {
                "id": 2,
                "name": "Bob Smith",
                "email": "bob@example.com",
                "roles": ["user"]
            }
        ],
        "meta": {
            "total": 2,
            "page": 1
        }
    }
    """;

// Advanced formatting with text blocks
public class QueryBuilder {
    public String buildComplexQuery(String table, List<String> columns,
                                  Map<String, Object> filters, String orderBy) {
        String columnList = String.join(", ", columns);

        String whereClause = filters.entrySet().stream()
            .map(entry -> entry.getKey() + " = ?")
            .collect(Collectors.joining(" AND "));

        return """
            SELECT %s
            FROM %s
            WHERE %s
            ORDER BY %s
            LIMIT 100
            """.formatted(columnList, table, whereClause, orderBy);
    }
}

// Configuration files
String dockerFile = """
    FROM openjdk:15-jdk-alpine

    WORKDIR /app

    COPY target/*.jar app.jar

    EXPOSE 8080

    ENTRYPOINT ["java", "-jar", "app.jar"]
    """;

// Code generation
public String generateJavaClass(String className, List<String> fields) {
    String fieldDeclarations = fields.stream()
        .map(field -> "    private String " + field + ";")
        .collect(Collectors.joining("\n"));

    String getters = fields.stream()
        .map(field -> """
                public String get%s() {
                    return %s;
                }
                """.formatted(capitalize(field), field))
        .collect(Collectors.joining("\n"));

    return """
        public class %s {
        %s

        %s
        }
        """.formatted(className, fieldDeclarations, getters);
}
```

#### 2. **Sealed Classes (Preview)**
Control which classes can extend or implement a class/interface.

```java
// Sealed class definition
public sealed class Shape
    permits Circle, Rectangle, Triangle {

    protected final String color;

    protected Shape(String color) {
        this.color = color;
    }

    public abstract double area();
}

// Permitted subclasses
public final class Circle extends Shape {
    private final double radius;

    public Circle(String color, double radius) {
        super(color);
        this.radius = radius;
    }

    @Override
    public double area() {
        return Math.PI * radius * radius;
    }
}

public final class Rectangle extends Shape {
    private final double width, height;

    public Rectangle(String color, double width, double height) {
        super(color);
        this.width = width;
        this.height = height;
    }

    @Override
    public double area() {
        return width * height;
    }
}

// Sealed class can be extended by non-sealed classes
public non-sealed class Triangle extends Shape {
    private final double base, height;

    public Triangle(String color, double base, double height) {
        super(color);
        this.base = base;
        this.height = height;
    }

    @Override
    public double area() {
        return 0.5 * base * height;
    }
}

// Triangle can now be extended by any class
public class RightTriangle extends Triangle {
    public RightTriangle(String color, double side1, double side2) {
        super(color, side1, side2);
    }
}

// Pattern matching with sealed classes (future benefit)
public String describeShape(Shape shape) {
    return switch (shape) {
        case Circle c -> "Circle with radius " + c.radius;
        case Rectangle r -> "Rectangle " + r.width + "x" + r.height;
        case Triangle t -> "Triangle with base " + t.base;
        // No default needed - compiler knows all possibilities
    };
}

// Sealed interfaces
public sealed interface Vehicle permits Car, Motorcycle, Truck {}

public final class Car implements Vehicle {
    private final String model;
    private final int doors;

    // Implementation
}

public final class Motorcycle implements Vehicle {
    private final String model;
    private final int engineSize;

    // Implementation
}

// Sealed hierarchies for state machines
public sealed interface ConnectionState
    permits Connecting, Connected, Disconnected, Failed {
}

public record Connecting() implements ConnectionState {}
public record Connected(String sessionId) implements ConnectionState {}
public record Disconnected() implements ConnectionState {}
public record Failed(String error) implements ConnectionState {}

// Perfect for handling all states
public void handleConnectionState(ConnectionState state) {
    switch (state) {
        case Connecting() -> showSpinner();
        case Connected(var sessionId) -> onConnected(sessionId);
        case Disconnected() -> onDisconnected();
        case Failed(var error) -> showError(error);
    }
}
```

#### 3. **Hidden Classes**
Classes that cannot be discovered by other classes and have a limited lifecycle.

```java
// Hidden classes are primarily for framework developers
// They're created programmatically, not through source code

// Example: Creating hidden classes for dynamic proxies or code generation
import java.lang.invoke.MethodHandles;

public class HiddenClassExample {
    public static void createHiddenClass() throws Exception {
        // Bytecode for a simple class
        byte[] classBytes = generateClassBytecode();

        // Create hidden class
        MethodHandles.Lookup lookup = MethodHandles.lookup();
        Class<?> hiddenClass = lookup
            .defineHiddenClass(classBytes, true,
                              MethodHandles.Lookup.ClassOption.NESTMATE)
            .lookupClass();

        // Hidden class properties:
        // - Cannot be discovered by Class.forName()
        // - Cannot be referenced by other classes' bytecode
        // - Automatically unloaded when no longer referenced
        // - Perfect for generated code that should not leak

        Object instance = hiddenClass.getConstructor().newInstance();
        System.out.println("Created instance of hidden class: " + instance);
    }
}

// Benefits for frameworks:
// 1. Better memory management (automatic cleanup)
// 2. Security (cannot be accessed by malicious code)
// 3. Performance (JVM can optimize better)
// 4. Isolation (no namespace pollution)
```

#### 4. **Pattern Matching for instanceof (Second Preview)**
Improvements to pattern matching with more robust semantics.

```java
// Enhanced pattern matching (second preview)
public class PatternMatchingEnhancements {

    // Pattern matching with complex conditions
    public String processData(Object data) {
        if (data instanceof String str && str.length() > 5) {
            return "Long string: " + str.toUpperCase();
        } else if (data instanceof Integer num && num > 100) {
            return "Large number: " + num * 2;
        } else if (data instanceof List<?> list && !list.isEmpty()) {
            return "Non-empty list with " + list.size() + " items";
        }
        return "Other data type";
    }

    // Pattern matching in complex expressions
    public boolean isValidUser(Object obj) {
        return obj instanceof User user &&
               user.isActive() &&
               user.getEmail() != null &&
               user.getEmail().contains("@");
    }

    // Nested pattern matching
    public void handleRequest(Object request) {
        if (request instanceof HttpRequest httpReq) {
            if (httpReq.getBody() instanceof JsonData jsonData &&
                jsonData.hasField("userId")) {
                processJsonRequest(httpReq, jsonData);
            } else if (httpReq.getBody() instanceof XmlData xmlData) {
                processXmlRequest(httpReq, xmlData);
            }
        } else if (request instanceof WebSocketMessage wsMsg) {
            handleWebSocketMessage(wsMsg);
        }
    }
}
```

### 🛠 **JVM and Performance Improvements**

#### **Edwards-Curve Digital Signature Algorithm**
Support for EdDSA cryptographic signatures.

```java
// EdDSA signature support
import java.security.KeyPairGenerator;
import java.security.Signature;

KeyPairGenerator keyGen = KeyPairGenerator.getInstance("Ed25519");
var keyPair = keyGen.generateKeyPair();

Signature signature = Signature.getInstance("Ed25519");
signature.initSign(keyPair.getPrivate());
signature.update("Hello World".getBytes());
byte[] signatureBytes = signature.sign();

// Verify signature
signature.initVerify(keyPair.getPublic());
signature.update("Hello World".getBytes());
boolean verified = signature.verify(signatureBytes);
```

#### **Z Garbage Collector (Production Ready)**
ZGC is now production-ready with significant improvements.

```bash
# ZGC is now production ready
java -XX:+UseZGC -XX:+UnlockExperimentalVMOptions MyApp

# ZGC improvements in Java 15:
# - Reduced allocation rate impact
# - Better NUMA awareness
# - Improved concurrent thread handling
# - Support for class unloading

# ZGC configuration options
java -XX:+UseZGC \
     -XX:SoftMaxHeapSize=30g \
     -XX:+UncommitUnusedPages \
     -Xlog:gc*:gc.log MyApp
```

#### **Shenandoah Garbage Collector (Production Ready)**
Shenandoah GC graduates from experimental status.

```bash
# Shenandoah is now production ready
java -XX:+UseShenandoahGC MyApp

# Shenandoah optimizations
java -XX:+UseShenandoahGC \
     -XX:ShenandoahGCHeuristics=adaptive \
     -XX:+ShenandoahUncommit \
     -Xlog:gc*:shenandoah.log MyApp
```

### 🔧 **API Enhancements**

#### **CharSequence.isEmpty() Default Method**
Convenience method for checking empty character sequences.

```java
// New default method in CharSequence
public interface CharSequence {
    default boolean isEmpty() {
        return this.length() == 0;
    }
}

// Usage
String str = "";
StringBuilder sb = new StringBuilder();
StringBuffer buf = new StringBuffer();

// All CharSequence implementations now have isEmpty()
if (str.isEmpty()) { /* ... */ }
if (sb.isEmpty()) { /* ... */ }
if (buf.isEmpty()) { /* ... */ }

// Custom CharSequence implementations get it for free
public class MyCharSequence implements CharSequence {
    // Only need to implement length(), charAt(), subSequence()
    // isEmpty() is inherited
}
```

#### **TreeMap Methods**
New methods for better TreeMap usability.

```java
TreeMap<Integer, String> map = new TreeMap<>();
map.put(1, "one");
map.put(3, "three");
map.put(5, "five");
map.put(7, "seven");
map.put(9, "nine");

// New methods for ranges
NavigableMap<Integer, String> headMap = map.headMap(5, true);  // <= 5
NavigableMap<Integer, String> tailMap = map.tailMap(5, false); // > 5
NavigableMap<Integer, String> subMap = map.subMap(3, true, 7, false); // 3 <= key < 7

// Null-safe operations
String value = map.getOrDefault(2, "default");
map.putIfAbsent(2, "two");
```

---

## Java 16 (March 2021)

**Major Theme**: Records and pattern matching finalized, vector API incubator

### 🔥 Major Features

#### 1. **Records (Final)**
Records become a permanent feature with additional capabilities.

```java
// Records are now standard (no longer preview)
public record Point(int x, int y) {
    // Compact constructor for validation
    public Point {
        if (x < 0 || y < 0) {
            throw new IllegalArgumentException("Coordinates must be non-negative");
        }
    }

    // Additional constructors
    public Point() {
        this(0, 0);
    }

    public Point(int coordinate) {
        this(coordinate, coordinate);
    }

    // Additional methods
    public double distanceFromOrigin() {
        return Math.sqrt(x * x + y * y);
    }

    public Point translate(int dx, int dy) {
        return new Point(x + dx, y + dy);
    }
}

// Generic records
public record Pair<T, U>(T first, U second) {
    public static <T, U> Pair<T, U> of(T first, U second) {
        return new Pair<>(first, second);
    }
}

// Usage
var coordinates = new Point(10, 20);
var nameAge = new Pair<>("Alice", 30);
var keyValue = Pair.of("username", "alice123");

// Records with interfaces
public interface Drawable {
    void draw();
}

public record Circle(int x, int y, int radius) implements Drawable {
    @Override
    public void draw() {
        System.out.println("Drawing circle at (" + x + ", " + y + ") with radius " + radius);
    }
}

// Complex record example - immutable data transfer
public record UserProfile(
    String username,
    String email,
    LocalDateTime lastLogin,
    Set<String> roles,
    Map<String, String> preferences
) {
    public UserProfile {
        // Defensive copying for mutable fields
        roles = Set.copyOf(roles);
        preferences = Map.copyOf(preferences);
    }

    // Convenience methods
    public boolean hasRole(String role) {
        return roles.contains(role);
    }

    public String getPreference(String key, String defaultValue) {
        return preferences.getOrDefault(key, defaultValue);
    }

    public UserProfile withLastLogin(LocalDateTime newLastLogin) {
        return new UserProfile(username, email, newLastLogin, roles, preferences);
    }
}
```

#### 2. **Pattern Matching for instanceof (Final)**
Pattern matching is now a standard feature.

```java
// Pattern matching is now standard
public class ShapeProcessor {

    public double calculateArea(Object shape) {
        if (shape instanceof Circle circle) {
            return Math.PI * circle.radius() * circle.radius();
        } else if (shape instanceof Rectangle rectangle) {
            return rectangle.width() * rectangle.height();
        } else if (shape instanceof Triangle triangle) {
            return 0.5 * triangle.base() * triangle.height();
        }
        throw new IllegalArgumentException("Unknown shape: " + shape);
    }

    // Pattern matching with complex conditions
    public String analyzeShape(Object shape) {
        if (shape instanceof Circle circle && circle.radius() > 10) {
            return "Large circle with area " + (Math.PI * circle.radius() * circle.radius());
        } else if (shape instanceof Rectangle rect && rect.width() == rect.height()) {
            return "Square with side " + rect.width();
        } else if (shape instanceof Triangle triangle &&
                   triangle.base() == triangle.height()) {
            return "Isosceles right triangle";
        }
        return "Regular shape";
    }

    // Nested pattern matching
    public void processContainer(Object container) {
        if (container instanceof Box box) {
            if (box.contents() instanceof Electronics electronics &&
                electronics.warranty() > 365) {
                handleWarrantyElectronics(electronics);
            } else if (box.contents() instanceof Books books) {
                processBooks(books);
            }
        }
    }
}

// Records with pattern matching
public record Response(int status, String message, Object data) {}

public void handleResponse(Response response) {
    if (response.status() == 200 && response.data() instanceof UserData userData) {
        System.out.println("Success: User " + userData.name() + " processed");
    } else if (response.status() >= 400 && response.message() instanceof String error) {
        System.err.println("Error " + response.status() + ": " + error);
    }
}
```

#### 3. **Sealed Classes (Second Preview)**
Improved sealed classes with better syntax and features.

```java
// Sealed classes second preview with improvements
public sealed interface Result<T, E>
    permits Success, Failure {
}

public record Success<T, E>(T value) implements Result<T, E> {}
public record Failure<T, E>(E error) implements Result<T, E> {}

// Usage with pattern matching (future enhancement)
public <T, E> void handleResult(Result<T, E> result) {
    switch (result) {
        case Success<T, E> success -> processSuccess(success.value());
        case Failure<T, E> failure -> processError(failure.error());
    }
}

// Sealed class hierarchy for expression trees
public sealed interface Expression
    permits BinaryExpression, UnaryExpression, Literal {
}

public sealed interface BinaryExpression extends Expression
    permits Addition, Subtraction, Multiplication, Division {
}

public record Addition(Expression left, Expression right) implements BinaryExpression {}
public record Subtraction(Expression left, Expression right) implements BinaryExpression {}
public record Multiplication(Expression left, Expression right) implements BinaryExpression {}
public record Division(Expression left, Expression right) implements BinaryExpression {}

public sealed interface UnaryExpression extends Expression
    permits Negation, Absolute {
}

public record Negation(Expression operand) implements UnaryExpression {}
public record Absolute(Expression operand) implements UnaryExpression {}

public record Literal(double value) implements Expression {}

// Expression evaluator with exhaustive pattern matching
public double evaluate(Expression expr) {
    return switch (expr) {
        case Addition(var left, var right) -> evaluate(left) + evaluate(right);
        case Subtraction(var left, var right) -> evaluate(left) - evaluate(right);
        case Multiplication(var left, var right) -> evaluate(left) * evaluate(right);
        case Division(var left, var right) -> evaluate(left) / evaluate(right);
        case Negation(var operand) -> -evaluate(operand);
        case Absolute(var operand) -> Math.abs(evaluate(operand));
        case Literal(var value) -> value;
    };
}
```

### 🛠 **JVM and Performance**

#### **Vector API (Incubator)**
SIMD (Single Instruction, Multiple Data) operations for better performance.

```java
// Vector API for SIMD operations (incubator)
// Requires --add-modules jdk.incubator.vector

import jdk.incubator.vector.*;

public class VectorExample {
    static final VectorSpecies<Integer> SPECIES = IntVector.SPECIES_256;

    public void vectorizedAddition(int[] a, int[] b, int[] result) {
        int length = a.length;
        int loopBound = SPECIES.loopBound(length);

        // Vectorized loop
        for (int i = 0; i < loopBound; i += SPECIES.length()) {
            IntVector va = IntVector.fromArray(SPECIES, a, i);
            IntVector vb = IntVector.fromArray(SPECIES, b, i);
            IntVector vc = va.add(vb);
            vc.intoArray(result, i);
        }

        // Handle remaining elements
        for (int i = loopBound; i < length; i++) {
            result[i] = a[i] + b[i];
        }
    }

    // Complex vector operations
    public double dotProduct(double[] a, double[] b) {
        var species = DoubleVector.SPECIES_PREFERRED;
        var sum = DoubleVector.zero(species);
        int loopBound = species.loopBound(a.length);

        for (int i = 0; i < loopBound; i += species.length()) {
            var va = DoubleVector.fromArray(species, a, i);
            var vb = DoubleVector.fromArray(species, b, i);
            sum = va.fma(vb, sum); // fused multiply-add
        }

        double result = sum.reduceLanes(VectorOperators.ADD);

        // Handle remaining elements
        for (int i = loopBound; i < a.length; i++) {
            result += a[i] * b[i];
        }

        return result;
    }
}
```

#### **Foreign Linker API (Incubator)**
Call native code without JNI overhead.

```java
// Foreign Linker API (incubator)
// Requires --add-modules jdk.incubator.foreign

import jdk.incubator.foreign.*;

public class ForeignLinkerExample {
    public static void callNativeFunction() {
        // Define function signature
        FunctionDescriptor descriptor = FunctionDescriptor.of(
            CLinker.C_INT,        // return type
            CLinker.C_POINTER     // parameter type
        );

        // Get native function
        MethodHandle strlen = CLinker.getInstance()
            .downcallHandle(
                CLinker.systemLookup().lookup("strlen").get(),
                MethodType.methodType(int.class, MemoryAddress.class),
                descriptor
            );

        // Allocate native memory for string
        try (MemorySegment nativeString = CLinker.toCString("Hello World!")) {
            int length = (int) strlen.invokeExact(nativeString.address());
            System.out.println("String length: " + length);
        } catch (Throwable e) {
            e.printStackTrace();
        }
    }
}
```

#### **Elastic Metaspace**
Better memory management for class metadata.

```bash
# Elastic Metaspace improvements:
# - Returns unused memory to OS more promptly
# - Better handling of class unloading
# - Reduced fragmentation
# - More predictable memory usage

# Monitor metaspace usage
java -XX:+UseG1GC \
     -Xlog:metaspace*:metaspace.log \
     -XX:MetaspaceSize=128m \
     MyApp
```

### 🔧 **API and Language Enhancements**

#### **Stream.toList()**
Convenient method to collect streams to lists.

```java
// Before Java 16 - verbose collection
List<String> names = people.stream()
    .map(Person::getName)
    .collect(Collectors.toList());

// Java 16+ - convenient toList()
List<String> names = people.stream()
    .map(Person::getName)
    .toList(); // Returns unmodifiable list

// Important: toList() returns unmodifiable list
List<String> mutableNames = people.stream()
    .map(Person::getName)
    .collect(Collectors.toList()); // Mutable

List<String> immutableNames = people.stream()
    .map(Person::getName)
    .toList(); // Immutable

// Complex stream processing
List<String> processedData = dataStream
    .filter(Objects::nonNull)
    .map(String::toUpperCase)
    .filter(s -> s.length() > 3)
    .sorted()
    .toList();

// Parallel streams
List<Integer> results = IntStream.range(0, 1000000)
    .parallel()
    .filter(n -> n % 2 == 0)
    .map(n -> n * n)
    .boxed()
    .toList();
```

#### **Day Period Support**
Enhanced date/time formatting with day periods.

```java
// Day period formatting (AM/PM alternatives)
DateTimeFormatter formatter = DateTimeFormatter.ofPattern("h:mm B");

LocalTime morning = LocalTime.of(10, 30);
LocalTime afternoon = LocalTime.of(14, 30);
LocalTime evening = LocalTime.of(19, 30);
LocalTime night = LocalTime.of(23, 30);

System.out.println(morning.format(formatter));    // "10:30 in the morning"
System.out.println(afternoon.format(formatter));  // "2:30 in the afternoon"
System.out.println(evening.format(formatter));    // "7:30 in the evening"
System.out.println(night.format(formatter));      // "11:30 at night"

// Locale-specific day periods
DateTimeFormatter germanFormatter = DateTimeFormatter.ofPattern("h:mm B", Locale.GERMAN);
System.out.println(afternoon.format(germanFormatter)); // Localized output
```

#### **InvocationHandler::invokeDefault**
Default method support in dynamic proxies.

```java
// Enhanced proxy support for default methods
interface Calculator {
    int add(int a, int b);
    int multiply(int a, int b);

    default int square(int a) {
        return multiply(a, a);
    }
}

// Proxy with default method support
Calculator proxy = (Calculator) Proxy.newProxyInstance(
    Calculator.class.getClassLoader(),
    new Class[]{Calculator.class},
    new InvocationHandler() {
        @Override
        public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
            if (method.isDefault()) {
                // Java 16+ - invoke default method
                return InvocationHandler.invokeDefault(proxy, method, args);
            }

            // Handle abstract methods
            return switch (method.getName()) {
                case "add" -> (int) args[0] + (int) args[1];
                case "multiply" -> (int) args[0] * (int) args[1];
                default -> throw new UnsupportedOperationException(method.getName());
            };
        }
    });

int result = proxy.square(5); // Uses default method implementation
```

---

## Java 17 (September 2021) - LTS

**Major Theme**: Long-term support with sealed classes finalized and enhanced random generators

### 🔥 Major Features

#### 1. **Sealed Classes (Final)**
Sealed classes become a standard feature with complete functionality.

```java
// Sealed classes are now standard
public sealed class Vehicle permits Car, Motorcycle, Truck {
    protected final String brand;
    protected final int year;

    protected Vehicle(String brand, int year) {
        this.brand = brand;
        this.year = year;
    }

    public String getBrand() { return brand; }
    public int getYear() { return year; }
    public abstract double calculateInsurance();
}

public final class Car extends Vehicle {
    private final int doors;
    private final String fuelType;

    public Car(String brand, int year, int doors, String fuelType) {
        super(brand, year);
        this.doors = doors;
        this.fuelType = fuelType;
    }

    @Override
    public double calculateInsurance() {
        double base = 500.0;
        if (doors > 2) base += 100;
        if ("electric".equals(fuelType)) base *= 0.9;
        return base;
    }

    public int getDoors() { return doors; }
    public String getFuelType() { return fuelType; }
}

public final class Motorcycle extends Vehicle {
    private final int engineSize;

    public Motorcycle(String brand, int year, int engineSize) {
        super(brand, year);
        this.engineSize = engineSize;
    }

    @Override
    public double calculateInsurance() {
        return 300.0 + (engineSize * 0.1);
    }

    public int getEngineSize() { return engineSize; }
}

public non-sealed class Truck extends Vehicle {
    private final double maxWeight;

    public Truck(String brand, int year, double maxWeight) {
        super(brand, year);
        this.maxWeight = maxWeight;
    }

    @Override
    public double calculateInsurance() {
        return 800.0 + (maxWeight * 0.05);
    }

    public double getMaxWeight() { return maxWeight; }
}

// Truck can be extended (non-sealed)
public class DeliveryTruck extends Truck {
    private final String company;

    public DeliveryTruck(String brand, int year, double maxWeight, String company) {
        super(brand, year, maxWeight);
        this.company = company;
    }

    @Override
    public double calculateInsurance() {
        return super.calculateInsurance() * 1.2; // Commercial rate
    }
}

// Perfect for pattern matching and exhaustive switch
public String describeVehicle(Vehicle vehicle) {
    return switch (vehicle) {
        case Car car -> String.format("Car: %s %d with %d doors (%s)",
                                     car.getBrand(), car.getYear(),
                                     car.getDoors(), car.getFuelType());
        case Motorcycle moto -> String.format("Motorcycle: %s %d with %dcc engine",
                                             moto.getBrand(), moto.getYear(),
                                             moto.getEngineSize());
        case Truck truck -> String.format("Truck: %s %d with %.1ft capacity",
                                         truck.getBrand(), truck.getYear(),
                                         truck.getMaxWeight());
    };
}

// Sealed interfaces for state management
public sealed interface OrderStatus permits Pending, Processing, Shipped, Delivered, Cancelled {}

public record Pending(LocalDateTime orderTime) implements OrderStatus {}
public record Processing(LocalDateTime startTime, String assignedWorker) implements OrderStatus {}
public record Shipped(LocalDateTime shipTime, String trackingNumber) implements OrderStatus {}
public record Delivered(LocalDateTime deliveryTime, String recipient) implements OrderStatus {}
public record Cancelled(LocalDateTime cancelTime, String reason) implements OrderStatus {}

// Exhaustive switch with sealed interfaces
public void handleOrderStatus(OrderStatus status) {
    switch (status) {
        case Pending(var orderTime) ->
            System.out.println("Order pending since " + orderTime);
        case Processing(var startTime, var worker) ->
            System.out.println("Being processed by " + worker + " since " + startTime);
        case Shipped(var shipTime, var tracking) ->
            System.out.println("Shipped on " + shipTime + ", tracking: " + tracking);
        case Delivered(var deliveryTime, var recipient) ->
            System.out.println("Delivered to " + recipient + " on " + deliveryTime);
        case Cancelled(var cancelTime, var reason) ->
            System.out.println("Cancelled on " + cancelTime + ": " + reason);
    }
}
```

#### 2. **Enhanced Pseudo-Random Number Generators**
New random number generator API with multiple algorithms.

```java
// New RandomGenerator interface with multiple implementations
import java.util.random.*;

public class RandomGeneratorExamples {

    public void demonstrateRandomGenerators() {
        // Different random generator algorithms
        RandomGenerator random1 = RandomGeneratorFactory.of("L32X64MixRandom").create();
        RandomGenerator random2 = RandomGeneratorFactory.of("Xoshiro256PlusPlus").create();
        RandomGenerator random3 = RandomGeneratorFactory.of("L128X256MixRandom").create();

        // List all available algorithms
        RandomGeneratorFactory.all()
            .forEach(factory -> {
                System.out.println("Algorithm: " + factory.name());
                System.out.println("  Period: " + factory.period());
                System.out.println("  Statistically independent: " + factory.isStatisticallyIndependent());
                System.out.println("  Stochastically independent: " + factory.isStochasticIndependent());
                System.out.println("  Hardware: " + factory.isHardware());
                System.out.println();
            });

        // Create splittable random generator
        SplittableRandomGenerator splittable =
            RandomGeneratorFactory.of("L128X256MixRandom").create(12345L);

        // Split for parallel processing
        List<SplittableRandomGenerator> generators = splittable
            .splits(10)  // Create 10 independent generators
            .collect(Collectors.toList());

        // Use different generators in parallel streams
        List<Integer> randomNumbers = generators.parallelStream()
            .flatMapToInt(gen -> gen.ints(1000, 1, 101))  // 1000 numbers 1-100
            .boxed()
            .collect(Collectors.toList());

        System.out.println("Generated " + randomNumbers.size() + " random numbers");
    }

    // Monte Carlo simulation with splittable generators
    public double estimatePi(int samples) {
        SplittableRandomGenerator baseRandom =
            RandomGeneratorFactory.of("L128X256MixRandom").create();

        return baseRandom.splits(Runtime.getRuntime().availableProcessors())
            .parallel()
            .mapToDouble(random -> {
                long insideCircle = random.doubles(samples / Runtime.getRuntime().availableProcessors())
                    .map(x -> random.nextDouble())
                    .mapToLong(x -> (x * x + random.nextDouble() * random.nextDouble() <= 1) ? 1 : 0)
                    .sum();
                return 4.0 * insideCircle / (samples / Runtime.getRuntime().availableProcessors());
            })
            .average()
            .orElse(0.0);
    }

    // Secure random with better performance
    public void secureRandomExample() {
        RandomGenerator secureRandom = RandomGeneratorFactory.of("SecureRandom").create();

        // Generate cryptographically secure random data
        byte[] secureBytes = new byte[32];
        secureRandom.nextBytes(secureBytes);

        // Generate secure passwords
        String charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*";
        StringBuilder password = new StringBuilder();
        for (int i = 0; i < 12; i++) {
            password.append(charset.charAt(secureRandom.nextInt(charset.length())));
        }
        System.out.println("Secure password: " + password);
    }
}
```

### 🛠 **Performance and JVM Improvements**

#### **Context-Specific Deserialization Filters**
Enhanced security for Java serialization.

```java
// Context-specific deserialization filters
public class SerializationSecurity {

    public static void setupGlobalFilter() {
        // Set global deserialization filter
        ObjectInputFilter globalFilter = ObjectInputFilter.Config
            .createFilter("java.lang.*;java.util.*;com.example.safe.**;!*");
        ObjectInputFilter.Config.setSerialFilter(globalFilter);
    }

    public static <T> T deserializeWithFilter(byte[] data, Class<T> expectedType)
            throws IOException, ClassNotFoundException {

        ByteArrayInputStream bis = new ByteArrayInputStream(data);
        ObjectInputStream ois = new ObjectInputStream(bis);

        // Set specific filter for this deserialization
        ObjectInputFilter filter = ObjectInputFilter.Config.createFilter(
            expectedType.getName() + ";java.lang.*;java.util.*;!*"
        );
        ois.setObjectInputFilter(filter);

        @SuppressWarnings("unchecked")
        T result = (T) ois.readObject();
        return result;
    }

    // Custom filter implementation
    static class SafeObjectFilter implements ObjectInputFilter {
        private final Set<String> allowedClasses;
        private final long maxBytes;

        public SafeObjectFilter(Set<String> allowedClasses, long maxBytes) {
            this.allowedClasses = Set.copyOf(allowedClasses);
            this.maxBytes = maxBytes;
        }

        @Override
        public Status checkInput(FilterInfo info) {
            if (info.references() > 1000) return Status.REJECTED;
            if (info.depth() > 10) return Status.REJECTED;
            if (info.streamBytes() > maxBytes) return Status.REJECTED;

            Class<?> clazz = info.serialClass();
            if (clazz != null) {
                String className = clazz.getName();
                if (allowedClasses.contains(className) ||
                    className.startsWith("java.lang.") ||
                    className.startsWith("java.util.")) {
                    return Status.ALLOWED;
                } else {
                    return Status.REJECTED;
                }
            }

            return Status.UNDECIDED;
        }
    }
}
```

#### **Foreign Function & Memory API (Incubator)**
Improved foreign function interface replacing JNI.

```java
// Foreign Function & Memory API (incubator in Java 17)
// Requires --add-modules jdk.incubator.foreign

import jdk.incubator.foreign.*;

public class ForeignFunctionExample {

    public static void callLibraryFunction() {
        // Define function signature
        FunctionDescriptor descriptor = FunctionDescriptor.of(
            CLinker.C_INT,      // return type
            CLinker.C_POINTER,  // first parameter
            CLinker.C_INT       // second parameter
        );

        // Get function from library
        SymbolLookup stdlib = CLinker.systemLookup();
        MemoryAddress mallocAddr = stdlib.lookup("malloc").orElseThrow();

        MethodHandle malloc = CLinker.getInstance().downcallHandle(
            mallocAddr,
            MethodType.methodType(MemoryAddress.class, long.class),
            FunctionDescriptor.of(CLinker.C_POINTER, CLinker.C_LONG)
        );

        try {
            // Allocate memory
            MemoryAddress ptr = (MemoryAddress) malloc.invoke(1024L);

            // Use the allocated memory
            MemorySegment segment = MemorySegment.ofAddress(ptr, 1024, ResourceScope.newImplicitScope());
            MemoryAccess.setByteAtOffset(segment, 0, (byte) 42);
            byte value = MemoryAccess.getByteAtOffset(segment, 0);

            System.out.println("Value from native memory: " + value);

        } catch (Throwable e) {
            e.printStackTrace();
        }
    }

    // Memory segments for direct memory management
    public static void memorySegmentExample() {
        try (ResourceScope scope = ResourceScope.newConfinedScope()) {
            // Allocate native memory
            MemorySegment segment = MemorySegment.allocateNative(1024, scope);

            // Write structured data
            MemoryAccess.setIntAtOffset(segment, 0, 42);
            MemoryAccess.setIntAtOffset(segment, 4, 84);
            MemoryAccess.setDoubleAtOffset(segment, 8, 3.14159);

            // Read structured data
            int value1 = MemoryAccess.getIntAtOffset(segment, 0);
            int value2 = MemoryAccess.getIntAtOffset(segment, 4);
            double pi = MemoryAccess.getDoubleAtOffset(segment, 8);

            System.out.printf("Values: %d, %d, %.5f%n", value1, value2, pi);

            // Memory automatically freed when scope closes
        }
    }
}
```

#### **Vector API (Second Incubator)**
Improved SIMD support with more operations.

```java
// Enhanced Vector API (second incubator)
import jdk.incubator.vector.*;

public class VectorAPIEnhancements {

    // Matrix multiplication using vectors
    public static void matrixMultiply(double[][] a, double[][] b, double[][] result) {
        var species = DoubleVector.SPECIES_PREFERRED;
        int rows = a.length;
        int cols = b[0].length;
        int common = b.length;

        for (int i = 0; i < rows; i++) {
            for (int j = 0; j < cols; j += species.length()) {
                var sum = DoubleVector.zero(species);

                for (int k = 0; k < common; k++) {
                    var aVal = DoubleVector.broadcast(species, a[i][k]);
                    var bVec = DoubleVector.fromArray(species, b[k], j);
                    sum = aVal.fma(bVec, sum);
                }

                sum.intoArray(result[i], j);
            }
        }
    }

    // Image processing with vectors
    public static void applyFilter(int[] pixels, int[] filter, int[] result) {
        var species = IntVector.SPECIES_256;
        int loopBound = species.loopBound(pixels.length);

        for (int i = 0; i < loopBound; i += species.length()) {
            var pixelVec = IntVector.fromArray(species, pixels, i);
            var filterVec = IntVector.fromArray(species, filter, i);

            // Apply filter (multiply and shift)
            var filtered = pixelVec.mul(filterVec).lanewise(VectorOperators.ASHR, 8);
            filtered.intoArray(result, i);
        }

        // Handle remaining elements
        for (int i = loopBound; i < pixels.length; i++) {
            result[i] = (pixels[i] * filter[i]) >> 8;
        }
    }

    // Statistical operations
    public static record Statistics(double mean, double variance, double stdDev) {}

    public static Statistics calculateStatistics(double[] data) {
        var species = DoubleVector.SPECIES_PREFERRED;
        var sum = DoubleVector.zero(species);
        var sumSquares = DoubleVector.zero(species);
        int loopBound = species.loopBound(data.length);

        // Vectorized sum and sum of squares
        for (int i = 0; i < loopBound; i += species.length()) {
            var vec = DoubleVector.fromArray(species, data, i);
            sum = sum.add(vec);
            sumSquares = sumSquares.add(vec.mul(vec));
        }

        double totalSum = sum.reduceLanes(VectorOperators.ADD);
        double totalSumSquares = sumSquares.reduceLanes(VectorOperators.ADD);

        // Handle remaining elements
        for (int i = loopBound; i < data.length; i++) {
            totalSum += data[i];
            totalSumSquares += data[i] * data[i];
        }

        double mean = totalSum / data.length;
        double variance = (totalSumSquares / data.length) - (mean * mean);
        double stdDev = Math.sqrt(variance);

        return new Statistics(mean, variance, stdDev);
    }
}
```

### 🔧 **API Enhancements**

#### **New macOS Rendering Pipeline**
Native Apple Metal API support for better macOS performance.

```bash
# Enable Metal rendering pipeline on macOS
java -Dsun.java2d.metal=true MySwingApplication

# Fallback options
java -Dsun.java2d.opengl=true MySwingApplication  # OpenGL
java -Dsun.java2d.noddraw=true MySwingApplication  # Software rendering
```

#### **Deprecate the Applet API**
Applet API marked for removal in future versions.

```java
// Applet API deprecated for removal
// @Deprecated(forRemoval = true, since = "9")
public class MyApplet extends java.applet.Applet {
    // This API is deprecated and will be removed
    // Migrate to standalone applications or web technologies
}

// Migration alternatives:
// 1. Convert to standalone Swing/JavaFX applications
// 2. Use web technologies (HTML5/JavaScript)
// 3. Create Java Web Start replacements
// 4. Use modern deployment strategies (containers, native packages)
```

#### **Remove RMI Activation**
RMI Activation system removed from the platform.

```java
// RMI Activation removed - use alternatives:

// Before (RMI Activation - removed):
// java.rmi.activation.* classes no longer available

// Alternative 1: Standard RMI
public interface Calculator extends Remote {
    int add(int a, int b) throws RemoteException;
}

public class CalculatorImpl extends UnicastRemoteObject implements Calculator {
    public CalculatorImpl() throws RemoteException {
        super();
    }

    @Override
    public int add(int a, int b) throws RemoteException {
        return a + b;
    }
}

// Alternative 2: Modern alternatives
// - HTTP-based APIs (REST, GraphQL)
// - Message queues (RabbitMQ, Apache Kafka)
// - gRPC for high-performance RPC
// - WebSocket for real-time communication
```

#### **Strongly Encapsulate JDK Internals**
Strong encapsulation of internal APIs by default.

```bash
# Internal APIs now strongly encapsulated
# These will cause warnings or errors:
# - sun.misc.Unsafe
# - com.sun.* packages
# - jdk.internal.* packages

# Temporary workarounds (not recommended for production):
java --add-opens java.base/sun.nio.ch=ALL-UNNAMED \
     --add-opens java.base/java.nio=ALL-UNNAMED \
     MyApplication

# Better approach: Use supported APIs
# - Use MethodHandles instead of Unsafe
# - Use public APIs instead of internal ones
# - Migrate to modern alternatives
```

## Java 18 (March 2022)

**Major Theme**: UTF-8 by default, simple web server, and code snippets in Javadoc

### 🔥 Major Features

#### 1. **UTF-8 by Default**
UTF-8 is now the default charset for Standard Java APIs.

```java
// Before Java 18 - charset could vary by platform
FileReader reader = new FileReader("file.txt"); // Platform default encoding
Files.readAllLines(Paths.get("file.txt")); // Platform default encoding

// Java 18+ - UTF-8 by default
FileReader reader = new FileReader("file.txt"); // UTF-8 by default
Files.readAllLines(Paths.get("file.txt")); // UTF-8 by default

// Explicit charset still supported
FileReader reader = new FileReader("file.txt", StandardCharsets.ISO_8859_1);
Files.readAllLines(Paths.get("file.txt"), StandardCharsets.UTF_16);

// Affected APIs now default to UTF-8:
// - FileReader, FileWriter
// - InputStreamReader, OutputStreamWriter (with default constructor)
// - PrintStream, PrintWriter (with default constructor)
// - Files methods (readAllLines, readString, writeString, etc.)
// - Formatter, Scanner (with default constructor)

// Check default charset
System.out.println("Default charset: " + Charset.defaultCharset());
System.out.println("File encoding: " + System.getProperty("file.encoding"));

// Cross-platform consistency
public void writeConfigFile(Map<String, String> config) throws IOException {
    // Now consistently UTF-8 across all platforms
    try (PrintWriter writer = new PrintWriter("config.properties")) {
        config.forEach((key, value) ->
            writer.println(key + "=" + value));
    }
}
```

#### 2. **Simple Web Server**
Built-in HTTP server for prototyping and testing.

```bash
# Start simple web server from command line
jwebserver
# Serves current directory on http://localhost:8000

# Custom options
jwebserver -p 9000 -d /path/to/directory -o info
# -p: port (default 8000)
# -d: directory to serve (default current directory)
# -o: output level (none, info, verbose)
```

```java
// Programmatic HTTP server
import com.sun.net.httpserver.*;
import java.net.InetSocketAddress;

public class SimpleWebServerExample {
    public static void main(String[] args) throws Exception {
        // Create simple file server
        HttpServer server = SimpleFileServer.createFileServer(
            new InetSocketAddress(8080),
            Path.of("./web-content"),
            SimpleFileServer.OutputLevel.VERBOSE
        );

        // Start the server
        server.start();
        System.out.println("Server started on http://localhost:8080");

        // Custom handler
        server.createContext("/api/hello", exchange -> {
            String response = "Hello from Java 18 Simple Web Server!";
            exchange.getResponseHeaders().set("Content-Type", "text/plain");
            exchange.sendResponseHeaders(200, response.getBytes().length);
            exchange.getResponseBody().write(response.getBytes());
            exchange.close();
        });

        // JSON API endpoint
        server.createContext("/api/time", exchange -> {
            String json = """
                {
                    "timestamp": "%s",
                    "timezone": "%s",
                    "formatted": "%s"
                }
                """.formatted(
                    Instant.now().toString(),
                    ZoneId.systemDefault().toString(),
                    LocalDateTime.now().format(DateTimeFormatter.RFC_1123_DATE_TIME)
                );

            exchange.getResponseHeaders().set("Content-Type", "application/json");
            exchange.sendResponseHeaders(200, json.getBytes().length);
            exchange.getResponseBody().write(json.getBytes());
            exchange.close();
        });

        // Keep server running
        Runtime.getRuntime().addShutdownHook(new Thread(server::stop));
    }
}

// Production-ready web server configuration
public class ProductionWebServer {
    public static HttpServer createProductionServer() throws IOException {
        HttpServer server = HttpServer.create(new InetSocketAddress(8080), 100);

        // Thread pool for handling requests
        server.setExecutor(Executors.newFixedThreadPool(50));

        // Static file serving with caching
        server.createContext("/static", new StaticFileHandler());

        // API endpoints
        server.createContext("/api/v1/users", new UserHandler());
        server.createContext("/api/v1/orders", new OrderHandler());

        return server;
    }
}
```

#### 3. **Code Snippets in Javadoc**
Enhanced Javadoc with executable code snippets.

```java
/**
 * Utility class for string operations.
 *
 * {@snippet :
 * // Basic usage example
 * String result = StringUtils.reverse("hello");
 * assert result.equals("olleh");
 * }
 *
 * @since 18
 */
public class StringUtils {

    /**
     * Reverses a string.
     *
     * {@snippet :
     * String original = "Java 18";
     * String reversed = StringUtils.reverse(original);
     * System.out.println(reversed); // "81 avaJ"
     * }
     *
     * {@snippet :
     * // Handle edge cases
     * assert StringUtils.reverse("").equals("");
     * assert StringUtils.reverse("a").equals("a");
     * assert StringUtils.reverse(null) == null;
     * }
     *
     * @param str the string to reverse
     * @return the reversed string, or null if input is null
     */
    public static String reverse(String str) {
        if (str == null) return null;
        return new StringBuilder(str).reverse().toString();
    }

    /**
     * Splits a string by delimiter with advanced options.
     *
     * {@snippet file="StringUtilsExample.java" region="advanced-split"}
     *
     * @param input the string to split
     * @param delimiter the delimiter pattern
     * @param trimResults whether to trim whitespace
     * @param ignoreEmpty whether to ignore empty strings
     * @return array of split strings
     */
    public static String[] advancedSplit(String input, String delimiter,
                                       boolean trimResults, boolean ignoreEmpty) {
        // Implementation here
        return new String[0];
    }
}

// External file: StringUtilsExample.java
// @start region="advanced-split"
public void demonstrateAdvancedSplit() {
    String csv = " apple , banana , , cherry , ";

    // Basic split
    String[] basic = csv.split(",");
    System.out.println(Arrays.toString(basic));
    // [" apple ", " banana ", " ", " cherry ", " "]

    // Advanced split with options
    String[] advanced = StringUtils.advancedSplit(csv, ",", true, true);
    System.out.println(Arrays.toString(advanced));
    // ["apple", "banana", "cherry"]
}
// @end
```

### 🛠 **Performance and JVM Improvements**

#### **Vector API (Third Incubator)**
Further improvements to SIMD operations.

```java
// Vector API third incubator with more operations
import jdk.incubator.vector.*;

public class VectorAPI3rdIncubator {

    // Trigonometric operations with vectors
    public static void vectorizedSineCosine(double[] angles, double[] sines, double[] cosines) {
        var species = DoubleVector.SPECIES_PREFERRED;
        int loopBound = species.loopBound(angles.length);

        for (int i = 0; i < loopBound; i += species.length()) {
            var angleVec = DoubleVector.fromArray(species, angles, i);

            // Compute sine and cosine in parallel
            var sineVec = angleVec.lanewise(VectorOperators.SIN);
            var cosineVec = angleVec.lanewise(VectorOperators.COS);

            sineVec.intoArray(sines, i);
            cosineVec.intoArray(cosines, i);
        }

        // Handle remaining elements
        for (int i = loopBound; i < angles.length; i++) {
            sines[i] = Math.sin(angles[i]);
            cosines[i] = Math.cos(angles[i]);
        }
    }

    // Vector-based searching
    public static int vectorizedSearch(int[] array, int target) {
        var species = IntVector.SPECIES_PREFERRED;
        var targetVec = IntVector.broadcast(species, target);
        int loopBound = species.loopBound(array.length);

        for (int i = 0; i < loopBound; i += species.length()) {
            var arrayVec = IntVector.fromArray(species, array, i);
            var mask = arrayVec.eq(targetVec);

            if (mask.anyTrue()) {
                // Found match, find first occurrence
                for (int j = i; j < i + species.length(); j++) {
                    if (array[j] == target) return j;
                }
            }
        }

        // Search remaining elements
        for (int i = loopBound; i < array.length; i++) {
            if (array[i] == target) return i;
        }

        return -1; // Not found
    }

    // Parallel reduction with vectors
    public static double vectorizedVariance(double[] data, double mean) {
        var species = DoubleVector.SPECIES_PREFERRED;
        var meanVec = DoubleVector.broadcast(species, mean);
        var sumSquares = DoubleVector.zero(species);
        int loopBound = species.loopBound(data.length);

        for (int i = 0; i < loopBound; i += species.length()) {
            var dataVec = DoubleVector.fromArray(species, data, i);
            var diff = dataVec.sub(meanVec);
            sumSquares = diff.fma(diff, sumSquares);
        }

        double result = sumSquares.reduceLanes(VectorOperators.ADD);

        // Handle remaining elements
        for (int i = loopBound; i < data.length; i++) {
            double diff = data[i] - mean;
            result += diff * diff;
        }

        return result / (data.length - 1);
    }
}
```

#### **Foreign Function & Memory API (Second Incubator)**
Improved native interop capabilities.

```java
// Enhanced Foreign Function & Memory API
// Requires --add-modules jdk.incubator.foreign

import jdk.incubator.foreign.*;

public class ForeignAPIEnhanced {

    // Structured data access
    public static void structuredDataExample() {
        // Define struct layout
        GroupLayout pointLayout = MemoryLayout.structLayout(
            ValueLayout.JAVA_INT.withName("x"),
            ValueLayout.JAVA_INT.withName("y")
        ).withName("Point");

        try (MemorySession session = MemorySession.openConfined()) {
            MemorySegment point = MemorySegment.allocateNative(pointLayout, session);

            // Set values using var handles
            VarHandle xHandle = pointLayout.varHandle(PathElement.groupElement("x"));
            VarHandle yHandle = pointLayout.varHandle(PathElement.groupElement("y"));

            xHandle.set(point, 10);
            yHandle.set(point, 20);

            // Read values
            int x = (int) xHandle.get(point);
            int y = (int) yHandle.get(point);

            System.out.printf("Point: (%d, %d)%n", x, y);
        }
    }

    // Array operations with native memory
    public static void nativeArrayExample() {
        try (MemorySession session = MemorySession.openConfined()) {
            // Allocate array of 1000 integers
            MemorySegment intArray = MemorySegment.allocateNative(
                ValueLayout.JAVA_INT, 1000, session);

            // Fill array
            for (int i = 0; i < 1000; i++) {
                intArray.setAtIndex(ValueLayout.JAVA_INT, i, i * i);
            }

            // Process array
            long sum = 0;
            for (int i = 0; i < 1000; i++) {
                sum += intArray.getAtIndex(ValueLayout.JAVA_INT, i);
            }

            System.out.println("Sum of squares: " + sum);
        }
    }

    // Callback functions from native code
    public static void callbackExample() {
        try (MemorySession session = MemorySession.openConfined()) {
            // Define callback signature
            FunctionDescriptor callbackSig = FunctionDescriptor.of(
                ValueLayout.JAVA_INT,
                ValueLayout.JAVA_INT,
                ValueLayout.JAVA_INT
            );

            // Create callback method handle
            MethodHandle callback = MethodHandles.lookup().findStatic(
                ForeignAPIEnhanced.class, "addCallback",
                MethodType.methodType(int.class, int.class, int.class)
            );

            // Create upcall stub
            MemorySegment callbackSegment = Linker.nativeLinker()
                .upcallStub(callback, callbackSig, session);

            // Use callback with native function
            System.out.println("Callback segment: " + callbackSegment);
        } catch (NoSuchMethodException | IllegalAccessException e) {
            e.printStackTrace();
        }
    }

    // Callback method
    public static int addCallback(int a, int b) {
        System.out.printf("Native code called back with: %d + %d%n", a, b);
        return a + b;
    }
}
```

#### **Pattern Matching for switch (Second Preview)**
Enhanced pattern matching capabilities.

```java
// Pattern matching for switch (second preview)
public class PatternMatchingSwitch {

    // Basic pattern matching switch
    public String describe(Object obj) {
        return switch (obj) {
            case Integer i -> "Integer: " + i;
            case String s -> "String of length " + s.length();
            case Long l -> "Long: " + l;
            case Double d -> "Double: " + d;
            case null -> "null value";
            default -> "Unknown type: " + obj.getClass().getSimpleName();
        };
    }

    // Guarded patterns
    public String categorizeNumber(Object obj) {
        return switch (obj) {
            case Integer i when i < 0 -> "Negative integer";
            case Integer i when i == 0 -> "Zero";
            case Integer i when i > 0 && i <= 100 -> "Small positive integer";
            case Integer i -> "Large positive integer";
            case Double d when d.isNaN() -> "Not a number";
            case Double d when d.isInfinite() -> "Infinite double";
            case Double d -> "Finite double: " + d;
            case null -> "null value";
            default -> "Not a number type";
        };
    }

    // Pattern matching with sealed classes
    public sealed interface Response permits Success, Error, Loading {}
    public record Success(String data) implements Response {}
    public record Error(String message, int code) implements Response {}
    public record Loading() implements Response {}

    public String handleResponse(Response response) {
        return switch (response) {
            case Success(var data) -> "Got data: " + data;
            case Error(var message, var code) when code >= 500 ->
                "Server error (" + code + "): " + message;
            case Error(var message, var code) when code >= 400 ->
                "Client error (" + code + "): " + message;
            case Error(var message, var code) ->
                "Unknown error (" + code + "): " + message;
            case Loading() -> "Still loading...";
        };
    }

    // Complex nested pattern matching
    public record Point(int x, int y) {}
    public record Circle(Point center, int radius) {}
    public record Rectangle(Point topLeft, Point bottomRight) {}

    public String analyzeShape(Object shape) {
        return switch (shape) {
            case Circle(Point(var x, var y), var radius) when radius > 10 ->
                "Large circle at (" + x + ", " + y + ") with radius " + radius;
            case Circle(Point(var x, var y), var radius) ->
                "Small circle at (" + x + ", " + y + ") with radius " + radius;
            case Rectangle(Point(var x1, var y1), Point(var x2, var y2)) when
                Math.abs(x2 - x1) == Math.abs(y2 - y1) ->
                "Square from (" + x1 + ", " + y1 + ") to (" + x2 + ", " + y2 + ")";
            case Rectangle(Point(var x1, var y1), Point(var x2, var y2)) ->
                "Rectangle from (" + x1 + ", " + y1 + ") to (" + x2 + ", " + y2 + ")";
            default -> "Unknown shape";
        };
    }
}
```

### 🔧 **Tools and API Enhancements**

#### **Internet Address Resolution SPI**
Pluggable hostname resolution.

```java
// Custom hostname resolver
public class CustomAddressResolver extends InetAddressResolverProvider {

    @Override
    public InetAddressResolver get(Configuration configuration) {
        return new InetAddressResolver() {
            @Override
            public Stream<InetAddress> lookupByName(String host,
                    LookupPolicy lookupPolicy) throws UnknownHostException {

                // Custom resolution logic
                if ("custom.example.com".equals(host)) {
                    return Stream.of(InetAddress.getByName("127.0.0.1"));
                }

                // Fallback to system resolver
                return InetAddress.getAllByName(host);
            }

            @Override
            public String lookupByAddress(byte[] addr) throws UnknownHostException {
                // Custom reverse lookup logic
                return InetAddress.getByAddress(addr).getHostName();
            }
        };
    }

    @Override
    public String name() {
        return "CustomResolver";
    }
}

// Usage
public class NetworkExample {
    public static void testCustomResolver() {
        try {
            // Will use custom resolver if configured
            InetAddress address = InetAddress.getByName("custom.example.com");
            System.out.println("Resolved to: " + address.getHostAddress());
        } catch (UnknownHostException e) {
            e.printStackTrace();
        }
    }
}
```

#### **Deprecated finalization**
Finalization marked for future removal.

```java
// Finalization is deprecated - use alternatives
public class ResourceManagement {

    // ❌ Deprecated - don't use finalization
    @Deprecated(forRemoval = true)
    protected void finalize() throws Throwable {
        // This is unreliable and deprecated
        super.finalize();
    }

    // ✅ Better alternatives:

    // 1. try-with-resources
    public static class AutoCloseableResource implements AutoCloseable {
        @Override
        public void close() {
            System.out.println("Resource closed");
        }
    }

    // 2. Cleaner API
    private static final Cleaner cleaner = Cleaner.create();

    public static class CleanableResource {
        private final Cleaner.Cleanable cleanable;
        private final State state;

        private static class State implements Runnable {
            private boolean closed = false;

            @Override
            public void run() {
                if (!closed) {
                    // Cleanup logic here
                    System.out.println("Cleaner: Resource cleaned up");
                    closed = true;
                }
            }
        }

        public CleanableResource() {
            this.state = new State();
            this.cleanable = cleaner.register(this, state);
        }

        public void close() {
            if (!state.closed) {
                state.run();
                cleanable.clean();
            }
        }
    }

    // 3. Manual resource management
    public static void processWithManualCleanup() {
        SomeResource resource = new SomeResource();
        try {
            // Use resource
            resource.doWork();
        } finally {
            // Explicit cleanup
            resource.cleanup();
        }
    }
}
```

---

## Java 19 (September 2022)

**Major Theme**: Virtual threads preview, structured concurrency, and pattern matching refinements

### 🔥 Major Features

#### 1. **Virtual Threads (Preview)**
Lightweight threads for high-throughput concurrent applications.

```java
// Virtual threads - revolutionary concurrency model
public class VirtualThreadsExample {

    // Creating virtual threads
    public static void basicVirtualThreads() throws InterruptedException {
        // Create virtual thread
        Thread virtualThread = Thread.ofVirtual()
            .name("virtual-1")
            .start(() -> {
                System.out.println("Hello from virtual thread: " +
                    Thread.currentThread().getName());
                try {
                    Thread.sleep(1000);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                }
            });

        virtualThread.join();

        // Thread builder for multiple virtual threads
        Thread.Builder virtualBuilder = Thread.ofVirtual().name("worker-", 0);

        List<Thread> threads = IntStream.range(0, 10)
            .mapToObj(i -> virtualBuilder.start(() -> {
                System.out.println("Worker " + i + " on thread " +
                    Thread.currentThread().getName());
                try {
                    Thread.sleep(2000);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                }
            }))
            .toList();

        // Wait for all threads
        for (Thread t : threads) {
            t.join();
        }
    }

    // Virtual thread executor
    public static void virtualThreadExecutor() {
        try (ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor()) {

            // Submit many tasks - each gets its own virtual thread
            List<Future<String>> futures = IntStream.range(0, 10_000)
                .mapToObj(i -> executor.submit(() -> {
                    Thread.sleep(1000); // Simulate I/O
                    return "Task " + i + " completed by " +
                           Thread.currentThread().getName();
                }))
                .toList();

            // Collect results
            for (Future<String> future : futures) {
                System.out.println(future.get());
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    // HTTP client with virtual threads
    public static void webCrawlerExample() {
        List<String> urls = List.of(
            "https://example.com",
            "https://httpbin.org/delay/1",
            "https://httpbin.org/delay/2",
            "https://jsonplaceholder.typicode.com/posts/1"
        );

        HttpClient client = HttpClient.newHttpClient();

        try (ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor()) {
            List<CompletableFuture<String>> futures = urls.stream()
                .map(url -> CompletableFuture.supplyAsync(() -> {
                    try {
                        HttpRequest request = HttpRequest.newBuilder()
                            .uri(URI.create(url))
                            .build();
                        HttpResponse<String> response = client.send(request,
                            HttpResponse.BodyHandlers.ofString());
                        return "URL: " + url + " -> Status: " + response.statusCode() +
                               " (Thread: " + Thread.currentThread().getName() + ")";
                    } catch (Exception e) {
                        return "URL: " + url + " -> Error: " + e.getMessage();
                    }
                }, executor))
                .toList();

            futures.forEach(future -> {
                try {
                    System.out.println(future.get());
                } catch (Exception e) {
                    e.printStackTrace();
                }
            });
        }
    }

    // Virtual threads vs traditional threads comparison
    public static void performanceComparison() {
        int taskCount = 100_000;

        // Measure virtual threads
        long virtualStart = System.currentTimeMillis();
        try (ExecutorService virtualExecutor =
             Executors.newVirtualThreadPerTaskExecutor()) {

            CountDownLatch virtualLatch = new CountDownLatch(taskCount);
            for (int i = 0; i < taskCount; i++) {
                virtualExecutor.submit(() -> {
                    try {
                        Thread.sleep(100); // Simulate I/O
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                    } finally {
                        virtualLatch.countDown();
                    }
                });
            }
            virtualLatch.await();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        long virtualTime = System.currentTimeMillis() - virtualStart;

        // Measure platform threads (with reasonable pool size)
        long platformStart = System.currentTimeMillis();
        try (ExecutorService platformExecutor =
             Executors.newFixedThreadPool(200)) {

            CountDownLatch platformLatch = new CountDownLatch(taskCount);
            for (int i = 0; i < taskCount; i++) {
                platformExecutor.submit(() -> {
                    try {
                        Thread.sleep(100); // Simulate I/O
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                    } finally {
                        platformLatch.countDown();
                    }
                });
            }
            platformLatch.await();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        long platformTime = System.currentTimeMillis() - platformStart;

        System.out.printf("Virtual threads: %d ms%n", virtualTime);
        System.out.printf("Platform threads: %d ms%n", platformTime);
        System.out.printf("Virtual threads are %.2fx faster%n",
                         (double) platformTime / virtualTime);
    }
}
```

#### 2. **Structured Concurrency (Incubator)**
Organize concurrent tasks as a single unit of work.

```java
// Structured concurrency - organize related tasks
// Requires --add-modules jdk.incubator.concurrent

import jdk.incubator.concurrent.StructuredTaskScope;

public class StructuredConcurrencyExample {

    // Basic structured concurrency
    public static String fetchUserData(String userId) throws Exception {
        try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {

            // Launch concurrent subtasks
            Future<String> userProfile = scope.fork(() -> fetchUserProfile(userId));
            Future<List<String>> userOrders = scope.fork(() -> fetchUserOrders(userId));
            Future<String> userPreferences = scope.fork(() -> fetchUserPreferences(userId));

            // Wait for all subtasks to complete
            scope.join();           // Wait for all tasks
            scope.throwIfFailed();  // Throw if any task failed

            // All tasks completed successfully
            return combineUserData(
                userProfile.resultNow(),
                userOrders.resultNow(),
                userPreferences.resultNow()
            );
        }
    }

    // Shutdown on success - return first successful result
    public static String findFirstAvailableServer(List<String> servers) throws Exception {
        try (var scope = new StructuredTaskScope.ShutdownOnSuccess<String>()) {

            // Try all servers concurrently
            for (String server : servers) {
                scope.fork(() -> checkServerHealth(server));
            }

            // Wait for first success
            scope.join();
            return scope.result(); // First successful result
        }
    }

    // Custom structured task scope
    public static class CollectAllResults<T> extends StructuredTaskScope<T> {
        private final List<T> results = new CopyOnWriteArrayList<>();
        private final List<Exception> exceptions = new CopyOnWriteArrayList<>();

        @Override
        protected void handleComplete(Subtask<? extends T> subtask) {
            switch (subtask.state()) {
                case SUCCESS -> results.add(subtask.get());
                case FAILED -> exceptions.add((Exception) subtask.exception());
                default -> {} // UNAVAILABLE - task was cancelled
            }
        }

        public List<T> getResults() {
            return List.copyOf(results);
        }

        public List<Exception> getExceptions() {
            return List.copyOf(exceptions);
        }
    }

    // Using custom scope
    public static List<String> fetchAllUserData(List<String> userIds) throws Exception {
        try (var scope = new CollectAllResults<String>()) {

            // Launch tasks for all users
            for (String userId : userIds) {
                scope.fork(() -> fetchUserProfile(userId));
            }

            scope.join();

            // Get all successful results
            List<String> results = scope.getResults();
            List<Exception> errors = scope.getExceptions();

            if (!errors.isEmpty()) {
                System.out.println("Some tasks failed: " + errors.size());
                errors.forEach(e -> System.err.println("Error: " + e.getMessage()));
            }

            return results;
        }
    }

    // Complex example: Web service aggregation
    public record WeatherData(String location, double temperature, String condition) {}
    public record NewsData(String headline, String summary) {}
    public record StockData(String symbol, double price, double change) {}
    public record DashboardData(WeatherData weather, List<NewsData> news, List<StockData> stocks) {}

    public static DashboardData buildDashboard(String location, List<String> stockSymbols)
            throws Exception {
        try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {

            // Concurrent API calls
            Future<WeatherData> weather = scope.fork(() ->
                fetchWeatherData(location));
            Future<List<NewsData>> news = scope.fork(() ->
                fetchLatestNews());
            Future<List<StockData>> stocks = scope.fork(() ->
                fetchStockData(stockSymbols));

            // Wait for all data
            scope.join();
            scope.throwIfFailed();

            return new DashboardData(
                weather.resultNow(),
                news.resultNow(),
                stocks.resultNow()
            );
        }
    }

    // Helper methods (simulated)
    private static String fetchUserProfile(String userId) throws Exception {
        Thread.sleep(100); // Simulate network call
        return "Profile for " + userId;
    }

    private static List<String> fetchUserOrders(String userId) throws Exception {
        Thread.sleep(150); // Simulate network call
        return List.of("Order1", "Order2", "Order3");
    }

    private static String fetchUserPreferences(String userId) throws Exception {
        Thread.sleep(80); // Simulate network call
        return "Dark theme, English";
    }

    private static String combineUserData(String profile, List<String> orders, String prefs) {
        return String.format("User{profile='%s', orders=%s, preferences='%s'}",
                           profile, orders, prefs);
    }

    private static String checkServerHealth(String server) throws Exception {
        Thread.sleep(new Random().nextInt(1000)); // Simulate variable response time
        if (Math.random() > 0.3) { // 70% success rate
            return server + " is healthy";
        } else {
            throw new Exception(server + " is down");
        }
    }

    private static WeatherData fetchWeatherData(String location) throws Exception {
        Thread.sleep(200);
        return new WeatherData(location, 22.5, "Sunny");
    }

    private static List<NewsData> fetchLatestNews() throws Exception {
        Thread.sleep(300);
        return List.of(
            new NewsData("Tech News", "Latest technology updates"),
            new NewsData("World News", "Global events summary")
        );
    }

    private static List<StockData> fetchStockData(List<String> symbols) throws Exception {
        Thread.sleep(250);
        return symbols.stream()
            .map(symbol -> new StockData(symbol, Math.random() * 1000, Math.random() * 10 - 5))
            .toList();
    }
}
```

#### 3. **Pattern Matching for switch (Third Preview)**
Further refinements to pattern matching syntax.

```java
// Pattern matching for switch (third preview)
public class PatternMatchingRefinements {

    // Refined guard syntax
    public String processValue(Object value) {
        return switch (value) {
            case Integer i when i > 0 -> "Positive: " + i;
            case Integer i when i < 0 -> "Negative: " + i;
            case Integer i -> "Zero";
            case String s when s.isEmpty() -> "Empty string";
            case String s when s.length() == 1 -> "Single character: " + s;
            case String s -> "String: " + s.substring(0, Math.min(10, s.length()));
            case null -> "null";
            default -> "Unknown: " + value.getClass().getSimpleName();
        };
    }

    // Pattern matching with records and deconstruction
    public sealed interface JsonValue {}
    public record JsonString(String value) implements JsonValue {}
    public record JsonNumber(double value) implements JsonValue {}
    public record JsonBoolean(boolean value) implements JsonValue {}
    public record JsonArray(List<JsonValue> values) implements JsonValue {}
    public record JsonObject(Map<String, JsonValue> fields) implements JsonValue {}
    public record JsonNull() implements JsonValue {}

    public String formatJson(JsonValue json, int indent) {
        String spaces = " ".repeat(indent * 2);
        return switch (json) {
            case JsonString(var s) -> "\"" + s + "\"";
            case JsonNumber(var n) -> String.valueOf(n);
            case JsonBoolean(var b) -> String.valueOf(b);
            case JsonNull() -> "null";
            case JsonArray(var values) when values.isEmpty() -> "[]";
            case JsonArray(var values) -> {
                StringBuilder sb = new StringBuilder("[\n");
                for (int i = 0; i < values.size(); i++) {
                    sb.append(spaces).append("  ")
                      .append(formatJson(values.get(i), indent + 1));
                    if (i < values.size() - 1) sb.append(",");
                    sb.append("\n");
                }
                sb.append(spaces).append("]");
                yield sb.toString();
            }
            case JsonObject(var fields) when fields.isEmpty() -> "{}";
            case JsonObject(var fields) -> {
                StringBuilder sb = new StringBuilder("{\n");
                var entries = fields.entrySet().iterator();
                while (entries.hasNext()) {
                    var entry = entries.next();
                    sb.append(spaces).append("  \"")
                      .append(entry.getKey()).append("\": ")
                      .append(formatJson(entry.getValue(), indent + 1));
                    if (entries.hasNext()) sb.append(",");
                    sb.append("\n");
                }
                sb.append(spaces).append("}");
                yield sb.toString();
            }
        };
    }

    // Complex nested pattern matching
    public sealed interface Expression {}
    public record Literal(int value) implements Expression {}
    public record Variable(String name) implements Expression {}
    public record Binary(String operator, Expression left, Expression right) implements Expression {}
    public record Unary(String operator, Expression operand) implements Expression {}

    public int evaluate(Expression expr, Map<String, Integer> variables) {
        return switch (expr) {
            case Literal(var value) -> value;
            case Variable(var name) -> variables.getOrDefault(name, 0);
            case Binary("+", var left, var right) ->
                evaluate(left, variables) + evaluate(right, variables);
            case Binary("-", var left, var right) ->
                evaluate(left, variables) - evaluate(right, variables);
            case Binary("*", var left, var right) ->
                evaluate(left, variables) * evaluate(right, variables);
            case Binary("/", var left, var right) -> {
                int rightVal = evaluate(right, variables);
                if (rightVal == 0) throw new ArithmeticException("Division by zero");
                yield evaluate(left, variables) / rightVal;
            }
            case Binary(var op, var left, var right) ->
                throw new IllegalArgumentException("Unknown operator: " + op);
            case Unary("-", var operand) -> -evaluate(operand, variables);
            case Unary(var op, var operand) ->
                throw new IllegalArgumentException("Unknown unary operator: " + op);
        };
    }

    // Pattern matching for optimization
    public Expression optimize(Expression expr) {
        return switch (expr) {
            // Constant folding
            case Binary("+", Literal(var a), Literal(var b)) -> new Literal(a + b);
            case Binary("-", Literal(var a), Literal(var b)) -> new Literal(a - b);
            case Binary("*", Literal(var a), Literal(var b)) -> new Literal(a * b);
            case Binary("/", Literal(var a), Literal(var b)) when b != 0 ->
                new Literal(a / b);

            // Algebraic simplifications
            case Binary("+", var e, Literal(0)) -> optimize(e);
            case Binary("+", Literal(0), var e) -> optimize(e);
            case Binary("*", var e, Literal(1)) -> optimize(e);
            case Binary("*", Literal(1), var e) -> optimize(e);
            case Binary("*", var e, Literal(0)) -> new Literal(0);
            case Binary("*", Literal(0), var e) -> new Literal(0);

            // Recursive optimization
            case Binary(var op, var left, var right) ->
                new Binary(op, optimize(left), optimize(right));
            case Unary(var op, var operand) ->
                new Unary(op, optimize(operand));

            // Base cases
            default -> expr;
        };
    }
}
```

### 🛠 **Additional Features**

#### **Vector API (Fourth Incubator)**
Continued improvements to SIMD operations.

```java
// Vector API fourth incubator with performance optimizations
import jdk.incubator.vector.*;

public class VectorAPI4thIncubator {

    // Optimized image processing
    public static void blurImage(int[] input, int[] output, int width, int height, int blurRadius) {
        var species = IntVector.SPECIES_PREFERRED;

        for (int y = blurRadius; y < height - blurRadius; y++) {
            int loopBound = species.loopBound(width - 2 * blurRadius);

            for (int x = blurRadius; x < blurRadius + loopBound; x += species.length()) {
                var sum = IntVector.zero(species);
                var count = 0;

                // Accumulate values in blur radius
                for (int dy = -blurRadius; dy <= blurRadius; dy++) {
                    for (int dx = -blurRadius; dx <= blurRadius; dx++) {
                        int index = (y + dy) * width + x + dx;
                        if (index >= 0 && index < input.length - species.length()) {
                            var pixels = IntVector.fromArray(species, input, index);
                            sum = sum.add(pixels);
                            count++;
                        }
                    }
                }

                // Average and store result
                var avg = sum.div(count);
                avg.intoArray(output, y * width + x);
            }

            // Handle remaining pixels
            for (int x = blurRadius + loopBound; x < width - blurRadius; x++) {
                int sum = 0, count = 0;
                for (int dy = -blurRadius; dy <= blurRadius; dy++) {
                    for (int dx = -blurRadius; dx <= blurRadius; dx++) {
                        sum += input[(y + dy) * width + (x + dx)];
                        count++;
                    }
                }
                output[y * width + x] = sum / count;
            }
        }
    }

    // Vectorized string operations
    public static boolean vectorizedStringEquals(byte[] str1, byte[] str2) {
        if (str1.length != str2.length) return false;

        var species = ByteVector.SPECIES_PREFERRED;
        int loopBound = species.loopBound(str1.length);

        // Vectorized comparison
        for (int i = 0; i < loopBound; i += species.length()) {
            var vec1 = ByteVector.fromArray(species, str1, i);
            var vec2 = ByteVector.fromArray(species, str2, i);
            var mask = vec1.eq(vec2);

            if (!mask.allTrue()) {
                return false; // Found difference
            }
        }

        // Check remaining bytes
        for (int i = loopBound; i < str1.length; i++) {
            if (str1[i] != str2[i]) return false;
        }

        return true;
    }

    // Cryptographic operations
    public static void xorCipher(byte[] data, byte[] key, byte[] result) {
        var species = ByteVector.SPECIES_PREFERRED;
        int keyLen = key.length;
        int loopBound = species.loopBound(data.length);

        for (int i = 0; i < loopBound; i += species.length()) {
            var dataVec = ByteVector.fromArray(species, data, i);

            // Create repeating key pattern
            byte[] keyBlock = new byte[species.length()];
            for (int j = 0; j < species.length(); j++) {
                keyBlock[j] = key[(i + j) % keyLen];
            }
            var keyVec = ByteVector.fromArray(species, keyBlock, 0);

            // XOR operation
            var resultVec = dataVec.lanewise(VectorOperators.XOR, keyVec);
            resultVec.intoArray(result, i);
        }

        // Handle remaining bytes
        for (int i = loopBound; i < data.length; i++) {
            result[i] = (byte) (data[i] ^ key[i % keyLen]);
        }
    }
}
```

#### **Foreign Function & Memory API (Preview)**
Foreign function interface graduates from incubator.

```java
// Foreign Function & Memory API (preview in Java 19)
// Requires --enable-preview --add-modules java.base

import java.lang.foreign.*;

public class ForeignAPIPreview {

    // Memory layout definitions
    public static final GroupLayout POINT_LAYOUT = MemoryLayout.structLayout(
        ValueLayout.JAVA_INT.withName("x"),
        ValueLayout.JAVA_INT.withName("y")
    ).withName("Point");

    public static final VarHandle X_HANDLE = POINT_LAYOUT.varHandle(
        MemoryLayout.PathElement.groupElement("x"));
    public static final VarHandle Y_HANDLE = POINT_LAYOUT.varHandle(
        MemoryLayout.PathElement.groupElement("y"));

    // Structured memory access
    public static void structuredMemoryExample() {
        try (Arena arena = Arena.openConfined()) {
            // Allocate memory for point
            MemorySegment point = arena.allocate(POINT_LAYOUT);

            // Set values
            X_HANDLE.set(point, 0L, 42);
            Y_HANDLE.set(point, 0L, 84);

            // Read values
            int x = (int) X_HANDLE.get(point, 0L);
            int y = (int) Y_HANDLE.get(point, 0L);

            System.out.printf("Point: (%d, %d)%n", x, y);
        }
    }

    // Native library integration
    public static void nativeLibraryExample() {
        // Load native library
        SymbolLookup stdlib = Linker.nativeLinker().defaultLookup();

        // Get function symbols
        MemorySegment strlenSymbol = stdlib.find("strlen")
            .orElseThrow(() -> new RuntimeException("strlen not found"));

        // Create function descriptor
        FunctionDescriptor strlenDesc = FunctionDescriptor.of(
            ValueLayout.JAVA_LONG, ValueLayout.ADDRESS);

        // Get method handle
        Linker linker = Linker.nativeLinker();
        MethodHandle strlen = linker.downcallHandle(strlenSymbol, strlenDesc);

        try (Arena arena = Arena.openConfined()) {
            // Allocate C string
            MemorySegment cString = arena.allocateUtf8String("Hello, World!");

            // Call native function
            long length = (long) strlen.invoke(cString);
            System.out.println("String length: " + length);

        } catch (Throwable e) {
            e.printStackTrace();
        }
    }

    // Callback functions
    public static void callbackExample() {
        try (Arena arena = Arena.openConfined()) {
            // Define callback signature
            FunctionDescriptor callbackDesc = FunctionDescriptor.of(
                ValueLayout.JAVA_INT,
                ValueLayout.JAVA_INT,
                ValueLayout.JAVA_INT);

            // Create Java method handle
            MethodHandle javaCallback = MethodHandles.lookup()
                .findStatic(ForeignAPIPreview.class, "addNumbers",
                          MethodType.methodType(int.class, int.class, int.class));

            // Create upcall stub
            MemorySegment callbackStub = Linker.nativeLinker()
                .upcallStub(javaCallback, callbackDesc, arena);

            System.out.println("Created callback stub: " + callbackStub);

            // Use callback stub with native code
            // (This would typically be passed to a native function)

        } catch (NoSuchMethodException | IllegalAccessException e) {
            e.printStackTrace();
        }
    }

    // Callback method
    public static int addNumbers(int a, int b) {
        System.out.printf("Java callback called with %d + %d%n", a, b);
        return a + b;
    }

    // Array processing with native memory
    public static void arrayProcessingExample() {
        try (Arena arena = Arena.openConfined()) {
            int size = 1_000_000;

            // Allocate native arrays
            MemorySegment inputArray = arena.allocateArray(ValueLayout.JAVA_DOUBLE, size);
            MemorySegment outputArray = arena.allocateArray(ValueLayout.JAVA_DOUBLE, size);

            // Fill input array
            for (int i = 0; i < size; i++) {
                inputArray.setAtIndex(ValueLayout.JAVA_DOUBLE, i, Math.sin(i * 0.001));
            }

            // Process array (example: square all values)
            long startTime = System.nanoTime();
            for (int i = 0; i < size; i++) {
                double value = inputArray.getAtIndex(ValueLayout.JAVA_DOUBLE, i);
                outputArray.setAtIndex(ValueLayout.JAVA_DOUBLE, i, value * value);
            }
            long endTime = System.nanoTime();

            System.out.printf("Processed %d elements in %.2f ms%n",
                            size, (endTime - startTime) / 1_000_000.0);
        }
    }
}
```

---

## Java 20 (March 2023)

**Major Theme**: Scoped values, record patterns, and virtual threads improvements

### 🔥 Major Features

#### 1. **Scoped Values (Incubator)**
Efficient sharing of data across method calls within a thread.

```java
// Scoped values - better alternative to ThreadLocal
// Requires --add-modules jdk.incubator.concurrent

import jdk.incubator.concurrent.ScopedValue;

public class ScopedValuesExample {

    // Define scoped values
    private static final ScopedValue<String> USER_ID = ScopedValue.newInstance();
    private static final ScopedValue<String> REQUEST_ID = ScopedValue.newInstance();
    private static final ScopedValue<SecurityContext> SECURITY_CONTEXT = ScopedValue.newInstance();

    public static void main(String[] args) {
        // Handle incoming request
        handleRequest("user123", "req456", new SecurityContext("admin"));
    }

    // Handle request with scoped values
    public static void handleRequest(String userId, String requestId, SecurityContext secCtx) {
        // Bind scoped values for this request
        ScopedValue.where(USER_ID, userId)
                   .where(REQUEST_ID, requestId)
                   .where(SECURITY_CONTEXT, secCtx)
                   .run(() -> {
                       processRequest();
                   });
    }

    public static void processRequest() {
        // Access scoped values without parameter passing
        logInfo("Processing request");

        // Call business logic
        validateUser();
        processData();
        auditAction();
    }

    public static void validateUser() {
        String userId = USER_ID.get(); // No null checks needed
        SecurityContext security = SECURITY_CONTEXT.get();

        logInfo("Validating user: " + userId);

        if (!security.hasPermission("READ_DATA")) {
            throw new SecurityException("Access denied");
        }
    }

    public static void processData() {
        String userId = USER_ID.get();
        logInfo("Processing data for user: " + userId);

        // Complex processing...
        performCalculation();
    }

    public static void performCalculation() {
        // Deep in call stack, still have access to scoped values
        String userId = USER_ID.get();
        String requestId = REQUEST_ID.get();

        logInfo("Performing calculation");

        // Simulate work
        try {
            Thread.sleep(100);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    public static void auditAction() {
        String userId = USER_ID.get();
        String requestId = REQUEST_ID.get();
        SecurityContext security = SECURITY_CONTEXT.get();

        String auditMessage = String.format(
            "Action performed by user %s (request %s) with role %s",
            userId, requestId, security.getRole()
        );

        logInfo("AUDIT: " + auditMessage);
    }

    private static void logInfo(String message) {
        String userId = USER_ID.isBound() ? USER_ID.get() : "unknown";
        String requestId = REQUEST_ID.isBound() ? REQUEST_ID.get() : "unknown";

        System.out.printf("[%s][%s] %s%n", userId, requestId, message);
    }

    // Web service example with scoped values
    public static class WebService {
        private static final ScopedValue<HttpContext> HTTP_CONTEXT = ScopedValue.newInstance();

        public void handleHttpRequest(HttpRequest request) {
            HttpContext context = new HttpContext(request);

            ScopedValue.where(HTTP_CONTEXT, context)
                       .run(() -> {
                           try {
                               processHttpRequest();
                           } catch (Exception e) {
                               handleError(e);
                           }
                       });
        }

        private void processHttpRequest() {
            HttpContext context = HTTP_CONTEXT.get();

            // Route based on URL
            String path = context.getRequest().getPath();
            switch (path) {
                case "/users" -> handleUsersEndpoint();
                case "/orders" -> handleOrdersEndpoint();
                case "/products" -> handleProductsEndpoint();
                default -> handleNotFound();
            }
        }

        private void handleUsersEndpoint() {
            HttpContext context = HTTP_CONTEXT.get();

            // Access request data without parameter passing
            String method = context.getRequest().getMethod();
            Map<String, String> headers = context.getRequest().getHeaders();

            // Process based on method
            switch (method) {
                case "GET" -> getUsers();
                case "POST" -> createUser();
                case "PUT" -> updateUser();
                case "DELETE" -> deleteUser();
            }
        }

        private void getUsers() {
            // Deep method still has access to HTTP context
            HttpContext context = HTTP_CONTEXT.get();
            String userAgent = context.getRequest().getHeaders().get("User-Agent");

            logInfo("Getting users, User-Agent: " + userAgent);
        }

        private void handleError(Exception e) {
            HttpContext context = HTTP_CONTEXT.get();
            String requestId = context.getRequestId();

            System.err.printf("Error processing request %s: %s%n", requestId, e.getMessage());
        }

        // Placeholder methods
        private void handleOrdersEndpoint() { logInfo("Handling orders"); }
        private void handleProductsEndpoint() { logInfo("Handling products"); }
        private void handleNotFound() { logInfo("Not found"); }
        private void createUser() { logInfo("Creating user"); }
        private void updateUser() { logInfo("Updating user"); }
        private void deleteUser() { logInfo("Deleting user"); }
    }

    // Performance comparison: ScopedValue vs ThreadLocal
    public static void performanceComparison() {
        // ThreadLocal approach
        ThreadLocal<String> threadLocal = new ThreadLocal<>();

        long tlStart = System.nanoTime();
        for (int i = 0; i < 1_000_000; i++) {
            threadLocal.set("value" + i);
            String value = threadLocal.get();
            threadLocal.remove(); // Important for cleanup
        }
        long tlTime = System.nanoTime() - tlStart;

        // ScopedValue approach
        ScopedValue<String> scopedValue = ScopedValue.newInstance();

        long svStart = System.nanoTime();
        for (int i = 0; i < 1_000_000; i++) {
            final int index = i;
            ScopedValue.where(scopedValue, "value" + index)
                       .run(() -> {
                           String value = scopedValue.get();
                           // No manual cleanup needed
                       });
        }
        long svTime = System.nanoTime() - svStart;

        System.out.printf("ThreadLocal: %.2f ms%n", tlTime / 1_000_000.0);
        System.out.printf("ScopedValue: %.2f ms%n", svTime / 1_000_000.0);
        System.out.printf("ScopedValue is %.2fx %s%n",
                         Math.abs((double) tlTime / svTime),
                         svTime < tlTime ? "faster" : "slower");
    }

    // Supporting classes
    public static class SecurityContext {
        private final String role;

        public SecurityContext(String role) { this.role = role; }
        public String getRole() { return role; }
        public boolean hasPermission(String permission) {
            return "admin".equals(role) || "READ_DATA".equals(permission);
        }
    }

    public static class HttpContext {
        private final HttpRequest request;
        private final String requestId;

        public HttpContext(HttpRequest request) {
            this.request = request;
            this.requestId = java.util.UUID.randomUUID().toString();
        }

        public HttpRequest getRequest() { return request; }
        public String getRequestId() { return requestId; }
    }

    public static class HttpRequest {
        private final String method;
        private final String path;
        private final Map<String, String> headers;

        public HttpRequest(String method, String path, Map<String, String> headers) {
            this.method = method;
            this.path = path;
            this.headers = headers;
        }

        public String getMethod() { return method; }
        public String getPath() { return path; }
        public Map<String, String> getHeaders() { return headers; }
    }
}
```

#### 2. **Record Patterns (Second Preview)**
Pattern matching with record destructuring.

```java
// Record patterns - destructure records in pattern matching
public class RecordPatternsExample {

    // Basic records for examples
    public record Point(int x, int y) {}
    public record Circle(Point center, int radius) {}
    public record Rectangle(Point topLeft, Point bottomRight) {}
    public record Person(String name, int age, Address address) {}
    public record Address(String street, String city, String zipCode) {}

    // Basic record pattern matching
    public String describePoint(Object obj) {
        return switch (obj) {
            case Point(var x, var y) -> "Point at (" + x + ", " + y + ")";
            case null -> "null point";
            default -> "Not a point";
        };
    }

    // Nested record patterns
    public String describeShape(Object shape) {
        return switch (shape) {
            case Circle(Point(var x, var y), var radius) ->
                String.format("Circle at (%d, %d) with radius %d", x, y, radius);

            case Rectangle(Point(var x1, var y1), Point(var x2, var y2))
                when Math.abs(x2 - x1) == Math.abs(y2 - y1) ->
                String.format("Square from (%d, %d) to (%d, %d)", x1, y1, x2, y2);

            case Rectangle(Point(var x1, var y1), Point(var x2, var y2)) ->
                String.format("Rectangle from (%d, %d) to (%d, %d)", x1, y1, x2, y2);

            default -> "Unknown shape";
        };
    }

    // Complex nested patterns
    public String formatPersonInfo(Object obj) {
        return switch (obj) {
            case Person(var name, var age, Address(var street, var city, var zip))
                when age >= 18 ->
                String.format("Adult: %s, %d years old, lives at %s, %s %s",
                            name, age, street, city, zip);

            case Person(var name, var age, Address(var street, var city, var zip)) ->
                String.format("Minor: %s, %d years old, lives at %s, %s %s",
                            name, age, street, city, zip);

            case Person(var name, var age, null) ->
                String.format("Person: %s, %d years old, no address", name, age);

            default -> "Not a person";
        };
    }

    // Financial records example
    public record Money(long cents, String currency) {}
    public record Account(String id, String owner, Money balance) {}
    public record Transaction(String id, Account from, Account to, Money amount,
                            LocalDateTime timestamp) {}

    public String processTransaction(Transaction transaction) {
        return switch (transaction) {
            // Large transactions
            case Transaction(var id,
                           Account(var fromId, var fromOwner, Money(var fromBalance, var curr1)),
                           Account(var toId, var toOwner, Money(var toBalance, var curr2)),
                           Money(var amount, var transCurr),
                           var timestamp)
                when amount > 100000 && curr1.equals(curr2) && curr2.equals(transCurr) ->
                String.format("LARGE TRANSFER [%s]: %s -> %s, Amount: %.2f %s",
                            id, fromOwner, toOwner, amount / 100.0, transCurr);

            // Currency mismatch
            case Transaction(var id,
                           Account(var fromId, var fromOwner, Money(var fromBalance, var curr1)),
                           Account(var toId, var toOwner, Money(var toBalance, var curr2)),
                           Money(var amount, var transCurr),
                           var timestamp)
                when !curr1.equals(curr2) || !curr2.equals(transCurr) ->
                "CURRENCY MISMATCH: Transaction requires conversion";

            // Regular transaction
            case Transaction(var id,
                           Account(var fromId, var fromOwner, var fromBalance),
                           Account(var toId, var toOwner, var toBalance),
                           Money(var amount, var currency),
                           var timestamp) ->
                String.format("Transfer [%s]: %s -> %s, Amount: %.2f %s",
                            id, fromOwner, toOwner, amount / 100.0, currency);
        };
    }

    // Tree structure with record patterns
    public sealed interface TreeNode {}
    public record Leaf(int value) implements TreeNode {}
    public record Branch(TreeNode left, TreeNode right) implements TreeNode {}

    public int sumTree(TreeNode node) {
        return switch (node) {
            case Leaf(var value) -> value;
            case Branch(var left, var right) -> sumTree(left) + sumTree(right);
        };
    }

    public int maxDepth(TreeNode node) {
        return switch (node) {
            case Leaf(var value) -> 1;
            case Branch(var left, var right) -> 1 + Math.max(maxDepth(left), maxDepth(right));
        };
    }

    // Pattern matching for tree transformations
    public TreeNode optimizeTree(TreeNode node) {
        return switch (node) {
            // Optimize: Branch with two leaves
            case Branch(Leaf(var a), Leaf(var b)) when a == 0 -> new Leaf(b);
            case Branch(Leaf(var a), Leaf(var b)) when b == 0 -> new Leaf(a);
            case Branch(Leaf(var a), Leaf(var b)) when a == 1 -> new Leaf(b);
            case Branch(Leaf(var a), Leaf(var b)) when b == 1 -> new Leaf(a);

            // Recursive optimization
            case Branch(var left, var right) ->
                new Branch(optimizeTree(left), optimizeTree(right));

            // Leaves are already optimized
            case Leaf(var value) -> node;
        };
    }

    // HTTP request/response patterns
    public record HttpRequest(String method, String path, Map<String, String> headers, String body) {}
    public record HttpResponse(int status, Map<String, String> headers, String body) {}
    public record ApiEndpoint(String method, String path) {}

    public HttpResponse routeRequest(HttpRequest request) {
        return switch (request) {
            case HttpRequest("GET", var path, var headers, var body)
                when path.startsWith("/api/users/") ->
                handleGetUser(path.substring("/api/users/".length()));

            case HttpRequest("POST", "/api/users", var headers, var body)
                when headers.containsKey("Content-Type") &&
                     headers.get("Content-Type").equals("application/json") ->
                handleCreateUser(body);

            case HttpRequest("PUT", var path, var headers, var body)
                when path.matches("/api/users/\\d+") ->
                handleUpdateUser(extractUserId(path), body);

            case HttpRequest("DELETE", var path, var headers, var body)
                when path.matches("/api/users/\\d+") ->
                handleDeleteUser(extractUserId(path));

            case HttpRequest(var method, "/health", var headers, var body) ->
                new HttpResponse(200, Map.of("Content-Type", "text/plain"), "OK");

            default ->
                new HttpResponse(404, Map.of("Content-Type", "text/plain"), "Not Found");
        };
    }

    // Helper methods
    private HttpResponse handleGetUser(String userId) {
        return new HttpResponse(200, Map.of("Content-Type", "application/json"),
                              "{\"id\":\"" + userId + "\",\"name\":\"User " + userId + "\"}");
    }

    private HttpResponse handleCreateUser(String body) {
        return new HttpResponse(201, Map.of("Content-Type", "application/json"),
                              "{\"id\":\"new-user\",\"message\":\"User created\"}");
    }

    private HttpResponse handleUpdateUser(String userId, String body) {
        return new HttpResponse(200, Map.of("Content-Type", "application/json"),
                              "{\"id\":\"" + userId + "\",\"message\":\"User updated\"}");
    }

    private HttpResponse handleDeleteUser(String userId) {
        return new HttpResponse(204, Map.of(), "");
    }

    private String extractUserId(String path) {
        return path.substring(path.lastIndexOf('/') + 1);
    }
}
```

#### 3. **Virtual Threads (Second Preview)**
Improvements to virtual threads performance and debugging.

```java
// Virtual threads second preview with improvements
public class VirtualThreadsImprovements {

    // Improved virtual thread creation and management
    public static void demonstrateImprovements() throws Exception {
        // Better thread naming and grouping
        ThreadFactory virtualFactory = Thread.ofVirtual()
            .name("worker-pool-", 0)
            .factory();

        // Virtual thread executor with better resource management
        try (ExecutorService executor = Executors.newThreadPerTaskExecutor(virtualFactory)) {

            // Submit CPU-bound and I/O-bound tasks
            List<Future<String>> futures = new ArrayList<>();

            // I/O-bound tasks (virtual threads excel here)
            for (int i = 0; i < 1000; i++) {
                final int taskId = i;
                futures.add(executor.submit(() -> {
                    try {
                        // Simulate I/O operation
                        Thread.sleep(100 + (int)(Math.random() * 200));
                        return "I/O Task " + taskId + " completed on " +
                               Thread.currentThread().getName();
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                        return "Task " + taskId + " interrupted";
                    }
                }));
            }

            // CPU-bound tasks (still benefit from virtual threads for coordination)
            for (int i = 0; i < 100; i++) {
                final int taskId = i;
                futures.add(executor.submit(() -> {
                    // Simulate CPU work
                    long result = fibonacci(30);
                    return "CPU Task " + taskId + " result: " + result +
                           " on " + Thread.currentThread().getName();
                }));
            }

            // Collect results
            System.out.println("Processing " + futures.size() + " tasks...");
            long startTime = System.currentTimeMillis();

            for (Future<String> future : futures) {
                String result = future.get();
                if (futures.indexOf(future) < 5) { // Print first few
                    System.out.println(result);
                }
            }

            long endTime = System.currentTimeMillis();
            System.out.printf("All tasks completed in %d ms%n", endTime - startTime);
        }
    }

    // Virtual threads with structured concurrency improvements
    public static void structuredConcurrencyWithVirtualThreads() throws Exception {
        try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {

            // Submit subtasks that will run on virtual threads
            Future<String> userService = scope.fork(() -> callUserService());
            Future<String> orderService = scope.fork(() -> callOrderService());
            Future<String> inventoryService = scope.fork(() -> callInventoryService());
            Future<String> paymentService = scope.fork(() -> callPaymentService());

            // Join all tasks - virtual threads make this highly efficient
            scope.join();
            scope.throwIfFailed();

            // Combine results
            String result = String.format("Services called successfully: %s, %s, %s, %s",
                userService.resultNow(),
                orderService.resultNow(),
                inventoryService.resultNow(),
                paymentService.resultNow());

            System.out.println(result);
        }
    }

    // Web crawler using virtual threads
    public static void webCrawlerWithVirtualThreads() {
        List<String> urls = generateUrls(100); // Generate 100 URLs
        Set<String> visited = ConcurrentHashMap.newKeySet();
        Queue<String> toVisit = new ConcurrentLinkedQueue<>(urls);

        try (ExecutorService crawler = Executors.newVirtualThreadPerTaskExecutor()) {

            List<Future<Void>> crawlerTasks = new ArrayList<>();

            // Start multiple crawler workers
            for (int i = 0; i < 20; i++) {
                final int workerId = i;
                crawlerTasks.add(crawler.submit(() -> {
                    crawlWorker(workerId, toVisit, visited);
                    return null;
                }));
            }

            // Wait for all crawlers to complete
            for (Future<Void> task : crawlerTasks) {
                task.get();
            }

            System.out.printf("Crawling completed. Visited %d URLs%n", visited.size());
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    private static void crawlWorker(int workerId, Queue<String> toVisit, Set<String> visited) {
        HttpClient client = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(5))
            .build();

        String url;
        while ((url = toVisit.poll()) != null) {
            if (visited.contains(url)) continue;

            try {
                // Mark as visited early to avoid duplicates
                visited.add(url);

                HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(url))
                    .timeout(Duration.ofSeconds(10))
                    .build();

                HttpResponse<String> response = client.send(request,
                    HttpResponse.BodyHandlers.ofString());

                System.out.printf("Worker %d: %s -> %d (%s)%n",
                    workerId, url, response.statusCode(),
                    Thread.currentThread().getName());

                // Extract and queue new URLs (simplified)
                extractAndQueueUrls(response.body(), toVisit, visited);

                // Simulate processing time
                Thread.sleep(50);

            } catch (Exception e) {
                System.err.printf("Worker %d failed to crawl %s: %s%n",
                    workerId, url, e.getMessage());
            }
        }

        System.out.printf("Worker %d finished%n", workerId);
    }

    // Producer-consumer pattern with virtual threads
    public static void producerConsumerExample() throws Exception {
        BlockingQueue<String> queue = new LinkedBlockingQueue<>(1000);
        AtomicBoolean producersDone = new AtomicBoolean(false);

        try (ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor()) {

            // Start multiple producers
            List<Future<Void>> producers = new ArrayList<>();
            for (int i = 0; i < 5; i++) {
                final int producerId = i;
                producers.add(executor.submit(() -> {
                    produce(producerId, queue);
                    return null;
                }));
            }

            // Start multiple consumers
            List<Future<Void>> consumers = new ArrayList<>();
            for (int i = 0; i < 10; i++) {
                final int consumerId = i;
                consumers.add(executor.submit(() -> {
                    consume(consumerId, queue, producersDone);
                    return null;
                }));
            }

            // Wait for producers to finish
            for (Future<Void> producer : producers) {
                producer.get();
            }
            producersDone.set(true);

            // Wait for consumers to finish
            for (Future<Void> consumer : consumers) {
                consumer.get();
            }

            System.out.println("Producer-consumer example completed");
        }
    }

    private static void produce(int producerId, BlockingQueue<String> queue) {
        try {
            for (int i = 0; i < 100; i++) {
                String item = String.format("Item-%d-%d", producerId, i);
                queue.put(item);
                System.out.printf("Producer %d produced: %s%n", producerId, item);
                Thread.sleep(10); // Simulate work
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    private static void consume(int consumerId, BlockingQueue<String> queue,
                              AtomicBoolean producersDone) {
        try {
            while (!producersDone.get() || !queue.isEmpty()) {
                String item = queue.poll(100, TimeUnit.MILLISECONDS);
                if (item != null) {
                    System.out.printf("Consumer %d consumed: %s%n", consumerId, item);
                    Thread.sleep(20); // Simulate processing
                }
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    // Helper methods
    private static long fibonacci(int n) {
        if (n <= 1) return n;
        return fibonacci(n - 1) + fibonacci(n - 2);
    }

    private static String callUserService() throws Exception {
        Thread.sleep(100);
        return "UserService:OK";
    }

    private static String callOrderService() throws Exception {
        Thread.sleep(150);
        return "OrderService:OK";
    }

    private static String callInventoryService() throws Exception {
        Thread.sleep(120);
        return "InventoryService:OK";
    }

    private static String callPaymentService() throws Exception {
        Thread.sleep(200);
        return "PaymentService:OK";
    }

    private static List<String> generateUrls(int count) {
        return IntStream.range(1, count + 1)
            .mapToObj(i -> "https://httpbin.org/delay/" + (i % 3))
            .collect(Collectors.toList());
    }

    private static void extractAndQueueUrls(String content, Queue<String> toVisit,
                                          Set<String> visited) {
        // Simplified URL extraction
        if (content.contains("httpbin")) {
            String newUrl = "https://httpbin.org/status/200";
            if (!visited.contains(newUrl)) {
                toVisit.offer(newUrl);
            }
        }
    }
}
```

### 🛠 **Additional Enhancements**

#### **Vector API (Fifth Incubator)**
Performance improvements and new operations.

#### **Foreign Function & Memory API (Second Preview)**
Enhanced memory management and performance optimizations.

---

## Java 21 (September 2023) - LTS

**Major Theme**: Long-term support with virtual threads, pattern matching, and sequenced collections

### 🔥 Major Features

#### 1. **Virtual Threads (Final)**
Virtual threads become a standard feature with production-ready performance.

```java
// Virtual threads are now standard and production-ready
public class VirtualThreadsProduction {

    // High-throughput server simulation
    public static void simulateHighThroughputServer() throws Exception {
        ServerSocket serverSocket = new ServerSocket(8080);
        System.out.println("Server started on port 8080");

        // Use virtual threads for connection handling
        try (ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor()) {

            while (true) {
                Socket clientSocket = serverSocket.accept();

                // Each connection gets its own virtual thread
                executor.submit(() -> handleClient(clientSocket));
            }
        }
    }

    private static void handleClient(Socket clientSocket) {
        try (BufferedReader in = new BufferedReader(
                new InputStreamReader(clientSocket.getInputStream()));
             PrintWriter out = new PrintWriter(
                clientSocket.getOutputStream(), true)) {

            String inputLine;
            while ((inputLine = in.readLine()) != null) {
                // Simulate processing with I/O operations
                Thread.sleep(100); // Database query simulation

                String response = processRequest(inputLine);
                out.println(response);

                if ("bye".equalsIgnoreCase(inputLine)) {
                    break;
                }
            }
        } catch (Exception e) {
            System.err.println("Error handling client: " + e.getMessage());
        } finally {
            try {
                clientSocket.close();
            } catch (Exception e) {
                // Log error
            }
        }
    }

    // Reactive programming with virtual threads
    public static void reactiveStreamProcessing() {
        // Generate continuous stream of data
        BlockingQueue<String> dataStream = new LinkedBlockingQueue<>();

        // Producer
        Thread.ofVirtual().name("data-producer").start(() -> {
            int counter = 0;
            while (counter < 10000) {
                try {
                    dataStream.put("data-" + counter++);
                    Thread.sleep(1); // High frequency data
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    break;
                }
            }
        });

        // Multiple processing pipelines
        try (ExecutorService processor = Executors.newVirtualThreadPerTaskExecutor()) {

            // Pipeline 1: Real-time analytics
            processor.submit(() -> {
                while (true) {
                    try {
                        String data = dataStream.poll(1, TimeUnit.SECONDS);
                        if (data == null) break;

                        // Complex analytics
                        analyzeDataRealTime(data);
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                        break;
                    }
                }
            });

            // Pipeline 2: Data enrichment
            processor.submit(() -> {
                while (true) {
                    try {
                        String data = dataStream.poll(1, TimeUnit.SECONDS);
                        if (data == null) break;

                        // Enrich data with external services
                        enrichData(data);
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                        break;
                    }
                }
            });

            // Pipeline 3: Batch processing
            processor.submit(() -> {
                List<String> batch = new ArrayList<>();
                while (true) {
                    try {
                        String data = dataStream.poll(1, TimeUnit.SECONDS);
                        if (data == null) {
                            if (!batch.isEmpty()) {
                                processBatch(batch);
                                batch.clear();
                            }
                            break;
                        }

                        batch.add(data);
                        if (batch.size() >= 100) {
                            processBatch(new ArrayList<>(batch));
                            batch.clear();
                        }
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                        break;
                    }
                }
            });

            Thread.sleep(5000); // Let it run for 5 seconds
        }
    }

    // Microservices communication with virtual threads
    public static void microservicesOrchestration(String orderId) throws Exception {
        record ServiceResult(String service, String result) {}

        try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {

            // Call multiple microservices concurrently
            Future<ServiceResult> userService = executor.submit(() -> {
                try {
                    Thread.sleep(150); // Network call simulation
                    return new ServiceResult("UserService",
                        callExternalService("http://user-service/api/user/" + orderId));
                } catch (Exception e) {
                    return new ServiceResult("UserService", "Error: " + e.getMessage());
                }
            });

            Future<ServiceResult> inventoryService = executor.submit(() -> {
                try {
                    Thread.sleep(200); // Network call simulation
                    return new ServiceResult("InventoryService",
                        callExternalService("http://inventory-service/api/check/" + orderId));
                } catch (Exception e) {
                    return new ServiceResult("InventoryService", "Error: " + e.getMessage());
                }
            });

            Future<ServiceResult> paymentService = executor.submit(() -> {
                try {
                    Thread.sleep(300); // Network call simulation
                    return new ServiceResult("PaymentService",
                        callExternalService("http://payment-service/api/authorize/" + orderId));
                } catch (Exception e) {
                    return new ServiceResult("PaymentService", "Error: " + e.getMessage());
                }
            });

            Future<ServiceResult> shippingService = executor.submit(() -> {
                try {
                    Thread.sleep(100); // Network call simulation
                    return new ServiceResult("ShippingService",
                        callExternalService("http://shipping-service/api/calculate/" + orderId));
                } catch (Exception e) {
                    return new ServiceResult("ShippingService", "Error: " + e.getMessage());
                }
            });

            // Collect all results
            List<ServiceResult> results = List.of(
                userService.get(),
                inventoryService.get(),
                paymentService.get(),
                shippingService.get()
            );

            // Process combined results
            System.out.println("Microservices Results for Order " + orderId + ":");
            results.forEach(result ->
                System.out.printf("  %s: %s%n", result.service(), result.result()));
        }
    }

    // Helper methods
    private static String processRequest(String request) {
        return "Processed: " + request.toUpperCase();
    }

    private static void analyzeDataRealTime(String data) {
        // Simulate real-time analytics
        try {
            Thread.sleep(10);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    private static void enrichData(String data) {
        // Simulate data enrichment
        try {
            Thread.sleep(50);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    private static void processBatch(List<String> batch) {
        System.out.printf("Processing batch of %d items%n", batch.size());
        try {
            Thread.sleep(200); // Batch processing time
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    private static String callExternalService(String url) {
        // Simulate HTTP call
        return "Response from " + url;
    }
}
```

#### 2. **Sequenced Collections**
New collection interfaces with well-defined encounter order.

```java
// Sequenced Collections - ordered access to collection elements
import java.util.*;

public class SequencedCollectionsExample {

    public static void demonstrateSequencedList() {
        // SequencedList extends List
        SequencedList<String> list = new ArrayList<>();

        // Add elements
        list.add("first");
        list.add("second");
        list.add("third");

        // New sequenced operations
        System.out.println("First: " + list.getFirst()); // "first"
        System.out.println("Last: " + list.getLast());   // "third"

        // Add to beginning and end
        list.addFirst("zero");
        list.addLast("fourth");

        System.out.println("List: " + list); // [zero, first, second, third, fourth]

        // Remove from beginning and end
        String removedFirst = list.removeFirst(); // "zero"
        String removedLast = list.removeLast();   // "fourth"

        System.out.println("Removed first: " + removedFirst);
        System.out.println("Removed last: " + removedLast);
        System.out.println("Final list: " + list); // [first, second, third]

        // Reversed view
        SequencedList<String> reversed = list.reversed();
        System.out.println("Reversed: " + reversed); // [third, second, first]

        // Modifying reversed view affects original
        reversed.addFirst("new-last");
        System.out.println("Original after reverse modification: " + list);
        // [first, second, third, new-last]
    }

    public static void demonstrateSequencedSet() {
        // SequencedSet extends Set
        SequencedSet<String> set = new LinkedHashSet<>();

        set.add("apple");
        set.add("banana");
        set.add("cherry");

        // First and last elements
        System.out.println("First: " + set.getFirst()); // "apple"
        System.out.println("Last: " + set.getLast());   // "cherry"

        // Add to beginning (if not already present)
        set.addFirst("apricot");
        set.addLast("date");

        System.out.println("Set: " + set); // [apricot, apple, banana, cherry, date]

        // Remove from ends
        set.removeFirst(); // removes "apricot"
        set.removeLast();  // removes "date"

        System.out.println("After removal: " + set); // [apple, banana, cherry]

        // Reversed view
        SequencedSet<String> reversedSet = set.reversed();
        System.out.println("Reversed set: " + reversedSet); // [cherry, banana, apple]
    }

    public static void demonstrateSequencedMap() {
        // SequencedMap extends Map
        SequencedMap<String, Integer> map = new LinkedHashMap<>();

        map.put("first", 1);
        map.put("second", 2);
        map.put("third", 3);

        // Access first and last entries
        Map.Entry<String, Integer> firstEntry = map.firstEntry();
        Map.Entry<String, Integer> lastEntry = map.lastEntry();

        System.out.println("First entry: " + firstEntry); // first=1
        System.out.println("Last entry: " + lastEntry);   // third=3

        // Remove first and last entries
        map.pollFirstEntry(); // removes and returns first entry
        map.pollLastEntry();  // removes and returns last entry

        System.out.println("After polling: " + map); // {second=2}

        // Add more entries
        map.put("alpha", 100);
        map.put("beta", 200);
        map.put("gamma", 300);

        // Reversed view
        SequencedMap<String, Integer> reversedMap = map.reversed();
        System.out.println("Reversed map: " + reversedMap);
        // {gamma=300, beta=200, alpha=100, second=2}

        // Modifying reversed view
        reversedMap.put("omega", 999); // Adds to end of original
        System.out.println("Original after reverse modification: " + map);
        // {second=2, alpha=100, beta=200, gamma=300, omega=999}
    }

    // Practical example: LRU Cache using SequencedMap
    public static class LRUCache<K, V> {
        private final int capacity;
        private final SequencedMap<K, V> cache = new LinkedHashMap<>();

        public LRUCache(int capacity) {
            this.capacity = capacity;
        }

        public V get(K key) {
            V value = cache.remove(key); // Remove from current position
            if (value != null) {
                cache.put(key, value); // Add to end (most recent)
            }
            return value;
        }

        public void put(K key, V value) {
            cache.remove(key); // Remove if already exists

            // Check capacity
            if (cache.size() >= capacity) {
                cache.pollFirstEntry(); // Remove least recently used
            }

            cache.put(key, value); // Add as most recent
        }

        public void printCache() {
            System.out.println("Cache (LRU -> MRU): " + cache);
        }

        // New sequenced operations make this easier
        public K getLeastRecentlyUsed() {
            Map.Entry<K, V> first = cache.firstEntry();
            return first != null ? first.getKey() : null;
        }

        public K getMostRecentlyUsed() {
            Map.Entry<K, V> last = cache.lastEntry();
            return last != null ? last.getKey() : null;
        }
    }

    public static void demonstrateLRUCache() {
        LRUCache<String, String> cache = new LRUCache<>(3);

        cache.put("A", "Value A");
        cache.put("B", "Value B");
        cache.put("C", "Value C");
        cache.printCache(); // {A=Value A, B=Value B, C=Value C}

        // Access A (moves to end)
        cache.get("A");
        cache.printCache(); // {B=Value B, C=Value C, A=Value A}

        // Add D (evicts B)
        cache.put("D", "Value D");
        cache.printCache(); // {C=Value C, A=Value A, D=Value D}

        System.out.println("LRU: " + cache.getLeastRecentlyUsed()); // C
        System.out.println("MRU: " + cache.getMostRecentlyUsed());  // D
    }

    // Practical example: Navigation history
    public static class NavigationHistory<T> {
        private final SequencedList<T> history = new ArrayList<>();
        private final int maxSize;

        public NavigationHistory(int maxSize) {
            this.maxSize = maxSize;
        }

        public void visit(T item) {
            // Remove if already in history
            history.remove(item);

            // Add to end
            history.addLast(item);

            // Maintain max size
            while (history.size() > maxSize) {
                history.removeFirst();
            }
        }

        public T goBack() {
            if (history.size() <= 1) return null;

            history.removeLast(); // Remove current
            return history.getLast(); // Return previous
        }

        public T getCurrentPage() {
            return history.isEmpty() ? null : history.getLast();
        }

        public List<T> getRecentHistory(int count) {
            SequencedList<T> reversed = history.reversed();
            return reversed.stream().limit(count).toList();
        }

        public void printHistory() {
            System.out.println("History (oldest -> newest): " + history);
        }
    }

    public static void demonstrateNavigationHistory() {
        NavigationHistory<String> nav = new NavigationHistory<>(5);

        nav.visit("Home");
        nav.visit("Products");
        nav.visit("About");
        nav.visit("Contact");
        nav.printHistory(); // [Home, Products, About, Contact]

        System.out.println("Current: " + nav.getCurrentPage()); // Contact

        String previous = nav.goBack();
        System.out.println("Went back to: " + previous); // About
        nav.printHistory(); // [Home, Products, About]

        // Recent history
        List<String> recent = nav.getRecentHistory(3);
        System.out.println("Recent (newest first): " + recent); // [About, Products, Home]
    }
}
```

#### 3. **Pattern Matching for switch and Record Patterns (Final)**
Pattern matching becomes standard with full functionality.

```java
// Pattern matching for switch and record patterns are now final
public class PatternMatchingFinal {

    // Comprehensive pattern matching examples
    public sealed interface JsonValue {}
    public record JsonObject(Map<String, JsonValue> fields) implements JsonValue {}
    public record JsonArray(List<JsonValue> elements) implements JsonValue {}
    public record JsonString(String value) implements JsonValue {}
    public record JsonNumber(double value) implements JsonValue {}
    public record JsonBoolean(boolean value) implements JsonValue {}
    public record JsonNull() implements JsonValue {}

    // Advanced JSON processing with pattern matching
    public static String formatJson(JsonValue json, int indentLevel) {
        String indent = "  ".repeat(indentLevel);

        return switch (json) {
            case JsonString(var s) -> "\"" + escapeString(s) + "\"";
            case JsonNumber(var n) -> formatNumber(n);
            case JsonBoolean(var b) -> String.valueOf(b);
            case JsonNull() -> "null";

            case JsonArray(var elements) when elements.isEmpty() -> "[]";
            case JsonArray(var elements) -> {
                StringBuilder sb = new StringBuilder("[\n");
                for (int i = 0; i < elements.size(); i++) {
                    sb.append(indent).append("  ")
                      .append(formatJson(elements.get(i), indentLevel + 1));
                    if (i < elements.size() - 1) sb.append(",");
                    sb.append("\n");
                }
                yield sb.append(indent).append("]").toString();
            }

            case JsonObject(var fields) when fields.isEmpty() -> "{}";
            case JsonObject(var fields) -> {
                StringBuilder sb = new StringBuilder("{\n");
                var entries = fields.entrySet().toArray(Map.Entry[]::new);
                for (int i = 0; i < entries.length; i++) {
                    var entry = entries[i];
                    sb.append(indent).append("  \"")
                      .append(escapeString((String) entry.getKey()))
                      .append("\": ")
                      .append(formatJson((JsonValue) entry.getValue(), indentLevel + 1));
                    if (i < entries.length - 1) sb.append(",");
                    sb.append("\n");
                }
                yield sb.append(indent).append("}").toString();
            }
        };
    }

    // Complex nested pattern matching for data validation
    public static ValidationResult validateUser(JsonValue json) {
        return switch (json) {
            case JsonObject(var fields) when
                fields.containsKey("name") &&
                fields.containsKey("email") &&
                fields.containsKey("age") -> {

                var nameValidation = validateName(fields.get("name"));
                var emailValidation = validateEmail(fields.get("email"));
                var ageValidation = validateAge(fields.get("age"));

                yield ValidationResult.combine(nameValidation, emailValidation, ageValidation);
            }

            case JsonObject(var fields) ->
                ValidationResult.error("Missing required fields: name, email, age");

            case null -> ValidationResult.error("User data cannot be null");

            default -> ValidationResult.error("User data must be an object");
        };
    }

    private static ValidationResult validateName(JsonValue nameValue) {
        return switch (nameValue) {
            case JsonString(var name) when name.trim().length() >= 2 ->
                ValidationResult.success();
            case JsonString(var name) ->
                ValidationResult.error("Name must be at least 2 characters");
            case null ->
                ValidationResult.error("Name cannot be null");
            default ->
                ValidationResult.error("Name must be a string");
        };
    }

    private static ValidationResult validateEmail(JsonValue emailValue) {
        return switch (emailValue) {
            case JsonString(var email) when email.contains("@") && email.contains(".") ->
                ValidationResult.success();
            case JsonString(var email) ->
                ValidationResult.error("Invalid email format");
            case null ->
                ValidationResult.error("Email cannot be null");
            default ->
                ValidationResult.error("Email must be a string");
        };
    }

    private static ValidationResult validateAge(JsonValue ageValue) {
        return switch (ageValue) {
            case JsonNumber(var age) when age >= 0 && age <= 150 ->
                ValidationResult.success();
            case JsonNumber(var age) ->
                ValidationResult.error("Age must be between 0 and 150");
            case null ->
                ValidationResult.error("Age cannot be null");
            default ->
                ValidationResult.error("Age must be a number");
        };
    }

    // Database query result pattern matching
    public record User(String name, String email, int age) {}
    public record Order(String id, String userId, List<OrderItem> items, double total) {}
    public record OrderItem(String productId, int quantity, double price) {}

    public sealed interface QueryResult {}
    public record Success(Object data) implements QueryResult {}
    public record NotFound() implements QueryResult {}
    public record DatabaseError(String message) implements QueryResult {}
    public record ValidationError(List<String> errors) implements QueryResult {}

    public static String handleQueryResult(QueryResult result) {
        return switch (result) {
            case Success(User(var name, var email, var age)) ->
                String.format("User found: %s (%s), age %d", name, email, age);

            case Success(Order(var id, var userId, var items, var total)) when items.size() > 5 ->
                String.format("Large order %s for user %s: %d items, total $%.2f",
                            id, userId, items.size(), total);

            case Success(Order(var id, var userId, var items, var total)) ->
                String.format("Order %s for user %s: %d items, total $%.2f",
                            id, userId, items.size(), total);

            case Success(var data) ->
                "Query successful: " + data.toString();

            case NotFound() ->
                "No records found";

            case DatabaseError(var message) ->
                "Database error: " + message;

            case ValidationError(var errors) when errors.size() == 1 ->
                "Validation error: " + errors.get(0);

            case ValidationError(var errors) ->
                "Multiple validation errors: " + String.join(", ", errors);
        };
    }

    // HTTP request routing with pattern matching
    public record HttpRequest(String method, String path, Map<String, String> params) {}
    public record Route(String method, String pattern) {}

    public static String routeRequest(HttpRequest request) {
        return switch (request) {
            // REST API endpoints
            case HttpRequest("GET", var path, var params) when path.equals("/api/users") ->
                handleGetUsers(params);

            case HttpRequest("GET", var path, var params) when path.matches("/api/users/\\d+") ->
                handleGetUser(extractId(path), params);

            case HttpRequest("POST", "/api/users", var params) ->
                handleCreateUser(params);

            case HttpRequest("PUT", var path, var params) when path.matches("/api/users/\\d+") ->
                handleUpdateUser(extractId(path), params);

            case HttpRequest("DELETE", var path, var params) when path.matches("/api/users/\\d+") ->
                handleDeleteUser(extractId(path));

            // Special endpoints
            case HttpRequest("GET", "/health", var params) ->
                "{\"status\":\"healthy\",\"timestamp\":\"" + java.time.Instant.now() + "\"}";

            case HttpRequest("GET", "/metrics", var params) when params.containsKey("auth") ->
                generateMetrics();

            // Default cases
            case HttpRequest("OPTIONS", var path, var params) ->
                "Allow: GET, POST, PUT, DELETE, OPTIONS";

            case HttpRequest(var method, var path, var params) ->
                String.format("404 Not Found: %s %s", method, path);
        };
    }

    // Helper methods and classes
    private static String escapeString(String s) {
        return s.replace("\"", "\\\"").replace("\n", "\\n").replace("\t", "\\t");
    }

    private static String formatNumber(double n) {
        return n == (long) n ? String.valueOf((long) n) : String.valueOf(n);
    }

    private static String extractId(String path) {
        return path.substring(path.lastIndexOf('/') + 1);
    }

    private static String handleGetUsers(Map<String, String> params) {
        return "GET /api/users with params: " + params;
    }

    private static String handleGetUser(String id, Map<String, String> params) {
        return "GET /api/users/" + id + " with params: " + params;
    }

    private static String handleCreateUser(Map<String, String> params) {
        return "POST /api/users with params: " + params;
    }

    private static String handleUpdateUser(String id, Map<String, String> params) {
        return "PUT /api/users/" + id + " with params: " + params;
    }

    private static String handleDeleteUser(String id) {
        return "DELETE /api/users/" + id;
    }

    private static String generateMetrics() {
        return "{\"requests\":1000,\"errors\":5,\"uptime\":\"24h\"}";
    }

    public static class ValidationResult {
        private final boolean success;
        private final List<String> errors;

        private ValidationResult(boolean success, List<String> errors) {
            this.success = success;
            this.errors = errors;
        }

        public static ValidationResult success() {
            return new ValidationResult(true, List.of());
        }

        public static ValidationResult error(String error) {
            return new ValidationResult(false, List.of(error));
        }

        public static ValidationResult combine(ValidationResult... results) {
            List<String> allErrors = new ArrayList<>();
            boolean allSuccess = true;

            for (ValidationResult result : results) {
                allErrors.addAll(result.errors);
                allSuccess = allSuccess && result.success;
            }

            return new ValidationResult(allSuccess, allErrors);
        }

        public boolean isSuccess() { return success; }
        public List<String> getErrors() { return errors; }
    }
}
```

---

## Java 22 (March 2024) - Modern Concurrency and Performance

### Major Theme: Continued Virtual Thread Refinement and Language Modernization

Java 22 continued building on the foundation laid by Java 21, with focus on performance optimization, language refinement, and developer experience improvements.

### Key Features

#### 1. **Unnamed Variables & Patterns (Preview - JEP 456)**
A quality-of-life improvement for developers to indicate unused variables or patterns explicitly, improving code clarity and IDE support.

```java
// Traditional approach with unused variables
for (Order order : orders) {
    if (order instanceof PriorityOrder po) {
        // Only need the pattern check, not the bound variable
        processPriorityOrder();
    }
}

try (var resource = createResource()) {
    // resource not used in this example
    performOperation();
} catch (Exception e) {
    // Exception details not needed
    logError();
}

// With unnamed variables - clearer intent
for (var _ : orders) {
    // Explicitly showing we don't use the variable
    processOrder();
}

try (var _ = createResource()) {
    performOperation();
} catch (Exception _) {
    logError();
}

// Pattern matching with unnamed patterns
switch (shape) {
    case Circle(var _, double radius) ->
        calculateCircleArea(radius); // Don't need center point
    case Rectangle(var _, var _, double width, double height) ->
        calculateRectangleArea(width, height); // Don't need position
}

// Lambda expressions
stream.forEach(_ -> incrementCounter()); // Clearly unused parameter
map.forEach((_, value) -> processValue(value)); // Only need the value
```

#### 2. **Foreign Function & Memory API (Preview - JEP 454)**
Continued evolution of Panama Project for safe and efficient access to foreign functions and memory.

```java
import java.lang.foreign.*;
import java.lang.foreign.MemoryLayout.*;
import static java.lang.foreign.ValueLayout.*;

// Memory allocation and management
try (Arena arena = Arena.ofConfined()) {
    // Allocate native memory
    MemorySegment segment = arena.allocate(1024);

    // Write data to memory
    segment.setAtIndex(JAVA_INT, 0, 42);
    segment.setAtIndex(JAVA_DOUBLE, 1, 3.14159);

    // Read data back
    int value = segment.getAtIndex(JAVA_INT, 0);
    double pi = segment.getAtIndex(JAVA_DOUBLE, 1);
}

// Define a C struct layout
MemoryLayout POINT_LAYOUT = structLayout(
    JAVA_DOUBLE.withName("x"),
    JAVA_DOUBLE.withName("y")
);

// Foreign function calls
Linker linker = Linker.nativeLinker();
SymbolLookup stdlib = linker.defaultLookup();

// Find native function
MemorySegment strlen = stdlib.find("strlen").orElseThrow();

// Create function descriptor
FunctionDescriptor descriptor = FunctionDescriptor.of(JAVA_LONG, ADDRESS);

// Create method handle
MethodHandle strlenHandle = linker.downcallHandle(strlen, descriptor);

// Call native function
try (Arena arena = Arena.ofConfined()) {
    MemorySegment cString = arena.allocateUtf8String("Hello World");
    long length = (long) strlenHandle.invokeExact(cString);
    System.out.println("String length: " + length); // Output: 11
}
```

#### 3. **Stream Gatherers (Preview - JEP 461)**
New intermediate stream operations that provide more flexibility than existing operations.

```java
import static java.util.stream.Gatherers.*;

List<String> words = List.of("apple", "banana", "cherry", "date", "elderberry");

// Fixed-size windowing
List<List<String>> windowedWords = words.stream()
    .gather(windowFixed(3))
    .toList();
// Result: [[apple, banana, cherry], [date, elderberry]]

// Sliding window
List<List<String>> slidingWindows = words.stream()
    .gather(windowSliding(3))
    .toList();
// Result: [[apple, banana, cherry], [banana, cherry, date], [cherry, date, elderberry]]

// Fold operation (similar to reduce but more flexible)
List<Integer> numbers = List.of(1, 2, 3, 4, 5);
List<Integer> cumulativeSum = numbers.stream()
    .gather(fold(() -> 0, Integer::sum))
    .toList();
// Result: [1, 3, 6, 10, 15]
```

#### 4. **Statements before super() (Preview - JEP 447)**
Allow statements that don't reference the instance being constructed to appear before explicit constructor invocations.

```java
// With Java 22 preview feature
public class ImprovedValidation extends Parent {
    private final String name;
    private final int age;

    public ImprovedValidation(String name, int age) {
        // Validation can now happen before super()
        Objects.requireNonNull(name, "Name cannot be null");
        if (age < 0) {
            throw new IllegalArgumentException("Age must be non-negative");
        }

        // Prepare data for parent constructor
        String processedName = name.trim().toLowerCase();

        super(processedName, age); // Can now use processed data

        this.name = processedName;
        this.age = age;
    }
}
```

### Performance Improvements

#### JVM Enhancements
- **Enhanced Virtual Thread Performance**: Further optimizations to virtual thread creation and context switching
- **Improved G1 Garbage Collector**: Better handling of large heaps and reduced pause times
- **JIT Compiler Optimizations**: Enhanced profile-guided optimizations and better inlining decisions

---

## Java 23 (September 2024) - Latest Innovations

### Major Theme: Stabilizing Modern Features and Performance Excellence

Java 23 focuses on stabilizing preview features from previous releases while introducing new capabilities for modern application development.

### Key Features

#### 1. **Flexible Constructor Bodies (Second Preview - JEP 482)**
Enhanced version of statements before super() with additional flexibility.

```java
public class AdvancedValidation extends BaseValidator {
    private final String processedInput;
    private final ValidationLevel level;

    public AdvancedValidation(String input, ValidationLevel requestedLevel) {
        // Complex pre-processing before super() call
        String trimmed = input != null ? input.trim() : "";

        // Conditional logic based on input
        ValidationLevel actualLevel = switch (requestedLevel) {
            case STRICT -> {
                if (trimmed.isEmpty()) {
                    throw new IllegalArgumentException("Strict mode requires non-empty input");
                }
                yield ValidationLevel.STRICT;
            }
            case LENIENT -> ValidationLevel.LENIENT;
            case AUTO -> trimmed.length() > 100 ? ValidationLevel.STRICT : ValidationLevel.LENIENT;
        };

        // Call parent with processed values
        super(actualLevel);

        this.processedInput = trimmed;
        this.level = actualLevel;
    }
}
```

#### 2. **Primitive Types in Patterns (Preview - JEP 455)**
Pattern matching extended to work with primitive types.

```java
// Primitive pattern matching in switch expressions
public String formatNumber(Object value) {
    return switch (value) {
        case int i when i < 0 -> "Negative integer: " + i;
        case int i when i == 0 -> "Zero";
        case int i when i > 1000 -> "Large integer: " + i;
        case int i -> "Integer: " + i;
        case double d when d < 0.0 -> "Negative double: " + d;
        case double d when Double.isNaN(d) -> "Not a number";
        case double d -> String.format("Double: %.2f", d);
        case boolean b -> "Boolean: " + b;
        case char c when Character.isDigit(c) -> "Digit character: " + c;
        case char c -> "Character: " + c;
        default -> "Unknown type: " + value.getClass().getSimpleName();
    };
}
```

#### 3. **Module Import Declarations (Preview - JEP 476)**
Simplify importing entire modules instead of individual classes.

```java
// Module import (preview feature)
import module java.base;  // Imports commonly used classes from java.base

public class ModuleImportExample {
    // Can now use classes without individual imports
    public void processData() {
        List<String> data = new ArrayList<>();
        Set<String> unique = new HashSet<>();
        Map<String, Integer> counts = new HashMap<>();

        // Stream operations available without imports
        Stream.of("a", "b", "c")
              .collect(Collectors.toList());
    }
}
```

#### 4. **ZGC Generational (Standard - JEP 474)**
The generational ZGC becomes a standard feature, providing excellent low-latency garbage collection.

```java
// JVM flags for ZGC Generational (now standard)
// -XX:+UseZGC -XX:+UseGenerationalZGC

public class LowLatencyApplication {
    private final List<UserSession> activeSessions = new CopyOnWriteArrayList<>();

    public void handleHighThroughputOperations() {
        // ZGC Generational excels with:
        // 1. Young object allocation (sessions, temporary objects)
        // 2. Long-lived objects (cache, configuration)
        // 3. Predictable low pause times regardless of heap size

        for (int i = 0; i < 100_000; i++) {
            UserSession session = new UserSession(
                UUID.randomUUID().toString(),
                Instant.now(),
                generateUserData()
            );

            processSession(session);

            if (session.isPremiumUser()) {
                activeSessions.add(session); // Promoted to old generation
            }
        }
    }
}
```

### Performance Improvements

#### JVM Excellence
- **ZGC Generational**: Now standard, providing sub-millisecond pause times for applications of any size
- **Virtual Thread Performance**: Continued optimizations for creation and context switching overhead
- **JIT Compiler**: Enhanced C2 compiler with better optimization for modern patterns

---

## Java Version Evolution Summary

### LTS (Long Term Support) Releases
- **Java 8 (2014)**: Lambda expressions, Streams, new Date/Time API
- **Java 11 (2018)**: HTTP Client, local variable type inference, new String methods
- **Java 17 (2021)**: Sealed classes, pattern matching, text blocks
- **Java 21 (2023)**: Virtual threads, record patterns, string templates

### Major Paradigm Shifts

#### Java 8: Functional Programming
- Lambda expressions enabled functional programming style
- Streams revolutionized collection processing
- Optional reduced null pointer exceptions
- CompletableFuture improved asynchronous programming

#### Java 9+: Module System and Regular Releases
- Project Jigsaw introduced module system
- 6-month release cycle started
- Enhanced API evolution process
- Incubator and preview feature system

#### Java 17: Modern Language Features
- Pattern matching began transformation of conditional logic
- Records reduced boilerplate for data classes
- Sealed classes enabled exhaustive modeling
- Text blocks improved string handling

#### Java 21: Concurrency Revolution
- Virtual threads changed scalability paradigm
- Structured concurrency improved task coordination
- Pattern matching matured significantly
- Performance reached new heights

### Future Outlook

#### Project Valhalla (Value Types)
Expected features:
- Inline classes for memory efficiency
- Specialized generics avoiding boxing
- Better performance for data-intensive applications

#### Project Loom (Concurrency)
Continued evolution:
- Enhanced structured concurrency
- Better virtual thread integration
- Improved debugging and observability

#### Project Panama (Native Interop)
Ongoing development:
- Foreign Function & Memory API finalization
- Better native library integration
- Reduced JNI dependency

#### Project Leyden (Static Images)
Future capabilities:
- Ahead-of-time compilation
- Reduced memory footprint
- Faster startup times

### Recommended Adoption Strategy

#### For New Projects
1. **Start with Latest LTS**: Java 21 offers stable, production-ready features
2. **Consider Java 23**: For latest innovations if team is comfortable with non-LTS
3. **Plan for Regular Updates**: Embrace 6-month release cycle for continuous improvement

#### For Existing Projects
1. **Java 8 to 11**: Focus on module system and new APIs
2. **Java 11 to 17**: Embrace records, pattern matching, and text blocks
3. **Java 17 to 21**: Migrate to virtual threads for scalability
4. **Java 21 to 23**: Adopt latest language enhancements and performance improvements

---

*This comprehensive guide covers Java evolution from version 1.5 through 23, highlighting the major features, code examples, and migration considerations that have shaped modern Java development. Each version represents significant progress in making Java more expressive, performant, and developer-friendly while maintaining its core principles of reliability and backward compatibility.*