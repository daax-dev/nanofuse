# C Keywords and Syntax Reference

## C23 Keywords (Latest Standard)

### Core Keywords (32 keywords from C89/C90)
```c
auto        break       case        char        const       continue    default     do
double      else        enum        extern      float       for         goto        if
int         long        register    return      short       signed      sizeof      static
struct      switch      typedef     union       unsigned    void        volatile    while
```

### C99 Additions (5 keywords)
```c
inline      restrict    _Bool       _Complex    _Imaginary
```

### C11 Additions (7 keywords)
```c
_Alignas            _Alignof            _Atomic             _Static_assert
_Noreturn           _Thread_local       _Generic
```

### C23 Additions (8 keywords)
```c
typeof              typeof_unqual       _BitInt
_Decimal128         _Decimal32          _Decimal64
_Alignof            _Alignas            // Enhanced with new semantics
```

## Reserved Identifiers

### Standard Library Reserved Patterns
- **_[A-Z]**: Identifiers starting with underscore followed by uppercase letter
- **__**: Identifiers starting with double underscore
- **_[a-z]**: Identifiers starting with underscore in file scope

### Examples of Reserved Identifiers
```c
// Reserved - don't use these patterns
_MAX_SIZE          // Underscore + uppercase
__builtin_*        // Double underscore
_exit             // File scope underscore + lowercase
```

## Data Types and Type Specifiers

### Basic Types
```c
// Character types
char                    // At least 8 bits
signed char            // Explicitly signed 8-bit
unsigned char          // Unsigned 8-bit

// Integer types
short                  // At least 16 bits
int                    // Natural word size, at least 16 bits
long                   // At least 32 bits
long long              // At least 64 bits (C99+)

// Floating-point types
float                  // Single precision
double                 // Double precision
long double            // Extended precision

// C99 exact-width types (stdint.h)
int8_t, uint8_t
int16_t, uint16_t
int32_t, uint32_t
int64_t, uint64_t

// C23 bit-precise integers
_BitInt(N)             // Signed N-bit integer
unsigned _BitInt(N)    // Unsigned N-bit integer
```

### Type Qualifiers
```c
const           // Read-only
volatile        // Prevents optimization
restrict        // Pointer aliasing hint (C99+)
_Atomic         // Atomic operations (C11+)
```

### Storage Classes
```c
auto            // Automatic storage (default for local variables)
register        // Hint for register storage
static          // Static storage, file/function scope
extern          // External linkage
_Thread_local   // Thread-local storage (C11+)
```

## Operators

### Arithmetic Operators
```c
+       // Addition
-       // Subtraction
*       // Multiplication
/       // Division
%       // Modulo (remainder)

++      // Increment (prefix/postfix)
--      // Decrement (prefix/postfix)

+       // Unary plus
-       // Unary minus
```

### Relational and Logical Operators
```c
==      // Equal to
!=      // Not equal to
<       // Less than
<=      // Less than or equal to
>       // Greater than
>=      // Greater than or equal to

&&      // Logical AND
||      // Logical OR
!       // Logical NOT
```

### Bitwise Operators
```c
&       // Bitwise AND
|       // Bitwise OR
^       // Bitwise XOR
~       // Bitwise NOT (complement)
<<      // Left shift
>>      // Right shift
```

### Assignment Operators
```c
=       // Simple assignment

+=      // Addition assignment
-=      // Subtraction assignment
*=      // Multiplication assignment
/=      // Division assignment
%=      // Modulo assignment

&=      // Bitwise AND assignment
|=      // Bitwise OR assignment
^=      // Bitwise XOR assignment
<<=     // Left shift assignment
>>=     // Right shift assignment
```

### Other Operators
```c
&       // Address-of
*       // Dereference
->      // Member access through pointer
.       // Member access
[]      // Array subscript
()      // Function call
,       // Comma operator
?:      // Ternary conditional
sizeof  // Size operator
_Alignof // Alignment operator (C11+)
typeof   // Type operator (C23)
```

## Operator Precedence and Associativity

| Precedence | Operator | Description | Associativity |
|------------|----------|-------------|---------------|
| 1 (highest) | `() [] -> .` | Function call, array subscript, member access | Left-to-right |
| 2 | `! ~ ++ -- + - * & (type) sizeof _Alignof` | Unary operators | Right-to-left |
| 3 | `* / %` | Multiplicative | Left-to-right |
| 4 | `+ -` | Additive | Left-to-right |
| 5 | `<< >>` | Shift | Left-to-right |
| 6 | `< <= > >=` | Relational | Left-to-right |
| 7 | `== !=` | Equality | Left-to-right |
| 8 | `&` | Bitwise AND | Left-to-right |
| 9 | `^` | Bitwise XOR | Left-to-right |
| 10 | `|` | Bitwise OR | Left-to-right |
| 11 | `&&` | Logical AND | Left-to-right |
| 12 | `||` | Logical OR | Left-to-right |
| 13 | `?:` | Conditional | Right-to-left |
| 14 | `= += -= *= /= %= &= |= ^= <<= >>=` | Assignment | Right-to-left |
| 15 (lowest) | `,` | Comma | Left-to-right |

