# Java Types and Collections - Comprehensive Reference

## Primitive Data Types

### Numeric Types
```java
public class PrimitiveTypes {
    // Integer types
    byte byteValue = 127;           // 8-bit: -128 to 127
    short shortValue = 32767;       // 16-bit: -32,768 to 32,767
    int intValue = 2147483647;      // 32-bit: -2^31 to 2^31-1
    long longValue = 9223372036854775807L; // 64-bit: -2^63 to 2^63-1

    // Floating-point types
    float floatValue = 3.14f;       // 32-bit IEEE 754
    double doubleValue = 3.14159265358979; // 64-bit IEEE 754

    // Character and boolean
    char charValue = 'A';           // 16-bit Unicode character
    boolean booleanValue = true;    // true or false

    // Demonstrating ranges and precision
    public void demonstrateNumericTypes() {
        System.out.println("Byte range: " + Byte.MIN_VALUE + " to " + Byte.MAX_VALUE);
        System.out.println("Int range: " + Integer.MIN_VALUE + " to " + Integer.MAX_VALUE);
        System.out.println("Float precision: " + Float.MIN_VALUE + " to " + Float.MAX_VALUE);

        // Automatic promotion
        byte b = 10;
        int i = b;  // byte to int promotion
        long l = i; // int to long promotion

        // Explicit casting
        int largeInt = 1000;
        byte smallByte = (byte) largeInt; // May lose data
    }
}
```

### Wrapper Classes and Autoboxing
```java
public class WrapperClasses {
    public void demonstrateWrappers() {
        // Wrapper classes for primitives
        Byte byteWrapper = Byte.valueOf((byte) 10);
        Short shortWrapper = Short.valueOf((short) 100);
        Integer intWrapper = Integer.valueOf(1000);
        Long longWrapper = Long.valueOf(10000L);
        Float floatWrapper = Float.valueOf(3.14f);
        Double doubleWrapper = Double.valueOf(3.14159);
        Character charWrapper = Character.valueOf('A');
        Boolean booleanWrapper = Boolean.valueOf(true);

        // Autoboxing and unboxing
        Integer autoBoxed = 42;        // int to Integer
        int unboxed = autoBoxed;       // Integer to int

        // Integer caching (-128 to 127)
        Integer i1 = 100;
        Integer i2 = 100;
        System.out.println(i1 == i2);  // true (cached)

        Integer i3 = 200;
        Integer i4 = 200;
        System.out.println(i3 == i4);  // false (not cached)
        System.out.println(i3.equals(i4)); // true (value comparison)

        // Utility methods
        int parsed = Integer.parseInt("123");
        String stringified = Integer.toString(456);
        int max = Integer.max(10, 20);
        int min = Integer.min(10, 20);
    }
}
```

## String and Text Processing

### String Operations
```java
public class StringOperations {
    public void stringBasics() {
        // String literal vs new String
        String literal = "Hello";        // String pool
        String newString = new String("Hello"); // Heap

        System.out.println(literal == newString);        // false
        System.out.println(literal.equals(newString));   // true

        // String immutability
        String original = "Hello";
        String modified = original.concat(" World"); // Creates new String
        System.out.println(original);    // Still "Hello"
        System.out.println(modified);    // "Hello World"

        // Common string operations
        String text = "  Java Programming  ";
        System.out.println(text.trim());                    // "Java Programming"
        System.out.println(text.toUpperCase());            // "  JAVA PROGRAMMING  "
        System.out.println(text.toLowerCase());            // "  java programming  "
        System.out.println(text.substring(2, 6));          // "Java"
        System.out.println(text.contains("Java"));         // true
        System.out.println(text.indexOf("Programming"));   // 7
        System.out.println(text.replace("Java", "Python")); // "  Python Programming  "

        // String splitting and joining
        String data = "apple,banana,cherry";
        String[] fruits = data.split(",");
        String joined = String.join("|", fruits); // "apple|banana|cherry"
    }

    // StringBuilder for mutable strings
    public String efficientStringBuilding(String[] parts) {
        StringBuilder sb = new StringBuilder();
        for (String part : parts) {
            sb.append(part).append(" ");
        }
        return sb.toString().trim();
    }

    // StringBuffer for thread-safe mutable strings
    public String threadSafeStringBuilding(String[] parts) {
        StringBuffer sb = new StringBuffer();
        for (String part : parts) {
            sb.append(part).append(" ");
        }
        return sb.toString().trim();
    }
}
```

