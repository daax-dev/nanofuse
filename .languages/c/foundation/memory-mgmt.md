# C Memory Management Comprehensive Guide

## Memory Layout in C Programs

### Program Memory Segments
```c
#include <stdio.h>
#include <stdlib.h>

// Global/Static data segment (initialized)
int global_initialized = 42;
static int static_initialized = 100;

// BSS segment (uninitialized global/static)
int global_uninitialized;
static int static_uninitialized;

// Text segment (code)
void function_in_text_segment(void) {
    // Stack segment (local variables, parameters)
    int local_variable = 10;
    char local_array[100];

    // Heap segment (dynamic allocation)
    int* heap_memory = malloc(sizeof(int) * 10);

    printf("Stack variable address: %p\n", (void*)&local_variable);
    printf("Heap memory address: %p\n", (void*)heap_memory);
    printf("Global variable address: %p\n", (void*)&global_initialized);
    printf("Function address: %p\n", (void*)function_in_text_segment);

    free(heap_memory);
}
```

### Stack vs Heap Comparison

| Aspect | Stack | Heap |
|--------|-------|------|
| **Speed** | Very fast | Slower |
| **Size** | Limited (typically 1-8MB) | Large (limited by system memory) |
| **Management** | Automatic | Manual |
| **Allocation Pattern** | LIFO (Last In, First Out) | Random access |
| **Fragmentation** | None | Can occur |
| **Thread Safety** | Thread-local | Shared, needs synchronization |

## Dynamic Memory Allocation

### Standard Memory Allocation Functions

#### malloc() - Memory Allocation
```c
#include <stdlib.h>
#include <stdio.h>

void demonstrate_malloc(void) {
    // Allocate memory for 10 integers
    int* numbers = malloc(10 * sizeof(int));

    if (numbers == NULL) {
        fprintf(stderr, "Memory allocation failed\n");
        exit(EXIT_FAILURE);
    }

    // Initialize allocated memory
    for (int i = 0; i < 10; i++) {
        numbers[i] = i * i;
    }

    // Use the memory...
    for (int i = 0; i < 10; i++) {
        printf("%d ", numbers[i]);
    }
    printf("\n");

    // Always free allocated memory
    free(numbers);
    numbers = NULL;  // Avoid dangling pointer
}
```

#### calloc() - Cleared Allocation
```c
void demonstrate_calloc(void) {
    // Allocate and zero-initialize memory for 10 integers
    int* numbers = calloc(10, sizeof(int));

    if (numbers == NULL) {
        fprintf(stderr, "Memory allocation failed\n");
        exit(EXIT_FAILURE);
    }

    // Memory is already zero-initialized
    printf("Calloc initialized values: ");
    for (int i = 0; i < 10; i++) {
        printf("%d ", numbers[i]);  // All zeros
    }
    printf("\n");

    free(numbers);
}
```

#### realloc() - Memory Reallocation
```c
#include <string.h>

void demonstrate_realloc(void) {
    // Initial allocation
    int* numbers = malloc(5 * sizeof(int));
    if (!numbers) exit(EXIT_FAILURE);

    // Initialize
    for (int i = 0; i < 5; i++) {
        numbers[i] = i + 1;
    }

    // Expand the array
    int* temp = realloc(numbers, 10 * sizeof(int));
    if (!temp) {
        free(numbers);  // Original memory still valid
        exit(EXIT_FAILURE);
    }
    numbers = temp;

    // Initialize new elements
    for (int i = 5; i < 10; i++) {
        numbers[i] = i + 1;
    }

    // Print all elements
    for (int i = 0; i < 10; i++) {
        printf("%d ", numbers[i]);
    }
    printf("\n");

    free(numbers);
}
```

#### free() - Memory Deallocation
```c
void demonstrate_proper_free(void) {
    char* buffer = malloc(100);
    if (!buffer) return;

    // Use the buffer...
    strcpy(buffer, "Hello, World!");
    printf("%s\n", buffer);

    // Proper deallocation
    free(buffer);
    buffer = NULL;  // Prevent use-after-free bugs

    // Attempting to use buffer now would be undefined behavior
    // printf("%s\n", buffer);  // DON'T DO THIS
}
```

## Memory Management Patterns

