# Unique Features of TypeScript/JavaScript

## JavaScript's Distinctive Features

### Prototypal Inheritance
JavaScript uses prototypal inheritance, which is fundamentally different from class-based inheritance:

```javascript
// Prototype chain example
function Animal(name) {
  this.name = name;
}

Animal.prototype.speak = function() {
  return `${this.name} makes a sound`;
};

function Dog(name, breed) {
  Animal.call(this, name);
  this.breed = breed;
}

// Set up prototype chain
Dog.prototype = Object.create(Animal.prototype);
Dog.prototype.constructor = Dog;

Dog.prototype.speak = function() {
  return `${this.name} barks`;
};

const myDog = new Dog('Rex', 'German Shepherd');
console.log(myDog.speak()); // 'Rex barks'
console.log(myDog.__proto__ === Dog.prototype); // true
console.log(myDog.__proto__.__proto__ === Animal.prototype); // true

// Direct prototype manipulation
const obj = { x: 1 };
const child = Object.create(obj);
child.y = 2;

console.log(child.x); // 1 (inherited)
console.log(child.y); // 2 (own property)

// Prototype pollution (be careful!)
Object.prototype.newMethod = function() {
  return 'added to all objects';
};

console.log({}.newMethod()); // 'added to all objects'
```

### Dynamic This Binding
JavaScript's `this` context is determined by how a function is called, not where it's defined:

```javascript
const obj = {
  name: 'Object',
  regularFunction: function() {
    return this.name; // 'this' depends on call site
  },
  arrowFunction: () => {
    return this.name; // 'this' is lexically bound (usually undefined/window)
  },

  methodWithCallback: function() {
    setTimeout(function() {
      console.log(this.name); // 'this' is window/undefined
    }, 100);

    setTimeout(() => {
      console.log(this.name); // 'this' is obj (lexically bound)
    }, 200);
  }
};

// Different ways to call functions affect 'this'
const fn = obj.regularFunction;
console.log(fn()); // undefined (no object context)
console.log(obj.regularFunction()); // 'Object'
console.log(fn.call(obj)); // 'Object' (explicit binding)
console.log(fn.apply(obj)); // 'Object' (explicit binding)
console.log(fn.bind(obj)()); // 'Object' (permanent binding)

// Method extraction loses context
const { regularFunction } = obj;
console.log(regularFunction()); // undefined

// Arrow functions don't have their own 'this'
const ArrowClass = () => {
  this.value = 42; // Error or unexpected behavior
};

// Regular function constructor
function RegularClass() {
  this.value = 42; // Works as expected
}
```

### Hoisting Behavior
JavaScript moves declarations to the top of their scope during compilation:

```javascript
// Variable hoisting
console.log(x); // undefined (not ReferenceError)
var x = 5;

// Function hoisting
console.log(hoistedFunction()); // 'I am hoisted!'

function hoistedFunction() {
  return 'I am hoisted!';
}

// Let/const temporal dead zone
console.log(y); // ReferenceError: Cannot access 'y' before initialization
let y = 10;

console.log(z); // ReferenceError: Cannot access 'z' before initialization
const z = 20;

// Function expression hoisting
console.log(notHoisted); // undefined
console.log(notHoisted()); // TypeError: notHoisted is not a function

var notHoisted = function() {
  return 'Not hoisted';
};

// Hoisting with loops
for (var i = 0; i < 3; i++) {
  setTimeout(() => console.log(i), 100); // Prints 3, 3, 3
}

for (let j = 0; j < 3; j++) {
  setTimeout(() => console.log(j), 100); // Prints 0, 1, 2
}
```

### Closures and Lexical Scoping
JavaScript functions have access to variables in their lexical scope even after the outer function returns:

