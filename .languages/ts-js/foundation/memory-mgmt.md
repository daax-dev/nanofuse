# Memory Management in TypeScript/JavaScript

## Overview

JavaScript uses automatic memory management through garbage collection, but understanding how memory works is crucial for writing efficient applications. This guide covers memory allocation, garbage collection, common memory issues, and optimization strategies.

## JavaScript Memory Model

### Memory Allocation

#### Primitive Values (Stack)
```javascript
// Primitives are stored on the stack
let number = 42;          // Number: 8 bytes
let boolean = true;       // Boolean: 4 bytes  
let string = 'hello';     // String: reference to heap
let symbol = Symbol('id'); // Symbol: reference to heap
let bigint = 123n;        // BigInt: reference to heap
let undefined_val;        // undefined: special value
let null_val = null;      // null: special value
```

#### Reference Values (Heap)
```javascript
// Objects, arrays, functions stored on heap
const obj = { name: 'John' };      // Object on heap
const arr = [1, 2, 3];            // Array on heap
const func = () => console.log();  // Function on heap
const date = new Date();           // Date object on heap
const regex = /pattern/;           // RegExp object on heap
```

### Memory Lifecycle

1. **Allocation**: Memory is allocated when variables are declared
2. **Usage**: Memory is read from and written to during execution
3. **Deallocation**: Memory is freed when no longer accessible

```javascript
function memoryLifecycle() {
  // 1. Allocation - memory allocated for variables
  let localVar = 'temporary';
  let obj = { data: new Array(1000).fill(0) };
  
  // 2. Usage - memory is used
  console.log(localVar);
  obj.data.push(42);
  
  // 3. Deallocation - happens automatically when function exits
  // localVar and obj become eligible for garbage collection
}

memoryLifecycle();
// Variables are now out of scope and can be garbage collected
```

## Garbage Collection

### Mark and Sweep Algorithm

```javascript
// The garbage collector marks reachable objects
function demonstrateReachability() {
  let root = {
    child: {
      grandchild: { value: 42 }
    }
  };
  
  let orphan = { value: 'unreachable' };
  
  // root and its children are reachable from global scope
  // orphan becomes unreachable when function exits
  
  return root; // root remains reachable
}

let globalRef = demonstrateReachability();
// Only objects reachable from globalRef survive garbage collection
```

### Generational Garbage Collection

```javascript
// V8 uses generational GC - objects are categorized by age

// Young generation (short-lived objects)
function createShortLived() {
  let temp = new Array(1000); // Likely collected quickly
  return temp.length;
}

// Old generation (long-lived objects)
const longLived = {
  cache: new Map(),
  config: { setting: 'value' }
}; // Survives multiple GC cycles, moved to old generation

// Frequent allocations in young generation
for (let i = 0; i < 1000; i++) {
  createShortLived(); // These objects are collected frequently
}
```

## Memory Leaks and Prevention

### Common Memory Leak Patterns

#### 1. Global Variables
```javascript
// BAD: Accidental globals
function createLeak() {
  // Forgot 'let/const/var' - creates global variable
  leakedVariable = 'This creates a global!';
  
  // 'this' refers to global object in non-strict mode
  this.anotherLeak = 'Another global variable';
}

// GOOD: Proper variable declarations
function preventLeak() {
  'use strict'; // Prevents accidental globals
  const localVariable = 'This is local';
  
  // Explicitly create global if needed
  if (typeof window !== 'undefined') {
    window.intentionalGlobal = 'Intentional';
  }
}
```

#### 2. Event Listeners
```javascript
// BAD: Event listeners not removed
class LeakyComponent {
  constructor() {
    this.data = new Array(1000).fill('data');
    
    // Event listener holds reference to 'this'
    document.addEventListener('click', this.handleClick.bind(this));
  }
  
  handleClick() {
    console.log(this.data.length);
  }
  
  // Component destroyed but event listener remains
}

// GOOD: Clean up event listeners
class CleanComponent {
  constructor() {
    this.data = new Array(1000).fill('data');
    this.boundHandleClick = this.handleClick.bind(this);
    
    document.addEventListener('click', this.boundHandleClick);
  }
  
  handleClick() {
    console.log(this.data.length);
  }
  
  destroy() {
    // Remove event listener to prevent memory leak
    document.removeEventListener('click', this.boundHandleClick);
    this.data = null;
  }
}
```

