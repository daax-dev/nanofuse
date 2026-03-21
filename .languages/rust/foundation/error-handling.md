# Error Handling in Rust

## Core Philosophy

Rust uses **explicit error handling** through the type system, making errors part of the function's contract. No exceptions - all errors are values.

### Key Principles
- **Fail fast**: Detect errors early
- **Explicit handling**: Errors are part of function signatures
- **Type safety**: Impossible to ignore errors accidentally
- **Composability**: Error handling patterns compose well

## The Two Types of Failures

### 1. Recoverable Errors - `Result<T, E>`
```rust
// Function that can fail
fn divide(a: f64, b: f64) -> Result<f64, String> {
    if b == 0.0 {
        Err("Division by zero".to_string())
    } else {
        Ok(a / b)
    }
}

// Usage
match divide(10.0, 2.0) {
    Ok(result) => println!("Result: {}", result),
    Err(error) => println!("Error: {}", error),
}
```

### 2. Unrecoverable Errors - `panic!`
```rust
// Immediate program termination
panic!("Something went terribly wrong!");

// Panic with formatted message
panic!("Index {} out of bounds for length {}", index, length);

// Conditional panic
assert!(x > 0, "x must be positive, got {}", x);
assert_eq!(actual, expected);
assert_ne!(left, right);
```

## Result<T, E> Deep Dive

### Definition
```rust
enum Result<T, E> {
    Ok(T),   // Success case with value
    Err(E),  // Error case with error value
}
```

### Common Result Methods
```rust
let result: Result<i32, &str> = Ok(42);

// Check if Ok or Err
result.is_ok();     // true
result.is_err();    // false

// Extract values (panics if wrong variant)
result.unwrap();           // 42 (panics if Err)
result.expect("Custom message"); // 42 (panics with message if Err)
result.unwrap_err();       // Panics because it's Ok

// Safe extraction with defaults
result.unwrap_or(0);       // 42 (or 0 if Err)
result.unwrap_or_else(|e| 0); // 42 (or call closure if Err)

// Transform values
result.map(|x| x * 2);           // Ok(84)
result.map_err(|e| format!("Error: {}", e)); // Transform error

// Chain operations
result
    .map(|x| x * 2)
    .and_then(|x| if x > 100 { Err("Too big") } else { Ok(x) });
```

## The `?` Operator (Error Propagation)

### Basic Usage
```rust
use std::fs::File;
use std::io::{self, Read};

// Without ? operator (verbose)
fn read_file_verbose(path: &str) -> Result<String, io::Error> {
    let mut file = match File::open(path) {
        Ok(file) => file,
        Err(error) => return Err(error),
    };
    
    let mut contents = String::new();
    match file.read_to_string(&mut contents) {
        Ok(_) => Ok(contents),
        Err(error) => Err(error),
    }
}

// With ? operator (concise)
fn read_file_concise(path: &str) -> Result<String, io::Error> {
    let mut file = File::open(path)?;
    let mut contents = String::new();
    file.read_to_string(&mut contents)?;
    Ok(contents)
}

// Even more concise
fn read_file_shortest(path: &str) -> Result<String, io::Error> {
    std::fs::read_to_string(path)
}
```

### How `?` Works
1. If `Result` is `Ok(value)`, extracts `value`
2. If `Result` is `Err(error)`, returns `Err(error)` early
3. Automatically converts error types using `From` trait

## Option<T> for Nullable Values

### Definition
```rust
enum Option<T> {
    Some(T),  // Value is present
    None,     // Value is absent
}
```

### Common Option Methods
```rust
let some_value: Option<i32> = Some(42);
let no_value: Option<i32> = None;

// Check presence
some_value.is_some();  // true
some_value.is_none();  // false

// Extract values
some_value.unwrap();         // 42 (panics if None)
some_value.expect("No value"); // 42 (panics with message if None)
some_value.unwrap_or(0);     // 42 (or default if None)
some_value.unwrap_or_else(|| expensive_computation());

// Transform values
some_value.map(|x| x * 2);   // Some(84)
no_value.map(|x| x * 2);     // None

// Chain operations
some_value
    .map(|x| x * 2)
    .filter(|&x| x > 50)
    .or_else(|| Some(0));
```