```javascript
// Basic closure
function outerFunction(x) {
  return function innerFunction(y) {
    return x + y; // 'x' is captured in closure
  };
}

const addFive = outerFunction(5);
console.log(addFive(3)); // 8

// Closure-based module pattern
const createCounter = () => {
  let count = 0;

  return {
    increment: () => ++count,
    decrement: () => --count,
    get: () => count
  };
};

const counter = createCounter();
console.log(counter.increment()); // 1
console.log(counter.get()); // 1
// count is private, can't be accessed directly

// Closure with loops (classic problem)
const functions = [];
for (var i = 0; i < 3; i++) {
  functions.push(function() {
    return i; // All functions reference the same 'i'
  });
}

functions.forEach(fn => console.log(fn())); // 3, 3, 3

// Solution with closure
const functionsFixed = [];
for (let i = 0; i < 3; i++) {
  functionsFixed.push(function() {
    return i; // Each function captures its own 'i'
  });
}

functionsFixed.forEach(fn => console.log(fn())); // 0, 1, 2

// IIFE solution
const functionsIIFE = [];
for (var i = 0; i < 3; i++) {
  functionsIIFE.push((function(index) {
    return function() {
      return index;
    };
  })(i));
}

functionsIIFE.forEach(fn => console.log(fn())); // 0, 1, 2
```

### Event Loop and Asynchronous Nature
JavaScript's single-threaded event loop with non-blocking I/O:

```javascript
// Event loop demonstration
console.log('1');

setTimeout(() => console.log('2'), 0);

Promise.resolve().then(() => console.log('3'));

console.log('4');

// Output: 1, 4, 3, 2
// Explanation:
// 1. Synchronous code runs first (1, 4)
// 2. Microtasks (Promises) run before macrotasks (setTimeout)
// 3. setTimeout callback runs last

// Microtask vs Macrotask queue
const promise = Promise.resolve();

promise.then(() => console.log('Promise 1'));
promise.then(() => console.log('Promise 2'));

setTimeout(() => console.log('setTimeout 1'), 0);
setTimeout(() => console.log('setTimeout 2'), 0);

queueMicrotask(() => console.log('queueMicrotask'));

// Output: Promise 1, Promise 2, queueMicrotask, setTimeout 1, setTimeout 2
```

### Type Coercion and Truthiness
JavaScript's automatic type conversion and unique truthiness rules:

```javascript
// Type coercion examples
console.log('5' + 3); // '53' (string concatenation)
console.log('5' - 3); // 2 (numeric subtraction)
console.log('5' * 3); // 15 (numeric multiplication)
console.log('5' / 3); // 1.6666... (numeric division)

// Boolean coercion
console.log(Boolean('')); // false
console.log(Boolean('0')); // true (non-empty string)
console.log(Boolean(0)); // false
console.log(Boolean([])); // true (empty array)
console.log(Boolean({})); // true (empty object)
console.log(Boolean(null)); // false
console.log(Boolean(undefined)); // false

// Equality coercion
console.log([] == false); // true
console.log([] == 0); // true
console.log('0' == false); // true
console.log(null == undefined); // true
console.log(null === undefined); // false

// NaN peculiarity
console.log(NaN === NaN); // false
console.log(Number.isNaN(NaN)); // true
console.log(isNaN('hello')); // true (coerces 'hello' to NaN)
console.log(Number.isNaN('hello')); // false (no coercion)

// Object to primitive conversion
const obj = {
  valueOf: () => 42,
  toString: () => 'object'
};

console.log(obj + 1); // 43 (uses valueOf)
console.log(`${obj}`); // 'object' (uses toString)
```

### First-Class Functions
Functions are values that can be assigned, passed, and returned:

```javascript
// Functions as values
const functions = [
  function add(a, b) { return a + b; },
  function multiply(a, b) { return a * b; },
  (a, b) => a - b
];

const operation = functions[0];
console.log(operation(5, 3)); // 8

// Higher-order functions
function withLogging(fn) {
  return function(...args) {
    console.log(`Calling function with args: ${args}`);
    const result = fn(...args);
    console.log(`Function returned: ${result}`);
    return result;
  };
}

const loggedAdd = withLogging((a, b) => a + b);
loggedAdd(2, 3); // Logs and returns 5

// Function composition
const compose = (f, g) => (x) => f(g(x));
const addOne = x => x + 1;
const multiplyByTwo = x => x * 2;

const addOneThenMultiplyByTwo = compose(multiplyByTwo, addOne);
console.log(addOneThenMultiplyByTwo(3)); // 8 ((3 + 1) * 2)

// Currying
const curry = (fn) => {
  return function curried(...args) {
    if (args.length >= fn.length) {
      return fn(...args);
    } else {
      return (...nextArgs) => curried(...args, ...nextArgs);
    }
  };
};

const curriedAdd = curry((a, b, c) => a + b + c);
console.log(curriedAdd(1)(2)(3)); // 6
console.log(curriedAdd(1, 2)(3)); // 6
console.log(curriedAdd(1)(2, 3)); // 6
```

