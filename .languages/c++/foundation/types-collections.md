# C++ Types and Collections

## Overview

Modern C++ provides a rich type system and comprehensive collection library through the Standard Template Library (STL). This guide covers fundamental types, container classes, iterators, algorithms, and modern collection utilities.

## Fundamental Types

### Built-in Types

```cpp
#include <cstdint>
#include <limits>
#include <type_traits>

// Integer types
int8_t   i8;     // signed 8-bit
uint8_t  u8;     // unsigned 8-bit
int16_t  i16;    // signed 16-bit
uint16_t u16;    // unsigned 16-bit
int32_t  i32;    // signed 32-bit
uint32_t u32;    // unsigned 32-bit
int64_t  i64;    // signed 64-bit
uint64_t u64;    // unsigned 64-bit

// Fast and least types
int_fast32_t fast32;  // fastest type with at least 32 bits
int_least32_t least32; // smallest type with at least 32 bits

// Pointer-sized integers
intptr_t  ptr_int;    // signed integer that can hold a pointer
uintptr_t ptr_uint;   // unsigned integer that can hold a pointer
ptrdiff_t ptr_diff;   // signed integer for pointer arithmetic
size_t    size_type;  // unsigned integer for sizes and indices

// Floating point types
float       f32;      // 32-bit floating point
double      f64;      // 64-bit floating point
long double f128;     // extended precision (often 80 or 128-bit)

// Character types
char     c;           // basic character type
char8_t  c8;          // UTF-8 character (C++20)
char16_t c16;         // UTF-16 character
char32_t c32;         // UTF-32 character
wchar_t  wc;          // wide character

// Boolean type
bool flag = true;

// Void type and nullptr
void* ptr = nullptr;
std::nullptr_t null_ptr = nullptr;

// Type limits and properties
void type_info_example() {
    std::cout << "int max: " << std::numeric_limits<int>::max() << std::endl;
    std::cout << "float epsilon: " << std::numeric_limits<float>::epsilon() << std::endl;
    std::cout << "double digits: " << std::numeric_limits<double>::digits10 << std::endl;

    // Type traits
    static_assert(std::is_integral_v<int>);
    static_assert(std::is_floating_point_v<double>);
    static_assert(std::is_signed_v<int>);
    static_assert(std::is_unsigned_v<size_t>);
}
```

### Type Deduction and Auto

```cpp
#include <vector>
#include <string>
#include <map>

void type_deduction_examples() {
    // Auto type deduction
    auto i = 42;                    // int
    auto f = 3.14;                  // double
    auto s = "hello";               // const char*
    auto str = std::string("hello"); // std::string

    // Auto with complex types
    auto vec = std::vector<int>{1, 2, 3, 4, 5};
    auto it = vec.begin();          // std::vector<int>::iterator
    auto lambda = [](int x) { return x * 2; }; // lambda type

    // decltype
    int x = 42;
    decltype(x) y = x;              // int
    decltype(auto) z = x;           // int
    decltype(auto) ref = (x);       // int& (note the parentheses)

    // Function return type deduction
    auto multiply = [](auto a, auto b) -> decltype(a * b) {
        return a * b;
    };

    // Template argument deduction (C++17)
    std::vector v{1, 2, 3, 4, 5};   // std::vector<int>
    std::pair p{42, "hello"};       // std::pair<int, const char*>
}

// Function templates with auto
template<typename T>
auto process_container(const T& container) -> decltype(container.size()) {
    return container.size();
}

// Generic lambda (C++14)
auto generic_lambda = [](auto&& x, auto&& y) {
    return std::forward<decltype(x)>(x) + std::forward<decltype(y)>(y);
};
```

## STL Containers

### Sequence Containers

#### vector - Dynamic Array

