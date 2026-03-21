# Java Error Handling - Comprehensive Guide

## Exception Hierarchy and Types

### Exception Class Hierarchy
```
java.lang.Throwable
├── java.lang.Error (Unchecked)
│   ├── OutOfMemoryError
│   ├── StackOverflowError
│   ├── VirtualMachineError
│   └── AssertionError
└── java.lang.Exception
    ├── java.lang.RuntimeException (Unchecked)
    │   ├── NullPointerException
    │   ├── IllegalArgumentException
    │   ├── IndexOutOfBoundsException
    │   ├── ClassCastException
    │   └── ConcurrentModificationException
    └── Checked Exceptions
        ├── IOException
        ├── SQLException
        ├── ClassNotFoundException
        └── InterruptedException
```

### Checked vs Unchecked Exceptions

#### Checked Exceptions
Checked exceptions must be either handled with try-catch or declared in the method signature with throws.

```java
// Must be handled or declared
public void readFile(String filename) throws IOException {
    FileReader file = new FileReader(filename); // May throw IOException
    // File operations
}

// Or handled with try-catch
public void readFileWithHandling(String filename) {
    try {
        FileReader file = new FileReader(filename);
        // File operations
    } catch (IOException e) {
        System.err.println("Error reading file: " + e.getMessage());
    }
}

// Common checked exceptions
public void checkedExceptionExamples() {
    try {
        // IOException and subclasses
        Files.readAllLines(Paths.get("data.txt"));

        // SQLException
        Connection conn = DriverManager.getConnection("jdbc:mysql://localhost/db");

        // ClassNotFoundException
        Class.forName("com.example.MyClass");

        // InterruptedException
        Thread.sleep(1000);

    } catch (IOException e) {
        System.err.println("I/O error: " + e.getMessage());
    } catch (SQLException e) {
        System.err.println("Database error: " + e.getMessage());
    } catch (ClassNotFoundException e) {
        System.err.println("Class not found: " + e.getMessage());
    } catch (InterruptedException e) {
        System.err.println("Thread interrupted: " + e.getMessage());
        Thread.currentThread().interrupt(); // Restore interrupt status
    }
}
```

#### Unchecked Exceptions (Runtime Exceptions)
Runtime exceptions don't need to be declared or caught, but can be handled if needed.

```java
public class RuntimeExceptionExamples {

    public void nullPointerExceptionExample() {
        String str = null;
        // throws NullPointerException at runtime
        int length = str.length();
    }

    public void illegalArgumentExceptionExample(int age) {
        if (age < 0) {
            throw new IllegalArgumentException("Age cannot be negative: " + age);
        }
    }

    public void indexOutOfBoundsExample() {
        List<String> list = Arrays.asList("a", "b", "c");
        // throws IndexOutOfBoundsException
        String item = list.get(10);
    }

    public void classCastExceptionExample() {
        Object obj = "Hello";
        // throws ClassCastException
        Integer number = (Integer) obj;
    }

    // Handling runtime exceptions (optional)
    public void safeExecution() {
        try {
            riskyOperation();
        } catch (NullPointerException e) {
            System.err.println("Null pointer error: " + e.getMessage());
        } catch (IllegalArgumentException e) {
            System.err.println("Invalid argument: " + e.getMessage());
        } catch (RuntimeException e) {
            System.err.println("Runtime error: " + e.getMessage());
        }
    }
}
```

## Exception Handling Mechanisms

### Try-Catch-Finally Blocks

#### Basic Try-Catch
```java
public void basicTryCatch() {
    try {
        // Code that may throw exception
        int result = divide(10, 0);
        System.out.println("Result: " + result);
    } catch (ArithmeticException e) {
        // Handle specific exception
        System.err.println("Division by zero error: " + e.getMessage());
    }
}

private int divide(int a, int b) {
    return a / b; // May throw ArithmeticException
}
```