## TypeScript-Specific Features

### Structural Type System
TypeScript uses structural typing (duck typing) rather than nominal typing:

```typescript
// Structural typing example
interface Point {
  x: number;
  y: number;
}

interface Vector {
  x: number;
  y: number;
}

// These are compatible despite different names
const point: Point = { x: 1, y: 2 };
const vector: Vector = point; // No error

// Function that accepts any object with x and y
function distance(p: { x: number; y: number }): number {
  return Math.sqrt(p.x * p.x + p.y * p.y);
}

distance(point); // OK
distance(vector); // OK
distance({ x: 3, y: 4, z: 5 }); // OK (excess properties allowed in arguments)

// Excess property checking in object literals
const strictPoint: Point = { x: 1, y: 2, z: 3 }; // Error: excess property 'z'

// But this works (assignment to variable first)
const obj = { x: 1, y: 2, z: 3 };
const flexiblePoint: Point = obj; // OK
```

### Type Inference and Flow Control
TypeScript's sophisticated type inference system:

```typescript
// Basic inference
let x = 42; // inferred as number
let y = 'hello'; // inferred as string
let z = [1, 2, 3]; // inferred as number[]

// Contextual typing
window.addEventListener('click', (e) => {
  // 'e' is inferred as MouseEvent
  console.log(e.clientX, e.clientY);
});

// Best common type inference
let mixed = [1, 'hello', true]; // inferred as (string | number | boolean)[]

// Control flow analysis
function processValue(value: string | number) {
  if (typeof value === 'string') {
    // TypeScript knows value is string here
    return value.toUpperCase();
  } else {
    // TypeScript knows value is number here
    return value.toFixed(2);
  }
}

// Discriminated unions with control flow
type Result =
  | { success: true; data: string }
  | { success: false; error: string };

function handleResult(result: Result) {
  if (result.success) {
    // TypeScript knows result.data exists here
    console.log(result.data);
  } else {
    // TypeScript knows result.error exists here
    console.log(result.error);
  }
}

// Type narrowing with in operator
type Fish = { swim: () => void };
type Bird = { fly: () => void };

function move(animal: Fish | Bird) {
  if ('swim' in animal) {
    animal.swim(); // TypeScript knows it's Fish
  } else {
    animal.fly(); // TypeScript knows it's Bird
  }
}
```

### Advanced Type Manipulations
TypeScript's powerful type-level programming:

```typescript
// Conditional types
type NonNullable<T> = T extends null | undefined ? never : T;
type Example1 = NonNullable<string | null>; // string

// Mapped types
type Optional<T> = {
  [P in keyof T]?: T[P];
};

type PartialUser = Optional<{ name: string; age: number }>;
// Result: { name?: string; age?: number }

// Template literal types
type EventNames = 'click' | 'scroll' | 'focus';
type EventHandlers = {
  [K in EventNames as `on${Capitalize<K>}`]: (event: Event) => void;
};
// Result: { onClick: ..., onScroll: ..., onFocus: ... }

// Recursive types
type JsonValue =
  | string
  | number
  | boolean
  | null
  | JsonValue[]
  | { [key: string]: JsonValue };

// Utility types combination
type DeepReadonly<T> = {
  readonly [P in keyof T]: T[P] extends object ? DeepReadonly<T[P]> : T[P];
};

// Type-level string manipulation
type CamelCase<S extends string> = S extends `${infer P1}_${infer P2}${infer P3}`
  ? `${P1}${Uppercase<P2>}${CamelCase<P3>}`
  : S;

type CamelCased = CamelCase<'hello_world_foo'>; // 'helloWorldFoo'

// Higher-kinded types simulation
interface Functor<T> {
  map<U>(f: (x: T) => U): Functor<U>;
}

class Maybe<T> implements Functor<T> {
  constructor(private value: T | null) {}

  map<U>(f: (x: T) => U): Maybe<U> {
    return this.value === null
      ? new Maybe<U>(null)
      : new Maybe<U>(f(this.value));
  }

  static some<T>(value: T): Maybe<T> {
    return new Maybe(value);
  }

  static none<T>(): Maybe<T> {
    return new Maybe<T>(null);
  }
}
```