```cpp
#include <vector>
#include <algorithm>
#include <numeric>

void vector_examples() {
    // Construction
    std::vector<int> v1;                    // empty vector
    std::vector<int> v2(10);                // 10 default-initialized elements
    std::vector<int> v3(10, 42);            // 10 elements, all 42
    std::vector<int> v4{1, 2, 3, 4, 5};     // initializer list
    std::vector<int> v5(v4);                // copy constructor
    std::vector<int> v6(v4.begin() + 1, v4.end() - 1); // range constructor

    // Capacity and size
    v1.reserve(100);                        // reserve capacity
    std::cout << "Size: " << v1.size() << std::endl;
    std::cout << "Capacity: " << v1.capacity() << std::endl;
    std::cout << "Max size: " << v1.max_size() << std::endl;

    // Element access
    v4[0] = 10;                            // unchecked access
    v4.at(1) = 20;                         // checked access (throws)
    int& front = v4.front();               // first element
    int& back = v4.back();                 // last element
    int* data_ptr = v4.data();             // raw data pointer

    // Modifiers
    v1.push_back(42);                      // add element at end
    v1.emplace_back(43);                   // construct element in place
    v1.pop_back();                         // remove last element

    v1.insert(v1.begin() + 1, 99);        // insert single element
    v1.insert(v1.end(), 3, 100);          // insert multiple copies
    v1.insert(v1.end(), {101, 102, 103}); // insert from initializer list

    v1.erase(v1.begin());                  // erase single element
    v1.erase(v1.begin(), v1.begin() + 2); // erase range

    v1.clear();                            // remove all elements
    v1.resize(5);                          // resize container
    v1.shrink_to_fit();                    // reduce capacity to size

    // Algorithms with vectors
    std::vector<int> numbers = {5, 2, 8, 1, 9, 3};

    std::sort(numbers.begin(), numbers.end());
    auto it = std::lower_bound(numbers.begin(), numbers.end(), 5);

    int sum = std::accumulate(numbers.begin(), numbers.end(), 0);

    std::transform(numbers.begin(), numbers.end(), numbers.begin(),
                   [](int x) { return x * x; });
}

// Custom allocator example
template<typename T>
class DebugAllocator {
public:
    using value_type = T;

    T* allocate(size_t n) {
        std::cout << "Allocating " << n << " objects of size " << sizeof(T) << std::endl;
        return std::allocator<T>{}.allocate(n);
    }

    void deallocate(T* p, size_t n) {
        std::cout << "Deallocating " << n << " objects" << std::endl;
        std::allocator<T>{}.deallocate(p, n);
    }
};

void custom_allocator_example() {
    std::vector<int, DebugAllocator<int>> debug_vec;
    debug_vec.push_back(42);
    debug_vec.resize(10);
}
```

#### deque - Double-Ended Queue

```cpp
#include <deque>

void deque_examples() {
    std::deque<int> dq{1, 2, 3, 4, 5};

    // Double-ended operations
    dq.push_front(0);                      // add to front
    dq.push_back(6);                       // add to back
    dq.pop_front();                        // remove from front
    dq.pop_back();                         // remove from back

    // Random access (like vector)
    dq[2] = 99;
    int value = dq.at(3);

    // Iterators work the same as vector
    for (auto it = dq.begin(); it != dq.end(); ++it) {
        std::cout << *it << " ";
    }

    // Efficient insertion/deletion at both ends
    // Better than vector for front operations
    // Slightly more overhead than vector for random access
}
```

#### list - Doubly Linked List

```cpp
#include <list>

void list_examples() {
    std::list<int> lst{3, 1, 4, 1, 5, 9, 2, 6};

    // No random access, only bidirectional iterators
    lst.push_front(0);
    lst.push_back(7);

    // Efficient insertion/deletion anywhere
    auto it = std::find(lst.begin(), lst.end(), 4);
    lst.insert(it, 99);                    // insert before 4
    lst.erase(it);                         // remove 4

    // List-specific operations
    lst.sort();                            // sort the list
    lst.reverse();                         // reverse the list
    lst.unique();                          // remove consecutive duplicates

    std::list<int> other{10, 20, 30};
    lst.merge(other);                      // merge sorted lists

    // Splice operations (move elements between lists)
    std::list<int> source{100, 200, 300};
    lst.splice(lst.end(), source);         // move all from source to lst
}
```

#### forward_list - Singly Linked List

```cpp
#include <forward_list>

void forward_list_examples() {
    std::forward_list<int> flst{1, 2, 3, 4, 5};

    // Only forward iterators
    flst.push_front(0);
    // No push_back() - only front operations

    // Insert after position
    auto it = flst.begin();
    std::advance(it, 2);
    flst.insert_after(it, 99);

    // Erase after position
    flst.erase_after(flst.begin());

    // Forward list specific operations
    flst.sort();
    flst.reverse();
    flst.unique();

    // Memory efficient - only stores next pointer
    // Use when you need minimal memory overhead
}
```

#### array - Fixed-Size Array

