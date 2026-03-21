# Java Programming Idioms and Best Practices

## Object Creation and Management

### Static Factory Methods Over Constructors
**Idiom**: Use static factory methods instead of public constructors when possible

Static factory methods provide more flexibility and expressiveness than constructors.

```java
// Prefer this
public static Optional<String> of(String value) {
    return value != null ? new Optional<>(value) : Optional.empty();
}

// Over this
public Optional(String value) {
    this.value = value;
}
```

**Advantages:**
- Named methods improve code readability
- Can return subclasses or cached instances
- Don't have to create new objects every time
- Can have multiple factory methods with different purposes

**Common Patterns:**
- `valueOf()` for type conversion
- `of()` for instance creation
- `from()` for conversion from other types
- `getInstance()` for singleton-like behavior

### Builder Pattern for Complex Objects
**Idiom**: Use the Builder pattern for objects with multiple optional parameters

The Builder pattern eliminates telescoping constructor anti-pattern and improves readability.

```java
public class Person {
    private final String name;
    private final int age;
    private final String email;
    private final String phone;
    
    private Person(Builder builder) {
        this.name = builder.name;
        this.age = builder.age;
        this.email = builder.email;
        this.phone = builder.phone;
    }
    
    public static class Builder {
        private String name; // required
        private int age;     // required
        private String email;
        private String phone;
        
        public Builder(String name, int age) {
            this.name = name;
            this.age = age;
        }
        
        public Builder email(String email) {
            this.email = email;
            return this;
        }
        
        public Builder phone(String phone) {
            this.phone = phone;
            return this;
        }
        
        public Person build() {
            return new Person(this);
        }
    }
}

// Usage
Person person = new Person.Builder("John Doe", 30)
    .email("john@example.com")
    .phone("555-1234")
    .build();
```

### Prefer Immutable Objects
**Idiom**: Create immutable objects whenever possible

Immutable objects are thread-safe, easier to reason about, and prevent many common bugs.

```java
public final class ImmutablePoint {
    private final int x;
    private final int y;
    
    public ImmutablePoint(int x, int y) {
        this.x = x;
        this.y = y;
    }
    
    public int getX() { return x; }
    public int getY() { return y; }
    
    public ImmutablePoint move(int dx, int dy) {
        return new ImmutablePoint(x + dx, y + dy);
    }
    
    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (!(obj instanceof ImmutablePoint)) return false;
        ImmutablePoint point = (ImmutablePoint) obj;
        return x == point.x && y == point.y;
    }
    
    @Override
    public int hashCode() {
        return Objects.hash(x, y);
    }
}
```

## Method Design

### Accept Interfaces, Return Concrete Classes
**Idiom**: Method parameters should be as general as possible, return types as specific as useful

This maximizes flexibility for callers while providing useful guarantees about return values.

```java
// Good: accepts any List implementation
public ArrayList<String> processItems(List<String> items) {
    ArrayList<String> result = new ArrayList<>();
    // process items
    return result; // specific return type provides guarantees
}

// Less flexible: restricts input to specific implementation
public List<String> processItems(ArrayList<String> items) {
    // implementation
}
```

### Defensive Copying for Mutable Parameters
**Idiom**: Make defensive copies of mutable parameters and return values

Prevents external modification of internal state through reference sharing.

```java
public class DateRange {
    private final Date start;
    private final Date end;
    
    public DateRange(Date start, Date end) {
        this.start = new Date(start.getTime()); // defensive copy
        this.end = new Date(end.getTime());     // defensive copy
        
        if (this.start.after(this.end)) {
            throw new IllegalArgumentException("Start after end");
        }
    }
    
    public Date getStart() {
        return new Date(start.getTime()); // defensive copy
    }
    
    public Date getEnd() {
        return new Date(end.getTime()); // defensive copy
    }
}
```

### Null Object Pattern Over Null Returns
**Idiom**: Return empty collections or Optional instead of null

Eliminates null pointer exceptions and simplifies client code.

```java
// Good: never returns null
public List<String> getItems() {
    return items.isEmpty() ? Collections.emptyList() : new ArrayList<>(items);
}

public Optional<User> findUser(String id) {
    return users.containsKey(id) ? Optional.of(users.get(id)) : Optional.empty();
}

// Avoid: forces null checks on callers
public List<String> getItems() {
    return items.isEmpty() ? null : items;
}
```

## Exception Handling

### Fail-Fast Principle
**Idiom**: Validate parameters immediately and throw exceptions early

Early validation makes bugs easier to diagnose and prevents corruption of object state.

```java
public void processOrder(Order order, Customer customer) {
    Objects.requireNonNull(order, "Order cannot be null");
    Objects.requireNonNull(customer, "Customer cannot be null");
    
    if (order.getItems().isEmpty()) {
        throw new IllegalArgumentException("Order must have at least one item");
    }
    
    if (!customer.isActive()) {
        throw new IllegalStateException("Customer account is not active");
    }
    
    // proceed with processing
}
```