### Declaration Merging
TypeScript allows merging multiple declarations of the same name:

```typescript
// Interface merging
interface Window {
  customProperty: string;
}

interface Window {
  anotherProperty: number;
}

// Now Window has both properties
declare const window: Window;
window.customProperty = 'hello';
window.anotherProperty = 42;

// Namespace merging
namespace MyNamespace {
  export interface Config {
    apiUrl: string;
  }
}

namespace MyNamespace {
  export interface Config {
    timeout: number;
  }

  export function createClient(config: Config) {
    // config has both apiUrl and timeout
    return { url: config.apiUrl, timeout: config.timeout };
  }
}

// Module augmentation
declare module 'express' {
  interface Request {
    user?: { id: string; name: string };
  }
}

// Now all Express requests have optional user property
```

### Ambient Declarations
TypeScript's way to describe existing JavaScript:

```typescript
// Declare global variables
declare const VERSION: string;
declare const API_ENDPOINT: string;

// Declare global functions
declare function gtag(command: string, ...args: any[]): void;

// Declare modules
declare module 'my-library' {
  export function doSomething(): void;
  export interface Options {
    verbose: boolean;
  }
}

// Declare global augmentations
declare global {
  interface Window {
    myGlobalFunction: () => void;
  }

  namespace NodeJS {
    interface ProcessEnv {
      NODE_ENV: 'development' | 'production' | 'test';
      API_KEY: string;
    }
  }
}

// Using declared types
const version = VERSION; // TypeScript knows it's a string
gtag('config', 'GA_MEASUREMENT_ID');

// Triple-slash directives
/// <reference path="./custom.d.ts" />
/// <reference types="node" />
```

## Runtime vs Compile-Time Features

### TypeScript Compilation Process
TypeScript operates in two phases: type checking and compilation:

```typescript
// This TypeScript code...
interface User {
  name: string;
  age: number;
}

class UserService {
  private users: User[] = [];

  addUser(user: User): void {
    this.users.push(user);
  }

  findUser(name: string): User | undefined {
    return this.users.find(u => u.name === name);
  }
}

// Compiles to this JavaScript (roughly)
class UserService {
  constructor() {
    this.users = [];
  }

  addUser(user) {
    this.users.push(user);
  }

  findUser(name) {
    return this.users.find(u => u.name === name);
  }
}
```

### Runtime Type Checking
Since TypeScript types are erased at runtime, you need runtime validation:

```typescript
// Compile-time types don't exist at runtime
interface ApiResponse {
  data: string;
  status: number;
}

// This won't work at runtime
function processApiResponse(response: unknown): ApiResponse {
  // return response as ApiResponse; // Unsafe!

  // Need runtime validation
  if (
    typeof response === 'object' &&
    response !== null &&
    typeof (response as any).data === 'string' &&
    typeof (response as any).status === 'number'
  ) {
    return response as ApiResponse;
  }

  throw new Error('Invalid API response format');
}

// Runtime type validation with libraries
import { z } from 'zod';

const ApiResponseSchema = z.object({
  data: z.string(),
  status: z.number()
});

type ApiResponse = z.infer<typeof ApiResponseSchema>;

function processApiResponseSafe(response: unknown): ApiResponse {
  return ApiResponseSchema.parse(response); // Throws if invalid
}
```

### Unique Language Behaviors

JavaScript has several unique behaviors not found in most languages:

