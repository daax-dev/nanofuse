# Memory Management in Rust

## Core Memory Safety Philosophy

Rust achieves **memory safety without garbage collection** through:
- **Compile-time ownership checking**
- **Zero-cost abstractions**
- **Explicit memory management**
- **Prevention of undefined behavior**

## The Ownership System

### Three Core Rules
1. **Each value has exactly one owner**
2. **There can be multiple immutable references OR one mutable reference**
3. **References must be valid for their entire lifetime**

### Ownership Transfer (Move Semantics)
```rust
// String ownership transfer
let s1 = String::from("hello");
let s2 = s1;  // s1 moved to s2, s1 no longer valid
// println!("{}", s1); // ERROR: value borrowed after move

// Copy types (stack-allocated) are copied, not moved
let x = 5;
let y = x;  // x is copied to y, both remain valid
println!("{} and {}", x, y); // OK
```

### Clone vs Copy
```rust
// Clone: explicit deep copy
let s1 = String::from("hello");
let s2 = s1.clone();  // Expensive heap allocation
println!("{} and {}", s1, s2); // Both valid

// Copy trait: implicit bitwise copy for stack types
#[derive(Copy, Clone)]
struct Point {
    x: i32,
    y: i32,
}

let p1 = Point { x: 1, y: 2 };
let p2 = p1;  // Copied automatically
println!("{:?} and {:?}", p1, p2); // Both valid
```

## Borrowing and References

### Immutable References
```rust
let s = String::from("hello");
let r1 = &s;      // Immutable reference
let r2 = &s;      // Multiple immutable refs OK
println!("{} and {}", r1, r2);
// s is still valid here
```

### Mutable References
```rust
let mut s = String::from("hello");
let r = &mut s;   // Mutable reference
r.push_str(" world");
// Cannot use s while r is in scope
// println!("{}", s); // ERROR: cannot borrow as immutable
println!("{}", r); // OK
// r goes out of scope here, s can be used again
println!("{}", s); // OK
```

### Reference Rules Enforcement
```rust
let mut s = String::from("hello");

// VALID: Multiple immutable references
let r1 = &s;
let r2 = &s;
println!("{} and {}", r1, r2);
// r1 and r2 go out of scope here

// VALID: One mutable reference after immutable refs end
let r3 = &mut s;
println!("{}", r3);

// INVALID: Cannot mix mutable and immutable references
/*
let r4 = &s;      // Immutable reference
let r5 = &mut s;  // ERROR: cannot borrow as mutable
*/
```

## Lifetimes

### Lifetime Annotations
```rust
// Function with lifetime parameters
fn longest<'a>(x: &'a str, y: &'a str) -> &'a str {
    if x.len() > y.len() {
        x
    } else {
        y
    }
}

// Multiple lifetimes
fn complex<'a, 'b>(x: &'a str, y: &'b str) -> &'a str {
    x  // Return value tied to 'a lifetime
}
```

### Lifetime in Structs
```rust
struct ImportantExcerpt<'a> {
    part: &'a str,
}

impl<'a> ImportantExcerpt<'a> {
    fn level(&self) -> i32 {
        3
    }

    // Lifetime elision - input and output lifetimes
    fn announce_and_return_part(&self, announcement: &str) -> &str {
        println!("Attention please: {}", announcement);
        self.part
    }
}

fn main() {
    let novel = String::from("Call me Ishmael. Some years ago...");
    let first_sentence = novel.split('.').next().expect("Could not find a '.'");
    let i = ImportantExcerpt {
        part: first_sentence,
    };
    // novel must live as long as i
}
```

### Static Lifetime
```rust
// String literals have 'static lifetime
let s: &'static str = "I have a static lifetime.";

// Static variables
static HELLO_WORLD: &str = "Hello, world!";

// Functions can require 'static lifetime
fn takes_static(s: &'static str) {
    println!("{}", s);
}
```

## Memory Allocation Strategies

### Stack Allocation
```rust
// Stack-allocated data
let x = 5;              // i32 on stack
let arr = [1, 2, 3, 4]; // Array on stack
let tuple = (1, "hello"); // Tuple on stack

// Fixed-size types go on stack
struct Point {
    x: f64,
    y: f64,
}
let point = Point { x: 1.0, y: 2.0 }; // Stack allocated
```

