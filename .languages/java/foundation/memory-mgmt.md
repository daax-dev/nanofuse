# Java Memory Management - Complete Guide

## JVM Memory Model

### Heap Memory Structure
```
JVM Memory Layout:
┌─────────────────────────────────────────┐
│                Heap Memory               │
├─────────────────────────────────────────┤
│  Young Generation    │  Old Generation  │
├──────────────────────┴──────────────────┤
│ Eden | S0 | S1       │ Tenured Space    │
├─────────────────────────────────────────┤
│            Method Area/Metaspace         │
├─────────────────────────────────────────┤
│              Direct Memory               │
└─────────────────────────────────────────┘

Non-Heap Memory:
┌─────────────────────────────────────────┐
│ PC Registers │ JVM Stack │ Native Stack │
└─────────────────────────────────────────┘
```

### Memory Areas Detailed

#### Heap Memory
```java
// Object creation - allocated in heap
public class MemoryExample {
    // Instance variables stored in heap with object
    private String name;
    private List<String> items;

    public void demonstrateHeapAllocation() {
        // All objects created here go to heap
        String str = new String("Hello");           // Heap
        StringBuilder sb = new StringBuilder();     // Heap
        List<String> list = new ArrayList<>();     // Heap
        Map<String, Object> map = new HashMap<>();  // Heap

        // Even array objects go to heap
        int[] numbers = new int[1000];              // Heap
        String[] words = new String[100];           // Heap
    }
}
```

#### Stack Memory
```java
public class StackMemoryExample {
    public void stackAllocation() {
        // Local variables and method parameters on stack
        int localInt = 42;              // Stack
        double localDouble = 3.14;      // Stack
        boolean flag = true;            // Stack

        // Object references on stack, objects on heap
        String str = "Hello";           // Reference on stack, object on heap
        List<String> list = new ArrayList<>(); // Reference on stack, object on heap

        methodWithParameters(localInt, str);
    }

    private void methodWithParameters(int param, String strParam) {
        // Parameters stored on stack for this method frame
        int localVar = param * 2;      // Local variable on stack

        // Each method call creates new stack frame
        recursiveMethod(0);
    }

    private void recursiveMethod(int depth) {
        if (depth < 5) {
            int localDepth = depth + 1; // Each frame has its own variables
            recursiveMethod(localDepth);
        }
        // Stack frame destroyed when method returns
    }
}
```

#### Method Area/Metaspace (Java 8+)
```java
public class MetaspaceExample {
    // Class-level constants stored in method area
    public static final String CONSTANT = "Global Constant";
    private static int staticCounter = 0;

    // Method bytecode stored in method area
    public static void staticMethod() {
        staticCounter++; // Static variables in method area
    }

    // Class metadata, method information stored in metaspace
    static {
        System.out.println("Class loading - stored in metaspace");
    }
}
```

## Garbage Collection Fundamentals

### Object Lifecycle
```java
public class ObjectLifecycleExample {
    public void demonstrateObjectLifecycle() {
        // 1. Object Creation - allocated in Eden space
        StringBuilder sb = new StringBuilder("Initial");

        // 2. Object Usage - may survive minor GC
        for (int i = 0; i < 1000; i++) {
            sb.append("data").append(i);
        }

        // 3. Object becomes eligible for GC when no longer referenced
        sb = null; // Reference removed, object eligible for collection

        // 4. Force GC suggestion (not guaranteed)
        System.gc(); // Request garbage collection

        // 5. Object finalization (if finalize() method overridden)
        // 6. Memory reclamation
    }

    // Demonstrating strong, weak, soft references
    public void referenceTypes() {
        String strongRef = new String("Strong Reference"); // Prevents GC

        // Weak reference - doesn't prevent GC
        WeakReference<String> weakRef = new WeakReference<>(new String("Weak"));

        // Soft reference - GC'd when memory is needed
        SoftReference<String> softRef = new SoftReference<>(new String("Soft"));

        // Phantom reference - for cleanup notification
        ReferenceQueue<String> refQueue = new ReferenceQueue<>();
        PhantomReference<String> phantomRef = new PhantomReference<>(
            new String("Phantom"), refQueue);
    }
}
```