### RAII-style Resource Management
```c
typedef struct {
    char* data;
    size_t size;
    size_t capacity;
} buffer_t;

// Constructor pattern
buffer_t* buffer_create(size_t initial_capacity) {
    buffer_t* buf = malloc(sizeof(buffer_t));
    if (!buf) return NULL;

    buf->data = malloc(initial_capacity);
    if (!buf->data) {
        free(buf);
        return NULL;
    }

    buf->size = 0;
    buf->capacity = initial_capacity;
    return buf;
}

// Destructor pattern
void buffer_destroy(buffer_t* buf) {
    if (buf) {
        free(buf->data);
        free(buf);
    }
}

// Growth pattern with error handling
int buffer_ensure_capacity(buffer_t* buf, size_t required_capacity) {
    if (!buf) return -1;

    if (required_capacity <= buf->capacity) {
        return 0;  // Already sufficient
    }

    size_t new_capacity = buf->capacity * 2;
    if (new_capacity < required_capacity) {
        new_capacity = required_capacity;
    }

    char* new_data = realloc(buf->data, new_capacity);
    if (!new_data) {
        return -1;  // Allocation failed
    }

    buf->data = new_data;
    buf->capacity = new_capacity;
    return 0;
}
```

### Object Pool Pattern
```c
#define POOL_SIZE 100

typedef struct {
    void* objects[POOL_SIZE];
    bool available[POOL_SIZE];
    size_t object_size;
    size_t count;
    pthread_mutex_t mutex;
} object_pool_t;

object_pool_t* pool_create(size_t object_size) {
    object_pool_t* pool = malloc(sizeof(object_pool_t));
    if (!pool) return NULL;

    pool->object_size = object_size;
    pool->count = 0;

    // Pre-allocate objects
    for (size_t i = 0; i < POOL_SIZE; i++) {
        pool->objects[i] = malloc(object_size);
        if (!pool->objects[i]) {
            // Cleanup on failure
            for (size_t j = 0; j < i; j++) {
                free(pool->objects[j]);
            }
            free(pool);
            return NULL;
        }
        pool->available[i] = true;
    }

    pool->count = POOL_SIZE;
    pthread_mutex_init(&pool->mutex, NULL);
    return pool;
}

void* pool_acquire(object_pool_t* pool) {
    if (!pool) return NULL;

    pthread_mutex_lock(&pool->mutex);

    for (size_t i = 0; i < pool->count; i++) {
        if (pool->available[i]) {
            pool->available[i] = false;
            pthread_mutex_unlock(&pool->mutex);
            return pool->objects[i];
        }
    }

    pthread_mutex_unlock(&pool->mutex);
    return NULL;  // Pool exhausted
}

void pool_release(object_pool_t* pool, void* object) {
    if (!pool || !object) return;

    pthread_mutex_lock(&pool->mutex);

    for (size_t i = 0; i < pool->count; i++) {
        if (pool->objects[i] == object) {
            pool->available[i] = true;
            break;
        }
    }

    pthread_mutex_unlock(&pool->mutex);
}
```

### Memory Arena/Linear Allocator
```c
typedef struct {
    char* memory;
    size_t size;
    size_t used;
} arena_t;

arena_t* arena_create(size_t size) {
    arena_t* arena = malloc(sizeof(arena_t));
    if (!arena) return NULL;

    arena->memory = malloc(size);
    if (!arena->memory) {
        free(arena);
        return NULL;
    }

    arena->size = size;
    arena->used = 0;
    return arena;
}

void* arena_alloc(arena_t* arena, size_t size) {
    if (!arena) return NULL;

    // Align to pointer size
    size_t aligned_size = (size + sizeof(void*) - 1) & ~(sizeof(void*) - 1);

    if (arena->used + aligned_size > arena->size) {
        return NULL;  // Arena exhausted
    }

    void* ptr = arena->memory + arena->used;
    arena->used += aligned_size;
    return ptr;
}

void arena_reset(arena_t* arena) {
    if (arena) {
        arena->used = 0;  // Reset without freeing individual allocations
    }
}

void arena_destroy(arena_t* arena) {
    if (arena) {
        free(arena->memory);
        free(arena);
    }
}
```

## Common Memory Errors

### Memory Leaks
```c
// BAD: Memory leak
void memory_leak_example(void) {
    char* buffer = malloc(1000);
    // ... use buffer ...
    // Missing free(buffer) - MEMORY LEAK!
}

// GOOD: Proper cleanup
void proper_cleanup_example(void) {
    char* buffer = malloc(1000);
    if (!buffer) return;

    // ... use buffer ...

    free(buffer);
    buffer = NULL;
}

// GOOD: RAII-style with goto cleanup
int complex_function(void) {
    char* buffer1 = NULL;
    char* buffer2 = NULL;
    FILE* file = NULL;
    int result = -1;

    buffer1 = malloc(1000);
    if (!buffer1) goto cleanup;

    buffer2 = malloc(2000);
    if (!buffer2) goto cleanup;

    file = fopen("data.txt", "r");
    if (!file) goto cleanup;

    // ... process ...
    result = 0;  // Success

cleanup:
    if (file) fclose(file);
    free(buffer2);
    free(buffer1);
    return result;
}
```