### Heap Allocation with Box<T>
```rust
// Heap allocation
let b = Box::new(5);        // i32 on heap
let large_box = Box::new([0; 1000]); // Large array on heap

// Recursive types require heap allocation
#[derive(Debug)]
enum List {
    Cons(i32, Box<List>),
    Nil,
}

let list = List::Cons(1,
    Box::new(List::Cons(2,
        Box::new(List::Cons(3,
            Box::new(List::Nil))))));
```

### Dynamic Collections
```rust
// Vec<T> - growable array
let mut v = Vec::new();
v.push(1);
v.push(2);
v.push(3);

// String - growable UTF-8 string
let mut s = String::new();
s.push_str("hello");
s.push(' ');
s.push_str("world");

// HashMap - key-value pairs
use std::collections::HashMap;
let mut scores = HashMap::new();
scores.insert(String::from("Blue"), 10);
scores.insert(String::from("Yellow"), 50);
```

## Smart Pointers

### Reference Counting - Rc<T>
```rust
use std::rc::Rc;

let data = Rc::new(String::from("shared data"));
let reference1 = Rc::clone(&data);  // Increment reference count
let reference2 = Rc::clone(&data);  // Increment reference count

println!("Reference count: {}", Rc::strong_count(&data)); // 3
// When all Rc pointers drop, data is deallocated
```

### Interior Mutability - RefCell<T>
```rust
use std::cell::RefCell;
use std::rc::Rc;

#[derive(Debug)]
struct Node {
    value: i32,
    children: RefCell<Vec<Rc<Node>>>,
}

let leaf = Rc::new(Node {
    value: 3,
    children: RefCell::new(vec![]),
});

let branch = Rc::new(Node {
    value: 5,
    children: RefCell::new(vec![Rc::clone(&leaf)]),
});

// Mutate through immutable reference
leaf.children.borrow_mut().push(Rc::clone(&branch));
```

### Atomic Reference Counting - Arc<T>
```rust
use std::sync::Arc;
use std::thread;

let counter = Arc::new(5);
let mut handles = vec![];

for _ in 0..10 {
    let counter = Arc::clone(&counter);
    let handle = thread::spawn(move || {
        println!("Counter value: {}", counter);
    });
    handles.push(handle);
}

for handle in handles {
    handle.join().unwrap();
}
```

### Mutex for Thread-Safe Mutation
```rust
use std::sync::{Arc, Mutex};
use std::thread;

let counter = Arc::new(Mutex::new(0));
let mut handles = vec![];

for _ in 0..10 {
    let counter = Arc::clone(&counter);
    let handle = thread::spawn(move || {
        let mut num = counter.lock().unwrap();
        *num += 1;
    });
    handles.push(handle);
}

for handle in handles {
    handle.join().unwrap();
}

println!("Result: {}", *counter.lock().unwrap());
```

## Memory Layout and Optimization

### Data Structure Sizes
```rust
use std::mem;

println!("i32: {} bytes", mem::size_of::<i32>());         // 4
println!("i64: {} bytes", mem::size_of::<i64>());         // 8
println!("&i32: {} bytes", mem::size_of::<&i32>());       // 8 (64-bit)
println!("String: {} bytes", mem::size_of::<String>());   // 24
println!("Vec<i32>: {} bytes", mem::size_of::<Vec<i32>>()); // 24
println!("Option<i32>: {} bytes", mem::size_of::<Option<i32>>()); // 8
println!("Result<i32, String>: {} bytes", mem::size_of::<Result<i32, String>>()); // 32
```

### Alignment and Padding
```rust
#[repr(C)]
struct Padded {
    a: u8,  // 1 byte
    // 3 bytes padding
    b: u32, // 4 bytes
    c: u8,  // 1 byte
    // 3 bytes padding
}
// Total: 12 bytes due to alignment requirements

#[repr(packed)]
struct Packed {
    a: u8,  // 1 byte
    b: u32, // 4 bytes
    c: u8,  // 1 byte
}
// Total: 6 bytes (no padding)
```

## Unsafe Rust and Manual Memory Management

### Raw Pointers
```rust
let mut num = 5;

// Create raw pointers
let r1 = &num as *const i32;     // Immutable raw pointer
let r2 = &mut num as *mut i32;   // Mutable raw pointer

// Dereference raw pointers (unsafe)
unsafe {
    println!("r1 is: {}", *r1);
    println!("r2 is: {}", *r2);
}
```

