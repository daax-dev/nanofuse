# C Types and Collections System

## Fundamental Data Types

### Character Types
```c
#include <limits.h>
#include <stdio.h>

void demonstrate_char_types(void) {
    char c1 = 'A';                    // Implementation-defined signedness
    signed char c2 = -128;            // Explicitly signed (-128 to 127)
    unsigned char c3 = 255;           // Unsigned (0 to 255)

    printf("char: %d, signed char: %d, unsigned char: %u\n", c1, c2, c3);
    printf("CHAR_BIT: %d bits per char\n", CHAR_BIT);
    printf("CHAR_MIN: %d, CHAR_MAX: %d\n", CHAR_MIN, CHAR_MAX);
}
```

### Integer Types
```c
#include <stdint.h>
#include <inttypes.h>

void demonstrate_integer_types(void) {
    // Basic integer types
    short s = 32767;                  // At least 16 bits
    int i = 2147483647;              // At least 16 bits (usually 32)
    long l = 2147483647L;            // At least 32 bits
    long long ll = 9223372036854775807LL; // At least 64 bits

    // Unsigned variants
    unsigned short us = 65535U;
    unsigned int ui = 4294967295U;
    unsigned long ul = 4294967295UL;
    unsigned long long ull = 18446744073709551615ULL;

    // C99 exact-width types (stdint.h)
    int8_t i8 = 127;                 // Exactly 8 bits
    int16_t i16 = 32767;             // Exactly 16 bits
    int32_t i32 = 2147483647;        // Exactly 32 bits
    int64_t i64 = 9223372036854775807LL; // Exactly 64 bits

    uint8_t u8 = 255;
    uint16_t u16 = 65535;
    uint32_t u32 = 4294967295U;
    uint64_t u64 = 18446744073709551615ULL;

    // Fast and least-width types
    int_fast8_t if8 = 100;           // Fastest type at least 8 bits
    int_least8_t il8 = 100;          // Smallest type at least 8 bits

    // Pointer-sized integers
    intptr_t iptr = (intptr_t)&i;    // Can hold pointer value
    uintptr_t uptr = (uintptr_t)&i;  // Unsigned pointer-sized

    // Maximum width integers
    intmax_t imax = INTMAX_C(9223372036854775807);
    uintmax_t umax = UINTMAX_C(18446744073709551615);

    printf("Sizes: short=%zu, int=%zu, long=%zu, long long=%zu\n",
           sizeof(short), sizeof(int), sizeof(long), sizeof(long long));
}
```

### Floating-Point Types
```c
#include <float.h>
#include <math.h>

void demonstrate_float_types(void) {
    float f = 3.14159f;              // Single precision (usually 32-bit)
    double d = 3.141592653589793;    // Double precision (usually 64-bit)
    long double ld = 3.141592653589793238L; // Extended precision

    printf("float: %f (%.6g)\n", f, f);
    printf("double: %lf (%.15g)\n", d, d);
    printf("long double: %Lf (%.18Lg)\n", ld, ld);

    // Floating-point limits
    printf("FLT_DIG: %d decimal digits precision\n", FLT_DIG);
    printf("DBL_DIG: %d decimal digits precision\n", DBL_DIG);
    printf("LDBL_DIG: %d decimal digits precision\n", LDBL_DIG);

    // Special values
    float pos_inf = INFINITY;
    float neg_inf = -INFINITY;
    float not_a_number = NAN;

    printf("INFINITY: %f, -INFINITY: %f, NAN: %f\n",
           pos_inf, neg_inf, not_a_number);
}
```

### C23 Bit-Precise Integers
```c
// C23 feature - may not be available in all compilers yet
#if __STDC_VERSION__ >= 202311L
void demonstrate_bitint(void) {
    _BitInt(7) small_int = 63;       // 7-bit signed integer
    unsigned _BitInt(7) small_uint = 127; // 7-bit unsigned integer

    _BitInt(128) huge_int = 0;       // 128-bit integer

    printf("7-bit signed: %d, unsigned: %u\n", small_int, small_uint);
}
#endif
```

## Derived Types

