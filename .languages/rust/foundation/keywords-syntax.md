# Rust Keywords and Syntax Reference

## Keywords

### Strict Keywords
These keywords cannot be used as identifiers in any context.

#### Control Flow Keywords
- **`if`** - Conditional expression
- **`else`** - Alternative branch in conditional
- **`match`** - Pattern matching expression
- **`loop`** - Infinite loop
- **`while`** - Conditional loop
- **`for`** - Iterator-based loop
- **`break`** - Exit loop early
- **`continue`** - Skip to next loop iteration
- **`return`** - Return value from function

#### Declaration and Definition Keywords
- **`fn`** - Function definition
- **`let`** - Variable binding
- **`mut`** - Mutable binding modifier
- **`const`** - Compile-time constant
- **`static`** - Static variable
- **`struct`** - Structure definition
- **`enum`** - Enumeration definition
- **`union`** - Union definition (unsafe)
- **`trait`** - Trait definition
- **`impl`** - Implementation block
- **`type`** - Type alias

#### Module and Visibility Keywords
- **`mod`** - Module definition
- **`use`** - Import items into scope
- **`pub`** - Public visibility
- **`crate`** - Current crate root
- **`super`** - Parent module
- **`self`** - Current module or method receiver
- **`Self`** - Current type in trait/impl

#### Ownership and Borrowing Keywords
- **`move`** - Force move semantics in closures
- **`ref`** - Create reference in pattern matching
- **`&`** - Reference operator (borrowing)
- **`&mut`** - Mutable reference
- **`*`** - Dereference operator
- **`Box`** - Heap allocation (not keyword, but fundamental)

#### Memory Safety Keywords
- **`unsafe`** - Unsafe code block
- **`extern`** - External function interface

#### Generics and Lifetimes
- **`where`** - Generic bounds clause
- **`'a`** - Lifetime parameter syntax

#### Error Handling
- **`Result`** - Result type (not keyword, but fundamental)
- **`Option`** - Option type (not keyword, but fundamental)

#### Async Keywords
- **`async`** - Asynchronous function
- **`await`** - Await async operation

#### Miscellaneous
- **`as`** - Type casting
- **`in`** - Used in for loops and pattern matching

### Reserved Keywords
These are reserved for future use and cannot be used as identifiers.

- **`abstract`** - Reserved for future use
- **`become`** - Reserved for future use
- **`box`** - Reserved for future use
- **`do`** - Reserved for future use
- **`final`** - Reserved for future use
- **`macro`** - Reserved for future use
- **`override`** - Reserved for future use
- **`priv`** - Reserved for future use
- **`typeof`** - Reserved for future use
- **`unsized`** - Reserved for future use
- **`virtual`** - Reserved for future use
- **`yield`** - Reserved for future use

### Weak Keywords
These can be used as identifiers in some contexts.

- **`union`** - Can be used as identifier except in declaration position
- **`'static`** - Special lifetime, can be shadowed in some contexts

## Syntax Elements

### Comments

#### Line Comments
```rust
// Single line comment
/// Documentation comment for following item
//! Inner documentation comment for enclosing item
```

#### Block Comments
```rust
/* Block comment */
/** Outer block doc comment */
/*! Inner block doc comment */
```

### Literals

#### Integer Literals
```rust
// Decimal
42
1_000_000  // Underscores for readability

// Hexadecimal
0xff
0xFF

// Octal
0o755

// Binary
0b1010_1010

// Type suffixes
42u32      // u32
42i64      // i64
42usize    // usize
42isize    // isize
```

#### Floating Point Literals
```rust
3.14
3.14f32    // f32 suffix
3.14f64    // f64 suffix
2.5E10     // Scientific notation
2.5e-4
```

#### Boolean Literals
```rust
true
false
```

#### Character Literals
```rust
'a'          // Character
'\\n'        // Newline escape
'\\t'        // Tab escape
'\\''        // Single quote escape
'\\\\'       // Backslash escape
'\\0'        // Null character
'\\u{1F600}' // Unicode escape (😀)
```

#### String Literals
```rust
\"Hello\"              // Basic string
\"Line 1\\nLine 2\"     // With escape sequences
\"\\u{1F600}\"          // Unicode in string
\"C:\\\\path\\\\file\"   // Windows path with escapes

// Raw strings
r\"No escapes: \\n \\t\"
r#\"String with "quotes"\"#
r##\"String with \"# inside\"##

// Byte strings
b\"Hello\"              // &[u8]
br\"Raw byte string\"

// Multi-line strings
\"This is a \\
multi-line string\"
```

### Identifiers