#### Multiple Catch Blocks
```java
public void multipleCatchBlocks() {
    try {
        performOperations();
    } catch (FileNotFoundException e) {
        // Handle file not found
        System.err.println("File not found: " + e.getMessage());
        // Log and possibly create default file
    } catch (IOException e) {
        // Handle other I/O exceptions
        System.err.println("I/O error: " + e.getMessage());
        // Log and possibly retry
    } catch (NumberFormatException e) {
        // Handle number parsing errors
        System.err.println("Invalid number format: " + e.getMessage());
        // Use default values
    } catch (Exception e) {
        // Handle any other exception
        System.err.println("Unexpected error: " + e.getMessage());
        // Log for debugging
    }
}
```

#### Multi-Catch (Java 7+)
```java
public void multiCatch() {
    try {
        performOperations();
    } catch (IOException | SQLException e) {
        // Handle multiple exception types with same logic
        System.err.println("Database or I/O error: " + e.getMessage());
        logError(e);
    } catch (NumberFormatException | DateTimeParseException e) {
        // Handle parsing errors
        System.err.println("Parsing error: " + e.getMessage());
        useDefaultValues();
    }
}
```

#### Finally Block
```java
public void finallyBlockExample() {
    FileInputStream file = null;
    try {
        file = new FileInputStream("data.txt");
        // Process file
        processFile(file);
    } catch (IOException e) {
        System.err.println("Error processing file: " + e.getMessage());
    } finally {
        // Always executes (unless JVM terminates)
        if (file != null) {
            try {
                file.close();
            } catch (IOException e) {
                System.err.println("Error closing file: " + e.getMessage());
            }
        }
        System.out.println("Cleanup completed");
    }
}
```

### Try-With-Resources (Java 7+)

#### Basic Try-With-Resources
```java
public void tryWithResources() {
    // Automatically closes resources that implement AutoCloseable
    try (FileInputStream file = new FileInputStream("data.txt");
         BufferedReader reader = new BufferedReader(new InputStreamReader(file))) {

        String line;
        while ((line = reader.readLine()) != null) {
            System.out.println(line);
        }
    } catch (IOException e) {
        System.err.println("Error reading file: " + e.getMessage());
    }
    // Resources are automatically closed here
}
```

#### Multiple Resources
```java
public void multipleResources() {
    try (FileInputStream input = new FileInputStream("source.txt");
         FileOutputStream output = new FileOutputStream("destination.txt");
         BufferedReader reader = new BufferedReader(new InputStreamReader(input));
         BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(output))) {

        String line;
        while ((line = reader.readLine()) != null) {
            writer.write(line.toUpperCase());
            writer.newLine();
        }
    } catch (IOException e) {
        System.err.println("Error copying file: " + e.getMessage());
    }
}
```

#### Custom AutoCloseable Resources
```java
public class DatabaseConnection implements AutoCloseable {
    private Connection connection;

    public DatabaseConnection(String url) throws SQLException {
        this.connection = DriverManager.getConnection(url);
    }

    public void executeQuery(String sql) throws SQLException {
        try (PreparedStatement stmt = connection.prepareStatement(sql);
             ResultSet rs = stmt.executeQuery()) {

            while (rs.next()) {
                // Process results
            }
        }
    }

    @Override
    public void close() throws SQLException {
        if (connection != null && !connection.isClosed()) {
            connection.close();
            System.out.println("Database connection closed");
        }
    }
}

// Usage
public void useDatabaseConnection() {
    try (DatabaseConnection db = new DatabaseConnection("jdbc:h2:mem:testdb")) {
        db.executeQuery("SELECT * FROM users");
    } catch (SQLException e) {
        System.err.println("Database error: " + e.getMessage());
    }
    // Database connection automatically closed
}
```

## Throwing Exceptions