### Use-After-Free
```c
// BAD: Use after free
void use_after_free_bug(void) {
    char* buffer = malloc(100);
    strcpy(buffer, "Hello");

    free(buffer);

    // BUG: Using freed memory
    printf("%s\n", buffer);  // Undefined behavior!
}

// GOOD: Null after free
void safe_pattern(void) {
    char* buffer = malloc(100);
    strcpy(buffer, "Hello");
    printf("%s\n", buffer);

    free(buffer);
    buffer = NULL;  // Prevent accidental use
}
```

### Double Free
```c
// BAD: Double free
void double_free_bug(void) {
    char* buffer = malloc(100);

    free(buffer);
    free(buffer);  // BUG: Double free!
}

// GOOD: Safe free macro
#define SAFE_FREE(ptr) do { free(ptr); (ptr) = NULL; } while(0)

void safe_free_example(void) {
    char* buffer = malloc(100);

    SAFE_FREE(buffer);
    SAFE_FREE(buffer);  // Safe - becomes free(NULL) which is valid
}
```

### Buffer Overruns
```c
// BAD: Buffer overrun
void buffer_overrun_bug(void) {
    char* buffer = malloc(10);

    // BUG: Writing beyond allocated memory
    strcpy(buffer, "This string is too long for the buffer!");

    free(buffer);
}

// GOOD: Bounds checking
void safe_string_copy(void) {
    const char* source = "This string is too long for the buffer!";
    size_t buffer_size = 10;

    char* buffer = malloc(buffer_size);
    if (!buffer) return;

    // Safe copy with bounds checking
    size_t source_len = strlen(source);
    if (source_len >= buffer_size) {
        fprintf(stderr, "String too long for buffer\n");
        free(buffer);
        return;
    }

    strcpy(buffer, source);
    free(buffer);
}
```

## Memory Debugging Tools

### Manual Debugging Techniques
```c
#ifdef DEBUG_MEMORY
    #define DEBUG_MALLOC(size) debug_malloc(size, __FILE__, __LINE__)
    #define DEBUG_FREE(ptr) debug_free(ptr, __FILE__, __LINE__)

    static size_t total_allocated = 0;
    static size_t allocation_count = 0;

    void* debug_malloc(size_t size, const char* file, int line) {
        void* ptr = malloc(size);
        if (ptr) {
            total_allocated += size;
            allocation_count++;
            printf("MALLOC: %zu bytes at %p (%s:%d) - Total: %zu\n",
                   size, ptr, file, line, total_allocated);
        }
        return ptr;
    }

    void debug_free(void* ptr, const char* file, int line) {
        if (ptr) {
            free(ptr);
            allocation_count--;
            printf("FREE: %p (%s:%d) - Outstanding: %zu\n",
                   ptr, file, line, allocation_count);
        }
    }
#else
    #define DEBUG_MALLOC(size) malloc(size)
    #define DEBUG_FREE(ptr) free(ptr)
#endif
```

### Valgrind Integration
```c
// Valgrind-friendly code patterns
#include <valgrind/valgrind.h>
#include <valgrind/memcheck.h>

void valgrind_example(void) {
    char* buffer = malloc(100);

    // Mark memory as uninitialized for Valgrind
    VALGRIND_MAKE_MEM_UNDEFINED(buffer, 100);

    // Initialize only part of the buffer
    memset(buffer, 0, 50);
    VALGRIND_MAKE_MEM_DEFINED(buffer, 50);

    // Use the buffer...

    free(buffer);
}
```

### AddressSanitizer (ASan) Compatible Code
```c
// Compile with: gcc -fsanitize=address -g
#include <sanitizer/asan_interface.h>

void asan_example(void) {
    char* buffer = malloc(100);

    // This will be caught by ASan
    // buffer[100] = 'x';  // Buffer overflow

    free(buffer);

    // This will be caught by ASan
    // buffer[0] = 'y';  // Use after free
}
```

## Advanced Memory Management