#### 3. Closures and Circular References
```javascript
// BAD: Closure retains unnecessary references
function createLeakyClosure() {
  const largeData = new Array(1000000).fill('data');
  const smallData = 'small';
  
  // This closure captures entire scope, including largeData
  return function() {
    return smallData; // Only needs smallData
  };
}

// GOOD: Minimize closure scope
function createCleanClosure() {
  const largeData = new Array(1000000).fill('data');
  const smallData = 'small';
  
  // Process large data and discard
  processLargeData(largeData);
  
  // Return closure that only captures what it needs
  return (function(data) {
    return function() {
      return data;
    };
  })(smallData);
}

// BAD: Circular references
function createCircularReference() {
  const parent = { name: 'parent' };
  const child = { name: 'child', parent: parent };
  parent.child = child; // Creates circular reference
  
  // Modern JS engines handle this, but can be problematic
  return parent;
}

// GOOD: Use WeakMap for relationships
const parentChildMap = new WeakMap();

function createCleanReference() {
  const parent = { name: 'parent' };
  const child = { name: 'child' };
  
  // Use WeakMap to avoid circular reference
  parentChildMap.set(parent, child);
  
  return parent;
}
```

#### 4. Timers and Intervals
```javascript
// BAD: Timers not cleared
class TimerComponent {
  constructor() {
    this.data = new Array(1000).fill('data');
    
    // Timer holds reference to 'this'
    this.intervalId = setInterval(() => {
      console.log(this.data.length);
    }, 1000);
  }
  
  // If destroy() is never called, timer keeps running
  destroy() {
    clearInterval(this.intervalId);
  }
}

// GOOD: Automatic timer cleanup
class SafeTimerComponent {
  constructor() {
    this.data = new Array(1000).fill('data');
    this.timers = new Set();
    
    this.startTimer();
  }
  
  startTimer() {
    const intervalId = setInterval(() => {
      console.log(this.data.length);
    }, 1000);
    
    this.timers.add(intervalId);
  }
  
  destroy() {
    // Clean up all timers
    this.timers.forEach(timerId => clearInterval(timerId));
    this.timers.clear();
    this.data = null;
  }
}
```

### Memory Leak Detection

```javascript
// Memory monitoring utilities
class MemoryMonitor {
  static getMemoryUsage() {
    if (performance.memory) {
      return {
        used: Math.round(performance.memory.usedJSHeapSize / 1048576) + 'MB',
        total: Math.round(performance.memory.totalJSHeapSize / 1048576) + 'MB',
        limit: Math.round(performance.memory.jsHeapSizeLimit / 1048576) + 'MB'
      };
    }
    return null;
  }
  
  static startMonitoring(interval = 5000) {
    const monitor = setInterval(() => {
      const memory = MemoryMonitor.getMemoryUsage();
      if (memory) {
        console.log('Memory Usage:', memory);
        
        // Alert if memory usage is high
        const usedMB = parseInt(memory.used);
        const limitMB = parseInt(memory.limit);
        
        if (usedMB > limitMB * 0.8) {
          console.warn('High memory usage detected!');
        }
      }
    }, interval);
    
    return monitor;
  }
}

// Usage
const memoryMonitor = MemoryMonitor.startMonitoring();

// Later...
clearInterval(memoryMonitor);
```

## Memory Optimization Strategies

### Object Pooling

```javascript
// Object pool to reduce garbage collection
class ObjectPool {
  constructor(createFn, resetFn, maxSize = 100) {
    this.createFn = createFn;
    this.resetFn = resetFn;
    this.pool = [];
    this.maxSize = maxSize;
  }
  
  acquire() {
    if (this.pool.length > 0) {
      return this.pool.pop();
    }
    return this.createFn();
  }
  
  release(obj) {
    if (this.pool.length < this.maxSize) {
      this.resetFn(obj);
      this.pool.push(obj);
    }
  }
}

// Example: Vector pool for game development
class Vector {
  constructor(x = 0, y = 0) {
    this.x = x;
    this.y = y;
  }
  
  reset(x = 0, y = 0) {
    this.x = x;
    this.y = y;
    return this;
  }
}

const vectorPool = new ObjectPool(
  () => new Vector(),
  (vec) => vec.reset(),
  50
);

// Usage
function calculatePhysics() {
  const velocity = vectorPool.acquire();
  const position = vectorPool.acquire();
  
  // Use vectors...
  velocity.reset(10, 5);
  position.reset(100, 200);
  
  // Return to pool when done
  vectorPool.release(velocity);
  vectorPool.release(position);
}
```