## Control Flow Statements

### Conditional Statements
```c
// if statement
if (condition) {
    // statements
}

// if-else statement
if (condition) {
    // statements
} else {
    // statements
}

// if-else if-else chain
if (condition1) {
    // statements
} else if (condition2) {
    // statements
} else {
    // statements
}

// switch statement
switch (expression) {
    case constant1:
        // statements
        break;
    case constant2:
        // statements
        break;
    default:
        // statements
        break;
}
```

### Loop Statements
```c
// while loop
while (condition) {
    // statements
}

// do-while loop
do {
    // statements
} while (condition);

// for loop
for (initialization; condition; increment) {
    // statements
}

// C99 for loop with declaration
for (int i = 0; i < n; i++) {
    // statements
}
```

### Jump Statements
```c
break;          // Exit loop or switch
continue;       // Skip to next iteration
return;         // Return from function
return value;   // Return value from function
goto label;     // Jump to labeled statement

label:          // Label for goto
```

## Declarations and Definitions

### Variable Declarations
```c
// Basic declaration
int x;

// Initialization
int x = 42;

// Multiple variables
int x, y, z;
int x = 1, y = 2, z = 3;

// C99: Mixed declarations and code
int x = func();
some_statement();
int y = x * 2;  // Declaration after statement

// C99: Designated initializers
int arr[5] = {[2] = 3, [4] = 5};
struct point p = {.x = 1, .y = 2};
```

### Array Declarations
```c
// Fixed-size array
int arr[10];

// Array with initialization
int arr[5] = {1, 2, 3, 4, 5};
int arr[] = {1, 2, 3, 4, 5};  // Size inferred

// Multidimensional arrays
int matrix[3][4];
int matrix[3][4] = {{1, 2, 3, 4}, {5, 6, 7, 8}, {9, 10, 11, 12}};

// C99: Variable-length arrays (VLA)
int n = 10;
int vla[n];

// C99: VLA with function parameters
void func(int n, int arr[n]);
```

### Pointer Declarations
```c
// Basic pointer
int *ptr;

// Pointer initialization
int x = 42;
int *ptr = &x;

// Pointer to pointer
int **ptr_to_ptr;

// Array of pointers
int *ptr_array[10];

// Pointer to array
int (*array_ptr)[10];

// Function pointer
int (*func_ptr)(int, int);
```

### Structure and Union Declarations
```c
// Structure definition
struct point {
    int x;
    int y;
};

// Structure variable declaration
struct point p1;
struct point p2 = {10, 20};

// C99: Designated initializers
struct point p3 = {.x = 10, .y = 20};

// Anonymous structures (C11)
struct {
    int x, y;
} anonymous_point;

// Union definition
union data {
    int i;
    float f;
    char c;
};

// Bit fields
struct flags {
    unsigned int is_valid : 1;
    unsigned int is_ready : 1;
    unsigned int reserved : 6;
};
```

### Function Declarations
```c
// Function declaration (prototype)
int add(int a, int b);

// Function definition
int add(int a, int b) {
    return a + b;
}

// C99: inline functions
inline int square(int x) {
    return x * x;
}

// C11: _Noreturn functions
_Noreturn void exit_program(void);

// Variadic functions
int printf(const char *format, ...);
```

### Typedef Declarations
```c
// Basic typedef
typedef int Integer;
typedef unsigned long size_t;

// Struct typedef
typedef struct {
    int x, y;
} Point;

// Function pointer typedef
typedef int (*BinaryOp)(int, int);

// Array typedef
typedef int Vector[3];
```

### Enumeration Declarations
```c
// Basic enumeration
enum color {
    RED,
    GREEN,
    BLUE
};

// Enumeration with explicit values
enum status {
    SUCCESS = 0,
    ERROR = -1,
    PENDING = 1
};

// C23: Enumeration with underlying type
enum : unsigned char {
    FLAG_A = 1,
    FLAG_B = 2,
    FLAG_C = 4
};
```

## Preprocessor Directives

### File Inclusion
```c
#include <stdio.h>      // System header
#include "myheader.h"   // User header
```

### Macro Definitions
```c
// Object-like macros
#define MAX_SIZE 100
#define PI 3.14159

// Function-like macros
#define SQUARE(x) ((x) * (x))
#define MAX(a, b) ((a) > (b) ? (a) : (b))

// Variadic macros (C99)
#define DEBUG_PRINT(fmt, ...) printf("DEBUG: " fmt, __VA_ARGS__)

// C23: __VA_OPT__
#define DEBUG(fmt, ...) printf("DEBUG: " fmt __VA_OPT__(,) __VA_ARGS__)
```

