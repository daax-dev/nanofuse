# Concurrency in Rust

## Core Concurrency Philosophy

Rust enables **fearless concurrency** through:
- **Ownership system prevents data races at compile time**
- **Send and Sync traits for thread safety**
- **Message passing preferred over shared state**
- **Zero-cost thread safety abstractions**

## Threading Basics

### Creating Threads
```rust
use std::thread;
use std::time::Duration;

// Basic thread creation
let handle = thread::spawn(|| {
    for i in 1..10 {
        println!("Thread: {}", i);
        thread::sleep(Duration::from_millis(1));
    }
});

// Main thread work
for i in 1..5 {
    println!("Main: {}", i);
    thread::sleep(Duration::from_millis(1));
}

// Wait for thread to finish
handle.join().unwrap();
```

### Moving Data into Threads
```rust
let v = vec![1, 2, 3];

// Move ownership to thread
let handle = thread::spawn(move || {
    println!("Vector: {:?}", v);
    // v is owned by this thread
});

// v is no longer accessible in main thread
handle.join().unwrap();
```

### Thread Builder
```rust
use std::thread;

let builder = thread::Builder::new()
    .name("worker-thread".into())
    .stack_size(32 * 1024); // 32KB stack

let handler = builder.spawn(|| {
    println!("Hello from {}", thread::current().name().unwrap());
}).unwrap();

handler.join().unwrap();
```

## Message Passing

### Channels (mpsc - Multiple Producer, Single Consumer)
```rust
use std::sync::mpsc;
use std::thread;

// Create channel
let (tx, rx) = mpsc::channel();

// Spawn thread with transmitter
thread::spawn(move || {
    let val = String::from("hi");
    tx.send(val).unwrap();
    // val is moved, cannot use it here
});

// Receive in main thread
let received = rx.recv().unwrap();
println!("Got: {}", received);
```

### Multiple Producers
```rust
use std::sync::mpsc;
use std::thread;
use std::time::Duration;

let (tx, rx) = mpsc::channel();

// Clone transmitter for multiple producers
let tx1 = tx.clone();
thread::spawn(move || {
    let vals = vec![
        String::from("hi"),
        String::from("from"),
        String::from("the"),
        String::from("thread"),
    ];

    for val in vals {
        tx1.send(val).unwrap();
        thread::sleep(Duration::from_secs(1));
    }
});

thread::spawn(move || {
    let vals = vec![
        String::from("more"),
        String::from("messages"),
        String::from("for"),
        String::from("you"),
    ];

    for val in vals {
        tx.send(val).unwrap();
        thread::sleep(Duration::from_secs(1));
    }
});

// Receive all messages
for received in rx {
    println!("Got: {}", received);
}
```

### Synchronous Channels
```rust
use std::sync::mpsc;
use std::thread;

// Bounded channel with buffer size of 0 (synchronous)
let (tx, rx) = mpsc::sync_channel(0);

thread::spawn(move || {
    println!("Sending...");
    tx.send(42).unwrap(); // Blocks until received
    println!("Sent!");
});

thread::sleep(std::time::Duration::from_secs(1));
println!("Receiving...");
let value = rx.recv().unwrap();
println!("Received: {}", value);
```

## Shared State Concurrency

### Mutex (Mutual Exclusion)
```rust
use std::sync::{Arc, Mutex};
use std::thread;

// Shared counter
let counter = Arc::new(Mutex::new(0));
let mut handles = vec![];

for _ in 0..10 {
    let counter = Arc::clone(&counter);
    let handle = thread::spawn(move || {
        let mut num = counter.lock().unwrap();
        *num += 1;
        // Lock is released when `num` goes out of scope
    });
    handles.push(handle);
}

for handle in handles {
    handle.join().unwrap();
}

println!("Result: {}", *counter.lock().unwrap());
```

### RwLock (Read-Write Lock)
```rust
use std::sync::{Arc, RwLock};
use std::thread;

let data = Arc::new(RwLock::new(5));

// Multiple readers
let mut handles = vec![];
for i in 0..5 {
    let data = Arc::clone(&data);
    let handle = thread::spawn(move || {
        let r = data.read().unwrap();
        println!("Reader {}: {}", i, *r);
        // Multiple readers can access simultaneously
    });
    handles.push(handle);
}

// Single writer
let data_clone = Arc::clone(&data);
let writer_handle = thread::spawn(move || {
    let mut w = data_clone.write().unwrap();
    *w += 1;
    println!("Writer updated value to {}", *w);
});

handles.push(writer_handle);

for handle in handles {
    handle.join().unwrap();
}
```

### Atomic Types
```rust
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;
use std::thread;

let counter = Arc::new(AtomicUsize::new(0));
let mut handles = vec![];

for _ in 0..10 {
    let counter = Arc::clone(&counter);
    let handle = thread::spawn(move || {
        for _ in 0..100 {
            counter.fetch_add(1, Ordering::SeqCst);
        }
    });
    handles.push(handle);
}

for handle in handles {
    handle.join().unwrap();
}

println!("Result: {}", counter.load(Ordering::SeqCst));
```