```cpp
#include <array>

void array_examples() {
    // Fixed size determined at compile time
    std::array<int, 5> arr{1, 2, 3, 4, 5};

    // Same interface as vector for element access
    arr[0] = 10;
    arr.at(1) = 20;
    int& front = arr.front();
    int& back = arr.back();

    // Size is compile-time constant
    constexpr size_t size = arr.size();
    static_assert(size == 5);

    // Can be used in constexpr contexts
    constexpr std::array<int, 3> constexpr_arr{1, 2, 3};
    constexpr int sum = constexpr_arr[0] + constexpr_arr[1] + constexpr_arr[2];

    // Stack allocated, no dynamic allocation
    // Cache-friendly, minimal overhead
    // Good for small, fixed-size collections
}
```

### Associative Containers

#### set and multiset

```cpp
#include <set>

void set_examples() {
    // Ordered sets (usually Red-Black trees)
    std::set<int> s{5, 2, 8, 1, 9, 2};     // duplicates removed
    std::multiset<int> ms{5, 2, 8, 1, 9, 2}; // duplicates allowed

    // Insertion
    s.insert(3);
    auto [it, inserted] = s.insert(5);     // returns pair<iterator, bool>
    s.emplace(7);                          // construct in place

    // Find operations
    auto find_it = s.find(5);
    if (find_it != s.end()) {
        std::cout << "Found: " << *find_it << std::endl;
    }

    size_t count = s.count(5);             // 0 or 1 for set, can be > 1 for multiset

    // Bounds
    auto lower = s.lower_bound(4);         // first element not less than 4
    auto upper = s.upper_bound(6);         // first element greater than 6
    auto range = s.equal_range(5);         // pair of lower_bound and upper_bound

    // Erase
    s.erase(5);                            // erase by key
    s.erase(s.find(8));                    // erase by iterator
    s.erase(s.lower_bound(1), s.upper_bound(3)); // erase range

    // Custom comparator
    std::set<std::string, std::greater<std::string>> reverse_set;

    // Custom type with custom comparator
    struct Person {
        std::string name;
        int age;
    };

    auto person_comp = [](const Person& a, const Person& b) {
        return a.age < b.age || (a.age == b.age && a.name < b.name);
    };

    std::set<Person, decltype(person_comp)> person_set(person_comp);
    person_set.emplace("Alice", 30);
    person_set.emplace("Bob", 25);
}
```

#### map and multimap

```cpp
#include <map>
#include <string>

void map_examples() {
    std::map<std::string, int> word_count;
    std::multimap<std::string, int> multi_map;

    // Insertion
    word_count["hello"] = 1;               // operator[] (creates if not exists)
    word_count.insert({"world", 1});       // insert pair
    word_count.emplace("foo", 42);         // construct in place

    // Safe insertion (doesn't overwrite)
    auto [it, inserted] = word_count.try_emplace("hello", 999);
    if (!inserted) {
        std::cout << "Key already exists with value: " << it->second << std::endl;
    }

    // Access
    int count = word_count["hello"];       // creates if not exists

    // Safe access
    if (auto it = word_count.find("hello"); it != word_count.end()) {
        std::cout << "hello: " << it->second << std::endl;
    }

    // at() method (throws if not found)
    try {
        int value = word_count.at("nonexistent");
    } catch (const std::out_of_range& e) {
        std::cout << "Key not found" << std::endl;
    }

    // Modify
    word_count["hello"]++;                 // increment
    word_count.insert_or_assign("world", 5); // insert or update

    // Iteration
    for (const auto& [key, value] : word_count) {
        std::cout << key << ": " << value << std::endl;
    }

    // Custom key type
    struct Point {
        int x, y;

        bool operator<(const Point& other) const {
            return std::tie(x, y) < std::tie(other.x, other.y);
        }
    };

    std::map<Point, std::string> point_names;
    point_names[{0, 0}] = "Origin";
    point_names[{1, 1}] = "Unit diagonal";
}
```

### Unordered Containers (C++11)

#### unordered_set and unordered_multiset

```cpp
#include <unordered_set>
#include <functional>

void unordered_set_examples() {
    std::unordered_set<int> us{1, 2, 3, 4, 5};

    // Hash-based, average O(1) operations
    us.insert(6);
    us.emplace(7);

    // Find operations
    auto it = us.find(3);
    bool exists = us.count(4) > 0;

    // Hash table properties
    std::cout << "Bucket count: " << us.bucket_count() << std::endl;
    std::cout << "Load factor: " << us.load_factor() << std::endl;
    std::cout << "Max load factor: " << us.max_load_factor() << std::endl;

    // Rehash
    us.rehash(100);                        // set minimum bucket count
    us.reserve(1000);                      // set capacity for elements

    // Custom hash for custom types
    struct Person {
        std::string name;
        int age;

        bool operator==(const Person& other) const {
            return name == other.name && age == other.age;
        }
    };

    struct PersonHash {
        size_t operator()(const Person& p) const {
            return std::hash<std::string>{}(p.name) ^
                   (std::hash<int>{}(p.age) << 1);
        }
    };

    std::unordered_set<Person, PersonHash> person_set;
    person_set.emplace("Alice", 30);
    person_set.emplace("Bob", 25);
}
```

