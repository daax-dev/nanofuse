# Rust Types and Collections

## Primitive Types

### Scalar Types

#### Integer Types
```rust
// Signed integers
let i8_val: i8 = 127;           // -128 to 127
let i16_val: i16 = 32_767;      // -32,768 to 32,767
let i32_val: i32 = 2_147_483_647; // Default integer type
let i64_val: i64 = 9_223_372_036_854_775_807;
let i128_val: i128 = 170_141_183_460_469_231_731_687_303_715_884_105_727;
let isize_val: isize = 100;     // Architecture-dependent

// Unsigned integers
let u8_val: u8 = 255;           // 0 to 255
let u16_val: u16 = 65_535;      // 0 to 65,535
let u32_val: u32 = 4_294_967_295;
let u64_val: u64 = 18_446_744_073_709_551_615;
let u128_val: u128 = 340_282_366_920_938_463_463_374_607_431_768_211_455;
let usize_val: usize = 100;     // Architecture-dependent, used for indexing

// Integer literals with type suffixes
let typed_int = 42u32;
let hex = 0xff;
let octal = 0o77;
let binary = 0b1111_0000;
let byte = b'A'; // u8 only
```

#### Floating Point Types
```rust
let f32_val: f32 = 3.14159;     // Single precision
let f64_val: f64 = 2.718281828; // Double precision (default)

// Scientific notation
let large = 1.23e10f64;
let small = 1.23e-10f32;

// Special values
let infinity = f64::INFINITY;
let neg_infinity = f64::NEG_INFINITY;
let not_a_number = f64::NAN;

// Operations
let sum = 5.0 + 10.0;
let difference = 95.5 - 4.3;
let product = 4.0 * 30.0;
let quotient = 56.7 / 32.2;
let remainder = 43.0 % 5.0;
```

#### Boolean Type
```rust
let t: bool = true;
let f: bool = false;

// Boolean operations
let and_result = t && f;  // false
let or_result = t || f;   // true
let not_result = !t;      // false
```

#### Character Type
```rust
let c: char = 'z';
let unicode_char: char = '😀';
let heart_eyed_cat = '😻';

// Character literals
let a = 'a';
let newline = '\n';
let tab = '\t';
let unicode = '\u{1F600}'; // Unicode escape

// Character methods
println!("Is alphabetic: {}", c.is_alphabetic());
println!("Is numeric: {}", '5'.is_numeric());
println!("To uppercase: {}", c.to_uppercase().collect::<String>());
```

### Compound Types

#### Tuple Type
```rust
// Basic tuple
let tup: (i32, f64, u8) = (500, 6.4, 1);

// Destructuring
let (x, y, z) = tup;
println!("x: {}, y: {}, z: {}", x, y, z);

// Accessing by index
let five_hundred = tup.0;
let six_point_four = tup.1;
let one = tup.2;

// Unit tuple (empty tuple)
let unit: () = ();

// Single element tuple (note the comma)
let single = (42,);

// Nested tuples
let nested: ((i32, i32), (f64, f64)) = ((1, 2), (3.0, 4.0));
```

#### Array Type
```rust
// Fixed-size arrays
let a: [i32; 5] = [1, 2, 3, 4, 5];
let months = [\"January\", \"February\", \"March\", \"April\", \"May\",
              \"June\", \"July\", \"August\", \"September\", \"October\",
              \"November\", \"December\"];

// Array with repeated values
let zeros = [0; 10]; // [0, 0, 0, 0, 0, 0, 0, 0, 0, 0]

// Accessing elements
let first = a[0];
let second = a[1];

// Array methods
println!(\"Length: {}\", a.len());
println!(\"Is empty: {}\", a.is_empty());

// Iterating
for element in a {
    println!(\"Value: {}\", element);
}

// Slicing
let slice: &[i32] = &a[1..4]; // [2, 3, 4]
```

## String Types