#### Valid Identifier Patterns
```rust
// Basic identifiers
variable_name
function_name
Type_Name
CONSTANT_NAME

// With numbers (not at start)
variable2
item_3

// Unicode identifiers (allowed)
变量
identifiér

// Raw identifiers (escape keywords)
r#type     // Use 'type' as identifier
r#match    // Use 'match' as identifier
```

### Operators

#### Arithmetic Operators
```rust
+    // Addition
-    // Subtraction
*    // Multiplication
/    // Division
%    // Remainder (modulo)
```

#### Comparison Operators
```rust
==   // Equal
!=   // Not equal
<    // Less than
<=   // Less than or equal
>    // Greater than
>=   // Greater than or equal
```

#### Logical Operators
```rust
&&   // Logical AND
||   // Logical OR
!    // Logical NOT
```

#### Bitwise Operators
```rust
&    // Bitwise AND
|    // Bitwise OR
^    // Bitwise XOR
!    // Bitwise NOT
<<   // Left shift
>>   // Right shift
```

#### Assignment Operators
```rust
=    // Assignment
+=   // Add and assign
-=   // Subtract and assign
*=   // Multiply and assign
/=   // Divide and assign
%=   // Remainder and assign
&=   // Bitwise AND and assign
|=   // Bitwise OR and assign
^=   // Bitwise XOR and assign
<<=  // Left shift and assign
>>=  // Right shift and assign
```

#### Other Operators
```rust
&    // Reference (borrow)
&mut // Mutable reference
*    // Dereference
..   // Range (exclusive)
..=  // Range (inclusive)
@    // Pattern binding
_    // Wildcard pattern
?    // Error propagation
```

### Punctuation

#### Delimiters
```rust
()   // Parentheses - function calls, tuples, grouping
[]   // Square brackets - arrays, indexing, attributes
{}   // Curly braces - blocks, structs, match arms

<>   // Angle brackets - generics
```

#### Separators
```rust
,    // Comma - separate items in lists
;    // Semicolon - statement terminator
:    // Colon - type annotations, struct field syntax
::   // Path separator
.    // Field access, method calls
..   // Range syntax, struct update syntax
...  // Variadic (deprecated, use ..)
```

#### Special Punctuation
```rust
->   // Function return type
=>   // Match arm separator
|    // Closure parameter separator, pattern alternatives
#    // Attribute prefix
$    // Macro variable prefix
```

### Attributes

#### Outer Attributes
```rust
#[derive(Debug, Clone)]
#[cfg(target_os = \"linux\")]
#[allow(dead_code)]
#[warn(unused_variables)]
#[deny(unsafe_code)]
#[forbid(unsafe_code)]
```

#### Inner Attributes
```rust
#![no_std]
#![no_main]
#![allow(unused)]
#![warn(missing_docs)]
```

### Macros

#### Macro Invocation
```rust
println!(\"Hello\");
vec![1, 2, 3];
format!(\"Number: {}\", 42);

// Different delimiters
macro_name!()
macro_name![]
macro_name!{}
```

#### Macro Definition
```rust
macro_rules! my_macro {
    () => {
        println!(\"Called with no arguments\");
    };
    ($x:expr) => {
        println!(\"Called with expression: {}\", $x);
    };
}
```

### Path Syntax

#### Absolute Paths
```rust
crate::module::function();
std::collections::HashMap::new();
::std::mem::size_of::<i32>();  // Global path
```

#### Relative Paths
```rust
super::parent_function();
self::current_module_item();
module::submodule::item();
```

#### Use Declarations
```rust
use std::collections::HashMap;
use std::io::{Read, Write};
use std::fs::File as FileHandle;
use std::collections::*;  // Glob import
```

### Pattern Syntax

#### Literal Patterns
```rust
match x {
    0 => \"zero\",
    1 | 2 => \"one or two\",
    3..=5 => \"three to five\",
    _ => \"other\",
}
```

#### Destructuring Patterns
```rust
// Tuple destructuring
let (x, y, z) = (1, 2, 3);

// Struct destructuring
struct Point { x: i32, y: i32 }
let Point { x, y } = point;
let Point { x: a, y: b } = point;  // Rename fields

// Enum destructuring
match option {
    Some(value) => value,
    None => 0,
}
```

#### Reference Patterns
```rust
match &value {
    &Some(ref x) => x,  // Match reference, bind by reference
    &None => &0,
}

// ref binding
let ref x = value;     // x is &T where value is T
let ref mut x = value; // x is &mut T
```

### Type Syntax

#### Primitive Types
```rust
bool        // Boolean
char        // Unicode character
i8, i16, i32, i64, i128, isize    // Signed integers
u8, u16, u32, u64, u128, usize    // Unsigned integers
f32, f64    // Floating point
str         // String slice (unsized)
()          // Unit type
!           // Never type
```