### Arrays
```c
void demonstrate_arrays(void) {
    // Fixed-size arrays
    int numbers[5] = {1, 2, 3, 4, 5};
    char string[20] = "Hello, World!";

    // Array initialization patterns
    int zeros[10] = {0};             // All elements zero
    int partial[10] = {1, 2, 3};     // 1,2,3,0,0,0,0,0,0,0

    // C99 designated initializers
    int designated[10] = {[0] = 1, [4] = 5, [9] = 10};

    // Multidimensional arrays
    int matrix[3][4] = {
        {1, 2, 3, 4},
        {5, 6, 7, 8},
        {9, 10, 11, 12}
    };

    // Array size calculation
    size_t array_length = sizeof(numbers) / sizeof(numbers[0]);
    printf("Array length: %zu\n", array_length);

    // Arrays decay to pointers when passed to functions
    print_array(numbers, array_length);
}

void print_array(int arr[], size_t length) {  // arr[] is same as int* arr
    for (size_t i = 0; i < length; i++) {
        printf("%d ", arr[i]);
    }
    printf("\n");
}
```

### Variable Length Arrays (C99)
```c
#include <stdlib.h>

void demonstrate_vla(int n) {
    // VLA on stack
    int vla[n];

    // Initialize VLA
    for (int i = 0; i < n; i++) {
        vla[i] = i * i;
    }

    // 2D VLA
    int matrix[n][n];
    for (int i = 0; i < n; i++) {
        for (int j = 0; j < n; j++) {
            matrix[i][j] = i * n + j;
        }
    }

    printf("VLA of size %d created\n", n);
}

// VLA in function parameters
void process_matrix(int rows, int cols, int matrix[rows][cols]) {
    for (int i = 0; i < rows; i++) {
        for (int j = 0; j < cols; j++) {
            matrix[i][j] *= 2;
        }
    }
}
```

### Pointers
```c
void demonstrate_pointers(void) {
    int x = 42;
    int* ptr = &x;                   // Pointer to int
    int** ptr_to_ptr = &ptr;         // Pointer to pointer to int

    printf("x = %d, *ptr = %d, **ptr_to_ptr = %d\n", x, *ptr, **ptr_to_ptr);

    // Pointer arithmetic
    int array[5] = {10, 20, 30, 40, 50};
    int* p = array;                  // Points to first element

    printf("Elements via pointer arithmetic:\n");
    for (int i = 0; i < 5; i++) {
        printf("*(p + %d) = %d\n", i, *(p + i));
    }

    // Function pointers
    int (*operation)(int, int);
    operation = add;
    printf("Function pointer result: %d\n", operation(10, 20));

    // Array of function pointers
    int (*operations[])(int, int) = {add, subtract, multiply};
    for (int i = 0; i < 3; i++) {
        printf("Operation %d: %d\n", i, operations[i](15, 5));
    }
}

int add(int a, int b) { return a + b; }
int subtract(int a, int b) { return a - b; }
int multiply(int a, int b) { return a * b; }
```

### Structures
```c
// Basic structure
struct point {
    int x;
    int y;
};

// Structure with different types
struct person {
    char name[50];
    int age;
    float height;
    struct point location;  // Nested structure
};

// Self-referential structure (linked list)
struct node {
    int data;
    struct node* next;
};

// Anonymous structures (C11)
struct {
    int anonymous_field;
} anonymous_struct;

void demonstrate_structures(void) {
    // Structure initialization
    struct point p1 = {10, 20};
    struct point p2 = {.x = 5, .y = 15}; // C99 designated initializers

    struct person john = {
        .name = "John Doe",
        .age = 30,
        .height = 5.9f,
        .location = {100, 200}
    };

    printf("Person: %s, Age: %d, Location: (%d, %d)\n",
           john.name, john.age, john.location.x, john.location.y);

    // Structure assignment
    struct point p3 = p1;  // Copies all fields

    // Pointer to structure
    struct person* person_ptr = &john;
    printf("Via pointer: %s, Age: %d\n", person_ptr->name, person_ptr->age);
}
```