## Collection Framework Overview

### Collection Hierarchy
```
Collection Interface
├── List (ordered, allows duplicates)
│   ├── ArrayList
│   ├── LinkedList
│   └── Vector (legacy, synchronized)
├── Set (no duplicates)
│   ├── HashSet
│   ├── LinkedHashSet
│   └── TreeSet (sorted)
└── Queue (FIFO operations)
    ├── PriorityQueue
    ├── ArrayDeque
    └── LinkedList

Map Interface (separate hierarchy)
├── HashMap
├── LinkedHashMap
├── TreeMap (sorted)
└── Hashtable (legacy, synchronized)
```

## List Implementations

### ArrayList
```java
public class ArrayListExample {
    public void demonstrateArrayList() {
        // Creation and initialization
        List<String> arrayList = new ArrayList<>();
        List<String> initializedList = new ArrayList<>(Arrays.asList("a", "b", "c"));
        List<String> sizedList = new ArrayList<>(100); // Initial capacity

        // Adding elements
        arrayList.add("first");
        arrayList.add(1, "second"); // Insert at index
        arrayList.addAll(Arrays.asList("third", "fourth"));

        // Accessing elements
        String first = arrayList.get(0);
        int size = arrayList.size();
        boolean isEmpty = arrayList.isEmpty();
        boolean contains = arrayList.contains("first");
        int index = arrayList.indexOf("second");

        // Modifying elements
        arrayList.set(0, "modified");
        arrayList.remove(0);            // Remove by index
        arrayList.remove("second");     // Remove by object

        // Iteration
        for (String item : arrayList) {
            System.out.println(item);
        }

        // Using iterator
        Iterator<String> iterator = arrayList.iterator();
        while (iterator.hasNext()) {
            String item = iterator.next();
            if (item.equals("unwanted")) {
                iterator.remove(); // Safe removal during iteration
            }
        }

        // Stream operations (Java 8+)
        arrayList.stream()
                 .filter(s -> s.startsWith("t"))
                 .map(String::toUpperCase)
                 .forEach(System.out::println);
    }
}
```

### LinkedList
```java
public class LinkedListExample {
    public void demonstrateLinkedList() {
        LinkedList<String> linkedList = new LinkedList<>();

        // LinkedList as List
        linkedList.add("element1");
        linkedList.add("element2");

        // LinkedList as Deque (double-ended queue)
        linkedList.addFirst("first");
        linkedList.addLast("last");

        String firstElement = linkedList.removeFirst();
        String lastElement = linkedList.removeLast();

        // Queue operations
        linkedList.offer("queued");    // Add to tail
        String polled = linkedList.poll(); // Remove from head

        // Stack operations
        linkedList.push("stacked");    // Add to head
        String popped = linkedList.pop(); // Remove from head

        // Performance consideration: LinkedList better for frequent insertions/deletions
        // ArrayList better for random access
    }
}
```

## Set Implementations