### Efficient Data Structures

```javascript
// Use appropriate data structures for memory efficiency

// BAD: Using objects as dictionaries
const dictionary = {};
for (let i = 0; i < 1000; i++) {
  dictionary[`key${i}`] = `value${i}`;
}

// GOOD: Use Map for better memory usage
const map = new Map();
for (let i = 0; i < 1000; i++) {
  map.set(`key${i}`, `value${i}`);
}

// Use WeakMap when keys are objects and you want automatic cleanup
const metadata = new WeakMap();
function attachMetadata(obj, meta) {
  metadata.set(obj, meta); // Automatically cleaned when obj is GC'd
}

// Use typed arrays for numeric data
const positions = new Float32Array(1000); // More memory efficient than regular array
const indices = new Uint16Array(500);     // For integer data

// Use ArrayBuffer for raw binary data
const buffer = new ArrayBuffer(1024);
const view = new DataView(buffer);
view.setFloat32(0, 3.14);
view.setInt32(4, 42);
```

### Lazy Loading and Pagination

```javascript
// Lazy loading to reduce memory usage
class LazyDataLoader {
  constructor(dataSource) {
    this.dataSource = dataSource;
    this.cache = new Map();
    this.maxCacheSize = 100;
  }
  
  async getData(id) {
    // Check cache first
    if (this.cache.has(id)) {
      return this.cache.get(id);
    }
    
    // Load data
    const data = await this.dataSource.load(id);
    
    // Add to cache with size limit
    if (this.cache.size >= this.maxCacheSize) {
      const firstKey = this.cache.keys().next().value;
      this.cache.delete(firstKey);
    }
    
    this.cache.set(id, data);
    return data;
  }
  
  clearCache() {
    this.cache.clear();
  }
}

// Virtual scrolling for large lists
class VirtualList {
  constructor(container, itemHeight, bufferSize = 5) {
    this.container = container;
    this.itemHeight = itemHeight;
    this.bufferSize = bufferSize;
    this.visibleItems = new Map();
    this.data = [];
  }
  
  setData(data) {
    this.data = data;
    this.render();
  }
  
  render() {
    const containerHeight = this.container.clientHeight;
    const scrollTop = this.container.scrollTop;
    
    const startIndex = Math.max(0, 
      Math.floor(scrollTop / this.itemHeight) - this.bufferSize);
    const endIndex = Math.min(this.data.length - 1,
      Math.ceil((scrollTop + containerHeight) / this.itemHeight) + this.bufferSize);
    
    // Remove items outside visible range
    for (const [index, element] of this.visibleItems) {
      if (index < startIndex || index > endIndex) {
        element.remove();
        this.visibleItems.delete(index);
      }
    }
    
    // Add items in visible range
    for (let i = startIndex; i <= endIndex; i++) {
      if (!this.visibleItems.has(i) && this.data[i]) {
        const element = this.createItem(this.data[i], i);
        this.visibleItems.set(i, element);
        this.container.appendChild(element);
      }
    }
  }
  
  createItem(data, index) {
    const element = document.createElement('div');
    element.style.position = 'absolute';
    element.style.top = `${index * this.itemHeight}px`;
    element.style.height = `${this.itemHeight}px`;
    element.textContent = data;
    return element;
  }
}
```

## Browser-Specific Memory Management

### DevTools Memory Profiling

```javascript
// Performance marks for memory profiling
function profileMemoryUsage(label) {
  performance.mark(`${label}-start`);
  
  return {
    end: () => {
      performance.mark(`${label}-end`);
      performance.measure(label, `${label}-start`, `${label}-end`);
      
      if (performance.memory) {
        console.log(`${label} - Memory:`, {
          used: performance.memory.usedJSHeapSize,
          total: performance.memory.totalJSHeapSize
        });
      }
    }
  };
}

// Usage
const profiler = profileMemoryUsage('data-processing');

// ... do memory-intensive work
processLargeDataSet();

profiler.end();
```

### Memory-Conscious Programming Patterns