### Bit Fields
```c
struct flags {
    unsigned int is_valid : 1;       // 1 bit
    unsigned int is_ready : 1;       // 1 bit
    unsigned int priority : 3;       // 3 bits (0-7)
    unsigned int reserved : 3;       // 3 bits reserved
};

struct packed_date {
    unsigned int day : 5;    // 1-31
    unsigned int month : 4;  // 1-12
    unsigned int year : 23;  // Large enough for years
};

void demonstrate_bit_fields(void) {
    struct flags f = {0};
    f.is_valid = 1;
    f.is_ready = 1;
    f.priority = 5;

    printf("Flags: valid=%u, ready=%u, priority=%u\n",
           f.is_valid, f.is_ready, f.priority);
    printf("Size of flags struct: %zu bytes\n", sizeof(struct flags));

    struct packed_date date = {.day = 15, .month = 8, .year = 2023};
    printf("Date: %u/%u/%u\n", date.day, date.month, date.year);
    printf("Size of date struct: %zu bytes\n", sizeof(struct packed_date));
}
```

### Unions
```c
union data {
    int i;
    float f;
    char c[4];
};

// Tagged union for type safety
struct tagged_union {
    enum { TYPE_INT, TYPE_FLOAT, TYPE_STRING } type;
    union {
        int i;
        float f;
        char* s;
    } value;
};

void demonstrate_unions(void) {
    union data d;

    d.i = 42;
    printf("As int: %d\n", d.i);

    d.f = 3.14f;
    printf("As float: %f\n", d.f);
    printf("As int (reinterpreted): %d\n", d.i); // Undefined behavior

    // Safe tagged union usage
    struct tagged_union tu;
    tu.type = TYPE_INT;
    tu.value.i = 100;

    switch (tu.type) {
        case TYPE_INT:
            printf("Tagged union int: %d\n", tu.value.i);
            break;
        case TYPE_FLOAT:
            printf("Tagged union float: %f\n", tu.value.f);
            break;
        case TYPE_STRING:
            printf("Tagged union string: %s\n", tu.value.s);
            break;
    }
}
```

### Enumerations
```c
// Basic enumeration
enum color {
    RED,     // 0
    GREEN,   // 1
    BLUE     // 2
};

// Enumeration with explicit values
enum status_code {
    SUCCESS = 0,
    ERROR_INVALID_INPUT = -1,
    ERROR_OUT_OF_MEMORY = -2,
    ERROR_FILE_NOT_FOUND = -3
};

// C23 typed enumerations (if supported)
#if __STDC_VERSION__ >= 202311L
enum : unsigned char {
    FLAG_A = 1,
    FLAG_B = 2,
    FLAG_C = 4,
    FLAG_D = 8
};
#endif

void demonstrate_enums(void) {
    enum color favorite = BLUE;
    enum status_code result = SUCCESS;

    printf("Favorite color: %d\n", favorite);
    printf("Status: %d\n", result);

    // Enums can be used in switch statements
    switch (favorite) {
        case RED:
            printf("Red is chosen\n");
            break;
        case GREEN:
            printf("Green is chosen\n");
            break;
        case BLUE:
            printf("Blue is chosen\n");
            break;
    }
}
```

## Type Qualifiers

### const Qualifier
```c
void demonstrate_const(void) {
    const int constant_value = 42;
    // constant_value = 43; // Error: cannot modify const

    const int* ptr_to_const = &constant_value;    // Pointer to const int
    // *ptr_to_const = 43; // Error: cannot modify through pointer

    int value = 10;
    int* const const_ptr = &value;               // const pointer to int
    *const_ptr = 20;  // OK: can modify value
    // const_ptr = &constant_value; // Error: cannot modify pointer

    const int* const const_ptr_to_const = &constant_value; // const pointer to const int
    // *const_ptr_to_const = 30; // Error: cannot modify value
    // const_ptr_to_const = &value; // Error: cannot modify pointer
}
```

### volatile Qualifier
```c
volatile int hardware_register;
volatile sig_atomic_t signal_flag;

void demonstrate_volatile(void) {
    volatile int sensor_reading = 0;

    // Compiler cannot optimize away repeated reads
    int first_read = sensor_reading;
    int second_read = sensor_reading;  // This read cannot be eliminated

    if (first_read != second_read) {
        printf("Sensor value changed: %d -> %d\n", first_read, second_read);
    }
}
```

