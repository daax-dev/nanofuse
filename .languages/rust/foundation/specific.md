# Rust-Specific Unique Features

## Ownership System - Rust's Defining Feature

### Zero-Cost Memory Safety
```rust
// Automatic memory management without garbage collection
fn process_data() -> String {
    let mut data = String::from(\"processing\");
    data.push_str(\" complete\");
    data // Ownership transferred out, no explicit cleanup needed
} // Memory automatically freed when data goes out of scope
```

### Move Semantics by Default
```rust
fn demonstrate_move() {
    let s1 = String::from(\"hello\");
    let s2 = s1; // Move, not copy - prevents double-free
    // println!(\"{}\", s1); // Compile error: value used after move
    println!(\"{}\", s2); // OK
}
```

### Borrow Checker - Compile-Time Reference Validation
```rust
fn borrow_checker_example() {
    let mut data = vec![1, 2, 3];

    // Multiple immutable borrows allowed
    let r1 = &data;
    let r2 = &data;
    println!(\"{:?} {:?}\", r1, r2);

    // Mutable borrow after immutable borrows end
    let r3 = &mut data;
    r3.push(4);
    // Cannot use r1, r2 here - borrow checker prevents it
}
```

## Traits - Rust's Interface System

### Trait Definitions and Implementations
```rust
// Define behavior without implementation details
trait Drawable {
    fn draw(&self);

    // Default implementation
    fn area(&self) -> f64 {
        0.0
    }
}

struct Circle { radius: f64 }
struct Rectangle { width: f64, height: f64 }

impl Drawable for Circle {
    fn draw(&self) {
        println!(\"Drawing circle with radius {}\", self.radius);
    }

    fn area(&self) -> f64 {
        3.14159 * self.radius * self.radius
    }
}

impl Drawable for Rectangle {
    fn draw(&self) {
        println!(\"Drawing rectangle {}x{}\", self.width, self.height);
    }

    fn area(&self) -> f64 {
        self.width * self.height
    }
}
```

### Trait Bounds and Generic Constraints
```rust
// Function with trait bounds
fn print_area<T: Drawable>(shape: T) {
    shape.draw();
    println!(\"Area: {}\", shape.area());
}

// Multiple trait bounds
fn complex_function<T>(item: T)
where
    T: Drawable + Clone + std::fmt::Debug
{
    item.draw();
    let copy = item.clone();
    println!(\"Debug: {:?}\", copy);
}
```

### Associated Types
```rust
trait Iterator {
    type Item; // Associated type

    fn next(&mut self) -> Option<Self::Item>;
}

struct Counter {
    current: usize,
    max: usize,
}

impl Iterator for Counter {
    type Item = usize; // Concrete associated type

    fn next(&mut self) -> Option<Self::Item> {
        if self.current < self.max {
            let current = self.current;
            self.current += 1;
            Some(current)
        } else {
            None
        }
    }
}
```

### Operator Overloading Through Traits
```rust
use std::ops::Add;

#[derive(Debug, Clone, Copy)]
struct Point {
    x: i32,
    y: i32,
}

impl Add for Point {
    type Output = Point;

    fn add(self, other: Point) -> Point {
        Point {
            x: self.x + other.x,
            y: self.y + other.y,
        }
    }
}

let p1 = Point { x: 1, y: 2 };
let p2 = Point { x: 3, y: 4 };
let p3 = p1 + p2; // Uses the Add trait
println!(\"{:?}\", p3); // Point { x: 4, y: 6 }
```

## Pattern Matching and Destructuring

### Match Expression - Exhaustive Pattern Matching
```rust
enum Coin {
    Penny,
    Nickel,
    Dime,
    Quarter(String), // With associated data
}

fn value_in_cents(coin: Coin) -> u8 {
    match coin {
        Coin::Penny => {
            println!(\"Lucky penny!\");
            1
        },
        Coin::Nickel => 5,
        Coin::Dime => 10,
        Coin::Quarter(state) => {
            println!(\"State quarter from {}!\", state);
            25
        },
    }
}
```