### HashSet
```java
public class HashSetExample {
    public void demonstrateHashSet() {
        Set<String> hashSet = new HashSet<>();

        // Adding elements (no duplicates)
        hashSet.add("apple");
        hashSet.add("banana");
        hashSet.add("apple"); // Duplicate, won't be added
        System.out.println(hashSet.size()); // 2

        // Set operations
        Set<String> set1 = new HashSet<>(Arrays.asList("a", "b", "c"));
        Set<String> set2 = new HashSet<>(Arrays.asList("b", "c", "d"));

        // Union
        Set<String> union = new HashSet<>(set1);
        union.addAll(set2); // {a, b, c, d}

        // Intersection
        Set<String> intersection = new HashSet<>(set1);
        intersection.retainAll(set2); // {b, c}

        // Difference
        Set<String> difference = new HashSet<>(set1);
        difference.removeAll(set2); // {a}

        // Custom objects in HashSet require proper equals() and hashCode()
        Set<Person> people = new HashSet<>();
        people.add(new Person("John", 30));
        people.add(new Person("John", 30)); // Duplicate if equals/hashCode implemented correctly
    }

    private static class Person {
        private String name;
        private int age;

        public Person(String name, int age) {
            this.name = name;
            this.age = age;
        }

        @Override
        public boolean equals(Object obj) {
            if (this == obj) return true;
            if (obj == null || getClass() != obj.getClass()) return false;
            Person person = (Person) obj;
            return age == person.age && Objects.equals(name, person.name);
        }

        @Override
        public int hashCode() {
            return Objects.hash(name, age);
        }
    }
}
```

### TreeSet
```java
public class TreeSetExample {
    public void demonstrateTreeSet() {
        // TreeSet maintains sorted order
        Set<String> treeSet = new TreeSet<>();
        treeSet.add("zebra");
        treeSet.add("apple");
        treeSet.add("banana");

        System.out.println(treeSet); // [apple, banana, zebra] - sorted

        // NavigableSet operations
        NavigableSet<String> navigableSet = new TreeSet<>(treeSet);
        System.out.println(navigableSet.first());      // "apple"
        System.out.println(navigableSet.last());       // "zebra"
        System.out.println(navigableSet.higher("banana")); // "zebra"
        System.out.println(navigableSet.lower("banana"));  // "apple"

        // Range operations
        NavigableSet<String> subset = navigableSet.subSet("apple", true, "zebra", false);
        System.out.println(subset); // [apple, banana]

        // Custom comparator
        Set<String> lengthSorted = new TreeSet<>(Comparator.comparing(String::length));
        lengthSorted.addAll(Arrays.asList("apple", "pie", "banana"));
        System.out.println(lengthSorted); // [pie, apple, banana] - sorted by length
    }
}
```

## Map Implementations

### HashMap
```java
public class HashMapExample {
    public void demonstrateHashMap() {
        Map<String, Integer> map = new HashMap<>();

        // Basic operations
        map.put("apple", 5);
        map.put("banana", 3);
        map.put("cherry", 8);

        Integer appleCount = map.get("apple");
        Integer orangeCount = map.getOrDefault("orange", 0);

        // Check operations
        boolean hasApple = map.containsKey("apple");
        boolean hasValue5 = map.containsValue(5);

        // Modification operations
        map.put("apple", 10);        // Update existing
        map.putIfAbsent("date", 2);  // Add only if key doesn't exist
        map.remove("banana");
        map.remove("cherry", 8);     // Remove only if key-value matches

        // Java 8+ operations
        map.compute("apple", (key, val) -> val == null ? 1 : val + 1);
        map.computeIfAbsent("elderberry", key -> key.length());
        map.computeIfPresent("apple", (key, val) -> val * 2);

        // Merge operation
        map.merge("apple", 5, Integer::sum); // Add 5 to existing value

        // Iteration
        for (Map.Entry<String, Integer> entry : map.entrySet()) {
            System.out.println(entry.getKey() + ": " + entry.getValue());
        }

        // Lambda iteration (Java 8+)
        map.forEach((key, value) -> System.out.println(key + " = " + value));

        // Stream operations
        map.entrySet().stream()
           .filter(entry -> entry.getValue() > 5)
           .forEach(entry -> System.out.println(entry.getKey()));
    }
}
```