### restrict Qualifier (C99)
```c
// restrict tells compiler that pointer is the only way to access the object
void copy_array(int* restrict dest, const int* restrict src, size_t n) {
    for (size_t i = 0; i < n; i++) {
        dest[i] = src[i];  // Compiler can optimize knowing no aliasing
    }
}

void demonstrate_restrict(void) {
    int source[10] = {1, 2, 3, 4, 5, 6, 7, 8, 9, 10};
    int destination[10];

    copy_array(destination, source, 10);

    for (int i = 0; i < 10; i++) {
        printf("%d ", destination[i]);
    }
    printf("\n");
}
```

## Built-in Collections and Data Structures

### Strings
```c
#include <string.h>

void demonstrate_strings(void) {
    // String literals
    char* literal = "Hello, World!";  // In read-only memory
    char array[] = "Hello, World!";   // Mutable copy
    char buffer[100] = "Initial";     // Fixed-size buffer

    // String operations
    size_t len = strlen(literal);
    printf("String length: %zu\n", len);

    // Safe string operations
    strncpy(buffer, literal, sizeof(buffer) - 1);
    buffer[sizeof(buffer) - 1] = '\0';  // Ensure null termination

    strncat(buffer, " Appended", sizeof(buffer) - strlen(buffer) - 1);

    printf("Modified string: %s\n", buffer);

    // String comparison
    if (strcmp(literal, array) == 0) {
        printf("Strings are equal\n");
    }
}
```

### Dynamic Arrays
```c
typedef struct {
    int* data;
    size_t size;
    size_t capacity;
} dynamic_array_t;

dynamic_array_t* array_create(size_t initial_capacity) {
    dynamic_array_t* arr = malloc(sizeof(dynamic_array_t));
    if (!arr) return NULL;

    arr->data = malloc(initial_capacity * sizeof(int));
    if (!arr->data) {
        free(arr);
        return NULL;
    }

    arr->size = 0;
    arr->capacity = initial_capacity;
    return arr;
}

int array_push_back(dynamic_array_t* arr, int value) {
    if (!arr) return -1;

    if (arr->size >= arr->capacity) {
        size_t new_capacity = arr->capacity * 2;
        int* new_data = realloc(arr->data, new_capacity * sizeof(int));
        if (!new_data) return -1;

        arr->data = new_data;
        arr->capacity = new_capacity;
    }

    arr->data[arr->size++] = value;
    return 0;
}

int array_get(dynamic_array_t* arr, size_t index) {
    if (!arr || index >= arr->size) return -1;
    return arr->data[index];
}

void array_destroy(dynamic_array_t* arr) {
    if (arr) {
        free(arr->data);
        free(arr);
    }
}
```

### Linked Lists
```c
typedef struct list_node {
    int data;
    struct list_node* next;
} list_node_t;

typedef struct {
    list_node_t* head;
    size_t size;
} linked_list_t;

linked_list_t* list_create(void) {
    linked_list_t* list = malloc(sizeof(linked_list_t));
    if (!list) return NULL;

    list->head = NULL;
    list->size = 0;
    return list;
}

int list_push_front(linked_list_t* list, int data) {
    if (!list) return -1;

    list_node_t* new_node = malloc(sizeof(list_node_t));
    if (!new_node) return -1;

    new_node->data = data;
    new_node->next = list->head;
    list->head = new_node;
    list->size++;
    return 0;
}

int list_pop_front(linked_list_t* list) {
    if (!list || !list->head) return -1;

    list_node_t* old_head = list->head;
    int data = old_head->data;

    list->head = old_head->next;
    free(old_head);
    list->size--;

    return data;
}

void list_destroy(linked_list_t* list) {
    if (!list) return;

    list_node_t* current = list->head;
    while (current) {
        list_node_t* next = current->next;
        free(current);
        current = next;
    }

    free(list);
}
```