### String vs &str
```rust
// String - owned, growable
let mut s = String::new();
s.push_str(\"hello\");
s.push(' ');
s.push('w');
s.push_str(\"orld\");

// String from literal
let s1 = String::from(\"hello\");
let s2 = \"hello\".to_string();

// &str - string slice, borrowed
let string_slice: &str = \"hello world\";
let slice_from_string: &str = &s1;

// String manipulation
let concatenated = format!(\"{} {}\", \"hello\", \"world\");
let replaced = s1.replace(\"l\", \"L\");
let uppercase = s1.to_uppercase();
let trimmed = \"  hello  \".trim();

// Slicing strings (be careful with UTF-8)
let hello = \"Здравствуй\";
let s = &hello[0..4]; // First 4 bytes, not characters!
```

### String Methods
```rust
let s = String::from(\"Hello, world!\");

// Length and capacity
println!(\"Length: {}\", s.len());           // 13 bytes
println!(\"Capacity: {}\", s.capacity());    // May be larger
println!(\"Is empty: {}\", s.is_empty());    // false

// Searching
println!(\"Contains 'world': {}\", s.contains(\"world\"));
println!(\"Starts with 'Hello': {}\", s.starts_with(\"Hello\"));
println!(\"Ends with '!': {}\", s.ends_with(\"!\"));

// Finding
if let Some(index) = s.find(\"world\") {
    println!(\"'world' starts at index {}\", index);
}

// Splitting
let words: Vec<&str> = s.split(\", \").collect();
let lines: Vec<&str> = \"line1\\nline2\\nline3\".lines().collect();

// Trimming
let padded = \"  hello  \";
println!(\"Trimmed: '{}'\", padded.trim());
```

## Collections

### Vec<T> - Dynamic Array
```rust
// Creating vectors
let mut v1: Vec<i32> = Vec::new();
let v2 = vec![1, 2, 3, 4, 5];
let v3 = vec![0; 10]; // 10 zeros

// Adding elements
v1.push(1);
v1.push(2);
v1.push(3);

// Accessing elements
let third = &v2[2];           // Panics if index out of bounds
let third_safe = v2.get(2);   // Returns Option<&T>

match third_safe {
    Some(value) => println!(\"Third element: {}\", value),
    None => println!(\"No third element\"),
}

// Modifying elements
v1[0] = 10;

// Iterating
for element in &v2 {
    println!(\"Value: {}\", element);
}

// Mutable iteration
for element in &mut v1 {
    *element += 50;
}

// Vector methods
println!(\"Length: {}\", v2.len());
println!(\"Capacity: {}\", v2.capacity());
println!(\"Is empty: {}\", v2.is_empty());

// Removing elements
let last = v1.pop(); // Returns Option<T>
v1.remove(0);        // Remove at index, shifts elements
v1.clear();          // Remove all elements

// Vector operations
let mut v4 = vec![1, 2, 3];
let mut v5 = vec![4, 5, 6];
v4.append(&mut v5);  // v4 now contains [1, 2, 3, 4, 5, 6]

// Slicing
let slice: &[i32] = &v2[1..4];
```

### HashMap<K, V> - Key-Value Store
```rust
use std::collections::HashMap;

// Creating HashMap
let mut scores = HashMap::new();
scores.insert(String::from(\"Blue\"), 10);
scores.insert(String::from(\"Yellow\"), 50);

// From vectors
let teams = vec![String::from(\"Blue\"), String::from(\"Yellow\")];
let initial_scores = vec![10, 50];
let scores: HashMap<_, _> = teams.into_iter()
    .zip(initial_scores.into_iter())
    .collect();

// Accessing values
let team_name = String::from(\"Blue\");
let score = scores.get(&team_name); // Returns Option<&V>

match score {
    Some(s) => println!(\"Blue team score: {}\", s),
    None => println!(\"Blue team not found\"),
}

// Iterating
for (key, value) in &scores {
    println!(\"{}: {}\", key, value);
}

// Updating values
scores.insert(String::from(\"Blue\"), 25); // Overwrites

// Insert only if key doesn't exist
scores.entry(String::from(\"Red\")).or_insert(5);

// Update based on old value
let count = scores.entry(String::from(\"Blue\")).or_insert(0);
*count += 10;

// HashMap methods
println!(\"Contains Blue: {}\", scores.contains_key(\"Blue\"));
println!(\"Length: {}\", scores.len());
println!(\"Is empty: {}\", scores.is_empty());

// Removing
let removed = scores.remove(\"Yellow\");
scores.clear();
```