### Generational Garbage Collection
```java
public class GenerationalGCExample {
    public void youngGenerationExample() {
        // Objects typically allocated in Eden space
        List<String> shortLived = new ArrayList<>();
        for (int i = 0; i < 1000; i++) {
            shortLived.add("String " + i); // Allocated in Eden
        }
        // Most objects die young - collected in minor GC
        shortLived = null;
    }

    // Long-lived objects eventually promoted to Old Generation
    private static final Map<String, Object> cache = new HashMap<>();

    public void oldGenerationExample() {
        // Objects that survive multiple minor GCs move to Old Generation
        for (int i = 0; i < 10000; i++) {
            cache.put("key" + i, new LongLivedObject());

            // Occasionally trigger minor GC
            if (i % 1000 == 0) {
                System.gc();
            }
        }
        // Cache objects likely promoted to Old Generation
    }

    private static class LongLivedObject {
        private String data = "Long lived data";
        private long timestamp = System.currentTimeMillis();
    }
}
```

## Memory Leaks and Prevention

### Common Memory Leak Patterns
```java
public class MemoryLeakExamples {

    // 1. Static Collection Growing Indefinitely
    private static final List<String> staticList = new ArrayList<>();

    public void staticCollectionLeak() {
        // This grows forever - never collected
        staticList.add("New item " + System.currentTimeMillis());
    }

    // 2. Listener/Observer Not Removed
    private List<EventListener> listeners = new ArrayList<>();

    public void addListener(EventListener listener) {
        listeners.add(listener);
        // Memory leak if listeners not removed when done
    }

    public void removeListener(EventListener listener) {
        listeners.remove(listener); // Important to clean up!
    }

    // 3. Thread Local Variables
    private ThreadLocal<StringBuilder> threadLocalBuilder = new ThreadLocal<StringBuilder>() {
        @Override
        protected StringBuilder initialValue() {
            return new StringBuilder();
        }
    };

    public void threadLocalLeak() {
        StringBuilder sb = threadLocalBuilder.get();
        sb.append("Data");
        // Should clear when done: threadLocalBuilder.remove();
    }

    // 4. Unclosed Resources
    public void resourceLeak() throws IOException {
        FileInputStream fis = null;
        try {
            fis = new FileInputStream("file.txt");
            // Use file input stream
        } finally {
            // Without this, file handle leaks
            if (fis != null) {
                fis.close();
            }
        }
    }

    // Better approach with try-with-resources
    public void properResourceHandling() throws IOException {
        try (FileInputStream fis = new FileInputStream("file.txt")) {
            // Use file input stream - automatically closed
        }
    }

    // 5. Inner Class References
    public class OuterClass {
        private String data = "Outer data";

        public class InnerClass {
            public void doSomething() {
                System.out.println(data); // Holds reference to outer instance
            }
        }

        // Memory leak: inner class keeps outer instance alive
        public InnerClass createInner() {
            return new InnerClass();
        }

        // Better: static inner class
        public static class StaticInnerClass {
            public void doSomething() {
                System.out.println("No outer reference");
            }
        }
    }
}
```

### Memory Leak Prevention
```java
public class MemoryLeakPrevention {

    // Use WeakReference for cached objects
    private Map<String, WeakReference<ExpensiveObject>> cache = new ConcurrentHashMap<>();

    public ExpensiveObject getCachedObject(String key) {
        WeakReference<ExpensiveObject> ref = cache.get(key);
        ExpensiveObject obj = (ref != null) ? ref.get() : null;

        if (obj == null) {
            obj = new ExpensiveObject(key);
            cache.put(key, new WeakReference<>(obj));
        }

        return obj;
    }

    // Periodic cleanup of expired entries
    public void cleanupExpiredEntries() {
        cache.entrySet().removeIf(entry -> entry.getValue().get() == null);
    }

    // Proper singleton pattern - prevents memory accumulation
    private static final class SingletonHolder {
        private static final MemoryLeakPrevention INSTANCE = new MemoryLeakPrevention();
    }

    public static MemoryLeakPrevention getInstance() {
        return SingletonHolder.INSTANCE;
    }

    // Cleanup method for proper resource management
    public void cleanup() {
        cache.clear();
        // Clear other resources
    }

    private static class ExpensiveObject {
        private String data;

        public ExpensiveObject(String data) {
            this.data = data;
        }
    }
}
```