### Throw Statement
```java
public class ValidationService {

    public void validateAge(int age) {
        if (age < 0) {
            throw new IllegalArgumentException("Age cannot be negative: " + age);
        }
        if (age > 150) {
            throw new IllegalArgumentException("Age seems unrealistic: " + age);
        }
    }

    public void validateEmail(String email) {
        if (email == null || email.trim().isEmpty()) {
            throw new IllegalArgumentException("Email cannot be null or empty");
        }
        if (!email.contains("@")) {
            throw new IllegalArgumentException("Invalid email format: " + email);
        }
    }

    public void processPayment(double amount) {
        if (amount <= 0) {
            throw new IllegalArgumentException("Payment amount must be positive");
        }

        // Simulate payment processing
        boolean paymentSucceeded = processPaymentInternal(amount);
        if (!paymentSucceeded) {
            throw new RuntimeException("Payment processing failed for amount: " + amount);
        }
    }

    private boolean processPaymentInternal(double amount) {
        // Simulate payment processing logic
        return amount < 10000; // Simulated failure for large amounts
    }
}
```

### Throws Declaration
```java
public class FileProcessor {

    // Method declares checked exceptions it may throw
    public String readFileContent(String filename)
            throws IOException, SecurityException {

        // Check file permissions
        File file = new File(filename);
        if (!file.canRead()) {
            throw new SecurityException("No read permission for file: " + filename);
        }

        // Read file content - may throw IOException
        return Files.readString(Paths.get(filename));
    }

    // Method with multiple declared exceptions
    public void processConfigFile(String configFile)
            throws IOException, NumberFormatException, IllegalStateException {

        String content = readFileContent(configFile);

        // Parse configuration - may throw NumberFormatException
        Properties props = new Properties();
        props.load(new StringReader(content));

        String portStr = props.getProperty("port");
        if (portStr == null) {
            throw new IllegalStateException("Port configuration is missing");
        }

        int port = Integer.parseInt(portStr); // May throw NumberFormatException

        if (port < 1 || port > 65535) {
            throw new IllegalStateException("Invalid port number: " + port);
        }
    }

    // Catching and re-throwing with more context
    public void initializeSystem(String configFile) throws SystemInitializationException {
        try {
            processConfigFile(configFile);
        } catch (IOException e) {
            throw new SystemInitializationException(
                "Failed to read configuration file: " + configFile, e);
        } catch (NumberFormatException e) {
            throw new SystemInitializationException(
                "Invalid number format in configuration file: " + configFile, e);
        } catch (IllegalStateException e) {
            throw new SystemInitializationException(
                "Configuration error in file: " + configFile, e);
        }
    }
}

// Custom checked exception
class SystemInitializationException extends Exception {
    public SystemInitializationException(String message, Throwable cause) {
        super(message, cause);
    }
}
```

## Custom Exception Classes

### Creating Custom Exceptions
```java
// Custom checked exception
public class InsufficientFundsException extends Exception {
    private final double requestedAmount;
    private final double availableBalance;

    public InsufficientFundsException(double requestedAmount, double availableBalance) {
        super(String.format("Insufficient funds. Requested: %.2f, Available: %.2f",
              requestedAmount, availableBalance));
        this.requestedAmount = requestedAmount;
        this.availableBalance = availableBalance;
    }

    public InsufficientFundsException(String message, double requestedAmount, double availableBalance) {
        super(message);
        this.requestedAmount = requestedAmount;
        this.availableBalance = availableBalance;
    }

    public InsufficientFundsException(String message, Throwable cause,
                                    double requestedAmount, double availableBalance) {
        super(message, cause);
        this.requestedAmount = requestedAmount;
        this.availableBalance = availableBalance;
    }

    public double getRequestedAmount() {
        return requestedAmount;
    }

    public double getAvailableBalance() {
        return availableBalance;
    }

    public double getShortfall() {
        return requestedAmount - availableBalance;
    }
}

// Custom unchecked exception
public class InvalidAccountException extends RuntimeException {
    private final String accountId;

    public InvalidAccountException(String accountId) {
        super("Invalid account ID: " + accountId);
        this.accountId = accountId;
    }

    public InvalidAccountException(String accountId, String message) {
        super(message);
        this.accountId = accountId;
    }

    public InvalidAccountException(String accountId, String message, Throwable cause) {
        super(message, cause);
        this.accountId = accountId;
    }

    public String getAccountId() {
        return accountId;
    }
}

// Using custom exceptions
public class BankAccount {
    private String accountId;
    private double balance;

    public BankAccount(String accountId, double initialBalance) {
        if (accountId == null || accountId.trim().isEmpty()) {
            throw new InvalidAccountException(accountId, "Account ID cannot be null or empty");
        }
        this.accountId = accountId;
        this.balance = initialBalance;
    }

    public void withdraw(double amount) throws InsufficientFundsException {
        if (amount <= 0) {
            throw new IllegalArgumentException("Withdrawal amount must be positive");
        }

        if (amount > balance) {
            throw new InsufficientFundsException(amount, balance);
        }

        balance -= amount;
    }

    public void deposit(double amount) {
        if (amount <= 0) {
            throw new IllegalArgumentException("Deposit amount must be positive");
        }
        balance += amount;
    }

    public double getBalance() {
        return balance;
    }
}
```