### Hash Tables
```c
#define HASH_TABLE_SIZE 101

typedef struct hash_entry {
    char* key;
    int value;
    struct hash_entry* next;
} hash_entry_t;

typedef struct {
    hash_entry_t* buckets[HASH_TABLE_SIZE];
} hash_table_t;

unsigned int hash_function(const char* key) {
    unsigned int hash = 5381;
    int c;

    while ((c = *key++)) {
        hash = ((hash << 5) + hash) + c; // hash * 33 + c
    }

    return hash % HASH_TABLE_SIZE;
}

hash_table_t* hash_table_create(void) {
    hash_table_t* table = malloc(sizeof(hash_table_t));
    if (!table) return NULL;

    for (int i = 0; i < HASH_TABLE_SIZE; i++) {
        table->buckets[i] = NULL;
    }

    return table;
}

int hash_table_insert(hash_table_t* table, const char* key, int value) {
    if (!table || !key) return -1;

    unsigned int index = hash_function(key);

    // Check if key already exists
    hash_entry_t* entry = table->buckets[index];
    while (entry) {
        if (strcmp(entry->key, key) == 0) {
            entry->value = value;  // Update existing
            return 0;
        }
        entry = entry->next;
    }

    // Create new entry
    hash_entry_t* new_entry = malloc(sizeof(hash_entry_t));
    if (!new_entry) return -1;

    new_entry->key = malloc(strlen(key) + 1);
    if (!new_entry->key) {
        free(new_entry);
        return -1;
    }

    strcpy(new_entry->key, key);
    new_entry->value = value;
    new_entry->next = table->buckets[index];
    table->buckets[index] = new_entry;

    return 0;
}

int hash_table_get(hash_table_t* table, const char* key, int* value) {
    if (!table || !key || !value) return -1;

    unsigned int index = hash_function(key);
    hash_entry_t* entry = table->buckets[index];

    while (entry) {
        if (strcmp(entry->key, key) == 0) {
            *value = entry->value;
            return 0;  // Found
        }
        entry = entry->next;
    }

    return -1;  // Not found
}
```

## Type Conversions

### Implicit Conversions
```c
void demonstrate_implicit_conversions(void) {
    int i = 42;
    float f = i;        // int to float
    double d = f;       // float to double

    char c = 'A';
    int ascii = c;      // char to int

    // Integer promotion in expressions
    short s1 = 10, s2 = 20;
    int result = s1 + s2;  // shorts promoted to int

    printf("Conversions: %f, %lf, %d\n", f, d, ascii);
}
```

### Explicit Conversions (Casts)
```c
void demonstrate_explicit_conversions(void) {
    double d = 3.14159;
    int i = (int)d;                    // Explicit cast, truncates

    void* generic_ptr = &i;
    int* specific_ptr = (int*)generic_ptr;  // Cast void* to specific type

    // Function pointer cast
    void* func_ptr = (void*)demonstrate_explicit_conversions;
    void (*typed_func_ptr)(void) = (void(*)(void))func_ptr;

    printf("Cast result: %d\n", i);
}
```

### Compound Literals (C99)
```c
void demonstrate_compound_literals(void) {
    // Compound literal for array
    int* array_ptr = (int[]){1, 2, 3, 4, 5};

    // Compound literal for structure
    struct point* p = &(struct point){.x = 10, .y = 20};

    // Passing compound literals to functions
    print_point((struct point){30, 40});

    printf("Compound literal point: (%d, %d)\n", p->x, p->y);
}

void print_point(struct point p) {
    printf("Point: (%d, %d)\n", p.x, p.y);
}
```

## Type Alignment and Packing

### Alignment
```c
#include <stdalign.h>
#include <stddef.h>

struct aligned_struct {
    char c;                    // 1 byte
    // 3 bytes padding
    int i;                     // 4 bytes, aligned to 4-byte boundary
    char c2;                   // 1 byte
    // 7 bytes padding (on 64-bit systems)
    double d;                  // 8 bytes, aligned to 8-byte boundary
};

void demonstrate_alignment(void) {
    printf("Alignment requirements:\n");
    printf("char: %zu\n", alignof(char));
    printf("int: %zu\n", alignof(int));
    printf("double: %zu\n", alignof(double));
    printf("struct: %zu\n", alignof(struct aligned_struct));

    printf("\nSizes:\n");
    printf("char: %zu\n", sizeof(char));
    printf("int: %zu\n", sizeof(int));
    printf("double: %zu\n", sizeof(double));
    printf("struct: %zu\n", sizeof(struct aligned_struct));

    // Manual alignment
    char buffer[256];
    void* aligned_ptr = (void*)(((uintptr_t)buffer + 15) & ~15); // 16-byte aligned

    printf("Buffer: %p, Aligned: %p\n", (void*)buffer, aligned_ptr);
}
```

The C type system is both simple and powerful, providing the foundation for all data manipulation while allowing direct control over memory layout and performance characteristics.