### Advanced Pattern Matching
```rust
// Pattern matching with guards
fn categorize_number(x: Option<i32>) -> String {
    match x {
        Some(n) if n < 0 => \"Negative\".to_string(),
        Some(n) if n > 0 => \"Positive\".to_string(),
        Some(0) => \"Zero\".to_string(),
        None => \"No number\".to_string(),
    }
}

// Destructuring complex types
struct Point3D { x: i32, y: i32, z: i32 }

fn process_point(point: Point3D) {
    match point {
        Point3D { x: 0, y: 0, z: 0 } => println!(\"Origin\"),
        Point3D { x, y: 0, z: 0 } => println!(\"On x-axis at {}\", x),
        Point3D { x: 0, y, z: 0 } => println!(\"On y-axis at {}\", y),
        Point3D { x: 0, y: 0, z } => println!(\"On z-axis at {}\", z),
        Point3D { x, y, z } => println!(\"Point at ({}, {}, {})\", x, y, z),
    }
}

// Range patterns
fn classify_age(age: u8) -> &'static str {
    match age {
        0..=12 => \"Child\",
        13..=19 => \"Teenager\",
        20..=64 => \"Adult\",
        65..=u8::MAX => \"Senior\",
    }
}
```

## Lifetimes - Compile-Time Reference Validation

### Lifetime Annotations
```rust
// Explicit lifetime parameters
fn longest<'a>(x: &'a str, y: &'a str) -> &'a str {
    if x.len() > y.len() {
        x
    } else {
        y
    }
}

// Multiple lifetime parameters
fn announce_and_return<'a, 'b>(
    announcement: &'a str,
    value: &'b str
) -> &'b str {
    println!(\"Attention: {}\", announcement);
    value
}
```

### Lifetime Elision Rules
```rust
// These don't need explicit lifetime annotations due to elision rules:

// Rule 1: Each parameter gets its own lifetime
fn first_word(s: &str) -> &str { // Becomes: fn first_word<'a>(s: &'a str) -> &'a str
    &s[..1]
}

// Rule 2: If there's exactly one input lifetime, it's assigned to all outputs
fn process(input: &str) -> &str { // Becomes: fn process<'a>(input: &'a str) -> &'a str
    input
}

// Rule 3: If one parameter is &self or &mut self, its lifetime is assigned to all outputs
impl<T> MyStruct<T> {
    fn get_data(&self) -> &T { // Becomes: fn get_data<'a>(&'a self) -> &'a T
        &self.data
    }
}
```

### Static Lifetime
```rust
// String literals have 'static lifetime
let s: &'static str = \"I live for the entire program\";

// Static variables
static GLOBAL_VALUE: i32 = 42;

// Functions requiring static lifetime
fn store_reference<T: 'static>(data: T) -> T {
    data // T must live for the entire program
}
```

## Macros - Compile-Time Code Generation

### Declarative Macros (macro_rules!)
```rust
// Simple macro
macro_rules! say_hello {
    () => {
        println!(\"Hello, Rust!\");
    };
}

// Macro with parameters
macro_rules! create_function {
    ($func_name:ident) => {
        fn $func_name() {
            println!(\"You called {:?}()\", stringify!($func_name));
        }
    };
}

create_function!(foo);
create_function!(bar);

// Variable argument macros
macro_rules! find_min {
    ($x:expr) => ($x);
    ($x:expr, $($y:expr),+) => (
        std::cmp::min($x, find_min!($($y),+))
    );
}

let min = find_min!(5, 2, 8, 1, 9);
```

### Procedural Macros
```rust
use proc_macro::TokenStream;
use quote::quote;
use syn;

// Derive macro example (simplified)
#[proc_macro_derive(HelloMacro)]
pub fn hello_macro_derive(input: TokenStream) -> TokenStream {
    let ast = syn::parse(input).unwrap();
    impl_hello_macro(&ast)
}

fn impl_hello_macro(ast: &syn::DeriveInput) -> TokenStream {
    let name = &ast.ident;
    let gen = quote! {
        impl HelloMacro for #name {
            fn hello_macro() {
                println!(\"Hello, Macro! My name is {}!\", stringify!(#name));
            }
        }
    };
    gen.into()
}
```