### Memory Ordering
```rust
use std::sync::atomic::{AtomicBool, AtomicUsize, Ordering};
use std::sync::Arc;
use std::thread;

let flag = Arc::new(AtomicBool::new(false));
let counter = Arc::new(AtomicUsize::new(0));

let flag_clone = Arc::clone(&flag);
let counter_clone = Arc::clone(&counter);

// Producer thread
let producer = thread::spawn(move || {
    counter_clone.store(42, Ordering::Relaxed);
    flag_clone.store(true, Ordering::Release); // Release ordering
});

// Consumer thread
let consumer = thread::spawn(move || {
    while !flag.load(Ordering::Acquire) { // Acquire ordering
        thread::yield_now();
    }
    let value = counter.load(Ordering::Relaxed);
    println!("Value: {}", value); // Guaranteed to see 42
});

producer.join().unwrap();
consumer.join().unwrap();
```

## Advanced Concurrency Patterns

### Thread Pool
```rust
use std::sync::{mpsc, Arc, Mutex};
use std::thread;

pub struct ThreadPool {
    workers: Vec<Worker>,
    sender: mpsc::Sender<Job>,
}

type Job = Box<dyn FnOnce() + Send + 'static>;

impl ThreadPool {
    pub fn new(size: usize) -> ThreadPool {
        assert!(size > 0);

        let (sender, receiver) = mpsc::channel();
        let receiver = Arc::new(Mutex::new(receiver));
        let mut workers = Vec::with_capacity(size);

        for id in 0..size {
            workers.push(Worker::new(id, Arc::clone(&receiver)));
        }

        ThreadPool { workers, sender }
    }

    pub fn execute<F>(&self, f: F)
    where
        F: FnOnce() + Send + 'static,
    {
        let job = Box::new(f);
        self.sender.send(job).unwrap();
    }
}

struct Worker {
    id: usize,
    thread: thread::JoinHandle<()>,
}

impl Worker {
    fn new(id: usize, receiver: Arc<Mutex<mpsc::Receiver<Job>>>) -> Worker {
        let thread = thread::spawn(move || loop {
            let job = receiver.lock().unwrap().recv().unwrap();
            println!("Worker {} got a job; executing.", id);
            job();
        });

        Worker { id, thread }
    }
}
```

### Barrier Synchronization
```rust
use std::sync::{Arc, Barrier};
use std::thread;

let mut handles = Vec::with_capacity(10);
let barrier = Arc::new(Barrier::new(10));

for _ in 0..10 {
    let c = Arc::clone(&barrier);
    handles.push(thread::spawn(move|| {
        println!("Before wait");
        c.wait(); // All threads wait here until all 10 arrive
        println!("After wait");
    }));
}

for handle in handles {
    handle.join().unwrap();
}
```

### Condvar (Condition Variable)
```rust
use std::sync::{Arc, Mutex, Condvar};
use std::thread;

let pair = Arc::new((Mutex::new(false), Condvar::new()));
let pair2 = Arc::clone(&pair);

// Worker thread
thread::spawn(move|| {
    let (lock, cvar) = &*pair2;
    let mut started = lock.lock().unwrap();

    println!("Worker: starting work");
    thread::sleep(std::time::Duration::from_secs(2));

    *started = true;
    cvar.notify_one(); // Notify waiting thread
    println!("Worker: work done");
});

// Main thread waits
let (lock, cvar) = &*pair;
let mut started = lock.lock().unwrap();
while !*started {
    started = cvar.wait(started).unwrap(); // Wait for notification
}
println!("Main: worker has finished");
```

## Async Programming

### Basic Async/Await
```rust
use std::time::Duration;

// Async function
async fn say_hello() {
    println!("Hello");
    // Simulate async work
    tokio::time::sleep(Duration::from_secs(1)).await;
    println!("World");
}

#[tokio::main]
async fn main() {
    say_hello().await;
}
```

### Multiple Async Operations
```rust
use tokio::time::{Duration, sleep};

async fn fetch_data(id: u32) -> String {
    sleep(Duration::from_millis(100)).await;
    format!("Data {}", id)
}

#[tokio::main]
async fn main() {
    // Sequential execution
    let data1 = fetch_data(1).await;
    let data2 = fetch_data(2).await;
    println!("{}, {}", data1, data2);

    // Concurrent execution
    let (result1, result2) = tokio::join!(
        fetch_data(3),
        fetch_data(4)
    );
    println!("{}, {}", result1, result2);

    // Multiple concurrent operations
    let handles: Vec<_> = (1..=5)
        .map(|i| tokio::spawn(fetch_data(i)))
        .collect();

    let results: Vec<String> = futures::future::try_join_all(handles)
        .await
        .unwrap()
        .into_iter()
        .collect();

    println!("Results: {:?}", results);
}
```