### Custom Allocators
```c
// Simple slab allocator for fixed-size objects
typedef struct slab_block {
    struct slab_block* next;
} slab_block_t;

typedef struct {
    void* memory;
    slab_block_t* free_list;
    size_t object_size;
    size_t block_count;
} slab_allocator_t;

slab_allocator_t* slab_create(size_t object_size, size_t count) {
    slab_allocator_t* slab = malloc(sizeof(slab_allocator_t));
    if (!slab) return NULL;

    // Ensure minimum size for free list linkage
    if (object_size < sizeof(slab_block_t)) {
        object_size = sizeof(slab_block_t);
    }

    slab->memory = malloc(object_size * count);
    if (!slab->memory) {
        free(slab);
        return NULL;
    }

    slab->object_size = object_size;
    slab->block_count = count;

    // Initialize free list
    slab->free_list = NULL;
    char* ptr = (char*)slab->memory;
    for (size_t i = 0; i < count; i++) {
        slab_block_t* block = (slab_block_t*)ptr;
        block->next = slab->free_list;
        slab->free_list = block;
        ptr += object_size;
    }

    return slab;
}

void* slab_alloc(slab_allocator_t* slab) {
    if (!slab || !slab->free_list) {
        return NULL;
    }

    slab_block_t* block = slab->free_list;
    slab->free_list = block->next;
    return block;
}

void slab_free(slab_allocator_t* slab, void* ptr) {
    if (!slab || !ptr) return;

    slab_block_t* block = (slab_block_t*)ptr;
    block->next = slab->free_list;
    slab->free_list = block;
}
```

### Memory-Mapped Files
```c
#include <sys/mman.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>

typedef struct {
    void* data;
    size_t size;
    int fd;
} mmap_file_t;

mmap_file_t* mmap_file_open(const char* filename) {
    mmap_file_t* mf = malloc(sizeof(mmap_file_t));
    if (!mf) return NULL;

    mf->fd = open(filename, O_RDWR);
    if (mf->fd == -1) {
        free(mf);
        return NULL;
    }

    struct stat st;
    if (fstat(mf->fd, &st) == -1) {
        close(mf->fd);
        free(mf);
        return NULL;
    }

    mf->size = st.st_size;
    mf->data = mmap(NULL, mf->size, PROT_READ | PROT_WRITE, MAP_SHARED, mf->fd, 0);
    if (mf->data == MAP_FAILED) {
        close(mf->fd);
        free(mf);
        return NULL;
    }

    return mf;
}

void mmap_file_close(mmap_file_t* mf) {
    if (mf) {
        munmap(mf->data, mf->size);
        close(mf->fd);
        free(mf);
    }
}
```

## Best Practices

### 1. Always Check Allocation Results
```c
void* safe_alloc(size_t size) {
    void* ptr = malloc(size);
    if (!ptr) {
        fprintf(stderr, "Out of memory: failed to allocate %zu bytes\n", size);
        exit(EXIT_FAILURE);
    }
    return ptr;
}
```

### 2. Use Static Analysis Tools
```bash
# Clang Static Analyzer
clang --analyze -Xanalyzer -analyzer-checker=alpha.unix.cstring.NotNullTerminated file.c

# PVS-Studio, Coverity, etc.
```

### 3. Implement Memory Statistics
```c
typedef struct {
    size_t total_allocated;
    size_t peak_usage;
    size_t current_usage;
    size_t allocation_count;
    size_t free_count;
} memory_stats_t;

memory_stats_t g_memory_stats = {0};

void* tracked_malloc(size_t size) {
    void* ptr = malloc(size);
    if (ptr) {
        g_memory_stats.total_allocated += size;
        g_memory_stats.current_usage += size;
        g_memory_stats.allocation_count++;

        if (g_memory_stats.current_usage > g_memory_stats.peak_usage) {
            g_memory_stats.peak_usage = g_memory_stats.current_usage;
        }
    }
    return ptr;
}
```

### 4. Memory Alignment Considerations
```c
// Ensure proper alignment for performance
void* aligned_alloc_portable(size_t alignment, size_t size) {
    #if defined(_WIN32)
        return _aligned_malloc(size, alignment);
    #elif defined(__STDC_VERSION__) && __STDC_VERSION__ >= 201112L
        return aligned_alloc(alignment, size);
    #else
        // Fallback implementation
        void* ptr;
        if (posix_memalign(&ptr, alignment, size) != 0) {
            return NULL;
        }
        return ptr;
    #endif
}
```

Memory management in C requires discipline, understanding, and consistent application of best practices. The key is to establish patterns and stick to them throughout your codebase.