### HashSet<T> - Unique Values
```rust
use std::collections::HashSet;

// Creating HashSet
let mut books = HashSet::new();
books.insert(String::from(\"A Game of Thrones\"));
books.insert(String::from(\"To Kill a Mockingbird\"));
books.insert(String::from(\"The Lord of the Rings\"));

// From vector
let list = vec![1, 2, 3, 2, 4, 3, 5];
let unique: HashSet<_> = list.into_iter().collect();

// Checking membership
if books.contains(\"To Kill a Mockingbird\") {
    println!(\"Found the book!\");
}

// Set operations
let set1: HashSet<i32> = [1, 2, 3, 4].iter().cloned().collect();
let set2: HashSet<i32> = [3, 4, 5, 6].iter().cloned().collect();

// Union
let union: HashSet<_> = set1.union(&set2).collect();

// Intersection
let intersection: HashSet<_> = set1.intersection(&set2).collect();

// Difference
let difference: HashSet<_> = set1.difference(&set2).collect();

// Symmetric difference
let sym_diff: HashSet<_> = set1.symmetric_difference(&set2).collect();
```

### BTreeMap<K, V> and BTreeSet<T> - Sorted Collections
```rust
use std::collections::{BTreeMap, BTreeSet};

// BTreeMap - sorted by keys
let mut map = BTreeMap::new();
map.insert(3, \"three\");
map.insert(1, \"one\");
map.insert(2, \"two\");

// Iteration is in sorted order
for (key, value) in &map {
    println!(\"{}: {}\", key, value); // 1: one, 2: two, 3: three
}

// Range queries
let range: BTreeMap<_, _> = map.range(1..=2).collect();

// BTreeSet - sorted unique values
let mut set = BTreeSet::new();
set.insert(3);
set.insert(1);
set.insert(2);

for value in &set {
    println!(\"Value: {}\", value); // 1, 2, 3
}
```

### VecDeque<T> - Double-Ended Queue
```rust
use std::collections::VecDeque;

let mut deque = VecDeque::new();

// Push to both ends
deque.push_back(1);
deque.push_back(2);
deque.push_front(0);

// Pop from both ends
let back = deque.pop_back();   // Some(2)
let front = deque.pop_front(); // Some(0)

// Access by index
let middle = deque[0]; // 1

// Use as ring buffer
let mut ring_buffer = VecDeque::with_capacity(5);
for i in 0..10 {
    if ring_buffer.len() == 5 {
        ring_buffer.pop_front();
    }
    ring_buffer.push_back(i);
}
```

### BinaryHeap<T> - Priority Queue
```rust
use std::collections::BinaryHeap;

// Max heap by default
let mut heap = BinaryHeap::new();
heap.push(4);
heap.push(1);
heap.push(3);
heap.push(2);

while let Some(max) = heap.pop() {
    println!(\"Max: {}\", max); // 4, 3, 2, 1
}

// Min heap using Reverse wrapper
use std::cmp::Reverse;
let mut min_heap = BinaryHeap::new();
min_heap.push(Reverse(4));
min_heap.push(Reverse(1));
min_heap.push(Reverse(3));

while let Some(Reverse(min)) = min_heap.pop() {
    println!(\"Min: {}\", min); // 1, 3, 4
}
```

## Advanced Types

### Option<T>
```rust
// Represents optional values
let some_number = Some(5);
let some_string = Some(\"a string\");
let absent_number: Option<i32> = None;

// Pattern matching
match some_number {
    Some(i) => println!(\"Number: {}\", i),
    None => println!(\"No number\"),
}

// Convenient methods
let x: Option<i32> = Some(2);
assert_eq!(x.is_some(), true);
assert_eq!(x.is_none(), false);
assert_eq!(x.unwrap(), 2);
assert_eq!(x.unwrap_or(0), 2);
assert_eq!(x.map(|i| i * 2), Some(4));
```

### Result<T, E>
```rust
// Represents success or failure
fn divide(a: f64, b: f64) -> Result<f64, String> {
    if b != 0.0 {
        Ok(a / b)
    } else {
        Err(\"Division by zero\".to_string())
    }
}

match divide(10.0, 2.0) {
    Ok(result) => println!(\"Result: {}\", result),
    Err(error) => println!(\"Error: {}\", error),
}

// Convenient methods
let result = divide(10.0, 2.0);
assert!(result.is_ok());
assert_eq!(result.unwrap(), 5.0);
assert_eq!(result.unwrap_or(0.0), 5.0);
```