## Memory Optimization Techniques

### Object Pool Pattern
```java
public class ObjectPoolExample {
    // Pool for expensive-to-create objects
    private final BlockingQueue<StringBuilder> stringBuilderPool;
    private final int maxPoolSize;

    public ObjectPoolExample(int poolSize) {
        this.maxPoolSize = poolSize;
        this.stringBuilderPool = new ArrayBlockingQueue<>(poolSize);

        // Pre-populate pool
        for (int i = 0; i < poolSize; i++) {
            stringBuilderPool.offer(new StringBuilder());
        }
    }

    public StringBuilder borrowObject() {
        StringBuilder sb = stringBuilderPool.poll();
        if (sb == null) {
            // Pool empty, create new instance
            sb = new StringBuilder();
        } else {
            // Reset borrowed object
            sb.setLength(0);
        }
        return sb;
    }

    public void returnObject(StringBuilder sb) {
        if (sb != null && stringBuilderPool.size() < maxPoolSize) {
            sb.setLength(0); // Reset state
            stringBuilderPool.offer(sb);
        }
    }

    // Usage example
    public String processData(String input) {
        StringBuilder sb = borrowObject();
        try {
            sb.append(input).append(" processed");
            return sb.toString();
        } finally {
            returnObject(sb);
        }
    }
}
```

### Flyweight Pattern for Memory Efficiency
```java
public class FlyweightExample {
    // Intrinsic state (shared)
    private static final Map<String, CharacterType> characterTypes = new HashMap<>();

    // Factory for flyweight objects
    public static CharacterType getCharacterType(String fontFamily, int fontSize, Color color) {
        String key = fontFamily + fontSize + color.toString();
        return characterTypes.computeIfAbsent(key,
            k -> new CharacterType(fontFamily, fontSize, color));
    }

    // Flyweight class - intrinsic state only
    private static class CharacterType {
        private final String fontFamily;
        private final int fontSize;
        private final Color color;

        public CharacterType(String fontFamily, int fontSize, Color color) {
            this.fontFamily = fontFamily;
            this.fontSize = fontSize;
            this.color = color;
        }

        // Extrinsic state passed as parameters
        public void render(char character, int x, int y) {
            System.out.printf("Rendering '%c' at (%d,%d) with %s %d %s%n",
                character, x, y, fontFamily, fontSize, color);
        }
    }
}
```

### String Optimization
```java
public class StringOptimization {

    // Inefficient - creates many String objects
    public String inefficientConcatenation(String[] parts) {
        String result = "";
        for (String part : parts) {
            result += part; // Creates new String object each iteration
        }
        return result;
    }

    // Efficient - uses StringBuilder
    public String efficientConcatenation(String[] parts) {
        StringBuilder sb = new StringBuilder();
        for (String part : parts) {
            sb.append(part);
        }
        return sb.toString();
    }

    // String interning for repeated strings
    private static final Map<String, String> stringCache = new ConcurrentHashMap<>();

    public String getInternedString(String input) {
        return stringCache.computeIfAbsent(input, String::intern);
    }

    // Use String.join for simple concatenation
    public String joinStrings(String[] parts, String delimiter) {
        return String.join(delimiter, parts);
    }

    // Efficient string formatting
    public String formatMessage(String name, int age, double salary) {
        // Better than concatenation
        return String.format("Name: %s, Age: %d, Salary: %.2f", name, age, salary);
    }
}
```

## Memory Monitoring and Profiling