### Converting Between Option and Result
```rust
let opt: Option<i32> = Some(42);
let res: Result<i32, &str> = opt.ok_or("No value");

let res: Result<i32, &str> = Ok(42);
let opt: Option<i32> = res.ok(); // Discards error
```

## Custom Error Types

### Simple String Errors
```rust
type Result<T> = std::result::Result<T, String>;

fn parse_number(s: &str) -> Result<i32> {
    s.parse().map_err(|e| format!("Failed to parse '{}': {}", s, e))
}
```

### Enum-Based Errors
```rust
#[derive(Debug)]
enum MathError {
    DivisionByZero,
    NegativeSquareRoot,
    Overflow,
}

fn safe_divide(a: f64, b: f64) -> Result<f64, MathError> {
    if b == 0.0 {
        Err(MathError::DivisionByZero)
    } else {
        Ok(a / b)
    }
}

fn square_root(x: f64) -> Result<f64, MathError> {
    if x < 0.0 {
        Err(MathError::NegativeSquareRoot)
    } else {
        Ok(x.sqrt())
    }
}
```

### Struct-Based Errors with Context
```rust
#[derive(Debug)]
struct ParseError {
    message: String,
    line: usize,
    column: usize,
}

impl ParseError {
    fn new(message: String, line: usize, column: usize) -> Self {
        ParseError { message, line, column }
    }
}

impl std::fmt::Display for ParseError {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        write!(f, "Parse error at line {}, column {}: {}", 
               self.line, self.column, self.message)
    }
}

impl std::error::Error for ParseError {}
```

## Error Trait and Error Chains

### The Error Trait
```rust
use std::error::Error;
use std::fmt;

#[derive(Debug)]
struct CustomError {
    details: String,
}

impl CustomError {
    fn new(msg: &str) -> CustomError {
        CustomError {
            details: msg.to_string(),
        }
    }
}

impl fmt::Display for CustomError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.details)
    }
}

impl Error for CustomError {
    fn description(&self) -> &str {
        &self.details
    }
}
```

### Error Chaining
```rust
use std::error::Error;
use std::fmt;

#[derive(Debug)]
struct HighLevelError {
    source: Box<dyn Error>,
}

impl fmt::Display for HighLevelError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "High level error")
    }
}

impl Error for HighLevelError {
    fn source(&self) -> Option<&(dyn Error + 'static)> {
        Some(self.source.as_ref())
    }
}
```

## Third-Party Error Libraries

### anyhow - Easy Error Handling
```rust
use anyhow::{Context, Result};

fn read_config() -> Result<Config> {
    let contents = std::fs::read_to_string("config.toml")
        .context("Failed to read config file")?;
    
    let config = toml::from_str(&contents)
        .context("Failed to parse config")?;
    
    Ok(config)
}
```

### thiserror - Derive Error Implementations
```rust
use thiserror::Error;

#[derive(Error, Debug)]
pub enum DataStoreError {
    #[error("data store disconnected")]
    Disconnect(#[from] io::Error),
    
    #[error("the data for key `{0}` is not available")]
    Redaction(String),
    
    #[error("invalid header (expected {expected:?}, found {found:?})")]
    InvalidHeader {
        expected: String,
        found: String,
    },
}
```

## Error Handling Patterns

### Early Return Pattern
```rust
fn process_data(data: &str) -> Result<ProcessedData, ProcessError> {
    let parsed = parse_data(data)?;
    let validated = validate_data(parsed)?;
    let processed = transform_data(validated)?;
    Ok(processed)
}
```