#### unordered_map and unordered_multimap

```cpp
#include <unordered_map>

void unordered_map_examples() {
    std::unordered_map<std::string, int> word_count;

    // Same interface as map, but hash-based
    word_count["hello"] = 1;
    word_count["world"] = 2;

    // Better average performance than map for simple keys
    // Worse worst-case performance (hash collisions)

    // Custom hash function
    struct StringHash {
        size_t operator()(const std::string& str) const {
            // Custom hash implementation
            size_t hash = 0;
            for (char c : str) {
                hash = hash * 31 + c;
            }
            return hash;
        }
    };

    std::unordered_map<std::string, int, StringHash> custom_hash_map;

    // Performance tuning
    word_count.max_load_factor(0.75);     // default is usually 1.0
    word_count.reserve(1000);             // avoid rehashing
}
```

## Container Adaptors

### stack

```cpp
#include <stack>
#include <deque>
#include <vector>

void stack_examples() {
    // Default uses deque as underlying container
    std::stack<int> s1;

    // Can specify underlying container
    std::stack<int, std::vector<int>> s2;
    std::stack<int, std::deque<int>> s3;

    // Operations
    s1.push(1);
    s1.push(2);
    s1.push(3);

    while (!s1.empty()) {
        std::cout << s1.top() << " ";      // access top element
        s1.pop();                          // remove top element
    }

    std::cout << "Size: " << s1.size() << std::endl;
}
```

### queue and priority_queue

```cpp
#include <queue>
#include <vector>

void queue_examples() {
    // Regular queue (FIFO)
    std::queue<int> q;

    q.push(1);
    q.push(2);
    q.push(3);

    while (!q.empty()) {
        std::cout << q.front() << " ";     // access front element
        q.pop();                           // remove front element
    }

    // Priority queue (max-heap by default)
    std::priority_queue<int> pq;

    pq.push(3);
    pq.push(1);
    pq.push(4);
    pq.push(2);

    while (!pq.empty()) {
        std::cout << pq.top() << " ";      // outputs: 4 3 2 1
        pq.pop();
    }

    // Min-heap priority queue
    std::priority_queue<int, std::vector<int>, std::greater<int>> min_pq;

    // Custom comparator
    auto comp = [](const std::pair<int, std::string>& a,
                   const std::pair<int, std::string>& b) {
        return a.first > b.first;  // min-heap by first element
    };

    std::priority_queue<std::pair<int, std::string>,
                       std::vector<std::pair<int, std::string>>,
                       decltype(comp)> custom_pq(comp);

    custom_pq.emplace(3, "three");
    custom_pq.emplace(1, "one");
    custom_pq.emplace(4, "four");
}
```

## Iterators

### Iterator Categories

```cpp
#include <iterator>
#include <vector>
#include <list>
#include <forward_list>

void iterator_categories() {
    std::vector<int> vec{1, 2, 3, 4, 5};
    std::list<int> lst{1, 2, 3, 4, 5};
    std::forward_list<int> flst{1, 2, 3, 4, 5};

    // Random Access Iterator (vector, deque, array)
    auto vec_it = vec.begin();
    vec_it += 3;                           // jump to position
    vec_it -= 1;                           // jump backwards
    int value = vec_it[2];                 // random access
    auto distance = vec.end() - vec.begin(); // O(1) distance

    // Bidirectional Iterator (list, set, map)
    auto lst_it = lst.begin();
    ++lst_it;                              // forward
    --lst_it;                              // backward
    // lst_it += 3;                        // NOT allowed

    // Forward Iterator (forward_list, unordered_set)
    auto flst_it = flst.begin();
    ++flst_it;                             // forward only
    // --flst_it;                          // NOT allowed

    // Input Iterator (istream_iterator)
    std::istream_iterator<int> input_it(std::cin);
    int read_value = *input_it++;          // read and advance

    // Output Iterator (ostream_iterator)
    std::ostream_iterator<int> output_it(std::cout, " ");
    *output_it++ = 42;                     // write and advance
}
```

### Iterator Adaptors