## Exception Handling Best Practices

### Defensive Programming
```java
public class DefensiveProgramming {

    // Validate input parameters
    public String formatName(String firstName, String lastName) {
        Objects.requireNonNull(firstName, "First name cannot be null");
        Objects.requireNonNull(lastName, "Last name cannot be null");

        if (firstName.trim().isEmpty()) {
            throw new IllegalArgumentException("First name cannot be empty");
        }
        if (lastName.trim().isEmpty()) {
            throw new IllegalArgumentException("Last name cannot be empty");
        }

        return firstName.trim() + " " + lastName.trim();
    }

    // Null-safe operations
    public int getStringLength(String str) {
        return str != null ? str.length() : 0;
    }

    // Using Optional to avoid null returns
    public Optional<String> findUserById(String userId) {
        if (userId == null || userId.trim().isEmpty()) {
            return Optional.empty();
        }

        // Simulate database lookup
        String user = database.findUser(userId);
        return Optional.ofNullable(user);
    }

    // Fail-fast principle
    public void processOrder(Order order) {
        Objects.requireNonNull(order, "Order cannot be null");

        if (order.getItems().isEmpty()) {
            throw new IllegalStateException("Order must have at least one item");
        }

        if (order.getCustomerId() == null) {
            throw new IllegalStateException("Order must have a customer ID");
        }

        // Process order...
    }
}
```

### Resource Management
```java
public class ResourceManagement {

    // Proper resource cleanup
    public void copyFileOldWay(String source, String destination) throws IOException {
        FileInputStream input = null;
        FileOutputStream output = null;

        try {
            input = new FileInputStream(source);
            output = new FileOutputStream(destination);

            byte[] buffer = new byte[1024];
            int bytesRead;
            while ((bytesRead = input.read(buffer)) != -1) {
                output.write(buffer, 0, bytesRead);
            }
        } finally {
            // Cleanup in reverse order of creation
            if (output != null) {
                try {
                    output.close();
                } catch (IOException e) {
                    System.err.println("Error closing output stream: " + e.getMessage());
                }
            }
            if (input != null) {
                try {
                    input.close();
                } catch (IOException e) {
                    System.err.println("Error closing input stream: " + e.getMessage());
                }
            }
        }
    }

    // Better approach with try-with-resources
    public void copyFileNewWay(String source, String destination) throws IOException {
        try (FileInputStream input = new FileInputStream(source);
             FileOutputStream output = new FileOutputStream(destination)) {

            byte[] buffer = new byte[1024];
            int bytesRead;
            while ((bytesRead = input.read(buffer)) != -1) {
                output.write(buffer, 0, bytesRead);
            }
        }
        // Resources automatically closed
    }
}
```