## Cargo and Package Management

### Cargo.toml Configuration
```toml
[package]
name = \"my_project\"
version = \"0.1.0\"
edition = \"2021\"
authors = [\"Your Name <your.email@example.com>\"]
description = \"A sample Rust project\"
license = \"MIT OR Apache-2.0\"

[dependencies]
serde = { version = \"1.0\", features = [\"derive\"] }
tokio = { version = \"1.0\", features = [\"full\"] }
clap = \"4.0\"

[dev-dependencies]
criterion = \"0.4\"

[[bin]]
name = \"main_binary\"
path = \"src/main.rs\"

[[bin]]
name = \"other_binary\"
path = \"src/bin/other.rs\"

[profile.release]
lto = true
codegen-units = 1
```

### Workspace Management
```toml
# Workspace Cargo.toml
[workspace]
members = [
    \"crate_a\",
    \"crate_b\",
    \"tools/cli\",
]

[workspace.dependencies]
serde = \"1.0\"
tokio = \"1.0\"
```

## Attributes and Conditional Compilation

### Common Attributes
```rust
// Conditional compilation
#[cfg(target_os = \"windows\")]
fn windows_only_function() {
    println!(\"This only runs on Windows\");
}

#[cfg(feature = \"extra_functionality\")]
fn optional_feature() {
    println!(\"This requires the 'extra_functionality' feature\");
}

// Testing attributes
#[cfg(test)]
mod tests {
    #[test]
    fn test_something() {
        assert_eq!(2 + 2, 4);
    }

    #[test]
    #[should_panic]
    fn test_panic() {
        panic!(\"This test should panic\");
    }

    #[test]
    #[ignore]
    fn expensive_test() {
        // This test is ignored by default
    }
}

// Derive attributes
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
struct MyStruct {
    field: String,
}

// Allow/warn/deny attributes
#[allow(dead_code)]
fn unused_function() {}

#[warn(unused_variables)]
fn some_function() {
    let unused_var = 42; // Will generate warning
}

#[deny(unsafe_code)]
mod safe_module {
    // Any unsafe code in this module will cause compilation error
}
```

## Foreign Function Interface (FFI)

### Calling C Functions
```rust
// Declare external C functions
extern \"C\" {
    fn abs(input: i32) -> i32;
    fn sqrt(input: f64) -> f64;
}

fn main() {
    unsafe {
        println!(\"Absolute value of -3 according to C: {}\", abs(-3));
        println!(\"Square root of 9 according to C: {}\", sqrt(9.0));
    }
}
```

### Exposing Rust Functions to C
```rust
// Export Rust function to C
#[no_mangle]
pub extern \"C\" fn add_numbers(a: i32, b: i32) -> i32 {
    a + b
}

// Using repr(C) for C-compatible structs
#[repr(C)]
pub struct Point {
    x: f64,
    y: f64,
}

#[no_mangle]
pub extern \"C\" fn create_point(x: f64, y: f64) -> Point {
    Point { x, y }
}
```

## Unsafe Rust

### Unsafe Capabilities
```rust
unsafe {
    // 1. Dereference raw pointers
    let raw_ptr = 0x12345678 as *const i32;
    // let value = *raw_ptr; // Extremely dangerous!

    // 2. Call unsafe functions
    let mut vec = vec![1, 2, 3, 4, 5];
    let ptr = vec.as_mut_ptr();
    let len = vec.len();

    // Create slice from raw parts
    let slice = std::slice::from_raw_parts_mut(ptr, len);
    slice[0] = 10;

    // 3. Implement unsafe traits
    // (see Send/Sync examples below)

    // 4. Access/modify mutable static variables
    static mut COUNTER: usize = 0;
    COUNTER += 1;
    println!(\"COUNTER: {}\", COUNTER);

    // 5. Access union fields
    union MyUnion {
        f1: u32,
        f2: f32,
    }

    let mut u = MyUnion { f1: 1 };
    u.f1 = 5;
    let f = u.f2; // Accessing union field is unsafe
}
```

