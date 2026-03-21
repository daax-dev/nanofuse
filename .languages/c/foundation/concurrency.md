# C Concurrency and Parallel Programming

## Threading Models in C

### POSIX Threads (pthreads)
```c
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

// Thread function signature
void* thread_function(void* arg) {
    int thread_id = *(int*)arg;
    printf("Thread %d starting\n", thread_id);

    // Simulate work
    sleep(1);

    printf("Thread %d finishing\n", thread_id);
    return NULL;
}

void basic_pthread_example(void) {
    const int num_threads = 4;
    pthread_t threads[num_threads];
    int thread_ids[num_threads];

    // Create threads
    for (int i = 0; i < num_threads; i++) {
        thread_ids[i] = i;
        int result = pthread_create(&threads[i], NULL, thread_function, &thread_ids[i]);
        if (result != 0) {
            fprintf(stderr, "Error creating thread %d\n", i);
            exit(EXIT_FAILURE);
        }
    }

    // Wait for threads to complete
    for (int i = 0; i < num_threads; i++) {
        int result = pthread_join(threads[i], NULL);
        if (result != 0) {
            fprintf(stderr, "Error joining thread %d\n", i);
        }
    }

    printf("All threads completed\n");
}
```

### C11 Threads (threads.h)
```c
#include <threads.h>
#include <stdio.h>
#include <time.h>

int c11_thread_function(void* arg) {
    int thread_id = *(int*)arg;
    printf("C11 Thread %d starting\n", thread_id);

    // Sleep for 1 second
    struct timespec ts = {1, 0};
    thrd_sleep(&ts, NULL);

    printf("C11 Thread %d finishing\n", thread_id);
    return 0;
}

void c11_threading_example(void) {
    const int num_threads = 4;
    thrd_t threads[num_threads];
    int thread_ids[num_threads];

    // Create threads
    for (int i = 0; i < num_threads; i++) {
        thread_ids[i] = i;
        int result = thrd_create(&threads[i], c11_thread_function, &thread_ids[i]);
        if (result != thrd_success) {
            fprintf(stderr, "Error creating thread %d\n", i);
            exit(EXIT_FAILURE);
        }
    }

    // Wait for threads
    for (int i = 0; i < num_threads; i++) {
        int result;
        thrd_join(threads[i], &result);
        printf("Thread %d returned %d\n", i, result);
    }
}
```

## Synchronization Primitives

### Mutexes (Mutual Exclusion)
```c
#include <pthread.h>

// Global shared data
int shared_counter = 0;
pthread_mutex_t counter_mutex = PTHREAD_MUTEX_INITIALIZER;

void* increment_thread(void* arg) {
    int iterations = *(int*)arg;

    for (int i = 0; i < iterations; i++) {
        // Lock before accessing shared data
        pthread_mutex_lock(&counter_mutex);
        shared_counter++;
        pthread_mutex_unlock(&counter_mutex);
    }

    return NULL;
}

void mutex_example(void) {
    const int num_threads = 4;
    const int iterations = 1000;
    pthread_t threads[num_threads];
    int thread_iterations = iterations;

    for (int i = 0; i < num_threads; i++) {
        pthread_create(&threads[i], NULL, increment_thread, &thread_iterations);
    }

    for (int i = 0; i < num_threads; i++) {
        pthread_join(threads[i], NULL);
    }

    printf("Final counter value: %d (expected: %d)\n",
           shared_counter, num_threads * iterations);

    pthread_mutex_destroy(&counter_mutex);
}
```

### Reader-Writer Locks
```c
#include <pthread.h>

int shared_data = 0;
pthread_rwlock_t rw_lock = PTHREAD_RWLOCK_INITIALIZER;

void* reader_thread(void* arg) {
    int thread_id = *(int*)arg;

    for (int i = 0; i < 5; i++) {
        // Acquire read lock
        pthread_rwlock_rdlock(&rw_lock);

        printf("Reader %d: data = %d\n", thread_id, shared_data);
        usleep(100000);  // 0.1 second

        pthread_rwlock_unlock(&rw_lock);
        usleep(50000);   // Brief pause
    }

    return NULL;
}

void* writer_thread(void* arg) {
    int thread_id = *(int*)arg;

    for (int i = 0; i < 3; i++) {
        // Acquire write lock
        pthread_rwlock_wrlock(&rw_lock);

        shared_data++;
        printf("Writer %d: updated data to %d\n", thread_id, shared_data);
        usleep(200000);  // 0.2 second

        pthread_rwlock_unlock(&rw_lock);
        usleep(100000);  // Brief pause
    }

    return NULL;
}

void reader_writer_example(void) {
    pthread_t readers[3], writers[2];
    int reader_ids[3] = {1, 2, 3};
    int writer_ids[2] = {1, 2};

    // Create reader threads
    for (int i = 0; i < 3; i++) {
        pthread_create(&readers[i], NULL, reader_thread, &reader_ids[i]);
    }

    // Create writer threads
    for (int i = 0; i < 2; i++) {
        pthread_create(&writers[i], NULL, writer_thread, &writer_ids[i]);
    }

    // Wait for all threads
    for (int i = 0; i < 3; i++) {
        pthread_join(readers[i], NULL);
    }
    for (int i = 0; i < 2; i++) {
        pthread_join(writers[i], NULL);
    }

    pthread_rwlock_destroy(&rw_lock);
}
```