### TreeMap
```java
public class TreeMapExample {
    public void demonstrateTreeMap() {
        // TreeMap maintains sorted order of keys
        Map<String, Integer> treeMap = new TreeMap<>();
        treeMap.put("zebra", 1);
        treeMap.put("apple", 2);
        treeMap.put("banana", 3);

        System.out.println(treeMap); // {apple=2, banana=3, zebra=1} - sorted by keys

        // NavigableMap operations
        NavigableMap<String, Integer> navigableMap = new TreeMap<>(treeMap);

        String firstKey = navigableMap.firstKey();   // "apple"
        String lastKey = navigableMap.lastKey();     // "zebra"

        Map.Entry<String, Integer> firstEntry = navigableMap.firstEntry();
        Map.Entry<String, Integer> lastEntry = navigableMap.lastEntry();

        String higherKey = navigableMap.higherKey("banana"); // "zebra"
        String lowerKey = navigableMap.lowerKey("banana");   // "apple"

        // Range operations
        NavigableMap<String, Integer> subMap = navigableMap.subMap("apple", true, "zebra", false);
        System.out.println(subMap); // {apple=2, banana=3}

        // Custom comparator for reverse order
        Map<String, Integer> reverseMap = new TreeMap<>(Collections.reverseOrder());
        reverseMap.putAll(treeMap);
        System.out.println(reverseMap); // {zebra=1, banana=3, apple=2}
    }
}
```

## Queue and Deque Implementations

### PriorityQueue
```java
public class PriorityQueueExample {
    public void demonstratePriorityQueue() {
        // Natural ordering (min heap)
        PriorityQueue<Integer> minHeap = new PriorityQueue<>();
        minHeap.addAll(Arrays.asList(5, 2, 8, 1, 9));

        while (!minHeap.isEmpty()) {
            System.out.println(minHeap.poll()); // 1, 2, 5, 8, 9
        }

        // Max heap using reverse comparator
        PriorityQueue<Integer> maxHeap = new PriorityQueue<>(Collections.reverseOrder());
        maxHeap.addAll(Arrays.asList(5, 2, 8, 1, 9));

        while (!maxHeap.isEmpty()) {
            System.out.println(maxHeap.poll()); // 9, 8, 5, 2, 1
        }

        // Custom objects with comparator
        PriorityQueue<Task> taskQueue = new PriorityQueue<>(
            Comparator.comparing(Task::getPriority).reversed()
        );

        taskQueue.add(new Task("Low priority task", 1));
        taskQueue.add(new Task("High priority task", 10));
        taskQueue.add(new Task("Medium priority task", 5));

        while (!taskQueue.isEmpty()) {
            Task task = taskQueue.poll();
            System.out.println(task.getName() + " (Priority: " + task.getPriority() + ")");
        }
    }

    private static class Task {
        private String name;
        private int priority;

        public Task(String name, int priority) {
            this.name = name;
            this.priority = priority;
        }

        public String getName() { return name; }
        public int getPriority() { return priority; }
    }
}
```

### ArrayDeque
```java
public class ArrayDequeExample {
    public void demonstrateArrayDeque() {
        Deque<String> deque = new ArrayDeque<>();

        // Adding elements
        deque.addFirst("first");
        deque.addLast("last");
        deque.offerFirst("new first");
        deque.offerLast("new last");

        // Removing elements
        String first = deque.removeFirst();
        String last = deque.removeLast();
        String polledFirst = deque.pollFirst();
        String polledLast = deque.pollLast();

        // Peek operations
        String peekFirst = deque.peekFirst();
        String peekLast = deque.peekLast();

        // Use as stack (LIFO)
        deque.push("stack item 1");
        deque.push("stack item 2");
        String popped = deque.pop(); // "stack item 2"

        // Use as queue (FIFO)
        deque.offer("queue item 1");
        deque.offer("queue item 2");
        String polled = deque.poll(); // "queue item 1"
    }
}
```

## Specialized Collections