#### Compound Types
```rust
[T; N]           // Array type
[T]              // Slice type (unsized)
(T, U, V)        // Tuple type
&T               // Reference type
&mut T           // Mutable reference type
*const T         // Raw const pointer
*mut T           // Raw mutable pointer
Box<T>           // Owned pointer
```

#### Function Types
```rust
fn(i32) -> i32                    // Function pointer
fn(i32, i32) -> i32              // Multiple parameters
fn()                             // No parameters, no return
unsafe fn(i32) -> i32           // Unsafe function pointer
extern \"C\" fn(i32) -> i32       // C ABI function pointer
```

#### Closure Types
```rust
Fn(i32) -> i32      // Immutable closure trait
FnMut(i32) -> i32   // Mutable closure trait
FnOnce(i32) -> i32  // Move closure trait
```

#### Generic Types
```rust
Option<T>           // Generic enum
Result<T, E>        // Generic enum with two parameters
Vec<T>              // Generic struct
HashMap<K, V>       // Multiple type parameters
```

### Lifetime Syntax

#### Lifetime Parameters
```rust
'a, 'b, 'static     // Lifetime names
'_                  // Anonymous lifetime

// Function with lifetimes
fn longest<'a>(x: &'a str, y: &'a str) -> &'a str {
    if x.len() > y.len() { x } else { y }
}

// Struct with lifetimes
struct ImportantExcerpt<'a> {
    part: &'a str,
}

// Multiple lifetimes
fn complex<'a, 'b>(x: &'a str, y: &'b str) -> &'a str {
    x
}
```

### Expression Syntax

#### Block Expressions
```rust
{
    let x = 1;
    let y = 2;
    x + y  // Last expression is returned
}

// Unsafe blocks
unsafe {
    *raw_pointer
}
```

#### Control Flow Expressions
```rust
// If expressions
let result = if condition { 1 } else { 2 };

// Match expressions
let result = match value {
    Pattern1 => expression1,
    Pattern2 if guard => expression2,
    _ => default_expression,
};

// Loop expressions with labels
'outer: loop {
    'inner: loop {
        break 'outer;  // Break outer loop
    }
}
```

#### Range Expressions
```rust
1..10        // 1 to 9 (exclusive)
1..=10       // 1 to 10 (inclusive)
..10         // 0 to 9
1..          // 1 to end
..           // Full range
```

### Statement Syntax

#### Expression Statements
```rust
function_call();     // Expression statement (with semicolon)
x + y;              // Expression statement (return value ignored)
```

#### Declaration Statements
```rust
let x = 5;                    // Immutable binding
let mut y = 10;               // Mutable binding
let z: i32 = 15;             // With type annotation
let (a, b) = (1, 2);         // Destructuring binding
```

### Item Syntax

#### Function Items
```rust
fn function_name() {}
fn function_name() -> ReturnType {}
fn function_name(param: Type) {}
fn function_name<T>(param: T) where T: Trait {}

// Associated functions
impl SomeType {
    fn associated_function() {}
    fn method(&self) {}
    fn mutating_method(&mut self) {}
    fn consuming_method(self) {}
}
```

#### Type Items
```rust
// Struct definitions
struct UnitStruct;
struct TupleStruct(i32, String);
struct RecordStruct {
    field1: i32,
    field2: String,
}

// Enum definitions
enum SimpleEnum {
    Variant1,
    Variant2,
}

enum ComplexEnum {
    Unit,
    Tuple(i32, String),
    Struct { field1: i32, field2: String },
}

// Union definitions (unsafe)
union MyUnion {
    f1: u32,
    f2: f32,
}
```

#### Trait Items
```rust
trait MyTrait {
    type AssociatedType;                    // Associated type
    const CONSTANT: i32 = 10;              // Associated constant

    fn required_method(&self);              // Required method
    fn default_method(&self) {              // Default implementation
        // Default behavior
    }
}
```

### Special Syntax

#### Attribute Syntax
```rust
#[attribute]                    // Outer attribute
#[attribute(parameter)]         // Attribute with parameter
#[attribute = \"value\"]          // Attribute with value
#[cfg(feature = \"feature_name\")] // Conditional compilation
```

#### Macro Syntax
```rust
// Macro invocation patterns
macro_name!();                  // Empty
macro_name!(arg1, arg2);        // Arguments
macro_name! { key: value };     // Block form
macro_name![item1, item2];      // Array-like form
```

This comprehensive reference covers all major syntactic elements of the Rust programming language, providing a complete foundation for understanding and writing Rust code.