### Condition Variables
```c
#include <pthread.h>

typedef struct {
    int* buffer;
    int capacity;
    int count;
    int in;
    int out;
    pthread_mutex_t mutex;
    pthread_cond_t not_full;
    pthread_cond_t not_empty;
} bounded_buffer_t;

bounded_buffer_t* buffer_create(int capacity) {
    bounded_buffer_t* bb = malloc(sizeof(bounded_buffer_t));
    if (!bb) return NULL;

    bb->buffer = malloc(capacity * sizeof(int));
    if (!bb->buffer) {
        free(bb);
        return NULL;
    }

    bb->capacity = capacity;
    bb->count = 0;
    bb->in = 0;
    bb->out = 0;

    pthread_mutex_init(&bb->mutex, NULL);
    pthread_cond_init(&bb->not_full, NULL);
    pthread_cond_init(&bb->not_empty, NULL);

    return bb;
}

void buffer_put(bounded_buffer_t* bb, int item) {
    pthread_mutex_lock(&bb->mutex);

    // Wait while buffer is full
    while (bb->count == bb->capacity) {
        pthread_cond_wait(&bb->not_full, &bb->mutex);
    }

    // Add item to buffer
    bb->buffer[bb->in] = item;
    bb->in = (bb->in + 1) % bb->capacity;
    bb->count++;

    printf("Produced: %d (count: %d)\n", item, bb->count);

    // Signal that buffer is not empty
    pthread_cond_signal(&bb->not_empty);
    pthread_mutex_unlock(&bb->mutex);
}

int buffer_get(bounded_buffer_t* bb) {
    pthread_mutex_lock(&bb->mutex);

    // Wait while buffer is empty
    while (bb->count == 0) {
        pthread_cond_wait(&bb->not_empty, &bb->mutex);
    }

    // Remove item from buffer
    int item = bb->buffer[bb->out];
    bb->out = (bb->out + 1) % bb->capacity;
    bb->count--;

    printf("Consumed: %d (count: %d)\n", item, bb->count);

    // Signal that buffer is not full
    pthread_cond_signal(&bb->not_full);
    pthread_mutex_unlock(&bb->mutex);

    return item;
}
```

### Semaphores
```c
#include <semaphore.h>
#include <pthread.h>

sem_t semaphore;
int shared_resource = 0;

void* worker_thread(void* arg) {
    int thread_id = *(int*)arg;

    for (int i = 0; i < 3; i++) {
        // Wait for semaphore (decrement)
        sem_wait(&semaphore);

        printf("Thread %d acquired resource\n", thread_id);
        shared_resource++;
        sleep(1);  // Simulate work with resource
        printf("Thread %d releasing resource\n", thread_id);

        // Signal semaphore (increment)
        sem_post(&semaphore);

        sleep(1);  // Brief pause
    }

    return NULL;
}

void semaphore_example(void) {
    // Initialize semaphore with count of 2 (max 2 threads can access resource)
    sem_init(&semaphore, 0, 2);

    const int num_threads = 5;
    pthread_t threads[num_threads];
    int thread_ids[num_threads];

    for (int i = 0; i < num_threads; i++) {
        thread_ids[i] = i;
        pthread_create(&threads[i], NULL, worker_thread, &thread_ids[i]);
    }

    for (int i = 0; i < num_threads; i++) {
        pthread_join(threads[i], NULL);
    }

    sem_destroy(&semaphore);
}
```

## Atomic Operations (C11)