### Match vs ? Operator
```rust
// Use match when you need different handling per error
match risky_operation() {
    Ok(value) => process(value),
    Err(SpecificError::NetworkTimeout) => retry_operation(),
    Err(SpecificError::InvalidInput) => return Err("Bad input"),
    Err(other) => return Err(other.into()),
}

// Use ? when you just want to propagate
fn simple_case() -> Result<Output, Error> {
    let result = risky_operation()?;
    Ok(process(result))
}
```

### Multiple Error Types
```rust
// Using Box<dyn Error>
fn mixed_errors() -> Result<String, Box<dyn Error>> {
    let file = std::fs::File::open("data.txt")?;  // io::Error
    let number: i32 = "42".parse()?;              // ParseIntError
    Ok(format!("Number: {}", number))
}

// Using custom enum
#[derive(Debug)]
enum AppError {
    Io(std::io::Error),
    Parse(std::num::ParseIntError),
}

impl From<std::io::Error> for AppError {
    fn from(error: std::io::Error) -> Self {
        AppError::Io(error)
    }
}

impl From<std::num::ParseIntError> for AppError {
    fn from(error: std::num::ParseIntError) -> Self {
        AppError::Parse(error)
    }
}
```

## Error Recovery Strategies

### Retry Pattern
```rust
fn retry_operation<T, E, F>(mut operation: F, max_attempts: usize) -> Result<T, E>
where
    F: FnMut() -> Result<T, E>,
{
    for attempt in 1..=max_attempts {
        match operation() {
            Ok(result) => return Ok(result),
            Err(e) if attempt == max_attempts => return Err(e),
            Err(_) => {
                std::thread::sleep(std::time::Duration::from_millis(100 * attempt as u64));
            }
        }
    }
    unreachable!()
}
```

### Fallback Pattern
```rust
fn get_data_with_fallback() -> Result<Data, Error> {
    primary_data_source()
        .or_else(|_| secondary_data_source())
        .or_else(|_| default_data_source())
}
```

### Collect Results
```rust
fn process_all_items(items: Vec<Item>) -> Result<Vec<ProcessedItem>, ProcessError> {
    items
        .into_iter()
        .map(process_item)
        .collect()  // Stops at first error
}

// Continue processing despite errors
fn process_all_items_partial(items: Vec<Item>) -> (Vec<ProcessedItem>, Vec<ProcessError>) {
    items
        .into_iter()
        .map(process_item)
        .partition_map(|result| match result {
            Ok(item) => Either::Left(item),
            Err(error) => Either::Right(error),
        })
}
```

## Testing Error Conditions

### Unit Testing Errors
```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_division_by_zero() {
        let result = divide(10.0, 0.0);
        assert!(result.is_err());
        assert_eq!(result.unwrap_err(), "Division by zero");
    }

    #[test]
    fn test_successful_division() {
        let result = divide(10.0, 2.0);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), 5.0);
    }

    #[test]
    #[should_panic(expected = "assertion failed")]
    fn test_panic_condition() {
        assert!(false, "This should panic");
    }
}
```

## Performance Considerations

### Zero-Cost Error Handling
- `Result<T, E>` has same size as the larger of T or E
- `?` operator compiles to efficient branching
- No hidden allocations or exceptions

### Memory Layout
```rust
// Result is an enum, size = max(size_of::<T>(), size_of::<E>()) + discriminant
std::mem::size_of::<Result<i32, String>>(); // Size of String + discriminant
std::mem::size_of::<Result<i32, ()>>();     // Size of i32 + discriminant
```

## Best Practices

1. **Use Result for recoverable errors**, panic for unrecoverable ones
2. **Make error messages helpful** and actionable
3. **Use ? operator** for error propagation when appropriate
4. **Define custom error types** for complex applications
5. **Don't ignore errors** - handle them explicitly
6. **Use type aliases** for common Result types
7. **Consider error chaining** for debugging context
8. **Test error paths** as thoroughly as success paths
9. **Use library error types** (anyhow, thiserror) for convenience
10. **Document error conditions** in function documentation