## Custom Types

### Structs
```rust
// Named field struct
struct User {
    username: String,
    email: String,
    sign_in_count: u64,
    active: bool,
}

let user = User {
    email: String::from(\"someone@example.com\"),
    username: String::from(\"someusername123\"),
    active: true,
    sign_in_count: 1,
};

// Tuple struct
struct Color(i32, i32, i32);
struct Point(i32, i32, i32);

let black = Color(0, 0, 0);
let origin = Point(0, 0, 0);

// Unit struct
struct AlwaysEqual;
let subject = AlwaysEqual;

// Struct methods
impl User {
    fn full_info(&self) -> String {
        format!(\"{} <{}>\", self.username, self.email)
    }

    fn new(email: String, username: String) -> User {
        User {
            email,
            username,
            active: true,
            sign_in_count: 1,
        }
    }
}
```

### Enums
```rust
// Basic enum
enum IpAddrKind {
    V4,
    V6,
}

// Enum with data
enum IpAddr {
    V4(u8, u8, u8, u8),
    V6(String),
}

let home = IpAddr::V4(127, 0, 0, 1);
let loopback = IpAddr::V6(String::from(\"::1\"));

// Complex enum
enum Message {
    Quit,
    Move { x: i32, y: i32 },
    Write(String),
    ChangeColor(i32, i32, i32),
}

// Enum methods
impl Message {
    fn call(&self) {
        match self {
            Message::Quit => println!(\"Quit\"),
            Message::Move { x, y } => println!(\"Move to ({}, {})\", x, y),
            Message::Write(text) => println!(\"Text: {}\", text),
            Message::ChangeColor(r, g, b) => println!(\"Color: ({}, {}, {})\", r, g, b),
        }
    }
}
```

## Generic Types
```rust
// Generic struct
struct Point<T> {
    x: T,
    y: T,
}

let integer_point = Point { x: 5, y: 10 };
let float_point = Point { x: 1.0, y: 4.0 };

// Multiple generic parameters
struct Rectangle<T, U> {
    width: T,
    height: U,
}

// Generic enum
enum Option<T> {
    Some(T),
    None,
}

enum Result<T, E> {
    Ok(T),
    Err(E),
}

// Generic functions
fn largest<T: PartialOrd + Copy>(list: &[T]) -> T {
    let mut largest = list[0];

    for &item in list {
        if item > largest {
            largest = item;
        }
    }

    largest
}

// Generic implementations
impl<T> Point<T> {
    fn x(&self) -> &T {
        &self.x
    }
}

// Specific implementations
impl Point<f32> {
    fn distance_from_origin(&self) -> f32 {
        (self.x.powi(2) + self.y.powi(2)).sqrt()
    }
}
```

## Type Aliases
```rust
// Type alias for readability
type Kilometers = i32;
type Result<T> = std::result::Result<T, std::io::Error>;

let distance: Kilometers = 100;

fn read_file() -> Result<String> {
    std::fs::read_to_string(\"file.txt\")
}

// Complex type alias
type Thunk = Box<dyn Fn() + Send + 'static>;
let f: Thunk = Box::new(|| println!(\"hi\"));
```

## Trait Objects and Dynamic Dispatch
```rust
trait Draw {
    fn draw(&self);
}

struct Circle {
    radius: f64,
}

struct Rectangle {
    width: f64,
    height: f64,
}

impl Draw for Circle {
    fn draw(&self) {
        println!(\"Drawing circle with radius {}\", self.radius);
    }
}

impl Draw for Rectangle {
    fn draw(&self) {
        println!(\"Drawing rectangle {}x{}\", self.width, self.height);
    }
}

// Trait object - dynamic dispatch
let shapes: Vec<Box<dyn Draw>> = vec![
    Box::new(Circle { radius: 5.0 }),
    Box::new(Rectangle { width: 10.0, height: 20.0 }),
];

for shape in shapes {
    shape.draw();
}
```

This comprehensive overview covers Rust's type system and collections, providing the foundation for understanding data organization and manipulation in Rust programs.