```cpp
#include <iterator>
#include <algorithm>

void iterator_adaptors() {
    std::vector<int> vec{1, 2, 3, 4, 5};

    // Reverse iterator
    std::reverse_iterator<std::vector<int>::iterator> rit(vec.end());
    // or simply:
    auto rit2 = vec.rbegin();

    for (auto it = vec.rbegin(); it != vec.rend(); ++it) {
        std::cout << *it << " ";           // prints: 5 4 3 2 1
    }

    // Insert iterators
    std::vector<int> dest;

    // Back inserter
    std::back_insert_iterator<std::vector<int>> back_it(dest);
    *back_it++ = 10;
    // or using helper:
    auto back_inserter = std::back_inserter(dest);
    *back_inserter++ = 20;

    // Front inserter (for containers that support push_front)
    std::list<int> dest_list;
    auto front_inserter = std::front_inserter(dest_list);
    *front_inserter++ = 30;

    // Insert iterator (for any position)
    auto insert_it = std::inserter(vec, vec.begin() + 2);
    *insert_it++ = 99;

    // Move iterator (C++11)
    std::vector<std::string> source{"hello", "world", "foo", "bar"};
    std::vector<std::string> destination;

    std::move(std::make_move_iterator(source.begin()),
              std::make_move_iterator(source.end()),
              std::back_inserter(destination));

    // source strings are now in moved-from state
}
```

### Custom Iterators

```cpp
#include <iterator>

// Range-based iterator
template<typename T>
class Range {
private:
    T start, end, step;

public:
    Range(T end) : start(0), end(end), step(1) {}
    Range(T start, T end, T step = 1) : start(start), end(end), step(step) {}

    class iterator {
    private:
        T current;
        T step;

    public:
        using iterator_category = std::forward_iterator_tag;
        using value_type = T;
        using difference_type = std::ptrdiff_t;
        using pointer = T*;
        using reference = T&;

        iterator(T current, T step) : current(current), step(step) {}

        T operator*() const { return current; }

        iterator& operator++() {
            current += step;
            return *this;
        }

        iterator operator++(int) {
            iterator temp = *this;
            ++(*this);
            return temp;
        }

        bool operator==(const iterator& other) const {
            return current >= other.current;
        }

        bool operator!=(const iterator& other) const {
            return !(*this == other);
        }
    };

    iterator begin() { return iterator(start, step); }
    iterator end() { return iterator(end, step); }
};

void custom_iterator_example() {
    // Use custom range iterator
    for (int i : Range(0, 10, 2)) {
        std::cout << i << " ";             // prints: 0 2 4 6 8
    }

    for (int i : Range(10)) {
        std::cout << i << " ";             // prints: 0 1 2 3 4 5 6 7 8 9
    }
}
```

## STL Algorithms

### Non-Modifying Sequence Operations

```cpp
#include <algorithm>
#include <numeric>

void non_modifying_algorithms() {
    std::vector<int> vec{1, 2, 3, 2, 4, 5, 2, 6};

    // Find operations
    auto it = std::find(vec.begin(), vec.end(), 3);
    auto it2 = std::find_if(vec.begin(), vec.end(), [](int x) { return x > 4; });
    auto it3 = std::find_if_not(vec.begin(), vec.end(), [](int x) { return x < 10; });

    // Count operations
    size_t count_2 = std::count(vec.begin(), vec.end(), 2);
    size_t count_even = std::count_if(vec.begin(), vec.end(), [](int x) { return x % 2 == 0; });

    // All/Any/None operations
    bool all_positive = std::all_of(vec.begin(), vec.end(), [](int x) { return x > 0; });
    bool any_even = std::any_of(vec.begin(), vec.end(), [](int x) { return x % 2 == 0; });
    bool none_negative = std::none_of(vec.begin(), vec.end(), [](int x) { return x < 0; });

    // For each
    std::for_each(vec.begin(), vec.end(), [](int x) { std::cout << x << " "; });

    // Adjacent find
    auto adj_it = std::adjacent_find(vec.begin(), vec.end());

    // Search operations
    std::vector<int> pattern{4, 5};
    auto search_it = std::search(vec.begin(), vec.end(), pattern.begin(), pattern.end());

    // Mismatch
    std::vector<int> other{1, 2, 3, 9, 4, 5, 2, 6};
    auto [it1, it2] = std::mismatch(vec.begin(), vec.end(), other.begin());

    // Equal
    bool equal = std::equal(vec.begin(), vec.end(), other.begin());

    // Lexicographical compare
    std::vector<int> smaller{1, 2, 2};
    bool is_less = std::lexicographical_compare(smaller.begin(), smaller.end(),
                                               vec.begin(), vec.end());
}
```

