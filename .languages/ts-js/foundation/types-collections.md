# Types and Collections in TypeScript/JavaScript

## Table of Contents
- [JavaScript Primitive Types](#javascript-primitive-types)
- [JavaScript Object Types](#javascript-object-types)
- [TypeScript Type System](#typescript-type-system)
- [Advanced TypeScript Types](#advanced-typescript-types)
- [Collections and Data Structures](#collections-and-data-structures)
- [Specialized Collections](#specialized-collections)
- [Type Guards and Narrowing](#type-guards-and-narrowing)
- [Generic Programming](#generic-programming)

## JavaScript Primitive Types

### The Seven Primitive Types
```javascript
// 1. Number - IEEE 754 double precision floating point
const integer = 42;
const float = 3.14159;
const scientific = 1.23e10;
const hex = 0xFF; // 255
const binary = 0b1010; // 10
const octal = 0o755; // 493
const infinity = Infinity;
const negInfinity = -Infinity;
const notANumber = NaN;

// Number methods and properties
Number.isInteger(42); // true
Number.isNaN(NaN); // true
Number.parseFloat('3.14'); // 3.14
Number.parseInt('42', 10); // 42
Number.MAX_SAFE_INTEGER; // 9007199254740991
Number.MIN_SAFE_INTEGER; // -9007199254740991

// 2. String - UTF-16 encoded text
const singleQuote = 'Hello';
const doubleQuote = "World";
const template = `Hello ${singleQuote} ${doubleQuote}`;
const multiline = `
  Line 1
  Line 2
`;

// String methods
'hello'.charAt(1); // 'e'
'hello'.charCodeAt(1); // 101
'hello'.slice(1, 3); // 'el'
'hello'.substring(1, 3); // 'el'
'hello'.toUpperCase(); // 'HELLO'
'hello world'.split(' '); // ['hello', 'world']

// 3. Boolean - true or false
const isTrue = true;
const isFalse = false;

// Falsy values: false, 0, -0, 0n, "", null, undefined, NaN
// Everything else is truthy

// 4. Undefined - declared but not assigned
let undefinedVar;
console.log(undefinedVar); // undefined
console.log(typeof undefinedVar); // 'undefined'

// 5. Null - intentional absence of value
let nullVar = null;
console.log(nullVar); // null
console.log(typeof nullVar); // 'object' (this is a known quirk)

// 6. Symbol - unique identifier
const sym1 = Symbol();
const sym2 = Symbol('description');
const sym3 = Symbol('description');
console.log(sym2 !== sym3); // true - each Symbol is unique

// Global symbol registry
const globalSym1 = Symbol.for('mySymbol');
const globalSym2 = Symbol.for('mySymbol');
console.log(globalSym1 === globalSym2); // true

// 7. BigInt - arbitrary precision integers
const bigInt1 = 123n;
const bigInt2 = BigInt(123);
const bigInt3 = BigInt('123456789012345678901234567890');

// BigInt operations
const result = bigInt1 + bigInt2; // 246n
// Note: Cannot mix BigInt with regular numbers
// bigInt1 + 123; // TypeError!
const mixed = bigInt1 + BigInt(123); // OK: 246n
```

### Type Coercion and Conversion
```javascript
// Implicit type coercion
console.log('5' + 3); // '53' (string concatenation)
console.log('5' - 3); // 2 (numeric subtraction)
console.log(+'123'); // 123 (unary plus converts to number)
console.log(!!'hello'); // true (double negation converts to boolean)

// Explicit type conversion
String(123); // '123'
Number('123'); // 123
Boolean(1); // true

// Parsing
parseInt('123px'); // 123
parseFloat('123.45px'); // 123.45

// JSON conversion
const obj = { name: 'John', age: 30 };
const jsonString = JSON.stringify(obj); // '{"name":"John","age":30}'
const parsed = JSON.parse(jsonString); // { name: 'John', age: 30 }
```

## JavaScript Object Types

### Object Fundamentals
```javascript
// Object creation methods
const obj1 = {}; // Object literal
const obj2 = new Object(); // Constructor
const obj3 = Object.create(null); // No prototype
const obj4 = Object.create(Object.prototype); // Explicit prototype

// Property access
const person = {
  name: 'John',
  age: 30,
  'complex-key': 'value',
  123: 'numeric key'
};

person.name; // 'John' (dot notation)
person['age']; // 30 (bracket notation)
person['complex-key']; // 'value' (required for complex keys)
person[123]; // 'numeric key'

// Dynamic property access
const prop = 'name';
person[prop]; // 'John'

// Property descriptors
Object.defineProperty(person, 'id', {
  value: 1,
  writable: false,
  enumerable: false,
  configurable: false
});

Object.getOwnPropertyDescriptor(person, 'name');
// { value: 'John', writable: true, enumerable: true, configurable: true }
```

### Advanced Object Features
```javascript
// Getters and setters
const user = {
  firstName: 'John',
  lastName: 'Doe',
  
  get fullName() {
    return `${this.firstName} ${this.lastName}`;
  },
  
  set fullName(value) {
    [this.firstName, this.lastName] = value.split(' ');
  }
};

console.log(user.fullName); // 'John Doe'
user.fullName = 'Jane Smith';
console.log(user.firstName); // 'Jane'

// Computed property names
const prefix = 'user';
const dynamicObj = {
  [`${prefix}Name`]: 'John',
  [`${prefix}Age`]: 30
};

// Object methods
Object.keys(person); // ['name', 'age', 'complex-key', '123']
Object.values(person); // ['John', 30, 'value', 'numeric key']
Object.entries(person); // [['name', 'John'], ['age', 30], ...]

Object.assign({}, person, { city: 'New York' }); // Shallow copy with additional props
Object.freeze(person); // Make immutable
Object.seal(person); // Prevent property addition/deletion
```

## TypeScript Type System

### Basic Type Annotations
```typescript
// Primitive type annotations
let name: string = 'John';
let age: number = 30;
let isActive: boolean = true;
let value: null = null;
let undefinedValue: undefined = undefined;
let id: symbol = Symbol('id');
let bigNumber: bigint = 123n;

// Any type (avoid when possible)
let anything: any = 42;
anything = 'hello';
anything = true;

// Unknown type (safer than any)
let something: unknown = 42;
if (typeof something === 'string') {
  console.log(something.toUpperCase()); // Type narrowing required
}

// Never type (for functions that never return)
function throwError(message: string): never {
  throw new Error(message);
}

// Void type (for functions that return nothing)
function logMessage(message: string): void {
  console.log(message);
}
```

### Array and Tuple Types
```typescript
// Array types
const numbers: number[] = [1, 2, 3];
const strings: Array<string> = ['a', 'b', 'c'];
const mixed: (string | number)[] = ['hello', 42, 'world'];

// Readonly arrays
const readonlyNumbers: readonly number[] = [1, 2, 3];
const readonlyStrings: ReadonlyArray<string> = ['a', 'b', 'c'];
// readonlyNumbers.push(4); // Error: push doesn't exist on readonly array

// Tuple types
const coordinates: [number, number] = [10, 20];
const nameAndAge: [string, number] = ['John', 30];

// Named tuple elements
const namedCoordinates: [x: number, y: number] = [10, 20];

// Optional tuple elements
const optionalTuple: [string, number?] = ['John'];

// Rest elements in tuples
const restTuple: [string, ...number[]] = ['John', 1, 2, 3, 4];

// Tuple with specific length
type FixedArray<T, N extends number> = T[] & { length: N };
const fixedNumbers: FixedArray<number, 3> = [1, 2, 3];
```

### Object and Interface Types
```typescript
// Object type annotations
const person: {
  name: string;
  age: number;
  email?: string; // Optional property
  readonly id: number; // Readonly property
} = {
  name: 'John',
  age: 30,
  id: 1
};

// Interface declarations
interface User {
  readonly id: number;
  name: string;
  age: number;
  email?: string;
  preferences: {
    theme: 'light' | 'dark';
    notifications: boolean;
  };
}

// Interface extension
interface AdminUser extends User {
  permissions: string[];
  lastLogin: Date;
}

// Interface merging (declaration merging)
interface User {
  createdAt: Date;
}
// Now User interface includes createdAt property

// Index signatures
interface StringDictionary {
  [key: string]: string;
}

interface NumberDictionary {
  [key: string]: number;
  length: number; // OK, length is a number
  name: string; // Error, conflicts with index signature
}

// Method signatures
interface Calculator {
  add(a: number, b: number): number;
  subtract(a: number, b: number): number;
  multiply: (a: number, b: number) => number; // Alternative syntax
}
```

### Function Types
```typescript
// Function type annotations
const add: (a: number, b: number) => number = (a, b) => a + b;

// Function declarations with types
function multiply(a: number, b: number): number {
  return a * b;
}

// Optional parameters
function greet(name: string, greeting?: string): string {
  return `${greeting || 'Hello'}, ${name}!`;
}

// Default parameters
function createUser(name: string, age: number = 0): User {
  return { id: Date.now(), name, age };
}

// Rest parameters
function sum(...numbers: number[]): number {
  return numbers.reduce((total, num) => total + num, 0);
}

// Function overloads
function parseValue(input: string): string;
function parseValue(input: number): number;
function parseValue(input: boolean): boolean;
function parseValue(input: string | number | boolean): string | number | boolean {
  return input;
}

// Generic functions
function identity<T>(arg: T): T {
  return arg;
}

const stringResult = identity<string>('hello'); // Type: string
const numberResult = identity(42); // Type inferred as number
```

## Advanced TypeScript Types

### Union and Intersection Types
```typescript
// Union types
type StringOrNumber = string | number;
type Status = 'pending' | 'approved' | 'rejected';

function printValue(value: string | number): void {
  if (typeof value === 'string') {
    console.log(value.toUpperCase()); // TypeScript knows it's a string
  } else {
    console.log(value.toFixed(2)); // TypeScript knows it's a number
  }
}

// Discriminated unions
type Shape =
  | { kind: 'circle'; radius: number }
  | { kind: 'rectangle'; width: number; height: number }
  | { kind: 'triangle'; base: number; height: number };

function calculateArea(shape: Shape): number {
  switch (shape.kind) {
    case 'circle':
      return Math.PI * shape.radius ** 2;
    case 'rectangle':
      return shape.width * shape.height;
    case 'triangle':
      return (shape.base * shape.height) / 2;
    default:
      // Exhaustiveness check
      const _exhaustive: never = shape;
      return _exhaustive;
  }
}

// Intersection types
type Person = { name: string; age: number };
type Employee = { employeeId: string; department: string };
type PersonEmployee = Person & Employee;

const employee: PersonEmployee = {
  name: 'John',
  age: 30,
  employeeId: 'EMP001',
  department: 'Engineering'
};
```

### Literal and Template Literal Types
```typescript
// Literal types
type Direction = 'north' | 'south' | 'east' | 'west';
type DiceRoll = 1 | 2 | 3 | 4 | 5 | 6;
type EventType = 'click' | 'scroll' | 'resize';

// Template literal types
type Greeting = `hello-${string}`;
type EventName<T extends string> = `on${Capitalize<T>}`;
type ClickEvent = EventName<'click'>; // 'onClick'

// Template literal patterns
type CSSProperty = `--${string}`;
type DataAttribute = `data-${string}`;

type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'DELETE';
type Endpoint = `/api/${string}`;
type APIRoute = `${HTTPMethod} ${Endpoint}`;

// Advanced template literal types
type Split<S extends string, D extends string> =
  S extends `${infer T}${D}${infer U}` ?
    [T, ...Split<U, D>] :
    [S];

type SplitResult = Split<'a-b-c', '-'>; // ['a', 'b', 'c']
```

### Mapped Types
```typescript
// Built-in mapped types
type PartialUser = Partial<User>; // All properties optional
type RequiredUser = Required<User>; // All properties required
type ReadonlyUser = Readonly<User>; // All properties readonly
type UserKeys = keyof User; // Union of all property names
type UserName = Pick<User, 'name'>; // Extract specific properties
type UserWithoutId = Omit<User, 'id'>; // Exclude specific properties

// Custom mapped types
type Nullable<T> = {
  [P in keyof T]: T[P] | null;
};

type Optional<T, K extends keyof T> = {
  [P in Exclude<keyof T, K>]: T[P];
} & {
  [P in K]?: T[P];
};

// Advanced mapped types
type GetterType<T> = {
  [P in keyof T as `get${Capitalize<string & P>}`]: () => T[P];
};

type UserGetters = GetterType<User>;
// {
//   getName: () => string;
//   getAge: () => number;
//   getId: () => number;
// }

// Conditional mapped types
type NonFunctionPropertyNames<T> = {
  [K in keyof T]: T[K] extends Function ? never : K;
}[keyof T];

type NonFunctionProperties<T> = Pick<T, NonFunctionPropertyNames<T>>;
```

### Conditional Types
```typescript
// Basic conditional types
type IsString<T> = T extends string ? true : false;
type Test1 = IsString<string>; // true
type Test2 = IsString<number>; // false

// Conditional types with infer
type ReturnType<T> = T extends (...args: any[]) => infer R ? R : never;
type FunctionReturn = ReturnType<() => string>; // string

type ArrayElement<T> = T extends (infer U)[] ? U : never;
type Element = ArrayElement<string[]>; // string

// Distributive conditional types
type ToArray<T> = T extends any ? T[] : never;
type StringOrNumberArray = ToArray<string | number>; // string[] | number[]

// Complex conditional type example
type DeepReadonly<T> = {
  readonly [P in keyof T]: T[P] extends object
    ? T[P] extends Function
      ? T[P]
      : DeepReadonly<T[P]>
    : T[P];
};

// Utility conditional types
type NonNullable<T> = T extends null | undefined ? never : T;
type Flatten<T> = T extends Array<infer U> ? U : T;
type PromiseType<T> = T extends Promise<infer U> ? U : never;
```

## Collections and Data Structures

### Array Methods and Operations
```javascript
const numbers = [1, 2, 3, 4, 5];
const strings = ['apple', 'banana', 'cherry'];

// Transformation methods
const doubled = numbers.map(n => n * 2); // [2, 4, 6, 8, 10]
const uppercased = strings.map(s => s.toUpperCase()); // ['APPLE', 'BANANA', 'CHERRY']

// Filtering methods
const evens = numbers.filter(n => n % 2 === 0); // [2, 4]
const longWords = strings.filter(s => s.length > 5); // ['banana', 'cherry']

// Reduction methods
const sum = numbers.reduce((acc, n) => acc + n, 0); // 15
const max = numbers.reduce((acc, n) => Math.max(acc, n), -Infinity); // 5

// Finding methods
const firstEven = numbers.find(n => n % 2 === 0); // 2
const evenIndex = numbers.findIndex(n => n % 2 === 0); // 1
const hasEven = numbers.some(n => n % 2 === 0); // true
const allPositive = numbers.every(n => n > 0); // true

// Iteration methods
numbers.forEach((n, index) => console.log(`${index}: ${n}`));

// Array modification methods
const fruits = ['apple', 'banana'];
fruits.push('cherry'); // Add to end
fruits.unshift('grape'); // Add to beginning
const last = fruits.pop(); // Remove from end
const first = fruits.shift(); // Remove from beginning

// Splice method (add/remove at any position)
const items = [1, 2, 3, 4, 5];
items.splice(2, 1, 'inserted'); // Remove 1 item at index 2, insert 'inserted'
// items is now [1, 2, 'inserted', 4, 5]

// Array copying and combining
const copy = [...numbers]; // Shallow copy
const combined = [...numbers, ...strings]; // Combine arrays
const sliced = numbers.slice(1, 3); // [2, 3] (shallow copy of portion)

// Sorting
const sorted = [...numbers].sort((a, b) => a - b); // Ascending
const reversed = [...numbers].reverse(); // Reverse order
```

### TypeScript Array Types
```typescript
// Typed arrays
const typedNumbers: number[] = [1, 2, 3];
const typedStrings: Array<string> = ['a', 'b', 'c'];

// Array with union types
const mixed: (string | number | boolean)[] = ['hello', 42, true];

// Readonly arrays
const readonlyArray: ReadonlyArray<number> = [1, 2, 3];
// readonlyArray.push(4); // Error!

// Array utility types
type ArrayElement<T> = T extends ReadonlyArray<infer U> ? U : never;
type NumberArrayElement = ArrayElement<number[]>; // number

// Tuple operations
type Head<T extends readonly any[]> = T extends readonly [infer H, ...any[]] ? H : never;
type Tail<T extends readonly any[]> = T extends readonly [any, ...infer T] ? T : [];

type FirstNumber = Head<[1, 2, 3]>; // 1
type RestNumbers = Tail<[1, 2, 3]>; // [2, 3]

// Array methods with proper typing
function processNumbers(numbers: number[]): {
  doubled: number[];
  evens: number[];
  sum: number;
} {
  return {
    doubled: numbers.map(n => n * 2),
    evens: numbers.filter(n => n % 2 === 0),
    sum: numbers.reduce((acc, n) => acc + n, 0)
  };
}
```

### Set Collection
```javascript
// Set creation and basic operations
const uniqueNumbers = new Set([1, 2, 3, 3, 4, 4, 5]); // {1, 2, 3, 4, 5}
const emptySet = new Set();

// Set methods
uniqueNumbers.add(6); // Add element
uniqueNumbers.has(3); // true - check existence
uniqueNumbers.delete(2); // Remove element
uniqueNumbers.size; // Get size
uniqueNumbers.clear(); // Remove all elements

// Set iteration
const fruits = new Set(['apple', 'banana', 'cherry']);

// for...of loop
for (const fruit of fruits) {
  console.log(fruit);
}

// forEach method
fruits.forEach(fruit => console.log(fruit));

// Convert to array
const fruitsArray = [...fruits];
const fruitsArray2 = Array.from(fruits);

// Set operations
function union<T>(setA: Set<T>, setB: Set<T>): Set<T> {
  return new Set([...setA, ...setB]);
}

function intersection<T>(setA: Set<T>, setB: Set<T>): Set<T> {
  return new Set([...setA].filter(x => setB.has(x)));
}

function difference<T>(setA: Set<T>, setB: Set<T>): Set<T> {
  return new Set([...setA].filter(x => !setB.has(x)));
}

// Example usage
const set1 = new Set([1, 2, 3]);
const set2 = new Set([3, 4, 5]);

const unionSet = union(set1, set2); // {1, 2, 3, 4, 5}
const intersectionSet = intersection(set1, set2); // {3}
const differenceSet = difference(set1, set2); // {1, 2}
```

### Map Collection
```javascript
// Map creation and basic operations
const userMap = new Map([
  ['user1', { name: 'John', age: 30 }],
  ['user2', { name: 'Jane', age: 25 }]
]);

const emptyMap = new Map();

// Map methods
userMap.set('user3', { name: 'Bob', age: 35 }); // Add/update
const user1 = userMap.get('user1'); // Get value
const hasUser2 = userMap.has('user2'); // true - check existence
userMap.delete('user3'); // Remove entry
const mapSize = userMap.size; // Get size

// Map iteration
for (const [key, value] of userMap) {
  console.log(`${key}: ${value.name}`);
}

// Iterate over keys
for (const key of userMap.keys()) {
  console.log(key);
}

// Iterate over values
for (const value of userMap.values()) {
  console.log(value.name);
}

// forEach method
userMap.forEach((value, key) => {
  console.log(`${key}: ${value.name}`);
});

// Convert to arrays
const entries = [...userMap.entries()]; // [[key, value], ...]
const keys = [...userMap.keys()];
const values = [...userMap.values()];

// Object keys vs Map keys
const objAsMap = new Map(Object.entries({
  name: 'John',
  age: 30
}));

// Map can use any type as key
const complexMap = new Map();
const keyObj = { id: 1 };
const keyFunc = () => {};
const keyDate = new Date();

complexMap.set(keyObj, 'object key');
complexMap.set(keyFunc, 'function key');
complexMap.set(keyDate, 'date key');
complexMap.set(1, 'number key');
complexMap.set('1', 'string key');
```

### TypeScript Map and Set Types
```typescript
// Typed Map
const userMap: Map<string, User> = new Map();
userMap.set('user1', { id: 1, name: 'John', age: 30 });

// Typed Set
const uniqueIds: Set<number> = new Set([1, 2, 3]);

// Generic collections
class TypedCollection<K, V> {
  private map = new Map<K, V>();
  private keySet = new Set<K>();

  set(key: K, value: V): void {
    this.map.set(key, value);
    this.keySet.add(key);
  }

  get(key: K): V | undefined {
    return this.map.get(key);
  }

  has(key: K): boolean {
    return this.map.has(key);
  }

  keys(): Set<K> {
    return new Set(this.keySet);
  }

  values(): V[] {
    return [...this.map.values()];
  }

  size(): number {
    return this.map.size;
  }
}

// Usage
const collection = new TypedCollection<string, User>();
collection.set('user1', { id: 1, name: 'John', age: 30 });
```

## Specialized Collections

### WeakMap and WeakSet
```javascript
// WeakMap - keys must be objects, automatically garbage collected
const privateData = new WeakMap();

class User {
  constructor(name) {
    this.name = name;
    // Store private data using WeakMap
    privateData.set(this, {
      id: Math.random(),
      createdAt: new Date()
    });
  }

  getId() {
    return privateData.get(this)?.id;
  }

  getCreatedAt() {
    return privateData.get(this)?.createdAt;
  }
}

const user = new User('John');
console.log(user.getId()); // Access private data
// When user is garbage collected, private data is automatically removed

// WeakSet - values must be objects, automatically garbage collected
const visitedNodes = new WeakSet();

function traverseDOM(node) {
  if (visitedNodes.has(node)) {
    return; // Avoid infinite loops
  }

  visitedNodes.add(node);

  // Process node
  console.log(node.tagName);

  // Traverse children
  for (const child of node.children) {
    traverseDOM(child);
  }
}

// WeakMap for metadata
const elementMetadata = new WeakMap();

function attachMetadata(element, metadata) {
  elementMetadata.set(element, metadata);
}

function getMetadata(element) {
  return elementMetadata.get(element);
}

// When elements are removed from DOM, metadata is automatically cleaned up
```

### Typed Arrays
```javascript
// Typed arrays for binary data
const buffer = new ArrayBuffer(16); // 16 bytes
const int8View = new Int8Array(buffer); // 8-bit signed integers
const int16View = new Int16Array(buffer); // 16-bit signed integers
const int32View = new Int32Array(buffer); // 32-bit signed integers
const uint8View = new Uint8Array(buffer); // 8-bit unsigned integers
const float32View = new Float32Array(buffer); // 32-bit floats
const float64View = new Float64Array(buffer); // 64-bit floats

// Direct creation
const directInt32 = new Int32Array([1, 2, 3, 4]);
const directFloat32 = new Float32Array(10); // 10 elements, initialized to 0

// Typed array methods (similar to regular arrays)
const numbers = new Float32Array([1.1, 2.2, 3.3, 4.4]);
const doubled = numbers.map(x => x * 2);
const filtered = numbers.filter(x => x > 2);
const sum = numbers.reduce((acc, x) => acc + x, 0);

// Binary data manipulation
function createWaveform(frequency, sampleRate, duration) {
  const samples = sampleRate * duration;
  const waveform = new Float32Array(samples);

  for (let i = 0; i < samples; i++) {
    const t = i / sampleRate;
    waveform[i] = Math.sin(2 * Math.PI * frequency * t);
  }

  return waveform;
}

const wave = createWaveform(440, 44100, 1); // 440Hz for 1 second at 44.1kHz
```

### Custom Collection Classes
```typescript
// Generic Stack implementation
class Stack<T> {
  private items: T[] = [];

  push(item: T): void {
    this.items.push(item);
  }

  pop(): T | undefined {
    return this.items.pop();
  }

  peek(): T | undefined {
    return this.items[this.items.length - 1];
  }

  isEmpty(): boolean {
    return this.items.length === 0;
  }

  size(): number {
    return this.items.length;
  }

  toArray(): T[] {
    return [...this.items];
  }

  *[Symbol.iterator](): Iterator<T> {
    for (let i = this.items.length - 1; i >= 0; i--) {
      yield this.items[i];
    }
  }
}

// Generic Queue implementation
class Queue<T> {
  private items: T[] = [];

  enqueue(item: T): void {
    this.items.push(item);
  }

  dequeue(): T | undefined {
    return this.items.shift();
  }

  front(): T | undefined {
    return this.items[0];
  }

  isEmpty(): boolean {
    return this.items.length === 0;
  }

  size(): number {
    return this.items.length;
  }

  *[Symbol.iterator](): Iterator<T> {
    yield* this.items;
  }
}

// LRU Cache implementation
class LRUCache<K, V> {
  private capacity: number;
  private cache = new Map<K, V>();

  constructor(capacity: number) {
    this.capacity = capacity;
  }

  get(key: K): V | undefined {
    if (this.cache.has(key)) {
      // Move to end (most recently used)
      const value = this.cache.get(key)!;
      this.cache.delete(key);
      this.cache.set(key, value);
      return value;
    }
    return undefined;
  }

  set(key: K, value: V): void {
    if (this.cache.has(key)) {
      this.cache.delete(key);
    } else if (this.cache.size >= this.capacity) {
      // Remove least recently used (first item)
      const firstKey = this.cache.keys().next().value;
      this.cache.delete(firstKey);
    }

    this.cache.set(key, value);
  }

  has(key: K): boolean {
    return this.cache.has(key);
  }

  size(): number {
    return this.cache.size;
  }

  clear(): void {
    this.cache.clear();
  }
}

// Usage examples
const stack = new Stack<number>();
stack.push(1);
stack.push(2);
stack.push(3);

for (const item of stack) {
  console.log(item); // 3, 2, 1
}

const queue = new Queue<string>();
queue.enqueue('first');
queue.enqueue('second');
queue.enqueue('third');

const cache = new LRUCache<string, User>(3);
cache.set('user1', { id: 1, name: 'John', age: 30 });
cache.set('user2', { id: 2, name: 'Jane', age: 25 });
```

## Type Guards and Narrowing

### Built-in Type Guards
```typescript
// typeof type guards
function processValue(value: string | number): string {
  if (typeof value === 'string') {
    return value.toUpperCase(); // TypeScript knows it's a string
  } else {
    return value.toString(); // TypeScript knows it's a number
  }
}

// instanceof type guards
class Dog {
  bark(): void {
    console.log('Woof!');
  }
}

class Cat {
  meow(): void {
    console.log('Meow!');
  }
}

function makeSound(animal: Dog | Cat): void {
  if (animal instanceof Dog) {
    animal.bark(); // TypeScript knows it's a Dog
  } else {
    animal.meow(); // TypeScript knows it's a Cat
  }
}

// in operator type guards
type Fish = { swim: () => void };
type Bird = { fly: () => void };

function move(animal: Fish | Bird): void {
  if ('swim' in animal) {
    animal.swim(); // TypeScript knows it's a Fish
  } else {
    animal.fly(); // TypeScript knows it's a Bird
  }
}
```

### Custom Type Guards
```typescript
// Custom type guard functions
function isString(value: unknown): value is string {
  return typeof value === 'string';
}

function isNumber(value: unknown): value is number {
  return typeof value === 'number' && !isNaN(value);
}

function isArray<T>(value: unknown): value is T[] {
  return Array.isArray(value);
}

// Object type guards
interface User {
  id: number;
  name: string;
  email: string;
}

function isUser(obj: unknown): obj is User {
  return (
    typeof obj === 'object' &&
    obj !== null &&
    typeof (obj as User).id === 'number' &&
    typeof (obj as User).name === 'string' &&
    typeof (obj as User).email === 'string'
  );
}

// Generic type guard factory
function hasProperty<T extends object, K extends PropertyKey>(
  obj: T,
  key: K
): obj is T & Record<K, unknown> {
  return key in obj;
}

// Usage
function processUnknown(value: unknown): void {
  if (isString(value)) {
    console.log(value.toUpperCase()); // TypeScript knows it's a string
  } else if (isNumber(value)) {
    console.log(value.toFixed(2)); // TypeScript knows it's a number
  } else if (isUser(value)) {
    console.log(value.name); // TypeScript knows it's a User
  }
}

// Advanced type guards with generics
function isArrayOf<T>(
  value: unknown,
  itemGuard: (item: unknown) => item is T
): value is T[] {
  return Array.isArray(value) && value.every(itemGuard);
}

// Usage
const mixedArray: unknown = [1, 2, 3];
if (isArrayOf(mixedArray, isNumber)) {
  // TypeScript knows mixedArray is number[]
  console.log(mixedArray.map(n => n * 2));
}
```

### Assertion Functions
```typescript
// Assertion functions (TypeScript 3.7+)
function assert(condition: unknown, message?: string): asserts condition {
  if (!condition) {
    throw new Error(message || 'Assertion failed');
  }
}

function assertIsString(value: unknown): asserts value is string {
  if (typeof value !== 'string') {
    throw new Error('Expected string');
  }
}

function assertIsNumber(value: unknown): asserts value is number {
  if (typeof value !== 'number' || isNaN(value)) {
    throw new Error('Expected number');
  }
}

// Usage
function processValue(value: unknown): void {
  assertIsString(value);
  // TypeScript now knows value is a string
  console.log(value.toUpperCase());
}

// Generic assertion function
function assertType<T>(
  value: unknown,
  guard: (value: unknown) => value is T,
  message?: string
): asserts value is T {
  if (!guard(value)) {
    throw new Error(message || 'Type assertion failed');
  }
}

// Usage
function processUser(data: unknown): void {
  assertType(data, isUser, 'Invalid user data');
  // TypeScript now knows data is User
  console.log(data.name);
}
```

## Generic Programming

### Generic Functions and Classes
```typescript
// Generic function with constraints
function getProperty<T extends object, K extends keyof T>(
  obj: T,
  key: K
): T[K] {
  return obj[key];
}

const user = { name: 'John', age: 30, email: 'john@example.com' };
const name = getProperty(user, 'name'); // Type: string
const age = getProperty(user, 'age'); // Type: number

// Generic class with multiple type parameters
class Repository<T, K extends keyof T> {
  private items: T[] = [];
  private keyField: K;

  constructor(keyField: K) {
    this.keyField = keyField;
  }

  add(item: T): void {
    this.items.push(item);
  }

  findById(id: T[K]): T | undefined {
    return this.items.find(item => item[this.keyField] === id);
  }

  findAll(): T[] {
    return [...this.items];
  }

  removeById(id: T[K]): boolean {
    const index = this.items.findIndex(item => item[this.keyField] === id);
    if (index >= 0) {
      this.items.splice(index, 1);
      return true;
    }
    return false;
  }
}

// Usage
const userRepo = new Repository<User, 'id'>('id');
userRepo.add({ id: 1, name: 'John', age: 30 });
const foundUser = userRepo.findById(1); // Type: User | undefined

// Generic constraints with conditional types
type ApiResponse<T> = T extends string
  ? { message: T }
  : T extends number
  ? { value: T }
  : { data: T };

type StringResponse = ApiResponse<string>; // { message: string }
type NumberResponse = ApiResponse<number>; // { value: number }
type ObjectResponse = ApiResponse<User>; // { data: User }

// Generic utility functions
function deepClone<T>(obj: T): T {
  if (obj === null || typeof obj !== 'object') {
    return obj;
  }

  if (obj instanceof Date) {
    return new Date(obj.getTime()) as T;
  }

  if (obj instanceof Array) {
    return obj.map(item => deepClone(item)) as T;
  }

  if (typeof obj === 'object') {
    const cloned = {} as T;
    for (const key in obj) {
      if (obj.hasOwnProperty(key)) {
        cloned[key] = deepClone(obj[key]);
      }
    }
    return cloned;
  }

  return obj;
}

// Generic event emitter
class TypedEventEmitter<T extends Record<string, any[]>> {
  private listeners: {
    [K in keyof T]?: Array<(...args: T[K]) => void>
  } = {};

  on<K extends keyof T>(event: K, listener: (...args: T[K]) => void): void {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event]!.push(listener);
  }

  emit<K extends keyof T>(event: K, ...args: T[K]): void {
    const eventListeners = this.listeners[event];
    if (eventListeners) {
      eventListeners.forEach(listener => listener(...args));
    }
  }

  off<K extends keyof T>(event: K, listener: (...args: T[K]) => void): void {
    const eventListeners = this.listeners[event];
    if (eventListeners) {
      const index = eventListeners.indexOf(listener);
      if (index >= 0) {
        eventListeners.splice(index, 1);
      }
    }
  }
}

// Usage
type Events = {
  userCreated: [User];
  userUpdated: [User, User]; // old, new
  userDeleted: [number]; // id
};

const emitter = new TypedEventEmitter<Events>();

emitter.on('userCreated', (user) => {
  console.log('User created:', user.name); // Fully typed
});

emitter.on('userUpdated', (oldUser, newUser) => {
  console.log(`User updated: ${oldUser.name} -> ${newUser.name}`);
});

emitter.emit('userCreated', { id: 1, name: 'John', age: 30 });
```

This comprehensive guide covers the complete type system and collection types available in TypeScript and JavaScript, providing the foundation for building robust, type-safe applications with efficient data handling and manipulation.