### Exception Translation and Wrapping
```java
public class ServiceLayer {
    private DatabaseRepository repository;

    // Translate low-level exceptions to business exceptions
    public Customer findCustomer(String customerId) throws CustomerNotFoundException {
        try {
            return repository.findById(customerId);
        } catch (SQLException e) {
            if (e.getErrorCode() == 404) {
                throw new CustomerNotFoundException("Customer not found: " + customerId);
            } else {
                throw new DataAccessException("Database error while finding customer", e);
            }
        } catch (ConnectionException e) {
            throw new ServiceUnavailableException("Database connection failed", e);
        }
    }

    // Preserve exception chain
    public void updateCustomer(Customer customer) throws CustomerUpdateException {
        try {
            validateCustomer(customer);
            repository.save(customer);
        } catch (ValidationException e) {
            throw new CustomerUpdateException("Customer validation failed", e);
        } catch (SQLException e) {
            throw new CustomerUpdateException("Database error during update", e);
        }
    }

    private void validateCustomer(Customer customer) throws ValidationException {
        if (customer.getEmail() == null || !customer.getEmail().contains("@")) {
            throw new ValidationException("Invalid email address");
        }
    }
}

// Business exceptions
class CustomerNotFoundException extends Exception {
    public CustomerNotFoundException(String message) { super(message); }
}

class CustomerUpdateException extends Exception {
    public CustomerUpdateException(String message, Throwable cause) { super(message, cause); }
}

class DataAccessException extends RuntimeException {
    public DataAccessException(String message, Throwable cause) { super(message, cause); }
}

class ServiceUnavailableException extends RuntimeException {
    public ServiceUnavailableException(String message, Throwable cause) { super(message, cause); }
}

class ValidationException extends Exception {
    public ValidationException(String message) { super(message); }
}
```

### Logging and Monitoring
```java
public class ErrorHandlingWithLogging {
    private static final Logger logger = LoggerFactory.getLogger(ErrorHandlingWithLogging.class);

    public void processRequest(String requestId, String data) {
        try {
            logger.info("Processing request: {}", requestId);

            validateInput(data);
            processData(data);

            logger.info("Request processed successfully: {}", requestId);

        } catch (ValidationException e) {
            logger.warn("Validation failed for request {}: {}", requestId, e.getMessage());
            throw new BadRequestException("Invalid input data", e);
        } catch (ProcessingException e) {
            logger.error("Processing failed for request {}: {}", requestId, e.getMessage(), e);
            throw new InternalServerException("Processing failed", e);
        } catch (Exception e) {
            logger.error("Unexpected error processing request {}: {}", requestId, e.getMessage(), e);
            throw new InternalServerException("Unexpected error", e);
        }
    }

    // Structured error information
    public ErrorResponse handleException(Exception e, String requestId) {
        ErrorInfo errorInfo = new ErrorInfo();
        errorInfo.setTimestamp(Instant.now());
        errorInfo.setRequestId(requestId);

        if (e instanceof ValidationException) {
            errorInfo.setErrorCode("VALIDATION_ERROR");
            errorInfo.setMessage(e.getMessage());
            errorInfo.setSeverity(ErrorSeverity.WARNING);
        } else if (e instanceof ProcessingException) {
            errorInfo.setErrorCode("PROCESSING_ERROR");
            errorInfo.setMessage("Internal processing error");
            errorInfo.setSeverity(ErrorSeverity.ERROR);
        } else {
            errorInfo.setErrorCode("UNKNOWN_ERROR");
            errorInfo.setMessage("An unexpected error occurred");
            errorInfo.setSeverity(ErrorSeverity.CRITICAL);
        }

        logger.error("Error handled: {}", errorInfo, e);
        return new ErrorResponse(errorInfo);
    }
}
```