### Manual Memory Allocation
```rust
use std::alloc::{alloc, dealloc, Layout};

unsafe {
    // Allocate memory for 4 i32 values
    let layout = Layout::array::<i32>(4).unwrap();
    let ptr = alloc(layout) as *mut i32;

    if ptr.is_null() {
        panic!("Memory allocation failed");
    }

    // Initialize the memory
    for i in 0..4 {
        *ptr.add(i) = i as i32;
    }

    // Use the memory
    for i in 0..4 {
        println!("Value at index {}: {}", i, *ptr.add(i));
    }

    // Must manually deallocate
    dealloc(ptr as *mut u8, layout);
}
```

### Foreign Function Interface (FFI)
```rust
extern "C" {
    fn malloc(size: usize) -> *mut u8;
    fn free(ptr: *mut u8);
}

unsafe {
    let ptr = malloc(16);
    if !ptr.is_null() {
        // Use the allocated memory
        free(ptr);
    }
}
```

## Memory Safety Violations (Prevented)

### Use After Free (Prevented)
```rust
// This doesn't compile - Rust prevents use after free
/*
let reference_to_nothing = {
    let x = 5;
    &x  // ERROR: borrowed value does not live long enough
}; // x goes out of scope here
println!("{}", reference_to_nothing);
*/
```

### Double Free (Prevented)
```rust
// Rust prevents double free through ownership
let s1 = String::from("hello");
let s2 = s1; // Ownership moved, s1 no longer accessible
// drop(s1); // ERROR: value used after move
drop(s2); // OK, s2 owns the data
```

### Buffer Overflow (Prevented)
```rust
let v = vec![1, 2, 3];
// let value = v[10]; // Panics at runtime (bounds checking)

// Safe alternatives
if let Some(value) = v.get(10) {
    println!("Value: {}", value);
} else {
    println!("Index out of bounds");
}
```

## Performance Characteristics

### Zero-Cost Abstractions
```rust
// Iterator chains compile to efficient loops
let sum: i32 = (0..1_000_000)
    .filter(|x| x % 2 == 0)
    .map(|x| x * x)
    .sum();

// Equivalent to:
// let mut sum = 0;
// for i in 0..1_000_000 {
//     if i % 2 == 0 {
//         sum += i * i;
//     }
// }
```

### Memory Overhead
- **No garbage collector**: No GC pauses or overhead
- **Deterministic destructors**: RAII pattern with Drop trait
- **Stack allocation preferred**: Faster than heap allocation
- **Minimal runtime**: No hidden allocations

## Drop Trait and RAII

### Automatic Cleanup
```rust
struct FileWrapper {
    file: std::fs::File,
}

impl Drop for FileWrapper {
    fn drop(&mut self) {
        println!("FileWrapper is being dropped");
        // File is automatically closed when FileWrapper drops
    }
}

{
    let _wrapper = FileWrapper {
        file: std::fs::File::open("example.txt").unwrap(),
    };
    // File is automatically closed when _wrapper goes out of scope
}
```

### Custom Drop Implementation
```rust
struct CustomSmartPointer {
    data: String,
}

impl Drop for CustomSmartPointer {
    fn drop(&mut self) {
        println!("Dropping CustomSmartPointer with data `{}`!", self.data);
    }
}

fn main() {
    let c = CustomSmartPointer {
        data: String::from("my stuff"),
    };
    let d = CustomSmartPointer {
        data: String::from("other stuff"),
    };
    println!("CustomSmartPointers created.");
    // d drops first, then c (reverse order of creation)
}
```

## Memory Debugging and Tools

### Debugging Tools
- **Miri**: Interpreter for detecting undefined behavior
- **AddressSanitizer**: Runtime error detector
- **Valgrind**: Memory error detector (limited Rust support)

### Compile-Time Checks
```rust
// Rust prevents these at compile time:
// - Use after free
// - Double free
// - Buffer overflows
// - Data races
// - Null pointer dereferences
// - Memory leaks (in safe code)
```

## Best Practices

1. **Prefer stack allocation** over heap when possible
2. **Use ownership transfer** instead of cloning when performance matters
3. **Minimize lifetime complexity** through good design
4. **Use Rc/Arc only when necessary** for shared ownership
5. **Avoid unsafe code** unless absolutely required
6. **Use tools like clippy** to catch potential issues
7. **Profile memory usage** in performance-critical applications
8. **Understand the cost** of different allocation strategies
9. **Use appropriate smart pointers** for your use case
10. **Test with different allocation patterns** under load

This comprehensive understanding of Rust's memory management enables writing safe, efficient, and predictable programs without garbage collection overhead.