```javascript
// Automatic semicolon insertion
function returnObject() {
  return
  {
    name: 'John'
  }
}

console.log(returnObject()); // undefined (semicolon inserted after return)

// Function declarations vs expressions
console.log(declared()); // Works - function is hoisted
console.log(expressed()); // Error - variable is undefined

function declared() {
  return 'I am declared';
}

var expressed = function() {
  return 'I am expressed';
};

// With statements (avoid!)
const obj = { x: 1, y: 2 };
with (obj) {
  console.log(x + y); // 3 - but confusing and not allowed in strict mode
}

// Arguments object in functions
function oldStyleFunction() {
  console.log(arguments.length);
  console.log(Array.from(arguments));
}

oldStyleFunction(1, 2, 3); // 3, [1, 2, 3]

// Arrow functions don't have arguments
const modernFunction = (...args) => {
  console.log(args.length);
  console.log(args);
};

// delete operator quirks
const obj2 = { a: 1, b: 2 };
delete obj2.a; // true
console.log(obj2); // { b: 2 }

let x = 1;
delete x; // false in non-strict mode, SyntaxError in strict mode

// typeof null
console.log(typeof null); // 'object' (famous JavaScript quirk)
console.log(null instanceof Object); // false

// Array holes
const sparseArray = [1, , , 4];
console.log(sparseArray.length); // 4
console.log(sparseArray[1]); // undefined
console.log(1 in sparseArray); // false
console.log(sparseArray.map(x => x * 2)); // [2, empty, empty, 8]
```

## Ecosystem-Specific Features

### Module Systems Evolution
JavaScript/TypeScript supports multiple module systems:

```javascript
// CommonJS (Node.js)
const fs = require('fs');
const path = require('path');

module.exports = {
  readFile: fs.readFile,
  joinPath: path.join
};

// ES Modules
import fs from 'fs';
import { join } from 'path';

export const readFile = fs.readFile;
export const joinPath = join;

// AMD (RequireJS)
define(['fs', 'path'], function(fs, path) {
  return {
    readFile: fs.readFile,
    joinPath: path.join
  };
});

// UMD (Universal Module Definition)
(function(root, factory) {
  if (typeof define === 'function' && define.amd) {
    define(['fs', 'path'], factory);
  } else if (typeof module === 'object' && module.exports) {
    module.exports = factory(require('fs'), require('path'));
  } else {
    root.myModule = factory(root.fs, root.path);
  }
}(typeof self !== 'undefined' ? self : this, function(fs, path) {
  return {
    readFile: fs.readFile,
    joinPath: path.join
  };
}));
```

### Package.json Integration
TypeScript integrates deeply with npm's package.json:

```json
{
  "name": "my-typescript-project",
  "version": "1.0.0",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "exports": {
    ".": {
      "import": "./dist/index.esm.js",
      "require": "./dist/index.cjs.js",
      "types": "./dist/index.d.ts"
    },
    "./utils": {
      "import": "./dist/utils.esm.js",
      "require": "./dist/utils.cjs.js",
      "types": "./dist/utils.d.ts"
    }
  },
  "scripts": {
    "build": "tsc",
    "dev": "tsc --watch",
    "typecheck": "tsc --noEmit"
  },
  "devDependencies": {
    "typescript": "^5.0.0",
    "@types/node": "^18.0.0"
  }
}
```

### Browser vs Node.js Differences
JavaScript behaves differently in different environments:

```javascript
// Browser-specific
if (typeof window !== 'undefined') {
  // Browser environment
  console.log(window.location.href);
  document.getElementById('myElement');
  localStorage.setItem('key', 'value');

  // Web APIs
  fetch('/api/data');
  new WebSocket('ws://localhost:8080');
  navigator.geolocation.getCurrentPosition();
}

// Node.js-specific
if (typeof process !== 'undefined') {
  // Node.js environment
  console.log(process.env.NODE_ENV);
  console.log(__filename, __dirname);

  const fs = require('fs');
  const path = require('path');

  // Node.js APIs
  fs.readFileSync('file.txt');
  process.exit(0);
}

// Universal detection
const isNode = typeof process !== 'undefined' &&
               process.versions != null &&
               process.versions.node != null;

const isBrowser = typeof window !== 'undefined' &&
                  typeof document !== 'undefined';

// Polyfills for cross-platform compatibility
if (!globalThis.fetch) {
  globalThis.fetch = require('node-fetch');
}

// TypeScript environment detection
declare const process: any;
declare const window: any;

if (typeof process !== 'undefined') {
  // Node.js TypeScript code
}

if (typeof window !== 'undefined') {
  // Browser TypeScript code
}
```

These unique features make TypeScript and JavaScript distinct from other programming languages, providing both powerful capabilities and unique challenges that developers must understand to write effective code.