### EnumSet and EnumMap
```java
public class EnumCollectionsExample {
    enum Day {
        MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY
    }

    public void demonstrateEnumSet() {
        // EnumSet - efficient set implementation for enums
        EnumSet<Day> workDays = EnumSet.of(Day.MONDAY, Day.TUESDAY, Day.WEDNESDAY,
                                          Day.THURSDAY, Day.FRIDAY);
        EnumSet<Day> weekend = EnumSet.of(Day.SATURDAY, Day.SUNDAY);
        EnumSet<Day> allDays = EnumSet.allOf(Day.class);
        EnumSet<Day> noDays = EnumSet.noneOf(Day.class);

        // Range operation
        EnumSet<Day> midWeek = EnumSet.range(Day.TUESDAY, Day.THURSDAY);

        System.out.println("Work days: " + workDays);
        System.out.println("Weekend: " + weekend);
        System.out.println("Mid week: " + midWeek);
    }

    public void demonstrateEnumMap() {
        // EnumMap - efficient map implementation with enum keys
        EnumMap<Day, String> dayActivities = new EnumMap<>(Day.class);

        dayActivities.put(Day.MONDAY, "Start of work week");
        dayActivities.put(Day.FRIDAY, "TGIF!");
        dayActivities.put(Day.SATURDAY, "Weekend relaxation");
        dayActivities.put(Day.SUNDAY, "Prepare for next week");

        for (Map.Entry<Day, String> entry : dayActivities.entrySet()) {
            System.out.println(entry.getKey() + ": " + entry.getValue());
        }
    }
}
```

## Collection Utilities and Best Practices

### Collections Utility Class
```java
public class CollectionUtilitiesExample {
    public void demonstrateCollectionUtilities() {
        List<String> list = new ArrayList<>(Arrays.asList("c", "a", "b", "d"));

        // Sorting
        Collections.sort(list); // Natural order
        Collections.sort(list, Collections.reverseOrder()); // Reverse order
        Collections.shuffle(list); // Random order

        // Searching
        Collections.sort(list); // Must be sorted for binary search
        int index = Collections.binarySearch(list, "b");

        // Extremes
        String max = Collections.max(list);
        String min = Collections.min(list);

        // Frequency
        int frequency = Collections.frequency(list, "a");

        // Immutable collections
        List<String> immutableList = Collections.unmodifiableList(list);
        Set<String> immutableSet = Collections.unmodifiableSet(new HashSet<>(list));
        Map<String, Integer> originalMap = new HashMap<>();
        Map<String, Integer> immutableMap = Collections.unmodifiableMap(originalMap);

        // Empty collections
        List<String> emptyList = Collections.emptyList();
        Set<String> emptySet = Collections.emptySet();
        Map<String, Integer> emptyMap = Collections.emptyMap();

        // Singleton collections
        List<String> singletonList = Collections.singletonList("only");
        Set<String> singletonSet = Collections.singleton("only");
        Map<String, Integer> singletonMap = Collections.singletonMap("key", 1);

        // Synchronized collections
        List<String> syncList = Collections.synchronizedList(new ArrayList<>());
        Set<String> syncSet = Collections.synchronizedSet(new HashSet<>());
        Map<String, Integer> syncMap = Collections.synchronizedMap(new HashMap<>());

        // Fill and replace
        Collections.fill(list, "default");
        Collections.replaceAll(list, "default", "replaced");
    }
}
```

### Modern Collection Factory Methods (Java 9+)
```java
public class ModernCollectionFactories {
    public void demonstrateFactoryMethods() {
        // Immutable lists
        List<String> list = List.of("a", "b", "c");
        // list.add("d"); // UnsupportedOperationException

        // Immutable sets
        Set<String> set = Set.of("a", "b", "c");
        // No duplicate elements allowed

        // Immutable maps
        Map<String, Integer> map = Map.of(
            "a", 1,
            "b", 2,
            "c", 3
        );

        // For more than 10 entries, use Map.ofEntries
        Map<String, Integer> largeMap = Map.ofEntries(
            Map.entry("key1", 1),
            Map.entry("key2", 2),
            Map.entry("key3", 3)
        );

        // Converting streams to collections
        List<String> streamList = Stream.of("a", "b", "c")
                                       .collect(Collectors.toList());

        Set<String> streamSet = Stream.of("a", "b", "c")
                                     .collect(Collectors.toSet());

        Map<String, Integer> streamMap = Stream.of("apple", "banana", "cherry")
                                              .collect(Collectors.toMap(
                                                  Function.identity(),
                                                  String::length
                                              ));
    }
}
```

This comprehensive guide covers all Java types and collections, from primitive types to advanced collection operations, providing practical examples for effective data structure usage in Java applications.