### Conditional Compilation
```c
#ifdef SYMBOL
    // Code if SYMBOL is defined
#endif

#ifndef SYMBOL
    // Code if SYMBOL is not defined
#endif

#if defined(SYMBOL)
    // Code if SYMBOL is defined
#endif

#if CONDITION
    // Code if CONDITION is true
#elif OTHER_CONDITION
    // Code if OTHER_CONDITION is true
#else
    // Code if no conditions are true
#endif
```

### Other Preprocessor Directives
```c
#undef SYMBOL          // Undefine macro
#line 100 "file.c"     // Set line number and filename
#error "Error message" // Generate compilation error
#warning "Warning"     // Generate compilation warning (GCC extension)
#pragma directive      // Implementation-specific directive
```

### Predefined Macros
```c
__FILE__        // Current source file name
__LINE__        // Current line number
__DATE__        // Compilation date
__TIME__        // Compilation time
__STDC__        // Standard C compliance
__STDC_VERSION__ // C standard version

// C99 predefined macros
__func__        // Current function name

// C11 predefined macros
_Static_assert  // Compile-time assertion

// C23 predefined macros
typeof          // Type operator
typeof_unqual   // Unqualified type operator
```

## Comments

### C-style Comments
```c
/* Single line comment */

/*
 * Multi-line comment
 * spanning several lines
 */

/*
 * Documentation-style comment
 * @param x: The input value
 * @return: The processed value
 */
```

### C99 Single-line Comments
```c
// Single line comment
int x = 42; // End-of-line comment

// TODO: Implement this function
// FIXME: Memory leak in this function
// NOTE: This is a performance-critical section
```

## Literals

### Integer Literals
```c
// Decimal
42
42L         // long
42LL        // long long
42U         // unsigned
42UL        // unsigned long

// Octal (starts with 0)
042         // Octal 42 = decimal 34

// Hexadecimal (starts with 0x or 0X)
0x42        // Hex 42 = decimal 66
0X42

// C23: Binary literals
0b101010    // Binary = decimal 42
0B101010
```

### Floating-point Literals
```c
// Decimal notation
3.14
3.14f       // float
3.14L       // long double

// Scientific notation
1.23e4      // 1.23 × 10^4
1.23E-4     // 1.23 × 10^-4

// C99: Hexadecimal floating-point
0x1.23p4    // 1.23 × 2^4
```

### Character Literals
```c
// Basic character literals
'a'
'Z'
'0'

// Escape sequences
'\n'        // Newline
'\t'        // Tab
'\r'        // Carriage return
'\\'        // Backslash
'\''        // Single quote
'\"'        // Double quote
'\0'        // Null character

// Octal escape sequences
'\141'      // Octal for 'a'

// Hexadecimal escape sequences
'\x61'      // Hex for 'a'

// Unicode escape sequences (C99)
'\u0061'    // Unicode for 'a'
'\U00000061' // Extended Unicode for 'a'
```

### String Literals
```c
// Basic string literals
"Hello, World!"
"This is a string"

// String concatenation
"Hello, " "World!"  // Automatically concatenated

// Escape sequences in strings
"Line 1\nLine 2"
"Path: C:\\Program Files\\App"

// Raw strings (not in C, but similar effect with macros)
// Use R"(...)" in C++ or define custom macros

// Wide character strings (C99)
L"Wide string"

// C11: UTF-8, UTF-16, UTF-32 string literals
u8"UTF-8 string"
u"UTF-16 string"
U"UTF-32 string"
```

## C23 New Features Syntax

### typeof Operator
```c
// Basic typeof usage
int x = 42;
typeof(x) y = x;  // y has same type as x

// With expressions
typeof(func()) result;

// typeof_unqual for unqualified types
const int x = 42;
typeof_unqual(x) y = 0;  // y is int, not const int
```

### Enhanced Attributes
```c
// C23 attribute syntax
[[deprecated]] void old_function(void);
[[maybe_unused]] int debug_variable;
[[nodiscard]] int important_function(void);

// Multiple attributes
[[deprecated, maybe_unused]] int old_variable;
```

### Binary Literals
```c
// C23 binary integer literals
int flags = 0b11010110;
unsigned mask = 0b1111U;
```

### Enhanced Generic Selection
```c
// C11 _Generic enhanced in C23
#define abs(x) _Generic((x), \
    int: abs, \
    long: labs, \
    long long: llabs, \
    float: fabsf, \
    double: fabs, \
    long double: fabsl \
)(x)
```

This comprehensive syntax reference covers all major aspects of C programming language syntax from C89 through C23, providing both basic and advanced language constructs.