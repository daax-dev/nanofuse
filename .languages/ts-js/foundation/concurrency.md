# Concurrency, Parallelism, and Scaling in TypeScript/JavaScript

## Table of Contents
- [Understanding JavaScript Concurrency Model](#understanding-javascript-concurrency-model)
- [Event Loop Deep Dive](#event-loop-deep-dive)
- [Promises and Async Programming](#promises-and-async-programming)
- [Advanced Async Patterns](#advanced-async-patterns)
- [Web Workers and Parallelism](#web-workers-and-parallelism)
- [Node.js Concurrency](#nodejs-concurrency)
- [Scaling Patterns](#scaling-patterns)
- [Performance Optimization](#performance-optimization)
- [Error Handling in Concurrent Code](#error-handling-in-concurrent-code)

## Understanding JavaScript Concurrency Model

### Single-Threaded Nature
JavaScript runs on a single thread with an event-driven, non-blocking I/O model:

```javascript
// JavaScript is single-threaded - this runs sequentially
console.log('1. First');
console.log('2. Second');
console.log('3. Third');

// But asynchronous operations are non-blocking
console.log('1. Start');
setTimeout(() => console.log('2. Timeout callback'), 0);
console.log('3. End');
// Output: 1. Start, 3. End, 2. Timeout callback
```

### Concurrency vs Parallelism
```javascript
// Concurrency: Managing multiple tasks at once (time-slicing)
async function concurrentTasks() {
  console.log('Starting concurrent tasks');

  // These run concurrently, not in parallel
  const task1 = fetch('/api/data1');
  const task2 = fetch('/api/data2');
  const task3 = fetch('/api/data3');

  // Wait for all to complete
  const results = await Promise.all([task1, task2, task3]);
  return results;
}

// Parallelism: Actually running tasks simultaneously (different threads)
// This requires Web Workers or Worker Threads
function parallelTasks() {
  const worker1 = new Worker('cpu-intensive-task.js');
  const worker2 = new Worker('cpu-intensive-task.js');
  const worker3 = new Worker('cpu-intensive-task.js');

  // These run in parallel on different threads
  return Promise.all([
    new Promise(resolve => worker1.onmessage = e => resolve(e.data)),
    new Promise(resolve => worker2.onmessage = e => resolve(e.data)),
    new Promise(resolve => worker3.onmessage = e => resolve(e.data))
  ]);
}
```

## Event Loop Deep Dive

### Event Loop Components
```javascript
// Understanding the event loop phases
class EventLoopDemo {
  constructor() {
    this.phase = 1;
  }

  demonstrateEventLoop() {
    console.log('=== Event Loop Demo ===');

    // 1. Call Stack (synchronous execution)
    console.log('1. Synchronous code (call stack)');

    // 2. Microtask Queue (Promise callbacks, queueMicrotask)
    Promise.resolve().then(() => {
      console.log('3. Microtask (Promise.then)');
    });

    queueMicrotask(() => {
      console.log('4. Microtask (queueMicrotask)');
    });

    // 3. Macrotask Queue (setTimeout, setInterval, I/O)
    setTimeout(() => {
      console.log('5. Macrotask (setTimeout)');
    }, 0);

    setImmediate(() => {
      console.log('6. Macrotask (setImmediate) - Node.js only');
    });

    // 4. More synchronous code
    console.log('2. More synchronous code');

    // Output order: 1, 2, 3, 4, 5, (6 in Node.js)
  }
}

new EventLoopDemo().demonstrateEventLoop();
```

### Event Loop Phases (Node.js)
```javascript
// Node.js event loop has specific phases
const fs = require('fs');

console.log('=== Node.js Event Loop Phases ===');

// 1. Timer phase
setTimeout(() => console.log('Timer phase'), 0);

// 2. Pending callbacks phase
// (internal Node.js operations)

// 3. Poll phase
fs.readFile(__filename, () => {
  console.log('Poll phase (I/O callback)');

  // Check phase executions within I/O callback
  setTimeout(() => console.log('Timer within I/O'), 0);
  setImmediate(() => console.log('Check phase (setImmediate)'));
});

// 4. Check phase
setImmediate(() => console.log('Check phase'));

// 5. Close callbacks phase
const server = require('http').createServer();
server.on('close', () => console.log('Close callbacks phase'));
server.listen(0, () => server.close());

// Microtasks run between phases
process.nextTick(() => console.log('nextTick microtask'));
Promise.resolve().then(() => console.log('Promise microtask'));
```

### Custom Event Loop Control
```javascript
// Advanced event loop management
class TaskScheduler {
  constructor() {
    this.taskQueue = [];
    this.isRunning = false;
    this.maxTasksPerTick = 5;
  }

  schedule(task, priority = 0) {
    this.taskQueue.push({ task, priority, timestamp: Date.now() });
    this.taskQueue.sort((a, b) => b.priority - a.priority);

    if (!this.isRunning) {
      this.processQueue();
    }
  }

  async processQueue() {
    this.isRunning = true;

    while (this.taskQueue.length > 0) {
      const tasksThisTick = this.taskQueue.splice(0, this.maxTasksPerTick);

      // Process tasks
      await Promise.all(
        tasksThisTick.map(({ task }) => {
          try {
            return task();
          } catch (error) {
            console.error('Task failed:', error);
            return null;
          }
        })
      );

      // Yield control to allow other operations
      if (this.taskQueue.length > 0) {
        await new Promise(resolve => setTimeout(resolve, 0));
      }
    }

    this.isRunning = false;
  }

  // Time-slicing for CPU-intensive tasks
  async processWithTimeSlicing(heavyTask, timeSlice = 16) {
    return new Promise((resolve) => {
      const startTime = performance.now();

      const processor = () => {
        const sliceStart = performance.now();

        while (performance.now() - sliceStart < timeSlice) {
          const result = heavyTask();
          if (result !== undefined) {
            resolve(result);
            return;
          }
        }

        // Yield control and continue processing
        setTimeout(processor, 0);
      };

      processor();
    });
  }
}

// Usage
const scheduler = new TaskScheduler();

scheduler.schedule(async () => {
  console.log('High priority task');
}, 10);

scheduler.schedule(async () => {
  console.log('Low priority task');
}, 1);
```

## Promises and Async Programming

### Promise Fundamentals
```javascript
// Promise states and lifecycle
class AdvancedPromise {
  static create(executor) {
    return new Promise((resolve, reject) => {
      try {
        executor(resolve, reject);
      } catch (error) {
        reject(error);
      }
    });
  }

  // Promise factory patterns
  static immediate(value) {
    return Promise.resolve(value);
  }

  static delayed(value, ms) {
    return new Promise(resolve => {
      setTimeout(() => resolve(value), ms);
    });
  }

  static retry(promiseFactory, maxRetries = 3, delay = 1000) {
    return new Promise(async (resolve, reject) => {
      let lastError;

      for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
          const result = await promiseFactory();
          resolve(result);
          return;
        } catch (error) {
          lastError = error;

          if (attempt < maxRetries) {
            await new Promise(r => setTimeout(r, delay * attempt));
          }
        }
      }

      reject(lastError);
    });
  }

  static timeout(promise, ms, message = 'Operation timed out') {
    return Promise.race([
      promise,
      new Promise((_, reject) => {
        setTimeout(() => reject(new Error(message)), ms);
      })
    ]);
  }
}

// Usage examples
async function demonstratePromisePatterns() {
  // Retry pattern
  const fetchWithRetry = () => AdvancedPromise.retry(
    () => fetch('/api/unreliable-endpoint'),
    3,
    1000
  );

  // Timeout pattern
  const fetchWithTimeout = AdvancedPromise.timeout(
    fetch('/api/slow-endpoint'),
    5000,
    'Network request timed out'
  );

  // Combine patterns
  try {
    const result = await fetchWithTimeout;
    console.log('Success:', result);
  } catch (error) {
    console.error('Failed:', error);
  }
}
```

### Advanced Async/Await Patterns
```javascript
// Async iterator patterns
class AsyncDataProcessor {
  constructor(dataSource) {
    this.dataSource = dataSource;
  }

  async *processItems() {
    let hasMore = true;
    let offset = 0;
    const batchSize = 100;

    while (hasMore) {
      const batch = await this.dataSource.getBatch(offset, batchSize);

      if (batch.length === 0) {
        hasMore = false;
        break;
      }

      for (const item of batch) {
        yield await this.processItem(item);
      }

      offset += batchSize;
    }
  }

  async processItem(item) {
    // Simulate async processing
    await new Promise(resolve => setTimeout(resolve, 10));
    return { ...item, processed: true, timestamp: Date.now() };
  }

  // Parallel processing with concurrency control
  async processInParallel(concurrency = 5) {
    const results = [];
    const semaphore = new Semaphore(concurrency);
    const promises = [];

    for await (const item of this.processItems()) {
      const promise = semaphore.acquire().then(async (release) => {
        try {
          const result = await this.processItem(item);
          results.push(result);
          return result;
        } finally {
          release();
        }
      });

      promises.push(promise);
    }

    await Promise.all(promises);
    return results;
  }
}

// Semaphore for concurrency control
class Semaphore {
  constructor(limit) {
    this.limit = limit;
    this.current = 0;
    this.queue = [];
  }

  acquire() {
    return new Promise((resolve) => {
      if (this.current < this.limit) {
        this.current++;
        resolve(() => this.release());
      } else {
        this.queue.push(() => {
          this.current++;
          resolve(() => this.release());
        });
      }
    });
  }

  release() {
    this.current--;
    if (this.queue.length > 0) {
      const next = this.queue.shift();
      next();
    }
  }
}
```

## Advanced Async Patterns

### Producer-Consumer Pattern
```javascript
// Producer-Consumer with async queues
class AsyncQueue {
  constructor(maxSize = Infinity) {
    this.items = [];
    this.maxSize = maxSize;
    this.waitingConsumers = [];
    this.waitingProducers = [];
  }

  async enqueue(item) {
    if (this.items.length >= this.maxSize) {
      // Queue is full, wait for space
      await new Promise(resolve => {
        this.waitingProducers.push(resolve);
      });
    }

    this.items.push(item);

    // Notify waiting consumers
    if (this.waitingConsumers.length > 0) {
      const consumer = this.waitingConsumers.shift();
      consumer();
    }
  }

  async dequeue() {
    if (this.items.length === 0) {
      // Queue is empty, wait for items
      await new Promise(resolve => {
        this.waitingConsumers.push(resolve);
      });
    }

    const item = this.items.shift();

    // Notify waiting producers
    if (this.waitingProducers.length > 0) {
      const producer = this.waitingProducers.shift();
      producer();
    }

    return item;
  }

  get size() {
    return this.items.length;
  }

  get isEmpty() {
    return this.items.length === 0;
  }
}

// Producer-Consumer implementation
class ProducerConsumerSystem {
  constructor(queueSize = 10) {
    this.queue = new AsyncQueue(queueSize);
    this.isRunning = false;
    this.producers = [];
    this.consumers = [];
  }

  createProducer(name, generator) {
    const producer = async () => {
      console.log(`Producer ${name} started`);

      for await (const item of generator()) {
        await this.queue.enqueue({ item, producer: name, timestamp: Date.now() });
        console.log(`Producer ${name} produced:`, item);
      }

      console.log(`Producer ${name} finished`);
    };

    this.producers.push(producer);
    return producer;
  }

  createConsumer(name, processor) {
    const consumer = async () => {
      console.log(`Consumer ${name} started`);

      while (this.isRunning) {
        try {
          const data = await Promise.race([
            this.queue.dequeue(),
            new Promise((_, reject) =>
              setTimeout(() => reject(new Error('timeout')), 1000)
            )
          ]);

          console.log(`Consumer ${name} received:`, data);
          await processor(data);
        } catch (error) {
          if (error.message === 'timeout' && !this.isRunning) {
            break;
          }
        }
      }

      console.log(`Consumer ${name} finished`);
    };

    this.consumers.push(consumer);
    return consumer;
  }

  async start() {
    this.isRunning = true;

    const allTasks = [
      ...this.producers.map(p => p()),
      ...this.consumers.map(c => c())
    ];

    await Promise.all(allTasks);
  }

  stop() {
    this.isRunning = false;
  }
}

// Usage example
async function demonstrateProducerConsumer() {
  const system = new ProducerConsumerSystem(5);

  // Create data generator
  async function* dataGenerator(name, count) {
    for (let i = 0; i < count; i++) {
      yield `${name}-item-${i}`;
      await new Promise(resolve => setTimeout(resolve, Math.random() * 100));
    }
  }

  // Create processors
  async function processData(data) {
    await new Promise(resolve => setTimeout(resolve, Math.random() * 50));
    console.log(`Processed: ${data.item}`);
  }

  // Set up producers and consumers
  system.createProducer('Producer-A', () => dataGenerator('A', 10));
  system.createProducer('Producer-B', () => dataGenerator('B', 10));

  system.createConsumer('Consumer-1', processData);
  system.createConsumer('Consumer-2', processData);

  // Run system
  setTimeout(() => system.stop(), 5000);
  await system.start();
}
```

### Pipeline Pattern
```javascript
// Async pipeline for data processing
class AsyncPipeline {
  constructor() {
    this.stages = [];
  }

  addStage(name, processor, options = {}) {
    const stage = {
      name,
      processor,
      concurrency: options.concurrency || 1,
      queue: new AsyncQueue(options.bufferSize || 100),
      workers: []
    };

    // Create workers for this stage
    for (let i = 0; i < stage.concurrency; i++) {
      stage.workers.push(this.createWorker(stage, i));
    }

    this.stages.push(stage);
    return this;
  }

  createWorker(stage, workerId) {
    return async () => {
      console.log(`Worker ${stage.name}-${workerId} started`);

      while (true) {
        try {
          const data = await stage.queue.dequeue();

          if (data === null) break; // Poison pill

          const result = await stage.processor(data);

          // Send to next stage
          const nextStageIndex = this.stages.indexOf(stage) + 1;
          if (nextStageIndex < this.stages.length) {
            await this.stages[nextStageIndex].queue.enqueue(result);
          }

        } catch (error) {
          console.error(`Worker ${stage.name}-${workerId} error:`, error);
        }
      }

      console.log(`Worker ${stage.name}-${workerId} finished`);
    };\n  }\n  \n  async process(inputData) {\n    if (this.stages.length === 0) {\n      throw new Error('No stages defined');\n    }\n    \n    // Start all workers\n    const allWorkers = this.stages.flatMap(stage => \n      stage.workers.map(worker => worker())\n    );\n    \n    // Feed input data to first stage\n    const firstStage = this.stages[0];\n    for (const item of inputData) {\n      await firstStage.queue.enqueue(item);\n    }\n    \n    // Send poison pills to stop workers\n    setTimeout(async () => {\n      for (const stage of this.stages) {\n        for (let i = 0; i < stage.concurrency; i++) {\n          await stage.queue.enqueue(null);\n        }\n      }\n    }, 1000);\n    \n    await Promise.all(allWorkers);\n  }\n}\n\n// Usage example\nasync function demonstratePipeline() {\n  const pipeline = new AsyncPipeline()\n    .addStage('Parse', async (data) => {\n      await new Promise(resolve => setTimeout(resolve, 10));\n      return { ...data, parsed: true };\n    }, { concurrency: 2 })\n    .addStage('Transform', async (data) => {\n      await new Promise(resolve => setTimeout(resolve, 20));\n      return { ...data, transformed: true };\n    }, { concurrency: 3 })\n    .addStage('Validate', async (data) => {\n      await new Promise(resolve => setTimeout(resolve, 15));\n      return { ...data, validated: true };\n    }, { concurrency: 1 });\n  \n  const inputData = Array.from({ length: 50 }, (_, i) => ({ id: i, value: `item-${i}` }));\n  \n  await pipeline.process(inputData);\n}\n```\n\n## Web Workers and Parallelism\n\n### Web Workers for CPU-Intensive Tasks\n```javascript\n// Main thread: Worker manager\nclass WorkerPool {\n  constructor(workerScript, poolSize = navigator.hardwareConcurrency || 4) {\n    this.workerScript = workerScript;\n    this.poolSize = poolSize;\n    this.workers = [];\n    this.taskQueue = [];\n    this.availableWorkers = [];\n    \n    this.initializeWorkers();\n  }\n  \n  initializeWorkers() {\n    for (let i = 0; i < this.poolSize; i++) {\n      const worker = new Worker(this.workerScript);\n      worker.id = i;\n      \n      worker.onmessage = (event) => {\n        this.handleWorkerMessage(worker, event);\n      };\n      \n      worker.onerror = (error) => {\n        console.error(`Worker ${worker.id} error:`, error);\n      };\n      \n      this.workers.push(worker);\n      this.availableWorkers.push(worker);\n    }\n  }\n  \n  execute(task, transferables = []) {\n    return new Promise((resolve, reject) => {\n      const taskId = Date.now() + Math.random();\n      \n      this.taskQueue.push({\n        id: taskId,\n        task,\n        transferables,\n        resolve,\n        reject\n      });\n      \n      this.processQueue();\n    });\n  }\n  \n  processQueue() {\n    if (this.taskQueue.length === 0 || this.availableWorkers.length === 0) {\n      return;\n    }\n    \n    const task = this.taskQueue.shift();\n    const worker = this.availableWorkers.shift();\n    \n    worker.currentTask = task;\n    worker.postMessage(task.task, task.transferables);\n  }\n  \n  handleWorkerMessage(worker, event) {\n    const { currentTask } = worker;\n    \n    if (currentTask) {\n      if (event.data.error) {\n        currentTask.reject(new Error(event.data.error));\n      } else {\n        currentTask.resolve(event.data.result);\n      }\n      \n      worker.currentTask = null;\n      this.availableWorkers.push(worker);\n      \n      // Process next task in queue\n      this.processQueue();\n    }\n  }\n  \n  terminate() {\n    this.workers.forEach(worker => worker.terminate());\n    this.workers = [];\n    this.availableWorkers = [];\n  }\n}\n\n// Worker script (cpu-intensive-worker.js)\n/*\nself.onmessage = function(event) {\n  const { type, data } = event.data;\n  \n  try {\n    let result;\n    \n    switch (type) {\n      case 'fibonacci':\n        result = calculateFibonacci(data.n);\n        break;\n      case 'primes':\n        result = findPrimes(data.max);\n        break;\n      case 'matrixMultiply':\n        result = multiplyMatrices(data.a, data.b);\n        break;\n      default:\n        throw new Error('Unknown task type');\n    }\n    \n    self.postMessage({ result });\n  } catch (error) {\n    self.postMessage({ error: error.message });\n  }\n};\n\nfunction calculateFibonacci(n) {\n  if (n <= 1) return n;\n  let a = 0, b = 1;\n  for (let i = 2; i <= n; i++) {\n    [a, b] = [b, a + b];\n  }\n  return b;\n}\n\nfunction findPrimes(max) {\n  const primes = [];\n  const isPrime = new Array(max + 1).fill(true);\n  isPrime[0] = isPrime[1] = false;\n  \n  for (let i = 2; i <= max; i++) {\n    if (isPrime[i]) {\n      primes.push(i);\n      for (let j = i * i; j <= max; j += i) {\n        isPrime[j] = false;\n      }\n    }\n  }\n  \n  return primes;\n}\n\nfunction multiplyMatrices(a, b) {\n  const result = [];\n  for (let i = 0; i < a.length; i++) {\n    result[i] = [];\n    for (let j = 0; j < b[0].length; j++) {\n      result[i][j] = 0;\n      for (let k = 0; k < b.length; k++) {\n        result[i][j] += a[i][k] * b[k][j];\n      }\n    }\n  }\n  return result;\n}\n*/\n\n// Usage in main thread\nasync function demonstrateWorkerPool() {\n  const workerPool = new WorkerPool('cpu-intensive-worker.js', 4);\n  \n  // Parallel computation\n  const tasks = [\n    workerPool.execute({ type: 'fibonacci', data: { n: 40 } }),\n    workerPool.execute({ type: 'primes', data: { max: 100000 } }),\n    workerPool.execute({ type: 'fibonacci', data: { n: 35 } }),\n    workerPool.execute({ type: 'primes', data: { max: 50000 } })\n  ];\n  \n  try {\n    const results = await Promise.all(tasks);\n    console.log('All tasks completed:', results);\n  } catch (error) {\n    console.error('Task failed:', error);\n  } finally {\n    workerPool.terminate();\n  }\n}\n```\n\n### SharedArrayBuffer for Shared Memory\n```javascript\n// Shared memory between main thread and workers\nclass SharedMemoryManager {\n  constructor(size) {\n    // Note: SharedArrayBuffer requires cross-origin isolation\n    this.buffer = new SharedArrayBuffer(size);\n    this.int32Array = new Int32Array(this.buffer);\n    this.float64Array = new Float64Array(this.buffer);\n  }\n  \n  createWorkerWithSharedMemory(workerScript) {\n    const worker = new Worker(workerScript);\n    \n    // Transfer shared buffer to worker\n    worker.postMessage({\n      type: 'init',\n      sharedBuffer: this.buffer\n    });\n    \n    return worker;\n  }\n  \n  // Atomic operations for thread-safe access\n  atomicIncrement(index) {\n    return Atomics.add(this.int32Array, index, 1);\n  }\n  \n  atomicCompareAndSwap(index, expected, replacement) {\n    return Atomics.compareExchange(this.int32Array, index, expected, replacement);\n  }\n  \n  waitForWorkers(index, value, timeout = Infinity) {\n    return Atomics.wait(this.int32Array, index, value, timeout);\n  }\n  \n  notifyWorkers(index, count = Infinity) {\n    return Atomics.notify(this.int32Array, index, count);\n  }\n}\n\n// Shared memory worker script\n/*\nlet sharedBuffer;\nlet sharedInt32Array;\nlet sharedFloat64Array;\n\nself.onmessage = function(event) {\n  const { type, data } = event.data;\n  \n  switch (type) {\n    case 'init':\n      sharedBuffer = event.data.sharedBuffer;\n      sharedInt32Array = new Int32Array(sharedBuffer);\n      sharedFloat64Array = new Float64Array(sharedBuffer);\n      self.postMessage({ type: 'ready' });\n      break;\n      \n    case 'compute':\n      // Perform computation using shared memory\n      const { startIndex, endIndex } = data;\n      \n      for (let i = startIndex; i < endIndex; i++) {\n        // Atomic operations on shared memory\n        const currentValue = Atomics.load(sharedInt32Array, i);\n        const newValue = currentValue * 2;\n        Atomics.store(sharedInt32Array, i, newValue);\n      }\n      \n      // Notify completion\n      Atomics.add(sharedInt32Array, 0, 1); // Increment counter\n      Atomics.notify(sharedInt32Array, 0); // Wake waiting threads\n      \n      self.postMessage({ type: 'completed', data: { startIndex, endIndex } });\n      break;\n  }\n};\n*/\n```\n\n## Node.js Concurrency\n\n### Worker Threads\n```javascript\n// Node.js Worker Threads\nconst { Worker, isMainThread, parentPort, workerData } = require('worker_threads');\nconst os = require('os');\n\nif (isMainThread) {\n  // Main thread: Worker pool implementation\n  class NodeWorkerPool {\n    constructor(workerScript, poolSize = os.cpus().length) {\n      this.workerScript = workerScript;\n      this.poolSize = poolSize;\n      this.workers = [];\n      this.taskQueue = [];\n      this.availableWorkers = [];\n      \n      this.initializeWorkers();\n    }\n    \n    initializeWorkers() {\n      for (let i = 0; i < this.poolSize; i++) {\n        const worker = new Worker(this.workerScript);\n        worker.id = i;\n        \n        worker.on('message', (result) => {\n          this.handleWorkerMessage(worker, result);\n        });\n        \n        worker.on('error', (error) => {\n          console.error(`Worker ${worker.id} error:`, error);\n        });\n        \n        worker.on('exit', (code) => {\n          if (code !== 0) {\n            console.error(`Worker ${worker.id} exited with code ${code}`);\n          }\n        });\n        \n        this.workers.push(worker);\n        this.availableWorkers.push(worker);\n      }\n    }\n    \n    execute(task) {\n      return new Promise((resolve, reject) => {\n        this.taskQueue.push({\n          task,\n          resolve,\n          reject\n        });\n        \n        this.processQueue();\n      });\n    }\n    \n    processQueue() {\n      if (this.taskQueue.length === 0 || this.availableWorkers.length === 0) {\n        return;\n      }\n      \n      const { task, resolve, reject } = this.taskQueue.shift();\n      const worker = this.availableWorkers.shift();\n      \n      worker.currentResolve = resolve;\n      worker.currentReject = reject;\n      worker.postMessage(task);\n    }\n    \n    handleWorkerMessage(worker, result) {\n      if (worker.currentResolve) {\n        if (result.error) {\n          worker.currentReject(new Error(result.error));\n        } else {\n          worker.currentResolve(result.data);\n        }\n        \n        worker.currentResolve = null;\n        worker.currentReject = null;\n        this.availableWorkers.push(worker);\n        \n        this.processQueue();\n      }\n    }\n    \n    async terminate() {\n      await Promise.all(this.workers.map(worker => worker.terminate()));\n    }\n  }\n  \n  module.exports = { NodeWorkerPool };\n  \n} else {\n  // Worker thread implementation\n  parentPort.on('message', async (task) => {\n    try {\n      let result;\n      \n      switch (task.type) {\n        case 'cpu-intensive':\n          result = await performCPUIntensiveTask(task.data);\n          break;\n        case 'file-processing':\n          result = await processFile(task.data);\n          break;\n        default:\n          throw new Error(`Unknown task type: ${task.type}`);\n      }\n      \n      parentPort.postMessage({ data: result });\n    } catch (error) {\n      parentPort.postMessage({ error: error.message });\n    }\n  });\n  \n  async function performCPUIntensiveTask(data) {\n    // Simulate CPU-intensive work\n    const { iterations } = data;\n    let result = 0;\n    \n    for (let i = 0; i < iterations; i++) {\n      result += Math.sqrt(i);\n    }\n    \n    return result;\n  }\n  \n  async function processFile(data) {\n    const fs = require('fs').promises;\n    const path = require('path');\n    \n    const { filePath } = data;\n    const content = await fs.readFile(filePath, 'utf8');\n    \n    return {\n      filePath,\n      lineCount: content.split('\\n').length,\n      wordCount: content.split(/\\s+/).length,\n      charCount: content.length\n    };\n  }\n}\n```\n\n### Cluster Module for Scaling\n```javascript\n// Node.js cluster for multi-process scaling\nconst cluster = require('cluster');\nconst http = require('http');\nconst os = require('os');\n\nif (cluster.isMaster) {\n  console.log(`Master ${process.pid} is running`);\n  \n  const numCPUs = os.cpus().length;\n  \n  // Fork workers\n  for (let i = 0; i < numCPUs; i++) {\n    const worker = cluster.fork();\n    \n    // Handle worker messages\n    worker.on('message', (message) => {\n      console.log(`Master received message from worker ${worker.process.pid}:`, message);\n    });\n  }\n  \n  // Handle worker events\n  cluster.on('exit', (worker, code, signal) => {\n    console.log(`Worker ${worker.process.pid} died`);\n    console.log('Starting a new worker');\n    cluster.fork();\n  });\n  \n  cluster.on('online', (worker) => {\n    console.log(`Worker ${worker.process.pid} is online`);\n  });\n  \n  // Graceful shutdown\n  process.on('SIGTERM', () => {\n    console.log('Master received SIGTERM, shutting down gracefully');\n    \n    for (const worker of Object.values(cluster.workers)) {\n      worker.send('shutdown');\n    }\n  });\n  \n} else {\n  // Worker process\n  const server = http.createServer((req, res) => {\n    // CPU-intensive task simulation\n    const start = Date.now();\n    while (Date.now() - start < 100) {\n      // Busy wait\n    }\n    \n    res.writeHead(200);\n    res.end(`Hello from worker ${process.pid}\\n`);\n  });\n  \n  server.listen(8000, () => {\n    console.log(`Worker ${process.pid} started`);\n  });\n  \n  // Handle master messages\n  process.on('message', (message) => {\n    if (message === 'shutdown') {\n      server.close(() => {\n        process.exit(0);\n      });\n    }\n  });\n  \n  // Send heartbeat to master\n  setInterval(() => {\n    process.send({ type: 'heartbeat', pid: process.pid });\n  }, 5000);\n}\n```\n\n## Scaling Patterns\n\n### Load Balancing and Request Distribution\n```javascript\n// Advanced load balancer with health checking\nclass LoadBalancer {\n  constructor() {\n    this.backends = [];\n    this.healthCheckInterval = 5000;\n    this.healthCheckTimeout = 2000;\n    this.currentIndex = 0;\n    \n    this.startHealthChecks();\n  }\n  \n  addBackend(url, weight = 1) {\n    const backend = {\n      url,\n      weight,\n      healthy: true,\n      totalRequests: 0,\n      failedRequests: 0,\n      avgResponseTime: 0,\n      lastCheck: null\n    };\n    \n    this.backends.push(backend);\n    return this;\n  }\n  \n  // Round-robin with health checking\n  getHealthyBackend() {\n    const healthyBackends = this.backends.filter(b => b.healthy);\n    \n    if (healthyBackends.length === 0) {\n      throw new Error('No healthy backends available');\n    }\n    \n    const backend = healthyBackends[this.currentIndex % healthyBackends.length];\n    this.currentIndex++;\n    \n    return backend;\n  }\n  \n  // Weighted round-robin\n  getWeightedBackend() {\n    const healthyBackends = this.backends.filter(b => b.healthy);\n    const totalWeight = healthyBackends.reduce((sum, b) => sum + b.weight, 0);\n    \n    let random = Math.random() * totalWeight;\n    \n    for (const backend of healthyBackends) {\n      random -= backend.weight;\n      if (random <= 0) {\n        return backend;\n      }\n    }\n    \n    return healthyBackends[0];\n  }\n  \n  // Least connections\n  getLeastConnectionsBackend() {\n    const healthyBackends = this.backends.filter(b => b.healthy);\n    \n    return healthyBackends.reduce((least, current) => \n      current.totalRequests < least.totalRequests ? current : least\n    );\n  }\n  \n  async makeRequest(path, options = {}) {\n    const backend = this.getHealthyBackend();\n    const startTime = Date.now();\n    \n    try {\n      backend.totalRequests++;\n      \n      const response = await fetch(`${backend.url}${path}`, {\n        ...options,\n        timeout: 10000\n      });\n      \n      const responseTime = Date.now() - startTime;\n      backend.avgResponseTime = \n        (backend.avgResponseTime + responseTime) / 2;\n      \n      return response;\n      \n    } catch (error) {\n      backend.failedRequests++;\n      \n      // Mark as unhealthy if error rate is high\n      const errorRate = backend.failedRequests / backend.totalRequests;\n      if (errorRate > 0.1) { // 10% error rate threshold\n        backend.healthy = false;\n      }\n      \n      throw error;\n    }\n  }\n  \n  async healthCheck(backend) {\n    const controller = new AbortController();\n    const timeout = setTimeout(() => controller.abort(), this.healthCheckTimeout);\n    \n    try {\n      const response = await fetch(`${backend.url}/health`, {\n        signal: controller.signal,\n        method: 'GET'\n      });\n      \n      backend.healthy = response.ok;\n      backend.lastCheck = Date.now();\n      \n    } catch (error) {\n      backend.healthy = false;\n      backend.lastCheck = Date.now();\n      \n    } finally {\n      clearTimeout(timeout);\n    }\n  }\n  \n  startHealthChecks() {\n    setInterval(async () => {\n      const healthChecks = this.backends.map(backend => \n        this.healthCheck(backend)\n      );\n      \n      await Promise.allSettled(healthChecks);\n      \n      console.log('Health check completed:', \n        this.backends.map(b => ({ url: b.url, healthy: b.healthy }))\n      );\n    }, this.healthCheckInterval);\n  }\n  \n  getStats() {\n    return this.backends.map(backend => ({\n      url: backend.url,\n      healthy: backend.healthy,\n      totalRequests: backend.totalRequests,\n      failedRequests: backend.failedRequests,\n      errorRate: (backend.failedRequests / backend.totalRequests * 100).toFixed(2) + '%',\n      avgResponseTime: Math.round(backend.avgResponseTime) + 'ms'\n    }));\n  }\n}\n\n// Usage\nconst loadBalancer = new LoadBalancer()\n  .addBackend('http://server1:8080', 2)\n  .addBackend('http://server2:8080', 1)\n  .addBackend('http://server3:8080', 3);\n\n// Handle requests\nasync function handleRequest(req, res) {\n  try {\n    const response = await loadBalancer.makeRequest(req.url, {\n      method: req.method,\n      headers: req.headers,\n      body: req.body\n    });\n    \n    res.status(response.status);\n    response.body.pipe(res);\n    \n  } catch (error) {\n    res.status(500).json({ error: 'Service unavailable' });\n  }\n}\n```\n\n### Horizontal Scaling with Message Queues\n```javascript\n// Message queue implementation for scaling\nclass MessageQueue {\n  constructor(options = {}) {\n    this.queues = new Map();\n    this.subscribers = new Map();\n    this.dlq = new Map(); // Dead letter queue\n    this.maxRetries = options.maxRetries || 3;\n    this.retryDelay = options.retryDelay || 1000;\n  }\n  \n  async publish(queueName, message, options = {}) {\n    if (!this.queues.has(queueName)) {\n      this.queues.set(queueName, []);\n    }\n    \n    const enrichedMessage = {\n      id: Date.now() + Math.random(),\n      data: message,\n      timestamp: Date.now(),\n      retries: 0,\n      maxRetries: options.maxRetries || this.maxRetries,\n      delay: options.delay || 0,\n      priority: options.priority || 0\n    };\n    \n    const queue = this.queues.get(queueName);\n    \n    if (enrichedMessage.delay > 0) {\n      setTimeout(() => {\n        this.insertByPriority(queue, enrichedMessage);\n        this.notifySubscribers(queueName);\n      }, enrichedMessage.delay);\n    } else {\n      this.insertByPriority(queue, enrichedMessage);\n      this.notifySubscribers(queueName);\n    }\n  }\n  \n  insertByPriority(queue, message) {\n    let inserted = false;\n    for (let i = 0; i < queue.length; i++) {\n      if (message.priority > queue[i].priority) {\n        queue.splice(i, 0, message);\n        inserted = true;\n        break;\n      }\n    }\n    \n    if (!inserted) {\n      queue.push(message);\n    }\n  }\n  \n  subscribe(queueName, handler, options = {}) {\n    const subscription = {\n      handler,\n      concurrency: options.concurrency || 1,\n      activeJobs: 0,\n      autoAck: options.autoAck !== false\n    };\n    \n    if (!this.subscribers.has(queueName)) {\n      this.subscribers.set(queueName, []);\n    }\n    \n    this.subscribers.get(queueName).push(subscription);\n    \n    // Start processing\n    this.processQueue(queueName);\n  }\n  \n  async processQueue(queueName) {\n    const queue = this.queues.get(queueName) || [];\n    const subscribers = this.subscribers.get(queueName) || [];\n    \n    while (queue.length > 0 && subscribers.length > 0) {\n      // Find available subscriber\n      const availableSubscriber = subscribers.find(\n        sub => sub.activeJobs < sub.concurrency\n      );\n      \n      if (!availableSubscriber) {\n        break;\n      }\n      \n      const message = queue.shift();\n      availableSubscriber.activeJobs++;\n      \n      // Process message\n      this.processMessage(queueName, message, availableSubscriber)\n        .finally(() => {\n          availableSubscriber.activeJobs--;\n          // Continue processing\n          setImmediate(() => this.processQueue(queueName));\n        });\n    }\n  }\n  \n  async processMessage(queueName, message, subscriber) {\n    try {\n      await subscriber.handler(message.data, {\n        ack: () => this.ack(message),\n        nack: (requeue = true) => this.nack(queueName, message, requeue),\n        retry: () => this.retry(queueName, message)\n      });\n      \n      if (subscriber.autoAck) {\n        this.ack(message);\n      }\n      \n    } catch (error) {\n      console.error(`Message processing failed:`, error);\n      await this.retry(queueName, message);\n    }\n  }\n  \n  ack(message) {\n    // Message acknowledged - nothing more to do\n    console.log(`Message ${message.id} acknowledged`);\n  }\n  \n  nack(queueName, message, requeue = true) {\n    if (requeue) {\n      // Put message back in queue\n      const queue = this.queues.get(queueName);\n      this.insertByPriority(queue, message);\n    } else {\n      // Send to DLQ\n      this.sendToDLQ(queueName, message);\n    }\n  }\n  \n  async retry(queueName, message) {\n    message.retries++;\n    \n    if (message.retries <= message.maxRetries) {\n      // Exponential backoff\n      const delay = this.retryDelay * Math.pow(2, message.retries - 1);\n      \n      setTimeout(() => {\n        const queue = this.queues.get(queueName);\n        this.insertByPriority(queue, message);\n        this.notifySubscribers(queueName);\n      }, delay);\n      \n    } else {\n      this.sendToDLQ(queueName, message);\n    }\n  }\n  \n  sendToDLQ(queueName, message) {\n    const dlqName = `${queueName}.dlq`;\n    \n    if (!this.dlq.has(dlqName)) {\n      this.dlq.set(dlqName, []);\n    }\n    \n    this.dlq.get(dlqName).push({\n      ...message,\n      dlqTimestamp: Date.now(),\n      originalQueue: queueName\n    });\n    \n    console.log(`Message ${message.id} sent to DLQ: ${dlqName}`);\n  }\n  \n  notifySubscribers(queueName) {\n    setImmediate(() => this.processQueue(queueName));\n  }\n  \n  getStats() {\n    const stats = {};\n    \n    for (const [queueName, queue] of this.queues) {\n      stats[queueName] = {\n        pending: queue.length,\n        subscribers: this.subscribers.get(queueName)?.length || 0\n      };\n    }\n    \n    for (const [dlqName, messages] of this.dlq) {\n      stats[dlqName] = {\n        messages: messages.length\n      };\n    }\n    \n    return stats;\n  }\n}\n\n// Usage for horizontal scaling\nconst messageQueue = new MessageQueue();\n\n// Producer\nasync function processOrder(order) {\n  await messageQueue.publish('orders', order, {\n    priority: order.priority || 0,\n    maxRetries: 3\n  });\n}\n\n// Consumer 1\nmessageQueue.subscribe('orders', async (order, { ack, nack, retry }) => {\n  try {\n    console.log(`Processing order ${order.id} on worker 1`);\n    \n    // Simulate processing\n    await new Promise(resolve => setTimeout(resolve, Math.random() * 1000));\n    \n    if (Math.random() < 0.1) {\n      throw new Error('Random failure');\n    }\n    \n    console.log(`Order ${order.id} completed on worker 1`);\n    ack();\n    \n  } catch (error) {\n    console.error(`Order ${order.id} failed on worker 1:`, error.message);\n    retry();\n  }\n}, { concurrency: 3 });\n\n// Consumer 2\nmessageQueue.subscribe('orders', async (order, { ack, nack }) => {\n  console.log(`Processing order ${order.id} on worker 2`);\n  \n  // Different processing logic\n  await new Promise(resolve => setTimeout(resolve, Math.random() * 500));\n  \n  console.log(`Order ${order.id} completed on worker 2`);\n  ack();\n}, { concurrency: 2 });\n```\n\n## Performance Optimization\n\n### Async Performance Monitoring\n```javascript\n// Performance monitoring for async operations\nclass AsyncPerformanceMonitor {\n  constructor() {\n    this.metrics = new Map();\n    this.observers = [];\n  }\n  \n  // Decorator for monitoring async functions\n  monitor(name, options = {}) {\n    return (target, propertyKey, descriptor) => {\n      const originalMethod = descriptor.value;\n      \n      descriptor.value = async function(...args) {\n        return this.measureAsync(name, () => originalMethod.apply(this, args), options);\n      }.bind(this);\n      \n      return descriptor;\n    };\n  }\n  \n  async measureAsync(operationName, asyncOperation, options = {}) {\n    const startTime = performance.now();\n    const startMemory = this.getMemoryUsage();\n    \n    try {\n      const result = await asyncOperation();\n      \n      const duration = performance.now() - startTime;\n      const memoryDelta = this.getMemoryUsage() - startMemory;\n      \n      this.recordMetric(operationName, {\n        duration,\n        memoryDelta,\n        success: true,\n        timestamp: Date.now()\n      });\n      \n      return result;\n      \n    } catch (error) {\n      const duration = performance.now() - startTime;\n      \n      this.recordMetric(operationName, {\n        duration,\n        memoryDelta: this.getMemoryUsage() - startMemory,\n        success: false,\n        error: error.message,\n        timestamp: Date.now()\n      });\n      \n      throw error;\n    }\n  }\n  \n  recordMetric(operationName, metric) {\n    if (!this.metrics.has(operationName)) {\n      this.metrics.set(operationName, {\n        count: 0,\n        totalDuration: 0,\n        avgDuration: 0,\n        minDuration: Infinity,\n        maxDuration: 0,\n        successRate: 0,\n        errors: []\n      });\n    }\n    \n    const stats = this.metrics.get(operationName);\n    stats.count++;\n    stats.totalDuration += metric.duration;\n    stats.avgDuration = stats.totalDuration / stats.count;\n    stats.minDuration = Math.min(stats.minDuration, metric.duration);\n    stats.maxDuration = Math.max(stats.maxDuration, metric.duration);\n    \n    if (metric.success) {\n      stats.successRate = (stats.successRate * (stats.count - 1) + 1) / stats.count;\n    } else {\n      stats.successRate = stats.successRate * (stats.count - 1) / stats.count;\n      stats.errors.push({\n        error: metric.error,\n        timestamp: metric.timestamp\n      });\n    }\n    \n    // Notify observers\n    this.observers.forEach(observer => {\n      observer(operationName, metric, stats);\n    });\n  }\n  \n  getMemoryUsage() {\n    if (typeof performance !== 'undefined' && performance.memory) {\n      return performance.memory.usedJSHeapSize;\n    }\n    return 0;\n  }\n  \n  getStats(operationName) {\n    return this.metrics.get(operationName);\n  }\n  \n  getAllStats() {\n    const result = {};\n    for (const [name, stats] of this.metrics) {\n      result[name] = { ...stats };\n    }\n    return result;\n  }\n  \n  onMetric(callback) {\n    this.observers.push(callback);\n  }\n  \n  reset() {\n    this.metrics.clear();\n  }\n}\n\n// Usage\nconst monitor = new AsyncPerformanceMonitor();\n\n// Monitor specific operations\nclass DataService {\n  @monitor.monitor('fetchUserData', { threshold: 1000 })\n  async fetchUserData(userId) {\n    const response = await fetch(`/api/users/${userId}`);\n    return response.json();\n  }\n  \n  @monitor.monitor('processLargeDataset')\n  async processLargeDataset(data) {\n    // Simulate processing\n    return data.map(item => ({ ...item, processed: true }));\n  }\n}\n\n// Set up monitoring alerts\nmonitor.onMetric((operationName, metric, stats) => {\n  if (metric.duration > 2000) {\n    console.warn(`Slow operation detected: ${operationName} took ${metric.duration.toFixed(2)}ms`);\n  }\n  \n  if (stats.successRate < 0.95) {\n    console.error(`High error rate for ${operationName}: ${(stats.successRate * 100).toFixed(2)}%`);\n  }\n});\n```\n\n### Batch Processing and Rate Limiting\n```javascript\n// Advanced rate limiting and batch processing\nclass RateLimitedBatchProcessor {\n  constructor(options = {}) {\n    this.maxBatchSize = options.maxBatchSize || 10;\n    this.maxConcurrency = options.maxConcurrency || 3;\n    this.batchTimeout = options.batchTimeout || 1000;\n    this.rateLimit = options.rateLimit || { requests: 100, window: 60000 };\n    \n    this.pendingBatch = [];\n    this.activeBatches = 0;\n    this.requestWindow = [];\n    this.batchTimer = null;\n  }\n  \n  async process(item) {\n    // Check rate limit\n    await this.checkRateLimit();\n    \n    return new Promise((resolve, reject) => {\n      this.pendingBatch.push({ item, resolve, reject });\n      \n      // Start batch timer if this is the first item\n      if (this.pendingBatch.length === 1) {\n        this.startBatchTimer();\n      }\n      \n      // Process immediately if batch is full\n      if (this.pendingBatch.length >= this.maxBatchSize) {\n        this.processBatch();\n      }\n    });\n  }\n  \n  async checkRateLimit() {\n    const now = Date.now();\n    const windowStart = now - this.rateLimit.window;\n    \n    // Remove old requests\n    this.requestWindow = this.requestWindow.filter(time => time > windowStart);\n    \n    if (this.requestWindow.length >= this.rateLimit.requests) {\n      const oldestRequest = Math.min(...this.requestWindow);\n      const waitTime = oldestRequest + this.rateLimit.window - now;\n      \n      if (waitTime > 0) {\n        await new Promise(resolve => setTimeout(resolve, waitTime));\n      }\n    }\n    \n    this.requestWindow.push(now);\n  }\n  \n  startBatchTimer() {\n    if (this.batchTimer) return;\n    \n    this.batchTimer = setTimeout(() => {\n      if (this.pendingBatch.length > 0) {\n        this.processBatch();\n      }\n    }, this.batchTimeout);\n  }\n  \n  async processBatch() {\n    if (this.pendingBatch.length === 0) return;\n    if (this.activeBatches >= this.maxConcurrency) {\n      return; // Will be processed when a slot opens up\n    }\n    \n    const batch = this.pendingBatch.splice(0, this.maxBatchSize);\n    this.activeBatches++;\n    \n    if (this.batchTimer) {\n      clearTimeout(this.batchTimer);\n      this.batchTimer = null;\n    }\n    \n    try {\n      const items = batch.map(b => b.item);\n      const results = await this.processBatchItems(items);\n      \n      // Resolve individual promises\n      batch.forEach((b, index) => {\n        b.resolve(results[index]);\n      });\n      \n    } catch (error) {\n      // Reject all promises in the batch\n      batch.forEach(b => b.reject(error));\n      \n    } finally {\n      this.activeBatches--;\n      \n      // Process next batch if needed\n      if (this.pendingBatch.length > 0) {\n        setImmediate(() => this.processBatch());\n      }\n    }\n  }\n  \n  async processBatchItems(items) {\n    // Override this method with actual batch processing logic\n    console.log(`Processing batch of ${items.length} items`);\n    \n    // Simulate batch processing\n    await new Promise(resolve => setTimeout(resolve, 100));\n    \n    return items.map(item => ({ ...item, processed: true, batchedAt: Date.now() }));\n  }\n  \n  getStats() {\n    return {\n      pendingItems: this.pendingBatch.length,\n      activeBatches: this.activeBatches,\n      requestsInWindow: this.requestWindow.length,\n      windowStartsAt: new Date(Date.now() - this.rateLimit.window).toISOString()\n    };\n  }\n}\n\n// Usage\nclass APIBatchProcessor extends RateLimitedBatchProcessor {\n  async processBatchItems(items) {\n    // Actual API call with batch of items\n    const response = await fetch('/api/batch', {\n      method: 'POST',\n      headers: { 'Content-Type': 'application/json' },\n      body: JSON.stringify({ items })\n    });\n    \n    if (!response.ok) {\n      throw new Error(`Batch API failed: ${response.statusText}`);\n    }\n    \n    return response.json();\n  }\n}\n\nconst batchProcessor = new APIBatchProcessor({\n  maxBatchSize: 50,\n  maxConcurrency: 3,\n  batchTimeout: 2000,\n  rateLimit: { requests: 1000, window: 60000 }\n});\n\n// Process individual items - they'll be automatically batched\nasync function processItems(items) {\n  const results = await Promise.all(\n    items.map(item => batchProcessor.process(item))\n  );\n  \n  return results;\n}\n```\n\nThis comprehensive guide covers the essential aspects of concurrency, parallelism, and scaling in TypeScript/JavaScript. Understanding these concepts and patterns is crucial for building efficient, scalable applications that can handle high loads and complex asynchronous operations.