### Modifying Sequence Operations

```cpp
#include <algorithm>

void modifying_algorithms() {
    std::vector<int> vec{1, 2, 3, 4, 5};
    std::vector<int> dest(vec.size());

    // Copy operations
    std::copy(vec.begin(), vec.end(), dest.begin());
    std::copy_if(vec.begin(), vec.end(), dest.begin(), [](int x) { return x % 2 == 0; });
    std::copy_n(vec.begin(), 3, dest.begin());
    std::copy_backward(vec.begin(), vec.end(), dest.end());

    // Move operations
    std::vector<std::string> strings{"hello", "world", "foo"};
    std::vector<std::string> moved_strings(strings.size());
    std::move(strings.begin(), strings.end(), moved_strings.begin());

    // Fill operations
    std::fill(vec.begin(), vec.end(), 42);
    std::fill_n(vec.begin(), 3, 99);

    // Generate operations
    std::generate(vec.begin(), vec.end(), []() { return rand() % 100; });
    std::generate_n(vec.begin(), 3, []() { return rand() % 100; });

    // Transform operations
    std::vector<int> source{1, 2, 3, 4, 5};
    std::vector<int> transformed(source.size());

    std::transform(source.begin(), source.end(), transformed.begin(),
                   [](int x) { return x * x; });

    // Binary transform
    std::vector<int> other{2, 3, 4, 5, 6};
    std::transform(source.begin(), source.end(), other.begin(), transformed.begin(),
                   [](int a, int b) { return a + b; });

    // Replace operations
    std::replace(vec.begin(), vec.end(), 42, 0);
    std::replace_if(vec.begin(), vec.end(), [](int x) { return x > 50; }, 100);
    std::replace_copy(vec.begin(), vec.end(), dest.begin(), 99, -1);

    // Remove operations (does not actually remove, returns new end)
    auto new_end = std::remove(vec.begin(), vec.end(), 99);
    vec.erase(new_end, vec.end());  // actually remove

    auto new_end2 = std::remove_if(vec.begin(), vec.end(), [](int x) { return x < 0; });
    vec.erase(new_end2, vec.end());

    // Unique (remove consecutive duplicates)
    std::vector<int> with_dups{1, 1, 2, 2, 2, 3, 1, 1};
    auto unique_end = std::unique(with_dups.begin(), with_dups.end());
    with_dups.erase(unique_end, with_dups.end());

    // Reverse
    std::reverse(vec.begin(), vec.end());
    std::reverse_copy(vec.begin(), vec.end(), dest.begin());

    // Rotate
    std::rotate(vec.begin(), vec.begin() + 2, vec.end());

    // Shuffle
    std::random_device rd;
    std::mt19937 g(rd());
    std::shuffle(vec.begin(), vec.end(), g);
}
```

### Sorting and Related Operations

```cpp
#include <algorithm>
#include <random>

void sorting_algorithms() {
    std::vector<int> vec{3, 1, 4, 1, 5, 9, 2, 6, 5, 3};

    // Sorting
    std::sort(vec.begin(), vec.end());
    std::sort(vec.begin(), vec.end(), std::greater<int>()); // descending

    // Stable sort (maintains relative order of equal elements)
    std::stable_sort(vec.begin(), vec.end());

    // Partial sort
    std::partial_sort(vec.begin(), vec.begin() + 3, vec.end());

    // Nth element (partially sorts to put nth element in correct position)
    std::nth_element(vec.begin(), vec.begin() + vec.size()/2, vec.end());
    int median = vec[vec.size()/2];

    // Binary search operations (require sorted range)
    std::sort(vec.begin(), vec.end());

    bool found = std::binary_search(vec.begin(), vec.end(), 5);

    auto lower = std::lower_bound(vec.begin(), vec.end(), 5);
    auto upper = std::upper_bound(vec.begin(), vec.end(), 5);
    auto [equal_lower, equal_upper] = std::equal_range(vec.begin(), vec.end(), 5);

    // Heap operations
    std::make_heap(vec.begin(), vec.end());

    vec.push_back(10);
    std::push_heap(vec.begin(), vec.end());

    std::pop_heap(vec.begin(), vec.end());
    vec.pop_back();

    std::sort_heap(vec.begin(), vec.end());

    // Min/Max operations
    auto [min_it, max_it] = std::minmax_element(vec.begin(), vec.end());

    // Set operations (require sorted ranges)
    std::vector<int> set1{1, 2, 3, 4, 5};
    std::vector<int> set2{3, 4, 5, 6, 7};
    std::vector<int> result;

    std::set_union(set1.begin(), set1.end(),
                   set2.begin(), set2.end(),
                   std::back_inserter(result));

    result.clear();
    std::set_intersection(set1.begin(), set1.end(),
                         set2.begin(), set2.end(),
                         std::back_inserter(result));

    result.clear();
    std::set_difference(set1.begin(), set1.end(),
                       set2.begin(), set2.end(),
                       std::back_inserter(result));

    // Permutations
    std::vector<int> perm{1, 2, 3};
    do {
        for (int x : perm) std::cout << x << " ";
        std::cout << std::endl;
    } while (std::next_permutation(perm.begin(), perm.end()));
}
```