### Async Channels
```rust
use tokio::sync::mpsc;

#[tokio::main]
async fn main() {
    let (tx, mut rx) = mpsc::channel(32);

    tokio::spawn(async move {
        for i in 0..10 {
            tx.send(i).await.unwrap();
        }
    });

    while let Some(message) = rx.recv().await {
        println!("Received: {}", message);
    }
}
```

## Send and Sync Traits

### Send Trait
```rust
// Types that can be transferred between threads
// Most types implement Send automatically

// Examples of Send types:
// - i32, String, Vec<T> where T: Send
// - Box<T> where T: Send

// Examples of !Send types:
// - Rc<T> (not thread-safe reference counting)
// - raw pointers (*const T, *mut T)

fn is_send<T: Send>(t: T) -> T { t }

let value = String::from("Hello");
let sent_value = is_send(value); // Compiles because String: Send
```

### Sync Trait
```rust
// Types that are safe to share references between threads
// T: Sync means &T is Send

// Examples of Sync types:
// - i32, String (immutable data)
// - Mutex<T>, RwLock<T>, AtomicUsize
// - Arc<T> where T: Sync

// Examples of !Sync types:
// - Cell<T>, RefCell<T> (interior mutability without thread safety)
// - Rc<T>

fn is_sync<T: Sync>(t: &T) -> &T { t }

let value = 42i32;
let sync_ref = is_sync(&value); // Compiles because i32: Sync
```

### Custom Send/Sync Implementation
```rust
// Usually auto-implemented, but can be manual for unsafe types
struct MyBox(*mut u8);

// SAFETY: MyBox owns the data it points to
unsafe impl Send for MyBox {}

// SAFETY: MyBox is safe to share references between threads
unsafe impl Sync for MyBox {}

// Or explicitly opt-out
// impl !Send for MyBox {}
// impl !Sync for MyBox {}
```

## Deadlock Prevention

### Lock Ordering
```rust
use std::sync::{Arc, Mutex};
use std::thread;

fn transfer(from: &Mutex<i32>, to: &Mutex<i32>, amount: i32) {
    // Always acquire locks in the same order to prevent deadlock
    let (first, second) = if from as *const _ < to as *const _ {
        (from, to)
    } else {
        (to, from)
    };

    let _guard1 = first.lock().unwrap();
    let _guard2 = second.lock().unwrap();

    // Perform transfer logic here
}
```

### Timeout-based Locking
```rust
use std::sync::{Arc, Mutex};
use std::time::Duration;

let data = Arc::new(Mutex::new(0));

// Try to acquire lock with timeout
match data.try_lock() {
    Ok(mut guard) => {
        *guard += 1;
        println!("Lock acquired and data modified");
    },
    Err(_) => {
        println!("Could not acquire lock");
    }
}
```

## Performance Considerations

### Cache-Friendly Data Structures
```rust
use std::sync::atomic::{AtomicUsize, Ordering};

// False sharing example - avoid this
struct BadCounter {
    counter1: AtomicUsize,
    counter2: AtomicUsize, // Likely on same cache line
}

// Better: separate cache lines
#[repr(align(64))] // Cache line size
struct GoodCounter {
    counter: AtomicUsize,
}
```

### Work Stealing
```rust
use rayon::prelude::*;

// Parallel iteration with work stealing
let numbers: Vec<i32> = (0..1_000_000).collect();
let sum: i32 = numbers.par_iter().sum();

// Parallel processing with custom pool
let pool = rayon::ThreadPoolBuilder::new()
    .num_threads(8)
    .build()
    .unwrap();

pool.install(|| {
    let result: Vec<i32> = (0..1000)
        .into_par_iter()
        .map(|x| x * x)
        .collect();
});
```

## Testing Concurrent Code

### Testing for Data Races
```rust
#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::{Arc, Mutex};
    use std::thread;

    #[test]
    fn test_concurrent_counter() {
        let counter = Arc::new(Mutex::new(0));
        let mut handles = vec![];

        for _ in 0..100 {
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

        assert_eq!(*counter.lock().unwrap(), 100);
    }
}
```

## Best Practices

1. **Prefer message passing** over shared state when possible
2. **Use Arc<Mutex<T>>** for shared mutable state
3. **Consider RwLock** for read-heavy workloads
4. **Use atomic types** for simple shared counters/flags
5. **Avoid long-held locks** to prevent contention
6. **Be consistent with lock ordering** to prevent deadlocks
7. **Use channels** for producer-consumer patterns
8. **Consider async/await** for I/O-bound concurrency
9. **Profile concurrent code** to identify bottlenecks
10. **Test thoroughly** with tools like ThreadSanitizer

Rust's concurrency model eliminates data races at compile time while providing powerful abstractions for both shared-state and message-passing concurrency.