### Use Checked Exceptions for Recoverable Conditions
**Idiom**: Use checked exceptions for conditions the caller can reasonably be expected to recover from

Reserve unchecked exceptions for programming errors and unrecoverable conditions.

```java
// Checked exception for recoverable condition
public void saveToFile(String data, String filename) throws IOException {
    // File operations that might fail due to external conditions
}

// Unchecked exception for programming error
public void setAge(int age) {
    if (age < 0) {
        throw new IllegalArgumentException("Age cannot be negative");
    }
}
```

### Exception Chaining for Context
**Idiom**: Chain exceptions to preserve stack traces while adding context

Provides complete error information without losing original cause.

```java
public void processData() throws DataProcessingException {
    try {
        // some operation that might fail
    } catch (IOException e) {
        throw new DataProcessingException("Failed to process data file", e);
    }
}
```

## Resource Management

### Try-With-Resources for Automatic Cleanup
**Idiom**: Use try-with-resources for all AutoCloseable resources

Ensures resources are closed even if exceptions occur.

```java
// Good: automatic resource management
public String readFile(String filename) throws IOException {
    try (BufferedReader reader = Files.newBufferedReader(Paths.get(filename))) {
        return reader.lines()
                    .collect(Collectors.joining("\n"));
    }
}

// Multiple resources
public void copyFile(String source, String destination) throws IOException {
    try (InputStream in = Files.newInputStream(Paths.get(source));
         OutputStream out = Files.newOutputStream(Paths.get(destination))) {
        in.transferTo(out);
    }
}
```

## Collections and Streams

### Prefer Streams for Data Processing
**Idiom**: Use streams for complex data transformations and filtering

Streams provide readable, composable operations on collections.

```java
// Good: readable data processing pipeline
public List<String> getActiveUserNames(List<User> users) {
    return users.stream()
                .filter(User::isActive)
                .map(User::getName)
                .sorted()
                .collect(Collectors.toList());
}

// Complex aggregation example
public Map<Department, Double> getAverageSalaryByDepartment(List<Employee> employees) {
    return employees.stream()
                   .collect(Collectors.groupingBy(
                       Employee::getDepartment,
                       Collectors.averagingDouble(Employee::getSalary)));
}
```

### Use Method References When Possible
**Idiom**: Prefer method references over lambda expressions when they improve readability

Method references are often more concise and expressive than lambda expressions.

```java
// Method references (preferred)
users.forEach(System.out::println);
users.stream().map(User::getName);
users.stream().filter(Objects::nonNull);

// Equivalent lambda expressions (more verbose)
users.forEach(user -> System.out.println(user));
users.stream().map(user -> user.getName());
users.stream().filter(user -> user != null);
```

### Collect to Specific Collection Types When Needed
**Idiom**: Specify collection types when order or performance characteristics matter

Choose appropriate collection types for specific use cases.

```java
// Preserve order with LinkedHashSet
Set<String> orderedUniqueNames = users.stream()
    .map(User::getName)
    .collect(Collectors.toCollection(LinkedHashSet::new));

// Use TreeSet for sorted unique elements
Set<String> sortedNames = users.stream()
    .map(User::getName)
    .collect(Collectors.toCollection(TreeSet::new));

// Use ArrayList for indexed access
List<String> namesList = users.stream()
    .map(User::getName)
    .collect(Collectors.toCollection(ArrayList::new));
```

## Equality and Hashing

### Always Override hashCode() When Overriding equals()
**Idiom**: Maintain the equals-hashCode contract

Objects that are equal must have the same hash code.

```java
@Override
public boolean equals(Object obj) {
    if (this == obj) return true;
    if (!(obj instanceof Person)) return false;
    Person person = (Person) obj;
    return age == person.age &&
           Objects.equals(name, person.name) &&
           Objects.equals(email, person.email);
}

@Override
public int hashCode() {
    return Objects.hash(name, age, email);
}
```

### Use Objects.equals() for Null-Safe Comparisons
**Idiom**: Handle null values safely in equality comparisons

```java
// Good: null-safe comparison
public boolean equals(Object obj) {
    if (this == obj) return true;
    if (!(obj instanceof Book)) return false;
    Book book = (Book) obj;
    return Objects.equals(title, book.title) &&
           Objects.equals(author, book.author);
}

// Avoid: manual null checks
public boolean equals(Object obj) {
    if (this == obj) return true;
    if (!(obj instanceof Book)) return false;
    Book book = (Book) obj;
    return (title == null ? book.title == null : title.equals(book.title)) &&
           (author == null ? book.author == null : author.equals(book.author));
}
```

## Enum Usage Patterns

### Use Enums for Constants
**Idiom**: Prefer enums over public static final constants

Enums provide type safety, namespace control, and additional functionality.