### Numeric Algorithms

```cpp
#include <numeric>

void numeric_algorithms() {
    std::vector<int> vec{1, 2, 3, 4, 5};

    // Accumulate (reduce)
    int sum = std::accumulate(vec.begin(), vec.end(), 0);
    int product = std::accumulate(vec.begin(), vec.end(), 1, std::multiplies<int>());

    // Reduce (C++17) - parallel version of accumulate
    int parallel_sum = std::reduce(std::execution::par, vec.begin(), vec.end(), 0);

    // Transform reduce
    std::vector<int> other{2, 3, 4, 5, 6};
    int dot_product = std::transform_reduce(vec.begin(), vec.end(),
                                          other.begin(), 0);

    // Inner product
    int inner_prod = std::inner_product(vec.begin(), vec.end(),
                                       other.begin(), 0);

    // Partial sum
    std::vector<int> partial_sums(vec.size());
    std::partial_sum(vec.begin(), vec.end(), partial_sums.begin());

    // Adjacent difference
    std::vector<int> differences(vec.size());
    std::adjacent_difference(vec.begin(), vec.end(), differences.begin());

    // Iota (fill with incrementing values)
    std::vector<int> sequence(10);
    std::iota(sequence.begin(), sequence.end(), 1);  // 1, 2, 3, ..., 10

    // GCD and LCM (C++17)
    int gcd_result = std::gcd(12, 18);  // 6
    int lcm_result = std::lcm(12, 18);  // 36
}
```

## Modern Collection Utilities

### Range-based for loops (C++11)

```cpp
void range_based_examples() {
    std::vector<int> vec{1, 2, 3, 4, 5};

    // Basic range-based for
    for (int x : vec) {
        std::cout << x << " ";
    }

    // By reference (can modify)
    for (int& x : vec) {
        x *= 2;
    }

    // By const reference (efficient, no copy)
    for (const int& x : vec) {
        std::cout << x << " ";
    }

    // Auto type deduction
    for (auto x : vec) {          // copy
        std::cout << x << " ";
    }

    for (auto& x : vec) {         // reference
        x += 1;
    }

    for (const auto& x : vec) {   // const reference
        std::cout << x << " ";
    }

    // Works with initializer lists
    for (auto x : {1, 2, 3, 4, 5}) {
        std::cout << x << " ";
    }

    // Works with C-style arrays
    int arr[] = {1, 2, 3, 4, 5};
    for (int x : arr) {
        std::cout << x << " ";
    }
}
```

### Structured Bindings (C++17)

```cpp
void structured_binding_examples() {
    // With pairs
    std::map<std::string, int> word_count = {{"hello", 5}, {"world", 3}};

    for (const auto& [word, count] : word_count) {
        std::cout << word << ": " << count << std::endl;
    }

    // With tuples
    std::tuple<int, std::string, double> data{42, "hello", 3.14};
    auto [number, text, value] = data;

    // With arrays
    int arr[3] = {1, 2, 3};
    auto [a, b, c] = arr;

    // With custom types
    struct Point {
        int x, y;
    };

    Point p{10, 20};
    auto [x, y] = p;

    // Function returning multiple values
    auto get_min_max(const std::vector<int>& vec) {
        auto [min_it, max_it] = std::minmax_element(vec.begin(), vec.end());
        return std::make_pair(*min_it, *max_it);
    }

    std::vector<int> numbers{3, 1, 4, 1, 5};
    auto [min_val, max_val] = get_min_max(numbers);
}
```

### std::optional (C++17)