### Basic Atomic Types and Operations
```c
#include <stdatomic.h>
#include <pthread.h>

// Atomic variables
atomic_int atomic_counter = 0;
atomic_bool atomic_flag = false;
atomic_long atomic_sum = 0;

void* atomic_increment_thread(void* arg) {
    int iterations = *(int*)arg;

    for (int i = 0; i < iterations; i++) {
        // Atomic increment
        atomic_fetch_add(&atomic_counter, 1);

        // Atomic addition to sum
        atomic_fetch_add(&atomic_sum, i);
    }

    return NULL;
}

void atomic_operations_example(void) {
    const int num_threads = 4;
    const int iterations = 1000;
    pthread_t threads[num_threads];
    int thread_iterations = iterations;

    for (int i = 0; i < num_threads; i++) {
        pthread_create(&threads[i], NULL, atomic_increment_thread, &thread_iterations);
    }

    for (int i = 0; i < num_threads; i++) {
        pthread_join(threads[i], NULL);
    }

    printf("Atomic counter: %d (expected: %d)\n",
           atomic_counter, num_threads * iterations);
    printf("Atomic sum: %ld\n", atomic_sum);
}
```

### Memory Ordering
```c
#include <stdatomic.h>

atomic_int data = 0;
atomic_bool ready = false;

void* producer_thread(void* arg) {
    // Write data first
    atomic_store_explicit(&data, 42, memory_order_relaxed);

    // Then signal ready (release semantics)
    atomic_store_explicit(&ready, true, memory_order_release);

    return NULL;
}

void* consumer_thread(void* arg) {
    // Wait until ready (acquire semantics)
    while (!atomic_load_explicit(&ready, memory_order_acquire)) {
        // Busy wait
    }

    // Read data (guaranteed to see the write due to acquire-release)
    int value = atomic_load_explicit(&data, memory_order_relaxed);
    printf("Consumer read: %d\n", value);

    return NULL;
}

void memory_ordering_example(void) {
    pthread_t producer, consumer;

    pthread_create(&consumer, NULL, consumer_thread, NULL);
    pthread_create(&producer, NULL, producer_thread, NULL);

    pthread_join(producer, NULL);
    pthread_join(consumer, NULL);
}
```

### Compare-and-Swap Operations
```c
#include <stdatomic.h>

atomic_int shared_value = 0;

bool atomic_increment_if_less_than(atomic_int* value, int threshold) {
    int current = atomic_load(value);
    int desired;

    do {
        if (current >= threshold) {
            return false;  // Value is already >= threshold
        }

        desired = current + 1;

        // Attempt to update: if current value is still 'current', set it to 'desired'
    } while (!atomic_compare_exchange_weak(value, &current, desired));

    return true;  // Successfully incremented
}

void* cas_thread(void* arg) {
    int threshold = *(int*)arg;

    for (int i = 0; i < 1000; i++) {
        if (atomic_increment_if_less_than(&shared_value, threshold)) {
            // Successfully incremented
        } else {
            // Value reached threshold
            break;
        }
    }

    return NULL;
}

void compare_and_swap_example(void) {
    const int num_threads = 4;
    const int threshold = 100;
    pthread_t threads[num_threads];
    int thread_threshold = threshold;

    for (int i = 0; i < num_threads; i++) {
        pthread_create(&threads[i], NULL, cas_thread, &thread_threshold);
    }

    for (int i = 0; i < num_threads; i++) {
        pthread_join(threads[i], NULL);
    }

    printf("Final value: %d\n", atomic_load(&shared_value));
}
```

## Lock-Free Programming

### Lock-Free Stack
```c
#include <stdatomic.h>
#include <stdlib.h>

typedef struct stack_node {
    int data;
    struct stack_node* next;
} stack_node_t;

typedef struct {
    atomic(stack_node_t*) head;
} lockfree_stack_t;

void stack_init(lockfree_stack_t* stack) {
    atomic_store(&stack->head, NULL);
}

void stack_push(lockfree_stack_t* stack, int data) {
    stack_node_t* new_node = malloc(sizeof(stack_node_t));
    new_node->data = data;

    stack_node_t* current_head = atomic_load(&stack->head);

    do {
        new_node->next = current_head;
    } while (!atomic_compare_exchange_weak(&stack->head, &current_head, new_node));
}

bool stack_pop(lockfree_stack_t* stack, int* result) {
    stack_node_t* current_head = atomic_load(&stack->head);

    do {
        if (current_head == NULL) {
            return false;  // Stack is empty
        }
    } while (!atomic_compare_exchange_weak(&stack->head, &current_head, current_head->next));

    *result = current_head->data;
    free(current_head);
    return true;
}
```