### Runtime Memory Information
```java
public class MemoryMonitoring {

    public void printMemoryUsage() {
        Runtime runtime = Runtime.getRuntime();

        long maxMemory = runtime.maxMemory();     // Maximum heap size
        long totalMemory = runtime.totalMemory(); // Current heap size
        long freeMemory = runtime.freeMemory();   // Free memory in heap
        long usedMemory = totalMemory - freeMemory;

        System.out.printf("Max Memory: %d MB%n", maxMemory / (1024 * 1024));
        System.out.printf("Total Memory: %d MB%n", totalMemory / (1024 * 1024));
        System.out.printf("Used Memory: %d MB%n", usedMemory / (1024 * 1024));
        System.out.printf("Free Memory: %d MB%n", freeMemory / (1024 * 1024));
        System.out.printf("Memory Usage: %.2f%%%n",
            (double) usedMemory / maxMemory * 100);
    }

    // Using MemoryMXBean for detailed information
    public void detailedMemoryInfo() {
        MemoryMXBean memoryBean = ManagementFactory.getMemoryMXBean();

        MemoryUsage heapUsage = memoryBean.getHeapMemoryUsage();
        MemoryUsage nonHeapUsage = memoryBean.getNonHeapMemoryUsage();

        System.out.println("Heap Memory Usage:");
        printMemoryUsage("Heap", heapUsage);

        System.out.println("\nNon-Heap Memory Usage:");
        printMemoryUsage("Non-Heap", nonHeapUsage);

        // Individual memory pools
        List<MemoryPoolMXBean> memoryPools = ManagementFactory.getMemoryPoolMXBeans();
        for (MemoryPoolMXBean pool : memoryPools) {
            System.out.printf("\nMemory Pool: %s (%s)%n",
                pool.getName(), pool.getType());
            printMemoryUsage(pool.getName(), pool.getUsage());
        }
    }

    private void printMemoryUsage(String name, MemoryUsage usage) {
        System.out.printf("%s - Init: %d MB, Used: %d MB, Committed: %d MB, Max: %d MB%n",
            name,
            usage.getInit() / (1024 * 1024),
            usage.getUsed() / (1024 * 1024),
            usage.getCommitted() / (1024 * 1024),
            usage.getMax() / (1024 * 1024));
    }

    // Memory threshold notifications
    public void setupMemoryAlerts() {
        MemoryMXBean memoryBean = ManagementFactory.getMemoryMXBean();

        // Set up notification listener
        NotificationEmitter emitter = (NotificationEmitter) memoryBean;
        emitter.addNotificationListener(new MemoryNotificationListener(), null, null);

        // Enable memory threshold
        List<MemoryPoolMXBean> pools = ManagementFactory.getMemoryPoolMXBeans();
        for (MemoryPoolMXBean pool : pools) {
            if (pool.getType() == MemoryType.HEAP && pool.isUsageThresholdSupported()) {
                long maxMemory = pool.getUsage().getMax();
                long threshold = (long) (maxMemory * 0.8); // 80% threshold
                pool.setUsageThreshold(threshold);
            }
        }
    }

    private static class MemoryNotificationListener implements NotificationListener {
        @Override
        public void handleNotification(Notification notification, Object handback) {
            if (notification.getType().equals(MemoryNotificationInfo.MEMORY_THRESHOLD_EXCEEDED)) {
                System.err.println("Memory threshold exceeded: " + notification.getMessage());
                // Take corrective action: clear caches, trigger GC, etc.
            }
        }
    }
}
```

### Heap Dump Analysis
```java
public class HeapDumpUtility {

    // Programmatically create heap dump
    public void createHeapDump(String filePath) {
        try {
            MBeanServer server = ManagementFactory.getPlatformMBeanServer();
            HotSpotDiagnosticMXBean mxBean = ManagementFactory.newPlatformMXBeanProxy(
                server, "com.sun.management:type=HotSpotDiagnostic", HotSpotDiagnosticMXBean.class);

            mxBean.dumpHeap(filePath, true);
            System.out.println("Heap dump created: " + filePath);
        } catch (Exception e) {
            System.err.println("Failed to create heap dump: " + e.getMessage());
        }
    }

    // Force garbage collection for analysis
    public void analyzeMemoryBeforeAfterGC() {
        Runtime runtime = Runtime.getRuntime();

        System.out.println("Before GC:");
        printMemoryStats(runtime);

        System.gc();

        // Wait for GC to complete
        try {
            Thread.sleep(1000);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }

        System.out.println("After GC:");
        printMemoryStats(runtime);
    }

    private void printMemoryStats(Runtime runtime) {
        long total = runtime.totalMemory();
        long free = runtime.freeMemory();
        long used = total - free;

        System.out.printf("Used: %d MB, Free: %d MB, Total: %d MB%n",
            used / (1024 * 1024), free / (1024 * 1024), total / (1024 * 1024));
    }
}
```

This comprehensive guide covers Java's memory management from basic concepts to advanced optimization techniques, providing practical examples for understanding and managing memory in Java applications.