```cpp
#include <optional>

std::optional<int> safe_divide(int a, int b) {
    if (b == 0) {
        return std::nullopt;
    }
    return a / b;
}

std::optional<std::string> find_user(int id) {
    // Simulate database lookup
    if (id == 42) {
        return "Alice";
    }
    return std::nullopt;
}

void optional_examples() {
    // Basic usage
    auto result = safe_divide(10, 2);
    if (result) {
        std::cout << "Result: " << *result << std::endl;
    } else {
        std::cout << "Division by zero!" << std::endl;
    }

    // Using has_value()
    if (result.has_value()) {
        std::cout << "Value: " << result.value() << std::endl;
    }

    // Using value_or()
    int safe_result = safe_divide(10, 0).value_or(-1);

    // Chaining operations
    auto user = find_user(42);
    if (user) {
        std::cout << "Found user: " << *user << std::endl;
    }

    // Transform operation (C++23 proposal, might not be available)
    // auto length = user.transform([](const std::string& s) { return s.length(); });
}
```

### std::variant (C++17)

```cpp
#include <variant>

void variant_examples() {
    std::variant<int, std::string, double> value;

    // Assignment
    value = 42;
    value = std::string("hello");
    value = 3.14;

    // Access by index
    try {
        int i = std::get<int>(value);       // throws if wrong type
    } catch (const std::bad_variant_access& e) {
        std::cout << "Wrong type access" << std::endl;
    }

    // Safe access
    if (std::holds_alternative<double>(value)) {
        double d = std::get<double>(value);
        std::cout << "Double value: " << d << std::endl;
    }

    // Visitor pattern
    std::visit([](auto&& arg) {
        using T = std::decay_t<decltype(arg)>;
        if constexpr (std::is_same_v<T, int>) {
            std::cout << "Integer: " << arg << std::endl;
        } else if constexpr (std::is_same_v<T, std::string>) {
            std::cout << "String: " << arg << std::endl;
        } else if constexpr (std::is_same_v<T, double>) {
            std::cout << "Double: " << arg << std::endl;
        }
    }, value);

    // Generic visitor
    auto visitor = [](const auto& v) {
        std::cout << "Value: " << v << std::endl;
    };
    std::visit(visitor, value);
}
```

### std::any (C++17)

```cpp
#include <any>

void any_examples() {
    std::any value;

    // Assignment
    value = 42;
    value = std::string("hello");
    value = 3.14;

    // Type checking
    if (value.type() == typeid(double)) {
        std::cout << "Contains double" << std::endl;
    }

    // Casting
    try {
        double d = std::any_cast<double>(value);
        std::cout << "Double: " << d << std::endl;
    } catch (const std::bad_any_cast& e) {
        std::cout << "Bad cast: " << e.what() << std::endl;
    }

    // Reset
    value.reset();
    if (!value.has_value()) {
        std::cout << "Empty any" << std::endl;
    }

    // In-place construction
    value.emplace<std::vector<int>>(10, 42);
    auto& vec = std::any_cast<std::vector<int>&>(value);
}
```

## Performance Considerations

### Container Selection Guidelines

```cpp
// Quick reference for container selection:

void container_selection_guide() {
    // Sequential access, frequent insertion/deletion at ends
    // -> std::deque

    // Random access, infrequent insertion/deletion
    // -> std::vector

    // Frequent insertion/deletion in middle, no random access needed
    // -> std::list

    // Small, fixed size, known at compile time
    // -> std::array

    // Unique elements, need ordering
    // -> std::set

    // Key-value pairs, need ordering
    // -> std::map

    // Unique elements, don't need ordering, frequent lookups
    // -> std::unordered_set

    // Key-value pairs, don't need ordering, frequent lookups
    // -> std::unordered_map

    // LIFO operations
    // -> std::stack

    // FIFO operations
    // -> std::queue

    // Priority-based access
    // -> std::priority_queue
}

// Memory layout considerations
void memory_layout_examples() {
    // Contiguous memory (cache-friendly)
    std::vector<int> vec;           // excellent cache locality
    std::array<int, 100> arr;       // excellent cache locality
    std::deque<int> deq;            // good cache locality

    // Node-based (pointer chasing, cache-unfriendly)
    std::list<int> list;            // poor cache locality
    std::set<int> set;              // poor cache locality
    std::map<int, int> map;         // poor cache locality

    // Hash-based (depends on implementation)
    std::unordered_set<int> uset;   // variable cache locality
    std::unordered_map<int, int> umap; // variable cache locality
}
```

This comprehensive guide covers the C++ type system and STL collections, providing practical examples and modern best practices for production use.