### Lock-Free Queue (Michael & Scott Algorithm)
```c
typedef struct queue_node {
    atomic(int) data;
    atomic(struct queue_node*) next;
} queue_node_t;

typedef struct {
    atomic(queue_node_t*) head;
    atomic(queue_node_t*) tail;
} lockfree_queue_t;

void queue_init(lockfree_queue_t* queue) {
    queue_node_t* dummy = malloc(sizeof(queue_node_t));
    atomic_store(&dummy->next, NULL);
    atomic_store(&queue->head, dummy);
    atomic_store(&queue->tail, dummy);
}

void queue_enqueue(lockfree_queue_t* queue, int data) {
    queue_node_t* new_node = malloc(sizeof(queue_node_t));
    atomic_store(&new_node->data, data);
    atomic_store(&new_node->next, NULL);

    queue_node_t* tail;
    queue_node_t* next;

    while (true) {
        tail = atomic_load(&queue->tail);
        next = atomic_load(&tail->next);

        if (tail == atomic_load(&queue->tail)) {  // Are tail and next consistent?
            if (next == NULL) {
                // Last node, try to link new node
                if (atomic_compare_exchange_weak(&tail->next, &next, new_node)) {
                    // Enqueue done, try to advance tail
                    atomic_compare_exchange_weak(&queue->tail, &tail, new_node);
                    return;
                }
            } else {
                // Not the last node, try to advance tail
                atomic_compare_exchange_weak(&queue->tail, &tail, next);
            }
        }
    }
}
```

## Thread Pool Pattern
```c
#include <pthread.h>
#include <stdlib.h>
#include <stdbool.h>

typedef struct task {
    void (*function)(void* arg);
    void* arg;
    struct task* next;
} task_t;

typedef struct {
    pthread_t* threads;
    int thread_count;
    task_t* task_queue_head;
    task_t* task_queue_tail;
    pthread_mutex_t queue_mutex;
    pthread_cond_t queue_condition;
    bool shutdown;
} thread_pool_t;

void* worker_thread(void* pool_ptr) {
    thread_pool_t* pool = (thread_pool_t*)pool_ptr;

    while (true) {
        pthread_mutex_lock(&pool->queue_mutex);

        // Wait for tasks or shutdown signal
        while (pool->task_queue_head == NULL && !pool->shutdown) {
            pthread_cond_wait(&pool->queue_condition, &pool->queue_mutex);
        }

        if (pool->shutdown) {
            pthread_mutex_unlock(&pool->queue_mutex);
            break;
        }

        // Get task from queue
        task_t* task = pool->task_queue_head;
        pool->task_queue_head = task->next;
        if (pool->task_queue_head == NULL) {
            pool->task_queue_tail = NULL;
        }

        pthread_mutex_unlock(&pool->queue_mutex);

        // Execute task
        task->function(task->arg);
        free(task);
    }

    return NULL;
}

thread_pool_t* thread_pool_create(int thread_count) {
    thread_pool_t* pool = malloc(sizeof(thread_pool_t));
    if (!pool) return NULL;

    pool->threads = malloc(thread_count * sizeof(pthread_t));
    if (!pool->threads) {
        free(pool);
        return NULL;
    }

    pool->thread_count = thread_count;
    pool->task_queue_head = NULL;
    pool->task_queue_tail = NULL;
    pool->shutdown = false;

    pthread_mutex_init(&pool->queue_mutex, NULL);
    pthread_cond_init(&pool->queue_condition, NULL);

    // Create worker threads
    for (int i = 0; i < thread_count; i++) {
        pthread_create(&pool->threads[i], NULL, worker_thread, pool);
    }

    return pool;
}

void thread_pool_add_task(thread_pool_t* pool, void (*function)(void*), void* arg) {
    task_t* task = malloc(sizeof(task_t));
    task->function = function;
    task->arg = arg;
    task->next = NULL;

    pthread_mutex_lock(&pool->queue_mutex);

    if (pool->task_queue_tail == NULL) {
        pool->task_queue_head = task;
        pool->task_queue_tail = task;
    } else {
        pool->task_queue_tail->next = task;
        pool->task_queue_tail = task;
    }

    pthread_cond_signal(&pool->queue_condition);
    pthread_mutex_unlock(&pool->queue_mutex);
}

void thread_pool_destroy(thread_pool_t* pool) {
    if (!pool) return;

    pthread_mutex_lock(&pool->queue_mutex);
    pool->shutdown = true;
    pthread_cond_broadcast(&pool->queue_condition);
    pthread_mutex_unlock(&pool->queue_mutex);

    // Wait for all threads to finish
    for (int i = 0; i < pool->thread_count; i++) {
        pthread_join(pool->threads[i], NULL);
    }

    // Clean up remaining tasks
    task_t* current = pool->task_queue_head;
    while (current) {
        task_t* next = current->next;
        free(current);
        current = next;
    }

    pthread_mutex_destroy(&pool->queue_mutex);
    pthread_cond_destroy(&pool->queue_condition);
    free(pool->threads);
    free(pool);
}
```