```java
// Good: type-safe enum
public enum Status {
    PENDING,
    APPROVED,
    REJECTED;
    
    public boolean isComplete() {
        return this == APPROVED || this == REJECTED;
    }
}

// Avoid: string constants
public class StatusConstants {
    public static final String PENDING = "PENDING";
    public static final String APPROVED = "APPROVED";
    public static final String REJECTED = "REJECTED";
}
```

### Enum with Behavior
**Idiom**: Add methods to enums for enum-specific behavior

Enums can contain methods and implement interfaces.

```java
public enum Operation {
    PLUS("+") {
        @Override
        public double apply(double x, double y) {
            return x + y;
        }
    },
    MINUS("-") {
        @Override
        public double apply(double x, double y) {
            return x - y;
        }
    };
    
    private final String symbol;
    
    Operation(String symbol) {
        this.symbol = symbol;
    }
    
    public abstract double apply(double x, double y);
    
    @Override
    public String toString() {
        return symbol;
    }
}
```

## Modern Java Idioms (Java 8+)

### Optional for Null-Safe Method Chaining
**Idiom**: Use Optional to eliminate null pointer exceptions

```java
// Good: null-safe chain
public String getCustomerCityName(Long customerId) {
    return customerService.findById(customerId)
        .map(Customer::getAddress)
        .map(Address::getCity)
        .map(City::getName)
        .orElse("Unknown");
}

// Traditional null checking (more verbose)
public String getCustomerCityName(Long customerId) {
    Customer customer = customerService.findById(customerId);
    if (customer != null) {
        Address address = customer.getAddress();
        if (address != null) {
            City city = address.getCity();
            if (city != null) {
                return city.getName();
            }
        }
    }
    return "Unknown";
}
```

### Functional Interface Usage
**Idiom**: Leverage functional interfaces for flexible API design

```java
// Flexible filtering with Predicate
public <T> List<T> filter(List<T> items, Predicate<T> predicate) {
    return items.stream()
                .filter(predicate)
                .collect(Collectors.toList());
}

// Flexible transformation with Function
public <T, R> List<R> transform(List<T> items, Function<T, R> mapper) {
    return items.stream()
                .map(mapper)
                .collect(Collectors.toList());
}

// Usage examples
List<String> activeUsernames = filter(users, User::isActive)
    .stream()
    .map(User::getUsername)
    .collect(Collectors.toList());
```

### Records for Data Classes (Java 14+)
**Idiom**: Use records for immutable data containers

Records eliminate boilerplate for simple data classes.

```java
// Modern: using records
public record Point(int x, int y) {
    // Validation in compact constructor
    public Point {
        if (x < 0 || y < 0) {
            throw new IllegalArgumentException("Coordinates must be non-negative");
        }
    }
    
    // Additional methods if needed
    public double distanceFromOrigin() {
        return Math.sqrt(x * x + y * y);
    }
}

// Traditional class equivalent (much more verbose)
public final class Point {
    private final int x;
    private final int y;
    
    public Point(int x, int y) {
        if (x < 0 || y < 0) {
            throw new IllegalArgumentException("Coordinates must be non-negative");
        }
        this.x = x;
        this.y = y;
    }
    
    public int x() { return x; }
    public int y() { return y; }
    
    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (!(obj instanceof Point)) return false;
        Point point = (Point) obj;
        return x == point.x && y == point.y;
    }
    
    @Override
    public int hashCode() {
        return Objects.hash(x, y);
    }
    
    @Override
    public String toString() {
        return "Point[x=" + x + ", y=" + y + "]";
    }
    
    public double distanceFromOrigin() {
        return Math.sqrt(x * x + y * y);
    }
}
```

## Performance-Related Idioms

### StringBuilder for String Concatenation
**Idiom**: Use StringBuilder for multiple string concatenations

StringBuilder is more efficient than repeated string concatenation operations.

```java
// Good: for multiple concatenations
public String buildMessage(List<String> items) {
    StringBuilder sb = new StringBuilder();
    sb.append("Items: ");
    for (String item : items) {
        sb.append(item).append(", ");
    }
    if (sb.length() > 7) { // "Items: ".length()
        sb.setLength(sb.length() - 2); // remove trailing ", "
    }
    return sb.toString();
}

// OK: for simple concatenations (Java optimizes this)
public String formatName(String first, String last) {
    return first + " " + last;
}
```

### Use Enhanced For-Loop for Iteration
**Idiom**: Prefer enhanced for-loop for simple iteration

More readable and less error-prone than traditional for-loop.

```java
// Good: enhanced for-loop
for (String item : items) {
    process(item);
}

// Use traditional for-loop only when you need the index
for (int i = 0; i < items.size(); i++) {
    String item = items.get(i);
    System.out.println(i + ": " + item);
}
```

---

*Last Updated: January 2025*
*Based on "Effective Java", community best practices, and modern Java development patterns*