```javascript
// Stream processing instead of loading all data
async function* processLargeFile(fileHandle) {
  const reader = fileHandle.stream().getReader();
  let buffer = '';
  
  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      
      buffer += new TextDecoder().decode(value);
      
      // Process line by line to avoid memory buildup
      let newlineIndex;
      while ((newlineIndex = buffer.indexOf('\n')) !== -1) {
        const line = buffer.slice(0, newlineIndex);
        buffer = buffer.slice(newlineIndex + 1);
        
        yield processLine(line);
      }
    }
    
    // Process remaining buffer
    if (buffer.length > 0) {
      yield processLine(buffer);
    }
  } finally {
    reader.releaseLock();
  }
}

// Batch processing with cleanup
async function processBatches(items, batchSize = 100) {
  const results = [];
  
  for (let i = 0; i < items.length; i += batchSize) {
    const batch = items.slice(i, i + batchSize);
    const batchResults = await processBatch(batch);
    
    results.push(...batchResults);
    
    // Force garbage collection opportunity
    if (i % (batchSize * 10) === 0) {
      await new Promise(resolve => setTimeout(resolve, 0));
    }
  }
  
  return results;
}
```

## Node.js Memory Management

### Process Memory Monitoring

```javascript
// Node.js memory monitoring
function monitorNodeMemory() {
  const memUsage = process.memoryUsage();
  
  return {
    rss: Math.round(memUsage.rss / 1024 / 1024) + 'MB',        // Resident Set Size
    heapTotal: Math.round(memUsage.heapTotal / 1024 / 1024) + 'MB', // Total heap
    heapUsed: Math.round(memUsage.heapUsed / 1024 / 1024) + 'MB',   // Used heap
    external: Math.round(memUsage.external / 1024 / 1024) + 'MB',   // External memory
    arrayBuffers: Math.round(memUsage.arrayBuffers / 1024 / 1024) + 'MB' // ArrayBuffers
  };
}

// Set up periodic monitoring
setInterval(() => {
  const memory = monitorNodeMemory();
  console.log('Node.js Memory Usage:', memory);
  
  // Alert on high memory usage
  const heapUsedMB = parseInt(memory.heapUsed);
  if (heapUsedMB > 500) { // Alert if over 500MB
    console.warn('High memory usage detected!');
  }
}, 10000);

// Manual garbage collection (for debugging)
if (global.gc) {
  console.log('Before GC:', process.memoryUsage());
  global.gc();
  console.log('After GC:', process.memoryUsage());
}
```

### Stream Processing for Large Data

```javascript
const fs = require('fs');
const readline = require('readline');

// Process large files with streams
async function processLargeFile(filePath) {
  const fileStream = fs.createReadStream(filePath);
  const rl = readline.createInterface({
    input: fileStream,
    crlfDelay: Infinity
  });
  
  let lineCount = 0;
  
  for await (const line of rl) {
    // Process line without loading entire file into memory
    await processLine(line);
    lineCount++;
    
    // Periodic memory check
    if (lineCount % 10000 === 0) {
      console.log(`Processed ${lineCount} lines`, monitorNodeMemory());
    }
  }
  
  return lineCount;
}

async function processLine(line) {
  // Process individual line
  const data = JSON.parse(line);
  // ... process data
  return data;
}
```

## Best Practices Summary

1. **Avoid Memory Leaks**:
   - Remove event listeners
   - Clear timers and intervals
   - Break circular references
   - Nullify large object references

2. **Optimize Data Structures**:
   - Use appropriate collection types (Map, Set, WeakMap, WeakSet)
   - Use typed arrays for numeric data
   - Implement object pooling for frequently created objects

3. **Manage Large Datasets**:
   - Use pagination and virtual scrolling
   - Implement lazy loading
   - Stream process large files
   - Batch operations with cleanup breaks

4. **Monitor Memory Usage**:
   - Use browser DevTools
   - Implement memory monitoring in production
   - Profile memory usage regularly
   - Set up alerts for high memory usage

5. **Write Memory-Conscious Code**:
   - Minimize closure scope
   - Avoid accidental globals
   - Clean up resources explicitly
   - Use appropriate algorithms and data structures

Memory management in JavaScript requires understanding the garbage collection process, avoiding common leak patterns, and implementing efficient data handling strategies. While the garbage collector handles most memory management automatically, conscious programming practices are essential for building performant, scalable applications.