## Parallel Algorithms

### Parallel Merge Sort
```c
#include <pthread.h>
#include <string.h>

typedef struct {
    int* array;
    int* temp;
    int left;
    int right;
    int depth;
    int max_depth;
} sort_args_t;

void merge(int* array, int* temp, int left, int mid, int right) {
    int i = left, j = mid + 1, k = left;

    while (i <= mid && j <= right) {
        if (array[i] <= array[j]) {
            temp[k++] = array[i++];
        } else {
            temp[k++] = array[j++];
        }
    }

    while (i <= mid) temp[k++] = array[i++];
    while (j <= right) temp[k++] = array[j++];

    for (i = left; i <= right; i++) {
        array[i] = temp[i];
    }
}

void* parallel_merge_sort(void* arg) {
    sort_args_t* args = (sort_args_t*)arg;

    if (args->left >= args->right) {
        return NULL;
    }

    int mid = args->left + (args->right - args->left) / 2;

    if (args->depth < args->max_depth) {
        // Create threads for parallel execution
        pthread_t left_thread, right_thread;
        sort_args_t left_args = {args->array, args->temp, args->left, mid,
                                args->depth + 1, args->max_depth};
        sort_args_t right_args = {args->array, args->temp, mid + 1, args->right,
                                 args->depth + 1, args->max_depth};

        pthread_create(&left_thread, NULL, parallel_merge_sort, &left_args);
        pthread_create(&right_thread, NULL, parallel_merge_sort, &right_args);

        pthread_join(left_thread, NULL);
        pthread_join(right_thread, NULL);
    } else {
        // Sequential execution for small subproblems
        sort_args_t left_args = {args->array, args->temp, args->left, mid,
                                args->depth + 1, args->max_depth};
        sort_args_t right_args = {args->array, args->temp, mid + 1, args->right,
                                 args->depth + 1, args->max_depth};

        parallel_merge_sort(&left_args);
        parallel_merge_sort(&right_args);
    }

    merge(args->array, args->temp, args->left, mid, args->right);
    return NULL;
}
```

## Performance Considerations

### Cache-Friendly Concurrency
```c
// False sharing example - BAD
struct bad_counter {
    volatile int counter1;  // These will likely be on the same cache line
    volatile int counter2;  // Causing false sharing between threads
};

// Cache-friendly version - GOOD
struct good_counter {
    volatile int counter1;
    char padding[64 - sizeof(int)];  // Pad to cache line boundary
    volatile int counter2;
    char padding2[64 - sizeof(int)];
};

// Even better - use atomic operations
struct atomic_counter {
    atomic_int counter1;
    char padding[64 - sizeof(atomic_int)];
    atomic_int counter2;
    char padding2[64 - sizeof(atomic_int)];
};
```

### NUMA-Aware Programming
```c
#ifdef __linux__
#include <numa.h>

void numa_aware_allocation(void) {
    if (numa_available() != -1) {
        // Allocate memory on the current CPU's NUMA node
        void* memory = numa_alloc_local(1024 * 1024);

        // Use memory...

        numa_free(memory, 1024 * 1024);
    }
}
#endif
```

## Best Practices

1. **Minimize Shared State**: Reduce the amount of data shared between threads
2. **Use Appropriate Synchronization**: Choose the right tool for the job
3. **Avoid Deadlocks**: Always acquire locks in the same order
4. **Prefer Atomic Operations**: For simple operations, atomics are often faster than mutexes
5. **Consider Cache Effects**: Be aware of false sharing and cache line bouncing
6. **Test Thoroughly**: Concurrent bugs are often hard to reproduce and debug

Concurrency in C requires careful design and thorough testing, but when done correctly, it can significantly improve application performance on multi-core systems.