### Safe Abstractions Over Unsafe Code
```rust
struct SafeWrapper {
    data: *mut u8,
    len: usize,
    capacity: usize,
}

impl SafeWrapper {
    fn new() -> Self {
        SafeWrapper {
            data: std::ptr::null_mut(),
            len: 0,
            capacity: 0,
        }
    }

    fn push(&mut self, value: u8) {
        unsafe {
            // Unsafe implementation hidden behind safe interface
            if self.len == self.capacity {
                self.grow();
            }

            std::ptr::write(self.data.add(self.len), value);
            self.len += 1;
        }
    }

    unsafe fn grow(&mut self) {
        // Growth implementation
        let new_capacity = if self.capacity == 0 { 1 } else { self.capacity * 2 };
        let new_data = std::alloc::alloc(
            std::alloc::Layout::array::<u8>(new_capacity).unwrap()
        ) as *mut u8;

        if !self.data.is_null() {
            std::ptr::copy(self.data, new_data, self.len);
            std::alloc::dealloc(
                self.data as *mut u8,
                std::alloc::Layout::array::<u8>(self.capacity).unwrap()
            );
        }

        self.data = new_data;
        self.capacity = new_capacity;
    }
}
```

## Zero-Cost Abstractions Examples

### Iterator Optimization
```rust
// High-level iterator code
let numbers: Vec<i32> = (0..1_000_000).collect();
let sum: i32 = numbers
    .iter()
    .filter(|&&x| x % 2 == 0)
    .map(|&x| x * x)
    .sum();

// Compiles to equivalent of:
// let mut sum = 0;
// for i in 0..1_000_000 {
//     if i % 2 == 0 {
//         sum += i * i;
//     }
// }
```

### Generic Monomorphization
```rust
fn generic_function<T: std::fmt::Display>(value: T) {
    println!(\"Value: {}\", value);
}

// Each call creates a specialized version at compile time
generic_function(42);        // Creates generic_function_i32
generic_function(\"hello\");    // Creates generic_function_str
generic_function(3.14);      // Creates generic_function_f64
```

## Rust's Module System

### Module Organization
```rust
// src/lib.rs or src/main.rs
mod front_of_house {
    pub mod hosting {
        pub fn add_to_waitlist() {}
        fn seat_at_table() {}
    }

    mod serving {
        fn take_order() {}
        fn serve_order() {}
        fn take_payment() {}
    }
}

pub fn eat_at_restaurant() {
    // Absolute path
    crate::front_of_house::hosting::add_to_waitlist();

    // Relative path
    front_of_house::hosting::add_to_waitlist();
}
```

### Use Declarations and Re-exports
```rust
// Bringing items into scope
use std::collections::HashMap;
use std::collections::*; // Glob import
use std::io::{self, Write}; // Import module and trait
use std::fmt::Result as FmtResult; // Aliasing

// Re-exporting
pub use crate::front_of_house::hosting;

// External crate usage
use rand::Rng;
use serde::{Serialize, Deserialize};
```

## Edition System - Backward Compatible Evolution

### Edition Migration
```rust
// Rust 2015 (original)
extern crate serde;

// Rust 2018 improvements
use serde::Serialize; // No need for extern crate

// async/await syntax
async fn fetch_data() -> String {
    // Implementation
}

// Rust 2021 additions
// Disjoint captures in closures
// IntoIterator for arrays
// Panic macro consistency
```

These unique features make Rust distinct from other programming languages, providing memory safety, zero-cost abstractions, and powerful expressiveness while maintaining high performance and preventing common programming errors at compile time.