### Circuit Breaker Pattern
```java
public class CircuitBreaker {
    private enum State { CLOSED, OPEN, HALF_OPEN }

    private State state = State.CLOSED;
    private int failureCount = 0;
    private long lastFailureTime = 0;
    private final int failureThreshold;
    private final long timeout;

    public CircuitBreaker(int failureThreshold, long timeoutMs) {
        this.failureThreshold = failureThreshold;
        this.timeout = timeoutMs;
    }

    public <T> T execute(Supplier<T> operation) throws CircuitBreakerException {
        if (state == State.OPEN) {
            if (System.currentTimeMillis() - lastFailureTime < timeout) {
                throw new CircuitBreakerException("Circuit breaker is OPEN");
            } else {
                state = State.HALF_OPEN;
            }
        }

        try {
            T result = operation.get();
            onSuccess();
            return result;
        } catch (Exception e) {
            onFailure();
            throw new CircuitBreakerException("Operation failed", e);
        }
    }

    private void onSuccess() {
        failureCount = 0;
        state = State.CLOSED;
    }

    private void onFailure() {
        failureCount++;
        lastFailureTime = System.currentTimeMillis();

        if (failureCount >= failureThreshold) {
            state = State.OPEN;
        }
    }
}
```

## Advanced Error Handling Patterns

### Retry Pattern
```java
public class RetryHandler {

    public <T> T executeWithRetry(Supplier<T> operation, int maxAttempts, long delayMs)
            throws MaxAttemptsExceededException {

        int attempt = 0;
        Exception lastException = null;

        while (attempt < maxAttempts) {
            try {
                return operation.get();
            } catch (Exception e) {
                lastException = e;
                attempt++;

                if (attempt >= maxAttempts) {
                    break;
                }

                try {
                    Thread.sleep(delayMs);
                } catch (InterruptedException ie) {
                    Thread.currentThread().interrupt();
                    throw new RuntimeException("Retry interrupted", ie);
                }
            }
        }

        throw new MaxAttemptsExceededException(
            "Failed after " + maxAttempts + " attempts", lastException);
    }

    // Exponential backoff retry
    public <T> T executeWithExponentialBackoff(Supplier<T> operation, int maxAttempts)
            throws MaxAttemptsExceededException {

        int attempt = 0;
        Exception lastException = null;

        while (attempt < maxAttempts) {
            try {
                return operation.get();
            } catch (Exception e) {
                lastException = e;
                attempt++;

                if (attempt >= maxAttempts) {
                    break;
                }

                long delay = (long) Math.pow(2, attempt) * 1000; // Exponential backoff
                try {
                    Thread.sleep(delay);
                } catch (InterruptedException ie) {
                    Thread.currentThread().interrupt();
                    throw new RuntimeException("Retry interrupted", ie);
                }
            }
        }

        throw new MaxAttemptsExceededException(
            "Failed after " + maxAttempts + " attempts", lastException);
    }
}
```

### Error Aggregation
```java
public class ValidationService {

    public static class ValidationErrors {
        private final List<String> errors = new ArrayList<>();

        public void add(String error) {
            errors.add(error);
        }

        public boolean hasErrors() {
            return !errors.isEmpty();
        }

        public List<String> getErrors() {
            return Collections.unmodifiableList(errors);
        }

        public void throwIfHasErrors() throws ValidationException {
            if (hasErrors()) {
                throw new ValidationException("Validation failed: " +
                    String.join(", ", errors));
            }
        }
    }

    public void validateUser(User user) throws ValidationException {
        ValidationErrors errors = new ValidationErrors();

        if (user.getName() == null || user.getName().trim().isEmpty()) {
            errors.add("Name is required");
        }

        if (user.getEmail() == null || !user.getEmail().contains("@")) {
            errors.add("Valid email is required");
        }

        if (user.getAge() < 0 || user.getAge() > 150) {
            errors.add("Age must be between 0 and 150");
        }

        if (user.getPhone() == null || user.getPhone().length() < 10) {
            errors.add("Valid phone number is required");
        }

        errors.throwIfHasErrors();
    }
}
```

This comprehensive guide covers all aspects of Java error handling, from basic exception types to advanced patterns and best practices, providing the foundation for robust error management in Java applications.